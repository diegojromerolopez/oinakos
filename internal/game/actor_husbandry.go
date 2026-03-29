package game

import (
	"fmt"
	"image/color"
	"math"
)

func (a *Actor) updateHusbandry(ctx *SystemContext) {
	if !a.IsAlive() { return }

	// Ticking down cooldowns for all animals (milkable or not)
	if a.MilkCooldownTicks > 0 {
		a.MilkCooldownTicks--
	}

	if a.ActionState != ActorMilking { return }

	// 1. Locate the target animal
	var targetAnimal *Actor
	for _, c := range ctx.World.Characters {
		if c.Name == a.TargetActorID {
			targetAnimal = &c.Actor
			break
		}
	}

	if targetAnimal == nil {
		a.ActionState = ActorIdle
		a.LastAIReasoning = "Animal missing!"
		return
	}

	// 2. Proximity check
	dist := math.Sqrt(math.Pow(a.X-targetAnimal.X, 2) + math.Pow(a.Y-targetAnimal.Y, 2))
	if dist > 2.0 {
		a.ActionState = ActorIdle
		a.LastAIReasoning = "Too far to milk!"
		return
	}

	// 3. Animal state check
	if !targetAnimal.IsAlive() || !targetAnimal.Config.Stats.IsMilkable {
		a.ActionState = ActorIdle
		a.LastAIReasoning = "Cannot milk this!"
		return
	}

	if targetAnimal.MilkCooldownTicks > 0 {
		a.ActionState = ActorIdle
		a.LastAIReasoning = "Already milked recently."
		return
	}

	// 4. Progress milking
	a.WorkTicks++
	
	// Feedback every 1s
	if a.WorkTicks%60 == 0 {
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: "milking...", X: a.X, Y: a.Y, Life: 45, Color: ColorHeal,
		})
	}

	// 5. Success check at completion
	if a.WorkTicks >= 300 { // 5s to milk
		a.WorkTicks = 0
		a.ActionState = ActorIdle
		
		if !a.CheckAbilitySuccess("milk", 0) {
			a.LastAIReasoning = "Fumbled the milking!"
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Fumble!", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm })
			return
		}

		// Set cooldown on animal
		targetAnimal.MilkCooldownTicks = targetAnimal.RawStats.MilkCooldown
		
		// Reward: Yield scales with Husbandry
		litres := a.GetAbilityYield("milk")
		
		milkConfig := ctx.Registries.Objects.Get("bucket_milk")
		if milkConfig != nil {
			inst := &ItemInstance{Config: milkConfig}
			// In a more complex system, we'd store the 'litres' on the item
			a.Inventory = append(a.Inventory, inst)
			
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: fmt.Sprintf("%.1fL Milk!", litres), X: a.X, Y: a.Y, Life: 90, Color: ColorHeal,
			})
			a.LastAIReasoning = fmt.Sprintf("Gathered %.1fL milk from %s", litres, targetAnimal.Name)
		}
	}
}

func (a *Actor) updateOwnership(ctx *SystemContext) {
	if !a.IsAlive() { return }

	if a.ActionState != ActorStashing { return }

	// Stashing logic: Move items from inventory to the owned chest
	if a.OwnedChestID == "" {
		a.ActionState = ActorIdle
		return
	}

	// Find the chest obstacle
	var chest *Obstacle
	for _, o := range ctx.World.Obstacles {
		if o.ID == a.OwnedChestID {
			chest = o
			break
		}
	}

	if chest == nil || !chest.Alive {
		a.ActionState = ActorIdle
		a.OwnedChestID = "" // Clear invalid ownership
		return
	}

	// Proximity
	dist := math.Sqrt(math.Pow(a.X-chest.X, 2) + math.Pow(a.Y-chest.Y, 2))
	if dist > 1.5 {
		a.ActionState = ActorIdle
		return
	}

	// Move items
	if len(a.Inventory) > 0 {
		if !a.CheckAbilitySuccess("stash", 0) {
			a.ActionState = ActorIdle
			return
		}

		item := a.Inventory[0]
		a.Inventory = a.Inventory[1:]
		
		// Obstacles don't have inventories yet, so we just "store" them into the generic chest state
		// but we track the weight if we want to be realistic later.
		chest.TotalWeight += item.Config.Weight
		
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: "Stashed " + item.Config.Name, X: a.X, Y: a.Y, Life: 60, Color: color.RGBA{180, 180, 255, 255},
		})
	}

	a.ActionState = ActorIdle
}
