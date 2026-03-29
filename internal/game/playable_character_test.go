package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestPlayableCharacterStats(t *testing.T) {
	mc := &Character{
		Actor: Actor{
			Level: 1,
			Config: &EntityConfig{
				Attributes: PrimaryAttributeConfig{Strength: IntInterval{Min: 5, Max: 5}},
				Stats:      EntityStatsConfig{HealthMin: IntInterval{Min: 10, Max: 10}},
			},
			PrimaryAttributes: PrimaryAttributes{Strength: 5},
			AgeTicks:          25.0 * float64(TicksPerYear),
		},
	}
	mc.SyncStats(NewObjectRegistry())

	if att := mc.GetTotalAttack(); att != 10 {
		t.Errorf("GetTotalAttack(Level 1): got %d, want 10", att)
	}

	mc.Level = 10
	// 10 * 1.15^9 = 10 * 3.5178 = 35
	if att := mc.GetTotalAttack(); att != 35 {
		t.Errorf("GetTotalAttack(Level 10): got %d, want 35", att)
	}
}

func TestPlayableCharacterXPAndLevelUp(t *testing.T) {
	mc := &Character{
		Actor: Actor{
			Level: 1,
			XP:    0,
		},
	}

	mc.AddXP(50)
	if mc.XP != 50 {
		t.Errorf("XP after 50: got %d, want 50", mc.XP)
	}
	if mc.Level != 1 {
		t.Errorf("Level after 50 XP: got %d, want 1", mc.Level)
	}

	mc.AddXP(100)
	if mc.Level != 2 {
		t.Errorf("Level after 150 XP: got %d, want 2", mc.Level)
	}
	if mc.XP != 150 {
		t.Errorf("XP after Level Up: got %d, want 150", mc.XP)
	}
}

func TestPlayableCharacterTakeDamage(t *testing.T) {
	ctx := NewTestContext()
	mc := &Character{Actor: Actor{
		State: State{
			HealthPoints:    100,
			MaxHealthPoints: 100,
		},
		Config: &EntityConfig{ID: "player"},
	}}
	mc.TakeDamage(20, nil, ctx)
	if mc.State.HealthPoints != 80 {
		t.Errorf("Health after damage: got %d, want 80", mc.State.HealthPoints)
	}
	mc.TakeDamage(80, nil, ctx)
	if mc.State.HealthPoints != 0 || mc.ActionState != ActorIncapacitated {
		t.Errorf("Health after fatal damage: got %d, state=%v, want 0, state=ActorIncapacitated", mc.State.HealthPoints, mc.ActionState)
	}
	if !mc.IsAlive() {
		t.Error("Character should still be 'alive' (incapacitated) at 0 HP")
	}

	// Death threshold (-10% of 100 = -10)
	mc.TakeDamage(10, nil, ctx)
	if mc.State.HealthPoints != -10 || mc.ActionState != ActorDead {
		t.Errorf("Health after irremediable damage: got %d, state=%v, want -10, state=ActorDead", mc.State.HealthPoints, mc.ActionState)
	}
	if mc.IsAlive() {
		t.Error("Character should be truly dead at -10 HP")
	}
}

func TestPlayableCharacterGetters(t *testing.T) {
	mc := &Character{Actor: Actor{BaseDefense: 10}}
	if mc.GetTotalDefense() != 10 {
		t.Errorf("GetTotalDefense: got %d, want 10", mc.GetTotalDefense())
	}
	if mc.GetTotalProtection() != 0 {
		t.Errorf("GetTotalProtection: got %d, want 0", mc.GetTotalProtection())
	}

	mc.Slots = make(map[string]*ItemInstance)
	armorConfig := &ObjectConfig{
		Slot: "body",
		Effects: map[string]StatEffect{
			"protection": {Increase: 5},
		},
	}
	mc.Slots["body"] = NewItemInstance("leather_armor", armorConfig, 0, 0)
	mc.UpdateEffects()
	if mc.GetTotalProtection() != 5 {
		t.Errorf("GetTotalProtection with armor: got %d, want 5", mc.GetTotalProtection())
	}
}

func TestPlayableCharacterCheckAttackHits(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.Weapon = &Weapon{Name: "TestWeapon", Damage: Damage{Min: 10, Max: 10}}
	mc.Facing = DirSE

	npc := &Character{Actor: Actor{X: 1, Y: 0.5, ActionState: ActorIdle}}
	ctx.World.Characters = []*Character{npc}
	ctx.World.PlayableCharacter = mc
	mc.CheckAttackHits(ctx, "")
}

func TestPlayableCharacterCollisionCircle(t *testing.T) {
	mc := NewCharacter(10, 10, nil, 1, true, nil)
	c := mc.GetCollisionCircle()
	if c.Radius <= 0 {
		t.Error("Collision circle should have radius > 0")
	}
}

func TestPlayableCharacterCollision(t *testing.T) {
	mc := NewCharacter(10, 10, nil, 1, true, nil)
	colliders := []*Obstacle{NewObstacle("test_mc_collider", 10.5, 10.5, &ObstacleArchetype{Passable: false})}
	if !mc.checkCollisionAt(10.5, 10.5, colliders) {
		t.Error("Expected collision at 10.5, 10.5")
	}
	if mc.checkCollisionAt(20, 20, colliders) {
		t.Error("Expected no collision at 20, 20")
	}
}

func TestPlayableCharacterUpdate_Full(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.State.HealthPoints = mc.State.MaxHealthPoints
	ctx.World.PlayableCharacter = mc
	mockInput := ctx.Input.(*MockInputManager)

	// Update when dead
	mc.ActionState = ActorDead
	mc.Update(ctx)
	if mc.ActionState != ActorDead {
		t.Error("Dead mc should stay dead")
	}

	// Update drinking
	mc.ActionState = ActorDrinking
	mc.Tick = 0
	mc.State.Thirst = 50.0
	mc.Update(ctx)
	if mc.ActionState != ActorDrinking {
		t.Error("Should stay drinking")
	}
	mc.Tick = 60
	mc.Update(ctx)
	if mc.ActionState != ActorIdle {
		t.Error("Should be idle after drink timer")
	}

	// Update attacking
	mc.ActionState = ActorAttacking
	mc.Tick = 14
	mc.Update(ctx)
	if mc.Tick != 15 {
		t.Error("Tick should advance")
	}
	mc.Tick = 30
	mc.Update(ctx)
	if mc.ActionState != ActorIdle {
		t.Error("Should be idle after attack anim")
	}

	// Movement Input checks
	mc.ActionState = ActorIdle
	mockInput.PressedKeys[engine.KeyW] = true
	mockInput.PressedKeys[engine.KeyD] = true
	mc.Update(ctx)
	if mc.ActionState != ActorWalking {
		t.Error("Should be walking on input")
	}
	if mc.Facing != DirNE {
		t.Errorf("Expected Facing DirNE, got %v", mc.Facing)
	}

	mc.X = 0
	mc.Y = 0
	delete(mockInput.PressedKeys, engine.KeyW)
	mockInput.PressedKeys[engine.KeyS] = true
	mc.Update(ctx)
	if mc.Facing != DirSE {
		t.Errorf("Expected Facing DirSE, got %v", mc.Facing)
	}

	// Test clamp boundaries
	mc.X = 1000
	mc.Y = 1000
	ctx.World.CurrentMapType = &MapType{MapWidth: 100, MapHeight: 100}
	mockInput.PressedKeys[engine.KeyD] = true // Move right edge
	mc.Update(ctx)
	if mc.X > 50 || mc.Y > 50 {
		t.Error("Position not clamped correctly")
	}
}
