package game

import (
	"oinakos/internal/engine"
	"testing"
)


func TestTryPickup_Success(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter
	
	item := &ItemInstance{
		Config: &ObjectConfig{
			Name:   "Test Item",
			Weight: 5,
		},
		Pickable: true,
	}
	
	success := g.TryPickup(&pc.Actor, item)
	
	if !success {
		t.Error("expected TryPickup to succeed")
	}
	if len(pc.Inventory) != 1 {
		t.Errorf("expected 1 item in inventory, got %d", len(pc.Inventory))
	}
	if item.Pickable {
		t.Error("item should not be pickable after pickup")
	}
}

func TestTryPickup_FullInventory(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter
	pc.Config.MaxItems = 1
	
	item1 := &ItemInstance{Config: &ObjectConfig{Name: "Item 1", Weight: 1}, Pickable: true}
	item2 := &ItemInstance{Config: &ObjectConfig{Name: "Item 2", Weight: 1}, Pickable: true}
	
	g.TryPickup(&pc.Actor, item1)
	success := g.TryPickup(&pc.Actor, item2)
	
	if success {
		t.Error("expected TryPickup to fail due to full inventory")
	}
	if len(pc.Inventory) != 1 {
		t.Errorf("expected 1 item in inventory, got %d", len(pc.Inventory))
	}
}

func TestTryPickup_TooHeavy(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter
	pc.MaxWeight = 10
	
	item := &ItemInstance{
		Config: &ObjectConfig{
			Name:   "Heavy Item",
			Weight: 20,
		},
		Pickable: true,
	}
	
	success := g.TryPickup(&pc.Actor, item)
	
	if success {
		t.Error("expected TryPickup to fail due to weight")
	}
}

func TestTryDrop(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter
	
	item := &ItemInstance{
		Config: &ObjectConfig{
			Name: "Test Item",
		},
		Pickable: false,
	}
	pc.Inventory = append(pc.Inventory, item)
	
	success := g.TryDrop(&pc.Actor, 0)
	
	if !success {
		t.Error("expected TryDrop to succeed")
	}
	if len(pc.Inventory) != 0 {
		t.Errorf("expected empty inventory, got %d items", len(pc.Inventory))
	}
	if len(g.World.Items) != 1 {
		t.Errorf("expected 1 item in world, got %d", len(g.World.Items))
	}
	if !g.World.Items[0].Pickable {
		t.Error("dropped item should be pickable")
	}
}


type BetterMockInput struct {
	GenericMockInput
	JustPressedKeys []engine.Key
}

func (m *BetterMockInput) IsKeyJustPressed(k engine.Key) bool {
	for _, pk := range m.JustPressedKeys {
		if pk == k { return true }
	}
	return false
}

func TestUpdatePickups_Success(t *testing.T) {
	g := setupTestGame()
	g.playableCharacter.X, g.playableCharacter.Y = 0, 0
	
	mockInput := &BetterMockInput{}
	mockInput.JustPressedKeys = []engine.Key{engine.KeySpace}
	g.input = mockInput
	
	item := &ItemInstance{
		Config: &ObjectConfig{
			Name:   "Test Item",
			Weight: 1,
		},
		X:        1, // Within range 2.5
		Y:        1,
		Pickable: true,
	}
	g.World.Items = append(g.World.Items, item)
	
	g.UpdatePickups()
	
	if len(g.World.Items) != 0 {
		t.Error("item should have been picked up and removed from world")
	}
	if len(g.playableCharacter.Inventory) != 1 {
		t.Error("item should be in inventory")
	}
}

func TestDropAllItems(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter
	
	item1 := &ItemInstance{Config: &ObjectConfig{Name: "Item 1", Weight: 1}, Pickable: false}
	item2 := &ItemInstance{Config: &ObjectConfig{Name: "Item 2", Weight: 1}, Pickable: false}
	
	pc.Inventory = append(pc.Inventory, item1)
	pc.Slots = make(map[string]*ItemInstance)
	pc.Slots["head"] = item2
	
	g.DropAllItems(&pc.Actor)
	
	if len(pc.Inventory) != 0 {
		t.Error("Inventory should be empty")
	}
	if pc.Slots["head"] != nil {
		t.Error("Head slot should be empty")
	}
	if len(g.World.Items) != 2 {
		t.Errorf("World should have 2 items, got %d", len(g.World.Items))
	}
}
func TestDropEquippedItem(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter
	item := &ItemInstance{Config: &ObjectConfig{ID: "sword", Name: "Sword", Weight: 2}}
	pc.Slots = map[string]*ItemInstance{"weapon": item}
	
	g.DropEquippedItem(&pc.Actor, "weapon")
	
	if pc.Slots["weapon"] != nil {
		t.Error("Expected weapon slot to be empty after dropping")
	}
	if len(g.World.Items) != 1 {
		t.Error("Expected item to be in world after dropping")
	}
}

func TestDropSpecificItem(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter
	item := &ItemInstance{Config: &ObjectConfig{ID: "gold", Name: "Gold", Weight: 1}}
	pc.Inventory = []*ItemInstance{item}
	
	g.DropSpecificItem(&pc.Actor, item)
	
	if len(pc.Inventory) != 0 {
		t.Error("Expected inventory to be empty after dropping specific item")
	}
	if len(g.World.Items) != 1 {
		t.Error("Expected item to be in world after dropping")
	}
}
