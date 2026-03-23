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

func (c *Character) CheckAttackHits(ctx *SystemContext) {
	attackDist := 2.5
	if c.Config != nil && c.Config.Stats.AttackRange > 0 { attackDist = c.Config.Stats.AttackRange }
	if c.Weapon != nil { attackDist = c.Weapon.GetMaxDistance() }
	if c.State == ActorChopping || c.State == ActorDigging {
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
	if c.State == ActorChopping || c.State == ActorDigging {
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
		if target == c || !target.IsAlive() || (target.Alignment == c.Alignment && !c.IsPlayerControlled) { continue }
		
		// Use directional check for characters unless harvesting
		checkX, checkY := avgX, avgY
		if c.State == ActorChopping || c.State == ActorDigging { checkX, checkY = c.X, c.Y }
		
		if math.Sqrt(math.Pow(checkX-target.X, 2) + math.Pow(checkY-target.Y, 2)) < hitCircle.Radius { 
			c.hitCharacter(&target.Actor, ctx); hitSomething = true 
		}
	}

	// Harvesting / Obstacle logic
	// If harvesting, we only want to hit the SINGLE NEAREST obstacle.
	var bestTarget *Obstacle
	bestDist := 999.0

	for _, o := range ctx.World.Obstacles {
		if !o.Alive || o.Archetype == nil || !o.Archetype.Destructible { continue }
		
		distToCenter := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
		inRange := engine.CheckCirclePolygonCollision(hitCircle, o.GetFootprint()) || distToCenter <= 3.0

		if !inRange { continue }

		if c.State == ActorChopping || c.State == ActorDigging {
			// Check if it's the right type for the action
			isTree := (o.Archetype.Type == TypeTree) || strings.Contains(strings.ToLower(o.ID), "tree") || strings.Contains(strings.ToLower(o.Archetype.Name), "tree")
			isCorrectType := (c.State == ActorChopping && isTree) || (c.State == ActorDigging && !isTree)
			
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
			chopPower := power
			if c.State == ActorChopping { chopPower *= 5 }
			
			if ctx.Audio != nil { ctx.Audio.PlayRandomSound("footstep_wood") }
			
			// Account for loss (leaves, branches, sawdust) - random 0-50% of the force
			randLoss := rand.Float64() * (float64(chopPower) * 0.5)
			o.WeightLeft -= (float64(chopPower) + randLoss)
			if o.WeightLeft < 0 { o.WeightLeft = 0 }
			
			// Spawning manageable logs (5-25 units each)
			if woodConfig := ctx.Registries.Objects.Objects["wood"]; woodConfig != nil && o.WeightLeft > 0 {
				probability := float64(chopPower) / 100.0
				if rand.Float64() < probability || o.WeightLeft <= 30.0 {
					numLogs := 1 + rand.Intn(3)
					if c.State == ActorChopping { numLogs += 1 }
					
					var totalLogWeight float64
					for i := 0; i < numLogs && o.WeightLeft > 0; i++ {
						w := 5.0 + rand.Float64()*20.0
						if w > o.WeightLeft { w = o.WeightLeft }
						
						dropX := o.X + (rand.Float64()*1.2 - 0.6)
						dropY := o.Y + (rand.Float64()*1.2 - 0.6)
						
						item := NewItemInstance(fmt.Sprintf("wood_%d", rand.Int()), woodConfig, dropX, dropY)
						item.Weight = w
						ctx.World.Items = append(ctx.World.Items, item)
						
						o.WeightLeft -= w
						totalLogWeight += w
					}
					
					if totalLogWeight > 0 {
						ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("+%.1f mass", totalLogWeight), X: o.X, Y: o.Y - 1, Life: 60, Color: ColorHeal })
					}
				}
			}
			
			c.DegradeWeapon(ctx)
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("%.1f mass rem", o.WeightLeft), X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
			
			if o.WeightLeft <= 0 {
				if stumpArch, ok := ctx.Registries.Obstacles.Archetypes["stump"]; ok {
					o.Archetype, o.Health, o.MaxHealth, o.Alive = stumpArch, stumpArch.Health, stumpArch.Health, true
				} else { o.Alive = false }
				DebugLog("Tree Fell! Mass reservoir depleted.")
			}
			hitSomething = true
		} else if !isTree && c.State == ActorDigging {
			// This part is handled by the digging logic below if no obstacle is hit, 
			// but if an obstacle IS hit while digging (and it's not a tree), we handle it here if it's destructive.
			// Currently most stones are Obstacles.
			if power > 0 {
				o.TakeDamage(power)
				c.DegradeWeapon(ctx)
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("-%d", power), X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
				hitSomething = true
			}
		}
	}

	if !hitSomething && c.State == ActorDigging && c.Weapon != nil && (strings.Contains(strings.ToLower(c.Weapon.Name), "pike") || strings.Contains(strings.ToLower(c.Weapon.Name), "pickaxe")) {
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
				c.Health = 0; c.die(nil, ctx)
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
	if float64(c.Health) < float64(act.Health)*0.2 { c.Alignment, c.Behavior = AlignmentNeutral, BehaviorFlee
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

func (c *Character) executeMovement(ctx *SystemContext, dx, dy float64, obstacles []*Obstacle, flee bool) {
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
	}
}

func (c *Character) executeAttack(ctx *SystemContext, isTargetPlayer bool, dx, dy float64) {
	if c.State != ActorAttacking {
		if isTargetPlayer {
			if rand.Float64() < 0.1 && c.Config != nil && c.Config.Dialogues != nil {
				if bark := c.Config.Dialogues.PickCombatBark(); bark != "" && ctx.Log != nil { ctx.Log(fmt.Sprintf("%s: %s", c.Name, bark), LogNPC) }
			}
			if rand.Float64() < 0.3 && ctx.Audio != nil && c.Config != nil { ctx.Audio.PlayRandomSound(c.Config.SoundID + "/attack") }
		}
		c.State, c.Tick = ActorAttacking, 0
	}
	if c.AttackTimer >= c.AttackCooldown {
		c.AttackTimer = 0
		if c.Weapon != nil && c.Weapon.IsRanged() {
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag > 0 {
				pSpd := c.Config.Stats.ProjectileSpeed
				if pSpd <= 0 { pSpd = 0.5 }
				ctx.World.Projectiles = append(ctx.World.Projectiles, NewProjectile(c.X, c.Y, dx/mag, dy/mag, pSpd, c.GetTotalAttack(), false, 100.0))
			}
		} else { c.CheckAttackHits(ctx) }
	}
}
