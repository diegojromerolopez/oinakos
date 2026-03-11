package game

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"

	"oinakos/internal/engine"
)

// NPCType is kept for legacy compatibility if needed,
// but ConfigID is preferred for dynamic loading.
type NPCType string

// NPCState, BehaviorType, Alignment, Direction and their constants are defined in actor.go


type NPC struct {
	Actor // Embedded shared state

	// Archetype is a convenience pointer to Actor.Config (which is *EntityConfig = *Archetype).
	// It exists to avoid updating hundreds of n.Archetype references throughout the codebase.
	Archetype      *Archetype

	AttackCooldown int // frames between attacks
	AttackTimer    int

	// Behavior fields
	Behavior   BehaviorType
	WanderDirX float64
	WanderDirY float64
	PatrolStartX, PatrolStartY float64
	PatrolEndX, PatrolEndY     float64
	PatrolHeading              bool // true = toward End, false = toward Start

	TargetActor *Actor
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
		Actor: Actor{
			X:           x,
			Y:           y,
			Config:      archetype,
			State:       NPCIdle,
			Facing:      DirSE,
			Level:       level,
			Alignment:   AlignmentEnemy, // Default to Enemy
			Group:       archetype.Group,
			LeaderID:    archetype.LeaderID,
			MustSurvive: archetype.MustSurvive,
		},
		Archetype: archetype,
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
	case "escort":
		n.Behavior = BehaviorEscort
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

// GetTotalAttack, GetTotalDefense, GetTotalProtection, calculateStat,
// checkCollisionAt, GetFootprint, Heal, IsAlive are inherited from the embedded Actor.

// NPC overrides GetFootprint to use n.Archetype if available.
func (n *NPC) GetFootprint() engine.Polygon {
	if n.Archetype == nil {
		return engine.Polygon{}
	}
	return n.Archetype.GetFootprint().Transformed(n.X, n.Y)
}

// NPC overrides checkCollisionAt to use n.Archetype.
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


func (n *NPC) Update(playableCharacter *PlayableCharacter, obstacles []*Obstacle, allNPCs []*NPC, projectiles *[]*Projectile, fts *[]*FloatingText, mapW, mapH float64, audio AudioManager) {
	n.Tick++

	if n.HitTimer > 0 {
		n.HitTimer--
	}

	var playerDist float64
	if playableCharacter != nil {
		playerDist = math.Sqrt(math.Pow(n.X-playableCharacter.X, 2) + math.Pow(n.Y-playableCharacter.Y, 2))
	}

	if n.State == NPCDead {
		if n.DeadTimer == 0 {
			// FIRST TICK OF DEATH: Ensure position is safe for the corpse.
			if n.Archetype != nil {
				n.X, n.Y = findSafePosition(n.X, n.Y, n.Archetype.GetFootprint(), obstacles)
			}
		}
		n.DeadTimer++
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
		// Allies should stay near the player but fight nearby enemies
		const followDistThreshold = 8.0 // Dist at which they stop fighting and rejoin player
		const rejoinDistTarget = 3.0   // Ideal distance to maintain from player
		const enemyDetectionRange = 12.0

		if playableCharacter != nil && playerDist > followDistThreshold {
			// Player is too far, stop whatever we are doing and rejoin
			n.TargetActor = &playableCharacter.Actor
			targetX, targetY = playableCharacter.X, playableCharacter.Y
			hasTarget = true
			isTargetPlayer = true
		} else {
			// Close enough to consider fighting
			nearestEnemy := gNearestEnemy(n, playableCharacter, allNPCs, enemyDetectionRange)
			if nearestEnemy != nil {
				n.TargetActor = &nearestEnemy.Actor
				targetX, targetY = nearestEnemy.X, nearestEnemy.Y
				hasTarget = true
			} else if playableCharacter != nil && playerDist > rejoinDistTarget {
				// No enemies, stay close to player
				n.TargetActor = &playableCharacter.Actor
				targetX, targetY = playableCharacter.X, playableCharacter.Y
				hasTarget = true
				isTargetPlayer = true
			} else {
				// Close enough and no enemies
				n.TargetActor = nil
				n.State = NPCIdle
				return
			}
		}
	} else if n.Alignment == AlignmentNeutral {
		if n.Behavior == BehaviorFlee {
			if n.TargetActor != nil && n.TargetActor.IsAlive() {
				if playableCharacter != nil && n.TargetActor == &playableCharacter.Actor {
					isTargetPlayer = true
				}
				dx := n.X - n.TargetActor.X
				dy := n.Y - n.TargetActor.Y
				mag := math.Sqrt(dx*dx + dy*dy)
				if mag > 0 {
					targetX, targetY = n.X+(dx/mag)*5.0, n.Y+(dy/mag)*5.0
				} else {
					targetX, targetY = n.X+1, n.Y+1
				}
				hasTarget = true
			} else {
				// No target to flee from, so wander
				n.Behavior = BehaviorWander
				n.TargetActor = nil
			}
		}

		if !hasTarget {
			// Strictly wander, ignore player and NPCs
			n.TargetActor = nil
			if n.Tick%120 == 0 {
				n.WanderDirX = rand.Float64()*2 - 1
				n.WanderDirY = rand.Float64()*2 - 1
			}
			targetX, targetY = n.X+n.WanderDirX, n.Y+n.WanderDirY
			hasTarget = true
		}
	}

	// Check for leader death
	if n.LeaderID != "" && n.Alignment != AlignmentNeutral {
		leaderAlive := false
		for _, other := range allNPCs {
			if other.Archetype != nil && other.Archetype.ID == n.LeaderID && other.IsAlive() {
				leaderAlive = true
				break
			}
		}
		if !leaderAlive {
			DebugLog("NPC [%s] becomes NEUTRAL because leader [%s] died!", n.Name, n.LeaderID)
			n.Alignment = AlignmentNeutral
			n.TargetActor = nil
			n.Behavior = BehaviorWander
		}
	}

	// Reassess target for certain behaviors
	if n.Behavior == BehaviorChaotic || n.Behavior == BehaviorNpcFighter {
		// Clear existing target to force reassessment in the switch below
		n.TargetActor = nil
	}

	if !hasTarget {
		// Behavior Logic (Traditional Enemy behavior)
		if n.TargetActor != nil && n.TargetActor.IsAlive() {
			targetX, targetY = n.TargetActor.X, n.TargetActor.Y
			hasTarget = true
			if playableCharacter != nil && n.TargetActor == &playableCharacter.Actor {
				isTargetPlayer = true
			}
		} else {
			switch n.Behavior {
			case BehaviorKnightHunter:
				if playableCharacter != nil && playableCharacter.IsAlive() {
					n.TargetActor = &playableCharacter.Actor
					targetX, targetY = playableCharacter.X, playableCharacter.Y
					hasTarget = true
					isTargetPlayer = true
				}
			case BehaviorNpcFighter:
				// Find nearest living NPC that isn't me
				var minDist = 999.0
				var nearestNPC *NPC
				for _, other := range allNPCs {
					if other == n || !other.IsAlive() {
						continue
					}
					// Only fight NPCs that are enemies
					if !n.isEnemy(other, allNPCs) {
						continue
					}
					dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
					if dist < minDist {
						minDist = dist
						nearestNPC = other
					}
				}
				if nearestNPC != nil {
					n.TargetActor = &nearestNPC.Actor
					targetX, targetY = nearestNPC.X, nearestNPC.Y
					hasTarget = true
				}
			case BehaviorChaotic:
				// Find nearest living actor (NPC or PlayableCharacter) that isn't me
				var minDist = 999.0
				var nearestActor *Actor
				var playerDist = 999.0
				if playableCharacter != nil {
					playerDist = math.Sqrt(math.Pow(n.X-playableCharacter.X, 2) + math.Pow(n.Y-playableCharacter.Y, 2))
				}

				for _, other := range allNPCs {
					if other == n || !other.IsAlive() {
						continue
					}
					dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
					if dist < minDist {
						minDist = dist
						nearestActor = &other.Actor
					}
				}

				if playableCharacter != nil && playableCharacter.IsAlive() && playerDist < minDist {
					n.TargetActor = &playableCharacter.Actor
					targetX, targetY = playableCharacter.X, playableCharacter.Y
					hasTarget = true
					isTargetPlayer = true
				} else if nearestActor != nil {
					n.TargetActor = nearestActor
					targetX, targetY = nearestActor.X, nearestActor.Y
					hasTarget = true
				}
			case BehaviorEscort:
				if playableCharacter != nil {
					// Follow player, but attack nearest enemy if in range
					const seekEnemyRange = 12.0
					nearestEnemy := gNearestEnemy(n, playableCharacter, allNPCs, seekEnemyRange)

					if nearestEnemy != nil {
						n.TargetActor = &nearestEnemy.Actor
						targetX, targetY = nearestEnemy.X, nearestEnemy.Y
						hasTarget = true
					} else {
						// Follow player if too far
						if playerDist > 3.0 {
							n.TargetActor = &playableCharacter.Actor
							targetX, targetY = playableCharacter.X, playableCharacter.Y
							hasTarget = true
							isTargetPlayer = true
						} else {
							// Stay idle if near player and no enemies
							n.State = NPCIdle
							return
						}
					}
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
					targetX, targetY = playableCharacter.X, playableCharacter.Y
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

		// Check alignment before attacking
		canAttack := false
		if n.TargetActor != nil {
			if n.Alignment != n.TargetActor.Alignment {
				canAttack = true
			}
		}

		if canAttack {
			if n.State != NPCAttacking && isTargetPlayer {
				// Chance to say an attack line when starting to attack the playableCharacter
				if rand.Float64() < 0.3 {
					if audio != nil && n.Archetype != nil {
						audio.PlayRandomSound(n.Archetype.SoundID + "/attack")
					}
				}
			}
			n.State = NPCAttacking
		} else {
			// If it's an ally near the target (like follow player), just stay idle or walk
			n.State = NPCIdle
			return
		}
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
					targetProtection = playableCharacter.GetTotalProtection()
					// We need a wrapper or cast if we want to use the same interface,
					// but for now let's just do it directly.
					attk := float64(n.GetTotalAttack())
					def := float64(playableCharacter.GetTotalDefense())
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
						DebugLog("NPC [%s] attacks Player: HIT for %d damage (roll: %d/%d)", n.Name, finalDmg, roll, hitChance)
						playableCharacter.TakeDamage(finalDmg, audio)

						*fts = append(*fts, &FloatingText{
							Text:  fmt.Sprintf("-%d", finalDmg),
							X:     playableCharacter.X,
							Y:     playableCharacter.Y,
							Life:  45,
							Color: ColorHarm,
						})
					} else {
						// MISS
						DebugLog("NPC [%s] attacks Player: MISS (roll: %d/%d)", n.Name, roll, hitChance)
						*fts = append(*fts, &FloatingText{
							Text:  "MISS",
							X:     playableCharacter.X,
							Y:     playableCharacter.Y,
							Life:  45,
							Color: ColorMiss,
						})
					}
				} else {
					// NPC VS NPC
					var targetActor *Actor
					if n.TargetActor != nil && n.TargetActor.IsAlive() && n.TargetActor != &playableCharacter.Actor {
						targetActor = n.TargetActor
					}

					if targetActor == nil {
						n.TargetActor = nil
						return
					}
					targetProtection = targetActor.GetTotalProtection()

					attk := float64(n.GetTotalAttack())
					def := float64(targetActor.GetTotalDefense())
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
						DebugLog("NPC [%s] attacks [%s]: HIT for %d damage (roll: %d/%d)", n.Name, targetActor.Name, finalDmg, roll, hitChance)
						
						// We need a way to call TakeDamage on the target. 
						// For now, let's find the NPC if it is one.
						var targetNPC *NPC
						for _, other := range allNPCs {
							if &other.Actor == targetActor {
								targetNPC = other
								break
							}
						}
						if targetNPC != nil {
							targetNPC.TakeDamage(finalDmg, nil, n, audio, allNPCs)
						}

						*fts = append(*fts, &FloatingText{
							Text:  fmt.Sprintf("-%d", finalDmg),
							X:     targetActor.X,
							Y:     targetActor.Y,
							Life:  45,
							Color: ColorHarm,
						})
					} else {
						// MISS
						DebugLog("NPC [%s] attacks [%s]: MISS (roll: %d/%d)", n.Name, targetActor.Name, roll, hitChance)
						*fts = append(*fts, &FloatingText{
							Text:  "MISS",
							X:     targetActor.X,
							Y:     targetActor.Y,
							Life:  45,
							Color: ColorMiss,
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

		if n.Tick%120 == 0 {
			DebugLog("NPC [%s] Moved to (%.2f, %.2f) | State: %v", n.Name, n.X, n.Y, n.State)
		}
	}

	// ALWAYS clamp to map boundaries
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
}

func (n *NPC) TakeDamage(amount int, attackerPlayer *PlayableCharacter, attackerNPC *NPC, audio AudioManager, allNPCs []*NPC) {
	if n.State == NPCDead {
		return
	}
	oldHealth := n.Health
	n.Health -= amount
	DebugLog("NPC Hit! [%s] Name: %s | Damage: %d | Health: %d -> %d", n.Alignment, n.Name, amount, oldHealth, n.Health)

	// Retaliation tracking
	var attackerHealth int
	var attacker *Actor
	if attackerPlayer != nil {
		attackerHealth = attackerPlayer.Health
		attacker = &attackerPlayer.Actor
	} else if attackerNPC != nil {
		attackerHealth = attackerNPC.Health
		attacker = &attackerNPC.Actor
	}

	if attacker != nil {
		n.TargetActor = attacker
		
		if float64(n.Health) < float64(attackerHealth)*0.2 {
			DebugLog("NPC [%s] is fleeing from [%s]! (Health: %d vs %d)", n.Name, attacker.Name, n.Health, attackerHealth)
			n.Alignment = AlignmentNeutral
			n.Behavior = BehaviorFlee
		} else {
			if n.Alignment != AlignmentEnemy || n.Behavior != BehaviorKnightHunter {
				DebugLog("NPC [%s] becomes AGGRESSIVE towards [%s]! (Health: %d vs %d)", n.Name, attacker.Name, n.Health, attackerHealth)
			}
			n.Alignment = AlignmentEnemy
			n.Behavior = BehaviorKnightHunter
			
			// GROUP ALERT: Alert all members of the same group in the "same zone" (e.g. 20 units)
			if n.Group != "" {
				for _, other := range allNPCs {
					if other == n || other.Alignment == AlignmentEnemy || !other.IsAlive() {
						continue
					}
					if other.Group == n.Group {
						dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
						if dist < 20.0 {
							DebugLog("NPC [%s] joining fight of group [%s]!", other.Name, n.Group)
							other.Alignment = AlignmentEnemy
							other.Behavior = BehaviorKnightHunter
							other.TargetActor = attacker
						}
					}
				}
			}
		}
	}

	// Start blood effect
	n.HitTimer = 30
	// Play hit sound
	if audio != nil && n.Archetype != nil {
		audio.PlayRandomSound(n.Archetype.SoundID + "/hit")
	}

	if n.Health <= 0 {
		DebugLog("NPC [%s] has been killed at (%.2f, %.2f)!", n.Name, n.X, n.Y)
		n.State = NPCDead
		if attackerPlayer != nil {
			attackerPlayer.Kills++
			if n.Archetype != nil && n.Archetype.ID != "" {
				attackerPlayer.MapKills[n.Archetype.ID]++
			}
			// Award XP from YAML-defined archetype value
			if n.Archetype != nil && n.Archetype.XP > 0 {
				attackerPlayer.AddXP(n.Archetype.XP)
			} else {
				// Fallback: 1 XP so every kill counts
				attackerPlayer.AddXP(1)
			}
		}
		if audio != nil && n.Archetype != nil {
			audio.PlayRandomSound(n.Archetype.SoundID + "/death")
		}
	}
}

// IsAlive is inherited from the embedded Actor.

func (n *NPC) isEnemy(other *NPC, allNPCs []*NPC) bool {
	// Traitor check: if 'other' has a leader whose alignment matches MINE,
	// but 'other' itself has switched to a different alignment, 'other' is a traitor.
	if other.LeaderID != "" {
		for _, potentialLeader := range allNPCs {
			if potentialLeader.Archetype != nil && potentialLeader.Archetype.ID == other.LeaderID && potentialLeader.IsAlive() {
				if other.Alignment != potentialLeader.Alignment && n.Alignment == potentialLeader.Alignment {
					return true
				}
				break
			}
		}
	}

	// Neutrals are never enemies to anyone, and don't have enemies
	// (Unless they were caught as traitors above)
	if n.Alignment == AlignmentNeutral || other.Alignment == AlignmentNeutral {
		return false
	}

	if n.Alignment != other.Alignment {
		return true
	}

	return false
}
func gNearestEnemy(n *NPC, playableCharacter *PlayableCharacter, allNPCs []*NPC, maxRange float64) *NPC {
	var nearest *NPC
	minDist := maxRange
	for _, other := range allNPCs {
		if other == n || !other.IsAlive() {
			continue
		}
		if n.isEnemy(other, allNPCs) {
			dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
			if dist < minDist {
				minDist = dist
				nearest = other
			}
		}
	}
	return nearest
}
