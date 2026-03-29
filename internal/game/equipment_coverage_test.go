package game

import (
	"testing"
)

func TestActor_EquipmentCoverage(t *testing.T) {
	ctx := NewTestContext()
	// Initialize Game inside ctx if missing
	if ctx.World.Game == nil {
		ctx.World.Game = &Game{
			World:      ctx.World,
			Registries: ctx.Registries,
		}
	}
	g := ctx.World.Game
	p := NewCharacter(0, 0, nil, 1, true, nil)
	g.playableCharacter = p
	
	// Mock items and registry
	if ctx.Registries.Objects == nil {
		ctx.Registries.Objects = NewObjectRegistry()
	}
	ctx.Registries.Objects.Objects["sword"] = &ObjectConfig{
		ID:   "sword",
		Name: "Sword",
		Type: "weapon",
		Slot: "weapon",
		Combat: &Weapon{
			Damage: Damage{Min: 5, Max: 10},
		},
	}
	ctx.Registries.Objects.Objects["armor"] = &ObjectConfig{
		ID:         "armor",
		Name:       "Armor",
		Type:       "armor",
		Slot:       "torso",
		Resistance: 50,
	}
	
	item := NewItemInstance("sword", ctx.Registries.Objects.Get("sword"), 0, 0)
	ctx.World.Items = append(ctx.World.Items, item)
	item.Pickable = true
	
	// 1. TryPickup
	if !g.TryPickup(&p.Actor, item) {
		t.Error("TryPickup failed although inventory should be empty")
	}
	found := false
	for _, it := range p.Inventory { if it == item { found = true } }
	if !found { t.Error("Failed to pickup item") }

	// 2. EquipItem
	if !p.EquipItem(item) {
		t.Error("EquipItem failed")
	}
	if p.Slots["weapon"] != item { t.Error("Failed to equip item") }
	if p.Weapon == nil { t.Error("Expected weapon struct to be initialized") }

	// 3. TryDrop
	p.Inventory = append(p.Inventory, item)
	if !g.TryDrop(&p.Actor, 0) {
		t.Error("TryDrop failed")
	}
	foundInWorld := false
	for _, it := range ctx.World.Items { if it == item { foundInWorld = true } }
	if !foundInWorld { t.Error("Item should be in world after drop") }

	// 4. LoadEquipment
	p.Config = &EntityConfig{
		Equipment: map[string]string{"torso": "armor"},
		Inventory: []string{"sword"},
	}
	p.LoadEquipment(ctx.Registries.Objects)
	if p.Slots["torso"] == nil { t.Error("Failed to load equipment") }
	if len(p.Inventory) == 0 { t.Error("Failed to load inventory") }
	
	// 5. EvaluateUpgrade
	betterSword := &ItemInstance{
		Config: &ObjectConfig{
			Slot: "weapon",
			Type: "weapon",
			Combat: &Weapon{Damage: Damage{Min: 20, Max: 30}},
		},
	}
	if !p.EvaluateUpgrade(betterSword) {
		t.Error("Expected better sword to be an upgrade")
	}

	// 6. CanCarry
	p.MaxWeight = 100
	if !p.CanCarry(50) { t.Error("Should be able to carry 50") }
	if p.CanCarry(101) { t.Error("Should not be able to carry 101") }

	// 7. ConsumeItem
	food := &ItemInstance{
		Config: &ObjectConfig{
			Name: "Bread",
			Hunger: 20,
			Effects: map[string]StatEffect{
				"health": {Increase: 10},
				"attack": {Increase: 5},
			},
		},
	}
	p.State.Hunger = 50
	p.State.HealthPoints = 50
	oldAttack := p.BaseAttack
	p.ConsumeItem(food, ctx)
	if p.State.Hunger >= 50 { t.Error("Hunger should decrease") }
	if p.BaseAttack <= oldAttack { t.Error("Permanent attack effect should apply") }
}
