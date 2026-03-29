package game

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func (c *Character) updateAI(ctx *SystemContext) {
	// Wisdom Moderation: Wiser NPCs are more resilient to sanity loss
	// Wisdom 100 -> 0.5x sanity loss, Wisdom 0 -> 1.5x sanity loss
	sanityResilience := 1.5 - (float64(c.PrimaryAttributes.Wisdom) * 0.01)
	if sanityResilience < 0.5 { sanityResilience = 0.5 }

	if !c.IsAlive() || c.IsIncapacitated() { return }
	if c.TemporalState.Sanity < (10.0 * sanityResilience) {
		c.updateChaotic(ctx, ctx.World.Obstacles)
		return
	}

	worldObstacles, mapW, mapH := ctx.World.Obstacles, ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight
	if c.HitTimer > 0 { c.HitTimer-- }

	// VILLAGE HIERARCHY (Item 5: Leaders & Groups)
	effectiveShift := c.Shift
	if c.LeaderID != "" {
		for _, char := range ctx.World.Characters {
			if char.Name == c.LeaderID && char.IsAlive() {
				// Follow leader's shift unless exhausted
				if char.Shift == ShiftWork && c.TemporalState.Fatigue > 15 {
					effectiveShift = ShiftWork
				}
				break
			}
		}
	}

	c.HandleSocial(ctx)
	var playerDist float64
	playableCharacter := ctx.World.PlayableCharacter
	if playableCharacter != nil { playerDist = math.Sqrt(math.Pow(c.X-playableCharacter.X, 2) + math.Pow(c.Y-playableCharacter.Y, 2)) }

	if c.LeaderID != "" {
		leaderAlive := false
		for _, char := range ctx.World.Characters { if char.Config != nil && char.Config.ID == c.LeaderID && char.IsAlive() { leaderAlive = true; break } }
		if !leaderAlive { c.Alignment, c.Behavior, c.LeaderID = AlignmentNeutral, BehaviorWander, "" }
	}

	// 8-hour Labor Cycle Management using effectiveShift
	if effectiveShift == ShiftSleep && (c.TemporalState.Hunger <= 80 && c.TemporalState.Thirst <= 80) {
		c.updateSleepCycle(ctx)
		return
	}

	if c.LeaderID != "" {
		leaderAlive := false
		for _, char := range ctx.World.Characters { if char.Config != nil && char.Config.ID == c.LeaderID && char.IsAlive() { leaderAlive = true; break } }
		if !leaderAlive { c.Alignment, c.Behavior, c.LeaderID = AlignmentNeutral, BehaviorWander, "" }
	}

	isHungry := c.TemporalState.Hunger > 70
	isThirsty := c.TemporalState.Thirst > 70
	isExhausted := c.TemporalState.Fatigue > 70

	if (isHungry || isThirsty || isExhausted) && c.State != ActorResting && c.State != ActorForaging && c.State != ActorEating && c.State != ActorDrinking {
		// 1. Check if NPC has consumables in inventory
		for i, item := range c.Inventory {
			if item != nil && item.Config != nil && item.Config.Type == "consumable" {
				// Only consume if it helps the current need
				shouldConsume := false
				if isHungry && item.Config.Hunger > 0 { shouldConsume = true }
				if isThirsty && item.Config.Thirst > 0 { shouldConsume = true }
				if isExhausted && item.Config.Fatigue > 0 { shouldConsume = true }
				if item.Config.Energy > 0 { shouldConsume = true } // Legacy support

				if shouldConsume {
					if c.ConsumeItem(item, ctx) {
						c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
						c.UpdateEffects()
						return
					}
				}
			}
		}

				// 2. Not enough energy? Check for closest foraging spot
		// Intellect influence: Smarter NPCs are more aware of danger (further away)
		// Intellect 100 -> 15.0 safety radius, Intellect 0 -> 5.0 safety radius
		safetyRadius := 5.0 + (float64(c.PrimaryAttributes.Intellect) * 0.1)
		isSafe := true
		for _, other := range ctx.World.Characters { 
			if other.IsAlive() && other.Alignment != c.Alignment { 
				if math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2)) < safetyRadius { 
					isSafe = false; break 
				} 
			} 
		}
		
		if isSafe {
			// Smarter NPCs search further, but everyone needs a baseline for survival
			// Base 30.0 for safety/necessities
			searchRadius := 30.0 + (float64(c.PrimaryAttributes.Intellect) * 0.1)
			
			// 2. Check for nearby consumables on the ground
			if isHungry || isThirsty {
				var nearestItem *ItemInstance
				minIDist := searchRadius
				for _, item := range ctx.World.Items {
					if item.Config == nil || item.Config.Type != "consumable" { continue }
					// Don't pick up expired food
					if item.Config.MaxHours > 0 && item.HoursLeft <= 0 { continue }
					// Check if item helps current need
					if isHungry && item.Config.Hunger <= 0 && item.Config.Energy <= 0 { continue }
					if isThirsty && item.Config.Thirst <= 0 { continue }
					
					dist := math.Sqrt(math.Pow(c.X-item.X, 2) + math.Pow(c.Y-item.Y, 2))
					if dist < minIDist { minIDist, nearestItem = dist, item }
				}
				if nearestItem != nil {
					if minIDist < 1.0 {
						if ctx.World.Game != nil && ctx.World.Game.TryPickup(&c.Actor, nearestItem) {
							return 
						}
					} else {
						c.LastAIReasoning, c.ThoughtTimer = "Picking up food", 180
						c.executeMovement(ctx, nearestItem.X-c.X, nearestItem.Y-c.Y, worldObstacles, false)
						return
					}
				}
			}

			// 3. Find nearest foraging spot if hungry
			if isHungry {
				var nearestForage *Obstacle
				minFDist := searchRadius
				for _, o := range worldObstacles {
					if !o.Alive { continue }
					isTree := (o.Archetype.Type == TypeTree) || strings.Contains(strings.ToLower(o.ID), "tree")
					isBush := (o.Archetype.Type == TypeBush) || strings.Contains(strings.ToLower(o.ID), "bush")
					if (isTree || isBush) && o.CooldownTicks <= 0 {
						dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
						if dist < minFDist { minFDist, nearestForage = dist, o }
					}
				}
				
				if nearestForage != nil {
					if minFDist < 1.5 {
						c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorForaging, 0, "Gathering food...", 180
						return
					} else {
						dx, dy := nearestForage.X-c.X, nearestForage.Y-c.Y
						c.LastAIReasoning, c.ThoughtTimer = "Heading to find food", 180
						c.executeMovement(ctx, dx, dy, worldObstacles, false)
						return
					}
				}

				// No forage? Seek Market/Trader if has Denarii
				if c.Denarii > 0 {
					var nearestTrader *Character
					minTDist := 50.0
					for _, other := range ctx.World.Characters {
						if other.IsAlive() && other.Behavior == BehaviorTrader {
							dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
							if dist < minTDist { minTDist, nearestTrader = dist, other }
						}
					}
					if nearestTrader != nil {
						c.LastAIReasoning = "Going to the market"
						c.executeMovement(ctx, nearestTrader.X-c.X, nearestTrader.Y-c.Y, worldObstacles, false)
						return
					}
				}
				
				// Has raw meat? Seek Oven
				hasRaw := false
				for _, it := range c.Inventory { if it != nil && it.Config != nil && it.Config.ID == "raw_meat" { hasRaw = true; break } }
				if hasRaw {
					var nearestOven *Obstacle
					minODist := 30.0
					for _, o := range worldObstacles {
						if o.Alive && (o.Archetype.ID == "oven" || o.Archetype.ID == "stone_oven") {
							dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
							if dist < minODist { minODist, nearestOven = dist, o }
						}
					}
					if nearestOven != nil {
						if minODist < 2.0 {
							c.State, c.Tick, c.LastAIReasoning = ActorForaging, 0, "Cooking food..." // Reuse foraging for now
						} else {
							c.LastAIReasoning = "Heading to cook"
							c.MoveTo(ctx, nearestOven.X, nearestOven.Y)
						}
						return
					}
				}
			}

			// Find nearest water source if thirsty
			if isThirsty {
				var nearestWell *Obstacle
				minWDist := 12.0
				for _, o := range worldObstacles {
					if o.Alive && strings.Contains(strings.ToLower(o.ID), "well") {
						dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
						if dist < minWDist { minWDist, nearestWell = dist, o }
					}
				}
				if nearestWell != nil {
					if minWDist < 2.0 {
						c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorDrinking, 0, "Drinking water...", 180
						return
					} else {
						c.LastAIReasoning, c.ThoughtTimer = "Heading to the well", 180
						c.MoveTo(ctx, nearestWell.X, nearestWell.Y)
						return
					}
				}

				// No well? Look for a river/water zone
				var nearestRiver *FloorZone
				minRDist := 25.0
				if ctx.World.CurrentMapType != nil {
					for _, fz := range ctx.World.CurrentMapType.FloorZones {
						if fz.Type == "river" || fz.Type == "water" {
							dist := fz.DistanceTo(c.X, c.Y)
							if dist < minRDist { minRDist, nearestRiver = dist, fz }
						}
					}
				}
				if nearestRiver != nil {
					if minRDist < 1.0 {
						c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorDrinking, 0, "Drinking from the river...", 180
						return
					} else {
						cx, cy := (nearestRiver.MinX+nearestRiver.MaxX)*0.5, (nearestRiver.MinY+nearestRiver.MaxY)*0.5
						c.LastAIReasoning, c.ThoughtTimer = "Heading to the water", 180
						c.MoveTo(ctx, cx, cy)
						return
					}
				}
			}

			// 3. Nighttime? NPCs should seek shelter and rest
			hour := ctx.World.State.Hour
			isNight := hour >= 22 || hour < 6
			if isNight && c.State != ActorResting && c.State != ActorAttacking && c.State != ActorWalking {
				var nearestShelter *Obstacle
				minSDist := 40.0 // NPCs are willing to walk a bit for a bed
				for _, o := range worldObstacles {
					if !o.Alive { continue }
					id := strings.ToLower(o.ID)
					if strings.Contains(id, "house") || strings.Contains(id, "tavern") || strings.Contains(id, "farm") || strings.Contains(id, "campfire") {
						dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
						if dist < minSDist { minSDist, nearestShelter = dist, o }
					}
				}
				if nearestShelter != nil {
					if minSDist < 1.2 {
						c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorResting, 0, "Turning in for the night...", 180
						return
					} else {
						c.LastAIReasoning, c.ThoughtTimer = "Seeking shelter for the night", 180
						c.MoveTo(ctx, nearestShelter.X, nearestShelter.Y)
						return
					}
				}
			}

			// 4. Exhausted? Seek rest even if daytime (napping)
			if isExhausted && c.State != ActorResting {
				// Seek nearest comfy spot
				var nearestBed *Obstacle
				minBDist := 25.0
				for _, o := range worldObstacles {
					if !o.Alive { continue }
					id := strings.ToLower(o.ID)
					if strings.Contains(id, "house") || strings.Contains(id, "tavern") || strings.Contains(id, "farm") || strings.Contains(id, "campfire") {
						dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
						if dist < minBDist { minBDist, nearestBed = dist, o }
					}
				}
				if nearestBed != nil {
					if minBDist < 1.2 {
						c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorResting, 0, "I need a nap...", 180
						return
					} else {
						c.LastAIReasoning, c.ThoughtTimer = "Heading to rest", 180
						c.MoveTo(ctx, nearestBed.X, nearestBed.Y)
						return
					}
				}
				// Spontaneous rest if truly desperate
				c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorResting, 0, "I'm exhausted...", 180
				return
			}
		}
	}

	// 5. Need relief? Seek Latrine
	isUrgent := c.TemporalState.Miccionate > 75 || c.TemporalState.Defecate > 75
	if isUrgent && c.State != ActorRelieving {
		var nearestLatrine *Obstacle
		minLDist := 30.0
		for _, o := range worldObstacles {
			if !o.Alive || o.Archetype == nil { continue }
			isLat := strings.Contains(strings.ToLower(o.ID), "latrine") || strings.Contains(strings.ToLower(o.ID), "toilet")
			if isLat {
				dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
				if dist < minLDist { minLDist, nearestLatrine = dist, o }
			}
		}
		if nearestLatrine != nil {
			if minLDist < 1.5 {
				c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorRelieving, 0, "Relieving myself...", 180
				return
			} else {
				c.LastAIReasoning, c.ThoughtTimer = "Heading to latrine", 180
				c.MoveTo(ctx, nearestLatrine.X, nearestLatrine.Y)
				return
			}
		}
	}

	// 6. Filthy? Seek Bath
	isFilthy := c.TemporalState.Hygiene < 40
	if isFilthy && c.State != ActorBathing {
		var nearestBath *Obstacle
		minBathDist := 30.0
		for _, o := range worldObstacles {
			if !o.Alive || o.Archetype == nil { continue }
			isBath := strings.Contains(strings.ToLower(o.ID), "bath") || strings.Contains(strings.ToLower(o.ID), "well")
			if isBath {
				dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
				if dist < minBathDist { minBathDist, nearestBath = dist, o }
			}
		}
		if nearestBath != nil {
			if minBathDist < 1.5 {
				c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorBathing, 0, "Bathing...", 180
				return
			} else {
				c.LastAIReasoning, c.ThoughtTimer = "Heading to bath", 180
				c.MoveTo(ctx, nearestBath.X, nearestBath.Y)
				return
			}
		}
	}

	if c.State == ActorResting {
		wakeUp := c.TemporalState.Fatigue >= 100
		for _, other := range ctx.World.Characters { if other.IsAlive() && other.Alignment != c.Alignment { if math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2)) < 6.0 { wakeUp = true; break } } }
		if wakeUp { c.State = ActorIdle } else { return }
	}

	if c.State == ActorDrinking {
		if c.TemporalState.Thirst >= 100 { c.State = ActorIdle } else { return }
	}

	if c.State == ActorCrouching || c.IsTrulyDead() || c.IsIncapacitated() {
		if c.IsTrulyDead() { if c.DeadTimer == 0 { c.X, c.Y = findSafePosition(c.X, c.Y, c.GetCollisionCircle(), worldObstacles) }
			c.DeadTimer++
		}
		return
	}

	if c.State == ActorAttacking {
		// If target dies during wind-up, go back to idle to find new target
		if c.TargetActor != nil && !c.TargetActor.IsAlive() {
			c.TargetActor, c.State = nil, ActorIdle
		} else {
			if c.Tick == 15 { c.CheckAttackHits(ctx, "") }
			if c.Tick >= 30 { c.State = ActorIdle }
			return
		}
	}
	if c.State == ActorChopping || c.State == ActorDigging || c.State == ActorForaging {
		if c.Tick == 15 { c.CheckAttackHits(ctx, "") }
		if c.Tick >= 30 { c.State = ActorIdle }
		return
	}

	if c.State == ActorCooking || c.State == ActorWorkshop { return }
	
		// 3. Maintenance: Cooking & Workshop AI
	if c.State == ActorIdle || c.State == ActorWalking {
		// Cooking: Need cooked meat? Cook if we have raw meat and near campfire
		hasRawMeat := false
		for _, it := range c.Inventory { if it != nil && it.Config != nil && it.Config.ID == "raw_meat" { hasRawMeat = true; break } }
		if hasRawMeat && c.TemporalState.Hunger > 40 {
			var fire *Obstacle
			minD := 10.0
			for _, o := range worldObstacles {
				if o.Alive && strings.Contains(strings.ToLower(o.ID), "campfire") {
					dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
					if dist < minD { minD, fire = dist, o }
				}
			}
			if fire != nil {
				if minD < 2.0 { c.State, c.Tick, c.LastAIChoice = ActorCooking, 0, "cook"; return }
				c.executeMovement(ctx, fire.X-c.X, fire.Y-c.Y, worldObstacles, false); return
			}
		}

		// Workshop: Repair gear if degraded
		needsRepair := false
		for _, it := range c.Slots { if it != nil && it.Config != nil && it.Resistance < it.Config.Resistance { needsRepair = true; break } }
		if needsRepair {
			var bench *Obstacle
			minD := 12.0
			for _, o := range worldObstacles {
				id := strings.ToLower(o.ID)
				if o.Alive && (strings.Contains(id, "workbench") || strings.Contains(id, "workshop") || strings.Contains(id, "anvil")) {
					dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
					if dist < minD { minD, bench = dist, o }
				}
			}
			if bench != nil {
				if minD < 2.0 { c.State, c.Tick, c.LastAIChoice = ActorWorkshop, 0, "workshop"; return }
				c.executeMovement(ctx, bench.X-c.X, bench.Y-c.Y, worldObstacles, false); return
			}
		}
	}

	if ctx.AIManager != nil && !c.AIDecisionPending && c.needsAIDecision(playerDist) {
		worldCtx := BuildWorldContext(ctx.World.Game, c)
		options := []string{"attack", "flee", "wander", "patrol", "trade", "forage", "cook", "rest"}
		ctx.AIManager.RequestDecision(context.Background(), c.Config.ID, worldCtx, options)
		c.AIDecisionPending = true
	}

	canLoot := c.MaxWeight > 0 && (c.Config == nil || !c.Config.IsAnimal)
	if c.Tick%30 == 0 && canLoot { c.TargetItem = c.findLootTarget(ctx.World.Items) }

	if c.TargetItem != nil && c.TargetItem.Pickable && canLoot {
		dx, dy := c.TargetItem.X-c.X, c.TargetItem.Y-c.Y
		if math.Sqrt(dx*dx+dy*dy) < 1.5 {
			if ctx.World.Game != nil && ctx.World.Game.TryPickup(&c.Actor, c.TargetItem) {
				c.EquipItem(c.TargetItem)
				itList := []*ItemInstance{}
				for _, it := range ctx.World.Items { if it != c.TargetItem { itList = append(itList, it) } }
				ctx.World.Items, c.TargetItem = itList, nil
				return
			}
			c.TargetItem, c.State = nil, ActorIdle
		} else { c.executeMovement(ctx, dx, dy, worldObstacles, false); return }
	}

	targetX, targetY, hasTarget, isTargetPlayer := c.findTarget(playableCharacter, ctx.World.Characters, playerDist)
	if !hasTarget {
		if c.Behavior == BehaviorHauler {
			c.updateHauler(ctx, worldObstacles)
			c.clampToMap(mapW, mapH)
			return
		} else if c.Behavior == BehaviorLumberjack {
			c.updateLumberjack(ctx, worldObstacles)
			c.clampToMap(mapW, mapH)
			return
		} else if c.Behavior == BehaviorFarmer {
			c.updateFarmer(ctx, worldObstacles)
			c.clampToMap(mapW, mapH)
			return
		} else if c.Behavior == BehaviorArtisan {
			c.updateArtisan(ctx, worldObstacles)
			c.clampToMap(mapW, mapH)
			return
		} else if c.Behavior == BehaviorTrader {
			c.clampToMap(mapW, mapH)
			return // Traders stay put or interact via context
		}

		if c.Behavior == BehaviorWander { c.updateWander(ctx, worldObstacles) } else if c.Behavior == BehaviorPatrol { c.updatePatrol(ctx, worldObstacles)
		} else if c.Alignment == AlignmentAlly && playableCharacter != nil && playerDist > 5.0 && playerDist < 20.0 { c.executeMovement(ctx, playableCharacter.X-c.X, playableCharacter.Y-c.Y, worldObstacles, false)
		} else { c.State = ActorIdle }
		c.clampToMap(mapW, mapH); return
	}

	dx, dy := targetX-c.X, targetY-c.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	attackRange := 1.4
	if c.RawStats.AttackRange > 0 { attackRange = c.RawStats.AttackRange }
	if c.Weapon != nil { attackRange = c.Weapon.GetMaxDistance() }

	canAttack := (c.TargetActor != nil && c.Alignment != c.TargetActor.Alignment) || c.Behavior == BehaviorChaotic
	tooClose := dist < attackRange*0.5 && c.Weapon != nil && c.Weapon.IsRanged()
	if tooClose && canAttack { c.executeMovement(ctx, dx, dy, worldObstacles, true) } else if dist < attackRange && canAttack { c.executeAttack(ctx, isTargetPlayer, dx, dy)
	} else { c.executeMovement(ctx, dx, dy, worldObstacles, c.Behavior == BehaviorFlee) }
	c.clampToMap(mapW, mapH)
}

func (c *Character) updateWander(ctx *SystemContext, obstacles []*Obstacle) {
	if c.Tick%120 == 0 || (c.WanderDirX == 0 && c.WanderDirY == 0) {
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

func (c *Character) findTarget(player *Character, others []*Character, playerDist float64) (float64, float64, bool, bool) {
	if c.TargetActor != nil && !c.TargetActor.IsAlive() {
		c.TargetActor = nil
	}
	var bestX, bestY float64; var hasTarget, isTargetPlayer bool; minDist := 15.0
	isTargetValid := func(other *Character) bool {
		if c.Relationships != nil {
			if sentiment, ok := c.Relationships[other.Name]; ok && sentiment < -20.0 {
				return true // Grudge!
			}
		}
		if c.Behavior == BehaviorChaotic { return true }
		if c.Alignment == AlignmentEnemy { return other.Alignment == AlignmentAlly || other.LeaderID != "" || other.Group != ""
		} else if c.Alignment == AlignmentAlly { return other.Alignment == AlignmentEnemy
		} else if c.Alignment == AlignmentNeutral { return c.TargetActor == &other.Actor }
		return false
	}
	if player != nil && player.IsAlive() && playerDist < minDist && isTargetValid(player) {
		minDist, bestX, bestY, hasTarget, isTargetPlayer, c.TargetActor = playerDist, player.X, player.Y, true, true, &player.Actor
	}
	for _, other := range others {
		if other == c || !other.IsAlive() { continue }
		dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
		if dist < minDist && isTargetValid(other) {
			minDist, bestX, bestY, hasTarget, isTargetPlayer, c.TargetActor = dist, other.X, other.Y, true, false, &other.Actor
		}
	}
	return bestX, bestY, hasTarget, isTargetPlayer
}

func (c *Character) findLootTarget(items []*ItemInstance) *ItemInstance {
	var best *ItemInstance; minDist := 10.0
	for _, it := range items {
		if !it.Pickable { continue }
		dist := math.Sqrt(math.Pow(c.X-it.X, 2) + math.Pow(c.Y-it.Y, 2))
		if dist < minDist { minDist, best = dist, it }
	}
	return best
}

func (c *Character) needsAIDecision(playerDist float64) bool {
	if playerDist < 10.0 || (c.TemporalState.HealthPoints < c.TemporalState.MaxHealthPoints/2 && playerDist < 20.0) {
		interval := 300
		if IsDebugEnabled() { interval = 60 }
		return (c.Tick - c.LastAIDecisionTick) >= interval
	}
	return false
}
// HandleSocial manages non-combat interactions with nearby NPCs.
func (c *Character) HandleSocial(ctx *SystemContext) {
	// Don't interact every tick
	if c.Tick%180 != 0 { return }

	for _, other := range ctx.World.Characters {
		if other == c || !other.IsAlive() { continue }
		
		dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
		if dist > 3.0 { continue }

		// 1. Check existing relationship (adjusted for hygiene)
		sentiment := c.GetEffectiveSentiment(&other.Actor)
		
		// 2. High Intellect NPCs are more likely to start interactions
		socialPromptChance := 0.05 + (float64(c.PrimaryAttributes.Intellect) * 0.001)
		
		if rand.Float64() < socialPromptChance {
			// Choose a social action
			action := "talk"
			
			// DOCTORS / HEALERS (Item 4: Sepsis Treatment)
			if (c.Behavior == BehaviorArtisan || c.GetAbilityYield("herbalism") > 60) && other.TemporalState.IsSeptic {
				action = "treat_infection"
			} else if c.TemporalState.Hunger > 75 || c.TemporalState.Thirst > 75 {
				action = "request_food"
			} else if sentiment < -30 && c.PrimaryAttributes.Strength > other.PrimaryAttributes.Strength {
				action = "intimidate"
			} else if (c.Behavior == BehaviorTrader || other.Behavior == BehaviorTrader || rand.Float64() < 0.1) && (c.Denarii > 0 || other.Denarii > 0) {
				action = "trade" // Item 2: Autonomous Economic Exchange
			} else if sentiment > 30 && c.Config.Gender != other.Config.Gender {
				action = "seduce"
			}

			switch action {
			case "treat_infection":
				other.TemporalState.IsSeptic = false
				other.ModifySentiment(c.Name, 15.0)
				if ctx.Log != nil && playerNear(c, ctx) {
					ctx.Log(fmt.Sprintf("%s treated %s's severe infection.", c.Name, other.Name), LogNPC)
				}
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Infection Cured!", X: other.X, Y: other.Y - 1.0, Life: 60, Color: ColorHeal })

			case "request_food":
				// Find if other has food
				foundIdx := -1
				for i, it := range other.Inventory {
					if it != nil && it.Config != nil && (it.Config.Hunger > 0 || it.Config.Thirst > 0) {
						foundIdx = i; break
					}
				}
				if foundIdx >= 0 {
					item := other.Inventory[foundIdx]
					// Give item if friendly or intimidated
					willGive := other.Relationships[c.Name] > 20 || other.Submission[c.Name] > 40
					if willGive {
						other.Inventory = append(other.Inventory[:foundIdx], other.Inventory[foundIdx+1:]...)
						c.Inventory = append(c.Inventory, item)
						c.ModifySentiment(other.Name, 10.0)
						other.ModifySentiment(c.Name, 5.0)
						if ctx.Log != nil && playerNear(c, ctx) {
							ctx.Log(fmt.Sprintf("%s gave some food to the hungry %s.", other.Name, c.Name), LogNPC)
						}
					} else {
						c.ModifySentiment(other.Name, -5.0) // Refusal breeds resentment
					}
				}
			case "trade":
				// Simple trade logic: c buys something from other if c lacks it
				for i, it := range other.Inventory {
					if it == nil || it.Config == nil { continue }
					price := int(float64(it.Config.Value) * (1.2 - (c.GetAbilityYield("trade") * 0.002)))
					if price < 1 { price = 1 }
					if c.Denarii >= price {
						c.Denarii -= price
						other.Denarii += price
						other.Inventory = append(other.Inventory[:i], other.Inventory[i+1:]...)
						c.Inventory = append(c.Inventory, it)
						c.ModifySentiment(other.Name, 2.0)
						other.ModifySentiment(c.Name, 2.0)
						if ctx.Log != nil && playerNear(c, ctx) {
							ctx.Log(fmt.Sprintf("%s bought %s from %s for %d denarii.", c.Name, it.Config.Name, other.Name, price), LogNPC)
						}
						break
					}
				}
			case "talk":
				// Neutral interaction improves sentiment slightly
				c.ModifySentiment(other.Name, 1.0)
				other.ModifySentiment(c.Name, 1.0)
				// Factional ripple (Item 3)
				c.ModifyGroupSentiment(ctx, other.Group, 0.1)
				if ctx.Log != nil && playerNear(c, ctx) {
					ctx.Log(fmt.Sprintf("%s and %s are chatting.", c.Name, other.Name), LogNPC)
				}
			case "intimidate":
				// Attacker rolls Intimidate vs Defender's Wisdom/Intellect (Willpower)
				if c.CompetitiveAttributeRoll(&other.Actor, "culture") {
					// Success! Subjugate.
					other.ModifySentiment(c.Name, -10.0)
					other.ModifySubmission(c.Name, 15.0)
					c.ModifyGroupSentiment(ctx, other.Group, -0.5)
					if ctx.Log != nil && playerNear(c, ctx) {
						ctx.Log(fmt.Sprintf("%s cowed %s into submission.", c.Name, other.Name), LogNPC)
					}
				} else {
					// Fail! Target gets angry.
					other.ModifySentiment(c.Name, -5.0)
					other.TemporalState.IsAngry = true
					c.ModifyGroupSentiment(ctx, other.Group, -1.0)
				}
			case "seduce":
				// Roll Art/Dexterity vs Health (Attractiveness/Resistance)
				if c.CompetitiveAttributeRoll(&other.Actor, "art") {
					c.RomanticInterest[other.Name] += 10.0
					other.RomanticInterest[c.Name] += 10.0
					c.ModifyGroupSentiment(ctx, other.Group, 0.5)
					if ctx.Log != nil && playerNear(c, ctx) {
						ctx.Log(fmt.Sprintf("%s shared a romantic moment with %s.", c.Name, other.Name), LogNPC)
					}
				}
			}
		}
	}
}

func (c *Character) ModifyGroupSentiment(ctx *SystemContext, otherGroup string, delta float64) {
	if c.Group == "" || otherGroup == "" || ctx.World.State.GroupSentiment == nil { return }
	if ctx.World.State.GroupSentiment[c.Group] == nil { ctx.World.State.GroupSentiment[c.Group] = make(map[string]float64) }
	ctx.World.State.GroupSentiment[c.Group][otherGroup] += delta
	// Clamping
	if ctx.World.State.GroupSentiment[c.Group][otherGroup] > 100 { ctx.World.State.GroupSentiment[c.Group][otherGroup] = 100 }
	if ctx.World.State.GroupSentiment[c.Group][otherGroup] < -100 { ctx.World.State.GroupSentiment[c.Group][otherGroup] = -100 }
}

func playerNear(c *Character, ctx *SystemContext) bool {
	pc := ctx.World.PlayableCharacter
	if pc == nil { return false }
	return math.Sqrt(math.Pow(c.X-pc.X, 2) + math.Pow(c.Y-pc.Y, 2)) < 15.0
}

func (c *Character) ApplyAIDecision(ctx *SystemContext, dec AIDecision) {
	c.AIDecisionPending, c.LastAIChoice, c.LastAIReasoning = false, dec.ChosenOption, dec.Reasoning
	c.ThoughtTimer = 180
	choice := strings.ToLower(dec.ChosenOption)
	if strings.Contains(choice, "attack") { c.Behavior = BehaviorNpcFighter
	} else if strings.Contains(choice, "flee") { c.Behavior = BehaviorFlee
	} else if strings.Contains(choice, "wander") || strings.Contains(choice, "talk") { c.Behavior = BehaviorWander
	} else if strings.Contains(choice, "patrol") { c.Behavior = BehaviorPatrol
	} else if strings.Contains(choice, "trade") { c.Behavior = BehaviorTrader
	} else if strings.Contains(choice, "forage") { c.State, c.Tick = ActorForaging, 0
	} else if strings.Contains(choice, "cook") {
		// 1. Check if we have raw meat
		hasRaw := false
		for _, it := range c.Inventory { if it != nil && it.Config != nil && it.Config.ID == "raw_meat" { hasRaw = true; break } }
		
		if hasRaw {
			// 2. Find closest campfire
			var fire *Obstacle
			minDist := 15.0
			for _, o := range ctx.World.Obstacles {
				if o.Alive && strings.Contains(strings.ToLower(o.ID), "campfire") {
					dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
					if dist < minDist { minDist, fire = dist, o }
				}
			}
			
			if fire != nil {
				if minDist < 1.5 {
					c.State, c.Tick = ActorCooking, 0
				} else {
					c.executeMovement(ctx, fire.X-c.X, fire.Y-c.Y, ctx.World.Obstacles, false)
				}
			} else { c.Behavior = BehaviorWander } // No fire? Just wander.
		} else { c.Behavior = BehaviorWander } // No meat? Just wander.
	} else if strings.Contains(choice, "rest") { c.State, c.Tick = ActorResting, 0 }
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
			c.State, c.Tick = ActorChopping, 0
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
				c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorChopping, 0, "Harvesting crops", 180
			} else {
				c.MoveTo(ctx, c.TargetObstacle.X, c.TargetObstacle.Y)
			}
			return
		}
	}

	// 2. Husbandry: Milk animals if none to harvest or in harvest season but none ready
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
			c.State, c.Tick, c.LastAIReasoning, c.TargetActorID = ActorMilking, 0, "Milking animal", nearestAnimal.Name
		} else {
			c.MoveTo(ctx, nearestAnimal.X, nearestAnimal.Y)
		}
		return
	}

	// 2. Plant crops in Spring/Summer 
	if isPlantingSeason && c.State == ActorIdle {
		// Are we in a farm field?
		inField := false
		if ctx.World.CurrentMapType != nil {
			for _, fz := range ctx.World.CurrentMapType.FloorZones {
				if fz.Type == "farm_field" && fz.Contains(c.X, c.Y) {
					inField = true
					break
				}
			}
		}

		if inField {
			// Check if already occupied
			occupied := false
			for _, o := range obstacles {
				if o.Alive && math.Abs(o.X-c.X) < 1.0 && math.Abs(o.Y-c.Y) < 1.0 {
					occupied = true
					break
				}
			}

			if !occupied && rand.Float64() < 0.05 {
				// Plant!
				cropIDs := []string{"crop_wheat", "crop_cabbage", "crop_turnip"}
				choice := cropIDs[rand.Intn(len(cropIDs))]
				if arch, ok := ctx.Registries.Obstacles.Archetypes[choice]; ok {
					newObs := NewObstacle(fmt.Sprintf("crop_%d", rand.Int()), c.X, c.Y, arch)
					ctx.World.Obstacles = append(ctx.World.Obstacles, newObs)
					c.LastAIReasoning, c.ThoughtTimer = "Sowing seeds", 180
					return
				}
			}
		} else {
			// Move towards nearest farm field
			var nearestField *FloorZone
			minD := 50.0
			if ctx.World.CurrentMapType != nil {
				for _, fz := range ctx.World.CurrentMapType.FloorZones {
					if fz.Type == "farm_field" {
						cx, cy := (fz.MinX+fz.MaxX)*0.5, (fz.MinY+fz.MaxY)*0.5
						dist := math.Sqrt(math.Pow(c.X-cx, 2) + math.Pow(c.Y-cy, 2))
						if dist < minD { minD, nearestField = dist, fz }
					}
				}
			}
			if nearestField != nil {
				cx, cy := (nearestField.MinX+nearestField.MaxX)*0.5, (nearestField.MinY+nearestField.MaxY)*0.5
				c.executeMovement(ctx, cx-c.X, cy-c.Y, obstacles, false)
				return
			}
		}
	}

	// Default to idle wander
	c.updateWander(ctx, obstacles)
}

func (c *Character) updateChaotic(ctx *SystemContext, obstacles []*Obstacle) {
	if c.TargetActor == nil || c.TargetActor.State == ActorDead {
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
			c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorAttacking, 0, "Tantrum!", 60
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
			c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorAttacking, 0, "Smashing things!", 60
		} else {
			c.executeMovement(ctx, dx, dy, obstacles, false)
		}
		return
	}

	c.updateWander(ctx, obstacles)
}
func (c *Character) updateSleepCycle(ctx *SystemContext) {
	if c.State == ActorResting {
		if c.TemporalState.Fatigue >= 100 { 
			c.State, c.Tick, c.LastAIReasoning = ActorIdle, 0, "Well rested!"
		}
		return
	}

	// Seek nearest bed/house
	var nearestHouse *Obstacle
	minDist := 40.0 // Look in the neighborhood
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
			c.State, c.Tick, c.LastAIReasoning = ActorResting, 0, "Sleeping at home..."
		} else {
			c.LastAIReasoning = "Heading home to sleep"
			c.executeMovement(ctx, nearestHouse.X-c.X, nearestHouse.Y-c.Y, ctx.World.Obstacles, false)
		}
	} else {
		// Just rest nearby
		c.State, c.Tick, c.LastAIReasoning = ActorResting, 0, "Resting on the grass..."
	}
}

func (c *Character) updateArtisan(ctx *SystemContext, obstacles []*Obstacle) {
	// 1. If carrying finished products, go to stockpile
	hasProduct := false
	for _, it := range c.Inventory {
		id := strings.ToLower(it.Config.ID)
		if id == "iron_ingot" || id == "bread" || strings.Contains(id, "tool") {
			hasProduct = true
			break
		}
	}
	if hasProduct {
		c.updateHauler(ctx, obstacles)
		return
	}

	// 2. Look for work
	if c.TargetObstacle == nil || !c.TargetObstacle.Alive {
		var nearestWorkstation *Obstacle
		minD := 20.0
		for _, o := range obstacles {
			if !o.Alive || o.Archetype == nil { continue }
			id := strings.ToLower(o.Archetype.ID)
			// Check for workbench, anvil, furnace, oven
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
			// Smarters NPCs choose better tasks
			c.State, c.Tick, c.LastAIReasoning, c.ThoughtTimer = ActorWorkshop, 0, "Working at the station", 300
		} else {
			c.MoveTo(ctx, c.TargetObstacle.X, c.TargetObstacle.Y)
		}
		return
	}

	// Default to idle wander
	c.updateWander(ctx, obstacles)
}
