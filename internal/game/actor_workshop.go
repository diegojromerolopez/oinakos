package game

import (
	"fmt"
	"math"
	"strings"
)

// ProcessCooking handles the multi-tick cycle for turning raw food into cooked meals.
func (c *Character) ProcessCooking(ctx *SystemContext) {
	if c.ActionState != ActorCooking { return }

	// 1. Requirement: Near a heat source (Campfire)
	nearFire := false
	for _, o := range ctx.World.Obstacles {
		if o.Alive && o.Archetype != nil && strings.Contains(strings.ToLower(o.Archetype.ID), "campfire") {
			dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
			if dist < 2.5 { nearFire = true; break }
		}
	}
	if !nearFire { c.ActionState = ActorIdle; return }

	// 2. Recipe Detection (Early check to see if pot is needed)
	foundMeat, foundVeg := -1, -1
	for i, it := range c.Inventory {
		if it == nil || it.Config == nil { continue }
		if it.Config.ID == "raw_meat" && foundMeat == -1 { foundMeat = i }
		if (it.Config.ID == "cabbage" || it.Config.ID == "turnip" || it.Config.ID == "apple" || it.Config.ID == "pear") && foundVeg == -1 { foundVeg = i }
	}

	// 3. Requirement: Utensils (Cooking Pot) - ONLY needed for stew (vegetable recipes)
	hasPot := false
	for _, it := range c.Inventory { if it != nil && it.Config != nil && it.Config.ID == "cooking_pot" { hasPot = true; break } }
	
	if foundVeg >= 0 && !hasPot {
		c.ActionState, c.LastAIReasoning = ActorIdle, "Need a cooking pot for stew!"
		return 
	}
	if foundMeat == -1 && foundVeg == -1 {
		c.ActionState, c.LastAIReasoning = ActorIdle, "Nothing to cook!"
		return
	}

	if c.Tick < 400 { return } // 6.6 seconds (400 ticks)

	// 3. Success Check
	if !c.CheckAbilitySuccess("cook", 0) {
		c.ActionState, c.Tick, c.LastAIReasoning = ActorIdle, 0, "Burned the meal!"
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "🔥 BURNT!", X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHarm })
		return
	}

	// 4. Recipe Processing
	if foundMeat >= 0 {
		c.ActionState, c.Tick = ActorIdle, 0
		cookedID := "meat" // Default cooked meat
		msg := "prepared some Cooked Meat!"
		floatMsg := "🍖 MEAT!"
		if foundVeg >= 0 { 
			cookedID = "stew" 
			msg = "prepared a Hearty Stew!"
			floatMsg = "🍜 STEW!"
		}
		
		cookedConfig := ctx.Registries.Objects.Objects[cookedID]
		if cookedConfig != nil {
			yield := c.GetAbilityYield("cook")
			
			// Remove ingredients (removing higher index first)
			if foundVeg >= 0 {
				if foundMeat > foundVeg {
					c.Inventory = append(c.Inventory[:foundMeat], c.Inventory[foundMeat+1:]...)
					c.Inventory = append(c.Inventory[:foundVeg], c.Inventory[foundVeg+1:]...)
				} else {
					c.Inventory = append(c.Inventory[:foundVeg], c.Inventory[foundVeg+1:]...)
					c.Inventory = append(c.Inventory[:foundMeat], c.Inventory[foundMeat+1:]...)
				}
			} else {
				c.Inventory = append(c.Inventory[:foundMeat], c.Inventory[foundMeat+1:]...)
			}

			it := NewItemInstance(cookedID, cookedConfig, c.X, c.Y)
			it.Resistance = int(yield * 0.2) // Quality bonus
			c.Inventory = append(c.Inventory, it)
			
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s %s", c.Name, msg), LogNPC) }
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: floatMsg, X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHeal })
		}
	} else {
		c.ActionState, c.Tick, c.LastAIReasoning = ActorIdle, 0, "Missing meat for cooking!"
	}
}

// ProcessWorkshop handles repairing gear, smelting, and crafting.
func (c *Character) ProcessWorkshop(ctx *SystemContext) {
	if c.ActionState != ActorWorkshop { return }

	// 1. Find nearest station
	var station *Obstacle
	for _, o := range ctx.World.Obstacles {
		if o.Archetype == nil { continue }
		id := strings.ToLower(o.Archetype.ID)
		if o.Alive && (strings.Contains(id, "bench") || strings.Contains(id, "workshop") || strings.Contains(id, "anvil") || strings.Contains(id, "furnace")) {
			dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
			if dist < 2.5 { station = o; break }
		}
	}
	if station == nil { c.ActionState = ActorIdle; return }

	stationID := strings.ToLower(station.Archetype.ID)
	isFurnace := strings.Contains(stationID, "furnace") || strings.Contains(stationID, "anvil")
	isBench := strings.Contains(stationID, "bench") || strings.Contains(stationID, "workshop")

	if c.Tick < 480 { return }

	c.ActionState, c.Tick = ActorIdle, 0

	// 1. Try Smelting (if furnace and has ore)
	if isFurnace {
		foundOre := -1
		for i, it := range c.Inventory {
			if it != nil && it.Config != nil && it.Config.ID == "iron_ore" { foundOre = i; break }
		}
		if foundOre >= 0 && c.CheckAbilitySuccess("smelt", 0) {
			yield := c.GetAbilityYield("smelt")
			c.Inventory = append(c.Inventory[:foundOre], c.Inventory[foundOre+1:]...)
			ingotID := "iron_ingot"
			ingotConfig := ctx.Registries.Objects.Objects[ingotID]
			if ingotConfig != nil {
				c.Inventory = append(c.Inventory, NewItemInstance(ingotID, ingotConfig, c.X, c.Y))
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s smelted high-purity iron (%.1f).", c.Name, yield), LogNPC) }
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "🔥 SMELTED", X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHeal })
				return
			}
		}
	}

	// 2. Try Repairing (if bench/anvil)
	repairedAnything := false
	if isBench || isFurnace {
		for _, it := range c.Slots {
			if it != nil && it.Config != nil && it.Resistance < it.Config.Resistance {
				if c.CheckAbilitySuccess("repair", 0) {
					it.Resistance = it.Config.Resistance
					repairedAnything = true
				}
			}
		}
	}

	if repairedAnything {
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s repaired their equipment.", c.Name), LogNPC) }
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "⚒ REPAIRED", X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHeal })
		c.ModifySentiment("self", 2.0)
		return
	}

	// 3. Try Crafting (if bench and has wood)
	if isBench {
		foundWood := -1
		for i, it := range c.Inventory {
			if it != nil && it.Config != nil && (it.Config.ID == "wood" || it.Config.ID == "timber") { foundWood = i; break }
		}
		if foundWood >= 0 && c.CheckAbilitySuccess("craft", 0) {
			yield := c.GetAbilityYield("craft")
			c.Inventory = append(c.Inventory[:foundWood], c.Inventory[foundWood+1:]...)
			toolID := "iron_sword" // Simplified: wood -> sword if you are an artisan
			if yield < 50 { toolID = "dagger" }
			toolConfig := ctx.Registries.Objects.Objects[toolID]
			if toolConfig != nil {
				c.Inventory = append(c.Inventory, NewItemInstance(toolID, toolConfig, c.X, c.Y))
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s crafted a %s (Quality: %.1f).", c.Name, toolConfig.Name, yield), LogNPC) }
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "🔨 CRAFTED", X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHeal })
			}
		}
	}
}
