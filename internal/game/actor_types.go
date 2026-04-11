package game

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type Alignment int

const (
	AlignmentNeutral Alignment = iota
	AlignmentFriendly
	AlignmentHostile
	AlignmentEnemy = AlignmentHostile
	AlignmentAlly  = AlignmentFriendly
)

func (a *Alignment) UnmarshalYAML(v *yaml.Node) error {
	var i int
	if err := v.Decode(&i); err == nil {
		*a = Alignment(i)
		return nil
	}
	var s string
	if err := v.Decode(&s); err == nil {
		switch strings.ToLower(s) {
		case "neutral":
			*a = AlignmentNeutral
			return nil
		case "friendly", "ally":
			*a = AlignmentFriendly
			return nil
		case "hostile", "enemy":
			*a = AlignmentHostile
			return nil
		default:
			return fmt.Errorf("unknown alignment: %s", s)
		}
	}
	return fmt.Errorf("invalid alignment format")
}

type Direction int

const (
	DirS Direction = iota
	DirSE
	DirE
	DirNE
	DirN
	DirNW
	DirW
	DirSW
)

type ActorState int

const (
	ActorIdle ActorState = iota
	ActorWalking
	ActorAttacking
	ActorHit
	ActorDead
	ActorInteracting
	ActorHarvesting
	ActorResting
	ActorThinking
	ActorIncapacitated
	ActorFeeding
	ActorBerserk
    ActorCrouching
    ActorDrinking
    ActorEating
    ActorBathing
    ActorCooking
    ActorForaging
    ActorChopping
    ActorDigging
    ActorMilking
    ActorStashing
    ActorRelieving
    ActorWorkshop
    ActorIntercourse
    ActorSocializing
    ActorGambling
)

func (s ActorState) String() string {
    switch s {
    case ActorIdle: return "Idle"
    case ActorWalking: return "Walking"
    case ActorAttacking: return "Attacking"
    case ActorHit: return "Hit"
    case ActorDead: return "Dead"
    case ActorInteracting: return "Interacting"
    case ActorHarvesting: return "Harvesting"
    case ActorResting: return "Resting"
    case ActorThinking: return "Thinking"
    case ActorIncapacitated: return "Incapacitated"
    case ActorFeeding: return "Feeding"
    case ActorBerserk: return "Berserk"
    case ActorCrouching: return "Crouching"
    case ActorDrinking: return "Drinking"
    case ActorEating: return "Eating"
    case ActorBathing: return "Bathing"
    case ActorCooking: return "Cooking"
    case ActorForaging: return "Foraging"
    case ActorChopping: return "Chopping"
    case ActorDigging: return "Digging"
    case ActorMilking: return "Milking"
    case ActorStashing: return "Stashing"
    case ActorRelieving: return "Relieving"
    case ActorWorkshop: return "Workshop"
    case ActorIntercourse: return "Intercourse"
    case ActorSocializing: return "Socializing"
    case ActorGambling: return "Gambling"
    default: return "Unknown"
    }
}

func (a Alignment) String() string {
	switch a {
	case AlignmentNeutral: return "neutral"
	case AlignmentFriendly: return "friendly"
	case AlignmentHostile: return "hostile"
	default: return "unknown"
	}
}

type AgeConfig struct {
    Current  FloatInterval `yaml:"current"`
    Max      FloatInterval `yaml:"max"`
    Rate     FloatInterval `yaml:"rate"`
}

func (c AgeConfig) Roll() AgeState {
    return AgeState{Current: c.Current.Roll(), Max: c.Max.Roll(), Rate: c.Rate.Roll()}
}

type LaborShift int

const (
	ShiftWork LaborShift = iota
	ShiftLeisure
	ShiftRest
)

type BehaviorType int

const (
    BehaviorNone BehaviorType = iota
    BehaviorChaos
    BehaviorFighter
    BehaviorHunter
    BehaviorWander
    BehaviorPatrol
    BehaviorEscort
    BehaviorFlee
    BehaviorTrader
    BehaviorHauler
    BehaviorLumberjack
    BehaviorFarmer
    BehaviorArtisan
    BehaviorCriminal
    BehaviorSlave
    BehaviorSlaver
    BehaviorChaotic      = BehaviorChaos
    BehaviorNpcFighter   = BehaviorFighter
    BehaviorKnightHunter = BehaviorHunter
)

type State struct {
	HealthPoints    int     `yaml:"health_points"`
	MaxHealthPoints int     `yaml:"max_health_points"`
	Sanity          float64 `yaml:"sanity"`
	Hunger          float64 `yaml:"hunger"`
	Thirst          float64 `yaml:"thirst"`
	Fatigue         float64 `yaml:"fatigue"`
	BladderLevel    float64 `yaml:"bladder_level"`
	BowelLevel      float64 `yaml:"bowel_level"`
	Hygiene         float64 `yaml:"hygiene"`
	Arousal         float64 `yaml:"arousal"`
	IsDrunk         bool    `yaml:"is_drunk"`
	AlcoholLevel    float64 `yaml:"alcohol_level"`
	Pain            float64 `yaml:"pain"`
	IsSick          bool    `yaml:"is_sick"`
	IsSeptic        bool    `yaml:"is_septic"`
	IsPoisoned      bool    `yaml:"is_poisoned"`
	IsConscious     bool    `yaml:"is_conscious"`
	Age             AgeState `yaml:"age"`
	IsAngry         bool    `yaml:"is_angry"`
	IsHypersexual   bool    `yaml:"is_hypersexual"`
	ThirstMax       float64 `yaml:"thirst_max"`
	HydrationBuffer int     `yaml:"hydration_buffer"`
}

type AgeState struct {
	Current float64 `yaml:"current"`
	Max     float64 `yaml:"max"`
	Rate    float64 `yaml:"rate"`
}

type PrimaryAttributes struct {
	Strength  int `yaml:"strength"`
	Dexterity int `yaml:"dexterity"`
	Health    int `yaml:"health"`
	Intellect int `yaml:"intellect"`
	Wisdom    int `yaml:"wisdom"`
}

type PhysicalTrauma struct {
	LeftArmLost  bool
	RightArmLost bool
	LeftLegLost  bool
	RightLegLost bool
	EyesLost     int
	BurnedAlive  bool
	SpineBroken  bool
}

type MemoryEvent struct {
	Tick      int
	Type      string
	Context   string
	EmotionalWeight float64
    Source    string
    Value     float64
}

type EntityStats struct {
	HealthPoints int
	HungerMax, ThirstMax, FatigueMax, Speed float64
	BaseAttack, BaseDefense, BaseProtection, AttackCooldown int
	AttackRange, ProjectileSpeed float64
	IsMilkable bool
	MilkCooldown int
	MaxWeight float64
	Age AgeState
	Survivalism string
}

type Loan struct {
	LenderUID string `yaml:"lender_uid"`
	Amount    int    `yaml:"amount"`
	Deadline  int    `yaml:"deadline"`
}

type Actor struct {
	X, Y   float64
	Z      float64
	VerticalVelocity float64
	State             State             `yaml:"state"`
	PrimaryAttributes PrimaryAttributes `yaml:"attributes"`
	RawStats          EntityStats       `yaml:"-"` 
	SkillValues       map[string]int    `yaml:"skills,omitempty"`
	
	BodyStatus map[string]int 
	Config *EntityConfig
	Facing Direction
	ActionState  ActorState
	Trauma PhysicalTrauma
	
	Path      []engine.Point
	PathTimer int 
	
	WanderDirX float64
	WanderDirY float64
	PatrolStartX, PatrolStartY float64
	PatrolEndX, PatrolEndY     float64
	PatrolHeading              bool
	
	WorkTicks     int
	LeisureTicks  int
	
	ID               string
	Name             string
	Denarii          int
	Alignment        Alignment
	Group            string
	LeaderID         string
	MustSurvive      bool
	IsTransexual     bool
	ParentID         string 
	FatherID         string 
	MasterID         string 
	UID              string 
	Debts            []Loan 
	
	AgeTicks         float64
	LifeStage        string 
	MortalityChecked bool 
	
	BaseAttack         int
	RangedAttack       int     
	BaseDefense        int
	BaseProtection     int
	BaseAttackCooldown int
	Speed              float64
	CriticalChance     float64
	
	AttackBonus     int
	DefenseBonus    int
	ProtectionBonus int
	SpeedBonus      float64
	MaxHealthBonus  int
	RegenPerSecond  int
	
	BodyTemperature      float64
	PreferredTemperature float64
	FluTicks             int
	SicknessTicks        int    
	Sickness             string 
	ContagionTimer       int
	
	Relationships      map[string]float64
	RomanticInterest   map[string]float64
	Submission         map[string]float64
	SexualOrientation  string 
	Memories           []MemoryEvent
	LastLoggedState    ActorState
	LastLoggedHP       int
	GroupSentiment     map[string]float64
	IsPregnant         bool
	GestationTicks     int
	MatingCooldown     int
	ArousalTimer       int
	
	Inventory []*ItemInstance
	Slots     map[string]*ItemInstance
	MaxWeight float64
	Weapon    *Weapon
	BaseWeapon *Weapon
	
	Shift              LaborShift
	Mood               MoodType
	LastAIChoice       string
	LastAIReasoning    string
	ThoughtTimer       int
	LastAIDecisionTick int
	LastTalkTick       int
	CurrentTile        string
	
	TargetActorID      string
	TargetActor        *Actor
	TargetObstacle     *Obstacle
	TargetItem         *ItemInstance
	TargetStockpile    *FloorZone
	
	Nourishment int
	Survivalism int
	Mate      float64
	Crafting    int
	Herbalism   int
	Trading     int
	Harvesting  int
	Husbandry   int
	Art         int
	Culture     int
	
	Kills    int
	MapKills map[string]int
	Level    int
	XP       int
	
	OwnedChestID      string
	MilkCooldownTicks int
	MeatQuantity      float64
	
	Tick               int
	HitTimer           int
	LastReaction       string
	SelectedModel      string
	DeadTimer          int
	CrouchTimer        int
	IsOccluded         bool
	IsTarget           bool
	UnconsciousTimer   int
	IsConscious        bool 

	Behavior   BehaviorType

	// Simulation: DF Features
	RotTicks           int  // Ticks since death (for miasma)
	GriefTicks         int  // Ticks of mourning
	LodgingTicks       int  // Ticks of paid inn stay
	Scale              float64 // Visual scale (1.0 = normal)
	LastCompanionTick  int  // Last time they had social companionship
	LastGamblingTick   int  // Last time they played at the Fortune Home
	
	// Performance Caching
	StatsStable        bool `yaml:"-"`
	LastVisionTick     int  `yaml:"-"`
	LastSyncTick       int  `yaml:"-"`
	
	mu                 sync.Mutex `yaml:"-"`
}

func (a *Actor) Lock() { a.mu.Lock() }
func (a *Actor) Unlock() { a.mu.Unlock() }

func (a *Actor) DistanceToObject(o *Obstacle) float64 {
	return math.Sqrt(math.Pow(a.X-o.X, 2) + math.Pow(a.Y-o.Y, 2))
}

const (
	StageBaby     = "Baby"
	StageKid      = "Kid"
	StageTeenager = "Teenager"
	StageAdult    = "Adult"
	StageElder    = "Elder"
)
