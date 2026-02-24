package game

import (
	"fmt"
	"io/fs"
	"log"
	"oinakos/internal/engine"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
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

type MapData struct {
	ID           string         `yaml:"id"`
	Name         string         `yaml:"name"`
	Type         ObjectiveType  `yaml:"type"`
	Description  string         `yaml:"description"`
	Difficulty   int            `yaml:"difficulty"`
	TargetRadius float64        `yaml:"target_radius"`
	TargetTime   float64        `yaml:"target_time"`
	TargetKills  map[string]int `yaml:"target_kills"`
	SpawnFreq    float64        `yaml:"spawn_frequency"`
	SpawnAmount  int            `yaml:"spawn_amount"`
	SpawnWeights map[string]int `yaml:"spawn_weights"`

	TargetNPC      *EntityConfig `yaml:"-"`
	TargetObstacle *Obstacle     `yaml:"-"`
	TargetPoint    engine.Point  `yaml:"-"`
	StartTime      float64       `yaml:"-"`
	IsCompleted    bool          `yaml:"-"`
}

type MapRegistry struct {
	Maps map[string]*MapData
	IDs  []string
}

func NewMapRegistry() *MapRegistry {
	return &MapRegistry{
		Maps: make(map[string]*MapData),
		IDs:  make([]string, 0),
	}
}

func (r *MapRegistry) LoadAll(assets fs.FS) error {
	const mapDir = "data/maps"

	entries, err := fs.ReadDir(assets, mapDir)
	if err != nil {
		return fmt.Errorf("failed to read maps directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		yamlPath := filepath.Join(mapDir, entry.Name())
		data, err := fs.ReadFile(assets, yamlPath)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", yamlPath, err)
			continue
		}

		var config MapData
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", yamlPath, err)
			continue
		}

		r.Maps[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Map config: %s (%s)", config.ID, config.Name)
	}

	return nil
}

type EntityConfig struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Stats struct {
		HealthMin      int     `yaml:"health_min"`
		HealthMax      int     `yaml:"health_max"`
		BaseAttack     int     `yaml:"base_attack"`
		BaseDefense    int     `yaml:"base_defense"`
		Speed          float64 `yaml:"speed"`
		AttackCooldown int     `yaml:"attack_cooldown"`
	} `yaml:"stats"`
	WeaponName string `yaml:"weapon"`
	Sprites    struct {
		Static string `yaml:"static"`
		Corpse string `yaml:"corpse"`
		Attack string `yaml:"attack"`
	} `yaml:"sprites"`
	Footprint []struct {
		X float64 `yaml:"x"`
		Y float64 `yaml:"y"`
	} `yaml:"footprint"`

	// Run-time loaded assets
	StaticImage *ebiten.Image `yaml:"-"`
	CorpseImage *ebiten.Image `yaml:"-"`
	AttackImage *ebiten.Image `yaml:"-"`
	Weapon      *Weapon       `yaml:"-"`
}

type NPCConfig = EntityConfig

type NPCRegistry struct {
	Configs map[string]*NPCConfig
	IDs     []string
}

func NewNPCRegistry() *NPCRegistry {
	return &NPCRegistry{
		Configs: make(map[string]*NPCConfig),
		IDs:     make([]string, 0),
	}
}

func (r *NPCRegistry) LoadAll(assets fs.FS) error {
	const baseDir = "data/npcs"

	entries, err := fs.ReadDir(assets, baseDir)
	if err != nil {
		return fmt.Errorf("failed to read npcs directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		yamlPath := filepath.Join(baseDir, entry.Name())

		data, err := fs.ReadFile(assets, yamlPath)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", yamlPath, err)
			continue
		}

		var config NPCConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", yamlPath, err)
			continue
		}

		// Load images assuming consistent asset folder naming: assets/images/npcs/<id>/
		imgDir := filepath.Join("assets/images/npcs", config.ID)
		config.StaticImage = loadSprite(assets, filepath.Join(imgDir, config.Sprites.Static), true)
		config.CorpseImage = loadSprite(assets, filepath.Join(imgDir, config.Sprites.Corpse), true)
		if config.Sprites.Attack != "" {
			config.AttackImage = loadSprite(assets, filepath.Join(imgDir, config.Sprites.Attack), true)
		}

		// Link weapon
		config.Weapon = GetWeaponByName(config.WeaponName)

		r.Configs[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded NPC config: %s (%s)", config.ID, config.Name)
	}

	return nil
}

type ObstacleConfig struct {
	ID              string        `yaml:"id"`
	Name            string        `yaml:"name"`
	Description     string        `yaml:"description"`
	Health          int           `yaml:"health"` // Base health
	ImagePath       string        `yaml:"image"`  // Relative path from data/obstacles/<id>/
	Scale           float64       `yaml:"scale"`
	FootprintWidth  float64       `yaml:"footprint_width"`
	FootprintHeight float64       `yaml:"footprint_height"`
	Image           *ebiten.Image `yaml:"-"`
}

type ObstacleRegistry struct {
	Configs map[string]*ObstacleConfig
	IDs     []string
}

func NewObstacleRegistry() *ObstacleRegistry {
	return &ObstacleRegistry{
		Configs: make(map[string]*ObstacleConfig),
		IDs:     make([]string, 0),
	}
}

func (r *ObstacleRegistry) LoadAll(assets fs.FS) error {
	const obsDir = "data/obstacles"

	entries, err := fs.ReadDir(assets, obsDir)
	if err != nil {
		return fmt.Errorf("failed to read obstacles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		yamlPath := filepath.Join(obsDir, entry.Name())
		data, err := fs.ReadFile(assets, yamlPath)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", yamlPath, err)
			continue
		}

		var config ObstacleConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", yamlPath, err)
			continue
		}

		// Load Sprites - ImagePath is relative to the assets root (e.g. assets/images/...)
		config.Image = loadSprite(assets, config.ImagePath, true)

		r.Configs[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Obstacle config: %s (%s)", config.ID, config.Name)
	}

	return nil
}

func (c *EntityConfig) GetFootprint() engine.Polygon {
	poly := engine.Polygon{Points: make([]engine.Point, len(c.Footprint))}
	for i, p := range c.Footprint {
		poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
	}
	return poly
}

func LoadPlayerConfig(assets fs.FS) (*EntityConfig, error) {
	const configPath = "data/player/character.yaml"
	data, err := fs.ReadFile(assets, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read player config %s: %w", configPath, err)
	}

	var config EntityConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal player config %s: %w", configPath, err)
	}

	imgDir := "assets/images/player"
	config.StaticImage = loadSprite(assets, filepath.Join(imgDir, config.Sprites.Static), true)
	config.CorpseImage = loadSprite(assets, filepath.Join(imgDir, config.Sprites.Corpse), true)
	if config.Sprites.Attack != "" {
		config.AttackImage = loadSprite(assets, filepath.Join(imgDir, config.Sprites.Attack), true)
	}

	// Link weapon
	config.Weapon = GetWeaponByName(config.WeaponName)

	return &config, nil
}

func GetWeaponByName(name string) *Weapon {
	switch name {
	case "Orcish Axe":
		return WeaponOrcishAxe
	case "Iron Broadsword":
		return WeaponIronBroadsword
	case "Rusty Sword":
		return WeaponRustySword
	case "Fists":
		return WeaponFists
	default:
		return WeaponFists
	}
}
