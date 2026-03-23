package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strings"
)

func (c *Character) CheckAttackHits(ctx *SystemContext) {
	attackDist := 2.5
	if c.Config != nil && c.Config.Stats.AttackRange > 0 { attackDist = c.Config.Stats.AttackRange }
	if c.Weapon != nil { attackDist = c.Weapon.GetMaxDistance() }
	atX, atY := c.X, c.Y
	switch c.Facing {
	case DirSE: atX += attackDist * 0.7; atY += attackDist * 0.35
	case DirSW: atX -= attackDist * 0.35; atY += attackDist * 0.7
	case DirNE: atX += attackDist * 0.7; atY -= attackDist * 0.35
	case DirNW: atX -= attackDist * 0.35; atY -= attackDist * 0.7
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
		if math.Sqrt(math.Pow(atX-target.X, 2) + math.Pow(atY-target.Y, 2)) < attackDist*1.2 { c.hitCharacter(&target.Actor, ctx); hitSomething = true }
	}
	
	for _, o := range ctx.World.Obstacles {
		if !o.Alive || o.Archetype == nil || !o.Archetype.Destructible { continue }
		if math.Sqrt(math.Pow(atX-o.X, 2) + math.Pow(atY-o.Y, 2)) < attackDist*1.5 {
			power := c.rollDamage()
			if strings.Contains(strings.ToLower(string(o.Archetype.Type)), "tree") {
				if c.State == ActorChopping && c.Weapon != nil && strings.Contains(strings.ToLower(c.Weapon.Name), "axe") {
					chopPower := power * 5
					if ctx.Audio != nil { ctx.Audio.PlayRandomSound("footstep_wood") }
					if woodConfig := ctx.Registries.Objects.Objects["wood"]; woodConfig != nil {
						dropX, dropY := o.X + (rand.Float64()*1.0 - 0.5), o.Y + (rand.Float64()*1.0 - 0.5)
						ctx.World.Items, ctx.World.FloatingTexts = append(ctx.World.Items, NewItemInstance("wood", woodConfig, dropX, dropY)), append(ctx.World.FloatingTexts, &FloatingText{ Text: "+Wood", X: dropX, Y: dropY, Life: 40, Color: color.RGBA{139, 69, 19, 255} })
					}
					o.ReduceTimber(chopPower)
					c.DegradeWeapon(ctx)
					ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("-%d", chopPower), X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
					if o.TimberLeft() <= 0 { if stumpArch, ok := ctx.Registries.Obstacles.Archetypes["stump"]; ok { o.Archetype, o.Health, o.MaxHealth, o.Alive = stumpArch, stumpArch.Health, stumpArch.Health, true } }
					hitSomething = true
				}
			} else if power > 0 {
				o.TakeDamage(power); c.DegradeWeapon(ctx); ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("-%d", power), X: o.X, Y: o.Y, Life: 45, Color: ColorHarm })
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
