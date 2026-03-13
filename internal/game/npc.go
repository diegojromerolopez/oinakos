package game

import (
	"fmt"
	"math"
	"math/rand"
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

func (n *NPC) executeAttack(playableCharacter *PlayableCharacter, allNPCs []*NPC, projectiles *[]*Projectile, fts *[]*FloatingText, audio AudioManager, isTargetPlayer bool, dx, dy, dist float64, logFunc func(string, LogCategory), archs *ArchetypeRegistry) {
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
						targetNPC.TakeDamage(finalDmg, nil, n, audio, allNPCs, archs, logFunc)
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

func (n *NPC) Update(playableCharacter *PlayableCharacter, obstacles []*Obstacle, allNPCs []*NPC, projectiles *[]*Projectile, fts *[]*FloatingText, mapW, mapH float64, audio AudioManager, logFunc func(string, LogCategory), archs *ArchetypeRegistry) {
	n.Tick++
	if n.HitTimer > 0 { n.HitTimer-- }
	var playerDist float64
	if playableCharacter != nil {
		playerDist = math.Sqrt(math.Pow(n.X-playableCharacter.X, 2) + math.Pow(n.Y-playableCharacter.Y, 2))
	}

	if n.State == NPCDead {
		if n.DeadTimer == 0 && n.Archetype != nil {
			n.X, n.Y = findSafePosition(n.X, n.Y, n.Archetype.GetFootprint(), obstacles)
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
			n.executeMovement(dx, dy, obstacles, true)
		} else {
			n.executeAttack(playableCharacter, allNPCs, projectiles, fts, audio, isTargetPlayer, dx, dy, dist, logFunc, archs)
		}
	} else {
		n.executeMovement(dx, dy, obstacles, false)
	}

	// ALWAYS clamp to map boundaries
	halfW, halfH := mapW/2, mapH/2
	if n.X < -halfW { n.X = -halfW }
	if n.X > halfW { n.X = halfW }
	if n.Y < -halfH { n.Y = -halfH }
	if n.Y > halfH { n.Y = halfH }
}

func (n *NPC) TakeDamage(amount int, attackerPlayer *PlayableCharacter, attackerNPC *NPC, audio AudioManager, allNPCs []*NPC, archs *ArchetypeRegistry, logFunc func(string, LogCategory)) {
	if n.State == NPCDead { return }
	n.Health -= amount
	n.HitTimer = 30
	if audio != nil && n.Archetype != nil { audio.PlayRandomSound(n.Archetype.SoundID + "/hit") }

	var attacker *Actor
	attackerHealth := 0
	if attackerPlayer != nil {
		attacker = &attackerPlayer.Actor
		attackerHealth = attackerPlayer.Health
	} else if attackerNPC != nil {
		attacker = &attackerNPC.Actor
		attackerHealth = attackerNPC.Health
	}

	if attacker != nil {
		n.TargetActor = attacker
		if float64(n.Health) < float64(attackerHealth)*0.2 {
			n.Alignment = AlignmentNeutral
			n.Behavior = BehaviorFlee
		} else {
			n.Alignment = AlignmentEnemy
			n.Behavior = BehaviorKnightHunter
			if n.Group != "" {
				for _, other := range allNPCs {
					if other == n || other.Alignment == AlignmentEnemy || !other.IsAlive() || other.Group != n.Group { continue }
					if math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2)) < 20.0 {
						other.Alignment = AlignmentEnemy
						other.Behavior = BehaviorKnightHunter
						other.TargetActor = attacker
					}
				}
			}
		}
	}

	if n.Health <= 0 {
		// Vampire Conversion Logic
		var infectorConfig *EntityConfig
		if attackerPlayer != nil && attackerPlayer.Config != nil && attackerPlayer.Config.IsVampire() {
			infectorConfig = attackerPlayer.Config
		} else if attackerNPC != nil && attackerNPC.Archetype != nil && attackerNPC.Archetype.IsVampire() {
			infectorConfig = attackerNPC.Archetype
		}

		if infectorConfig != nil && n.Archetype != nil && n.Archetype.IsConvertibleHuman() {
			if rand.Float64() < infectorConfig.Stats.InfectingProbability {
				vampID := "vampire_male"
				if infectorConfig.Gender == "female" {
					vampID = "vampire_female"
				}

				if archs != nil {
					if newArch, ok := archs.Archetypes[vampID]; ok {
						n.Archetype = newArch
						n.Actor.Config = newArch
						n.Health = newArch.Stats.HealthMin
						n.MaxHealth = n.Health
						n.State = NPCIdle
						n.HitTimer = 0
						n.TargetActor = nil

						if attackerPlayer != nil {
							n.Alignment = AlignmentAlly
						} else if attackerNPC != nil {
							n.Alignment = attackerNPC.Alignment
						}

						if logFunc != nil {
							logFunc(fmt.Sprintf("%s was converted into a vampire by %s!", n.Name, infectorConfig.Name), LogCombatRecovery)
						}
						if audio != nil {
							audio.PlayRandomSound("vampire/convert") // Generic name for now
						}
						return
					}
				}
			}
		}

		n.State = NPCDead
		if attackerPlayer != nil {
			attackerPlayer.Kills++
			if n.Archetype != nil {
				attackerPlayer.MapKills[n.Archetype.ID]++
				if n.Archetype.XP > 0 { attackerPlayer.AddXP(n.Archetype.XP) } else { attackerPlayer.AddXP(1) }
			}
		}
		if audio != nil && n.Archetype != nil { audio.PlayRandomSound(n.Archetype.SoundID + "/death") }
	}
}
