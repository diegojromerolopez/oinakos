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

func (mc *MainCharacter) Heal(amount int) {
	if mc.State == StateDead {
		return
	}
	oldHealth := mc.Health
	mc.Health += amount
	if mc.Health > mc.MaxHealth {
		mc.Health = mc.MaxHealth
	}
	if mc.Health > oldHealth {
		DebugLog("Player Healed! +%d | Health: %d -> %d", amount, oldHealth, mc.Health)
	}
}

type MainCharacterState int

const (
	StateIdle MainCharacterState = iota
	StateWalking
	StateAttacking
	StateDead
	StateDrinking
)

type Direction int

const (
	DirSE Direction = iota
	DirSW
	DirNE
	DirNW
	DirS
	DirW
	DirN
	DirE
)

type MainCharacter struct {
	X, Y          float64
	Config        *EntityConfig
	Speed         float64
	Facing        Direction
	State         MainCharacterState
	Tick          int
	Health        int
	MaxHealth     int
	Kills         int
	MapKills      map[string]int
	XP            int
	Level         int
	BaseAttack    int
	BaseDefense   int
	Weapon        *Weapon
	EquippedArmor map[ArmorSlot]*Armor
	HitTimer      int
	DeadTimer     int
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

func NewMainCharacter(x, y float64, config *EntityConfig) *MainCharacter {
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
	mc := &MainCharacter{
		X:           x,
		Y:           y,
		Config:      config,
		Facing:      DirSE,
		State:       StateIdle,
		Health:      config.Stats.HealthMin,
		MaxHealth:   config.Stats.HealthMin,
		Speed:       config.Stats.Speed,
		MapKills:    make(map[string]int),
		BaseAttack:  config.Stats.BaseAttack,
		BaseDefense: config.Stats.BaseDefense,
		Weapon:      config.Weapon,
		EquippedArmor: map[ArmorSlot]*Armor{
			SlotBody: ArmorLeather,
		},
		Level: 1,
	}
	// Random quality bonus for starting weapon
	if mc.Weapon != nil {
		mc.Weapon.Bonus = rand.Intn(4) // 0 to 3
	}
	return mc
}

func (mc *MainCharacter) GetTotalAttack() int {
	return mc.calculateStat(mc.BaseAttack, mc.Level)
}

func (mc *MainCharacter) GetTotalDefense() int {
	return mc.calculateStat(mc.BaseDefense, mc.Level)
}

func (mc *MainCharacter) GetTotalProtection() int {
	total := 0
	for _, a := range mc.EquippedArmor {
		if a != nil {
			total += a.Protection
		}
	}
	return total
}

func (mc *MainCharacter) calculateStat(base, level int) int {
	// Logarithmic scaling: stat = base + log2(level) * scalingFactor
	// scalingFactor = 10 for meaningful growth
	if level <= 1 {
		return base
	}
	bonus := int(math.Log2(float64(level)) * 10)
	return base + bonus
}

func (mc *MainCharacter) AddXP(amount int) {
	mc.XP += amount
	// Simple level up logic: level = XP / 100 + 1
	newLevel := mc.XP/100 + 1
	if newLevel > mc.Level {
		mc.Level = newLevel
		// Optionally heal on level up
		mc.Health = mc.MaxHealth
	}
}

func (mc *MainCharacter) TakeDamage(amount int, audio AudioManager) {
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
		if audio != nil && mc.Config != nil && mc.Config.MainCharacter != "" {
			audio.PlayRandomSound(mc.Config.MainCharacter + "/death")
		}
	} else {
		if audio != nil && mc.Config != nil && mc.Config.MainCharacter != "" {
			audio.PlayRandomSound(mc.Config.MainCharacter + "/hit")
		}
	}
}

func (mc *MainCharacter) IsAlive() bool {
	return mc.State != StateDead
}

func (mc *MainCharacter) GetFootprint() engine.Polygon {
	return engine.Polygon{Points: []engine.Point{
		{X: -0.2, Y: -0.1}, {X: 0.2, Y: -0.1}, {X: 0.3, Y: 0}, {X: 0.2, Y: 0.1}, {X: -0.2, Y: 0.1}, {X: -0.3, Y: 0},
	}}.Transformed(mc.X, mc.Y)
}

func (mc *MainCharacter) checkCollisionAt(newX, newY float64, obstacles []*Obstacle) bool {
	pFootprint := engine.Polygon{Points: []engine.Point{
		{X: -0.2, Y: -0.1}, {X: 0.2, Y: -0.1}, {X: 0.3, Y: 0}, {X: 0.2, Y: 0.1}, {X: -0.2, Y: 0.1}, {X: -0.3, Y: 0},
	}}.Transformed(newX, newY)

	for _, o := range obstacles {
		if !o.Alive {
			continue
		}
		if engine.CheckCollision(pFootprint, o.GetFootprint()) {
			return true
		}
	}
	return false
}

func (mc *MainCharacter) Update(input engine.Input, audio AudioManager, obstacles []*Obstacle, npcs []*NPC, fts *[]*FloatingText, mapW, mapH float64) {
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
			mc.CheckAttackHits(npcs, obstacles, fts, audio)
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

	if mc.State == StateAttacking {
		mc.Tick++
		if mc.Tick > 20 { // 20 frames of attack pose
			mc.State = StateIdle
			mc.Tick = 0
		}
		// Allow small movement or rotation adjustments if desired,
		// but typically we lock movement during attack.
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
			if audio != nil && mc.Config != nil && mc.Config.MainCharacter != "" {
				audio.PlayRandomSound(mc.Config.MainCharacter + "/attack")
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

		moveX := dx * mc.Speed
		moveY := dy * mc.Speed

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

func (mc *MainCharacter) CheckAttackHits(npcs []*NPC, obstacles []*Obstacle, fts *[]*FloatingText, audio AudioManager) {
	attackDist := 0.9
	atX, atY := mc.X, mc.Y

	// Fix: Normalize attack center based on facing.
	// SE is "right" in isometric view generally (X increases, Y increases)
	// NE is X increases, Y decreases.
	// We want the attack to land in front of the mainCharacter.
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
				n.TakeDamage(finalDmg, mc, nil, audio, npcs)

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
