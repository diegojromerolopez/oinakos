package game

import (
	"math/rand"
	"strings"
)

// Character replaces the old NPC and PlayableCharacter structs.
type Character struct {
	Actor // Embedded shared state

	// AI and Behavior (inherited from NPC)
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
	AIDecisionPending    bool
	TargetActorForAI     *Actor
	TargetItem           *ItemInstance

	// Control state
	IsPlayerControlled bool
	PendingSkill       string // Currently executing attack skill (punch, slap, etc.)
}

var characterNames = []string{
	"Grog", "Zog", "Bob", "Drok", "Gorak", "Mug", "Snarl", "Thrak", "Vrog", "Kurg",
	"Hicks", "Miller", "Cooper", "Smith", "Potter", "Baker", "Carter", "Fisher",
}

func NewCharacter(x, y float64, config *EntityConfig, level int, isPlayer bool, objReg *ObjectRegistry) *Character {
	if config == nil {
		config = &EntityConfig{ID: "unknown", Name: "Unknown Entity"}
		config.Stats.HealthPoints = IntInterval{Min: 100, Max: 100}
		config.Stats.Speed = FloatInterval{Min: 0.1, Max: 0.1}
		config.Stats.AttackCooldown = IntInterval{Min: 60, Max: 60}
		config.Attributes = PrimaryAttributeConfig{
			Strength: IntInterval{Min: 100, Max: 100}, Dexterity: IntInterval{Min: 100, Max: 100},
			Health: IntInterval{Min: 100, Max: 100}, Intellect: IntInterval{Min: 100, Max: 100},
			Wisdom: IntInterval{Min: 100, Max: 100},
		}
		config.Weapon = WeaponConfig{Inline: WeaponTizon}
		config.Stats.MaxWeight = FloatInterval{Min: 100.0, Max: 100.0}
		config.Stats.Age = AgeConfig{Current: FloatInterval{Mean: 25.0, SD: 5.0, Mode: "normal"}, Rate: FloatInterval{Min: 1.0, Max: 1.0}} // Default adult age
	}

	attributes := config.Attributes.Roll()
	stats := config.Stats.Roll()
	skills := make(map[string]int)
	for id, interval := range config.Skills {
		skills[id] = interval.Roll()
	}
	
	c := &Character{
		Actor: Actor{
			X: x, Y: y, Config: config, ActionState: ActorIdle, Facing: DirSE, Level: level,
			Alignment: AlignmentEnemy, Group: config.Group, LeaderID: config.LeaderID, MustSurvive: config.MustSurvive,
			Name: config.Name,
			Denarii: config.Denarii,
			ID: config.ID,
			PrimaryAttributes: attributes,
			RawStats: stats,
			SkillValues: skills,
			IsTransexual: config.IsTransexual,
			BaseAttack: stats.BaseAttack,
			BaseDefense: stats.BaseDefense,
			BaseProtection: stats.BaseProtection,
			BaseAttackCooldown: stats.AttackCooldown,
			Speed: stats.Speed,
			MaxWeight: stats.MaxWeight,
		},
		IsPlayerControlled: isPlayer,
	}

	if isPlayer {
		c.Alignment = AlignmentAlly
	}

	// Dynamic Model Selection
	if len(config.Models) > 0 {
		models := make([]string, 0, len(config.Models))
		for k := range config.Models {
			models = append(models, k)
		}
		c.SelectedModel = models[rand.Intn(len(models))]
	}

	if !isPlayer {
		if config.Unique {
			c.Name = config.Name
		} else if len(config.Names) > 0 {
			c.Name = config.Names[rand.Intn(len(config.Names))]
		} else if config.Name != "" {
			c.Name = config.Name
		} else {
			c.Name = characterNames[rand.Intn(len(characterNames))]
		}

		switch config.Behavior {
		case "chaotic": c.Behavior = BehaviorChaotic
		case "fighter": c.Behavior = BehaviorNpcFighter
		case "hunter":  c.Behavior = BehaviorKnightHunter
		case "wander":  c.Behavior = BehaviorWander
		case "patrol":  c.Behavior = BehaviorPatrol
		case "escort":  c.Behavior = BehaviorEscort
		case "flee":    c.Behavior = BehaviorFlee
		case "trader":  c.Behavior = BehaviorTrader
		default:
			if config.Unique { c.Behavior = BehaviorWander } else { c.Behavior = BehaviorKnightHunter }
		}

		if c.Behavior == BehaviorWander {
			c.WanderDirX = rand.Float64()*2 - 1
			c.WanderDirY = rand.Float64()*2 - 1
		} else if c.Behavior == BehaviorPatrol {
			c.PatrolStartX = c.X
			c.PatrolStartY = c.Y
			c.PatrolEndX = c.X + (rand.Float64()*8 - 4)
			c.PatrolEndY = c.Y + (rand.Float64()*8 - 4)
			c.PatrolHeading = !c.PatrolHeading
		}
	}

	// Initial health values (clamped later by SyncStats)
	if config.Stats.HealthPoints.Min > 0 {
		minH := config.Stats.HealthPoints.Min
		maxH := config.Stats.HealthPoints.Max
		if maxH > minH {
			c.State.MaxHealthPoints = minH + rand.Intn(maxH-minH+1)
		} else {
			c.State.MaxHealthPoints = minH
		}
	} else {
		c.State.MaxHealthPoints = c.PrimaryAttributes.Health * 10
	}
	if c.State.MaxHealthPoints < 100 { c.State.MaxHealthPoints = 100 }
	c.State.HealthPoints = c.State.MaxHealthPoints
	c.State.Hygiene = 100.0
	c.State.IsConscious = true

	// Set LifeStage based on config.Archetype
	c.LifeStage = StageAdult
	if config != nil {
		arch := strings.ToLower(config.Archetype)
		if strings.Contains(arch, "baby") {
			c.LifeStage = "baby"
		} else if strings.Contains(arch, "kid") {
			c.LifeStage = "kid"
		} else if strings.Contains(arch, "teenager") {
			c.LifeStage = "teenager"
		} else if strings.Contains(arch, "elder") {
			c.LifeStage = "elder"
		}
	}
	// Initialize age ticks
	ageYears := stats.Age.Current
	if config.State.Age.Current > 0 {
		ageYears = config.State.Age.Current
	}
	c.AgeTicks = ageYears * float64(TicksPerYear)
	
	ageRate := stats.Age.Rate
	if config.State.Age.Current > 0 {
		// If age is specified in State, use its rate
		ageRate = config.State.Age.Rate
	}
	ageMax := stats.Age.Max
	if config.State.Age.Max > 0 {
		ageMax = config.State.Age.Max
	}

	c.State.Age = AgeState{Current: ageYears, Rate: ageRate, Max: ageMax}

	// Enforce 18+ for adults/elders
	if c.LifeStage == StageAdult || c.LifeStage == StageElder {
		minAgeTicks := 18.0 * float64(TicksPerYear)
		if c.AgeTicks < minAgeTicks {
			c.AgeTicks = minAgeTicks
		}
	}

	c.State.HealthPoints = c.State.MaxHealthPoints
	
	if config.State.Hunger > 0 {
		c.State.Hunger = config.State.Hunger
	}
	if config.State.Thirst > 0 {
		c.State.Thirst = config.State.Thirst
	}
	
	if config.State.Fatigue > 0 {
		c.State.Fatigue = config.State.Fatigue
	}
	
	if config.State.Hygiene > 0 {
		c.State.Hygiene = config.State.Hygiene
	} else {
		c.State.Hygiene = 100.0
	}
	if config.State.BladderLevel > 0 {
		c.State.BladderLevel = config.State.BladderLevel
	}
	if config.State.BowelLevel > 0 {
		c.State.BowelLevel = config.State.BowelLevel
	}

	c.State.Sanity = 100.0
	c.BodyTemperature = 37.0
	c.PreferredTemperature = 37.0
	
	// Centralized Primary & Derived Stats
	c.SyncStats(objReg)

	// Runtime health finalization and clamping
	if config.State.HealthPoints > 0 {
		c.State.HealthPoints = c.calculateStat(config.State.HealthPoints, c.Level)
	} else if c.State.HealthPoints == 0 {
		c.State.HealthPoints = c.State.MaxHealthPoints
	}
	
	// Final clamping to ensure we don't exceed max health from attributes
	if c.State.HealthPoints > c.State.MaxHealthPoints {
		c.State.HealthPoints = c.State.MaxHealthPoints
	}

	c.AttackCooldown = c.BaseAttackCooldown
	c.Slots = make(map[string]*ItemInstance)
	c.Inventory = make([]*ItemInstance, 0)

	c.Submission = make(map[string]float64)
	c.MapKills = make(map[string]int)

	// Load items from config for both player and NPCs
	c.LoadEquipment(objReg)

	return c
}

func (c *Character) clampToMap(mapW, mapH float64) {
	halfW, halfH := mapW/2, mapH/2
	if c.X < -halfW { c.X = -halfW }
	if c.X > halfW { c.X = halfW }
	if c.Y < -halfH { c.Y = -halfH }
	if c.Y > halfH { c.Y = halfH }
}
