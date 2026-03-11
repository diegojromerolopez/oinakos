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

type ObstacleActionType string

const (
	ActionHarm ObstacleActionType = "harm"
	ActionHeal ObstacleActionType = "heal"
)

type ObstacleActionConfig struct {
	Type                ObstacleActionType `yaml:"type"`
	Amount              int                `yaml:"amount"`
	Aura                float64            `yaml:"aura"`
	AlignmentLimit      string             `yaml:"alignment_limit"`      // "all", "ally", "enemy"
	RequiresInteraction bool               `yaml:"requires_interaction"` // e.g. the Well
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

func (t ObjectiveType) String() string {
	switch t {
	case ObjKillVIP:
		return "kill_vip"
	case ObjReachPortal:
		return "reach_portal"
	case ObjSurvive:
		return "survive"
	case ObjReachZone:
		return "reach_zone"
	case ObjKillCount:
		return "kill_count"
	case ObjReachBuilding:
		return "reach_building"
	case ObjProtectNPC:
		return "protect_npc"
	case ObjPacifist:
		return "pacifist"
	case ObjDestroyBuilding:
		return "destroy_building"
	}
	return "unknown"
}

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
	ID          string    `yaml:"id,omitempty"` // For internal mapping if needed
	Name        string    `yaml:"name,omitempty"`
	Archetype   string    `yaml:"archetype,omitempty"`
	ArchetypeID string    `yaml:"archetype_id,omitempty"`
	NPC         string    `yaml:"npc,omitempty"`
	NPCID       string    `yaml:"npc_id,omitempty"`
	X           float64   `yaml:"x"`
	Y           float64   `yaml:"y"`
	State       string    `yaml:"state,omitempty"` // e.g. "dead", empty means alive
	Alignment   Alignment `yaml:"alignment"`
	MustSurvive bool      `yaml:"must_survive,omitempty"`
}

type PreSpawnObstacle struct {
	ID          string   `yaml:"id"`
	Archetype   string   `yaml:"archetype"`
	ArchetypeID string   `yaml:"archetype_id,omitempty"`
	X           *float64 `yaml:"x,omitempty"`
	Y           *float64 `yaml:"y,omitempty"`
	Disabled    bool     `yaml:"disabled,omitempty"`
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

type FloorZone struct {
	Name      string           `yaml:"name"`
	Tile      string           `yaml:"tile"`
	Priority  int              `yaml:"priority"`
	Perimeter []FootprintPoint `yaml:"perimeter"`
	Polygon   engine.Polygon   `yaml:"-"`
}

func (fz *FloorZone) GetPolygon() engine.Polygon {
	if len(fz.Polygon.Points) > 0 {
		return fz.Polygon
	}
	pts := make([]engine.Point, len(fz.Perimeter))
	for i, pt := range fz.Perimeter {
		pts[i] = engine.Point{X: pt.X, Y: pt.Y}
	}
	fz.Polygon = engine.Polygon{Points: pts}
	return fz.Polygon
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
	NPCs            []Inhabitant       `yaml:"npcs,omitempty"`
	Spawns          []SpawnConfig      `yaml:"spawns"`
	Obstacles       []PreSpawnObstacle `yaml:"obstacles"`
	FloorTile       string             `yaml:"floor_tile"`
	FloorZones      []*FloorZone       `yaml:"floor_zones"`
	TargetPointRaw  *TargetPointConfig `yaml:"target_point"` // Optional YAML-supplied target point
	Player          *TargetPointConfig `yaml:"player,omitempty"`
	MapWidth        float64            `yaml:"-"` // Cartesian width
	MapHeight       float64            `yaml:"-"` // Cartesian height

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
		normalizedPath := filepath.ToSlash(fpath)
		dir := filepath.Dir(normalizedPath)
		// Only accept files directly in campaignDir or oinakos/campaignDir
		if dir != "data/campaigns" && dir != "oinakos/data/campaigns" {
			// Skip files in subdirectories (which are campaign maps)
			return nil
		}

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
	dirs := []string{"data/map_types", "data/maps", "data/campaigns"}
	for _, loadDir := range dirs {
		forEachYAML(assets, loadDir, func(fpath string, data []byte) error {
			normalizedPath := filepath.ToSlash(fpath)
			dir := filepath.Dir(normalizedPath)

			// Skip top-level files in campaigns (which are campaigns, not map levels)
			if dir == "data/campaigns" || dir == "oinakos/data/campaigns" {
				return nil
			}

			var config MapType
			if err := yaml.Unmarshal(data, &config); err != nil {
				log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
				return nil
			}

			// Detect if someone accidentally loaded a campaign object (it will have maps but no difficulty)
			// But for simplicity we just allow loading anything that passes Unmarshal.

			// Auto ID assignment
			if config.ID == "" {
				config.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
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

			// Only add if we actually parsed something meaningful or just unconditionally
			r.Types[config.ID] = &config

			// Skip adding campaign-specific maps to the UI selector list
			if strings.Contains(normalizedPath, "data/campaigns/") {
				return nil
			}

			// Add to ID list if not already there
			found := false
			for _, id := range r.IDs {
				if id == config.ID {
					found = true
					break
				}
			}
			if !found {
				r.IDs = append(r.IDs, config.ID)
			}
			return nil
		})
	}
	return nil
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
	Gender         string           `yaml:"gender,omitempty"`
	SoundID        string           `yaml:"-"` // ID used for audio lookups (e.g. "man_at_arms_male")
	PlayableCharacter  string           `yaml:"-"` // Set to config.ID when this is the active playable character
	PrimaryColor   string           `yaml:"primary_color,omitempty"`
	SecondaryColor string           `yaml:"secondary_color,omitempty"`
	XP             int              `yaml:"xp,omitempty"` // XP awarded on kill
	Group          string           `yaml:"group,omitempty"`
	LeaderID       string           `yaml:"leader,omitempty"`
	MustSurvive    bool             `yaml:"must_survive,omitempty"`

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

// PickAttackImage returns the sprite to display when this entity attacks.
// It follows the same logic as PickHitImage.
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

// PickHitImage returns the sprite to display when this entity is hit.
//   - If hit1.png and hit2.png both exist → randomly pick one.
//   - Else if hit.png exists              → use hit.png.
//   - Else                                → nil (caller uses static image).
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
	log.Printf(">>>> LOADING ARCHETYPE ASSETS: found %d archetypes in registry", len(r.Archetypes))
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
		// Only accept files directly in data/characters or oinakos/data/characters
		if dir != "data/characters" && dir != "oinakos/data/characters" {
			return nil
		}
		var config EntityConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}

		sanitizeEntityConfig(&config, fpath)

		// Set AssetDir for playable character sprites
		// assets/images/characters/variant_name/
		variantName := config.ID
		config.AssetDir = path.Join("assets/images/characters", variantName)
		config.AudioDir = path.Join("assets/audio/characters", variantName)
		config.SoundID = config.ID
		config.PlayableCharacter = config.ID

		// Link weapon
		config.Weapon = GetWeaponByName(config.WeaponName)

		r.Characters[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
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
			// Try archetype_gender if not already specific
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

		// Override if NPC has its own audio files in assets/audio/npcs/<id>
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
			// If it's in the NPC Registry, it's an NPC profile.
			// We should allow it to exist even without an archetype if it's defined in data/npcs,
			// though it might be missing base stats/assets if not self-contained.
			if config.Unique {
				DebugLog("Unique NPC %s has no archetype, using self-contained config", config.ID)
			} else if config.ArchetypeID != "" {
				log.Printf("Warning: NPC %s uses unknown archetype %s. Proceeding with self-contained config.", config.ID, config.ArchetypeID)
			}
		}

		// Inherit stats if they are empty
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

		// Copy images from archetype if NPC doesn't have its own
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

		// Load unique assets if they exist (overriding or initial)
		if config.AssetDir != "" {
			if _, err := fs.Stat(assets, staticPath); err == nil {
				config.StaticImage = graphics.LoadSprite(assets, staticPath, true)
			} else if config.Unique {
				// Fallback to characters directory for unique NPCs who are also playable heroes
				charDir := path.Join("assets/images/characters", config.ID)
				charStaticPath := path.Join(charDir, "static.png")
				if _, err := fs.Stat(assets, charStaticPath); err == nil {
					config.AssetDir = charDir
					config.StaticImage = graphics.LoadSprite(assets, charStaticPath, true)
					// Also update paths for other frames
					staticPath = charStaticPath
					backPath = path.Join(charDir, "back.png")
					corpsePath = path.Join(charDir, "corpse.png")
				}
			}

			// Load remaining frames from the (possibly updated) AssetDir
			if config.StaticImage != nil {
				if _, err := fs.Stat(assets, backPath); err == nil {
					config.BackImage = graphics.LoadSprite(assets, backPath, true)
				}
				if _, err := fs.Stat(assets, corpsePath); err == nil {
					config.CorpseImage = graphics.LoadSprite(assets, corpsePath, true)
				}
				
				// Optional variants
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

		// Final sanitize after merge and asset loading
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

		// We defer sanitization until LoadAssets, where we merge with Archetypes.
		// sanitizeEntityConfig(&config, fpath)

		// Set AssetDir for unique NPC sprites:
		// assets/images/npcs/<id>/
		config.AssetDir = path.Join("assets/images/npcs", config.ID)
		config.AudioDir = path.Join("assets/audio/npcs", config.ID)

		// Link weapon
		config.Weapon = GetWeaponByName(config.WeaponName)

		log.Printf("DEBUG: NPC Registry loading %s from %s", config.ID, fpath)
		r.NPCs[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		return nil
	})
}

type ObstacleArchetype struct {
	ID             string                 `yaml:"id"`
	Name           string                 `yaml:"name"`
	Type           ObstacleType           `yaml:"type"`
	Destructible   bool                   `yaml:"destructible"` // If false, cannot be damaged
	Description    string                 `yaml:"description"`
	Health         int                    `yaml:"health"`        // Base health (ignored if Destructible is false)
	CooldownTime   float64                `yaml:"cooldown_time"` // Base cooldown in minutes
	Footprint      []FootprintPoint       `yaml:"footprint"`
	FrameCount     int                    `yaml:"frame_count"`     // Total number of frames
	FramesPerRow   int                    `yaml:"frames_per_row"`  // For grid-based spritesheets (default 0 = single row)
	AnimationSpeed int                    `yaml:"animation_speed"` // Ticks per frame
	Actions        []ObstacleActionConfig `yaml:"actions,omitempty"`
	Image          interface{}            `yaml:"-"`
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
		return nil
	})
}
func (r *ObstacleRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics) {
	var jobs []*SpriteLoadJob
	for _, config := range r.Archetypes {
		// Derive image path from ID: assets/images/obstacles/<id>.png
		imagePath := path.Join("assets/images/obstacles", config.ID+".png")
		jobs = append(jobs, &SpriteLoadJob{
			Path: imagePath,
			Dest: &config.Image,
		})
	}
	loadSpritesParallel(assets, jobs, graphics)
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

func LoadPlayableCharacterConfig(assets fs.FS) (*EntityConfig, error) {
	if assets == nil {
		return &EntityConfig{}, nil
	}
	const configPath = "data/characters/oinakos.yaml"
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
		return nil, fmt.Errorf("failed to read playable character config: %w", err)
	}

	var config EntityConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal playable character config %s: %w", configPath, err)
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
	case "Pitchfork":
		return &Weapon{Name: "Pitchfork", MinDamage: 4, MaxDamage: 10}
	case "Gilded Pitchfork":
		return &Weapon{Name: "Gilded Pitchfork", MinDamage: 8, MaxDamage: 18}
	case "Shouts":
		return &Weapon{Name: "Shouts", MinDamage: 10, MaxDamage: 50}
	default:
		// Return a basic unarmed weapon instead of nil to prevent panics
		return WeaponFists
	}
}
