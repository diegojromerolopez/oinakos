package game

import (
	"fmt"
	"io/fs"
	"log"
	"oinakos/internal/engine"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ObjectiveType int

const (
	ObjKillVIP ObjectiveType = iota
	ObjReachPortal
	ObjSurvive
	ObjReachZone
	ObjKillCount
	ObjReachBuilding
	ObjProtectNPC
	ObjPacifist
	ObjDestroyBuilding
)

func (t *ObjectiveType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err == nil {
		switch s {
		case "kill_vip":
			*t = ObjKillVIP
			return nil
		case "reach_portal":
			*t = ObjReachPortal
			return nil
		case "survive":
			*t = ObjSurvive
			return nil
		case "reach_zone":
			*t = ObjReachZone
			return nil
		case "kill_count":
			*t = ObjKillCount
			return nil
		case "reach_building":
			*t = ObjReachBuilding
			return nil
		case "protect_npc":
			*t = ObjProtectNPC
			return nil
		case "pacifist":
			*t = ObjPacifist
			return nil
		case "destroy_building":
			*t = ObjDestroyBuilding
			return nil
		}
	}

	var i int
	if err := value.Decode(&i); err == nil {
		*t = ObjectiveType(i)
		return nil
	}

	return fmt.Errorf("unknown objective type: %v", value.Value)
}

type MapType struct {
	ID              string         `yaml:"id"`
	Name            string         `yaml:"name"`
	Type            ObjectiveType  `yaml:"type"`
	Description     string         `yaml:"description"`
	Difficulty      int            `yaml:"difficulty"`
	TargetRadius    float64        `yaml:"target_radius"`
	TargetTime      float64        `yaml:"target_time"`
	TargetKillCount int            `yaml:"target_kill_count"`
	TargetKills     map[string]int `yaml:"target_kills"`
	SpawnFreq       float64        `yaml:"spawn_frequency"`
	SpawnAmount     int            `yaml:"spawn_amount"`
	SpawnWeights    map[string]int `yaml:"spawn_weights"`
	WidthPixels     int            `yaml:"width_px"`
	HeightPixels    int            `yaml:"height_px"`
	MapWidth        float64        `yaml:"-"` // Cartesian width
	MapHeight       float64        `yaml:"-"` // Cartesian height

	TargetNPC      *EntityConfig `yaml:"-"`
	TargetObstacle *Obstacle     `yaml:"-"`
	TargetPoint    engine.Point  `yaml:"-"`
	StartTime      float64       `yaml:"-"`
	IsCompleted    bool          `yaml:"-"`
}

type MapTypeRegistry struct {
	Types map[string]*MapType
	IDs   []string
}

func NewMapTypeRegistry() *MapTypeRegistry {
	return &MapTypeRegistry{
		Types: make(map[string]*MapType),
		IDs:   make([]string, 0),
	}
}

func (r *MapTypeRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	const mapDir = "data/map_types"

	entries, err := fs.ReadDir(assets, mapDir)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory") {
			return nil // Optional directory
		}
		return fmt.Errorf("failed to read maps directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		yamlPath := path.Join(mapDir, entry.Name())
		data, err := fs.ReadFile(assets, yamlPath)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", yamlPath, err)
			continue
		}

		var config MapType
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", yamlPath, err)
			continue
		}

		sanitizeMapType(&config, yamlPath)

		if config.WidthPixels <= 0 {
			config.WidthPixels = 1000000
		}
		if config.HeightPixels <= 0 {
			config.HeightPixels = 1000000
		}
		config.MapWidth = float64(config.WidthPixels) / float64(engine.TileWidth)
		config.MapHeight = float64(config.HeightPixels) / float64(engine.TileHeight)

		r.Types[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Map config: %s (%s)", config.ID, config.Name)
	}

	return nil
}

type EntityConfig struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Names    []string `yaml:"names"`
	Behavior string   `yaml:"behavior"`
	Stats    struct {
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
	Sprites    struct {
		Static string `yaml:"static"`
		Corpse string `yaml:"corpse"`
		Attack string `yaml:"attack"`
	} `yaml:"sprites"`
	FootprintWidth  float64 `yaml:"footprint_width"`
	FootprintHeight float64 `yaml:"footprint_height"`
	Footprint       []struct {
		X float64 `yaml:"x"`
		Y float64 `yaml:"y"`
	} `yaml:"footprint"`

	// Run-time loaded assets
	AssetDir    string      `yaml:"-"`
	StaticImage interface{} `yaml:"-"`
	CorpseImage interface{} `yaml:"-"`
	AttackImage interface{} `yaml:"-"`
	Weapon      *Weapon     `yaml:"-"`
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
	for _, config := range r.Archetypes {
		if config.AssetDir == "" {
			continue
		}
		if config.Sprites.Static != "" {
			config.StaticImage = graphics.LoadSprite(assets, path.Join(config.AssetDir, config.Sprites.Static), true)
		}
		if config.Sprites.Corpse != "" {
			config.CorpseImage = graphics.LoadSprite(assets, path.Join(config.AssetDir, config.Sprites.Corpse), true)
		}
		if config.Sprites.Attack != "" {
			config.AttackImage = graphics.LoadSprite(assets, path.Join(config.AssetDir, config.Sprites.Attack), true)
		}
	}
}

func (r *ArchetypeRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	const baseDir = "data/archetypes"

	err := fs.WalkDir(assets, baseDir, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory") {
				return nil // Skip if baseDir doesn't exist
			}
			return err
		}
		if d.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
			return nil
		}

		// Get relative path from baseDir to the yaml file
		relPath, err := filepath.Rel(baseDir, fpath)
		if err != nil {
			return nil
		}
		// Dir of the YAML file (e.g. "orc" or "orc/male")
		subDir := filepath.Dir(relPath)
		if subDir == "." {
			subDir = ""
		}

		data, err := fs.ReadFile(assets, fpath)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", fpath, err)
			return nil
		}

		var config Archetype
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}

		sanitizeEntityConfig(&config, fpath)

		// Set AssetDir for NPC sprites based on directory structure:
		// assets/images/archetypes/<subDir>/<variantRootName>/
		variantName := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
		config.AssetDir = path.Join("assets/images/archetypes", subDir, variantName)

		// Load images assuming consistent asset folder naming: assets/images/archetypes/<subDir>/<variantFilenameRoot>/
		// If path is orc/female.yaml, imgDir should be assets/images/archetypes/orc/female/
		// imgDir := path.Join("assets/images/archetypes", subDir, filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))]))

		// config.StaticImage = loadSprite(assets, path.Join(imgDir, config.Sprites.Static), true)
		// config.CorpseImage = loadSprite(assets, path.Join(imgDir, config.Sprites.Corpse), true)

		// Link weapon
		config.Weapon = GetWeaponByName(config.WeaponName)

		r.Archetypes[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Archetype: %s (%s) from %s with AssetDir: %s", config.ID, config.Name, fpath, config.AssetDir)
		return nil
	})

	return err
}

type ObstacleArchetype struct {
	ID              string  `yaml:"id"`
	Name            string  `yaml:"name"`
	Type            string  `yaml:"type"` // "static", "well"
	Description     string  `yaml:"description"`
	Health          int     `yaml:"health"`        // Base health
	CooldownTime    float64 `yaml:"cooldown_time"` // Base cooldown in minutes
	ImagePath       string  `yaml:"image"`         // Relative path from data/obstacles/<id>/
	Scale           float64 `yaml:"scale"`
	FootprintWidth  float64 `yaml:"footprint_width"`
	FootprintHeight float64 `yaml:"footprint_height"`
	Footprint       []struct {
		X float64 `yaml:"x"`
		Y float64 `yaml:"y"`
	} `yaml:"footprint"`
	Image interface{} `yaml:"-"`
}

type ObstacleRegistry struct {
	Archetypes map[string]*ObstacleArchetype
	IDs        []string
}

func NewObstacleRegistry() *ObstacleRegistry {
	return &ObstacleRegistry{
		Archetypes: make(map[string]*ObstacleArchetype),
		IDs:        make([]string, 0),
	}
}

func (r *ObstacleRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	const obsDir = "data/obstacles"

	entries, err := fs.ReadDir(assets, obsDir)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory") {
			return nil // Optional directory
		}
		return fmt.Errorf("failed to read obstacles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		yamlPath := path.Join(obsDir, entry.Name())
		data, err := fs.ReadFile(assets, yamlPath)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", yamlPath, err)
			continue
		}

		var config ObstacleArchetype
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", yamlPath, err)
			continue
		}

		sanitizeObstacleArchetype(&config, yamlPath)

		// Store full path from data/obstacles/<id>/ logic
		config.Image = nil // Will be loaded in LoadAssets

		r.Archetypes[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Obstacle Archetype: %s (%s)", config.ID, config.Name)
	}

	return nil
}
func (r *ObstacleRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics) {
	for _, config := range r.Archetypes {
		if config.ImagePath != "" {
			config.Image = graphics.LoadSprite(assets, config.ImagePath, true)
		}
	}
}

func (c *EntityConfig) GetFootprint() engine.Polygon {
	if len(c.Footprint) == 0 {
		// Default rectangular footprint based on width/height
		w := c.FootprintWidth
		if w <= 0 {
			w = 0.3 // default fallback
		}
		h := c.FootprintHeight
		if h <= 0 {
			h = 0.3 // default fallback
		}
		return engine.Polygon{
			Points: []engine.Point{
				{X: -w / 2, Y: -h / 2},
				{X: w / 2, Y: -h / 2},
				{X: w / 2, Y: h / 2},
				{X: -w / 2, Y: h / 2},
			},
		}
	}
	poly := engine.Polygon{Points: make([]engine.Point, len(c.Footprint))}
	for i, p := range c.Footprint {
		poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
	}
	return poly
}

func LoadMainCharacterConfig(assets fs.FS) (*EntityConfig, error) {
	if assets == nil {
		return &EntityConfig{}, nil
	}
	const configPath = "data/characters/main/character.yaml"
	data, err := fs.ReadFile(assets, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read main character config %s: %w", configPath, err)
	}

	var config EntityConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal main character config %s: %w", configPath, err)
	}

	sanitizeEntityConfig(&config, configPath)

	// Link weapon
	config.Weapon = GetWeaponByName(config.WeaponName)

	return &config, nil
}

func GetWeaponByName(name string) *Weapon {
	switch name {
	case "Tizon":
		return WeaponTizon
	case "Orcish Axe":
		return WeaponOrcishAxe
	case "Iron Broadsword":
		return WeaponIronBroadsword
	case "Fists":
		return WeaponFists
	case "Cleaver":
		return &Weapon{Name: "Cleaver", MinDamage: 3, MaxDamage: 7}
	case "Trident":
		return &Weapon{Name: "Trident", MinDamage: 6, MaxDamage: 12}
	case "Whip":
		return &Weapon{Name: "Whip", MinDamage: 4, MaxDamage: 9}
	case "Bow":
		return &Weapon{Name: "Bow", MinDamage: 3, MaxDamage: 6}
	case "Dagger":
		return &Weapon{Name: "Dagger", MinDamage: 2, MaxDamage: 5}
	default:
		return WeaponFists
	}
}
