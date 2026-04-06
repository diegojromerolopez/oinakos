package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func (c *Character) handleSurvivalNeeds(ctx *SystemContext) bool {
	if c.ActionState == ActorDrinking || c.ActionState == ActorEating || c.ActionState == ActorResting || c.ActionState == ActorBathing || c.ActionState == ActorRelieving { 
		if c.Tick < 60 { return true } 
	}

	isStarving := c.State.Hunger > 90
	isDehydrated := c.State.Thirst > 90
	isHungry := c.State.Hunger > 70
	isThirsty := c.State.Thirst > 30 // Proactive threshold (lowered from 45)
	isUrgentThirst := c.State.Thirst > 60 // Urgent survival threshold
	isExhausted := c.State.Fatigue > 85
	isBursting := c.State.BladderLevel > 85 || c.State.BowelLevel > 85
	isDirty := c.State.Hygiene < 30
	isLowHealth := c.State.HealthPoints < c.State.MaxHealthPoints / 3
	inPain := c.State.Pain > 30
	searchRadius := 200.0 + (float64(c.PrimaryAttributes.Intellect) * 0.1)

	if inPain && c.ActionState == ActorIdle {
		var pSource *Character; minPDist := 10.0
		for _, other := range ctx.World.Characters {
			if other != c && other.IsAlive() && other.TargetActorID == c.Name {
				if d := math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)); d < minPDist { minPDist, pSource = d, other }
			}
		}
		if pSource != nil {
			if c.BaseAttack > 0 || c.Weapon != nil {
				c.TargetActor = &pSource.Actor
				c.ActionState = ActorAttacking
				return true
			} else {
				c.MoveTo(ctx, c.X + (c.X - pSource.X)*2, c.Y + (c.Y - pSource.Y)*2)
				return true
			}
		}
	}

	// 1. Immediate Health/Danger
	if isLowHealth && c.ActionState == ActorIdle {
		var nHealer *Character; minHDist := 200.0
		for _, other := range ctx.World.Characters {
			if other.IsAlive() && other != c && other.Config != nil {
				id := strings.ToLower(other.Config.ID)
				if strings.Contains(id, "doctor") || strings.Contains(id, "shaman") || strings.Contains(id, "cleric") || strings.Contains(id, "healer") {
					if d := math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)); d < minHDist { minHDist, nHealer = d, other }
				}
			}
		}
		if nHealer != nil {
			if minHDist > 2.0 { c.MoveTo(ctx, nHealer.X, nHealer.Y); return true }
			c.ActionState = ActorResting; return true
		}
	}

	// 2. Metabolic Consumption (Inventory first)
	for i, item := range c.Inventory {
		if item == nil || item.Config == nil { continue }
		
		// Handle Canteen/Refillable: Only use if AWAY from a local source to conserve water
		if (item.LiquidContent > 0 || item.Refillable) && isThirsty {
			if item.LiquidContent >= 0.25 { // At least one gulp
				// Find nearest well - don't waste bottle if one is close
				nearWell := false
				for _, o := range ctx.World.Obstacles {
					id, archID := strings.ToLower(o.ID), ""
					if o.Archetype != nil { archID = strings.ToLower(o.Archetype.ID) }
					if strings.Contains(id, "well") || strings.Contains(archID, "well") {
						if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < 30.0 { nearWell = true; break }
					}
				}
				if !nearWell || isUrgentThirst {
					c.State.Thirst -= 12.5 // Exactly one gulp's worth
					item.LiquidContent -= 0.25
					if item.LiquidContent < 0 { item.LiquidContent = 0 }
					c.ActionState, c.Tick = ActorDrinking, 0
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s drank from %s (Remaining: %.2fL).", c.Name, item.Config.Name, item.LiquidContent), LogNPC) }
					return true
				}
			}
		}

		hasNutrients := item.Config.Consumable || item.Config.Hunger > 0 || item.Config.Thirst > 0 || item.Config.Fatigue > 0 || item.Config.Energy > 0
		if !hasNutrients && item.Config.Effects != nil {
			if _, ok := item.Config.Effects["hunger"]; ok { hasNutrients = true }
			if _, ok := item.Config.Effects["thirst"]; ok { hasNutrients = true }
			if _, ok := item.Config.Effects["fatigue"]; ok { hasNutrients = true }
		}
		if hasNutrients {
			providesHunger := item.Config.Hunger > 0 || (item.Config.Effects != nil && item.Config.Effects["hunger"].Increase > 0)
			providesThirst := item.Config.Thirst > 0 || (item.Config.Effects != nil && item.Config.Effects["thirst"].Increase > 0)
			if isStarving || isDehydrated || (isHungry && providesHunger) || (isThirsty && providesThirst) {
				if c.ConsumeItem(item, ctx) {
					c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
					c.ActionState, c.Tick = ActorEating, 0
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s consumed %s to satisfy needs.", c.Name, item.Config.ID), LogNPC) }
					return true
				}
			}
		}
	}

	// 3. Hydration (Seeking Source)
	if isThirsty || isDehydrated {
		var nDrinkSrc *Obstacle; minWDist := 400.0
		for _, o := range ctx.World.Obstacles {
			if !o.Alive { continue }
			id, archID := strings.ToLower(o.ID), ""
			if o.Archetype != nil { archID = strings.ToLower(o.Archetype.ID) }
			if strings.Contains(id, "well") || strings.Contains(archID, "well") || strings.Contains(id, "river") || strings.Contains(archID, "river") {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minWDist { minWDist, nDrinkSrc = d, o }
			}
		}
		if nDrinkSrc != nil {
			if minWDist < 5.0 { // Relaxed for large well footprints
				c.ActionState, c.Tick = ActorDrinking, 0
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s is drinking from %s.", c.Name, nDrinkSrc.ID), LogNPC) }
				return true 
			}
			c.MoveTo(ctx, nDrinkSrc.X, nDrinkSrc.Y); return true
		}
		// Also seek Tavern separately (much larger footprint)
		var nTavern *Obstacle; minTDist := 400.0
		for _, o := range ctx.World.Obstacles {
			if !o.Alive { continue }
			id, archID := strings.ToLower(o.ID), ""
			if o.Archetype != nil { archID = strings.ToLower(o.Archetype.ID) }
			if strings.Contains(id, "tavern") || strings.Contains(archID, "tavern") {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minTDist { minTDist, nTavern = d, o }
			}
		}
		if nTavern != nil {
			if minTDist < 12.0 { 
				c.ActionState, c.Tick = ActorDrinking, 0
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s is drinking in %s.", c.Name, nTavern.ID), LogNPC) }
				return true
			}
			c.MoveTo(ctx, nTavern.X, nTavern.Y); return true
		}
		// Also seek River as FloorZone
		var nRiver *FloorZone; minRDist := 400.0
		if ctx.World != nil && ctx.World.CurrentMapType != nil {
			for _, fz := range ctx.World.CurrentMapType.FloorZones {
				if fz.Type == "river" || strings.Contains(strings.ToLower(fz.Name), "river") {
					d := fz.DistanceTo(c.X, c.Y)
					if d < minRDist { minRDist, nRiver = d, fz }
				}
			}
		}
		if nRiver != nil {
			if minRDist < 2.0 { c.ActionState, c.Tick = ActorDrinking, 0; return true }
			c.MoveTo(ctx, (nRiver.MinX+nRiver.MaxX)*0.5, (nRiver.MinY+nRiver.MaxY)*0.5)
			return true
		}
		if c.CurrentTile == "water.png" || strings.Contains(c.CurrentTile, "water") {
			c.ActionState, c.Tick = ActorDrinking, 0; return true
		}
	}

	// 4. Nutrition (Seeking Food/Cooking/Foraging)
	if isHungry || isStarving {
		// Try Cooking first
		hasIngredients, hasPot := false, false
		for _, it := range c.Inventory {
			if it == nil || it.Config == nil { continue }
			if it.Config.ID == "raw_meat" || strings.HasPrefix(it.Config.ID, "raw_meat_") || it.Config.ID == "cabbage" { hasIngredients = true }
			if it.Config.ID == "cooking_pot" { hasPot = true }
		}
		if hasIngredients && hasPot {
			var nFire *Obstacle; minFDist := 50.0
			for _, o := range ctx.World.Obstacles {
				if o.Alive && o.Archetype != nil && strings.Contains(strings.ToLower(o.Archetype.ID), "campfire") {
					if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minFDist { minFDist, nFire = d, o }
				}
			}
			if nFire != nil {
				if minFDist < 2.0 { c.ActionState, c.Tick = ActorCooking, 0; return true }
				c.MoveTo(ctx, nFire.X, nFire.Y); return true
			}
		}
		// Foraging fallback
		var nForage *Obstacle; minFDist := 200.0
		for _, o := range ctx.World.Obstacles {
			if !o.Alive { continue }
			id, archID := strings.ToLower(o.ID), ""
			if o.Archetype != nil { archID = strings.ToLower(o.Archetype.ID) }
			if strings.Contains(id, "bush") || strings.Contains(archID, "bush") || (o.Archetype != nil && o.Archetype.IsCrop) {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minFDist { minFDist, nForage = d, o }
			}
		}
		if nForage != nil {
			if minFDist < 2.0 { c.ActionState, c.Tick = ActorForaging, 0; return true }
			c.MoveTo(ctx, nForage.X, nForage.Y); return true
		}
	}

	// 5. Relief Needs
	if isBursting && c.ActionState == ActorIdle {
		var nToilet *Obstacle; minTDist := 50.0
		for _, o := range ctx.World.Obstacles {
			if !o.Alive { continue }
			id, archID := strings.ToLower(o.ID), ""
			if o.Archetype != nil { archID = strings.ToLower(o.Archetype.ID) }
			if strings.Contains(id, "toilet") || strings.Contains(archID, "toilet") || strings.Contains(id, "latrine") {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minTDist { minTDist, nToilet = d, o }
			}
		}
		if nToilet != nil {
			if minTDist > 2.0 { c.MoveTo(ctx, nToilet.X, nToilet.Y); return true }
		}
		c.ActionState, c.Tick = ActorRelieving, 0; return true
	}

	// 6. Rest
	if isExhausted && c.ActionState == ActorIdle {
		var nRest *Obstacle; minDist := 50.0
		for _, o := range ctx.World.Obstacles {
			id := strings.ToLower(o.ID)
			if o.Alive && (strings.Contains(id, "campfire") || strings.Contains(id, "bed")) {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minDist { minDist, nRest = d, o }
			}
		}
		if nRest != nil && minDist > 3.0 { c.MoveTo(ctx, nRest.X, nRest.Y); return true }
		c.ActionState, c.Tick = ActorResting, 0; return true
	}

	// 7. Hygiene
	if isDirty && c.ActionState == ActorIdle && !isStarving && !isDehydrated {
		var nWater *Obstacle; minWDist := 100.0
		for _, o := range ctx.World.Obstacles {
			if !o.Alive { continue }
			id := strings.ToLower(o.ID)
			if strings.Contains(id, "well") || strings.Contains(id, "bath") {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minWDist { minWDist, nWater = d, o }
			}
		}
		if nWater != nil {
			if minWDist > 2.0 { c.MoveTo(ctx, nWater.X, nWater.Y); return true }
			c.ActionState = ActorBathing; return true
		}
	}
	if isHungry && c.ActionState == ActorIdle {
		var nForage *Obstacle; minFDist := searchRadius
		for _, o := range ctx.World.Obstacles {
			if !o.Alive { continue }
			id := strings.ToLower(o.ID)
			archID := ""
			if o.Archetype != nil { archID = strings.ToLower(o.Archetype.ID) }
			if (strings.Contains(id, "tree") || strings.Contains(id, "bush") || strings.Contains(archID, "tree") || strings.Contains(archID, "bush") || (o.Archetype != nil && o.Archetype.IsCrop)) && o.CooldownTicks <= 0 {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minFDist { minFDist, nForage = d, o }
			}
		}
		if nForage != nil {
			if minFDist < 1.5 { c.ActionState, c.Tick = ActorForaging, 0; return true } else { c.MoveTo(ctx, nForage.X, nForage.Y); return true }
		}

		// 8a. Seek Corpse (Butcher)
		var nCorpse *Character; minCDist := searchRadius * 0.5
		for _, other := range ctx.World.Characters {
			if other == nil || other.IsAlive() || strings.Contains(other.Name, "Butchered") || (other.Actor.Config != nil && !other.Actor.Config.IsAnimal) { continue }
			if d := math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)); d < minCDist { minCDist, nCorpse = d, other }
		}
		if nCorpse != nil {
			if minCDist < 2.0 { c.Butcher(ctx, nCorpse); return true }
			c.MoveTo(ctx, nCorpse.X, nCorpse.Y); return true
		}

		// 8b. Seek Cooking (If has raw meat)
		hasRawFood := false
		for _, it := range c.Inventory { if it != nil && it.Config != nil && (it.Config.ID == "raw_meat" || strings.HasPrefix(it.Config.ID, "raw_meat_")) { hasRawFood = true; break } }
		if hasRawFood {
			var nFire *Obstacle; minFireDist := searchRadius
			for _, o := range ctx.World.Obstacles {
				if o.Alive && o.Archetype != nil && (strings.Contains(strings.ToLower(o.Archetype.ID), "campfire") || strings.Contains(strings.ToLower(o.Archetype.ID), "bakery")) {
					if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minFireDist { minFireDist, nFire = d, o }
				}
			}
			if nFire != nil {
				if minFireDist < 2.0 { c.ActionState, c.Tick = ActorCooking, 0; return true }
				c.MoveTo(ctx, nFire.X, nFire.Y); return true
			}
		}

		// 8c. Hunting (Target Prey)
		var nPrey *Character; minPDist := searchRadius
		for _, other := range ctx.World.Characters {
			if other == nil || !other.IsAlive() || other.Actor.Config == nil || !other.Actor.Config.IsAnimal || other.Alignment != AlignmentNeutral { continue }
			if d := math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)); d < minPDist { minPDist, nPrey = d, other }
		}
		if nPrey != nil {
			c.TargetActor = &nPrey.Actor
			c.ActionState = ActorAttacking // Or specialized state if we prefer, but Attacking triggers combat
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s is hunting %s.", c.Name, nPrey.Name), LogNPC) }
			return true
		}
	}

	// 8. Resting
	if isExhausted {
		var nRest *Obstacle; minDist := 50.0
		for _, o := range ctx.World.Obstacles {
			id := strings.ToLower(o.ID)
			if o.Alive && (strings.Contains(id, "tavern") || strings.Contains(id, "house") || strings.Contains(id, "farm") || strings.Contains(id, "campfire")) {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minDist { minDist, nRest = d, o }
			}
		}
		if nRest != nil && minDist > 3.0 { c.MoveTo(ctx, nRest.X, nRest.Y); return true }
		c.ActionState, c.Tick = ActorResting, 0
		return true
	}

	if c.checkEconomicSeeking(ctx) { return true }

	// 9. Wandering fallback (Stable movement)
	if (isHungry || isThirsty || isStarving || isDehydrated) && c.ActionState == ActorIdle {
		if len(c.Path) == 0 {
			tx, ty := c.X + (rand.Float64()*100 - 50), c.Y + (rand.Float64()*100 - 50)
			c.MoveTo(ctx, tx, ty)
		}
		return false
	}

	return false
}

func (c *Character) checkEconomicSeeking(ctx *SystemContext) bool {
	if (c.State.Hunger > 50 || c.State.Thirst > 50) && c.Denarii > 20 {
		var nTrader *Character; minTDist := 100.0
		for _, other := range ctx.World.Characters {
			if other == c || !other.IsAlive() { continue }
			isMerchant := other.Behavior == BehaviorTrader || strings.Contains(strings.ToLower(other.Config.ID), "merchant") || strings.Contains(strings.ToLower(other.Config.ID), "innkeeper")
			if isMerchant {
				if d := math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)); d < minTDist { minTDist, nTrader = d, other }
			}
		}
		if nTrader != nil {
			if minTDist < 2.0 { return false } 
			c.MoveTo(ctx, nTrader.X, nTrader.Y)
			return true
		}
	}

	hasTradeGoods := false
	for _, it := range c.Inventory { 
		if it != nil && it.Config != nil && (strings.Contains(it.Config.ID, "meat") || it.Config.Value > 20) { 
			hasTradeGoods = true; break 
		} 
	}
	
	if hasTradeGoods && c.Denarii < 50 {
		var nTrader *Character; minTDist := 100.0
		for _, other := range ctx.World.Characters {
			if other == c || !other.IsAlive() { continue }
			isMerchant := other.Behavior == BehaviorTrader || strings.Contains(strings.ToLower(other.Config.ID), "merchant") 
			if isMerchant {
				if d := math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)); d < minTDist { minTDist, nTrader = d, other }
			}
		}
		if nTrader != nil {
			if minTDist < 2.0 { return false } 
			c.MoveTo(ctx, nTrader.X, nTrader.Y)
			return true
		}
	}
	return false
}
