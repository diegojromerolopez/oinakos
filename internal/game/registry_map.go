package game

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

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

func (m *MapType) GetTileAt(x, y float64) string {
	resolvedTile := m.FloorTile
	highestPriority := -1
	for _, zone := range m.FloorZones {
		if zone.Priority > highestPriority {
			if zone.Contains(x, y) {
				resolvedTile = zone.Tile
				highestPriority = zone.Priority
			}
		}
	}
	return resolvedTile
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
