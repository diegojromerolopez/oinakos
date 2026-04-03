package game

import "oinakos/internal/engine"

type PrimaryAttributeConfig struct {
	Strength  IntInterval `yaml:"strength"`
	Dexterity IntInterval `yaml:"dexterity"`
	Health    IntInterval `yaml:"health"`
	Intellect IntInterval `yaml:"intellect"`
	Wisdom    IntInterval `yaml:"wisdom"`
}

func (c PrimaryAttributeConfig) Roll() PrimaryAttributes {
	return PrimaryAttributes{Strength: c.Strength.Roll(), Dexterity: c.Dexterity.Roll(), Health: c.Health.Roll(), Intellect: c.Intellect.Roll(), Wisdom: c.Wisdom.Roll()}
}

type EntityStatsConfig struct {
	HealthPoints IntInterval `yaml:"health_points"`
	HungerMax FloatInterval `yaml:"hunger_max"`; ThirstMax FloatInterval `yaml:"thirst_max"`; FatigueMax FloatInterval `yaml:"fatigue_max"`; Speed FloatInterval `yaml:"speed"`
	BaseAttack IntInterval `yaml:"base_attack"`; BaseDefense IntInterval `yaml:"base_defense"`; BaseProtection IntInterval `yaml:"base_protection"`
	Survivalism string `yaml:"survivalism"`; AttackCooldown IntInterval `yaml:"attack_cooldown"`; AttackRange FloatInterval `yaml:"attack_range"`; ProjectileSpeed FloatInterval `yaml:"projectile_speed"`
	IsMilkable bool `yaml:"is_milkable"`; MilkCooldown IntInterval `yaml:"milk_cooldown"`; MaxWeight FloatInterval `yaml:"max_weight"`; Age AgeConfig `yaml:"age"`
}

func (c EntityStatsConfig) Roll() EntityStats {
	return EntityStats{HealthPoints: c.HealthPoints.Roll(), HungerMax: c.HungerMax.Roll(), ThirstMax: c.ThirstMax.Roll(), FatigueMax: c.FatigueMax.Roll(), Speed: c.Speed.Roll(), BaseAttack: c.BaseAttack.Roll(), BaseDefense: c.BaseDefense.Roll(), BaseProtection: c.BaseProtection.Roll(), AttackCooldown: c.AttackCooldown.Roll(), AttackRange: c.AttackRange.Roll(), ProjectileSpeed: c.ProjectileSpeed.Roll(), IsMilkable: c.IsMilkable, MilkCooldown: c.MilkCooldown.Roll(), MaxWeight: c.MaxWeight.Roll(), Age: c.Age.Roll(), Survivalism: c.Survivalism}
}

type EntityConfig struct {
	ID string `yaml:"id"`; Name string `yaml:"name"`; Names []string `yaml:"names"`; Archetype string `yaml:"archetype,omitempty"`; Behavior string `yaml:"behavior"`; Attributes PrimaryAttributeConfig `yaml:"attributes"`; Stats EntityStatsConfig `yaml:"stats"`; Skills map[string]IntInterval `yaml:"skills,omitempty"`; State State `yaml:"state,omitempty"`; Actions *ActionConfig `yaml:"actions,omitempty"`; Abilities map[string]Ability `yaml:"abilities,omitempty"`; Weapon WeaponConfig `yaml:"weapon"`; CollisionRadius float64 `yaml:"collision_radius,omitempty"`
	Footprint []FootprintPoint `yaml:"footprint"`; Description string `yaml:"description,omitempty"`; Unique bool `yaml:"unique,omitempty"`; Gender string `yaml:"gender,omitempty"`; IsTransexual bool `yaml:"transexual,omitempty"`; SexualOrientation string `yaml:"sexual_orientation,omitempty"`; SoundID string `yaml:"-"`; PlayableCharacter string `yaml:"-"`; PrimaryColor string `yaml:"primary_color,omitempty"`; SecondaryColor string `yaml:"secondary_color,omitempty"`; XP int `yaml:"xp,omitempty"`; Group string `yaml:"group,omitempty"`; LeaderID string `yaml:"leader,omitempty"`; MustSurvive bool `yaml:"must_survive,omitempty"`; Playable bool `yaml:"playable,omitempty"`; MaxItems int `yaml:"max_items,omitempty"`; Equipment map[string]string `yaml:"equipment,omitempty"`; Inventory []string `yaml:"inventory,omitempty"`; Denarii int `yaml:"denarii,omitempty"`
	AssetDir, AudioDir string `yaml:"-"`; StaticImage, BackImage, CorpseImage, CrouchImage, AttackImage, Attack1Image, Attack2Image, HitImage, Hit1Image, Hit2Image, ChoppingImage, DiggingImage, PregnantImage, CookingImage engine.Image `yaml:"-"`
	Meat int `yaml:"meat,omitempty"`; IsAnimal bool `yaml:"is_animal,omitempty"`; CachedBaseFootprint *engine.Polygon `yaml:"-"`; Dialogues *DialogueRoot `yaml:"dialogues,omitempty"`; Models map[string]*ModelConfig `yaml:"-"`
}

type ModelConfig struct {
	ID string; StaticImage, BackImage, CorpseImage, CrouchImage, AttackImage, HitImage, PregnantImage, CookingImage engine.Image
}
