package game

import (
	"math"
	"math/rand"
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
	LastAIChoice         string
	LastAIReasoning      string
	TargetActorForAI     *Actor
	TargetItem           *ItemInstance

	// Control state
	IsPlayerControlled bool
}

var characterNames = []string{
	"Grog", "Zog", "Bob", "Drok", "Gorak", "Mug", "Snarl", "Thrak", "Vrog", "Kurg",
	"Hicks", "Miller", "Cooper", "Smith", "Potter", "Baker", "Carter", "Fisher",
}

func NewCharacter(x, y float64, config *EntityConfig, level int, isPlayer bool, objReg *ObjectRegistry) *Character {
	if config == nil {
		config = &EntityConfig{ID: "unknown", Name: "Unknown Entity"}
		config.Stats.HealthMin = 10
		config.Stats.HealthMax = 10
		config.Stats.Speed = 0.1
		config.Stats.AttackCooldown = 60
		config.Weapon = WeaponConfig{Inline: WeaponTizon}
	}
	
	c := &Character{
		Actor: Actor{
			X: x, Y: y, Config: config, State: ActorIdle, Facing: DirSE, Level: level,
			Alignment: AlignmentEnemy, Group: config.Group, LeaderID: config.LeaderID, MustSurvive: config.MustSurvive,
			Name: config.Name,
		},
		IsPlayerControlled: isPlayer,
	}

	if isPlayer {
		c.Alignment = AlignmentAlly
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
			c.PatrolHeading = true
		}
	}

	if config.Health > 0 {
		c.Health = config.Health
	} else if config.Stats.HealthMax > config.Stats.HealthMin {
		c.Health = config.Stats.HealthMin + rand.Intn(config.Stats.HealthMax-config.Stats.HealthMin+1)
	} else {
		c.Health = config.Stats.HealthMin
	}
	
	if config.Energy > 0 {
		c.Energy = config.Energy
	} else if config.Stats.EnergyMax > config.Stats.EnergyMin {
		c.Energy = config.Stats.EnergyMin + rand.Float64()*(config.Stats.EnergyMax-config.Stats.EnergyMin)
	} else if config.Stats.EnergyMin > 0 {
		c.Energy = config.Stats.EnergyMin
	} else {
		c.Energy = 100.0
	}
	c.BaseAttack = config.Stats.BaseAttack
	c.BaseDefense = config.Stats.BaseDefense
	c.Speed = config.Stats.Speed
	c.MaxWeight = config.MaxWeight
	c.Slots = make(map[string]*ItemInstance)
	c.Inventory = make([]*ItemInstance, 0)
	c.AttackCooldown = config.Stats.AttackCooldown
	c.Weapon = config.Weapon.Resolve(objReg) 
	if c.Weapon == nil {
		c.Weapon = WeaponFists
	}
	c.Health = c.calculateStat(c.Health, c.Level)
	c.MaxHealth = c.Health
	c.BaseAttack = c.calculateStat(c.BaseAttack, c.Level)
	c.BaseDefense = c.calculateStat(c.BaseDefense, c.Level)
	c.BaseProtection = c.calculateStat(config.Stats.BaseProtection, c.Level)
	c.MapKills = make(map[string]int)

	return c
}

func (c *Character) calculateStat(base int, level int) int {
	if level <= 1 { return base }
	return int(float64(base) * math.Pow(1.15, float64(level-1)))
}

func (c *Character) clampToMap(mapW, mapH float64) {
	halfW, halfH := mapW/2, mapH/2
	if c.X < -halfW { c.X = -halfW }
	if c.X > halfW { c.X = halfW }
	if c.Y < -halfH { c.Y = -halfH }
	if c.Y > halfH { c.Y = halfH }
}

func clampInt(v, min, max int) int {
	if v < min { return min }
	if v > max { return max }
	return v
}
