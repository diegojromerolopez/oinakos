package game

import (
	"oinakos/internal/engine"
	"testing"
	"testing/fstest"
)

func init() {
	isTestingEnvironment = true
}

func TestGame_Update(t *testing.T) {
	mockFS := fstest.MapFS{
		"data/map_types/type1.yaml": {
			Data: []byte(`id: "type1"
name: "Type One"
type: "kill_count"
difficulty: 1
width_px: 1000
height_px: 1000
`),
		},
	}
	g := NewGame(mockFS, &engine.MockGraphics{}, "type1", "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	g.isMainMenu = false
	g.isCharacterSelect = false

	// 1. Test Paused state
	g.isPaused = true
	if err := g.Update(); err != nil {
		t.Errorf("Update returned error while paused: %v", err)
	}

	// 2. Test Game Over state
	g.isPaused = false
	g.isGameOver = true
	if err := g.Update(); err != nil {
		t.Errorf("Update returned error while game over: %v", err)
	}

	// 3. Test Map Won state
	g.isGameOver = false
	g.isMapWon = true
	if err := g.Update(); err != nil {
		t.Errorf("Update returned error while map won: %v", err)
	}

	// 4. Test Normal Update
	g.isMapWon = false
	g.npcSpawnTimer = 0
	g.currentMapType.Spawns = nil // Ensure no auto-spawning
	if err := g.Update(); err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// 5. Test Entity Cleanup (Corpses should be retained, others cleaned up)
	g.characters = []*Character{{Actor: Actor{ActionState: ActorDead}}}
	g.projectiles = []*Projectile{{Alive: false}}
	g.floatingTexts = []*FloatingText{{Life: 0}}

	if err := g.Update(); err != nil {
		t.Errorf("Update failed during cleanup: %v", err)
	}

	if len(g.characters) != 1 || len(g.projectiles) != 0 || len(g.floatingTexts) != 0 {
		t.Errorf("Cleanup failed: npcs=%d (expected 1), projectiles=%d (expected 0), texts=%d (expected 0)", len(g.characters), len(g.projectiles), len(g.floatingTexts))
	}
}

func TestPlayableCharacterUpdate_Detailed(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.Weapon = WeaponTizon
	ctx.World.PlayableCharacter = mc

	// Test drinking state
	mc.ActionState = ActorDrinking
	mc.Tick = 0
	mc.State.Thirst = 50.0 // Ensure it doesn't immediately finish
	mc.Update(ctx)
	if mc.ActionState != ActorDrinking {
		t.Error("Should stay in drinking state")
	}
	mc.Tick = 60
	mc.Update(ctx)
	if mc.ActionState != ActorIdle {
		t.Error("Should transition to idle after drinking")
	}
}

func TestGame_BoundariesToggle(t *testing.T) {
	mockFS := fstest.MapFS{
		"data/map_types/type1.yaml": {
			Data: []byte(`id: "type1"
name: "Type One"
type: "all"
difficulty: 1
width_px: 1000
height_px: 1000
`),
		},
	}
	mockInput := NewMockInputManager()
	g := NewGame(mockFS, &engine.MockGraphics{}, "type1", "", "", mockInput, NewMockAudioManager(), false, "0.1-test")
	g.isMainMenu = false
	g.isCharacterSelect = false

	// Test Initial state
	if g.showBoundaries {
		t.Error("Initially showBoundaries should be false")
	}

	// 1. Test Toggle ON during Game
	mockInput.JustPressedKeys[engine.KeyTab] = true
	g.Update()
	if !g.showBoundaries {
		t.Error("showBoundaries should be true after first Tab press")
	}

	// Reset mock input for next update
	mockInput.JustPressedKeys[engine.KeyTab] = false
	g.Update()
	if !g.showBoundaries {
		t.Error("showBoundaries should stay true if Tab is NOT pressed")
	}

	// 2. Test Toggle OFF during Game
	mockInput.JustPressedKeys[engine.KeyTab] = true
	g.Update()
	if g.showBoundaries {
		t.Error("showBoundaries should be false after second Tab press")
	}

	// 3. Test Toggle while Paused
	g.isPaused = true
	mockInput.JustPressedKeys[engine.KeyTab] = true
	g.Update()
	if !g.showBoundaries {
		t.Error("showBoundaries should toggle even when game is paused")
	}

	// 4. Test Toggle during Game Over
	g.isPaused = false
	g.isGameOver = true
	mockInput.JustPressedKeys[engine.KeyTab] = true
	g.Update()
	if g.showBoundaries {
		t.Error("showBoundaries should toggle to false during GameOver")
	}
}

func TestNPCUpdate_Detailed(t *testing.T) {
	ctx := NewTestContext()
	n := NewCharacter(0, 0, nil, 1, false, nil)
	n.State.HealthPoints = 100
	n.Weapon = WeaponTizon
	n.Speed = 1.0 // Manually set speed since Archetype is nil
	mc := NewCharacter(10, 10, nil, 1, true, nil)
	mc.State.HealthPoints = 100
	ctx.World.PlayableCharacter = mc
	ctx.World.Characters = []*Character{n}

	// Test hunter behavior
	n.Behavior = BehaviorKnightHunter
	n.Update(ctx)
	// Should move towards mc
	if n.X == 0 && n.Y == 0 {
		t.Error("Hunter NPC should move")
	}

	// Test fighter behavior with other NPCs
	otherNpc := NewCharacter(5, 5, nil, 1, false, nil)
	otherNpc.Alignment = AlignmentAlly // Different alignment from n (Enemy)
	ctx.World.Characters = []*Character{otherNpc, n}        // allNPCs includes self
	n.Behavior = BehaviorNpcFighter
	n.X = 0
	n.Y = 0
	n.TargetActor = nil
	n.Update(ctx)
	if n.X == 0 && n.Y == 0 {
		t.Error("Fighter NPC should move towards other NPC")
	}

	// Test attack branch
	n.X = otherNpc.X + 0.1
	n.Y = otherNpc.Y + 0.1
	n.TargetActor = &otherNpc.Actor // Ensure it still targets otherNpc
	n.AttackTimer = 0
	n.Update(ctx)
	if n.ActionState != ActorAttacking {
		t.Errorf("NPC should be attacking, got state %v", n.ActionState)
	}
}

func TestCollisionDetailed(t *testing.T) {
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	obs := []*Obstacle{NewObstacle("test_obs_1", 1, 0, &ObstacleArchetype{ID: "test", Footprint: []FootprintPoint{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}})}

	// Test collision detection
	if !mc.checkCollisionAt(1, 0, obs) {
		t.Error("Should detect collision with obstacle")
	}
}

func TestNoSlidingMovement(t *testing.T) {
	ctx := NewTestContext()
	// Place character at 0,0
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.Speed = 1.0
	ctx.World.PlayableCharacter = mc

	// Place an obstacle at 1.2, 1.2 with 1x1 footprint (X range [0.7, 1.7], Y range [0.7, 1.7])
	obs := NewObstacle("test_obs", 1.2, 1.2, &ObstacleArchetype{
		ID: "test",
		Footprint: []FootprintPoint{
			{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5},
		},
	})
	ctx.World.Obstacles = []*Obstacle{obs}
	ctx.World.CurrentMapType = &MapType{MapWidth: 100, MapHeight: 100}

	// Try to move diagonally towards the obstacle
	ctx.Input.(*MockInputManager).PressedKeys[engine.KeyD] = true
	ctx.Input.(*MockInputManager).PressedKeys[engine.KeyS] = true

	// Initial position
	oldX, oldY := mc.X, mc.Y

	// Update
	mc.Update(ctx)

	// If it collided, X and Y should be exactly oldX and oldY
	if mc.X != oldX || mc.Y != oldY {
		t.Errorf("Character should have stopped at (%f, %f), but moved to (%f, %f)", oldX, oldY, mc.X, mc.Y)
	}

	if mc.ActionState != ActorIdle {
		t.Errorf("Character should be ActorIdle on collision, got %v", mc.ActionState)
	}
	
	if mc.Tick != 0 {
		t.Errorf("Character Tick should be reset to 0 on collision, got %d", mc.Tick)
	}
}

func TestNPCHitBranch_Detailed(t *testing.T) {
	ctx := NewTestContext()
	n := NewCharacter(0, 0, nil, 1, false, nil)
	n.State.HealthPoints = 100
	n.Weapon = WeaponTizon
	mc := NewCharacter(1, 0, nil, 1, true, nil)
	mc.State.HealthPoints = 100
	ctx.World.PlayableCharacter = mc
	ctx.World.Characters = []*Character{n}

	// Force a hit
	n.AttackTimer = 0
	n.ActionState = ActorIdle
	n.BaseAttack = 1000
	mc.BaseDefense = 0

	n.Update(ctx)
}

func TestPlayableCharacterTakeDamageDetailed(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.State.HealthPoints = 100
	mc.State.MaxHealthPoints = 100
	mc.TakeDamage(150, nil, ctx)
	if mc.State.HealthPoints != -10 || mc.ActionState != ActorDead {
		t.Errorf("Should be dead, health=%d, state=%v", mc.State.HealthPoints, mc.ActionState)
	}

	// Take damage while dead
	mc.TakeDamage(10, nil, ctx)
	if mc.State.HealthPoints != -10 { // Health should stay at the dead threshold
		t.Errorf("Health should stay at -10 (dead threshold), got %d", mc.State.HealthPoints)
	}
	if mc.ActionState != ActorDead {
		t.Errorf("State should remain ActorDead, got %v", mc.ActionState)
	}
}

func TestObstacleUpdate_Detailed(t *testing.T) {
	o := NewObstacle("test_obs_2", 0, 0, nil)
	o.Update()

	o.CooldownTicks = 10
	o.Update()
	if o.CooldownTicks != 9 {
		t.Errorf("Obstacle Update: CooldownTicks should decrease, got %d", o.CooldownTicks)
	}
}

func TestProjectileUpdate_Detailed(t *testing.T) {
	ctx := NewTestContext()
	p := NewProjectile(0, 0, 1, 0, 1.0, 10, true, 100.0)
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	ctx.World.PlayableCharacter = mc

	// Update until it hits nothing or expires
	p.Update(ctx)

	// Update with entities
	targetMc := NewCharacter(2, 0, nil, 1, true, nil)
	ctx.World.PlayableCharacter = targetMc
	obstacles := []*Obstacle{NewObstacle("test_obs_3", 5, 0, &ObstacleArchetype{ID: "test", Footprint: []FootprintPoint{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}}})}
	ctx.World.Obstacles = obstacles

	// Manually move projectile to hit targetMc
	p.X = 2
	p.Y = 0
	p.Alive = true
	p.Update(ctx)

	// Manually move projectile to hit obstacle
	p.Alive = true
	p.X = 5
	p.Y = 0
	p.Update(ctx)
}

func TestFloatingTextUpdate_Detailed(t *testing.T) {
	ft := &FloatingText{Life: 10}
	alive := ft.Update()
	if !alive {
		t.Error("FloatingText should be alive")
	}
	if ft.Life != 9 {
		t.Errorf("FloatingText life: got %d, want 9", ft.Life)
	}

	ft.Life = 1
	alive = ft.Update()
	if alive {
		t.Error("FloatingText should be finished")
	}
}

func TestObjKillVIP(t *testing.T) {
	mockFS := fstest.MapFS{
		"data/map_types/duel.yaml": {
			Data: []byte(`id: "duel"
name: "Demon Duel"
type: "kill_vip"
difficulty: 5
spawn_frequency: 0
`),
		},
	}
	g := NewGame(mockFS, &engine.MockGraphics{}, "duel", "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	g.isMainMenu = false
	g.isCharacterSelect = false

	// Spawn a boss (VIP)
	boss := NewCharacter(5, 5, nil, 10, false, nil)
	g.characters = []*Character{boss}

	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if g.isMapWon {
		t.Error("Map should not be won yet")
	}

	// Kill the boss
	boss.ActionState = ActorDead
	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if !g.isMapWon {
		t.Error("Map should be won after VIP death")
	}
}

func TestCombatCorpseRetention(t *testing.T) {
	mockFS := fstest.MapFS{
		"data/map_types/test.yaml": {
			Data: []byte(`id: "test"
name: "Test"
type: "kill_count"
difficulty: 1
`),
		},
	}
	g := NewGame(mockFS, &engine.MockGraphics{}, "test", "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	g.isMainMenu = false
	g.isCharacterSelect = false
	mc := g.playableCharacter

	npc := NewCharacter(0, 0, &EntityConfig{
		ID: "test_npc",
		Stats: EntityStatsConfig{
			HealthMin: IntInterval{Min: 5, Max: 5},
			HealthMax: IntInterval{Min: 5, Max: 5},
			BaseDefense: IntInterval{Min: 0, Max: 0},
		},
	}, 1, false, nil)

	npc.State.HealthPoints = 5
	g.characters = []*Character{npc}

	// Deal fatal damage
	ctx := &SystemContext{
		World: &World{PlayableCharacter: mc, Characters: []*Character{npc}},
		Audio: NewMockAudioManager(),
	}
	npc.TakeDamage(100, mc, ctx)

	if npc.ActionState != ActorDead {
		t.Fatalf("NPC should be dead")
	}

	g.Update() // Run one frame of the game loop

	found := false
	for _, n := range g.characters {
		if n == npc {
			found = true
		}
	}

	if !found {
		t.Fatalf("NPC Corpse was deleted from g.characters during Update() loop!")
	}
}
