package game

import (
	"testing"
)

func TestActor_WorkshopCoverage(t *testing.T) {
	ctx := NewTestContext()
	p := NewCharacter(0, 0, nil, 1, true, nil)
	p.Name = "Artisan"
	p.PrimaryAttributes.Intellect = 50
	
	// Mock registries
	if ctx.Registries.Objects == nil {
		ctx.Registries.Objects = NewObjectRegistry()
	}
	ctx.Registries.Objects.Objects["meat"] = &ObjectConfig{Name: "Cooked Meat"}
	ctx.Registries.Objects.Objects["iron_ingot"] = &ObjectConfig{Name: "Iron Ingot"}
	ctx.Registries.Objects.Objects["iron_sword"] = &ObjectConfig{Name: "Iron Sword"}
	
	// 1. ProcessCooking
	fire := &Obstacle{ID: "campfire_01", X: 1, Y: 1, Alive: true}
	ctx.World.Obstacles = append(ctx.World.Obstacles, fire)
	
	p.ActionState = ActorCooking
	p.Inventory = append(p.Inventory, &ItemInstance{Config: &ObjectConfig{ID: "raw_meat", Name: "Raw Meat"}})
	p.X, p.Y = 1, 1
	p.Tick = 301
	
	p.ProcessCooking(ctx)
	if p.ActionState != ActorIdle { t.Error("Should finish cooking") }

	// 2. ProcessWorkshop - Smelting
	furnace := &Obstacle{ID: "furnace_01", X: 5, Y: 5, Alive: true}
	ctx.World.Obstacles = append(ctx.World.Obstacles, furnace)
	
	p.ActionState = ActorWorkshop
	p.Inventory = append(p.Inventory, &ItemInstance{Config: &ObjectConfig{ID: "iron_ore", Name: "Iron Ore"}})
	p.X, p.Y = 5, 5
	p.Tick = 481
	
	p.ProcessWorkshop(ctx)
	if p.ActionState != ActorIdle { t.Error("Should finish workshop action") }

	// 3. ProcessWorkshop - Repairing
	p.ActionState = ActorWorkshop
	item := NewItemInstance("sword", &ObjectConfig{ID: "sword", Name: "Sword", Resistance: 100}, 0, 0)
	item.Resistance = 10
	p.Slots["weapon"] = item
	p.Tick = 481
	
	p.ProcessWorkshop(ctx)

	// 4. ProcessWorkshop - Crafting
	p.ActionState = ActorWorkshop
	p.Inventory = append(p.Inventory, &ItemInstance{Config: &ObjectConfig{ID: "wood", Name: "Wood"}})
	p.Tick = 481
	
	p.ProcessWorkshop(ctx)
}
