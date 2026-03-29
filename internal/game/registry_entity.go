package game

import (
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"strings"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type PrimaryAttributeConfig struct {
	Strength  IntInterval `yaml:"strength"`
	Dexterity IntInterval `yaml:"dexterity"`
	Health    IntInterval `yaml:"health"`
	Intellect IntInterval `yaml:"intellect"`
	Wisdom    IntInterval `yaml:"wisdom"`
}

func (c PrimaryAttributeConfig) Roll() PrimaryAttributes {
	return PrimaryAttributes{
		Strength:  c.Strength.Roll(),
		Dexterity: c.Dexterity.Roll(),
		Health:    c.Health.Roll(),
		Intellect: c.Intellect.Roll(),
		Wisdom:    c.Wisdom.Roll(),
	}
}

type EntityStatsConfig struct {
	HealthMin       IntInterval   `yaml:"health_points_min"`
	HealthMax       IntInterval   `yaml:"health_points_max"`
	HungerMax       FloatInterval `yaml:"hunger_max"`
	ThirstMax       FloatInterval `yaml:"thirst_max"`
	FatigueMax      FloatInterval `yaml:"fatigue_max"`
	Speed           FloatInterval `yaml:"speed"`
	BaseAttack      IntInterval   `yaml:"base_attack"`
	BaseDefense     IntInterval   `yaml:"base_defense"`
	BaseProtection  IntInterval   `yaml:"base_protection"`
	AttackCooldown  IntInterval   `yaml:"attack_cooldown"`
	AttackRange     FloatInterval `yaml:"attack_range"`
	ProjectileSpeed FloatInterval `yaml:"projectile_speed"`
	IsMilkable      bool          `yaml:"is_milkable"`
	MilkCooldown    IntInterval   `yaml:"milk_cooldown"`
	MaxWeight       FloatInterval `yaml:"max_weight"`
	Age             AgeConfig     `yaml:"age"`
}

func (c EntityStatsConfig) Roll() EntityStats {
	return EntityStats{
		HealthMin:       c.HealthMin.Roll(),
		HealthMax:       c.HealthMax.Roll(),
		HungerMax:       c.HungerMax.Roll(),
		ThirstMax:       c.ThirstMax.Roll(),
		FatigueMax:      c.FatigueMax.Roll(),
		Speed:           c.Speed.Roll(),
		BaseAttack:      c.BaseAttack.Roll(),
		BaseDefense:     c.BaseDefense.Roll(),
		BaseProtection:  c.BaseProtection.Roll(),
		AttackCooldown:  c.AttackCooldown.Roll(),
		AttackRange:     c.AttackRange.Roll(),
		ProjectileSpeed: c.ProjectileSpeed.Roll(),
		IsMilkable:      c.IsMilkable,
		MilkCooldown:    c.MilkCooldown.Roll(),
		MaxWeight:       c.MaxWeight.Roll(),
		Age:             c.Age.Roll(),
	}
}

type EntityConfig struct {
	ID                string                 `yaml:"id"`
	Name              string                 `yaml:"name"`
	Names             []string               `yaml:"names"`
	Archetype         string                 `yaml:"archetype,omitempty"`
	Behavior          string                 `yaml:"behavior"`
	Attributes        PrimaryAttributeConfig `yaml:"attributes"`
	Stats             EntityStatsConfig      `yaml:"stats"`
	Skills            map[string]IntInterval `yaml:"skills,omitempty"`
	State             TemporalState          `yaml:"state,omitempty"`
	Actions           *ActionConfig          `yaml:"actions,omitempty"`
	Abilities         map[string]Ability     `yaml:"abilities,omitempty"`
	Weapon      WeaponConfig  `yaml:"weapon"`
	CollisionRadius float64      `yaml:"collision_radius,omitempty"`

	Footprint      []FootprintPoint `yaml:"footprint"`
	Description    string           `yaml:"description,omitempty"`
	Unique         bool             `yaml:"unique,omitempty"`
	Gender         string           `yaml:"gender,omitempty"`
	IsTransexual   bool             `yaml:"transexual,omitempty"`
	SoundID        string           `yaml:"-"` // ID used for audio lookups (e.g. "man_at_arms_male")
	PlayableCharacter  string           `yaml:"-"` // Set to config.ID when this is the active playable character
	PrimaryColor   string           `yaml:"primary_color,omitempty"`
	SecondaryColor string           `yaml:"secondary_color,omitempty"`
	XP             int              `yaml:"xp,omitempty"` // XP awarded on kill
	Group          string           `yaml:"group,omitempty"`
	LeaderID       string           `yaml:"leader,omitempty"`
	MustSurvive    bool             `yaml:"must_survive,omitempty"`
	Playable       bool             `yaml:"playable,omitempty"`
	MaxItems       int              `yaml:"max_items,omitempty"`
	Equipment      map[string]string `yaml:"equipment,omitempty"` // map of slot name to object ID
	Inventory      []string         `yaml:"inventory,omitempty"`  // IDs of objects in backpack
	Denarii        int              `yaml:"denarii,omitempty"`

	// Run-time loaded assets
	AssetDir     string      `yaml:"-"`
	AudioDir     string      `yaml:"-"` // e.g. assets/audio/archetypes/orc/male
	StaticImage  engine.Image `yaml:"-"`
	BackImage    engine.Image `yaml:"-"` // back.png (instead of static.png when facing UP)
	CorpseImage  engine.Image `yaml:"-"`
	CrouchImage  engine.Image `yaml:"-"` // crouch.png (for picking up items)
	AttackImage  engine.Image `yaml:"-"` // attack.png (default)
	Attack1Image engine.Image `yaml:"-"` // attack1.png
	Attack2Image engine.Image `yaml:"-"` // attack2.png
	HitImage     engine.Image `yaml:"-"` // hit.png  (legacy / single hit frame)
	Hit1Image    engine.Image `yaml:"-"` // hit1.png (first variant)
	Hit2Image    engine.Image `yaml:"-"` // hit2.png (second variant, requires hit1.png)
	ChoppingImage engine.Image `yaml:"-"` // chopping.png
	DiggingImage  engine.Image `yaml:"-"` // digging.png
	PregnantImage engine.Image `yaml:"-"` // pregnant.png


	Meat           int              `yaml:"meat,omitempty"` // Amount of meat dropped on death
	IsAnimal       bool             `yaml:"is_animal,omitempty"`
	CachedBaseFootprint *engine.Polygon `yaml:"-"`

	Dialogues *DialogueRoot `yaml:"dialogues,omitempty"`

	// Models for variants
	Models map[string]*ModelConfig `yaml:"-"`
}

type ModelConfig struct {
	ID          string
	StaticImage engine.Image
	BackImage   engine.Image
	CorpseImage engine.Image
	CrouchImage engine.Image
	AttackImage engine.Image
	HitImage    engine.Image
	PregnantImage engine.Image
}


func (e *EntityConfig) PickAttackImage(seed int) engine.Image {
	if e.Attack1Image != nil {
		if e.Attack2Image != nil {
			if seed%2 == 0 {
				return e.Attack1Image
			}
			return e.Attack2Image
		}
		return e.Attack1Image
	}
	if e.AttackImage != nil {
		return e.AttackImage
	}
	return nil
}

func (e *EntityConfig) PickHitImage(seed int) engine.Image {
	if e.Hit1Image != nil {
		if e.Hit2Image != nil {
			if seed%2 == 0 {
				return e.Hit1Image
			}
			return e.Hit2Image
		}
		return e.Hit1Image
	}
	if e.HitImage != nil {
		return e.HitImage
	}
	return nil
}


func (c *EntityConfig) GetFootprint() engine.Polygon {
	if c.CachedBaseFootprint != nil {
		return *c.CachedBaseFootprint
	}

	var poly engine.Polygon
	if len(c.Footprint) == 0 {
		poly = engine.Polygon{Points: []engine.Point{
			{X: -0.15, Y: -0.15}, {X: 0.15, Y: -0.15},
			{X: 0.15, Y: 0.15}, {X: -0.15, Y: 0.15},
		}}
	} else {
		poly = engine.Polygon{Points: make([]engine.Point, len(c.Footprint))}
		for i, p := range c.Footprint {
			poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
		}
	}
	c.CachedBaseFootprint = &poly
	return poly
}

type Archetype = EntityConfig

type ArchetypeRegistry struct {
	Archetypes map[string]*Archetype
	IDs        []string
}

func NewArchetypeRegistry() *ArchetypeRegistry {
	return &ArchetypeRegistry{
		Archetypes: make(map[string]*Archetype),
		IDs:        make([]string, 0),
	}
}

func (r *ArchetypeRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, permitList map[string]bool, ls *LoadingState) {
	// Pre-scan for models subdirectory
	for _, config := range r.Archetypes {
		if config.AssetDir == "" { continue }
		modelDir := path.Join(config.AssetDir, "models")
		entries, err := fs.ReadDir(assets, modelDir)
		if err == nil {
			config.Models = make(map[string]*ModelConfig)
			for _, entry := range entries {
				if entry.IsDir() {
					config.Models[entry.Name()] = &ModelConfig{ID: entry.Name()}
				}
			}
		}
	}

	jobs := r.createLoadJobs(permitList)
	if len(jobs) > 0 {
		loadSpritesParallel(assets, jobs, graphics, ls)
	}
}

func (r *ArchetypeRegistry) createLoadJobs(permitList map[string]bool) []*SpriteLoadJob {
	var jobs []*SpriteLoadJob
	for _, config := range r.Archetypes {
		if permitList != nil && !permitList[config.ID] {
			continue
		}
		if config.AssetDir == "" {
			continue
		}
		
		addJob := func(assetPath string, target *engine.Image) {
			jobs = append(jobs, &SpriteLoadJob{
				Path: assetPath,
				Dest: target,
			})
		}
		
		// 1. Base files
		addJob(path.Join(config.AssetDir, "static.png"), &config.StaticImage)
		addJob(path.Join(config.AssetDir, "back.png"), &config.BackImage)
		addJob(path.Join(config.AssetDir, "corpse.png"), &config.CorpseImage)
		addJob(path.Join(config.AssetDir, "attack.png"), &config.AttackImage)
		addJob(path.Join(config.AssetDir, "attack1.png"), &config.Attack1Image)
		addJob(path.Join(config.AssetDir, "attack2.png"), &config.Attack2Image)
		addJob(path.Join(config.AssetDir, "hit.png"), &config.HitImage)
		addJob(path.Join(config.AssetDir, "hit1.png"), &config.Hit1Image)
		addJob(path.Join(config.AssetDir, "hit2.png"), &config.Hit2Image)
		addJob(path.Join(config.AssetDir, "crouch.png"), &config.CrouchImage)
		addJob(path.Join(config.AssetDir, "chopping.png"), &config.ChoppingImage)
		addJob(path.Join(config.AssetDir, "digging.png"), &config.DiggingImage)
		addJob(path.Join(config.AssetDir, "pregnant.png"), &config.PregnantImage)

		// 2. Models
		for modelID, mod := range config.Models {
			mDir := path.Join(config.AssetDir, "models", modelID)
			addJob(path.Join(mDir, "static.png"), &mod.StaticImage)
			addJob(path.Join(mDir, "back.png"), &mod.BackImage)
			addJob(path.Join(mDir, "corpse.png"), &mod.CorpseImage)
			addJob(path.Join(mDir, "attack.png"), &mod.AttackImage)
			addJob(path.Join(mDir, "hit.png"), &mod.HitImage)
			addJob(path.Join(mDir, "crouch.png"), &mod.CrouchImage)
			addJob(path.Join(mDir, "pregnant.png"), &mod.PregnantImage)
		}
	}
	return jobs
}

func (r *ArchetypeRegistry) CountAssets(permitList map[string]bool) int {
	return len(r.createLoadJobs(permitList))
}

func (r *ArchetypeRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	baseDirs := []string{"data/archetypes", "data/animals"}
	for _, baseDir := range baseDirs {
		err := forEachYAML(assets, baseDir, func(fpath string, data []byte) error {
			cleanRelPath := fpath
			if strings.HasPrefix(fpath, "oinakos"+string(filepath.Separator)) {
				cleanRelPath = fpath[len("oinakos"+string(filepath.Separator)):]
			}

			relPath, err := filepath.Rel(baseDir, cleanRelPath)
			if err != nil {
				return nil
			}
			subDir := filepath.Dir(relPath)
			if subDir == "." {
				subDir = ""
			}

			var config Archetype
			if err := yaml.Unmarshal(data, &config); err != nil {
				log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
				return nil
			}

			variantName := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
			if config.ID == "" {
				config.ID = variantName
			}

			sanitizeEntityConfig(&config, fpath)
			
			// Map asset and audio directories based on baseDir
			category := "archetypes"
			if baseDir == "data/animals" {
				category = "animals"
				config.IsAnimal = true
			}
			
			config.AssetDir = path.Join("assets/images", category, subDir, variantName)
			config.AudioDir = path.Join("assets/audio", category, subDir, variantName)

			// config.Weapon is now auto-loaded by YAML

			config.SoundID = config.ID

			r.Archetypes[config.ID] = &config
			r.IDs = append(r.IDs, config.ID)
			return nil
		})
		if err != nil {
			log.Printf("Warning: error loading from %s: %v", baseDir, err)
		}
	}
	return nil
}

type CharacterRegistry struct {
	Characters map[string]*EntityConfig
	IDs        []string
}

func NewCharacterRegistry() *CharacterRegistry {
	return &CharacterRegistry{
		Characters: make(map[string]*EntityConfig),
		IDs:        make([]string, 0),
	}
}

// PlayableIDs returns the subset of IDs whose characters have Playable == true,
// in the same order they were registered.
func (r *CharacterRegistry) PlayableIDs() []string {
	var ids []string
	for _, id := range r.IDs {
		if c, ok := r.Characters[id]; ok && c.Playable {
			ids = append(ids, id)
		}
	}
	return ids
}

func (r *CharacterRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	const baseDir = "data/characters"
	return forEachYAML(assets, baseDir, func(fpath string, data []byte) error {
		var config EntityConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}

		if config.ID == "" {
			config.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
		}

		sanitizeEntityConfig(&config, fpath)

		// Set asset and audio directories
		config.AssetDir = path.Join("assets/images/characters", config.ID)
		config.AudioDir = path.Join("assets/audio/characters", config.ID)
		config.SoundID = config.ID

		if config.Playable {
			config.PlayableCharacter = config.ID
		}

		r.Characters[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		return nil
	})
}

func (r *CharacterRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, archs *ArchetypeRegistry, permitList map[string]bool, ls *LoadingState) {
	jobs := r.createLoadJobs(assets, archs, permitList)
	if len(jobs) > 0 {
		loadSpritesParallel(assets, jobs, graphics, ls)
	}
}

func (r *CharacterRegistry) createLoadJobs(assets fs.FS, archs *ArchetypeRegistry, permitList map[string]bool) []*SpriteLoadJob {
	var jobs []*SpriteLoadJob
	for _, config := range r.Characters {
		if config.StaticImage != nil {
			continue
		}
		if permitList != nil && !permitList[config.ID] && !config.Playable {
			continue
		}
		
		// Fallback to archetype logic for non-playable characters
		lookupID := config.Archetype
		if config.Gender != "" && !strings.Contains(config.Archetype, config.Gender) {
			fullID := config.Archetype + "_" + config.Gender
			if _, exists := archs.Archetypes[fullID]; exists {
				lookupID = fullID
			}
		}

		arch, _ := archs.Archetypes[lookupID]

		// Audio ID resolution
		hasLocalAudio := false
		if config.AudioDir != "" {
			if entries, err := fs.ReadDir(assets, config.AudioDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						hasLocalAudio = true
						break
					}
				}
			}
		}

		if hasLocalAudio {
			config.SoundID = config.ID
		} else if arch != nil {
			config.SoundID = lookupID
		} else {
			config.SoundID = config.ID
		}

		// Asset loading jobs
		if config.AssetDir != "" {
			addJob := func(filename string, target *engine.Image, fallback engine.Image) {
				fpath := path.Join(config.AssetDir, filename)
				if _, err := fs.Stat(assets, fpath); err == nil {
					jobs = append(jobs, &SpriteLoadJob{Path: fpath, Dest: target})
				} else {
					if fallback != nil {
						*target = fallback
					}
				}
			}
			
			var archStatic, archBack, archCorpse, archCrouch engine.Image
			if arch != nil {
				archStatic, archBack, archCorpse, archCrouch = arch.StaticImage, arch.BackImage, arch.CorpseImage, arch.CrouchImage
			}

			addJob("static.png", &config.StaticImage, archStatic)
			addJob("back.png", &config.BackImage, archBack)
			addJob("corpse.png", &config.CorpseImage, archCorpse)
			addJob("crouch.png", &config.CrouchImage, archCrouch)
			
			addJob("attack.png", &config.AttackImage, nil)
			addJob("attack1.png", &config.Attack1Image, nil)
			addJob("attack2.png", &config.Attack2Image, nil)
			addJob("hit.png", &config.HitImage, nil)
			addJob("hit1.png", &config.Hit1Image, nil)
			addJob("hit2.png", &config.Hit2Image, nil)
			addJob("chopping.png", &config.ChoppingImage, nil)
			addJob("digging.png", &config.DiggingImage, nil)
		}
	}
	return jobs
}

func (r *CharacterRegistry) ProcessInheritance(archs *ArchetypeRegistry) {
	for _, config := range r.Characters {
		lookupID := config.Archetype
		if config.Gender != "" && !strings.Contains(config.Archetype, config.Gender) {
			fullID := config.Archetype + "_" + config.Gender
			if _, exists := archs.Archetypes[fullID]; exists {
				lookupID = fullID
			}
		}

		arch, _ := archs.Archetypes[lookupID]
		if arch != nil {
			if config.Stats.HealthMin.IsZero() { config.Stats.HealthMin = arch.Stats.HealthMin }
			if config.Stats.HealthMax.IsZero() { config.Stats.HealthMax = arch.Stats.HealthMax }
			if config.Stats.Speed.IsZero() { config.Stats.Speed = arch.Stats.Speed }
			if config.Stats.BaseAttack.IsZero() { config.Stats.BaseAttack = arch.Stats.BaseAttack }
			if config.Stats.ProjectileSpeed.IsZero() { config.Stats.ProjectileSpeed = arch.Stats.ProjectileSpeed }
			if config.Stats.AttackCooldown.IsZero() { config.Stats.AttackCooldown = arch.Stats.AttackCooldown }
			if config.Stats.BaseDefense.IsZero() { config.Stats.BaseDefense = arch.Stats.BaseDefense }
			
			// Inherit Attributes
			if config.Attributes.Strength.IsZero() { config.Attributes.Strength = arch.Attributes.Strength }
			if config.Attributes.Dexterity.IsZero() { config.Attributes.Dexterity = arch.Attributes.Dexterity }
			if config.Attributes.Health.IsZero() { config.Attributes.Health = arch.Attributes.Health }
			if config.Attributes.Intellect.IsZero() { config.Attributes.Intellect = arch.Attributes.Intellect }
			if config.Attributes.Wisdom.IsZero() { config.Attributes.Wisdom = arch.Attributes.Wisdom }
			
			// Inherit State
			if config.State.HealthPoints == 0 { config.State.HealthPoints = arch.State.HealthPoints }
			if config.State.MaxHealthPoints == 0 { config.State.MaxHealthPoints = arch.State.MaxHealthPoints }
			if config.State.Hunger == 0 { config.State.Hunger = arch.State.Hunger }
			if config.State.Thirst == 0 { config.State.Thirst = arch.State.Thirst }
			if config.State.Fatigue == 0 { config.State.Fatigue = arch.State.Fatigue }

			if config.PrimaryColor == "" { config.PrimaryColor = arch.PrimaryColor }
			if config.SecondaryColor == "" { config.SecondaryColor = arch.SecondaryColor }
			if len(config.Footprint) == 0 { config.Footprint = arch.Footprint }
			if config.Weapon.IsEmpty() { config.Weapon = arch.Weapon }
			if config.Dialogues == nil { config.Dialogues = arch.Dialogues }
		}
		sanitizeEntityConfig(config, config.ID)
	}
}

func (r *CharacterRegistry) CountAssets(assets fs.FS, archs *ArchetypeRegistry, permitList map[string]bool) int {
	return len(r.createLoadJobs(assets, archs, permitList))
}
