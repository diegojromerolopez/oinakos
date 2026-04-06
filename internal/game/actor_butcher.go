package game

import (
	"fmt"
)

// Butcher extracts raw_meat from a dead animal character.
func (c *Character) Butcher(ctx *SystemContext, target *Character) {
	if target == nil || target.IsAlive() || target.MeatQuantity <= 0 { return }
	
	// Butcher yield (1-2 units per action)
	qty := 1.0 + float64(int(c.GetAbilityYield("survivalism") * 0.02))
	if qty > target.MeatQuantity { qty = target.MeatQuantity }

	_, meatObj := ctx.Registries.Objects.RandomVariantID("raw_meat")
	if meatObj == nil { meatObj = ctx.Registries.Objects.Objects["meat"] }

	if meatObj != nil {
		it := NewItemInstance(meatObj.ID, meatObj, c.X, c.Y)
		c.Inventory = append(c.Inventory, it)
		target.MeatQuantity -= qty
		
		if ctx.Log != nil { 
			ctx.Log(fmt.Sprintf("%s butchered %s (Remaining: %.1f).", c.Name, target.Name, target.MeatQuantity), LogNPC) 
		}
	}

	if target.MeatQuantity <= 0 {
		// Mark as fully butchered
		target.Name = "Butchered " + target.Name
		target.RotTicks = TicksPerMonth // Mark for miasma/decay 
	}
}
