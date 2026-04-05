package game

import (
	"fmt"
	"math"
	"strings"

	"oinakos/internal/engine"
)

func (c *Character) Update(ctx *SystemContext) {
	c.Tick++
	c.AttackTimer++
	if (c.ID == "oinakos" || c.ID == "hero") {
		c.LastLoggedState, c.LastLoggedHP = c.ActionState, c.State.HealthPoints
	}
	c.SharedUpdate(ctx)
	c.ProcessCooking(ctx)
	c.ProcessWorkshop(ctx)
	if c.IsPlayerControlled { c.updatePlayer(ctx) } else { 
		if c.Tick%10 == 0 {
			c.updateAI(ctx)
		}
		c.ShareRumors(ctx)
	}
}

func (c *Character) ShareRumors(ctx *SystemContext) {
	if c.Tick%300 != 0 { return } // Spread rumors periodically
	for _, other := range ctx.World.Characters {
		if other == c || !other.IsAlive() { continue }
		dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
		if dist < 3.0 && len(c.Memories) > 0 {
			// Share the most significant memory (highest absolute value)
			var rumor *MemoryEvent
			maxAbs := 0.0
			for _, m := range c.Memories {
				abs := math.Abs(m.Value)
				if abs > maxAbs { maxAbs, rumor = abs, &m }
			}
			if rumor != nil {
				// Prevent infinite feedback - only share if the other person doesn't know it
				hasKnown := false
				for _, m := range other.Memories { if m.Source == rumor.Source && m.Type == rumor.Type { hasKnown = true; break } }
				if !hasKnown {
					// Rumors are slightly less impactful than first-hand accounts
					other.AddMemory(c.Tick, rumor.Type, rumor.Source, rumor.Value * 0.8)
					// Visual feedback or Log?
					if IsDebugEnabled() { DebugLog("%s told %s a rumor about %s!", c.Name, other.Name, rumor.Source) }
				}
			}
		}
	}
}

func (c *Character) Draw(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64, adultMode bool) {
	DrawActor(&c.Actor, screen, textRenderer, vectorRenderer, paletteShader, offsetX, offsetY, c.IsPlayerControlled, adultMode)
}

func (c *Character) DrawUI(game *Game, screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64, debug bool) {
	DrawActorUI(game, &c.Actor, screen, textRenderer, vectorRenderer, offsetX, offsetY, c.IsPlayerControlled, debug)
}

func (c *Character) updatePlayer(ctx *SystemContext) {
	worldObstacles := ctx.World.Obstacles
	mapW, mapH := ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight
	if c.ActionState == ActorCrouching || c.IsIncapacitated() { return }

	if c.ActionState == ActorDead {
		if c.DeadTimer == 0 { c.X, c.Y = findSafePosition(c.X, c.Y, c.GetCollisionCircle(), worldObstacles) }
		c.DeadTimer++
		return
	}

	if c.HitTimer > 0 { c.HitTimer-- }
	if c.ActionState == ActorAttacking {
		if c.Tick == 15 { c.CheckAttackHits(ctx, c.PendingSkill) }
		if c.Tick >= 30 { 
			c.ActionState = ActorIdle
			if c.State.Sanity <= 0 { c.ActionState = ActorBerserk }
			c.PendingSkill = "" 
		}
		return
	}
	if c.ActionState == ActorDrinking || c.ActionState == ActorEating {
		if c.Tick >= 30 { c.ActionState = ActorIdle }
		return
	}
	if c.ActionState == ActorBathing || c.ActionState == ActorRelieving {
		if c.Tick >= 60 {
			if c.ActionState == ActorRelieving {
				c.AlleviateProperly(ctx)
				c.SpawnDefecation(ctx)
			}
			c.ActionState = ActorIdle
		}
		return
	}
	if c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging {
		if c.Tick == 15 { c.CheckAttackHits(ctx, "") }
		if c.Tick >= 30 { c.ActionState = ActorIdle }
		return
	}

	var dx, dy float64
	simMode := false
	if ctx.World != nil && ctx.World.Game != nil && ctx.World.Game.settings != nil {
		simMode = ctx.World.Game.settings.AISimulationMode
	}

	if c.IsPlayerControlled && !simMode {
		if ctx.Input != nil {
			if ctx.Input.IsKeyJustPressed(ctx.Settings.GetKey("rest")) {
				if c.ActionState == ActorResting { c.ActionState = ActorIdle } else { c.ActionState = ActorResting; c.Tick = 0 }
				return
			}
			if c.ActionState == ActorResting {
				if ctx.Input.IsKeyPressed(ctx.Settings.GetKey("move_up")) || ctx.Input.IsKeyPressed(engine.KeyUp) || 
				   ctx.Input.IsKeyPressed(ctx.Settings.GetKey("move_left")) || ctx.Input.IsKeyPressed(engine.KeyLeft) || 
				   ctx.Input.IsKeyPressed(ctx.Settings.GetKey("move_down")) || ctx.Input.IsKeyPressed(engine.KeyDown) || 
				   ctx.Input.IsKeyPressed(ctx.Settings.GetKey("move_right")) || ctx.Input.IsKeyPressed(engine.KeyRight) || 
				   ctx.Input.IsKeyPressed(ctx.Settings.GetKey("attack")) {
					c.ActionState = ActorIdle
				} else { return }
			}
			if ctx.Input.IsKeyPressed(ctx.Settings.GetKey("move_up")) || ctx.Input.IsKeyPressed(engine.KeyUp) { dy -= 1 }
			if ctx.Input.IsKeyPressed(ctx.Settings.GetKey("move_down")) || ctx.Input.IsKeyPressed(engine.KeyDown) { dy += 1 }
			if ctx.Input.IsKeyPressed(ctx.Settings.GetKey("move_left")) || ctx.Input.IsKeyPressed(engine.KeyLeft) { dx -= 1 }
			if ctx.Input.IsKeyPressed(ctx.Settings.GetKey("move_right")) || ctx.Input.IsKeyPressed(engine.KeyRight) { dx += 1 }
			if ctx.Input.IsKeyPressed(ctx.Settings.GetKey("attack")) {
				for _, o := range worldObstacles {
					if o.Alive && o.Archetype != nil && o.CooldownTicks <= 0 {
						for _, action := range o.Archetype.Actions {
							if action.Type == ActionHeal && action.RequiresInteraction {
								if math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2)) < 1.5 {
									if action.Amount >= 999 { c.State.HealthPoints = c.State.MaxHealthPoints } else { c.Heal(action.Amount) }
									o.CooldownTicks = int(o.Archetype.CooldownTime * 3600)
									ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("+%d", action.Amount), X: c.X, Y: c.Y, Life: 45, Color: ColorHeal })
									c.ActionState, c.Tick = ActorDrinking, 0
									return
								}
							} else if action.Type == ActionBath && action.RequiresInteraction {
								if math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2)) < 1.5 {
									c.ActionState, c.Tick = ActorBathing, 0
									o.CooldownTicks = 600
									return
								}
							} else if action.Type == ActionAlleviate && action.RequiresInteraction {
								if math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2)) < 1.5 {
									c.ActionState, c.Tick = ActorRelieving, 0
									o.CooldownTicks = 300
									return
								}
							}
						}
					}
				}
				c.ActionState, c.Tick, c.PendingSkill = ActorAttacking, 0, ""
				if ctx.Audio != nil && c.Config != nil {
					prefix := c.Config.PlayableCharacter
					if prefix == "" { prefix = c.Config.ID }
					ctx.Audio.PlayRandomSound(prefix + "/attack")
				}
				return
			}
			if ctx.Input.IsKeyJustPressed(ctx.Settings.GetKey("punch")) {
				c.ActionState, c.Tick, c.PendingSkill = ActorAttacking, 0, AttackPunch
				if ctx.Audio != nil && c.Config != nil {
					prefix := c.Config.PlayableCharacter
					if prefix == "" { prefix = c.Config.ID }
					ctx.Audio.PlayRandomSound(prefix + "/attack")
				}
				return
			}
			if ctx.Input.IsKeyJustPressed(ctx.Settings.GetKey("slap")) {
				c.ActionState, c.Tick, c.PendingSkill = ActorAttacking, 0, AttackSlap
				if ctx.Audio != nil && c.Config != nil {
					prefix := c.Config.PlayableCharacter
					if prefix == "" { prefix = c.Config.ID }
					ctx.Audio.PlayRandomSound(prefix + "/attack")
				}
				return
			}
			if ctx.Input.IsKeyJustPressed(ctx.Settings.GetKey("chop")) {
				isAxeEquipped := c.Weapon != nil && strings.Contains(strings.ToLower(c.Weapon.Name), "axe")
				if !isAxeEquipped {
					for _, item := range c.Inventory {
						if item != nil && item.Config != nil && strings.Contains(strings.ToLower(item.Config.Name), "axe") {
							if c.Slots == nil { c.Slots = make(map[string]*ItemInstance) }
							c.Slots["weapon"] = item; c.Weapon = item.Config.Combat; c.UpdateEffects()
							isAxeEquipped = true
							if ctx.Audio != nil { ctx.Audio.PlayRandomSound("pickup") }
							break
						}
					}
				}
				if isAxeEquipped { 
					c.ActionState, c.Tick = ActorChopping, 0
					c.CheckAttackHits(ctx, "") // Instant feedback / hit on first frame
					return 
				} else if ctx.Log != nil { 
					ctx.Log(fmt.Sprintf("%s: I need an axe to chop wood! (Check inventory)", c.Name), LogWarning) 
				}
			}
			if ctx.Input.IsKeyJustPressed(ctx.Settings.GetKey("dig")) {
				isPikeEquipped := c.Weapon != nil && (strings.Contains(strings.ToLower(c.Weapon.Name), "pike") || strings.Contains(strings.ToLower(c.Weapon.Name), "pickaxe"))
				if !isPikeEquipped {
					for _, item := range c.Inventory {
						if item != nil && item.Config != nil {
							nameLow := strings.ToLower(item.Config.Name)
							if strings.Contains(nameLow, "pike") || strings.Contains(nameLow, "pickaxe") {
								if c.Slots == nil { c.Slots = make(map[string]*ItemInstance) }
								c.Slots["weapon"] = item; c.Weapon = item.Config.Combat; c.UpdateEffects()
								isPikeEquipped = true
								if ctx.Audio != nil { ctx.Audio.PlayRandomSound("pickup") }
								break
							}
						}
					}
				}
				if isPikeEquipped { c.ActionState, c.Tick = ActorDigging, 0; return } else if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s: I don't have a pike!", c.Name), LogNPC) }
			}
			if ctx.Input.IsKeyJustPressed(ctx.Settings.GetKey("forage")) {
				c.ActionState, c.Tick = ActorForaging, 0
				return
			}
		}

		if dx != 0 || dy != 0 {
			mag := math.Sqrt(dx*dx + dy*dy)
			mvX, mvY := (dx/mag)*c.Speed*c.GetSpeedModifier(ctx), (dy/mag)*c.Speed*c.GetSpeedModifier(ctx)
			if !c.checkCollisionAt(c.X+mvX, c.Y+mvY, worldObstacles) {
				c.X, c.Y, c.ActionState = c.X+mvX, c.Y+mvY, ActorWalking
				if c.Tick%30 == 0 && ctx.Audio != nil {
					sound := "footstep_grass"
					if c.CurrentTile == "water.png" || c.CurrentTile == "dark_water.png" { sound = "footstep_water" }
					if c.CurrentTile == "paved_ground.png" || c.CurrentTile == "big_stones.png" { sound = "footstep_stone" }
					ctx.Audio.PlayRandomSound(sound)
				}
			} else { c.ActionState, c.Tick = ActorIdle, 0 }
			c.updateFacing(dx, dy)
		} else if c.ActionState == ActorWalking { 
			c.ActionState, c.Tick = ActorIdle, 0 
		}
		c.clampToMap(mapW, mapH)
	} else { c.updateAI(ctx) }
}

func (c *Character) updateFacing(dx, dy float64) {
	if dx > 0 { if dy < 0 { c.Facing = DirNE } else if dy > 0 { c.Facing = DirSE } else { c.Facing = DirSE }
	} else if dx < 0 { if dy < 0 { c.Facing = DirNW } else if dy > 0 { c.Facing = DirSW } else { c.Facing = DirSW }
	} else { if dy < 0 { c.Facing = DirNE } else if dy > 0 { c.Facing = DirSW } }
}
