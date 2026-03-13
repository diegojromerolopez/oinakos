package game

import (
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

func (n *NPC) Update(playableCharacter *PlayableCharacter, obstacles []*Obstacle, allNPCs []*NPC, projectiles *[]*Projectile, fts *[]*FloatingText, mapW, mapH float64, audio AudioManager, logFunc func(string, LogCategory)) {
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
			n.executeAttack(playableCharacter, allNPCs, projectiles, fts, audio, isTargetPlayer, dx, dy, dist, logFunc)
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

func (n *NPC) TakeDamage(amount int, attackerPlayer *PlayableCharacter, attackerNPC *NPC, audio AudioManager, allNPCs []*NPC) {
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
