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

type EntityConfig struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Names       []string `yaml:"names"`
	ArchetypeID string   `yaml:"archetype,omitempty"`
	Behavior    string   `yaml:"behavior"`
	Stats       struct {
		HealthMin       int     `yaml:"health_min"`
		HealthMax       int     `yaml:"health_max"`
		Speed           float64 `yaml:"speed"`
		BaseAttack      int     `yaml:"base_attack"`
		BaseDefense     int     `yaml:"base_defense"`
		AttackCooldown      int     `yaml:"attack_cooldown"`
		AttackRange         float64 `yaml:"attack_range"`
		ProjectileSpeed     float64 `yaml:"projectile_speed"`
	} `yaml:"stats"`
	Actions    *ActionConfig `yaml:"actions,omitempty"`
	Weapon      WeaponConfig  `yaml:"weapon"`
	CollisionRadius float64      `yaml:"collision_radius,omitempty"`

	Footprint      []FootprintPoint `yaml:"footprint"`
	Description    string           `yaml:"description,omitempty"`
	Unique         bool             `yaml:"unique,omitempty"`
	Gender         string           `yaml:"gender,omitempty"`
	SoundID        string           `yaml:"-"` // ID used for audio lookups (e.g. "man_at_arms_male")
	PlayableCharacter  string           `yaml:"-"` // Set to config.ID when this is the active playable character
	PrimaryColor   string           `yaml:"primary_color,omitempty"`
	SecondaryColor string           `yaml:"secondary_color,omitempty"`
	XP             int              `yaml:"xp,omitempty"` // XP awarded on kill
	Group          string           `yaml:"group,omitempty"`
	LeaderID       string           `yaml:"leader,omitempty"`
	MustSurvive    bool             `yaml:"must_survive,omitempty"`
	Playable       bool             `yaml:"playable,omitempty"`
	MaxWeight      float64          `yaml:"max_weight,omitempty"`
	MaxItems       int              `yaml:"max_items,omitempty"`
	Equipment      map[string]string `yaml:"equipment,omitempty"` // map of slot name to object ID
	Inventory      []string         `yaml:"inventory,omitempty"`  // IDs of objects in backpack

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


	CachedBaseFootprint *engine.Polygon `yaml:"-"`

	Dialogues *DialogueRoot `yaml:"dialogues,omitempty"`
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
	jobs := r.createLoadJobs(permitList)
	if len(jobs) > 0 {
		loadSpritesParallel(assets, jobs, graphics, ls)
	}
}

func (r *ArchetypeRegistry) createLoadJobs(permitList map[string]bool) []*SpriteLoadJob {
	var jobs []*SpriteLoadJob
	for _, config := range r.Archetypes {
		if config.StaticImage != nil {
			continue
		}
		if permitList != nil && !permitList[config.ID] {
			continue
		}
		if config.AssetDir == "" {
			continue
		}
		
		addJob := func(filename string, target *engine.Image) {
			jobs = append(jobs, &SpriteLoadJob{
				Path: path.Join(config.AssetDir, filename),
				Dest: target,
			})
		}
		
		addJob("static.png", &config.StaticImage)
		addJob("back.png", &config.BackImage)
		addJob("corpse.png", &config.CorpseImage)
		addJob("attack.png", &config.AttackImage)
		addJob("attack1.png", &config.Attack1Image)
		addJob("attack2.png", &config.Attack2Image)
		addJob("hit.png", &config.HitImage)
		addJob("hit1.png", &config.Hit1Image)
		addJob("hit2.png", &config.Hit2Image)
		addJob("crouch.png", &config.CrouchImage)
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
	const baseDir = "data/archetypes"
	return forEachYAML(assets, baseDir, func(fpath string, data []byte) error {
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
		config.AssetDir = path.Join("assets/images/archetypes", subDir, variantName)
		config.AudioDir = path.Join("assets/audio/archetypes", subDir, variantName)

		// config.Weapon is now auto-loaded by YAML

		config.SoundID = config.ID

		r.Archetypes[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		return nil
	})
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
		lookupID := config.ArchetypeID
		if config.Gender != "" && !strings.Contains(config.ArchetypeID, config.Gender) {
			fullID := config.ArchetypeID + "_" + config.Gender
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

		// Stat fallbacks from archetype
		if arch != nil {
			if config.Stats.HealthMin == 0 { config.Stats.HealthMin = arch.Stats.HealthMin }
			if config.Stats.HealthMax == 0 { config.Stats.HealthMax = arch.Stats.HealthMax }
			if config.Stats.Speed == 0 { config.Stats.Speed = arch.Stats.Speed }
			if config.Stats.BaseAttack == 0 { config.Stats.BaseAttack = arch.Stats.BaseAttack }
			if config.Stats.ProjectileSpeed == 0 { config.Stats.ProjectileSpeed = arch.Stats.ProjectileSpeed }
			if config.Stats.AttackCooldown == 0 { config.Stats.AttackCooldown = arch.Stats.AttackCooldown }
			if config.Stats.BaseDefense == 0 { config.Stats.BaseDefense = arch.Stats.BaseDefense }
			if config.PrimaryColor == "" { config.PrimaryColor = arch.PrimaryColor }
			if config.SecondaryColor == "" { config.SecondaryColor = arch.SecondaryColor }
			if len(config.Footprint) == 0 { config.Footprint = arch.Footprint }
			if config.Weapon.IsEmpty() { config.Weapon = arch.Weapon }
			if config.Dialogues == nil { config.Dialogues = arch.Dialogues }
		}

		// Asset loading jobs
		if config.AssetDir != "" {
			addJob := func(filename string, target *engine.Image, fallback engine.Image) {
				fpath := path.Join(config.AssetDir, filename)
				if _, err := fs.Stat(assets, fpath); err == nil {
					jobs = append(jobs, &SpriteLoadJob{Path: fpath, Dest: target})
				} else {
					*target = fallback
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
		}
		sanitizeEntityConfig(config, config.ID)
	}
	return jobs
}

func (r *CharacterRegistry) CountAssets(assets fs.FS, archs *ArchetypeRegistry, permitList map[string]bool) int {
	return len(r.createLoadJobs(assets, archs, permitList))
}
