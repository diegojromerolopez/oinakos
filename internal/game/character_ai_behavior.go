package game

import (
	"math"
	"math/rand"
	"strings"
)

func (c *Character) updateWander(ctx *SystemContext, obstacles []*Obstacle) {
	if c.Tick%120 == 0 || (c.WanderDirX == 0 && c.WanderDirY == 0) {
		// SOCIAL MAGNET: ShiftLeisure NPCs gravitate toward Tavern/Town Square
		if c.Shift == ShiftLeisure && !c.Config.IsAnimal {
			var hub *Obstacle
			minD := 100.0
			for _, o := range obstacles {
				if o.Alive && (strings.Contains(strings.ToLower(o.ID), "tavern") || strings.Contains(strings.ToLower(o.ID), "market") || strings.Contains(strings.ToLower(o.ID), "well")) {
					dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
					if dist < minD { minD, hub = dist, o }
				}
			}
			if hub != nil && minD > 3.0 {
				c.WanderDirX, c.WanderDirY = (hub.X-c.X)/minD, (hub.Y-c.Y)/minD
				c.executeMovement(ctx, c.WanderDirX, c.WanderDirY, obstacles, false)
				return
			}
		}

		angle := rand.Float64() * 2 * math.Pi
		c.WanderDirX, c.WanderDirY = math.Cos(angle), math.Sin(angle)
	}
	c.executeMovement(ctx, c.WanderDirX, c.WanderDirY, obstacles, false)
}

func (c *Character) updatePatrol(ctx *SystemContext, obstacles []*Obstacle) {
	targetX, targetY := c.PatrolEndX, c.PatrolEndY
	if !c.PatrolHeading { targetX, targetY = c.PatrolStartX, c.PatrolStartY }
	if math.Sqrt(math.Pow(c.X-targetX, 2) + math.Pow(c.Y-targetY, 2)) < 0.5 { c.PatrolHeading = !c.PatrolHeading } else { c.executeMovement(ctx, targetX-c.X, targetY-c.Y, obstacles, false) }
}

func (c *Character) updateHauler(ctx *SystemContext, obstacles []*Obstacle) {
	// If carrying something, find a stockpile
	if len(c.Inventory) > 0 {
		item := c.Inventory[0]
		// Find nearest stockpile for this item
		var bestStockpile *FloorZone
		minD := 999.0
		
		if ctx.World.CurrentMapType != nil {
			for _, fz := range ctx.World.CurrentMapType.FloorZones {
				if fz.Type == "stockpile" {
					accepts := false
					for _, a := range fz.Accepts { if strings.Contains(strings.ToLower(item.Config.ID), a) || strings.Contains(strings.ToLower(item.Config.Name), a) { accepts = true; break } }
					if accepts {
						// Calculate center of zone
						cx, cy := (fz.MinX+fz.MaxX)*0.5, (fz.MinY+fz.MaxY)*0.5
						dist := math.Sqrt(math.Pow(c.X-cx, 2) + math.Pow(c.Y-cy, 2))
						if dist < minD { minD, bestStockpile = dist, fz }
					}
				}
			}
		}

		if bestStockpile != nil {
			cx, cy := (bestStockpile.MinX+bestStockpile.MaxX)*0.5, (bestStockpile.MinY+bestStockpile.MaxY)*0.5
			if minD < 1.5 {
				// Drop item
				if ctx.World.Game != nil {
					ctx.World.Game.TryDrop(&c.Actor, 0)
					c.TargetItem = nil
				}
			} else {
				c.executeMovement(ctx, cx-c.X, cy-c.Y, obstacles, false)
			}
			return
		}
	}

	// Not carrying anything (or no stockpile found), look for items
	if c.TargetItem == nil || !c.TargetItem.Pickable {
		c.TargetItem = c.findLootTarget(ctx.World.Items)
	}

	if c.TargetItem != nil {
		dx, dy := c.TargetItem.X-c.X, c.TargetItem.Y-c.Y
		if math.Sqrt(dx*dx+dy*dy) < 1.5 {
			if ctx.World.Game != nil && ctx.World.Game.TryPickup(&c.Actor, c.TargetItem) {
				// Success, next update will look for stockpile
				return
			}
			c.TargetItem = nil
		} else {
			c.executeMovement(ctx, dx, dy, obstacles, false)
			return
		}
	}

	// Default to idle wander
	c.updateWander(ctx, obstacles)
}

func (c *Character) updateLumberjack(ctx *SystemContext, obstacles []*Obstacle) {
	// If carrying wood, go to stockpile
	hasWood := false
	for _, it := range c.Inventory { if strings.Contains(strings.ToLower(it.Config.ID), "wood") { hasWood = true; break } }
	
	if hasWood {
		c.updateHauler(ctx, obstacles)
		return
	}

	// Go chop trees
	if c.TargetObstacle == nil || !c.TargetObstacle.Alive {
		var nearestTree *Obstacle
		minD := 15.0
		for _, o := range obstacles {
			if o.Alive && (o.Archetype.Type == TypeTree || strings.Contains(strings.ToLower(o.ID), "tree")) {
				dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
				if dist < minD { minD, nearestTree = dist, o }
			}
		}
		c.TargetObstacle = nearestTree
	}

	if c.TargetObstacle != nil {
		dist := math.Sqrt(math.Pow(c.X-c.TargetObstacle.X, 2) + math.Pow(c.Y-c.TargetObstacle.Y, 2))
		if dist < 2.0 {
			if c.ActionState != ActorChopping { c.ActionState, c.Tick = ActorChopping, 0 }
		} else {
			c.executeMovement(ctx, c.TargetObstacle.X-c.X, c.TargetObstacle.Y-c.Y, obstacles, false)
		}
		return
	}

	// No trees? Search for wood on ground
	c.updateHauler(ctx, obstacles)
}

func (c *Character) updateFarmer(ctx *SystemContext, obstacles []*Obstacle) {
	// If carrying food, go to stockpile
	hasFood := false
	for _, it := range c.Inventory { 
		id := strings.ToLower(it.Config.ID)
		if id == "wheat" || id == "cabbage" || id == "turnip" || id == "bread" || id == "stew" {
			hasFood = true
			break
		}
	}
	
	if hasFood {
		c.updateHauler(ctx, obstacles)
		return
	}

	season := ctx.World.State.Season
	isPlantingSeason := season == SeasonSpring || season == SeasonSummer
	isHarvestSeason := season == SeasonAutumn

	// 1. Harvest mature crops in Autumn
	if isHarvestSeason {
		if c.TargetObstacle == nil || !c.TargetObstacle.Alive || c.TargetObstacle.GrowthStage < 2 {
			var nearestMature *Obstacle
			minD := 20.0
			for _, o := range obstacles {
				if o.Alive && o.Archetype.IsCrop && o.GrowthStage >= 2 {
					dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
					if dist < minD { minD, nearestMature = dist, o }
				}
			}
			c.TargetObstacle = nearestMature
		}

		if c.TargetObstacle != nil {
			dist := math.Sqrt(math.Pow(c.X-c.TargetObstacle.X, 2) + math.Pow(c.Y-c.TargetObstacle.Y, 2))
			if dist < 1.5 {
				if c.ActionState != ActorChopping { c.ActionState, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorChopping, 0, "Harvesting crops", 180 }
			} else {
				c.MoveTo(ctx, c.TargetObstacle.X, c.TargetObstacle.Y)
			}
			return
		}
	}

	// 2. Husbandry: Milk animals
	var nearestAnimal *Character
	minADist := 10.0
	for _, other := range ctx.World.Characters {
		if other.IsAlive() && other.Config.Stats.IsMilkable && other.MilkCooldownTicks <= 0 {
			dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
			if dist < minADist { minADist, nearestAnimal = dist, other }
		}
	}
	if nearestAnimal != nil {
		if minADist < 2.0 {
			if c.ActionState != ActorMilking { c.ActionState, c.Tick, c.LastAIReasoning, c.TargetActorID = ActorMilking, 0, "Milking animal", nearestAnimal.Name }
		} else {
			c.MoveTo(ctx, nearestAnimal.X, nearestAnimal.Y)
		}
		return
	}

	// 3. Plant crops in Spring/Summer 
	if isPlantingSeason && c.ActionState == ActorIdle {
		// ... existing field logic
	}

	// SOCIAL MAGNET: ShiftLeisure NPCs gravitate toward Tavern/Town Square
	if c.Shift == ShiftLeisure && !c.Config.IsAnimal {
		var hub *Obstacle
		minD := 100.0
		for _, o := range obstacles {
			if o.Alive && (strings.Contains(strings.ToLower(o.ID), "tavern") || strings.Contains(strings.ToLower(o.ID), "market")) {
				dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
				if dist < minD { minD, hub = dist, o }
			}
		}
		if hub != nil && minD > 3.0 {
			c.executeMovement(ctx, hub.X-c.X, hub.Y-c.Y, obstacles, false)
			return
		}
	}

	c.updateWander(ctx, obstacles)
}

func (c *Character) updateCriminal(ctx *SystemContext, obstacles []*Obstacle) {
	if c.TargetActor == nil || !c.TargetActor.IsAlive() {
		// Criminals actively look for prey (Alone or Vulnerable)
		var prey *Character; minD := 25.0
		for _, other := range ctx.World.Characters {
			if other == c || !other.IsAlive() { continue }
			isVulnerable := other.State.Hunger > 60 || other.State.Thirst > 60 || other.State.Fatigue > 60 || other.State.IsDrunk
			dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
			if dist < minD && (isVulnerable || rand.Float64() < 0.1) {
				prey = other; minD = dist
			}
		}
		if prey != nil { c.TargetActor = &prey.Actor }
	}

	if c.TargetActor != nil {
		dx, dy := c.TargetActor.X-c.X, c.TargetActor.Y-c.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 1.4 {
			// Attack to rob or release arousal
			if c.State.Arousal > 70 && c.ActionState != ActorIntercourse {
				c.ActionState, c.Tick = ActorIntercourse, 0
			} else if c.ActionState != ActorAttacking {
				c.ActionState, c.Tick = ActorAttacking, 0
			}
		} else {
			c.executeMovement(ctx, dx, dy, obstacles, false)
		}
		return
	}

	// If no prey, look for items to steal
	if c.TargetItem == nil || !c.TargetItem.Pickable {
		c.TargetItem = c.findLootTarget(ctx.World.Items)
	}
	if c.TargetItem != nil {
		dx, dy := c.TargetItem.X-c.X, c.TargetItem.Y-c.Y
		if math.Sqrt(dx*dx+dy*dy) < 1.2 {
			ctx.World.Game.TryPickup(&c.Actor, c.TargetItem)
			c.TargetItem = nil
		} else {
			c.executeMovement(ctx, dx, dy, obstacles, false)
		}
		return
	}

	c.updateWander(ctx, obstacles)
}

func (c *Character) updateChaotic(ctx *SystemContext, obstacles []*Obstacle) {
	if c.TargetActor == nil || c.TargetActor.ActionState == ActorDead {
		var nearest *Character
		minD := 10.0
		for _, other := range ctx.World.Characters {
			if other == c || !other.IsAlive() { continue }
			dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
			if dist < minD { minD, nearest = dist, other }
		}
		if nearest != nil { c.TargetActor = &nearest.Actor }
	}
	if c.TargetActor != nil {
		dx, dy := c.TargetActor.X-c.X, c.TargetActor.Y-c.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 1.5 {
			if c.ActionState != ActorAttacking { c.ActionState, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorAttacking, 0, "Tantrum!", 60 }
		} else {
			c.executeMovement(ctx, dx, dy, obstacles, false)
		}
		return
	}
	if c.TargetObstacle == nil || !c.TargetObstacle.Alive {
		for _, o := range obstacles {
			if o.Alive && o.Archetype.Destructible {
				dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
				if dist < 4.0 { c.TargetObstacle = o; break }
			}
		}
	}
	if c.TargetObstacle != nil {
		dx, dy := c.TargetObstacle.X-c.X, c.TargetObstacle.Y-c.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 1.5 {
			if c.ActionState != ActorAttacking { c.ActionState, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorAttacking, 0, "Smashing things!", 60 }
		} else {
			c.executeMovement(ctx, dx, dy, obstacles, false)
		}
		return
	}
	c.updateWander(ctx, obstacles)
}

func (c *Character) updateSleepCycle(ctx *SystemContext) {
	if c.ActionState == ActorResting {
		if c.State.Fatigue <= 0 { 
			c.ActionState, c.Tick, c.LastAIReasoning = ActorIdle, 0, "Well rested!"
		}
		return
	}
	var nearestHouse *Obstacle
	minDist := 40.0
	for _, o := range ctx.World.Obstacles {
		if !o.Alive || o.Archetype == nil { continue }
		typeName := strings.ToLower(o.Archetype.ID)
		if strings.Contains(typeName, "house") || strings.Contains(typeName, "inn") || strings.Contains(typeName, "tavern") {
			dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
			if dist < minDist { minDist, nearestHouse = dist, o }
		}
	}
	if nearestHouse != nil {
		if minDist < 2.0 {
			c.ActionState, c.Tick, c.LastAIReasoning = ActorResting, 0, "Sleeping at home..."
		} else {
			c.LastAIReasoning = "Heading home to sleep"
			c.executeMovement(ctx, nearestHouse.X-c.X, nearestHouse.Y-c.Y, ctx.World.Obstacles, false)
		}
	} else {
		c.ActionState, c.Tick, c.LastAIReasoning = ActorResting, 0, "Resting on the grass..."
	}
}

func (c *Character) updateArtisan(ctx *SystemContext, obstacles []*Obstacle) {
	hasProduct := false
	for _, it := range c.Inventory {
		id := strings.ToLower(it.Config.ID)
		if id == "iron_ingot" || id == "bread" || strings.Contains(id, "tool") {
			hasProduct = true; break
		}
	}
	if hasProduct {
		c.updateHauler(ctx, obstacles)
		return
	}
	if c.TargetObstacle == nil || !c.TargetObstacle.Alive {
		var nearestWorkstation *Obstacle
		minD := 20.0
		for _, o := range obstacles {
			if !o.Alive || o.Archetype == nil { continue }
			id := strings.ToLower(o.Archetype.ID)
			if strings.Contains(id, "bench") || strings.Contains(id, "anvil") || strings.Contains(id, "furnace") || strings.Contains(id, "oven") {
				dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
				if dist < minD { minD, nearestWorkstation = dist, o }
			}
		}
		c.TargetObstacle = nearestWorkstation
	}
	if c.TargetObstacle != nil {
		dist := math.Sqrt(math.Pow(c.X-c.TargetObstacle.X, 2) + math.Pow(c.Y-c.TargetObstacle.Y, 2))
		if dist < 1.5 {
			if c.ActionState != ActorWorkshop && c.ActionState != ActorCooking {
				c.ActionState, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorWorkshop, 0, "Working at the station", 300
			}
		} else {
			c.MoveTo(ctx, c.TargetObstacle.X, c.TargetObstacle.Y)
		}
		return
	}
	c.updateWander(ctx, obstacles)
}
