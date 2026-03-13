package game

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"math"
	"math/rand"
	"os"

	"oinakos/internal/engine"
)

// Heal is now handled by Actor.Heal

// PlayableCharacterState, Direction, and their constants are defined in actor.go

type PlayableCharacter struct {
	Actor         // Embedded shared state
	Kills         int
	MapKills      map[string]int
}

func loadPlayerImage(assets fs.FS, path string) (image.Image, error) {
	var f fs.File
	var err error

	if assets != nil {
		f, err = assets.Open(path)
	}
	if err != nil || assets == nil {
		f, err = os.Open(path)
	}

	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func NewPlayableCharacter(x, y float64, config *EntityConfig) *PlayableCharacter {
	if config == nil {
		config = &EntityConfig{
			Name: "Knight",
		}
		config.Stats.HealthMin = 100
		config.Stats.BaseAttack = 20
		config.Stats.BaseDefense = 10
		config.Stats.Speed = 0.05
		config.Weapon = WeaponTizon
	}
	mc := &PlayableCharacter{
		Actor: Actor{
			X:           x,
			Y:           y,
			Config:      config,
			Facing:      DirSE,
			State:       StateIdle,
			Health:      config.Stats.HealthMin,
			MaxHealth:   config.Stats.HealthMin,
			Speed:       config.Stats.Speed,
			BaseAttack:  config.Stats.BaseAttack,
			BaseDefense: config.Stats.BaseDefense,
			Weapon:      config.Weapon,
			EquippedArmor: map[ArmorSlot]*Armor{
				SlotBody: ArmorLeather,
			},
			Level:     1,
			Alignment: AlignmentAlly,
			Name:      config.Name,
		},
		MapKills: make(map[string]int),
	}
	// Random quality bonus for starting weapon
	if mc.Weapon != nil {
		mc.Weapon.Bonus = rand.Intn(4) // 0 to 3
	}
	return mc
}

// GetTotalAttack, GetTotalDefense, GetTotalProtection, calculateStat, AddXP are now in Actor.

func (mc *PlayableCharacter) TakeDamage(amount int, audio AudioManager) {
	if mc.State == StateDead {
		return
	}
	oldHealth := mc.Health
	mc.Health -= amount
	mc.HitTimer = 15 // Show hit frame for 15 ticks
	DebugLog("Player Hit! Damage: %d | Health: %d -> %d", amount, oldHealth, mc.Health)
	if mc.Health <= 0 {
		mc.Health = 0
		mc.State = StateDead
		DebugLog("Player DIED at (%.2f, %.2f)", mc.X, mc.Y)
		if audio != nil && mc.Config != nil && mc.Config.PlayableCharacter != "" {
			audio.PlayRandomSound(mc.Config.PlayableCharacter + "/death")
		}
	} else {
		if audio != nil && mc.Config != nil && mc.Config.PlayableCharacter != "" {
			audio.PlayRandomSound(mc.Config.PlayableCharacter + "/hit")
		}
	}
}

// IsAlive is now in Actor.

// GetFootprint is now in Actor (uses Config.Footprint).
// PlayableCharacter currently uses a default hex footprint in its own version, 
// which is also the fallback in Actor.GetFootprint.

// checkCollisionAt is now in Actor.

func (mc *PlayableCharacter) Update(input engine.Input, audio AudioManager, obstacles []*Obstacle, npcs []*NPC, fts *[]*FloatingText, mapW, mapH float64, archs *ArchetypeRegistry, logFunc func(string, LogCategory)) {
	if mc.State == StateDead {
		if mc.DeadTimer == 0 {
			if mc.Config != nil {
				mc.X, mc.Y = findSafePosition(mc.X, mc.Y, mc.Config.GetFootprint(), obstacles)
			}
		}
		mc.DeadTimer++
		return
	}

	if mc.HitTimer > 0 {
		mc.HitTimer--
	}

	if mc.State == StateAttacking {
		mc.Tick++
		if mc.Tick == 15 {
			mc.CheckAttackHits(npcs, obstacles, fts, audio, archs, logFunc)
		}
		if mc.Tick > 30 {
			mc.State = StateIdle
			mc.Tick = 0
		}
		return
	}

	if mc.State == StateDrinking {
		mc.Tick++
		if mc.Tick > 60 { // 1 second drinking animation
			mc.State = StateIdle
			mc.Tick = 0
		}
		return
	}

	var dx, dy float64
	if input != nil {
		if input.IsKeyPressed(engine.KeyW) || input.IsKeyPressed(engine.KeyUp) {
			dy -= 1
		}
		if input.IsKeyPressed(engine.KeyS) || input.IsKeyPressed(engine.KeyDown) {
			dy += 1
		}
		if input.IsKeyPressed(engine.KeyA) || input.IsKeyPressed(engine.KeyLeft) {
			dx -= 1
		}
		if input.IsKeyPressed(engine.KeyD) || input.IsKeyPressed(engine.KeyRight) {
			dx += 1
		}

		if input.IsKeyPressed(engine.KeySpace) {
			// Check for interactive obstacles (like wells)
			for _, o := range obstacles {
				if o.Alive && o.Archetype != nil && o.CooldownTicks <= 0 {
					for _, action := range o.Archetype.Actions {
						if action.Type == ActionHeal && action.RequiresInteraction {
							dist := math.Sqrt(math.Pow(mc.X-o.X, 2) + math.Pow(mc.Y-o.Y, 2))
							if dist < 1.5 {
								// Use interactive healing
								if action.Amount >= 999 {
									mc.Health = mc.MaxHealth
								} else {
									mc.Heal(action.Amount)
								}
								// Use legacy cooldown_time logic (minutes to ticks)
								o.CooldownTicks = int(o.Archetype.CooldownTime * 60 * 60)
								DebugLog("Player used %s at (%.2f, %.2f) | Health: %d", o.Archetype.Name, o.X, o.Y, mc.Health)

								// Add healing text
								healVal := action.Amount
								if healVal > 1000 {
									healVal = mc.MaxHealth // Just a representative number for the text if it was a "Full Heal"
								}
								if fts != nil {
									*fts = append(*fts, &FloatingText{
										Text:  fmt.Sprintf("+%d", healVal),
										X:     mc.X,
										Y:     mc.Y,
										Life:  45,
										Color: ColorHeal,
									})
								}

								// Change state to drinking
								mc.State = StateDrinking
								mc.Tick = 0
								return
							}
						}
					}
				}
			}

			mc.State = StateAttacking
			mc.Tick = 0
			DebugLog("Player is Attacking! Pos: (%.2f, %.2f) | Facing: %v", mc.X, mc.Y, mc.Facing)
			if audio != nil && mc.Config != nil && mc.Config.PlayableCharacter != "" {
				audio.PlayRandomSound(mc.Config.PlayableCharacter + "/attack")
			}
			return
		}
	} // Close if input != nil

	if dx != 0 || dy != 0 {
		mc.State = StateWalking
		mc.Tick++

		mag := math.Sqrt(dx*dx + dy*dy)
		dx /= mag
		dy /= mag

		moveX := dx * mc.Speed * mc.GetSpeedModifier()
		moveY := dy * mc.Speed * mc.GetSpeedModifier()

		if !mc.checkCollisionAt(mc.X+moveX, mc.Y+moveY, obstacles) {
			mc.X += moveX
			mc.Y += moveY
		} else {
			if !mc.checkCollisionAt(mc.X+moveX, mc.Y, obstacles) {
				mc.X += moveX
			}
			if !mc.checkCollisionAt(mc.X, mc.Y+moveY, obstacles) {
				mc.Y += moveY
			}
		}

		if mc.Tick%30 == 0 {
			DebugLog("Player Moved to (%.2f, %.2f)", mc.X, mc.Y)
			if audio != nil {
				if mc.CurrentTile == "water.png" || mc.CurrentTile == "dark_water.png" {
					audio.PlayRandomSound("footstep_water")
				} else if mc.CurrentTile == "paved_ground.png" || mc.CurrentTile == "big_stones.png" {
					audio.PlayRandomSound("footstep_stone")
				} else {
					audio.PlayRandomSound("footstep_grass")
				}
			}
		}

		if dx > 0 {
			if dy < 0 {
				mc.Facing = DirNE
			} else if dy > 0 {
				mc.Facing = DirSE
			} else {
				mc.Facing = DirSE // Default for purely horizontal right
			}
		} else if dx < 0 {
			if dy < 0 {
				mc.Facing = DirNW
			} else if dy > 0 {
				mc.Facing = DirSW
			} else {
				mc.Facing = DirSW // Default for purely horizontal left
			}
		} else {
			// Purely vertical movement
			if dy < 0 {
				mc.Facing = DirNE // Up-Right in isometric
			} else if dy > 0 {
				mc.Facing = DirSW // Down-Left in isometric
			}
		}
	} else {
		mc.State = StateIdle
		mc.Tick = 0
	}

	// ALWAYS clamp to map boundaries
	halfW := mapW / 2
	halfH := mapH / 2
	if mc.X < -halfW {
		mc.X = -halfW
	}
	if mc.X > halfW {
		mc.X = halfW
	}
	if mc.Y < -halfH {
		mc.Y = -halfH
	}
	if mc.Y > halfH {
		mc.Y = halfH
	}
}

func (mc *PlayableCharacter) CheckAttackHits(npcs []*NPC, obstacles []*Obstacle, fts *[]*FloatingText, audio AudioManager, archs *ArchetypeRegistry, logFunc func(string, LogCategory)) {
	attackDist := 0.9
	atX, atY := mc.X, mc.Y

	// Fix: Normalize attack center based on facing.
	// SE is "right" in isometric view generally (X increases, Y increases)
	// NE is X increases, Y decreases.
	// We want the attack to land in front of the playableCharacter.
	switch mc.Facing {
	case DirSE:
		atX += attackDist
		atY += attackDist * 0.5
	case DirSW:
		atX -= attackDist * 0.5
		atY += attackDist
	case DirNE:
		atX += attackDist
		atY -= attackDist * 0.5
	case DirNW:
		atX -= attackDist * 0.5
		atY -= attackDist
	}

	for _, n := range npcs {
		if !n.IsAlive() {
			continue
		}
		// Generous circle check around the attack center
		dist := math.Sqrt(math.Pow(atX-n.X, 2) + math.Pow(atY-n.Y, 2))
		if dist < 1.6 { // Increased range for better feel
			// HIT ROLL — ratio-based so scaling doesn't break it
			// hitChance in [5, 95] based on relative attack vs defense
			attk := float64(mc.GetTotalAttack())
			def := float64(n.GetTotalDefense())
			if def <= 0 {
				def = 1
			}
			hitChance := int(attk / (attk + def) * 100)
			if hitChance < 5 {
				hitChance = 5
			}
			if hitChance > 95 {
				hitChance = 95
			}

			roll := rand.Intn(100) + 1
			if roll <= hitChance {
				// SUCCESSFUL HIT
				rawDmg := mc.Weapon.RollDamage()
				protection := n.GetTotalProtection()
				finalDmg := int(math.Max(1, float64(rawDmg-protection)))
				DebugLog("Player attacks NPC %s: HIT for %d damage (roll: %d/%d)", n.Name, finalDmg, roll, hitChance)
				n.TakeDamage(finalDmg, mc, nil, audio, npcs, archs, logFunc)

				*fts = append(*fts, &FloatingText{
					Text:  fmt.Sprintf("-%d", finalDmg),
					X:     n.X,
					Y:     n.Y,
					Life:  45,
					Color: ColorHarm,
				})
			} else {
				// MISS
				DebugLog("Player attacks NPC %s: MISS (roll: %d/%d)", n.Name, roll, hitChance)
				*fts = append(*fts, &FloatingText{
					Text:  "MISS",
					X:     n.X,
					Y:     n.Y,
					Life:  45,
					Color: ColorMiss,
				})
			}
		}
	}

	// OBSTACLE Damage
	for _, o := range obstacles {
		if !o.Alive {
			continue
		}
		// Circle check for obstacles
		dist := math.Sqrt(math.Pow(atX-o.X, 2) + math.Pow(atY-o.Y, 2))
		if dist < 1.8 { // Slightly larger radius for obstacles (easier to hit)
			if o.Archetype != nil && o.Archetype.Destructible {
				rawDmg := mc.Weapon.RollDamage()
				o.TakeDamage(rawDmg)

				*fts = append(*fts, &FloatingText{
					Text:  fmt.Sprintf("-%d", rawDmg),
					X:     o.X,
					Y:     o.Y,
					Life:  45,
					Color: ColorHarm,
				})
			}
		}
	}
}
