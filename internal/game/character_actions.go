package game

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strings"
	"oinakos/internal/engine"
)

func (c *Character) CheckAttackHits(ctx *SystemContext, skill string) {
	if ctx.Log != nil {
		actionDesc := "is attacking"
		if skill != "" { actionDesc = "is using " + skill }
		if c.State == ActorChopping { actionDesc = "is chopping" }
		if c.State == ActorDigging { actionDesc = "is digging" }
		if c.State == ActorForaging { actionDesc = "is foraging" }
		ctx.Log(fmt.Sprintf("[%s]: %s", c.Name, actionDesc), LogNPC)
	}
	attackDist := 2.5
	if c.RawStats.AttackRange > 0 { attackDist = c.RawStats.AttackRange }
	if c.Weapon != nil { attackDist = c.Weapon.GetMaxDistance() }
	if c.State == ActorChopping || c.State == ActorDigging || c.State == ActorForaging {
		attackDist = 5.0
	}
	atX, atY := c.X, c.Y
	switch c.Facing {
	case DirSE: atX += attackDist
	case DirSW: atY += attackDist
	case DirNE: atY -= attackDist
	case DirNW: atX -= attackDist
	}

	// [DEBUG] Logging for harvesting actions - ONLY for player to avoid log spam
	if c.IsPlayerControlled && (c.State == ActorChopping || c.State == ActorDigging) && ctx.Log != nil {
		nearestMatchDist := 999.0
		nearestMatchName := "None"
		nx, ny := 0.0, 0.0
		
		matchType := "tree"
		if c.State == ActorDigging { matchType = "dig-spot" }

		totalObstacles := len(ctx.World.Obstacles)
		for _, o := range ctx.World.Obstacles {
			if o.Archetype == nil { continue }
			isTree := (o.Archetype.Type == TypeTree) || strings.Contains(strings.ToLower(o.ID), "tree") || strings.Contains(strings.ToLower(o.Archetype.Name), "tree")
			if (c.State == ActorChopping && isTree) || (c.State == ActorDigging && !isTree) {
				dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
				if dist < nearestMatchDist {
					nearestMatchDist = dist
					nearestMatchName = o.Archetype.Name
					nx, ny = o.X, o.Y
				}
			}
		}
		msg := fmt.Sprintf("[DEBUG] %s: Player(%.2f, %.2f) | Obstacles in World: %d | Nearest %s: %s at (%.2f, %.2f) dist=%.2f",
			c.State.String(), c.X, c.Y, totalObstacles, matchType, nearestMatchName, nx, ny, nearestMatchDist)
		ctx.Log(msg, LogNPC)
		log.Printf("%s", msg)
	}

	// Define the hit area. For harvesting, we use a generous 360-degree circle around the player.
	// For normal attacks, we use a directional area to reward proper positioning.
	var hitCircle engine.Circle
	var avgX, avgY float64
	if c.State == ActorChopping || c.State == ActorDigging || c.State == ActorForaging {
		hitCircle = engine.Circle{X: c.X, Y: c.Y, Radius: attackDist}
	} else {
		avgX, avgY = (c.X+atX)*0.5, (c.Y+atY)*0.5
		hitCircle = engine.Circle{X: avgX, Y: avgY, Radius: attackDist * 0.75}
	}

	targets := ctx.World.Characters
	if ctx.World.PlayableCharacter != nil {
		found := false
		for _, t := range targets { if t == ctx.World.PlayableCharacter { found = true; break } }
		if !found { targets = append([]*Character{ctx.World.PlayableCharacter}, targets...) }
	}

	hitSomething := false
	for _, target := range targets {
		if target == c || (target.Alignment == c.Alignment && !c.IsPlayerControlled) { continue }
		
		// Allow hitting dead characters ONLY if we are butchering (Chopping)
		if !target.IsAlive() && c.State != ActorChopping { continue }
		
		// Use directional check for characters unless harvesting
		checkX, checkY := avgX, avgY
		if c.State == ActorChopping || c.State == ActorDigging || c.State == ActorForaging { checkX, checkY = c.X, c.Y }
		
		if math.Sqrt(math.Pow(checkX-target.X, 2) + math.Pow(checkY-target.Y, 2)) < hitCircle.Radius { 
			c.hitCharacter(&target.Actor, skill, ctx); hitSomething = true 
		}
	}

	// Harvesting / Obstacle logic
	// If harvesting, we only want to hit the SINGLE NEAREST obstacle.
	var bestTarget *Obstacle
	bestDist := 999.0

	for _, o := range ctx.World.Obstacles {
		if !o.Alive || o.Archetype == nil || !o.Archetype.Destructible { continue }
		
		distToCenter := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
		inRange := engine.CheckCirclePolygonCollision(hitCircle, o.GetFootprint()) || distToCenter <= attackDist

		if !inRange { continue }
		
		isHarvesting := c.State == ActorChopping || c.State == ActorDigging || c.State == ActorForaging
		if isHarvesting {
			// Check if it's the right type for the action
			isTree := (o.Archetype.Type == TypeTree) || strings.Contains(strings.ToLower(o.ID), "tree") || strings.Contains(strings.ToLower(o.Archetype.Name), "tree")
			isBush := (o.Archetype.Type == TypeBush) || strings.Contains(strings.ToLower(o.ID), "bush")
			isCampfire := strings.Contains(strings.ToLower(o.ID), "campfire")
			isCrop := o.Archetype != nil && o.Archetype.IsCrop
			
			isCorrectType := (c.State == ActorChopping && (isTree || isCrop)) || (c.State == ActorDigging && !isTree && !isCampfire && !isCrop) || (c.State == ActorForaging && (isTree || isBush || isCampfire))
			
			if isCorrectType && distToCenter < bestDist {
				bestDist = distToCenter
				bestTarget = o
			}
		} else {
			// Normal attack: hit EVERYTHING in range
			power := c.rollDamage()
			if power > 0 {
				o.TakeDamage(power)
				c.DegradeWeapon(ctx)
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("-%d", power), X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
				hitSomething = true
			}
		}
	}

	// Apply harvesting to the single best target found
	if bestTarget != nil {
		o := bestTarget
		power := c.rollDamage()
		isTree := (o.Archetype.Type == TypeTree) || strings.Contains(strings.ToLower(o.ID), "tree") || strings.Contains(strings.ToLower(o.Archetype.Name), "tree")
		hasAxe := c.Weapon != nil && strings.Contains(strings.ToLower(c.Weapon.Name), "axe")

		if isTree && hasAxe && (c.State == ActorChopping || c.State == ActorAttacking) {
			// Harvesting logic based on total mass (weight)
			// Success check: Use "chop" ability (falls back to Strength)
			if !c.CheckAbilitySuccess("chop", 0) {
				if IsDebugEnabled() { DebugLog("%s FAILED to chop", c.Name) }
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Miss!", X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
				return
			}

			chopYield := c.GetAbilityYield("chop")
			chopPower := power + int(chopYield * 0.5)
			if c.State == ActorChopping { chopPower *= 5 }
			
			if ctx.Audio != nil { ctx.Audio.PlayRandomSound("footstep_wood") }
			
			// Record weight before depletion so we can still spawn drops if tree is felled in one hit
			weightBefore := o.WeightLeft

			// Account for loss (leaves, branches, sawdust) - random 0-50% of the force
			randLoss := rand.Float64() * (float64(chopPower) * 0.3) // Reduced loss for higher skill?
			o.WeightLeft -= (float64(chopPower) + randLoss)
			if o.WeightLeft < 0 { o.WeightLeft = 0 }
			
			// Spawning manageable logs (5-25 units each) from what was available
			if woodConfig := ctx.Registries.Objects.Objects["wood"]; woodConfig != nil && weightBefore > 0 {
				probability := float64(chopPower) / 80.0
				if c.State == ActorChopping { probability *= 2.0 }
				if rand.Float64() < probability || o.WeightLeft <= 30.0 {
					numLogs := 1 + int(chopYield * 0.02) // More logs for more skill
					if c.State == ActorChopping { numLogs += 1 }
					
					logPool := o.WeightLeft * 0.2 // Max yield per hit
					if o.WeightLeft <= 0 { logPool = weightBefore }
					
					var totalLogWeight float64
					for i := 0; i < numLogs && logPool > 0; i++ {
						w := 5.0 + rand.Float64()*15.0 + (chopYield * 0.1)
						if w > logPool { w = logPool }
						
						dropX := o.X + (rand.Float64()*1.2 - 0.6)
						dropY := o.Y + (rand.Float64()*1.2 - 0.6)
						
						item := NewItemInstance(fmt.Sprintf("wood_%d", rand.Int()), woodConfig, dropX, dropY)
						item.Weight = w
						ctx.World.Items = append(ctx.World.Items, item)
						
						logPool -= w
						totalLogWeight += w
					}
					
					if totalLogWeight > 0 {
						ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("+%.1f logs", totalLogWeight), X: o.X, Y: o.Y - 1, Life: 60, Color: ColorHeal })
					}
				}
			}
			
			c.DegradeWeapon(ctx)
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("%.1f rem", o.WeightLeft), X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
			
			if o.WeightLeft <= 0 {
				if stumpArch, ok := ctx.Registries.Obstacles.Archetypes["stump"]; ok {
					o.Archetype, o.HealthPoints, o.MaxHealthPoints, o.Alive = stumpArch, stumpArch.HealthPoints, stumpArch.HealthPoints, true
				} else { o.Alive = false }
				DebugLog("Tree Fell! Mass reservoir depleted.")
			}
			hitSomething = true
		} else if o.Archetype != nil && o.Archetype.IsCrop && c.State == ActorChopping {
			if o.GrowthStage < 2 {
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Not mature", X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
				return
			}
			// Harvest!
			o.Alive = false
			harvestYield := c.GetAbilityYield("harvest_crop")
			if yieldObj := ctx.Registries.Objects.Objects[o.Archetype.Yield]; yieldObj != nil {
				num := 1 + int(harvestYield * 0.03) // More and better crops
				for i := 0; i < num; i++ {
					dropX := o.X + (rand.Float64()*0.4 - 0.2)
					dropY := o.Y + (rand.Float64()*0.4 - 0.2)
					it := NewItemInstance(fmt.Sprintf("%s_%d", yieldObj.ID, rand.Int()), yieldObj, dropX, dropY)
					// Skill bonus quality
					it.Resistance = int(harvestYield * 0.1)
					ctx.World.Items = append(ctx.World.Items, it)
				}
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("+%.0f Harvest!", harvestYield), X: o.X, Y: o.Y - 1, Life: 60, Color: ColorHeal })
			}
			hitSomething = true
		} else if !isTree && c.State == ActorDigging {
			// Success check: Use "dig" ability (falls back to Strength)
			if !c.CheckAbilitySuccess("dig", 0) {
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Hard rock!", X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
				return
			}
			digYield := c.GetAbilityYield("dig")
			actualPower := power + int(digYield * 0.4)
			if actualPower > 0 {
				o.TakeDamage(actualPower)
				c.DegradeWeapon(ctx)
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("-%d", actualPower), X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
				hitSomething = true
			}
		} else if c.State == ActorForaging {
			isCampfire := strings.Contains(strings.ToLower(o.ID), "campfire")
			isTree := (o.Archetype.Type == TypeTree) || strings.Contains(strings.ToLower(o.ID), "tree")
			isBush := (o.Archetype.Type == TypeBush) || strings.Contains(strings.ToLower(o.ID), "bush")

			if isCampfire {
				// Cooking logic - handled in update loop / ProcessCooking now but keeping fallback
				foundRaw := -1
				for i, item := range c.Inventory {
					if item != nil && item.Config != nil && item.Config.ID == "raw_meat" { foundRaw = i; break }
				}
				if foundRaw >= 0 {
					// Success check: Use "cook" ability (falls back to Intellect)
					if !c.CheckAbilitySuccess("cook", 0) {
						ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "*Burned!*", X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHarm })
						c.Inventory = append(c.Inventory[:foundRaw], c.Inventory[foundRaw+1:]...) // Lose the meat if burned
						return
					}
					cookYield := c.GetAbilityYield("cook")
					if cookedConfig := ctx.Registries.Objects.Objects["meat"]; cookedConfig != nil {
						c.Inventory = append(c.Inventory[:foundRaw], c.Inventory[foundRaw+1:]...)
						item := NewItemInstance("meat", cookedConfig, c.X, c.Y)
						item.Resistance = int(cookYield * 0.1)
						c.Inventory = append(c.Inventory, item)
						ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "*Cooking...*", X: o.X, Y: o.Y - 1, Life: 60, Color: ColorHeal })
						ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "+Cooked Meat", X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHeal })
						if ctx.Audio != nil { ctx.Audio.PlayRandomSound("footstep_water") } // Placeholder for sizzling
						hitSomething = true
					}
				} else if ctx.Log != nil {
					ctx.Log(fmt.Sprintf("%s: I need raw meat to cook here!", c.Name), LogNPC)
				}
			} else if isTree || isBush {
				// Foraging logic with cooldown
				if o.CooldownTicks > 0 {
					return
				}

				forageYield := c.GetAbilityYield("forage")
				itemID := "wild_fruit"
				if o.Archetype != nil && o.Archetype.Yield != "" {
					itemID = o.Archetype.Yield
				} else if isBush {
					itemID = "wild_berries"
				}
				
				// Success check: Use "forage" ability (falls back to Wisdom)
				if c.CheckAbilitySuccess("forage", 0) {
					if rand.Float64() < 0.3 { itemID = "wild_veg" }
					
					if itemConfig := ctx.Registries.Objects.Objects[itemID]; itemConfig != nil {
						num := 1
						if forageYield > 50 { num++ }
						if forageYield > 80 && rand.Float64() < 0.5 { num++ }

						for i := 0; i < num; i++ {
							dropX := o.X + (rand.Float64()*1.2 - 0.6)
							dropY := o.Y + (rand.Float64()*1.2 - 0.6)
							item := NewItemInstance(itemID, itemConfig, dropX, dropY)
							item.Resistance = int(forageYield * 0.2) // Quality
							ctx.World.Items = append(ctx.World.Items, item)
						}
						
						ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("+%s x%d", itemConfig.Name, num), X: o.X, Y: o.Y - 1, Life: 60, Color: ColorHeal })
						if ctx.Audio != nil { ctx.Audio.PlayRandomSound("footstep_grass") }
						hitSomething = true
						o.CooldownTicks = 600 // 10s cooldown
					}
				}
			}
		}
	}

	if !hitSomething && c.State == ActorDigging && c.Weapon != nil && (strings.Contains(strings.ToLower(c.Weapon.Name), "pike") || strings.Contains(strings.ToLower(c.Weapon.Name), "pickaxe")) {
		// Ground digging success check: Use "dig" ability (falls back to Strength)
		if !c.CheckAbilitySuccess("dig", 0) {
			return
		}
		gridX, gridY := int(math.Floor(atX)), int(math.Floor(atY))
		if mapType := ctx.World.CurrentMapType; mapType != nil {
			if mapType.Heightmap == nil { mapType.Heightmap = make(map[string]float64) }
			key := fmt.Sprintf("%d,%d", gridX, gridY)
			newZ := mapType.GetElevationAt(float64(gridX), float64(gridY)) - 0.5
			mapType.Heightmap[key], c.Z = newZ, newZ
			c.DegradeWeapon(ctx)
			if ctx.Audio != nil { ctx.Audio.PlayRandomSound("footstep_stone") }
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "*dig*", X: atX, Y: atY, Life: 30, Color: color.RGBA{150, 100, 50, 255} })

			isCaveIn := false
			for dx := -1; dx <= 1; dx++ { for dy := -1; dy <= 1; dy++ { if dx == 0 && dy == 0 { continue }
					if math.Abs(mapType.GetElevationAt(float64(gridX+dx), float64(gridY+dy)) - newZ) >= 6.0 { isCaveIn = true } } }
			if isCaveIn {
				c.TemporalState.HealthPoints = 0; c.die(nil, ctx)
				if ctx.Audio != nil { ctx.Audio.PlayRandomSound("rock_crumble") }
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "CAVE-IN!", X: c.X, Y: c.Y, Life: 90, Color: ColorHarm })
			} else {
				var oreID string
				if mapType.MineralMap != nil {
					key := fmt.Sprintf("%d,%d", gridX, gridY)
					if mType, exists := mapType.MineralMap[key]; exists { oreID = mType; delete(mapType.MineralMap, key) }
				}
				if oreID == "" && rand.Float64() < 0.40 { oreID = "stone" }
				if oreID != "" && ctx.Registries != nil && ctx.Registries.Objects != nil {
					if oreConfig := ctx.Registries.Objects.Objects[oreID]; oreConfig != nil {
						it := NewItemInstance(oreID, oreConfig, float64(gridX), float64(gridY)); it.Z = newZ
						ctx.World.Items, ctx.World.FloatingTexts = append(ctx.World.Items, it), append(ctx.World.FloatingTexts, &FloatingText{ Text: "+Ore", X: atX, Y: atY, Life: 40, Color: color.RGBA{200, 200, 0, 255} })
					}
				}
			}
		}
	}
}

func (c *Character) TakeDamage(amount int, attacker ActorInterface, ctx *SystemContext) {
	c.Actor.TakeDamage(amount, attacker, ctx)
	if c.IsAlive() && !c.IsPlayerControlled { c.handleAIReaction(attacker, ctx) }
}

func (c *Character) handleAIReaction(attacker ActorInterface, ctx *SystemContext) {
	if attacker == nil { return }
	act := attacker.GetActor()
	c.TargetActor = act
	if float64(c.TemporalState.HealthPoints) < float64(act.TemporalState.HealthPoints)*0.2 { c.Alignment, c.Behavior = AlignmentNeutral, BehaviorFlee
	} else {
		c.Alignment, c.Behavior = AlignmentEnemy, BehaviorKnightHunter
		if c.Group != "" {
			for _, other := range ctx.World.Characters {
				if other == c || other.Alignment == AlignmentEnemy || !other.IsAlive() || other.Group != c.Group { continue }
				if math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)) < 20.0 { other.Alignment, other.Behavior, other.TargetActor = AlignmentEnemy, BehaviorKnightHunter, act }
			}
		}
	}
}

func (c *Character) MoveTo(ctx *SystemContext, tx, ty float64) {
	c.ExecutePathTo(ctx, tx, ty)
}

func (c *Character) ExecutePathTo(ctx *SystemContext, tx, ty float64) {
	dist := math.Sqrt(math.Pow(c.X-tx, 2) + math.Pow(c.Y-ty, 2))
	if dist < 0.5 {
		c.Path = nil
		c.State = ActorIdle
		return
	}

	// Recalculate path if needed
	if len(c.Path) == 0 || c.PathTimer <= 0 {
		c.Path = c.Actor.FindPath(tx, ty, ctx)
		c.PathTimer = 120 + rand.Intn(60) // Recalculate every 2-3 seconds
	} else {
		c.PathTimer--
	}

	if len(c.Path) > 0 {
		nextPoint := c.Path[0]
		pDist := math.Sqrt(math.Pow(c.X-nextPoint.X, 2) + math.Pow(c.Y-nextPoint.Y, 2))
		if pDist < 0.4 {
			c.Path = c.Path[1:]
			if len(c.Path) > 0 {
				nextPoint = c.Path[0]
			}
		}
		c.executeMovement(ctx, nextPoint.X-c.X, nextPoint.Y-c.Y, ctx.World.Obstacles, false)
	} else {
		// Fallback to direct movement if pathfinding failed or target is too far
		c.executeMovement(ctx, tx-c.X, ty-c.Y, ctx.World.Obstacles, false)
	}
}

func (c *Character) executeMovement(ctx *SystemContext, dx, dy float64, obstacles []*Obstacle, flee bool) {
	if c.State == ActorIncapacitated { return }
	mag := math.Sqrt(dx*dx + dy*dy)
	if mag < 0.01 { return }
	mvX, mvY := dx/mag, dy/mag
	if flee { mvX, mvY = -mvX, -mvY }
	speed := c.Speed * c.GetSpeedModifier(ctx)
	if !c.checkCollisionAt(c.X+mvX*speed, c.Y+mvY*speed, obstacles) {
		c.X, c.Y, c.State = c.X+mvX*speed, c.Y+mvY*speed, ActorWalking
		c.updateFacing(mvX, mvY)
	} else {
		c.State = ActorIdle
		// If we hit something and we were following a path, maybe the path is invalid
		if len(c.Path) > 0 {
			c.PathTimer = 0 // Force recalculate next tick
		}
	}
}

func (c *Character) executeAttack(ctx *SystemContext, isTargetPlayer bool, dx, dy float64) {
	if c.State == ActorIncapacitated { return }
	if c.State != ActorAttacking {
		if isTargetPlayer {
			// Use Wisdom for combat bark frequency/relevance
			if c.CheckAttributeSuccess("wisdom", 0) && c.Config != nil && c.Config.Dialogues != nil {
				if bark := c.Config.Dialogues.PickCombatBark(); bark != "" && ctx.Log != nil { ctx.Log(fmt.Sprintf("%s: %s", c.Name, bark), LogNPC) }
			}
			if rand.Float64() < 0.3 && ctx.Audio != nil && c.Config != nil { ctx.Audio.PlayRandomSound(c.Config.SoundID + "/attack") }
		}
		c.State, c.Tick = ActorAttacking, 0
	}
	if c.AttackTimer >= c.AttackCooldown {
		c.AttackTimer = 0
		
		// NPC Tactical Choice: Restrain or Torture
		skill := c.PendingSkill
		if !c.IsPlayerControlled && c.TargetActor != nil {
			if c.TargetActor.IsIncapacitated() {
				skill = "torture"
			} else if isTargetPlayer && rand.Float64() < 0.15 { // 15% chance to attempt restraint
				skill = "restrain"
			}
		}

		if c.Weapon != nil && c.Weapon.IsRanged() && skill == "" {
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag > 0 {
				pSpd := c.RawStats.ProjectileSpeed
				if pSpd <= 0 { pSpd = 0.5 }
				ctx.World.Projectiles = append(ctx.World.Projectiles, NewProjectile(c.X, c.Y, dx/mag, dy/mag, pSpd, c.GetTotalAttack(), false, 100.0))
			}
		} else { 
			c.CheckAttackHits(ctx, skill) 
		}
	}
}
func (c *Character) Rest(ctx *SystemContext) {
	// Success check: Use "rest" ability (falls back to Health)
	if !c.CheckAbilitySuccess("rest", 0) {
		return
	}
	if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s]: is resting", c.Name), LogNPC) }
	c.TemporalState.Fatigue += 10.0
	if c.TemporalState.Fatigue > 100 { c.TemporalState.Fatigue = 100 }
	if ctx.World != nil {
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Resting...", X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHeal })
	}
}

func (c *Character) Milk(target *Actor, ctx *SystemContext) {
	if target == nil || !target.Config.Stats.IsMilkable { return }
	if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s]: is milking %s", c.Name, target.Name), LogNPC) }
	if target.MilkCooldownTicks > 0 {
		if c.IsPlayerControlled && ctx.Log != nil { ctx.Log("This animal needs time to recover.", LogNPC) }
		return 
	}

	// Success check: Use "milk" ability (falls back to Dexterity)
	if !c.CheckAbilitySuccess("milk", 0) {
		if ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Spilled!", X: target.X, Y: target.Y - 1, Life: 60, Color: ColorHarm })
		}
		target.MilkCooldownTicks = 300 // Short cooldown for fail
		return
	}

	if milkConfig := ctx.Registries.Objects.Objects["milk"]; milkConfig != nil {
		it := NewItemInstance("milk", milkConfig, target.X, target.Y)
		ctx.World.Items = append(ctx.World.Items, it)
		target.MilkCooldownTicks = target.RawStats.MilkCooldown
		if ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "+Milk", X: target.X, Y: target.Y - 1, Life: 60, Color: ColorHeal })
		}
	}
}
