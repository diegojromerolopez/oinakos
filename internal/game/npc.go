package game

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"oinakos/internal/engine"
)

// NPCType is kept for legacy compatibility if needed,
// but ConfigID is preferred for dynamic loading.
type NPCType string

type NPCState int

const (
	NPCIdle NPCState = iota
	NPCWalking
	NPCAttacking
	NPCDead
)

type BehaviorType int
type Alignment int

const (
	AlignmentEnemy Alignment = iota
	AlignmentNeutral
	AlignmentAlly
)

func (a Alignment) String() string {
	switch a {
	case AlignmentEnemy:
		return "ENEMY"
	case AlignmentNeutral:
		return "NEUTRAL"
	case AlignmentAlly:
		return "ALLY"
	default:
		return "UNKNOWN"
	}
}

const (
	BehaviorWander BehaviorType = iota
	BehaviorPatrol
	BehaviorKnightHunter
	BehaviorNpcFighter
	BehaviorChaotic
)

type NPC struct {
	X, Y           float64
	Archetype      *Archetype
	State          NPCState
	Facing         Direction
	Tick           int
	Health         int
	MaxHealth      int
	Speed          float64
	AttackCooldown int // frames between attacks
	AttackTimer    int
	Name           string
	Level          int
	XP             int
	BaseAttack     int
	BaseDefense    int
	Weapon         *Weapon

	// Behavior fields
	Behavior   BehaviorType
	WanderDirX float64
	WanderDirY float64
	// Combat & Effects
	BloodTimer                 int
	DeadTimer                  int
	PatrolStartX, PatrolStartY float64
	PatrolEndX, PatrolEndY     float64
	PatrolHeading              bool // true = toward End, false = toward Start

	TargetNPC    *NPC
	TargetPlayer *MainCharacter
	IsBoss       bool
	Alignment    Alignment
}

var npcNames = []string{
	"Grog", "Zog", "Bob", "Drok", "Gorak", "Mug", "Snarl", "Thrak", "Vrog", "Kurg",
	"Hicks", "Miller", "Cooper", "Smith", "Potter", "Baker", "Carter", "Fisher",
}

func NewNPC(x, y float64, archetype *Archetype, level int) *NPC {
	if archetype == nil {
		archetype = &Archetype{
			ID:   "unknown",
			Name: "Unknown Entity",
		}
		archetype.Stats.HealthMin = 10
		archetype.Stats.HealthMax = 10
		archetype.Weapon = WeaponTizon
	}
	n := &NPC{
		X:         x,
		Y:         y,
		Archetype: archetype,
		State:     NPCIdle,
		Facing:    DirSE,
		Level:     level,
		Alignment: AlignmentEnemy, // Default to Enemy
	}

	if archetype.Unique {
		n.Name = archetype.Name
	} else if len(archetype.Names) > 0 {
		n.Name = archetype.Names[rand.Intn(len(archetype.Names))]
	} else if archetype.Name != "" {
		n.Name = archetype.Name
	} else {
		n.Name = npcNames[rand.Intn(len(npcNames))]
	}

	// Load from archetype
	n.Health = archetype.Stats.HealthMin + rand.Intn(archetype.Stats.HealthMax-archetype.Stats.HealthMin+1)
	n.BaseAttack = archetype.Stats.BaseAttack
	n.BaseDefense = archetype.Stats.BaseDefense
	n.Speed = archetype.Stats.Speed
	n.AttackCooldown = archetype.Stats.AttackCooldown
	n.Weapon = archetype.Weapon

	// Set behavior from archetype
	switch archetype.Behavior {
	case "chaotic":
		n.Behavior = BehaviorChaotic
	case "fighter":
		n.Behavior = BehaviorNpcFighter
	case "hunter":
		n.Behavior = BehaviorKnightHunter
	case "wander":
		n.Behavior = BehaviorWander
	case "patrol":
		n.Behavior = BehaviorPatrol
	default:
		if archetype.Unique {
			n.Behavior = BehaviorWander
		} else {
			n.Behavior = BehaviorKnightHunter
		}
	}

	// Dynamic scaling based on level
	n.Health = n.calculateStat(n.Health, n.Level)
	n.MaxHealth = n.Health
	n.BaseAttack = n.calculateStat(n.BaseAttack, n.Level)
	n.BaseDefense = n.calculateStat(n.BaseDefense, n.Level)

	// Initialize random behavior if none provided
	if archetype == nil || archetype.Behavior == "" {
		n.Behavior = BehaviorType(rand.Intn(4))
	}

	// Pre-calculations for behaviors
	if n.Behavior == BehaviorWander {
		n.WanderDirX = rand.Float64()*2 - 1
		n.WanderDirY = rand.Float64()*2 - 1
	} else if n.Behavior == BehaviorPatrol {
		n.PatrolStartX = n.X
		n.PatrolStartY = n.Y
		// Patrol destination is somewhere nearby
		n.PatrolEndX = n.X + (rand.Float64()*8 - 4)
		n.PatrolEndY = n.Y + (rand.Float64()*8 - 4)
		n.PatrolHeading = true
	}

	return n
}

func (n *NPC) GetTotalAttack() int {
	return n.BaseAttack // Level is already baked during NewNPC for simple NPCs
}

func (n *NPC) GetTotalDefense() int {
	return n.BaseDefense
}

func (n *NPC) GetTotalProtection() int {
	return 0 // Placeholder for armor system?
}

func (n *NPC) calculateStat(base, level int) int {
	if level <= 1 {
		return base
	}
	bonus := int(math.Log2(float64(level)) * 10)
	return base + bonus
}

func (n *NPC) checkCollisionAt(newX, newY float64, obstacles []*Obstacle) bool {
	if n.Archetype == nil {
		return false
	}
	nFootprint := n.Archetype.GetFootprint().Transformed(newX, newY)
	for _, o := range obstacles {
		if !o.Alive {
			continue
		}
		if engine.CheckCollision(nFootprint, o.GetFootprint()) {
			return true
		}
	}
	return false
}

func (n *NPC) GetFootprint() engine.Polygon {
	if n.Archetype == nil {
		return engine.Polygon{}
	}
	return n.Archetype.GetFootprint().Transformed(n.X, n.Y)
}

func (n *NPC) Update(mainCharacter *MainCharacter, obstacles []*Obstacle, allNPCs []*NPC, projectiles *[]*Projectile, fts *[]*FloatingText, mapW, mapH float64, audio AudioManager) {
	n.Tick++

	if n.BloodTimer > 0 {
		n.BloodTimer--
	}

	if n.State == NPCDead {
		n.DeadTimer++
		return
	}

	// Custom Escort Logic
	if n.Archetype != nil && n.Archetype.ID == "escort" {
		n.updateEscort(mainCharacter, obstacles)
		return
	}

	if n.State == NPCAttacking {
		n.AttackTimer++
	}

	var targetX, targetY float64
	var hasTarget bool
	var isTargetPlayer bool

	// Override behavior based on alignment
	if n.Alignment == AlignmentAlly {
		// Try to find nearest enemy NPC
		var nearestEnemy *NPC
		var minDist = 999.0
		for _, other := range allNPCs {
			if other == n {
				continue
			}
			if !other.IsAlive() || other.Alignment != AlignmentEnemy {
				continue
			}
			dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
			if dist < 15.0 && dist < minDist { // 15 unit range for allies to notice enemies
				minDist = dist
				nearestEnemy = other
			}
		}

		if nearestEnemy != nil {
			n.TargetNPC = nearestEnemy
			n.TargetPlayer = nil
			targetX, targetY = nearestEnemy.X, nearestEnemy.Y
			hasTarget = true
		} else {
			// No enemies, follow player
			distToPlayer := math.Sqrt(math.Pow(n.X-mainCharacter.X, 2) + math.Pow(n.Y-mainCharacter.Y, 2))
			if distToPlayer > 3.0 {
				targetX, targetY = mainCharacter.X, mainCharacter.Y
				hasTarget = true
			} else {
				n.State = NPCIdle
				return
			}
		}
	} else if n.Alignment == AlignmentNeutral {
		// Strictly wander, ignore player and NPCs unless we add retaliation logic later
		if n.Tick%120 == 0 {
			n.WanderDirX = rand.Float64()*2 - 1
			n.WanderDirY = rand.Float64()*2 - 1
		}
		targetX, targetY = n.X+n.WanderDirX, n.Y+n.WanderDirY
		hasTarget = true
	}

	// Reassess target for certain behaviors
	if n.Behavior == BehaviorChaotic || n.Behavior == BehaviorNpcFighter {
		// Clear existing target to force reassessment in the switch below
		n.TargetPlayer = nil
		n.TargetNPC = nil
	}

	if !hasTarget {
		// Behavior Logic (Traditional Enemy behavior)
		if n.TargetPlayer != nil && n.TargetPlayer.IsAlive() {
			targetX, targetY = n.TargetPlayer.X, n.TargetPlayer.Y
			hasTarget = true
			isTargetPlayer = true
		} else if n.TargetNPC != nil && n.TargetNPC.IsAlive() {
			targetX, targetY = n.TargetNPC.X, n.TargetNPC.Y
			hasTarget = true
		} else {
			switch n.Behavior {
			case BehaviorKnightHunter:
				if mainCharacter.IsAlive() {
					n.TargetPlayer = mainCharacter
					targetX, targetY = mainCharacter.X, mainCharacter.Y
					hasTarget = true
					isTargetPlayer = true
				}
			case BehaviorNpcFighter:
				// Find nearest living NPC that isn't me
				var minDist = 999.0
				for _, other := range allNPCs {
					if other == n || !other.IsAlive() {
						continue
					}
					dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
					if dist < minDist {
						minDist = dist
						n.TargetNPC = other
					}
				}
				if n.TargetNPC != nil {
					targetX, targetY = n.TargetNPC.X, n.TargetNPC.Y
					hasTarget = true
				}
			case BehaviorChaotic:
				// Find nearest living actor (NPC or MainCharacter) that isn't me
				var minDist = 999.0
				var nearestNPC *NPC
				var playerDist = math.Sqrt(math.Pow(n.X-mainCharacter.X, 2) + math.Pow(n.Y-mainCharacter.Y, 2))

				for _, other := range allNPCs {
					if other == n || !other.IsAlive() {
						continue
					}
					dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
					if dist < minDist {
						minDist = dist
						nearestNPC = other
					}
				}

				if mainCharacter.IsAlive() && playerDist < minDist {
					n.TargetPlayer = mainCharacter
					n.TargetNPC = nil
					targetX, targetY = mainCharacter.X, mainCharacter.Y
					hasTarget = true
				} else if nearestNPC != nil {
					n.TargetNPC = nearestNPC
					n.TargetPlayer = nil
					targetX, targetY = nearestNPC.X, nearestNPC.Y
					hasTarget = true
				}
			case BehaviorWander:
				if n.Tick%120 == 0 {
					n.WanderDirX = rand.Float64()*2 - 1
					n.WanderDirY = rand.Float64()*2 - 1
				}
				targetX, targetY = n.X+n.WanderDirX, n.Y+n.WanderDirY
				hasTarget = true
			case BehaviorPatrol:
				if n.PatrolHeading {
					targetX, targetY = n.PatrolEndX, n.PatrolEndY
					if math.Sqrt(math.Pow(n.X-n.PatrolEndX, 2)+math.Pow(n.Y-n.PatrolEndY, 2)) < 0.5 {
						n.PatrolHeading = false
					}
				} else {
					targetX, targetY = n.PatrolStartX, n.PatrolStartY
					if math.Sqrt(math.Pow(n.X-n.PatrolStartX, 2)+math.Pow(n.Y-n.PatrolStartY, 2)) < 0.5 {
						n.PatrolHeading = true
					}
				}
				hasTarget = true
			default:
				// Fallback for generic Enemy alignment: attack player
				if n.Alignment == AlignmentEnemy {
					targetX, targetY = mainCharacter.X, mainCharacter.Y
					hasTarget = true
					isTargetPlayer = true
				}
			}
		}
	}

	if !hasTarget {
		n.State = NPCIdle
		return
	}

	dx := targetX - n.X
	dy := targetY - n.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	attackRange := 1.0
	if n.Archetype != nil && n.Archetype.Stats.AttackRange > 1.0 {
		attackRange = n.Archetype.Stats.AttackRange
	}

	isRanged := attackRange > 1.0

	if dist < attackRange {
		if isRanged && dist < attackRange-2.0 {
			// Kite away if too close
			kMag := math.Sqrt(dx*dx + dy*dy)
			if kMag > 0 {
				moveX := -(dx / kMag) * n.Speed
				moveY := -(dy / kMag) * n.Speed
				// Sliding collision for NPC kiting
				if !n.checkCollisionAt(n.X+moveX, n.Y+moveY, obstacles) {
					n.X += moveX
					n.Y += moveY
				} else {
					if !n.checkCollisionAt(n.X+moveX, n.Y, obstacles) {
						n.X += moveX
					}
					if !n.checkCollisionAt(n.X, n.Y+moveY, obstacles) {
						n.Y += moveY
					}
				}
			}
			n.State = NPCWalking
			return
		}

		if n.State != NPCAttacking && isTargetPlayer {
			// Chance to say a menace line when starting to attack the mainCharacter
			if rand.Float64() < 0.3 {
				msgNum := rand.Intn(5) + 1
				if audio != nil {
					switch n.Archetype.ID {
					case "orc":
						audio.PlaySound(fmt.Sprintf("orc_menace_%d", msgNum))
					case "demon":
						audio.PlaySound(fmt.Sprintf("demon_menace_%d", msgNum))
					case "peasant":
						audio.PlaySound(fmt.Sprintf("peasant_menace_%d", msgNum))
					}
				}
			}
		}

		n.State = NPCAttacking
		if n.AttackTimer >= n.AttackCooldown {
			n.AttackTimer = 0

			if isRanged {
				// Spawn Projectile
				mag := math.Sqrt(dx*dx + dy*dy)
				if mag > 0 {
					pSpd := n.Archetype.Stats.ProjectileSpeed
					if pSpd <= 0 {
						pSpd = 0.15 // fallback default
					}

					proj := NewProjectile(n.X, n.Y, dx/mag, dy/mag, pSpd, n.GetTotalAttack(), false, 100.0)
					*projectiles = append(*projectiles, proj)

					if n.Archetype != nil && n.Archetype.ID == "stultus" {
						*fts = append(*fts, &FloatingText{
							Text:  "SHOUT!",
							X:     n.X,
							Y:     n.Y,
							Life:  30,
							Color: color.RGBA{255, 255, 0, 255},
						})
					}
				}
			} else {
				// MELEE HIT ROLL
				var targetProtection int

				if isTargetPlayer {
					targetProtection = mainCharacter.GetTotalProtection()
					// We need a wrapper or cast if we want to use the same interface,
					// but for now let's just do it directly.
					attk := float64(n.GetTotalAttack())
					def := float64(mainCharacter.GetTotalDefense())
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
						rawDmg := n.Weapon.RollDamage()
						finalDmg := int(math.Max(1, float64(rawDmg-targetProtection)))
						log.Printf("NPC %s attacks Player: HIT for %d damage (roll: %d/%d)", n.Name, finalDmg, roll, hitChance)
						mainCharacter.TakeDamage(finalDmg, audio)

						*fts = append(*fts, &FloatingText{
							Text:  fmt.Sprintf("%d", finalDmg),
							X:     mainCharacter.X,
							Y:     mainCharacter.Y,
							Life:  45,
							Color: color.RGBA{255, 100, 100, 255},
						})
					} else {
						// MISS
						log.Printf("NPC %s attacks Player: MISS (roll: %d/%d)", n.Name, roll, hitChance)
						*fts = append(*fts, &FloatingText{
							Text:  "MISS",
							X:     mainCharacter.X,
							Y:     mainCharacter.Y,
							Life:  45,
							Color: color.RGBA{200, 200, 200, 255},
						})
					}
				} else {
					// NPC VS NPC
					if n.TargetNPC == nil || !n.TargetNPC.IsAlive() {
						n.TargetNPC = nil
						return
					}
					targetProtection = n.TargetNPC.GetTotalProtection()

					attk := float64(n.GetTotalAttack())
					def := float64(n.TargetNPC.GetTotalDefense())
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
						rawDmg := n.Weapon.RollDamage()
						finalDmg := int(math.Max(1, float64(rawDmg-targetProtection)))
						log.Printf("NPC %s attacks NPC %s: HIT for %d damage (roll: %d/%d)", n.Name, n.TargetNPC.Name, finalDmg, roll, hitChance)
						n.TargetNPC.TakeDamage(finalDmg, nil, n, audio)

						*fts = append(*fts, &FloatingText{
							Text:  fmt.Sprintf("%d", finalDmg),
							X:     n.TargetNPC.X,
							Y:     n.TargetNPC.Y,
							Life:  45,
							Color: color.RGBA{255, 100, 100, 255},
						})
					} else {
						// MISS
						log.Printf("NPC %s attacks NPC %s: MISS (roll: %d/%d)", n.Name, n.TargetNPC.Name, roll, hitChance)
						*fts = append(*fts, &FloatingText{
							Text:  "MISS",
							X:     n.TargetNPC.X,
							Y:     n.TargetNPC.Y,
							Life:  45,
							Color: color.RGBA{150, 150, 150, 255},
						})
					}
				}
			}
		}
	} else {
		// Move toward target
		n.State = NPCWalking
		mag := math.Sqrt(dx*dx + dy*dy)
		if mag > 0 {
			ndx := dx / mag
			ndy := dy / mag

			moveX := ndx * n.Speed
			moveY := ndy * n.Speed

			// Sliding collision for NPC
			if !n.checkCollisionAt(n.X+moveX, n.Y+moveY, obstacles) {
				n.X += moveX
				n.Y += moveY
			} else {
				if !n.checkCollisionAt(n.X+moveX, n.Y, obstacles) {
					n.X += moveX
				}
				if !n.checkCollisionAt(n.X, n.Y+moveY, obstacles) {
					n.Y += moveY
				}
			}

			// Clamp to map boundaries
			halfW := mapW / 2
			halfH := mapH / 2
			if n.X < -halfW {
				n.X = -halfW
			}
			if n.X > halfW {
				n.X = halfW
			}
			if n.Y < -halfH {
				n.Y = -halfH
			}
			if n.Y > halfH {
				n.Y = halfH
			}

			// Facing direction
			if ndx > 0 {
				if ndy < 0 {
					n.Facing = DirNE
				} else {
					n.Facing = DirSE
				}
			} else if ndx < 0 {
				if ndy < 0 {
					n.Facing = DirNW
				} else {
					n.Facing = DirSW
				}
			}
		}
	}
}

func (n *NPC) TakeDamage(amount int, attackerPlayer *MainCharacter, attackerNPC *NPC, audio AudioManager) {
	if n.State == NPCDead {
		return
	}
	n.Health -= amount

	// Retaliation tracking
	if attackerPlayer != nil {
		n.TargetPlayer = attackerPlayer
		n.TargetNPC = nil
		// Neutral or Ally NPCs become enemies if hit by the player
		if n.Alignment != AlignmentEnemy {
			log.Printf("NPC %s was %s and is now an ENEMY due to player attack!", n.Name, n.Alignment)
			n.Alignment = AlignmentEnemy
			n.Behavior = BehaviorKnightHunter
		}
	} else if attackerNPC != nil {
		n.TargetNPC = attackerNPC
		n.TargetPlayer = nil
	}

	// Start blood effect
	n.BloodTimer = 30
	// Play hit sound
	if audio != nil && n.Archetype != nil {
		switch n.Archetype.ID {
		case "orc":
			audio.PlaySound("orc_hit")
		case "demon":
			audio.PlaySound("demon_hit")
		case "peasant":
			audio.PlaySound("peasant_hit")
		}
	}

	if n.Health <= 0 {
		log.Printf("NPC %s has been killed!", n.Name)
		n.State = NPCDead
		if attackerPlayer != nil {
			attackerPlayer.Kills++
			// Award XP from YAML-defined archetype value
			if n.Archetype != nil && n.Archetype.XP > 0 {
				attackerPlayer.XP += n.Archetype.XP
			} else {
				// Fallback: 1 XP so every kill counts
				attackerPlayer.XP += 1
			}
		}
		if audio != nil && n.Archetype != nil {
			switch n.Archetype.ID {
			case "orc", "orc_male", "orc_female":
				audio.PlaySound("orc_death")
			case "demon", "demon_male", "demon_female":
				audio.PlaySound("demon_death")
			case "peasant", "peasant_male", "peasant_female":
				audio.PlaySound("peasant_death")
			}
		}
	}
}

func (n *NPC) IsAlive() bool {
	return n.State != NPCDead
}

func (n *NPC) updateEscort(mainCharacter *MainCharacter, obstacles []*Obstacle) {
	dx := mainCharacter.X - n.X
	dy := mainCharacter.Y - n.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	// Follow mainCharacter if further than 5 units, but don't get closer than 3
	if dist > 5.0 {
		n.State = NPCWalking
		speed := 0.02
		if n.Archetype != nil {
			speed = n.Archetype.Stats.Speed
			if speed == 0 {
				speed = 0.02
			}
		}

		vx := (dx / dist) * speed
		vy := (dy / dist) * speed

		if math.Abs(vx) > math.Abs(vy) {
			if vx > 0 {
				n.Facing = DirSE
			} else {
				n.Facing = DirNW
			}
		} else {
			if vy > 0 {
				n.Facing = DirSW
			} else {
				n.Facing = DirNE
			}
		}

		newX := n.X + vx
		newY := n.Y + vy

		// Simple collision avoidance for escort
		collision := false
		bounds := n.GetFootprint().Transformed(newX, newY)
		for _, o := range obstacles {
			if engine.CheckCollision(bounds, o.GetFootprint()) {
				collision = true
				break
			}
		}
		if !collision {
			n.X = newX
			n.Y = newY
		}
	} else {
		n.State = NPCIdle
	}
}
