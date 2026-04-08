package game

type ObstacleActionType string
const ( ActionHarm ObstacleActionType = "harm"; ActionHeal ObstacleActionType = "heal"; ActionBath ObstacleActionType = "bath"; ActionAlleviate ObstacleActionType = "alleviate"; ActionSoiling ObstacleActionType = "soiling" )

type ObstacleActionConfig struct {
	Type ObstacleActionType `yaml:"type"`; Amount int `yaml:"amount"`; Aura float64 `yaml:"aura"`; AlignmentLimit string `yaml:"alignment_limit"`; RequiresInteraction bool `yaml:"requires_interaction"`
}

type ActionConfig struct { OnKill []KillAction `yaml:"on_kill"` }
type KillAction struct { Type string `yaml:"type"`; Probability float64 `yaml:"probability"`; Effect ActionEffect `yaml:"effect"` }
type ActionEffect struct { Victim *VictimEffect `yaml:"victim,omitempty"`; Attacker *AttackerEffect `yaml:"attacker,omitempty"` }
type VictimEffect struct {
	Transform   string `yaml:"transform,omitempty"`
	Alignment   string `yaml:"alignment,omitempty"`
	SpawnCorpse *bool  `yaml:"spawn_corpse,omitempty"`
	CorpseImage string `yaml:"corpse_image,omitempty"`
}
type AttackerEffect struct { Heal int `yaml:"heal,omitempty"` }
type StatEffect struct { Increase float64 `yaml:"increase"` }

type AbilityEffect struct {
	StunChance float64 `yaml:"stun_chance,omitempty"`; Duration float64 `yaml:"duration,omitempty"`; KnockbackDistance float64 `yaml:"knockback_distance,omitempty"`; ArmorBreakPercentage float64 `yaml:"armor_break_percentage,omitempty"`; PierceTargets int `yaml:"pierce_targets,omitempty"`; PoisonDamagePerSecond int `yaml:"poison_damage_per_second,omitempty"`; Probability float64 `yaml:"probability,omitempty"`
}

type Ability struct {
	Damage string `yaml:"damage,omitempty"`; Yield string `yaml:"yield,omitempty"`; RequiredWeapon string `yaml:"required_weapon,omitempty"`; ParentAttribute string `yaml:"parent_attribute,omitempty"`; Effects []AbilityEffect `yaml:"effects"`
}
