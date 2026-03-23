package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestPlayerUpdate_Comprehensive(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.Speed = 1.0
	ctx.World.PlayableCharacter = mc
	ctx.World.CurrentMapType = &MapType{MapWidth: 100, MapHeight: 100}
	input := ctx.Input.(*MockInputManager)

	// 1. Test Resting toggle (KeyR)
	input.JustPressedKeys[engine.KeyR] = true
	mc.updatePlayer(ctx)
	if mc.State != ActorResting {
		t.Error("Player should be resting after KeyR")
	}
	input.JustPressedKeys[engine.KeyR] = false
	
	// Break rest by moving
	input.PressedKeys[engine.KeyW] = true
	mc.updatePlayer(ctx)
	if mc.State != ActorWalking {
		t.Error("Player should break rest when a movement key is pressed and transition to walking")
	}
	input.PressedKeys[engine.KeyW] = false

	// 2. Test Healing Interaction (KeySpace near a well)
	wellArch := &ObstacleArchetype{
		ID: "well",
		Actions: []ObstacleActionConfig{
			{Type: ActionHeal, RequiresInteraction: true, Amount: 10},
		},
		CooldownTime: 1.0,
	}
	well := NewObstacle("well1", 1, 1, wellArch)
	ctx.World.Obstacles = []*Obstacle{well}
	
	mc.X, mc.Y = 0.5, 0.5
	mc.Health = 50
	mc.MaxHealth = 100
	input.PressedKeys[engine.KeySpace] = true
	mc.updatePlayer(ctx)
	if mc.Health != 60 {
		t.Errorf("Healing failed: got %d, want 60", mc.Health)
	}
	if mc.State != ActorDrinking {
		t.Errorf("Expected ActorDrinking state, got %v", mc.State)
	}
	input.PressedKeys[engine.KeySpace] = false
	mc.State = ActorIdle // Reset for next tests

	// 3. Test Axe auto-equip (KeyC)
	axeConfig := &ObjectConfig{ID: "axe", Name: "Lumberjack Axe", Combat: &Weapon{Name: "Axe", Damage: Damage{Min: 5, Max: 5}}}
	mc.Inventory = []*ItemInstance{NewItemInstance("axe", axeConfig, 0, 0)}
	input.JustPressedKeys[engine.KeyC] = true
	mc.updatePlayer(ctx)
	if mc.Weapon == nil || mc.Weapon.Name != "Axe" {
		t.Error("Axe should be equipped automatically")
	}
	if mc.State != ActorChopping {
		t.Errorf("Expected ActorChopping state, got %v", mc.State)
	}
	input.JustPressedKeys[engine.KeyC] = false
	mc.State = ActorIdle

	// 4. Test Pike auto-equip (KeyV)
	pikeConfig := &ObjectConfig{ID: "pike", Name: "Mining Pickaxe", Combat: &Weapon{Name: "Pike", Damage: Damage{Min: 5, Max: 5}}}
	mc.Inventory = []*ItemInstance{NewItemInstance("pike", pikeConfig, 0, 0)}
	input.JustPressedKeys[engine.KeyV] = true
	mc.updatePlayer(ctx)
	if mc.Weapon == nil || mc.Weapon.Name != "Pike" {
		t.Error("Pike should be equipped automatically")
	}
	if mc.State != ActorDigging {
		t.Errorf("Expected ActorDigging state, got %v", mc.State)
	}
	input.JustPressedKeys[engine.KeyV] = false
	mc.State = ActorIdle

	// 5. Test Movement Sounds on different tiles
	mc.X, mc.Y = 0, 0
	input.PressedKeys[engine.KeyD] = true
	mc.Tick = 30 // trigger sound (Tick % 30 == 0)
	
	// Water
	mc.CurrentTile = "water.png"
	mc.updatePlayer(ctx)
	// Stone
	mc.Tick = 60
	mc.CurrentTile = "paved_ground.png"
	mc.updatePlayer(ctx)
	
	input.PressedKeys[engine.KeyD] = false

	// 6. Test Map Clamping
	mc.X = 49.9
	mc.Y = 0
	input.PressedKeys[engine.KeyD] = true
	mc.Speed = 10.0
	mc.updatePlayer(ctx) // moved past 50.0 (half width)
	if mc.X > 50.0 {
		t.Errorf("X should be clamped to 50.0, got %f", mc.X)
	}
	input.PressedKeys[engine.KeyD] = false
}
