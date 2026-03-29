package game

import (
	"testing"
	"oinakos/internal/engine"
)

type mockInput struct{}

func (m *mockInput) IsKeyPressed(key engine.Key) bool { return false }
func (m *mockInput) IsKeyJustPressed(key engine.Key) bool { return key == engine.KeySpace }
func (m *mockInput) AppendJustPressedKeys(keys []engine.Key) []engine.Key { return keys }
func (m *mockInput) AppendInputChars(chars []rune) []rune { return chars }
func (m *mockInput) MousePosition() (x, y int) { return 0, 0 }
func (m *mockInput) IsMouseButtonPressed(button engine.MouseButton) bool { return false }
func (m *mockInput) IsMouseButtonJustPressed(button engine.MouseButton) bool { return false }
func (m *mockInput) Wheel() (x, y float64) { return 0, 0 }
func (m *mockInput) SetCursorMode(mode engine.CursorMode) {}

func TestObjectLifecycle(t *testing.T) {

	g := &Game{
		width:  1280,
		height: 720,
		Registries: &RegistryContainer{
			Objects: NewObjectRegistry(),
		},
	}
	g.World = NewWorld()
	g.World.Game = g
	sword := &ObjectConfig{
		ID: "test_sword",
		Name: "Test Sword",
		Weight: 5.0,
		Type: "weapon",
		Slot: "weapon",
		Combat: &Weapon{Name: "Test Sword", Damage: Damage{Min: 5, Max: 10}},
	}
	g.Registries.Objects.Objects[sword.ID] = sword
	
	// Setup character
	config := &EntityConfig{ID: "hero", Stats: EntityStatsConfig{MaxWeight: FloatInterval{Min: 10.0, Max: 10.0}}}
	char := NewCharacter(0, 0, config, 1, true, nil)
	g.playableCharacter = char
	g.World.PlayableCharacter = char
	
	// 1. Test Pickup
	it := NewItemInstance(sword.ID, sword, 0.1, 0.1)
	g.World.Items = []*ItemInstance{it}
	
	// Create mock input
	g.input = &mockInput{}

	// Move close and trigger pickup
	g.UpdatePickups()
	
	if len(char.Inventory) != 1 {
		t.Errorf("Expected 1 item in inventory, got %d", len(char.Inventory))
	}
	if len(g.World.Items) != 0 {
		t.Errorf("Expected 0 items in world after pickup, got %d", len(g.World.Items))
	}
	
	// 2. Test Weight Limit
	it2 := NewItemInstance("heavy", &ObjectConfig{Name: "Heavy", Weight: 100.0}, 0.1, 0.1)
	g.World.Items = []*ItemInstance{it2}
	g.UpdatePickups()
	if len(char.Inventory) != 1 {
		t.Errorf("Should not have picked up heavy item")
	}
	
	// 3. Test Drop
	success := g.TryDrop(&char.Actor, 0)
	if !success {
		t.Fatal("Failed to drop item")
	}
	if len(char.Inventory) != 0 {
		t.Errorf("Inventory should be empty after drop")
	}
	if len(g.World.Items) != 2 { // Test sword + heavy item that was never picked up
		t.Errorf("World should have 2 items now, got %d", len(g.World.Items))
	}

	// 4. Test Drop All Items on Death
	// Pick up the sword again. Use UpdatePickups to properly simulate removal from world.
	g.UpdatePickups()
	
	if len(char.Inventory) != 1 {
		t.Fatalf("Failed to re-pickup sword, inventory size: %d", len(char.Inventory))
	}
	
	ctx := &SystemContext{
		World: g.World,
		Registries: g.Registries,
	}
	
	// Kill character
	char.TakeDamage(1000, nil, ctx)
	
	if len(char.Inventory) != 0 {
		t.Errorf("Expected inventory to be empty after death, got %d", len(char.Inventory))
	}
	// The sword should be back on the ground, so items should be at least 2 again (sword + heavy)
	if len(g.World.Items) != 2 {
		t.Errorf("Expected 2 items in world after death drop, got %d", len(g.World.Items))
	}
}
