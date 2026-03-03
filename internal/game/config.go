package game

import (
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"oinakos/internal/engine"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// forEachYAML iterates over YAML files in a base directory from both the embedded FS
// and the local oinakos/data override directory.
func forEachYAML(assets fs.FS, baseDir string, callback func(fpath string, data []byte) error) error {
	// 1. Embedded assets
	if assets != nil {
		fs.WalkDir(assets, baseDir, func(fpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip if not found in embedded
			}
			if d.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
				return nil
			}
			data, err := fs.ReadFile(assets, fpath)
			if err == nil {
				callback(fpath, data)
			}
			return nil
		})
	}

	// 2. Local oinakos/data overrides
	localBaseDir := filepath.Join("oinakos", baseDir)
	if _, err := os.Stat(localBaseDir); err == nil {
		filepath.WalkDir(localBaseDir, func(fpath string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
				return nil
			}
			data, err := os.ReadFile(fpath)
			if err == nil {
				callback(fpath, data)
			}
			return nil
		})
	}
	return nil
}

type ObstacleType string

const (
	TypeBuilding ObstacleType = "building"
	TypeTree     ObstacleType = "tree"
	TypeRock     ObstacleType = "rock"
	TypeResource ObstacleType = "resource"
	TypeBush     ObstacleType = "bush"
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

func (a *Alignment) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err == nil {
		switch s {
		case "enemy":
			*a = AlignmentEnemy
			return nil
		case "neutral":
			*a = AlignmentNeutral
			return nil
		case "ally":
			*a = AlignmentAlly
			return nil
		}
	}
	var i int
	if err := value.Decode(&i); err == nil {
		*a = Alignment(i)
		return nil
	}
	return fmt.Errorf("unknown alignment: %v", value.Value)
}

type Inhabitant struct {
	ID        string    `yaml:"id,omitempty"` // For internal mapping if needed
	Name      string    `yaml:"name,omitempty"`
	Archetype string    `yaml:"archetype,omitempty"`
	NPC       string    `yaml:"npc,omitempty"`
	X         float64   `yaml:"x"`
	Y         float64   `yaml:"y"`
	State     string    `yaml:"state,omitempty"` // e.g. "dead", empty means alive
	Alignment Alignment `yaml:"alignment"`
}

type PreSpawnObstacle struct {
	ID        string   `yaml:"id"`
	Archetype string   `yaml:"archetype"`
	X         *float64 `yaml:"x,omitempty"`
	Y         *float64 `yaml:"y,omitempty"`
	Disabled  bool     `yaml:"disabled,omitempty"`
}

type SpawnConfig struct {
	Archetype   string    `yaml:"archetype"`
	Alignment   Alignment `yaml:"alignment"`
	Probability float64   `yaml:"probability"` // 0.0 to 1.0
	Frequency   float64   `yaml:"frequency"`   // seconds
	X           *float64  `yaml:"x,omitempty"`
	Y           *float64  `yaml:"y,omitempty"`

	Timer int `yaml:"-"` // Internal tick counter
}

// TargetPointConfig is used to parse target_point from YAML.
type TargetPointConfig struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}

type MapType struct {
	ID              string             `yaml:"id"`
	Name            string             `yaml:"name"`
	Type            ObjectiveType      `yaml:"type"`
	Description     string             `yaml:"description"`
	Difficulty      int                `yaml:"difficulty"`
	TargetRadius    float64            `yaml:"target_radius"`
	TargetTime      float64            `yaml:"target_time"`
	TargetKillCount int                `yaml:"target_kill_count"`
	TargetKills     map[string]int     `yaml:"target_kills"`
	WidthPixels     int                `yaml:"width_px"`
	HeightPixels    int                `yaml:"height_px"`
	Inhabitants     []Inhabitant       `yaml:"inhabitants"`
	Spawns          []SpawnConfig      `yaml:"spawns"`
	Obstacles       []PreSpawnObstacle `yaml:"obstacles"`
	FloorTile       string             `yaml:"floor_tile"`
	TargetPointRaw  *TargetPointConfig `yaml:"target_point"` // Optional YAML-supplied target point
	MapWidth        float64            `yaml:"-"`            // Cartesian width
	MapHeight       float64            `yaml:"-"`            // Cartesian height

	TargetNPC      *EntityConfig `yaml:"-"`
	TargetObstacle *Obstacle     `yaml:"-"`
	TargetPoint    engine.Point  `yaml:"-"` // Resolved at loadMapLevel time
	StartTime      float64       `yaml:"-"`
	IsCompleted    bool          `yaml:"-"`
}

type MapTypeRegistry struct {
	Types map[string]*MapType
	IDs   []string
}

type Campaign struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Maps        []string `yaml:"maps"` // Map IDs in sequence
}

type CampaignRegistry struct {
	Campaigns map[string]*Campaign
	IDs       []string
}

func NewCampaignRegistry() *CampaignRegistry {
	return &CampaignRegistry{
		Campaigns: make(map[string]*Campaign),
		IDs:       make([]string, 0),
	}
}

func (r *CampaignRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	const campaignDir = "data/campaigns"
	return forEachYAML(assets, campaignDir, func(fpath string, data []byte) error {
		var config Campaign
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}
		if config.ID == "" {
			config.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
		}
		r.Campaigns[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Campaign: %s (%s) from %s", config.ID, config.Name, fpath)
		return nil
	})
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
	return forEachYAML(assets, mapDir, func(fpath string, data []byte) error {
		var config MapType
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}
		sanitizeMapType(&config, fpath)
		if config.WidthPixels <= 0 {
			config.WidthPixels = 1000000
		}
		if config.HeightPixels <= 0 {
			config.HeightPixels = 1000000
		}
		config.MapWidth = float64(config.WidthPixels) / float64(engine.TileWidth)
		config.MapHeight = float64(config.HeightPixels) / float64(engine.TileHeight)
		if config.FloorTile == "" {
			config.FloorTile = "grass.png"
		}
		r.Types[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Map config: %s (%s) from %s", config.ID, config.Name, fpath)
		return nil
	})
}

type FootprintPoint struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}

// MarshalYAML emits the point as a plain YAML mapping without quoting the 'y'
// key. gopkg.in/yaml.v3 quotes 'y' by default because it is a YAML 1.1 boolean
// synonym. Using explicit !!float tags prevents that behaviour.
func (p FootprintPoint) MarshalYAML() (interface{}, error) {
	// Helper to ensure values always have a decimal point (e.g., 2 -> 2.0)
	// without using the explicit !!float tag.
	format := func(f float64) string {
		s := fmt.Sprintf("%g", f)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") {
			return s + ".0"
		}
		return s
	}
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "x"},
			{Kind: yaml.ScalarNode, Value: format(p.X)},
			{Kind: yaml.ScalarNode, Value: "y"},
			{Kind: yaml.ScalarNode, Value: format(p.Y)},
		},
	}, nil
}

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
	PrimaryColor   string           `yaml:"primary_color,omitempty"`
	SecondaryColor string           `yaml:"secondary_color,omitempty"`
	XP             int              `yaml:"xp,omitempty"` // XP awarded on kill

	// Run-time loaded assets
	AssetDir     string      `yaml:"-"`
	AudioDir     string      `yaml:"-"` // e.g. assets/audio/archetypes/orc/male
	StaticImage  interface{} `yaml:"-"`
	CorpseImage  interface{} `yaml:"-"`
	AttackImage  interface{} `yaml:"-"` // attack.png (default)
	Attack1Image interface{} `yaml:"-"` // attack1.png
	Attack2Image interface{} `yaml:"-"` // attack2.png
	HitImage     interface{} `yaml:"-"` // hit.png  (legacy / single hit frame)
	Hit1Image    interface{} `yaml:"-"` // hit1.png (first variant)
	Hit2Image    interface{} `yaml:"-"` // hit2.png (second variant, requires hit1.png)
	Weapon       *Weapon     `yaml:"-"`
}

// PickAttackImage returns the sprite to display when this entity attacks.
// It follows the same logic as PickHitImage.
func (e *EntityConfig) PickAttackImage() engine.Image {
	if a1, ok := e.Attack1Image.(engine.Image); ok {
		if a2, ok2 := e.Attack2Image.(engine.Image); ok2 {
			if rand.Intn(2) == 0 {
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

// PickHitImage returns the sprite to display when this entity is hit.
//   - If hit1.png and hit2.png both exist → randomly pick one.
//   - Else if hit.png exists              → use hit.png.
//   - Else                                → nil (caller uses static image).
func (e *EntityConfig) PickHitImage() engine.Image {
	if h1, ok := e.Hit1Image.(engine.Image); ok {
		if h2, ok2 := e.Hit2Image.(engine.Image); ok2 {
			if rand.Intn(2) == 0 {
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
		staticPath := path.Join(config.AssetDir, "static.png")
		if _, err := fs.Stat(assets, staticPath); err == nil {
			config.StaticImage = graphics.LoadSprite(assets, staticPath, true)
		}

		corpsePath := path.Join(config.AssetDir, "corpse.png")
		if _, err := fs.Stat(assets, corpsePath); err == nil {
			config.CorpseImage = graphics.LoadSprite(assets, corpsePath, true)
		}

		attackPath := path.Join(config.AssetDir, "attack.png")
		if _, err := fs.Stat(assets, attackPath); err == nil {
			config.AttackImage = graphics.LoadSprite(assets, attackPath, true)
		}
		attack1Path := path.Join(config.AssetDir, "attack1.png")
		if _, err := fs.Stat(assets, attack1Path); err == nil {
			config.Attack1Image = graphics.LoadSprite(assets, attack1Path, true)
		}
		attack2Path := path.Join(config.AssetDir, "attack2.png")
		if _, err := fs.Stat(assets, attack2Path); err == nil {
			config.Attack2Image = graphics.LoadSprite(assets, attack2Path, true)
		}

		hitPath := path.Join(config.AssetDir, "hit.png")
		if _, err := fs.Stat(assets, hitPath); err == nil {
			config.HitImage = graphics.LoadSprite(assets, hitPath, true)
		}
		hit1Path := path.Join(config.AssetDir, "hit1.png")
		if _, err := fs.Stat(assets, hit1Path); err == nil {
			config.Hit1Image = graphics.LoadSprite(assets, hit1Path, true)
		}
		hit2Path := path.Join(config.AssetDir, "hit2.png")
		if _, err := fs.Stat(assets, hit2Path); err == nil {
			config.Hit2Image = graphics.LoadSprite(assets, hit2Path, true)
		}
	}
}

func (r *ArchetypeRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
	const baseDir = "data/archetypes"
	return forEachYAML(assets, baseDir, func(fpath string, data []byte) error {
		// Get relative path from baseDir to the yaml file
		// NOTE: if fpath comes from oinakos/data, it will have that prefix.
		// We need to strip it to get the correct subDir/assetDir mapping.
		cleanRelPath := fpath
		if strings.HasPrefix(fpath, "oinakos"+string(filepath.Separator)) {
			cleanRelPath = fpath[len("oinakos"+string(filepath.Separator)):]
		}

		relPath, err := filepath.Rel(baseDir, cleanRelPath)
		if err != nil {
			return nil
		}
		// Dir of the YAML file (e.g. "orc" or "orc/male")
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

		// Set AssetDir for NPC sprites based on directory structure:
		// assets/images/archetypes/<subDir>/<variantRootName>/
		variantName := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
		config.AssetDir = path.Join("assets/images/archetypes", subDir, variantName)
		config.AudioDir = path.Join("assets/audio/archetypes", subDir, variantName)

		// Link weapon
		config.Weapon = GetWeaponByName(config.WeaponName)

		r.Archetypes[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Archetype: %s (%s) from %s with AssetDir: %s", config.ID, config.Name, fpath, config.AssetDir)
		return nil
	})
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
		// Inheritance from Archetype
		if config.ArchetypeID != "" && archs != nil {
			if arch, ok := archs.Archetypes[config.ArchetypeID]; ok {
				// Copy missing metadata/stats
				if config.Behavior == "" {
					config.Behavior = arch.Behavior
				}
				if config.Stats.HealthMin == 0 {
					config.Stats = arch.Stats
				}
				if len(config.Footprint) == 0 {
					config.Footprint = arch.Footprint
				}
				if config.WeaponName == "" {
					config.WeaponName = arch.WeaponName
					config.Weapon = arch.Weapon
				}

				// Copy images if NPC doesn't have its own
				staticPath := path.Join(config.AssetDir, "static.png")
				if _, err := fs.Stat(assets, staticPath); err != nil {
					config.StaticImage = arch.StaticImage
				}
				corpsePath := path.Join(config.AssetDir, "corpse.png")
				if _, err := fs.Stat(assets, corpsePath); err != nil {
					config.CorpseImage = arch.CorpseImage
				}
				attackPath := path.Join(config.AssetDir, "attack.png")
				if _, err := fs.Stat(assets, attackPath); err != nil {
					config.AttackImage = arch.AttackImage
				}
				attack1Path := path.Join(config.AssetDir, "attack1.png")
				if _, err := fs.Stat(assets, attack1Path); err != nil {
					config.Attack1Image = arch.Attack1Image
				}
				attack2Path := path.Join(config.AssetDir, "attack2.png")
				if _, err := fs.Stat(assets, attack2Path); err != nil {
					config.Attack2Image = arch.Attack2Image
				}
				hitPath := path.Join(config.AssetDir, "hit.png")
				if _, err := fs.Stat(assets, hitPath); err != nil {
					config.HitImage = arch.HitImage
				}
				hit1Path := path.Join(config.AssetDir, "hit1.png")
				if _, err := fs.Stat(assets, hit1Path); err != nil {
					config.Hit1Image = arch.Hit1Image
				}
				hit2Path := path.Join(config.AssetDir, "hit2.png")
				if _, err := fs.Stat(assets, hit2Path); err != nil {
					config.Hit2Image = arch.Hit2Image
				}
			}
		}

		// Load unique assets if they exist (overriding or initial)
		staticPath := path.Join(config.AssetDir, "static.png")
		if _, err := fs.Stat(assets, staticPath); err == nil {
			config.StaticImage = graphics.LoadSprite(assets, staticPath, true)
		}
		corpsePath := path.Join(config.AssetDir, "corpse.png")
		if _, err := fs.Stat(assets, corpsePath); err == nil {
			config.CorpseImage = graphics.LoadSprite(assets, corpsePath, true)
		}
		attackPath := path.Join(config.AssetDir, "attack.png")
		if _, err := fs.Stat(assets, attackPath); err == nil {
			config.AttackImage = graphics.LoadSprite(assets, attackPath, true)
		}
		attack1Path := path.Join(config.AssetDir, "attack1.png")
		if _, err := fs.Stat(assets, attack1Path); err == nil {
			config.Attack1Image = graphics.LoadSprite(assets, attack1Path, true)
		}
		attack2Path := path.Join(config.AssetDir, "attack2.png")
		if _, err := fs.Stat(assets, attack2Path); err == nil {
			config.Attack2Image = graphics.LoadSprite(assets, attack2Path, true)
		}

		hitPath := path.Join(config.AssetDir, "hit.png")
		if _, err := fs.Stat(assets, hitPath); err == nil {
			config.HitImage = graphics.LoadSprite(assets, hitPath, true)
		}
		hit1Path := path.Join(config.AssetDir, "hit1.png")
		if _, err := fs.Stat(assets, hit1Path); err == nil {
			config.Hit1Image = graphics.LoadSprite(assets, hit1Path, true)
		}
		hit2Path := path.Join(config.AssetDir, "hit2.png")
		if _, err := fs.Stat(assets, hit2Path); err == nil {
			config.Hit2Image = graphics.LoadSprite(assets, hit2Path, true)
		}
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

		sanitizeEntityConfig(&config, fpath)

		// Set AssetDir for unique NPC sprites:
		// assets/images/npcs/<id>/
		config.AssetDir = path.Join("assets/images/npcs", config.ID)

		// Link weapon
		config.Weapon = GetWeaponByName(config.WeaponName)

		r.NPCs[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Unique NPC: %s (%s) from %s", config.ID, config.Name, fpath)
		return nil
	})
}

type ObstacleArchetype struct {
	ID             string           `yaml:"id"`
	Name           string           `yaml:"name"`
	Type           ObstacleType     `yaml:"type"`
	Destructible   bool             `yaml:"destructible"` // If false, cannot be damaged
	Description    string           `yaml:"description"`
	Health         int              `yaml:"health"`        // Base health (ignored if Destructible is false)
	CooldownTime   float64          `yaml:"cooldown_time"` // Base cooldown in minutes
	Footprint      []FootprintPoint `yaml:"footprint"`
	FrameCount     int              `yaml:"frame_count"`     // Total number of frames
	FramesPerRow   int              `yaml:"frames_per_row"`  // For grid-based spritesheets (default 0 = single row)
	AnimationSpeed int              `yaml:"animation_speed"` // Ticks per frame
	Image          interface{}      `yaml:"-"`
}

func (a *ObstacleArchetype) IsWell() bool {
	return a.CooldownTime > 0
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
	return forEachYAML(assets, obsDir, func(fpath string, data []byte) error {
		var config ObstacleArchetype
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}

		sanitizeObstacleArchetype(&config, fpath)

		if config.ID == "" {
			config.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
		}

		r.Archetypes[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		log.Printf("Loaded Obstacle Archetype: %s (%s) from %s", config.ID, config.Name, fpath)
		return nil
	})
}
func (r *ObstacleRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics) {
	for _, config := range r.Archetypes {
		// Derive image path from ID: assets/images/obstacles/<id>.png
		imagePath := path.Join("assets/images/obstacles", config.ID+".png")
		config.Image = graphics.LoadSprite(assets, imagePath, true)
	}
}

func (c *EntityConfig) GetFootprint() engine.Polygon {
	if len(c.Footprint) == 0 {
		// Fallback for missing footprint
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

func LoadMainCharacterConfig(assets fs.FS) (*EntityConfig, error) {
	if assets == nil {
		return &EntityConfig{}, nil
	}
	const configPath = "data/characters/main/character.yaml"
	localPath := filepath.Join("oinakos", configPath)

	var data []byte
	var err error

	// Try local override first
	if _, errStat := os.Stat(localPath); errStat == nil {
		data, err = os.ReadFile(localPath)
	}
	if data == nil {
		data, err = fs.ReadFile(assets, configPath)
	}
	if err != nil {
		// Fallback to direct OS read of regular path
		data, err = os.ReadFile(configPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read main character config: %w", err)
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
	case "Gilded Pitchfork":
		return &Weapon{Name: "Gilded Pitchfork", MinDamage: 8, MaxDamage: 18}
	case "Shouts":
		return &Weapon{Name: "Shouts", MinDamage: 10, MaxDamage: 50}
	default:
		return WeaponFists
	}
}
