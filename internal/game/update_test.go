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
	g := NewGame(mockFS, "type1", "", NewMockInputManager(), NewMockAudioManager(), false)

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
	g.npcs = []*NPC{{State: NPCDead}}
	g.projectiles = []*Projectile{{Alive: false}}
	g.floatingTexts = []*FloatingText{{Life: 0}}

	if err := g.Update(); err != nil {
		t.Errorf("Update failed during cleanup: %v", err)
	}

	if len(g.npcs) != 1 || len(g.projectiles) != 0 || len(g.floatingTexts) != 0 {
		t.Errorf("Cleanup failed: npcs=%d (expected 1), projectiles=%d (expected 0), texts=%d (expected 0)", len(g.npcs), len(g.projectiles), len(g.floatingTexts))
	}
}

func TestMainCharacterUpdate_Detailed(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	mc.Weapon = WeaponTizon

	// Test drinking state
	mc.State = StateDrinking
	mc.Tick = 0
	fts := []*FloatingText{}
	mc.Update(NewMockInputManager(), nil, nil, nil, &fts, 100, 100)
	if mc.State != StateDrinking {
		t.Error("Should stay in drinking state")
	}
	mc.Tick = 60
	mc.Update(NewMockInputManager(), nil, nil, nil, &fts, 100, 100)
	if mc.State != StateIdle {
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
	g := NewGame(mockFS, "type1", "", mockInput, NewMockAudioManager(), false)

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
	n := NewNPC(0, 0, nil, 1)
	n.Weapon = WeaponTizon
	n.Speed = 1.0 // Manually set speed since Archetype is nil
	mc := NewMainCharacter(10, 10, nil)
	fts := []*FloatingText{}
	projs := []*Projectile{}

	// Test hunter behavior
	n.Behavior = BehaviorKnightHunter
	n.Update(mc, nil, nil, &projs, &fts, 100, 100, nil)
	// Should move towards mc
	if n.X == 0 && n.Y == 0 {
		t.Error("Hunter NPC should move")
	}

	// Test fighter behavior with other NPCs
	otherNpc := NewNPC(5, 5, nil, 1)
	otherNpc.Alignment = AlignmentAlly // Different alignment from n (Enemy)
	npcs := []*NPC{otherNpc, n}        // allNPCs includes self
	n.Behavior = BehaviorNpcFighter
	n.X = 0
	n.Y = 0
	n.TargetPlayer = nil
	n.TargetNPC = nil
	n.Update(mc, nil, npcs, &projs, &fts, 100, 100, nil)
	if n.X == 0 && n.Y == 0 {
		t.Error("Fighter NPC should move towards other NPC")
	}

	// Test attack branch
	n.X = otherNpc.X + 0.1
	n.Y = otherNpc.Y + 0.1
	n.TargetNPC = otherNpc // Ensure it still targets otherNpc
	n.TargetPlayer = nil
	n.AttackTimer = 0
	n.Update(mc, nil, npcs, &projs, &fts, 100, 100, nil)
	if n.State != NPCAttacking {
		t.Errorf("NPC should be attacking, got state %v", n.State)
	}
}

func TestCollisionDetailed(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	obs := []*Obstacle{NewObstacle("test_obs_1", 1, 0, &ObstacleArchetype{ID: "test", Footprint: []FootprintPoint{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}})}

	// Test collision detection
	if !mc.checkCollisionAt(1, 0, obs) {
		t.Error("Should detect collision with obstacle")
	}
}

func TestNPCHitBranch_Detailed(t *testing.T) {
	n := NewNPC(0, 0, nil, 1)
	n.Weapon = WeaponTizon
	mc := NewMainCharacter(1, 0, nil)
	projs := []*Projectile{}
	fts := []*FloatingText{}
	npcs := []*NPC{n}

	// Force a hit
	n.AttackTimer = 0
	n.State = NPCIdle
	n.BaseAttack = 1000
	mc.BaseDefense = 0

	n.Update(mc, nil, npcs, &projs, &fts, 100, 100, nil)
}

func TestMainCharacterTakeDamageDetailed(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	mc.Health = 100
	mc.TakeDamage(150, nil)
	if mc.Health != 0 || mc.State != StateDead {
		t.Errorf("Should be dead, health=%d, state=%v", mc.Health, mc.State)
	}

	// Take damage while dead
	mc.TakeDamage(10, nil)
	if mc.Health != 0 {
		t.Error("Health should stay 0")
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
	p := NewProjectile(0, 0, 1, 0, 1.0, 10, true, 100.0)
	mc := NewMainCharacter(0, 0, nil)
	fts := []*FloatingText{}

	// Update until it hits nothing or expires
	p.Update(mc, nil, &fts)

	// Update with entities
	targetMc := NewMainCharacter(2, 0, nil)
	obstacles := []*Obstacle{NewObstacle("test_obs_3", 5, 0, &ObstacleArchetype{ID: "test", Footprint: []FootprintPoint{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}}})}

	// Manually move projectile to hit targetMc
	p.X = 2
	p.Y = 0
	p.Alive = true
	p.Update(targetMc, obstacles, &fts)

	// Manually move projectile to hit obstacle
	p.Alive = true
	p.X = 5
	p.Y = 0
	p.Update(targetMc, obstacles, &fts)
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
	g := NewGame(mockFS, "duel", "", NewMockInputManager(), NewMockAudioManager(), false)

	// Spawn a boss (VIP)
	boss := NewNPC(5, 5, nil, 10)
	g.npcs = []*NPC{boss}

	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if g.isMapWon {
		t.Error("Map should not be won yet")
	}

	// Kill the boss
	boss.State = NPCDead
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
	g := NewGame(mockFS, "test", "", NewMockInputManager(), NewMockAudioManager(), false)
	mc := g.mainCharacter

	npc := NewNPC(0, 0, &Archetype{
		ID: "test_npc",
		Stats: struct {
			HealthMin       int     `yaml:"health_min"`
			HealthMax       int     `yaml:"health_max"`
			Speed           float64 `yaml:"speed"`
			BaseAttack      int     `yaml:"base_attack"`
			BaseDefense     int     `yaml:"base_defense"`
			AttackCooldown  int     `yaml:"attack_cooldown"`
			AttackRange     float64 `yaml:"attack_range"`
			ProjectileSpeed float64 `yaml:"projectile_speed"`
		}{HealthMin: 5, HealthMax: 5, BaseDefense: 0},
	}, 1)

	npc.Health = 5
	g.npcs = []*NPC{npc}

	// Deal fatal damage
	npc.TakeDamage(100, mc, nil, NewMockAudioManager(), []*NPC{npc})

	if npc.State != NPCDead {
		t.Fatalf("NPC should be dead")
	}

	g.Update() // Run one frame of the game loop

	found := false
	for _, n := range g.npcs {
		if n == npc {
			found = true
		}
	}

	if !found {
		t.Fatalf("NPC Corpse was deleted from g.npcs during Update() loop!")
	}
}
