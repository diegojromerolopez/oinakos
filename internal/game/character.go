package game

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"oinakos/internal/engine"
)

// Character replaces the old NPC and PlayableCharacter structs.
type Character struct {
	Actor // Embedded shared state

	// AI and Behavior (inherited from NPC)
	AttackCooldown int
	AttackTimer    int

	Behavior   BehaviorType
	WanderDirX float64
	WanderDirY float64
	PatrolStartX, PatrolStartY float64
	PatrolEndX, PatrolEndY     float64
	PatrolHeading              bool
	TargetActor *Actor
	HasInitiatedDialogue bool
	AIDecisionPending    bool
	LastAIChoice         string
	LastAIReasoning      string
	TargetActorForAI     *Actor
	TargetItem           *ItemInstance

	// Control state
	IsPlayerControlled bool
}

var characterNames = []string{
	"Grog", "Zog", "Bob", "Drok", "Gorak", "Mug", "Snarl", "Thrak", "Vrog", "Kurg",
	"Hicks", "Miller", "Cooper", "Smith", "Potter", "Baker", "Carter", "Fisher",
}

func NewCharacter(x, y float64, config *EntityConfig, level int, isPlayer bool) *Character {
	if config == nil {
		config = &EntityConfig{ID: "unknown", Name: "Unknown Entity"}
		config.Stats.HealthMin = 10
		config.Stats.HealthMax = 10
		config.Weapon = WeaponConfig{Inline: WeaponTizon}
	}
	
	c := &Character{
		Actor: Actor{
			X: x, Y: y, Config: config, State: ActorIdle, Facing: DirSE, Level: level,
			Alignment: AlignmentEnemy, Group: config.Group, LeaderID: config.LeaderID, MustSurvive: config.MustSurvive,
			Name: config.Name,
		},
		IsPlayerControlled: isPlayer,
	}

	if isPlayer {
		c.Alignment = AlignmentAlly
	}

	// Character-specific initialization (if not player)
	if !isPlayer {
		if config.Unique {
			c.Name = config.Name
		} else if len(config.Names) > 0 {
			c.Name = config.Names[rand.Intn(len(config.Names))]
		} else if config.Name != "" {
			c.Name = config.Name
		} else {
			c.Name = characterNames[rand.Intn(len(characterNames))]
		}

		switch config.Behavior {
		case "chaotic": c.Behavior = BehaviorChaotic
		case "fighter": c.Behavior = BehaviorNpcFighter
		case "hunter":  c.Behavior = BehaviorKnightHunter
		case "wander":  c.Behavior = BehaviorWander
		case "patrol":  c.Behavior = BehaviorPatrol
		case "escort":  c.Behavior = BehaviorEscort
		case "flee":    c.Behavior = BehaviorFlee
		default:
			if config.Unique { c.Behavior = BehaviorWander } else { c.Behavior = BehaviorKnightHunter }
		}

		if c.Behavior == BehaviorWander {
			c.WanderDirX = rand.Float64()*2 - 1
			c.WanderDirY = rand.Float64()*2 - 1
		} else if c.Behavior == BehaviorPatrol {
			c.PatrolStartX = c.X
			c.PatrolStartY = c.Y
			c.PatrolEndX = c.X + (rand.Float64()*8 - 4)
			c.PatrolEndY = c.Y + (rand.Float64()*8 - 4)
			c.PatrolHeading = true
		}
	}

	c.Health = config.Stats.HealthMin + rand.Intn(config.Stats.HealthMax-config.Stats.HealthMin+1)
	c.BaseAttack = config.Stats.BaseAttack
	c.BaseDefense = config.Stats.BaseDefense
	c.Speed = config.Stats.Speed
	c.MaxWeight = config.MaxWeight
	c.Slots = make(map[string]*ObjectConfig)
	c.Inventory = make([]*ObjectConfig, 0)
	c.AttackCooldown = config.Stats.AttackCooldown
	c.Weapon = config.Weapon.Resolve(nil) 

	c.Health = c.calculateStat(c.Health, c.Level)
	c.MaxHealth = c.Health
	c.BaseAttack = c.calculateStat(c.BaseAttack, c.Level)
	c.BaseDefense = c.calculateStat(c.BaseDefense, c.Level)
	c.MapKills = make(map[string]int)

	return c
}

func (c *Character) Update(ctx *SystemContext) {
	c.Tick++
	c.SharedUpdate(ctx)
	
	if c.IsPlayerControlled {
		c.updatePlayer(ctx)
	} else {
		c.updateAI(ctx)
	}
}

func (c *Character) Draw(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64) {
	DrawActor(&c.Actor, screen, textRenderer, vectorRenderer, paletteShader, offsetX, offsetY, c.IsPlayerControlled)
}

func (c *Character) DrawUI(game *Game, screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64, debug bool) {
	DrawActorUI(game, &c.Actor, screen, textRenderer, vectorRenderer, offsetX, offsetY, c.IsPlayerControlled, debug)
}

func (c *Character) updatePlayer(ctx *SystemContext) {
	worldObstacles := ctx.World.Obstacles
	mapW, mapH := ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight

	if c.State == ActorCrouching { return }

	if c.State == ActorDead {
		if c.DeadTimer == 0 {
			newX, newY := findSafePosition(c.X, c.Y, c.GetCollisionCircle(), worldObstacles)
			c.X, c.Y = newX, newY
		}
		c.DeadTimer++
		return
	}

	if c.HitTimer > 0 { c.HitTimer-- }

	if c.State == ActorAttacking {
		if c.Tick >= 30 { c.State = ActorIdle }
		return
	}

	if c.State == ActorDrinking {
		if c.Tick >= 60 { c.State = ActorIdle }
		return
	}

	var dx, dy float64
	simMode := false
	if ctx.World != nil && ctx.World.Game != nil && ctx.World.Game.settings != nil {
		simMode = ctx.World.Game.settings.AISimulationMode
	}

	if c.IsPlayerControlled && !simMode {
		if ctx.Input != nil {
			if ctx.Input.IsKeyPressed(engine.KeyW) || ctx.Input.IsKeyPressed(engine.KeyUp) { dy -= 1 }
			if ctx.Input.IsKeyPressed(engine.KeyS) || ctx.Input.IsKeyPressed(engine.KeyDown) { dy += 1 }
			if ctx.Input.IsKeyPressed(engine.KeyA) || ctx.Input.IsKeyPressed(engine.KeyLeft) { dx -= 1 }
			if ctx.Input.IsKeyPressed(engine.KeyD) || ctx.Input.IsKeyPressed(engine.KeyRight) { dx += 1 }

			if ctx.Input.IsKeyPressed(engine.KeySpace) {
				for _, o := range worldObstacles {
					if o.Alive && o.Archetype != nil && o.CooldownTicks <= 0 {
						for _, action := range o.Archetype.Actions {
							if action.Type == ActionHeal && action.RequiresInteraction {
								dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
								if dist < 1.5 {
									if action.Amount >= 999 { c.Health = c.MaxHealth } else { c.Heal(action.Amount) }
									o.CooldownTicks = int(o.Archetype.CooldownTime * 60 * 60)
									ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
										Text: fmt.Sprintf("+%d", action.Amount), X: c.X, Y: c.Y, Life: 45, Color: ColorHeal,
									})
									c.State = ActorDrinking
									c.Tick = 0
									return
								}
							}
						}
					}
				}

				c.State = ActorAttacking
				c.Tick = 0
				if ctx.Audio != nil && c.Config != nil {
					prefix := c.Config.PlayableCharacter
					if prefix == "" { prefix = c.Config.ID }
					ctx.Audio.PlayRandomSound(prefix + "/attack")
				}
				return
			}
		}

		if dx != 0 || dy != 0 {
			mag := math.Sqrt(dx*dx + dy*dy)
			dx /= mag
			dy /= mag
			moveX := dx * c.Speed * c.GetSpeedModifier()
			moveY := dy * c.Speed * c.GetSpeedModifier()

			if !c.checkCollisionAt(c.X+moveX, c.Y+moveY, worldObstacles) {
				c.X += moveX
				c.Y += moveY
				c.State = ActorWalking
				if c.Tick%30 == 0 && ctx.Audio != nil {
					sound := "footstep_grass"
					if c.CurrentTile == "water.png" || c.CurrentTile == "dark_water.png" { sound = "footstep_water" }
					if c.CurrentTile == "paved_ground.png" || c.CurrentTile == "big_stones.png" { sound = "footstep_stone" }
					ctx.Audio.PlayRandomSound(sound)
				}
			} else {
				c.State = ActorIdle; c.Tick = 0
			}
			c.updateFacing(dx, dy)
		} else {
			c.State = ActorIdle; c.Tick = 0
		}
		c.clampToMap(mapW, mapH)
	} else {
		c.updateAI(ctx)
	}
}

func (c *Character) updateAI(ctx *SystemContext) {
	worldObstacles := ctx.World.Obstacles
	mapW, mapH := ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight

	if c.HitTimer > 0 { c.HitTimer-- }
	var playerDist float64
	playableCharacter := ctx.World.PlayableCharacter
	if playableCharacter != nil {
		playerDist = math.Sqrt(math.Pow(c.X-playableCharacter.X, 2) + math.Pow(c.Y-playableCharacter.Y, 2))
	}

	if c.LeaderID != "" {
		leaderAlive := false
		for _, char := range ctx.World.Characters {
			if char.Config != nil && char.Config.ID == c.LeaderID && char.IsAlive() {
				leaderAlive = true
				break
			}
		}
		if !leaderAlive {
			c.Alignment = AlignmentNeutral
			c.Behavior = BehaviorWander
			c.LeaderID = "" // Clear it once transition is done
		}
	}

	if c.State == ActorCrouching || c.State == ActorDead {
		if c.State == ActorDead {
			if c.DeadTimer == 0 { c.X, c.Y = findSafePosition(c.X, c.Y, c.GetCollisionCircle(), worldObstacles) }
			c.DeadTimer++
		}
		return
	}

	if c.State == ActorAttacking {
		c.AttackTimer++
		if c.Tick >= 30 { c.State = ActorIdle }
		return
	}

	if ctx.AIManager != nil && !c.AIDecisionPending && c.needsAIDecision(playerDist) {
		worldCtx := BuildWorldContext(ctx.World.Game, c)
		options := []string{"attack", "flee", "wander", "patrol"}
		ctx.AIManager.RequestDecision(context.Background(), c.Config.ID, worldCtx, options)
		c.AIDecisionPending = true
	}

	if c.Tick%30 == 0 { c.TargetItem = c.findLootTarget(ctx.World.Items) }

	if c.TargetItem != nil && c.TargetItem.Pickable {
		dx := c.TargetItem.X - c.X
		dy := c.TargetItem.Y - c.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 1.5 {
			if ctx.World.Game != nil && ctx.World.Game.TryPickup(&c.Actor, c.TargetItem) {
				c.EquipItem(c.TargetItem.Config)
				itList := []*ItemInstance{}
				for _, it := range ctx.World.Items { if it != c.TargetItem { itList = append(itList, it) } }
				ctx.World.Items = itList
				c.TargetItem = nil
				return
			}
			c.TargetItem = nil
			c.State = ActorIdle
		} else {
			c.executeMovement(dx, dy, worldObstacles, false)
			return
		}
	}

	targetX, targetY, hasTarget, isTargetPlayer := c.findTarget(playableCharacter, ctx.World.Characters, playerDist)
	if !hasTarget {
		if c.Behavior == BehaviorWander {
			c.updateWander(worldObstacles)
		} else if c.Behavior == BehaviorPatrol {
			c.updatePatrol(worldObstacles)
		} else if c.Alignment == AlignmentAlly && playableCharacter != nil && playerDist > 5.0 && playerDist < 20.0 {
			c.executeMovement(playableCharacter.X-c.X, playableCharacter.Y-c.Y, worldObstacles, false)
		} else {
			c.State = ActorIdle
		}
		c.clampToMap(mapW, mapH)
		return
	}

	dx := targetX - c.X
	dy := targetY - c.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	attackRange := 1.4
	if c.Config != nil && c.Config.Stats.AttackRange > 0 { attackRange = c.Config.Stats.AttackRange }
	if c.Weapon != nil { attackRange = c.Weapon.GetMaxDistance() }

	canAttack := (c.TargetActor != nil && c.Alignment != c.TargetActor.Alignment)
	if c.Behavior == BehaviorChaotic { canAttack = true }
	
	tooClose := dist < attackRange*0.5 && c.Weapon != nil && c.Weapon.IsRanged()
	if tooClose && canAttack {
		c.executeMovement(dx, dy, worldObstacles, true) // Flee if too close
	} else if dist < attackRange && canAttack {
		c.executeAttack(ctx, isTargetPlayer, dx, dy)
	} else {
		flee := (c.Behavior == BehaviorFlee)
		c.executeMovement(dx, dy, worldObstacles, flee)
	}
	c.clampToMap(mapW, mapH)
}

func (c *Character) updateWander(obstacles []*Obstacle) {
	if c.Tick%120 == 0 || (c.WanderDirX == 0 && c.WanderDirY == 0) {
		angle := rand.Float64() * 2 * math.Pi
		c.WanderDirX = math.Cos(angle)
		c.WanderDirY = math.Sin(angle)
	}
	c.executeMovement(c.WanderDirX, c.WanderDirY, obstacles, false)
}

func (c *Character) updatePatrol(obstacles []*Obstacle) {
	targetX, targetY := c.PatrolEndX, c.PatrolEndY
	if !c.PatrolHeading {
		targetX, targetY = c.PatrolStartX, c.PatrolStartY
	}

	dist := math.Sqrt(math.Pow(c.X-targetX, 2) + math.Pow(c.Y-targetY, 2))
	if dist < 0.5 {
		c.PatrolHeading = !c.PatrolHeading
	} else {
		c.executeMovement(targetX-c.X, targetY-c.Y, obstacles, false)
	}
}

func (c *Character) updateFacing(dx, dy float64) {
	if dx > 0 {
		if dy < 0 { c.Facing = DirNE } else if dy > 0 { c.Facing = DirSE } else { c.Facing = DirSE }
	} else if dx < 0 {
		if dy < 0 { c.Facing = DirNW } else if dy > 0 { c.Facing = DirSW } else { c.Facing = DirSW }
	} else {
		if dy < 0 { c.Facing = DirNE } else if dy > 0 { c.Facing = DirSW }
	}
}

func (c *Character) clampToMap(mapW, mapH float64) {
	halfW, halfH := mapW/2, mapH/2
	if c.X < -halfW { c.X = -halfW }
	if c.X > halfW { c.X = halfW }
	if c.Y < -halfH { c.Y = -halfH }
	if c.Y > halfH { c.Y = halfH }
}

func (c *Character) CheckAttackHits(ctx *SystemContext) {
	attackDist := 1.4
	if c.Config != nil && c.Config.Stats.AttackRange > 0 { attackDist = c.Config.Stats.AttackRange }
	if c.Weapon != nil { attackDist = c.Weapon.GetMaxDistance() }
	
	atX, atY := c.X, c.Y
	// Precise hitbox offset based on facing
	switch c.Facing {
	case DirSE: atX += attackDist * 0.7; atY += attackDist * 0.35
	case DirSW: atX -= attackDist * 0.35; atY += attackDist * 0.7
	case DirNE: atX += attackDist * 0.7; atY -= attackDist * 0.35
	case DirNW: atX -= attackDist * 0.35; atY -= attackDist * 0.7
	}
	
	targets := ctx.World.Characters
	if ctx.World.PlayableCharacter != nil {
		found := false
		for _, t := range targets {
			if t == ctx.World.PlayableCharacter { found = true; break }
		}
		if !found {
			targets = append([]*Character{ctx.World.PlayableCharacter}, targets...)
		}
	}

	for _, target := range targets {
		if target == c || !target.IsAlive() { continue }
		if target.Alignment == c.Alignment && !c.IsPlayerControlled { continue } 
		dist := math.Sqrt(math.Pow(atX-target.X, 2) + math.Pow(atY-target.Y, 2))
		if dist < attackDist*1.2 { c.hitCharacter(target, ctx) }
	}
	
	for _, o := range ctx.World.Obstacles {
		if !o.Alive || o.Archetype == nil || !o.Archetype.Destructible { continue }
		dist := math.Sqrt(math.Pow(atX-o.X, 2) + math.Pow(atY-o.Y, 2))
		if dist < attackDist*1.5 {
			rawDmg := c.rollDamage()
			o.TakeDamage(rawDmg)
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: fmt.Sprintf("-%d", rawDmg), X: o.X, Y: o.Y, Life: 45, Color: ColorHarm,
			})
		}
	}
}

func (c *Character) hitCharacter(target *Character, ctx *SystemContext) {
	attk, def := float64(c.GetTotalAttack()), float64(target.GetTotalDefense())
	if def <= 0 { def = 1 }
	hitChance := clampInt(int(attk/(attk+def)*100), 5, 95)

	if rand.Intn(100)+1 <= hitChance {
		rawDmg := c.rollDamage()
		finalDmg := int(math.Max(1, float64(rawDmg-target.GetTotalProtection())))
		target.TakeDamage(finalDmg, c, ctx)
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: fmt.Sprintf("-%d", finalDmg), X: target.X, Y: target.Y, Life: 45, Color: ColorHarm,
		})
	} else {
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: "MISS", X: target.X, Y: target.Y, Life: 45, Color: ColorMiss,
		})
	}
}

func (c *Character) rollDamage() int {
	if c.Weapon != nil { return c.Weapon.RollDamage() }
	return c.BaseAttack + WeaponFists.RollDamage()
}

func (c *Character) TakeDamage(amount int, attacker ActorInterface, ctx *SystemContext) {
	if c.State == ActorDead { return }
	c.Health -= amount; c.HitTimer = 30
	prefix := "unknown"
	if c.Config != nil { prefix = c.Config.ID }
	if c.IsPlayerControlled && c.Config != nil && c.Config.PlayableCharacter != "" { prefix = c.Config.PlayableCharacter }
	if ctx.Audio != nil { ctx.Audio.PlayRandomSound(prefix + "/hit") }

	if c.Health <= 0 {
		c.Health = 0
		c.die(attacker, ctx)
	} else if !c.IsPlayerControlled {
		c.handleAIReaction(attacker, ctx)
	}
}

func (c *Character) die(attacker ActorInterface, ctx *SystemContext) {
	c.State = ActorDead
	if ctx.World != nil && ctx.World.Game != nil { ctx.World.Game.DropAllItems(&c.Actor) }
	prefix := "unknown"
	if c.Config != nil { prefix = c.Config.ID }
	if c.IsPlayerControlled && c.Config != nil && c.Config.PlayableCharacter != "" { prefix = c.Config.PlayableCharacter }
	if ctx.Audio != nil { ctx.Audio.PlayRandomSound(prefix + "/death") }
	if attacker != nil {
		if act := attacker.GetActor(); act != nil {
			act.Kills++
			if c.Config != nil {
				act.MapKills[c.Config.ID]++
				xp := c.Config.XP
				if xp <= 0 { xp = 1 }
				act.AddXP(xp)
			}
			
			// Process OnKill actions from the attacker's config
			if act.Config != nil && act.Config.Actions != nil {
				for _, action := range act.Config.Actions.OnKill {
					if rand.Float64() < action.Probability {
						c.applyKillAction(action, attacker, ctx)
					}
				}
			}
		}
	}
}

func (c *Character) applyKillAction(action KillAction, attacker ActorInterface, ctx *SystemContext) {
	if action.Type == "transform_victim" {
		e := action.Effect.Victim
		if e == nil { return }
		
		targetID := e.Transform
		// Replace {gender} if present
		targetID = strings.ReplaceAll(targetID, "{gender}", c.Config.Gender)
		
		var newConfig *EntityConfig
		var ok bool
		if ctx.Registries != nil {
			if ctx.Registries.Archetypes != nil {
				newConfig, ok = ctx.Registries.Archetypes.Archetypes[targetID]
			}
			if !ok && ctx.Registries.Characters != nil {
				newConfig, ok = ctx.Registries.Characters.Characters[targetID]
			}
		}
		
		if ok {
			c.Config = newConfig
			c.Health = newConfig.Stats.HealthMax
			c.MaxHealth = newConfig.Stats.HealthMax
			c.State = ActorIdle
			if e.Alignment == "inherit" {
				c.Alignment = attacker.GetActor().Alignment
			}
		}
	}
	
	if action.Type == "heal_attacker" || action.Effect.Attacker != nil {
		if action.Effect.Attacker != nil && action.Effect.Attacker.Heal > 0 {
			attacker.GetActor().Health += action.Effect.Attacker.Heal
			if attacker.GetActor().Health > attacker.GetActor().MaxHealth {
				attacker.GetActor().Health = attacker.GetActor().MaxHealth
			}
		}
	}
}

func (c *Character) handleAIReaction(attacker ActorInterface, ctx *SystemContext) {
	if attacker == nil { return }
	act := attacker.GetActor()
	c.TargetActor = act
	if float64(c.Health) < float64(act.Health)*0.2 {
		c.Alignment = AlignmentNeutral; c.Behavior = BehaviorFlee
	} else {
		c.Alignment = AlignmentEnemy; c.Behavior = BehaviorKnightHunter
		if c.Group != "" {
			for _, other := range ctx.World.Characters {
				if other == c || other.Alignment == AlignmentEnemy || !other.IsAlive() || other.Group != c.Group { continue }
				if math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)) < 20.0 {
					other.Alignment = AlignmentEnemy; other.Behavior = BehaviorKnightHunter; other.TargetActor = act
				}
			}
		}
	}
}

func (c *Character) executeMovement(dx, dy float64, obstacles []*Obstacle, flee bool) {
	mag := math.Sqrt(dx*dx + dy*dy)
	if mag < 0.01 { return }
	mvX, mvY := dx/mag, dy/mag
	if flee { mvX, mvY = -mvX, -mvY }
	speed := c.Speed * c.GetSpeedModifier()
	if !c.checkCollisionAt(c.X+mvX*speed, c.Y+mvY*speed, obstacles) {
		c.X += mvX * speed; c.Y += mvY * speed; c.State = ActorWalking; c.updateFacing(mvX, mvY)
	} else {
		c.State = ActorIdle
	}
}

func (c *Character) executeAttack(ctx *SystemContext, isTargetPlayer bool, dx, dy float64) {
	if c.State != ActorAttacking {
		if isTargetPlayer {
			if rand.Float64() < 0.1 && c.Config != nil && c.Config.Dialogues != nil {
				bark := c.Config.Dialogues.PickCombatBark()
				if bark != "" && ctx.Log != nil { ctx.Log(fmt.Sprintf("%s: %s", c.Name, bark), LogNPC) }
			}
			if rand.Float64() < 0.3 && ctx.Audio != nil && c.Config != nil {
				ctx.Audio.PlayRandomSound(c.Config.SoundID + "/attack")
			}
		}
		c.State = ActorAttacking
		c.Tick = 0
	}
	
	if c.AttackTimer >= c.AttackCooldown {
		c.AttackTimer = 0
		if c.Weapon != nil && c.Weapon.IsRanged() {
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag > 0 {
				pSpd := c.Config.Stats.ProjectileSpeed
				if pSpd <= 0 { pSpd = 0.5 }
				proj := NewProjectile(c.X, c.Y, dx/mag, dy/mag, pSpd, c.GetTotalAttack(), false, 100.0)
				ctx.World.Projectiles = append(ctx.World.Projectiles, proj)
			}
		} else { 
			c.CheckAttackHits(ctx) 
		}
	}
}

func (c *Character) findTarget(player *Character, others []*Character, playerDist float64) (float64, float64, bool, bool) {
	var bestX, bestY float64
	var hasTarget bool
	var isTargetPlayer bool
	minDist := 15.0

	isTargetValid := func(other *Character) bool {
		if c.Behavior == BehaviorChaotic { return true }
		if c.Alignment == AlignmentEnemy {
			if other.Alignment == AlignmentAlly { return true }
			if other.Alignment == AlignmentNeutral {
				// Enemy attacks Neutrals if they have a leader (traitor check)
				return other.LeaderID != "" || other.Group != ""
			}
			return false
		}
		if c.Alignment == AlignmentAlly {
			return other.Alignment == AlignmentEnemy
		}
		if c.Alignment == AlignmentNeutral {
			return c.TargetActor == &other.Actor
		}
		return false
	}

	// Check player
	if player != nil && player.IsAlive() && playerDist < minDist {
		if isTargetValid(player) {
			minDist = playerDist
			bestX, bestY = player.X, player.Y
			hasTarget = true
			isTargetPlayer = true
			c.TargetActor = &player.Actor
		}
	}

	// Check others
	for _, other := range others {
		if other == c || !other.IsAlive() { continue }
		dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
		if dist < minDist && isTargetValid(other) {
			minDist = dist
			bestX, bestY = other.X, other.Y
			hasTarget = true
			isTargetPlayer = false
			c.TargetActor = &other.Actor
		}
	}

	return bestX, bestY, hasTarget, isTargetPlayer
}

func (c *Character) findLootTarget(items []*ItemInstance) *ItemInstance {
	var best *ItemInstance; minDist := 10.0
	for _, it := range items {
		if !it.Pickable { continue }
		dist := math.Sqrt(math.Pow(c.X-it.X, 2) + math.Pow(c.Y-it.Y, 2))
		if dist < minDist { minDist = dist; best = it }
	}
	return best
}

func (c *Character) needsAIDecision(playerDist float64) bool {
	if playerDist < 10.0 || (c.Health < c.MaxHealth/2 && playerDist < 20.0) {
		interval := 300
		if IsDebugEnabled() { interval = 60 }
		return (c.Tick - c.LastAIDecisionTick) >= interval
	}
	return false
}

func (c *Character) ApplyAIDecision(dec AIDecision) {
	c.AIDecisionPending = false; c.LastAIChoice = dec.ChosenOption; c.LastAIReasoning = dec.Reasoning
	choice := strings.ToLower(dec.ChosenOption)
	if strings.Contains(choice, "attack") { c.Behavior = BehaviorNpcFighter
	} else if strings.Contains(choice, "flee") { c.Behavior = BehaviorFlee
	} else if strings.Contains(choice, "wander") || strings.Contains(choice, "talk") { c.Behavior = BehaviorWander
	} else if strings.Contains(choice, "patrol") { c.Behavior = BehaviorPatrol }
}
func (c *Character) calculateStat(base int, level int) int {
	if level <= 1 { return base }
	return int(float64(base) * math.Pow(1.15, float64(level-1)))
}

func clampInt(v, min, max int) int {
	if v < min { return min }
	if v > max { return max }
	return v
}
