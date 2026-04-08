package game

import (
	"testing"
)

func TestPickup_DropAdvanced(t *testing.T) {
	ctx := setupTestGame()
	mc := ctx.playableCharacter
	g := ctx // setupTestGame returns *Game
	
	// 1. Drop Equipped Item
	item := NewItemInstance("sword_iron", &ObjectConfig{ID: "sword_iron"}, 0, 0)
	mc.Slots["weapon"] = item
	success := g.DropEquippedItem(&mc.Actor, "weapon")
	if !success { t.Error("DropEquippedItem returned false") }
	if mc.Slots["weapon"] != nil { t.Error("Failed to drop equipped item") }
	if len(g.World.Items) != 1 { t.Error("Item did not appear in world") }
	
	// 2. Drop Specific Item from inventory
	item2 := NewItemInstance("gold_ore", &ObjectConfig{ID: "gold_ore"}, 0, 0)
	mc.Inventory = []*ItemInstance{item2}
	success = g.DropSpecificItem(&mc.Actor, item2)
	if !success { t.Error("DropSpecificItem returned false") }
	if len(mc.Inventory) != 0 { t.Error("Failed to drop specific item") }
	if len(g.World.Items) != 2 { t.Error("Item2 did not appear in world") }
	
	// 3. Drop Item not in inventory (force drop behavior)
	item3 := NewItemInstance("stone", &ObjectConfig{ID: "stone", Name: "Stone"}, 0, 0)
	success = g.DropSpecificItem(&mc.Actor, item3)
	if !success { t.Error("DropSpecificItem should return true even if not in inventory (appends and drops)") }
}
