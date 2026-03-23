package game

import (
	"fmt"
	"math"
	"strings"

	"oinakos/internal/engine"
)

func (c *Character) Update(ctx *SystemContext) {
	c.Tick++
	c.SharedUpdate(ctx)
	if c.IsPlayerControlled { c.updatePlayer(ctx) } else { c.updateAI(ctx) }
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
	if c.State == ActorCrouching || c.IsIncapacitated() { return }

	if c.State == ActorDead {
		if c.DeadTimer == 0 { c.X, c.Y = findSafePosition(c.X, c.Y, c.GetCollisionCircle(), worldObstacles) }
		c.DeadTimer++
		return
	}

	if c.HitTimer > 0 { c.HitTimer-- }
	if c.State == ActorAttacking {
		if c.Tick == 15 { c.CheckAttackHits(ctx) }
		if c.Tick >= 30 { c.State = ActorIdle }
		return
	}
	if c.State == ActorDrinking {
		if c.Tick >= 60 { c.State = ActorIdle }
		return
	}
	if c.State == ActorChopping || c.State == ActorDigging {
		if c.Tick == 15 { c.CheckAttackHits(ctx) }
		if c.Tick >= 30 { c.State = ActorIdle }
		return
	}

	var dx, dy float64
	simMode := false
	if ctx.World != nil && ctx.World.Game != nil && ctx.World.Game.settings != nil {
		simMode = ctx.World.Game.settings.AISimulationMode
	}

	if c.IsPlayerControlled && !simMode {
		if ctx.Input != nil {
			if ctx.Input.IsKeyJustPressed(engine.KeyR) {
				if c.State == ActorResting { c.State = ActorIdle } else { c.State = ActorResting; c.Tick = 0 }
				return
			}
			if c.State == ActorResting {
				if ctx.Input.IsKeyPressed(engine.KeyW) || ctx.Input.IsKeyPressed(engine.KeyUp) || ctx.Input.IsKeyPressed(engine.KeyA) || ctx.Input.IsKeyPressed(engine.KeyLeft) || ctx.Input.IsKeyPressed(engine.KeyS) || ctx.Input.IsKeyPressed(engine.KeyDown) || ctx.Input.IsKeyPressed(engine.KeyD) || ctx.Input.IsKeyPressed(engine.KeyRight) || ctx.Input.IsKeyPressed(engine.KeySpace) {
					c.State = ActorIdle
				} else { return }
			}
			if ctx.Input.IsKeyPressed(engine.KeyW) || ctx.Input.IsKeyPressed(engine.KeyUp) { dy -= 1 }
			if ctx.Input.IsKeyPressed(engine.KeyS) || ctx.Input.IsKeyPressed(engine.KeyDown) { dy += 1 }
			if ctx.Input.IsKeyPressed(engine.KeyA) || ctx.Input.IsKeyPressed(engine.KeyLeft) { dx -= 1 }
			if ctx.Input.IsKeyPressed(engine.KeyD) || ctx.Input.IsKeyPressed(engine.KeyRight) { dx += 1 }
			if ctx.Input.IsKeyPressed(engine.KeySpace) {
				for _, o := range worldObstacles {
					if o.Alive && o.Archetype != nil && o.CooldownTicks <= 0 {
						for _, action := range o.Archetype.Actions {
							if action.Type == ActionHeal && action.RequiresInteraction {
								if math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2)) < 1.5 {
									if action.Amount >= 999 { c.Health = c.MaxHealth } else { c.Heal(action.Amount) }
									o.CooldownTicks = int(o.Archetype.CooldownTime * 3600)
									ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("+%d", action.Amount), X: c.X, Y: c.Y, Life: 45, Color: ColorHeal })
									c.State, c.Tick = ActorDrinking, 0
									return
								}
							}
						}
					}
				}
				c.State, c.Tick = ActorAttacking, 0
				if ctx.Audio != nil && c.Config != nil {
					prefix := c.Config.PlayableCharacter
					if prefix == "" { prefix = c.Config.ID }
					ctx.Audio.PlayRandomSound(prefix + "/attack")
				}
				return
			}
			if ctx.Input.IsKeyJustPressed(engine.KeyC) {
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
					c.State, c.Tick = ActorChopping, 0
					c.CheckAttackHits(ctx) // Instant feedback / hit on first frame
					return 
				} else if ctx.Log != nil { 
					ctx.Log(fmt.Sprintf("%s: I need an axe to chop wood! (Check inventory)", c.Name), LogWarning) 
				}
			}
			if ctx.Input.IsKeyJustPressed(engine.KeyV) {
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
				if isPikeEquipped { c.State, c.Tick = ActorDigging, 0; return } else if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s: I don't have a pike!", c.Name), LogNPC) }
			}
		}

		if dx != 0 || dy != 0 {
			mag := math.Sqrt(dx*dx + dy*dy)
			mvX, mvY := (dx/mag)*c.Speed*c.GetSpeedModifier(ctx), (dy/mag)*c.Speed*c.GetSpeedModifier(ctx)
			if !c.checkCollisionAt(c.X+mvX, c.Y+mvY, worldObstacles) {
				c.X, c.Y, c.State = c.X+mvX, c.Y+mvY, ActorWalking
				if c.Tick%30 == 0 && ctx.Audio != nil {
					sound := "footstep_grass"
					if c.CurrentTile == "water.png" || c.CurrentTile == "dark_water.png" { sound = "footstep_water" }
					if c.CurrentTile == "paved_ground.png" || c.CurrentTile == "big_stones.png" { sound = "footstep_stone" }
					ctx.Audio.PlayRandomSound(sound)
				}
			} else { c.State, c.Tick = ActorIdle, 0 }
			c.updateFacing(dx, dy)
		} else { c.State, c.Tick = ActorIdle, 0 }
		c.clampToMap(mapW, mapH)
	} else { c.updateAI(ctx) }
}

func (c *Character) updateFacing(dx, dy float64) {
	if dx > 0 { if dy < 0 { c.Facing = DirNE } else if dy > 0 { c.Facing = DirSE } else { c.Facing = DirSE }
	} else if dx < 0 { if dy < 0 { c.Facing = DirNW } else if dy > 0 { c.Facing = DirSW } else { c.Facing = DirSW }
	} else { if dy < 0 { c.Facing = DirNE } else if dy > 0 { c.Facing = DirSW } }
}
