package game

import (
	"math"

	"oinakos/internal/engine"
)

// ActorState is the unified state enum for all living entities.
type ActorState int

const (
	ActorIdle      ActorState = iota
	ActorWalking
	ActorAttacking
	ActorDead
	ActorDrinking // Player-specific (well interaction)
)

// Backward-compatible aliases for PlayableCharacterState
type PlayableCharacterState = ActorState

const (
	StateIdle     = ActorIdle
	StateWalking  = ActorWalking
	StateAttacking = ActorAttacking
	StateDead     = ActorDead
	StateDrinking = ActorDrinking
)

// Backward-compatible aliases for NPCState
type NPCState = ActorState

const (
	NPCIdle      = ActorIdle
	NPCWalking   = ActorWalking
	NPCAttacking = ActorAttacking
	NPCDead      = ActorDead
)

// Direction represents isometric facing direction.
type Direction int

const (
	DirSE Direction = iota
	DirSW
	DirNE
	DirNW
)

// Alignment represents faction membership.
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

// BehaviorType controls NPC AI decision-making.
type BehaviorType int

const (
	BehaviorWander BehaviorType = iota
	BehaviorPatrol
	BehaviorKnightHunter
	BehaviorNpcFighter
	BehaviorChaotic
	BehaviorEscort
)

// Actor holds all runtime state shared between the player character and any NPC.
type Actor struct {
	X, Y        float64
	Config      *EntityConfig
	Facing      Direction
	State       ActorState
	Tick        int
	Health      int
	MaxHealth   int
	BaseAttack  int
	BaseDefense int
	Speed       float64
	Weapon      *Weapon
	Alignment   Alignment
	Group       string
	LeaderID    string
	MustSurvive bool
	Level       int
	XP          int
	Name        string

	// Timers
	HitTimer  int // How long to show hit sprite (BloodTimer on NPC)
	DeadTimer int // Ticks since death

	// Equipment
	EquippedArmor map[ArmorSlot]*Armor
}

// IsAlive returns true if this actor is not in the Dead state.
func (a *Actor) IsAlive() bool {
	return a.State != ActorDead
}

// Heal increases health up to MaxHealth.
func (a *Actor) Heal(amount int) {
	if a.State == ActorDead {
		return
	}
	oldHealth := a.Health
	a.Health += amount
	if a.Health > a.MaxHealth {
		a.Health = a.MaxHealth
	}
	if a.Health > oldHealth {
		DebugLog("Actor Healed [%s] %s! +%d | Health: %d -> %d", a.Alignment, a.Name, amount, oldHealth, a.Health)
	}
}

// calculateStat applies logarithmic level scaling.
func (a *Actor) calculateStat(base, level int) int {
	if level <= 1 {
		return base
	}
	bonus := int(math.Log2(float64(level)) * 10)
	return base + bonus
}

// GetTotalAttack returns the level-scaled attack value.
func (a *Actor) GetTotalAttack() int {
	return a.calculateStat(a.BaseAttack, a.Level)
}

// GetTotalDefense returns the level-scaled defense value.
func (a *Actor) GetTotalDefense() int {
	return a.calculateStat(a.BaseDefense, a.Level)
}

// GetTotalProtection returns the sum of all equipped armor.
func (a *Actor) GetTotalProtection() int {
	total := 0
	for _, armor := range a.EquippedArmor {
		if armor != nil {
			total += armor.Protection
		}
	}
	return total
}

// GetFootprint returns the collision polygon for this actor.
// If Config has a custom footprint, use it; otherwise fall back to a default hexagon.
func (a *Actor) GetFootprint() engine.Polygon {
	if a.Config != nil && len(a.Config.Footprint) > 0 {
		return a.Config.GetFootprint().Transformed(a.X, a.Y)
	}
	// Default fallback hexagon
	return engine.Polygon{Points: []engine.Point{
		{X: -0.2, Y: -0.1}, {X: 0.2, Y: -0.1}, {X: 0.3, Y: 0},
		{X: 0.2, Y: 0.1}, {X: -0.2, Y: 0.1}, {X: -0.3, Y: 0},
	}}.Transformed(a.X, a.Y)
}

// checkCollisionAt tests whether moving this actor to (newX, newY) would collide with any obstacle.
func (a *Actor) checkCollisionAt(newX, newY float64, obstacles []*Obstacle) bool {
	var fp engine.Polygon
	if a.Config != nil && len(a.Config.Footprint) > 0 {
		fp = a.Config.GetFootprint().Transformed(newX, newY)
	} else {
		fp = engine.Polygon{Points: []engine.Point{
			{X: -0.2, Y: -0.1}, {X: 0.2, Y: -0.1}, {X: 0.3, Y: 0},
			{X: 0.2, Y: 0.1}, {X: -0.2, Y: 0.1}, {X: -0.3, Y: 0},
		}}.Transformed(newX, newY)
	}
	for _, o := range obstacles {
		if !o.Alive {
			continue
		}
		if engine.CheckCollision(fp, o.GetFootprint()) {
			return true
		}
	}
	return false
}

// AddXP adds experience and handles level-up logic.
func (a *Actor) AddXP(amount int) {
	a.XP += amount
	newLevel := a.XP/100 + 1
	if newLevel > a.Level {
		a.Level = newLevel
		a.Health = a.MaxHealth
	}
}
