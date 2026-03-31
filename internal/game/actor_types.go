package game

import (
	"fmt"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
	"strings"
)

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
	ActorDrinking  
	ActorCrouching 
	ActorIncapacitated 
	ActorResting   
	ActorChopping  
	ActorDigging   
	ActorForaging  
	ActorEating    
	ActorBerserk   
	ActorCooking   
	ActorWorkshop  
	ActorMilking   
	ActorStashing  
	ActorIntercourse 
	ActorBathing     
	ActorRelieving   
	ActorFeeding     
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
	case ActorFeeding:       return "Feeding"
	default:                 return "Unknown"
	}
}

type PlayableCharacterState = ActorState
type NPCState = ActorState
type Direction int

const (
	DirSE Direction = iota
	DirSW
	DirNE
	DirNW
)

type Alignment int

const (
	AlignmentEnemy Alignment = iota
	AlignmentNeutral
	AlignmentAlly
)

func (a Alignment) String() string {
	switch a {
	case AlignmentEnemy: return "ENEMY"
	case AlignmentNeutral: return "NEUTRAL"
	case AlignmentAlly: return "ALLY"
	default: return "UNKNOWN"
	}
}

func (a *Alignment) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err == nil {
		switch strings.ToLower(s) {
		case "enemy": *a = AlignmentEnemy; return nil
		case "neutral": *a = AlignmentNeutral; return nil
		case "ally": *a = AlignmentAlly; return nil
		}
	}
	var i int
	if err := value.Decode(&i); err == nil {
		if i >= 0 && i <= 2 {
			*a = Alignment(i)
			return nil
		}
	}
	return fmt.Errorf("unknown alignment: %s", value.Value)
}

type BehaviorType int

const (
	BehaviorWander BehaviorType = iota
	BehaviorPatrol
	BehaviorKnightHunter
	BehaviorNpcFighter
	BehaviorChaotic
	BehaviorEscort
	BehaviorFlee
	BehaviorTrader 
	BehaviorHauler 
	BehaviorLumberjack 
	BehaviorFarmer     
	BehaviorArtisan    
)

type PhysicalTrauma struct {
	LeftArmLost   bool
	RightArmLost  bool
	LeftLegLost   bool
	RightLegLost  bool
	LeftHandLost  bool
	RightHandLost bool
	EyesLost      int  
	BurnedAlive   bool 
	SpineBroken   bool
}

type MemoryEvent struct {
	Tick   int     `json:"tick"`
	Type   string  `json:"type"`   
	Source string  `json:"source"` 
	Value  float64 `json:"value"`  
}

type PrimaryAttributes struct {
	Strength           int     
	Dexterity          int     
	Health             int     
	Intellect          int     
	Wisdom             int     
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
	Max     float64       `yaml:"max"` 
}

func (c AgeConfig) Roll() AgeState {
	return AgeState{
		Current: c.Current.Roll(),
		Rate:    c.Rate,
		Max:     c.Max,
	}
}

type State struct {
	HealthPoints     int     `yaml:"health_points,omitempty"`
	MaxHealthPoints  int
	Hunger           float64 `yaml:"hunger,omitempty"`
	Thirst           float64 `yaml:"thirst,omitempty"`
	Fatigue          float64 `yaml:"fatigue,omitempty"`
	Sanity           float64 `yaml:"sanity,omitempty"`
	IsPoisoned       bool    `yaml:"is_poisoned,omitempty"`
	IsAngry          bool    `yaml:"is_angry,omitempty"`
	IsConscious      bool    `yaml:"is_conscious,omitempty"`
	IsSeptic         bool    `yaml:"is_septic,omitempty"` 
	IsSick           bool    `yaml:"is_sick,omitempty"`   
	Arousal          float64 `yaml:"arousal,omitempty"`   
	Pain             float64 `yaml:"pain,omitempty"`      
	Hygiene          float64 `yaml:"hygiene,omitempty"`   
	BladderLevel     float64 `yaml:"bladder_level,omitempty"` 
	BowelLevel       float64 `yaml:"bowel_level,omitempty"`   
	AlcoholLevel     float64 `yaml:"alcohol_level,omitempty"`
	IsDrunk          bool    `yaml:"is_drunk,omitempty"`
	Age              AgeState `yaml:"age,omitempty"`
}

type AgeState struct {
	Current float64 `yaml:"current"` 
	Rate    float64 `yaml:"rate"`    
	Max     float64 `yaml:"max"`     
}

const (
	StageBaby     = "baby"
	StageKid      = "kid"
	StageTeenager = "teenager"
	StageAdult    = "peasant"
	StageElder    = "elder"
)

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
	
	AgeTicks         float64
	LifeStage        string 
	
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
	Memories           []MemoryEvent
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
	TargetObstacle     *Obstacle
	TargetItem         *ItemInstance
	TargetStockpile    *FloorZone
	
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
	
	Kills    int
	MapKills map[string]int
	Level    int
	XP       int
	
	OwnedChestID      string
	MilkCooldownTicks int
	MeatQuantity      float64
	
	Tick               int
	HitTimer           int
	DeadTimer          int
	CrouchTimer        int
	IsOccluded         bool
	IsTarget           bool
	IsConscious        bool 
	UnconsciousTimer   int
	
	LastReaction       string
	SelectedModel      string

	// Simulation: DF Features
	RotTicks           int  // Ticks since death (for miasma)
	GriefTicks         int  // Ticks of mourning
}
