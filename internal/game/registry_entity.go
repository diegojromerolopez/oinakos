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
		AttackCooldown  int     `yaml:"attack_cooldown"`
		AttackRange     float64 `yaml:"attack_range"`
		ProjectileSpeed float64 `yaml:"projectile_speed"`
	} `yaml:"stats"`
	WeaponName string `yaml:"weapon"`

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

	// Run-time loaded assets
	AssetDir     string      `yaml:"-"`
	AudioDir     string      `yaml:"-"` // e.g. assets/audio/archetypes/orc/male
	StaticImage  interface{} `yaml:"-"`
	BackImage    interface{} `yaml:"-"` // back.png (instead of static.png when facing UP)
	CorpseImage  interface{} `yaml:"-"`
	AttackImage  interface{} `yaml:"-"` // attack.png (default)
	Attack1Image interface{} `yaml:"-"` // attack1.png
	Attack2Image interface{} `yaml:"-"` // attack2.png
	HitImage     interface{} `yaml:"-"` // hit.png  (legacy / single hit frame)
	Hit1Image    interface{} `yaml:"-"` // hit1.png (first variant)
	Hit2Image    interface{} `yaml:"-"` // hit2.png (second variant, requires hit1.png)
	Weapon       *Weapon     `yaml:"-"`
}

func (e *EntityConfig) PickAttackImage(seed int) engine.Image {
	if a1, ok := e.Attack1Image.(engine.Image); ok {
		if a2, ok2 := e.Attack2Image.(engine.Image); ok2 {
			if seed%2 == 0 {
				return a1
			}
			return a2
		}
		return a1
	}
	if a, ok := e.AttackImage.(engine.Image); ok {
		return a
	}
	return nil
}

func (e *EntityConfig) PickHitImage(seed int) engine.Image {
	if h1, ok := e.Hit1Image.(engine.Image); ok {
		if h2, ok2 := e.Hit2Image.(engine.Image); ok2 {
			if seed%2 == 0 {
				return h1
			}
			return h2
		}
		return h1
	}
	if h, ok := e.HitImage.(engine.Image); ok {
		return h
	}
	return nil
}

func (c *EntityConfig) GetFootprint() engine.Polygon {
	if len(c.Footprint) == 0 {
		return engine.Polygon{Points: []engine.Point{
			{X: -0.15, Y: -0.15}, {X: 0.15, Y: -0.15},
			{X: 0.15, Y: 0.15}, {X: -0.15, Y: 0.15},
		}}
	}
	poly := engine.Polygon{Points: make([]engine.Point, len(c.Footprint))}
	for i, p := range c.Footprint {
		poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
	}
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

func (r *ArchetypeRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics) {
	var jobs []*SpriteLoadJob
	for _, config := range r.Archetypes {
		if config.AssetDir == "" {
			continue
		}
		
		addJob := func(filename string, target *interface{}) {
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
	}
	loadSpritesParallel(assets, jobs, graphics)
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

		sanitizeEntityConfig(&config, fpath)

		variantName := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
		config.AssetDir = path.Join("assets/images/archetypes", subDir, variantName)
		config.AudioDir = path.Join("assets/audio/archetypes", subDir, variantName)

		config.Weapon = GetWeaponByName(config.WeaponName)
		config.SoundID = config.ID

		r.Archetypes[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		return nil
	})
}

type PlayableCharacterRegistry struct {
	Characters map[string]*EntityConfig
	IDs        []string
}

func NewPlayableCharacterRegistry() *PlayableCharacterRegistry {
	return &PlayableCharacterRegistry{
		Characters: make(map[string]*EntityConfig),
		IDs:        make([]string, 0),
	}
}

func (r *PlayableCharacterRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	const baseDir = "data/characters"
	return forEachYAML(assets, baseDir, func(fpath string, data []byte) error {
		normalizedPath := filepath.ToSlash(fpath)
		dir := filepath.Dir(normalizedPath)
		if dir != "data/characters" && dir != "oinakos/data/characters" {
			return nil
		}
		var config EntityConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}

		sanitizeEntityConfig(&config, fpath)

		variantName := config.ID
		config.AssetDir = path.Join("assets/images/characters", variantName)
		config.AudioDir = path.Join("assets/audio/characters", variantName)
		config.SoundID = config.ID
		config.PlayableCharacter = config.ID

		config.Weapon = GetWeaponByName(config.WeaponName)

		r.Characters[config.ID] = &config
		if config.Playable {
			r.IDs = append(r.IDs, config.ID)
		}
		return nil
	})
}

func (r *PlayableCharacterRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics) {
	var jobs []*SpriteLoadJob
	for _, config := range r.Characters {
		if config.AssetDir == "" {
			continue
		}
		
		addJob := func(filename string, target *interface{}) {
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
	}
	loadSpritesParallel(assets, jobs, graphics)
}

type NPCRegistry struct {
	NPCs map[string]*EntityConfig
	IDs  []string
}

func NewNPCRegistry() *NPCRegistry {
	return &NPCRegistry{
		NPCs: make(map[string]*EntityConfig),
		IDs:  make([]string, 0),
	}
}

func (r *NPCRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, archs *ArchetypeRegistry) {
	for _, config := range r.NPCs {
		lookupID := config.ArchetypeID
		if config.Gender != "" && !strings.Contains(config.ArchetypeID, config.Gender) {
			fullID := config.ArchetypeID + "_" + config.Gender
			if _, exists := archs.Archetypes[fullID]; exists {
				lookupID = fullID
			}
		}

		arch, ok := archs.Archetypes[lookupID]
		if ok {
			config.SoundID = lookupID
		} else {
			config.SoundID = config.ID
		}

		if config.AudioDir != "" {
			if entries, err := fs.ReadDir(assets, config.AudioDir); err == nil {
				for _, e := range entries {
					if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".wav") {
						config.SoundID = config.ID
						break
					}
				}
			}
		}

		if !ok {
			if config.Unique {
				DebugLog("Unique NPC %s has no archetype, using self-contained config", config.ID)
			} else if config.ArchetypeID != "" {
				log.Printf("Warning: NPC %s uses unknown archetype %s. Proceeding with self-contained config.", config.ID, config.ArchetypeID)
			}
		}

		if arch != nil {
			if config.Stats.HealthMin == 0 {
				config.Stats.HealthMin = arch.Stats.HealthMin
			}
			if config.Stats.HealthMax == 0 {
				config.Stats.HealthMax = arch.Stats.HealthMax
			}
			if config.Stats.Speed == 0 {
				config.Stats.Speed = arch.Stats.Speed
			}
			if config.Stats.BaseAttack == 0 {
				config.Stats.BaseAttack = arch.Stats.BaseAttack
			}
			if config.Stats.ProjectileSpeed == 0 {
				config.Stats.ProjectileSpeed = arch.Stats.ProjectileSpeed
			}
			if config.Stats.AttackCooldown == 0 {
				config.Stats.AttackCooldown = arch.Stats.AttackCooldown
			}
			if config.Stats.BaseDefense == 0 {
				config.Stats.BaseDefense = arch.Stats.BaseDefense
			}

			if config.PrimaryColor == "" {
				config.PrimaryColor = arch.PrimaryColor
			}
			if config.SecondaryColor == "" {
				config.SecondaryColor = arch.SecondaryColor
			}

			if len(config.Footprint) == 0 {
				config.Footprint = arch.Footprint
			}
			if config.WeaponName == "" {
				config.WeaponName = arch.WeaponName
				config.Weapon = arch.Weapon
			}
		}

		staticPath := path.Join(config.AssetDir, "static.png")
		backPath := path.Join(config.AssetDir, "back.png")
		corpsePath := path.Join(config.AssetDir, "corpse.png")

		if arch != nil {
			if _, err := fs.Stat(assets, staticPath); err != nil {
				config.StaticImage = arch.StaticImage
			}
			if _, err := fs.Stat(assets, backPath); err != nil {
				config.BackImage = arch.BackImage
			}
			if _, err := fs.Stat(assets, corpsePath); err != nil {
				config.CorpseImage = arch.CorpseImage
			}
		}

		if config.AssetDir != "" {
			if _, err := fs.Stat(assets, staticPath); err == nil {
				config.StaticImage = graphics.LoadSprite(assets, staticPath, true)
			} else if config.Unique {
				charDir := path.Join("assets/images/characters", config.ID)
				charStaticPath := path.Join(charDir, "static.png")
				if _, err := fs.Stat(assets, charStaticPath); err == nil {
					config.AssetDir = charDir
					config.StaticImage = graphics.LoadSprite(assets, charStaticPath, true)
					staticPath = charStaticPath
					backPath = path.Join(charDir, "back.png")
					corpsePath = path.Join(charDir, "corpse.png")
				}
			}

			if config.StaticImage != nil {
				if _, err := fs.Stat(assets, backPath); err == nil {
					config.BackImage = graphics.LoadSprite(assets, backPath, true)
				}
				if _, err := fs.Stat(assets, corpsePath); err == nil {
					config.CorpseImage = graphics.LoadSprite(assets, corpsePath, true)
				}
				
				a1p := path.Join(config.AssetDir, "attack.png")
				if _, err := fs.Stat(assets, a1p); err == nil {
					config.AttackImage = graphics.LoadSprite(assets, a1p, true)
				}
				a1p = path.Join(config.AssetDir, "attack1.png")
				if _, err := fs.Stat(assets, a1p); err == nil {
					config.Attack1Image = graphics.LoadSprite(assets, a1p, true)
				}
				a2p := path.Join(config.AssetDir, "attack2.png")
				if _, err := fs.Stat(assets, a2p); err == nil {
					config.Attack2Image = graphics.LoadSprite(assets, a2p, true)
				}
				h1p := path.Join(config.AssetDir, "hit.png")
				if _, err := fs.Stat(assets, h1p); err == nil {
					config.HitImage = graphics.LoadSprite(assets, h1p, true)
				}
				h1p = path.Join(config.AssetDir, "hit1.png")
				if _, err := fs.Stat(assets, h1p); err == nil {
					config.Hit1Image = graphics.LoadSprite(assets, h1p, true)
				}
				h2p := path.Join(config.AssetDir, "hit2.png")
				if _, err := fs.Stat(assets, h2p); err == nil {
					config.Hit2Image = graphics.LoadSprite(assets, h2p, true)
				}
			}
		}

		if config.ID == "crimson_guard" {
			log.Printf("NPC %s: P=%s S=%s staticNil=%v", config.ID, config.PrimaryColor, config.SecondaryColor, config.StaticImage == nil)
		}

		sanitizeEntityConfig(config, config.ID)
	}
}

func (r *NPCRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	const baseDir = "data/npcs"
	return forEachYAML(assets, baseDir, func(fpath string, data []byte) error {
		var config EntityConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}

		config.AssetDir = path.Join("assets/images/npcs", config.ID)
		config.AudioDir = path.Join("assets/audio/npcs", config.ID)

		config.Weapon = GetWeaponByName(config.WeaponName)

		log.Printf("DEBUG: NPC Registry loading %s from %s", config.ID, fpath)
		r.NPCs[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		return nil
	})
}
