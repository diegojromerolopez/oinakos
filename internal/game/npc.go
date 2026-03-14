package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"oinakos/internal/engine"
)

// NPCType is kept for legacy compatibility if needed,
// but ConfigID is preferred for dynamic loading.
type NPCType string

type NPC struct {
	Actor // Embedded shared state

	Archetype      *Archetype
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
}

var npcNames = []string{
	"Grog", "Zog", "Bob", "Drok", "Gorak", "Mug", "Snarl", "Thrak", "Vrog", "Kurg",
	"Hicks", "Miller", "Cooper", "Smith", "Potter", "Baker", "Carter", "Fisher",
}

func NewNPC(x, y float64, archetype *Archetype, level int) *NPC {
	if archetype == nil {
		archetype = &Archetype{ID: "unknown", Name: "Unknown Entity"}
		archetype.Stats.HealthMin = 10
		archetype.Stats.HealthMax = 10
		archetype.Weapon = WeaponTizon
	}
	n := &NPC{
		Actor: Actor{
			X: x, Y: y, Config: archetype, State: NPCIdle, Facing: DirSE, Level: level,
			Alignment: AlignmentEnemy, Group: archetype.Group, LeaderID: archetype.LeaderID, MustSurvive: archetype.MustSurvive,
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

	n.Health = archetype.Stats.HealthMin + rand.Intn(archetype.Stats.HealthMax-archetype.Stats.HealthMin+1)
	n.BaseAttack = archetype.Stats.BaseAttack
	n.BaseDefense = archetype.Stats.BaseDefense
	n.Speed = archetype.Stats.Speed
	n.AttackCooldown = archetype.Stats.AttackCooldown
	n.Weapon = archetype.Weapon

	switch archetype.Behavior {
	case "chaotic": n.Behavior = BehaviorChaotic
	case "fighter": n.Behavior = BehaviorNpcFighter
	case "hunter":  n.Behavior = BehaviorKnightHunter
	case "wander":  n.Behavior = BehaviorWander
	case "patrol":  n.Behavior = BehaviorPatrol
	case "escort":  n.Behavior = BehaviorEscort
	default:
		if archetype.Unique { n.Behavior = BehaviorWander } else { n.Behavior = BehaviorKnightHunter }
	}

	n.Health = n.calculateStat(n.Health, n.Level)
	n.MaxHealth = n.Health
	n.BaseAttack = n.calculateStat(n.BaseAttack, n.Level)
	n.BaseDefense = n.calculateStat(n.BaseDefense, n.Level)

	if archetype.Behavior == "" {
		n.Behavior = BehaviorType(rand.Intn(4))
	}

	if n.Behavior == BehaviorWander {
		n.WanderDirX = rand.Float64()*2 - 1
		n.WanderDirY = rand.Float64()*2 - 1
	} else if n.Behavior == BehaviorPatrol {
		n.PatrolStartX = n.X
		n.PatrolStartY = n.Y
		n.PatrolEndX = n.X + (rand.Float64()*8 - 4)
		n.PatrolEndY = n.Y + (rand.Float64()*8 - 4)
		n.PatrolHeading = true
	}
	return n
}

func (n *NPC) GetFootprint() engine.Polygon {
	if n.Archetype == nil { return engine.Polygon{} }
	return n.Archetype.GetFootprint().Transformed(n.X, n.Y)
}

func (n *NPC) checkCollisionAt(newX, newY float64, obstacles []*Obstacle) bool {
	if n.Archetype == nil { return false }
	nFootprint := n.Archetype.GetFootprint().Transformed(newX, newY)
	for _, o := range obstacles {
		if !o.Alive { continue }
		if engine.CheckCollision(nFootprint, o.GetFootprint()) { return true }
	}
	return false
}

func (n *NPC) executeAttack(playableCharacter *PlayableCharacter, allNPCs []*NPC, projectiles *[]*Projectile, obstacles *[]*Obstacle, obsRegistry *ObstacleRegistry, fts *[]*FloatingText, audio AudioManager, isTargetPlayer bool, dx, dy, dist float64, logFunc func(string, LogCategory), archs *ArchetypeRegistry) {
	if n.State != NPCAttacking && isTargetPlayer {
		if rand.Float64() < 0.1 {
			if n.Archetype != nil && n.Archetype.Dialogues != nil {
				bark := n.Archetype.Dialogues.PickCombatBark()
				if bark != "" && logFunc != nil {
					logFunc(fmt.Sprintf("%s: %s", n.Name, bark), LogNPC)
				}
			}
		}
		if rand.Float64() < 0.3 {
			if audio != nil && n.Archetype != nil {
				audio.PlayRandomSound(n.Archetype.SoundID + "/attack")
			}
		}
	}
	n.State = NPCAttacking

	if n.AttackTimer >= n.AttackCooldown {
		n.AttackTimer = 0
		attackRange := 1.0
		if n.Archetype != nil && n.Archetype.Stats.AttackRange > 1.0 {
			attackRange = n.Archetype.Stats.AttackRange
		}
		isRanged := attackRange > 1.0

		if isRanged {
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag > 0 {
				pSpd := n.Archetype.Stats.ProjectileSpeed
				if pSpd <= 0 { pSpd = 0.15 }
				proj := NewProjectile(n.X, n.Y, dx/mag, dy/mag, pSpd, n.GetTotalAttack(), false, 100.0)
				*projectiles = append(*projectiles, proj)
			}
		} else {
			if isTargetPlayer {
				targetProtection := playableCharacter.GetTotalProtection()
				attk := float64(n.GetTotalAttack())
				def := float64(playableCharacter.GetTotalDefense())
				if def <= 0 { def = 1 }
				hitChance := int(attk / (attk + def) * 100)
				if hitChance < 5 { hitChance = 5 }
				if hitChance > 95 { hitChance = 95 }

				if rand.Intn(100)+1 <= hitChance {
					rawDmg := n.Weapon.RollDamage()
					finalDmg := int(math.Max(1, float64(rawDmg-targetProtection)))
					playableCharacter.TakeDamage(finalDmg, audio)
					*fts = append(*fts, &FloatingText{
						Text: fmt.Sprintf("-%d", finalDmg), X: playableCharacter.X, Y: playableCharacter.Y, Life: 45, Color: ColorHarm,
					})
				} else {
					*fts = append(*fts, &FloatingText{
						Text: "MISS", X: playableCharacter.X, Y: playableCharacter.Y, Life: 45, Color: ColorMiss,
					})
				}
			} else if n.TargetActor != nil && n.TargetActor.IsAlive() {
				targetActor := n.TargetActor
				targetProtection := targetActor.GetTotalProtection()
				attk := float64(n.GetTotalAttack())
				def := float64(targetActor.GetTotalDefense())
				if def <= 0 { def = 1 }
				hitChance := int(attk / (attk + def) * 100)
				if hitChance < 5 { hitChance = 5 }
				if hitChance > 95 { hitChance = 95 }

				if rand.Intn(100)+1 <= hitChance {
					rawDmg := n.Weapon.RollDamage()
					finalDmg := int(math.Max(1, float64(rawDmg-targetProtection)))
					var targetNPC *NPC
					for _, other := range allNPCs {
						if &other.Actor == targetActor {
							targetNPC = other
							break
						}
					}
					if targetNPC != nil {
						targetNPC.TakeDamage(finalDmg, n, audio, allNPCs, archs, obstacles, obsRegistry, logFunc)
					}
					*fts = append(*fts, &FloatingText{
						Text: fmt.Sprintf("-%d", finalDmg), X: targetActor.X, Y: targetActor.Y, Life: 45, Color: ColorHarm,
					})
				} else {
					*fts = append(*fts, &FloatingText{
						Text: "MISS", X: targetActor.X, Y: targetActor.Y, Life: 45, Color: ColorMiss,
					})
				}
			}
		}
	}
}

func (n *NPC) Update(playableCharacter *PlayableCharacter, obstacles *[]*Obstacle, obsRegistry *ObstacleRegistry, allNPCs []*NPC, projectiles *[]*Projectile, fts *[]*FloatingText, mapW, mapH float64, audio AudioManager, logFunc func(string, LogCategory), archs *ArchetypeRegistry) {
	n.Tick++
	var worldObstacles []*Obstacle
	if obstacles != nil {
		worldObstacles = *obstacles
	}
	if n.HitTimer > 0 { n.HitTimer-- }
	var playerDist float64
	if playableCharacter != nil {
		playerDist = math.Sqrt(math.Pow(n.X-playableCharacter.X, 2) + math.Pow(n.Y-playableCharacter.Y, 2))
	}

	if n.State == NPCDead {
		if n.DeadTimer == 0 && n.Archetype != nil {
			n.X, n.Y = findSafePosition(n.X, n.Y, n.Archetype.GetFootprint(), worldObstacles)
		}
		n.DeadTimer++
		return
	}

	if n.State == NPCAttacking { n.AttackTimer++ }

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
			n.Alignment = AlignmentNeutral
			n.TargetActor = nil
			n.Behavior = BehaviorWander
		}
	}

	targetX, targetY, hasTarget, isTargetPlayer := n.findTarget(playableCharacter, allNPCs, playerDist)

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

	// Check alignment before attacking
	canAttack := false
	if n.TargetActor != nil && n.Alignment != n.TargetActor.Alignment {
		canAttack = true
	}

	if dist < attackRange && canAttack {
		if isRanged && dist < attackRange-2.0 {
			n.executeMovement(dx, dy, worldObstacles, true)
		} else {
			n.executeAttack(playableCharacter, allNPCs, projectiles, obstacles, obsRegistry, fts, audio, isTargetPlayer, dx, dy, dist, logFunc, archs)
		}
	} else {
		n.executeMovement(dx, dy, worldObstacles, false)
	}

	// ALWAYS clamp to map boundaries
	halfW, halfH := mapW/2, mapH/2
	if n.X < -halfW { n.X = -halfW }
	if n.X > halfW { n.X = halfW }
	if n.Y < -halfH { n.Y = -halfH }
	if n.Y > halfH { n.Y = halfH }
}

func (n *NPC) TakeDamage(amount int, attacker ActorInterface, audio AudioManager, allNPCs []*NPC, archs *ArchetypeRegistry, obstacles *[]*Obstacle, obsRegistry *ObstacleRegistry, logFunc func(string, LogCategory)) {
	if n.State == NPCDead {
		return
	}
	n.Health -= amount
	n.HitTimer = 30
	if audio != nil && n.Archetype != nil {
		audio.PlayRandomSound(n.Archetype.SoundID + "/hit")
	}

	var actor *Actor
	attackerHealth := 0
	if attacker != nil {
		actor = attacker.GetActor()
		attackerHealth = actor.Health
	}

	if actor != nil {
		n.TargetActor = actor
		if float64(n.Health) < float64(attackerHealth)*0.2 {
			n.Alignment = AlignmentNeutral
			n.Behavior = BehaviorFlee
		} else {
			n.Alignment = AlignmentEnemy
			n.Behavior = BehaviorKnightHunter
			if n.Group != "" {
				for _, other := range allNPCs {
					if other == n || other.Alignment == AlignmentEnemy || !other.IsAlive() || other.Group != n.Group {
						continue
					}
					if math.Sqrt(math.Pow(n.X-other.X, 2)+math.Pow(n.Y-other.Y, 2)) < 20.0 {
						other.Alignment = AlignmentEnemy
						other.Behavior = BehaviorKnightHunter
						other.TargetActor = actor
					}
				}
			}
		}
	}

	if n.Health <= 0 {
		// Generic Kill Actions System
		var attackerConfig *EntityConfig
		var attackerNPC *NPC
		var attackerPlayer *PlayableCharacter

		if attacker != nil {
			if npc, ok := attacker.(*NPC); ok {
				attackerNPC = npc
				attackerConfig = npc.Archetype
			} else if pc, ok := attacker.(*PlayableCharacter); ok {
				attackerPlayer = pc
				attackerConfig = pc.Config
			}
		}

		if attackerConfig != nil && attackerConfig.Actions != nil && len(attackerConfig.Actions.OnKill) > 0 {
			for _, action := range attackerConfig.Actions.OnKill {
				// Probability check
				if rand.Float64() >= action.Probability {
					continue
				}

				if action.Type == "transform_victim" {
					eff := action.Effect.Victim
					if eff != nil && eff.Transform != "" && archs != nil {
						targetArchID := interpolateConfigString(eff.Transform, n)
						if newArch, ok := archs.Archetypes[targetArchID]; ok {
							n.Archetype = newArch
							n.Actor.Config = newArch
							n.Health = newArch.Stats.HealthMin
							n.MaxHealth = n.Health
							n.State = NPCIdle
							n.HitTimer = 0
							n.TargetActor = nil

							// Default alignment logic
							if attackerPlayer != nil {
								n.Alignment = AlignmentAlly
							} else if attackerNPC != nil {
								n.Alignment = attackerNPC.Alignment
							}

							// Manual override if set
							if eff.Alignment == "inherit" {
								if attackerNPC != nil {
									n.Alignment = attackerNPC.Alignment
								} else if attackerPlayer != nil {
									n.Alignment = AlignmentAlly
								}
							}

							if logFunc != nil {
								logFunc(fmt.Sprintf("%s was transformed into %s!", n.Name, newArch.Name), LogCombatRecovery)
							}

							if eff.SpawnCorpse != nil && !*eff.SpawnCorpse {
								return // Suppress death processing
							}
						}
					}
				} else if action.Type == "heal_attacker" {
					if action.Effect.Attacker != nil && action.Effect.Attacker.Heal > 0 {
						attacker.Heal(action.Effect.Attacker.Heal)
					}
				} else if action.Type == "incinerate" {
					eff := action.Effect.Victim
					if eff != nil && eff.CorpseImage != "" && obstacles != nil && obsRegistry != nil {
						staticID := interpolateConfigString(eff.CorpseImage, n)
						if arch, ok := obsRegistry.Archetypes[staticID]; ok {
							newObs := NewObstacle(staticID, n.X, n.Y, arch)
							*obstacles = append(*obstacles, newObs)
						}
						if eff.SpawnCorpse != nil && !*eff.SpawnCorpse {
							n.State = NPCDead
							return // Suppress standard death
						}
					}
				}
			}
		}

		n.State = NPCDead
		if attackerPlayer != nil {
			attackerPlayer.Kills++
			if n.Archetype != nil {
				attackerPlayer.MapKills[n.Archetype.ID]++
				if n.Archetype.XP > 0 {
					attackerPlayer.AddXP(n.Archetype.XP)
				} else {
					attackerPlayer.AddXP(1)
				}
			}
		}
		if audio != nil && n.Archetype != nil {
			audio.PlayRandomSound(n.Archetype.SoundID + "/death")
		}
	}
}

func interpolateConfigString(s string, victim *NPC) string {
	res := s
	if victim != nil {
		res = strings.ReplaceAll(res, "{gender}", victim.Archetype.Gender)
		// Default to "male" if gender is missing
		if victim.Archetype.Gender == "" {
			res = strings.ReplaceAll(res, "{gender}", "male")
		}
	}
	return res
}
