package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestPickupRawMeatFromGround(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter
	pc.X, pc.Y = 0, 0

	// 1. Manually add "raw_meat" to registry (as setupTestGame doesn't load YAMLs)
	meatConfig := &ObjectConfig{
		ID:     "raw_meat",
		Name:   "Raw Meat",
		Weight: 2.5,
	}
	g.Registries.Objects.Objects["raw_meat"] = meatConfig

	// 2. Place raw meat on the ground near the player
	item := NewItemInstance("raw_meat", meatConfig, 1.0, 1.0) // Within 2.5 pickup range
	item.Pickable = true
	g.World.Items = append(g.World.Items, item)

	// 3. Mock Spacebar input
	mockInput := &BetterMockInput{}
	mockInput.JustPressedKeys = []engine.Key{engine.KeySpace}
	g.input = mockInput

	// Initial checks
	if len(g.World.Items) != 1 {
		t.Fatalf("World should have 1 item, got %d", len(g.World.Items))
	}
	if len(pc.Inventory) != 0 {
		t.Fatalf("Player inventory should be empty, got %d items", len(pc.Inventory))
	}

	// 4. Update pickups
	g.UpdatePickups()

	// 5. Verify pickup
	if len(g.World.Items) != 0 {
		t.Errorf("Item should have been picked up from the ground")
	}
	if len(pc.Inventory) != 1 {
		t.Errorf("Player should have 1 item in inventory")
	} else {
		if pc.Inventory[0].Config.ID != "raw_meat" {
			t.Errorf("Expected raw_meat in inventory, got %s", pc.Inventory[0].Config.ID)
		}
	}
}
