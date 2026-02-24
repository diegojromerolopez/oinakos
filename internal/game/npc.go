package game

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"

	"oinakos/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

const (
	BehaviorWander BehaviorType = iota
	BehaviorPatrol
	BehaviorKnightHunter
	BehaviorNpcFighter
)

type NPC struct {
	X, Y           float64
	Config         *NPCConfig
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
	Behavior                   BehaviorType
	WanderDirX                 float64
	WanderDirY                 float64
	PatrolStartX, PatrolStartY float64
	PatrolEndX, PatrolEndY     float64
	PatrolHeading              bool // true = toward End, false = toward Start

	TargetNPC    *NPC
	TargetPlayer *Player
}

var npcNames = []string{
	"Grog", "Zog", "Bob", "Drok", "Gorak", "Mug", "Snarl", "Thrak", "Vrog", "Kurg",
	"Hicks", "Miller", "Cooper", "Smith", "Potter", "Baker", "Carter", "Fisher",
}

func NewNPC(x, y float64, config *NPCConfig, level int) *NPC {
	n := &NPC{
		X:      x,
		Y:      y,
		Config: config,
		State:  NPCIdle,
		Facing: DirSE,
		Level:  level,
	}

	n.Name = npcNames[rand.Intn(len(npcNames))]

	// Load from config
	n.Health = config.Stats.HealthMin + rand.Intn(config.Stats.HealthMax-config.Stats.HealthMin+1)
	n.BaseAttack = config.Stats.BaseAttack
	n.BaseDefense = config.Stats.BaseDefense
	n.Speed = config.Stats.Speed
	n.AttackCooldown = config.Stats.AttackCooldown
	n.Weapon = config.Weapon

	// Dynamic scaling based on level
	n.Health = n.calculateStat(n.Health, n.Level)
	n.MaxHealth = n.Health
	n.BaseAttack = n.calculateStat(n.BaseAttack, n.Level)
	n.BaseDefense = n.calculateStat(n.BaseDefense, n.Level)

	// Initialize random behavior
	n.Behavior = BehaviorType(rand.Intn(4))

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
	if n.Config == nil {
		return false
	}
	nFootprint := n.Config.GetFootprint().Transformed(newX, newY)
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
	if n.Config == nil {
		return engine.Polygon{}
	}
	return n.Config.GetFootprint().Transformed(n.X, n.Y)
}

func (n *NPC) Update(player *Player, obstacles []*Obstacle, allNPCs []*NPC, fts *[]*FloatingText) {
	if n.State == NPCDead {
		return
	}

	n.Tick++

	// Custom Escort Logic
	if n.Config != nil && n.Config.ID == "escort" {
		n.updateEscort(player, obstacles)
		return
	}

	if n.State == NPCAttacking {
		n.AttackTimer++
	}

	var targetX, targetY float64
	var hasTarget bool
	var isTargetPlayer bool

	// Behavior Logic
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
			targetX, targetY = player.X, player.Y
			hasTarget = true
			isTargetPlayer = true
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
	if dist < attackRange {
		if n.State != NPCAttacking && isTargetPlayer {
			// Chance to say a menace line when starting to attack the player
			if rand.Float64() < 0.3 {
				msgNum := rand.Intn(5) + 1
				switch n.Config.ID {
				case "orc":
					engine.PlaySound(fmt.Sprintf("orc_menace_%d", msgNum))
				case "demon":
					engine.PlaySound(fmt.Sprintf("demon_menace_%d", msgNum))
				case "peasant":
					engine.PlaySound(fmt.Sprintf("peasant_menace_%d", msgNum))
				}
			}
		}
		n.State = NPCAttacking
		if n.AttackTimer >= n.AttackCooldown {
			n.AttackTimer = 0
			// HIT ROLL
			var targetAttack, targetDefense int
			var targetProtection int

			if isTargetPlayer {
				targetAttack = n.GetTotalAttack()
				targetDefense = player.GetTotalDefense()
				targetProtection = player.GetTotalProtection()
				// We need a wrapper or cast if we want to use the same interface,
				// but for now let's just do it directly.
				hitChance := targetAttack - targetDefense
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
					player.TakeDamage(finalDmg)

					*fts = append(*fts, &FloatingText{
						Text:  fmt.Sprintf("%d", finalDmg),
						X:     player.X,
						Y:     player.Y,
						Life:  45,
						Color: color.RGBA{255, 100, 100, 255},
					})
				} else {
					// MISS
					*fts = append(*fts, &FloatingText{
						Text:  "MISS",
						X:     player.X,
						Y:     player.Y,
						Life:  45,
						Color: color.RGBA{220, 220, 220, 255},
					})
				}
			} else if n.TargetNPC != nil {
				targetAttack = n.GetTotalAttack()
				targetDefense = n.TargetNPC.GetTotalDefense()
				targetProtection = n.TargetNPC.GetTotalProtection()

				hitChance := targetAttack - targetDefense
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
					n.TargetNPC.TakeDamage(finalDmg, nil, n)

					*fts = append(*fts, &FloatingText{
						Text:  fmt.Sprintf("%d", finalDmg),
						X:     n.TargetNPC.X,
						Y:     n.TargetNPC.Y,
						Life:  45,
						Color: color.RGBA{255, 255, 255, 255},
					})
				} else {
					// MISS
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
	}
}

func (n *NPC) TakeDamage(amount int, attackerPlayer *Player, attackerNPC *NPC) {
	if n.State == NPCDead {
		return
	}
	n.Health -= amount

	// Retaliation tracking
	if attackerPlayer != nil {
		n.TargetPlayer = attackerPlayer
		n.TargetNPC = nil
	} else if attackerNPC != nil {
		n.TargetNPC = attackerNPC
		n.TargetPlayer = nil
	}

	// Play hit sound
	switch n.Config.ID {
	case "orc":
		engine.PlaySound("orc_hit")
	case "demon":
		engine.PlaySound("demon_hit")
	case "peasant":
		engine.PlaySound("peasant_hit")
	}

	if n.Health <= 0 {
		n.State = NPCDead
		if attackerPlayer != nil {
			attackerPlayer.Kills++
			// Award XP based on NPC type
			switch n.Config.ID {
			case "orc":
				attackerPlayer.XP += 8 + rand.Intn(8) // 8-15
			case "demon":
				attackerPlayer.XP += 10 + rand.Intn(11) // 10-20
			case "peasant":
				attackerPlayer.XP += 2 + rand.Intn(5) // 2-6
			}
		}
		switch n.Config.ID {
		case "orc":
			engine.PlaySound("orc_death")
		case "demon":
			engine.PlaySound("demon_death")
		case "peasant":
			engine.PlaySound("peasant_death")
		}
	}
}

func (n *NPC) IsAlive() bool {
	return n.State != NPCDead
}

func (n *NPC) updateEscort(player *Player, obstacles []*Obstacle) {
	dx := player.X - n.X
	dy := player.Y - n.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	// Follow player if further than 5 units, but don't get closer than 3
	if dist > 5.0 {
		n.State = NPCWalking
		speed := n.Config.Stats.Speed
		if speed == 0 {
			speed = 0.02
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

func (n *NPC) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if n.Config == nil {
		return
	}

	isoX, isoY := engine.CartesianToIso(n.X, n.Y)

	drawSprite := n.Config.StaticImage
	if n.State == NPCDead {
		drawSprite = n.Config.CorpseImage
	} else if n.State == NPCAttacking && n.Config.AttackImage != nil {
		drawSprite = n.Config.AttackImage
	}

	if drawSprite == nil {
		return
	}

	w, h := drawSprite.Size()
	op := &ebiten.DrawImageOptions{}
	scale := 0.25

	flip := 1.0
	if n.Facing == DirSE || n.Facing == DirNE {
		flip = -1.0
	}

	op.GeoM.Scale(scale*flip, scale)

	// Anchoring: if flipped, we need to translate differently
	tx := isoX + offsetX
	if flip < 0 {
		tx += float64(w) * scale / 2
	} else {
		tx -= float64(w) * scale / 2
	}

	ty := isoY + offsetY - float64(h)*scale*0.85

	// Procedural Animation Overrides
	if n.State == NPCDead {
		// Lie flat on the ground
		ty = isoY + offsetY - float64(h)*scale*0.5
	} else if n.State == NPCWalking {
		// Bobbing effect
		bob := math.Sin(float64(n.Tick)*0.2) * 2.0
		ty += bob
	} else if n.State == NPCAttacking {
		// Lunge effect
		lungeAmt := 0.0
		attackPhase := float64(n.AttackTimer) / float64(n.AttackCooldown)
		if attackPhase < 0.2 { // Quick forward lunge
			lungeAmt = (attackPhase / 0.2) * 5.0
		} else if attackPhase < 0.5 { // Hold slightly, then pull back
			lungeAmt = 5.0 - ((attackPhase-0.2)/0.3)*5.0
		}

		if flip < 0 {
			tx += lungeAmt // Lunge right
		} else {
			tx -= lungeAmt // Lunge left
		}
	}

	op.GeoM.Translate(tx, ty)

	screen.DrawImage(drawSprite, op)

	// Only draw UI for living NPCs
	if n.IsAlive() {
		// Draw Name at feet
		ebitenutil.DebugPrintAt(screen, n.Name, int(isoX+offsetX-20), int(isoY+offsetY+5))

		// Draw Health Bar above NPC
		barWidth := 40.0
		barHeight := 4.0
		bx := isoX + offsetX - barWidth/2
		by := isoY + offsetY - float64(h)*scale*0.9 // Floating above head area

		vector.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth), float32(barHeight), color.RGBA{100, 0, 0, 255}, false)
		hpFrac := float32(n.Health) / float32(n.MaxHealth)
		if hpFrac > 0 {
			vector.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth)*hpFrac, float32(barHeight), color.RGBA{0, 255, 0, 255}, false)
		}
	}
}
