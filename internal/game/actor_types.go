package game

import (
	"oinakos/internal/engine"
)

// ActorState is the unified state enum for all living entities.
type ActorState int
type LaborShift int

const (
	ShiftWork LaborShift = iota
	ShiftLeisure
	ShiftSleep
)

const (
	ActorIdle ActorState = iota
	ActorWalking
	ActorAttacking
	ActorDead
	ActorDrinking  // Player-specific (well interaction)
	ActorCrouching // Picking up items
	ActorIncapacitated // Down but not yet Truly Dead
	ActorResting   // Sleeping / Resting to regain energy
	ActorChopping  // Gathering timber
	ActorDigging   // Gathering ore/excavating
	ActorForaging  // Gathering fruits/veg from environment
	ActorEating    // Transitioning after consuming food
	ActorBerserk   // Psychotic break - hostile to all
	ActorCooking   // Preparing food at a campfire
	ActorWorkshop  // Repairing gear at a workbench
	ActorMilking   // Gathering milk from livestock
	ActorStashing  // Moving items to an owned chest
	ActorIntercourse // Sexual activity
	ActorBathing     // Taking a bath
	ActorRelieving   // Alleviating bodily needs
)

func (s ActorState) String() string {
	switch s {
	case ActorIdle:          return "Idle"
	case ActorWalking:       return "Walking"
	case ActorAttacking:     return "Attacking"
	case ActorDead:          return "Dead"
	case ActorDrinking:      return "Drinking"
	case ActorCrouching:     return "Crouching"
	case ActorIncapacitated: return "Incapacitated"
	case ActorResting:       return "Resting"
	case ActorChopping:      return "Chopping"
	case ActorDigging:       return "Digging"
	case ActorForaging:      return "Foraging"
	case ActorEating:        return "Eating"
	case ActorBerserk:       return "Berserk"
	case ActorCooking:       return "Cooking"
	case ActorWorkshop:      return "Workshop"
	case ActorMilking:       return "Milking"
	case ActorStashing:      return "Stashing"
	case ActorIntercourse:   return "Intercourse"
	case ActorBathing:       return "Bathing"
	case ActorRelieving:     return "Relieving"
	default:                 return "Unknown"
	}
}

// Backward-compatible aliases for PlayableCharacterState
type PlayableCharacterState = ActorState

// Backward-compatible aliases for NPCState
type NPCState = ActorState

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
	BehaviorFlee
	BehaviorTrader // NPC will open a trade window when interacted with
	BehaviorHauler // NPC will move items on the ground to stockpiles
	BehaviorLumberjack // NPC will chop trees and move wood to stockpiles
	BehaviorFarmer     // NPC will plant and harvest crops
	BehaviorArtisan    // NPC will smelt, craft, and repair at workbenches
)

// PhysicalTrauma tracks irreversible physical injuries.
type PhysicalTrauma struct {
	LeftArmLost   bool
	RightArmLost  bool
	LeftLegLost   bool
	RightLegLost  bool
	LeftHandLost  bool
	RightHandLost bool
	EyesLost      int  // 0, 1, or 2
	BurnedAlive   bool // Survivors of extreme fire
	SpineBroken   bool
}

// MemoryEvent tracks past interactions for AI decision making.
type MemoryEvent struct {
	Tick   int     `json:"tick"`
	Type   string  `json:"type"`   // "gift", "attack", "observed_kill"
	Source string  `json:"source"` // Actor ID
	Value  float64 `json:"value"`  // Sentiment change (+ or -)
}

type PrimaryAttributes struct {
	Strength           int     // Affects melee attacks, base_attack (0-100)
	Dexterity          int     // Affects ranged attacks, speed, cooldown (0-100)
	Health             int     // Affects max health points, thirst, hunger, base_defense (0-100)
	Intellect          int     // Persuasion, logic, planning (0-100)
	Wisdom             int     // Integrity, morale, code (0-100)
}

type EntityStats struct {
	HealthMin       int     `yaml:"health_min"`
	HealthMax       int     `yaml:"health_max"`
	HungerMax       float64 `yaml:"hunger_max"`
	ThirstMax       float64 `yaml:"thirst_max"`
	FatigueMax      float64 `yaml:"fatigue_max"`
	Speed           float64 `yaml:"speed"`
	BaseAttack      int     `yaml:"base_attack"`
	BaseDefense     int     `yaml:"base_defense"`
	BaseProtection  int     `yaml:"base_protection"`
	AttackCooldown  int     `yaml:"attack_cooldown"`
	AttackRange     float64 `yaml:"attack_range"`
	ProjectileSpeed float64 `yaml:"projectile_speed"`
	IsMilkable      bool    `yaml:"is_milkable"`
	MilkCooldown    int     `yaml:"milk_cooldown"`
	MaxWeight       float64       `yaml:"max_weight"`
	Age             AgeState      `yaml:"age"`
}

type AgeConfig struct {
	Current FloatInterval `yaml:"current"`
	Rate    float64       `yaml:"rate"`
	Max     float64       `yaml:"max"` // Max age character can reach (0 = infinite)
}

func (c AgeConfig) Roll() AgeState {
	return AgeState{
		Current: c.Current.Roll(),
		Rate:    c.Rate,
		Max:     c.Max,
	}
}

type TemporalState struct {
	HealthPoints     int     `yaml:"health_points,omitempty"`
	MaxHealthPoints  int
	Hunger           float64 `yaml:"hunger,omitempty"`
	Thirst           float64 `yaml:"thirst,omitempty"`
	Fatigue          float64 `yaml:"fatigue,omitempty"`
	Sanity           float64 `yaml:"sanity,omitempty"`
	IsPoisoned       bool    `yaml:"is_poisoned,omitempty"`
	IsAngry          bool    `yaml:"is_angry,omitempty"`
	IsConscious      bool    `yaml:"is_conscious,omitempty"`
	IsSeptic         bool    `yaml:"is_septic,omitempty"` // Wound infection
	IsSick           bool    `yaml:"is_sick,omitempty"`   // Common cold / stomach flu / etc.
	Arousal          float64 `yaml:"arousal,omitempty"`   // 0 to 100
	Pain             float64 `yaml:"pain,omitempty"`      // 0 to 100
	Hygiene          float64 `yaml:"hygiene,omitempty"`   // 100 to 0 (100 = clean, 0 = filthy)
	Miccionate       float64 `yaml:"miccionate,omitempty"` // 0 to 100 (100 = urgent)
	Defecate         float64 `yaml:"defecate,omitempty"`   // 0 to 100 (100 = urgent)
	AlcoholLevel     float64 `yaml:"alcohol_level,omitempty"`
	IsDrunk          bool    `yaml:"is_drunk,omitempty"`
	Age              AgeState `yaml:"age,omitempty"`
}

type AgeState struct {
	Current float64 `yaml:"current"` // Years
	Rate    float64 `yaml:"rate"`    // 1.0 normal, 0.0 vampire
	Max     float64 `yaml:"max"`     // Absolute death threshold
}

const (
	StageBaby     = "baby"
	StageKid      = "kid"
	StageTeenager = "teenager"
	StageAdult    = "peasant"
	StageElder    = "elder"
)

// Actor holds all runtime state shared between the player character and any NPC.
type Actor struct {
	X, Y   float64
	Z      float64
	VerticalVelocity float64
	TemporalState     TemporalState     `yaml:"state"`
	PrimaryAttributes PrimaryAttributes `yaml:"attributes"`
	RawStats          EntityStats       `yaml:"-"` // Rolled values from config
	SkillValues       map[string]int    `yaml:"skills,omitempty"`
	
	BodyStatus map[string]int // "head", "torso", "l_arm", "r_arm", "l_leg", "r_leg"
	Config *EntityConfig
	Facing Direction
	State  ActorState
	Trauma PhysicalTrauma
	
	// Navigation
	Path      []engine.Point
	PathTimer int // Ticks until next path recalculation
	
	// Mental Ticks
	WorkTicks     int
	LeisureTicks  int
	
	// Character Identity & Social
	ID               string
	Name             string
	Denarii          int
	Alignment        Alignment
	Group            string
	LeaderID         string
	MustSurvive      bool
	IsTransexual     bool
	ParentID         string // For offspring
	FatherID         string // For offspring
	
	// Life Stage & Aging
	AgeTicks         float64
	LifeStage        string // "baby", "kid", "teenager", "adult", "elder"
	
	// Combat Stats (derived)
	BaseAttack         int
	RangedAttack       int     
	BaseDefense        int
	BaseProtection     int
	BaseAttackCooldown int
	Speed              float64
	CriticalChance     float64
	
	// Bonus from items/buffs
	AttackBonus     int
	DefenseBonus    int
	ProtectionBonus int
	SpeedBonus      float64
	MaxHealthBonus  int
	RegenPerSecond  int
	
	// Thermodynamics & Illness
	BodyTemperature      float64
	PreferredTemperature float64
	FluTicks             int
	SicknessTicks        int    // General sickness duration
	Sickness             string // Type: "stomach sickness", etc.
	ContagionTimer       int
	
	// Reproductive & Relationships
	Relationships      map[string]float64
	RomanticInterest   map[string]float64
	Submission         map[string]float64
	Memories           []MemoryEvent
	IsPregnant         bool
	GestationTicks     int
	MatingCooldown     int
	ArousalTimer       int
	
	// Inventory & Equipment
	Inventory []*ItemInstance
	Slots     map[string]*ItemInstance
	MaxWeight float64
	Weapon    *Weapon
	BaseWeapon *Weapon
	
	// AI Decisions & State
	Shift              LaborShift
	Mood               MoodType
	LastAIChoice       string
	LastAIReasoning    string
	ThoughtTimer       int
	LastAIDecisionTick int
	LastTalkTick       int
	CurrentTile        string
	
	// Targets
	TargetActorID      string
	TargetObstacle     *Obstacle
	TargetItem         *ItemInstance
	TargetStockpile    *FloorZone
	
	// Productive Stats
	Nourishment int
	Survivalism int
	Mate        float64
	Crafting    int
	Herbalism   int
	Trading     int
	Harvesting  int
	Husbandry   int
	Art         int
	Culture     int
	
	// Counters & Progress
	Kills    int
	MapKills map[string]int
	Level    int
	XP       int
	
	// Ownership
	OwnedChestID      string
	MilkCooldownTicks int
	MeatQuantity      float64
	
	// Timers & Visual Flags
	Tick               int
	HitTimer           int
	DeadTimer          int
	CrouchTimer        int
	IsOccluded         bool
	IsTarget           bool
	IsConscious        bool // Mirroring TemporalState for easy access
	UnconsciousTimer   int
	
	LastReaction       string
	SelectedModel      string
}
