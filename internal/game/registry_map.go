package game

import (
	"fmt"
	"io/fs"
	"log"
	"math"
	"math/rand"
	"path/filepath"
	"strings"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type PreSpawnObject struct {
	ID string  `yaml:"id"`
	X  float64 `yaml:"x"`
	Y  float64 `yaml:"y"`
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
	Characters      []Inhabitant       `yaml:"characters,omitempty"`
	Spawns          []SpawnConfig      `yaml:"spawns"`
	Obstacles       []PreSpawnObstacle `yaml:"obstacles"`
	Objects         []PreSpawnObject   `yaml:"objects"`
	FloorTile       string             `yaml:"floor_tile"`
	FloorZones      []*FloorZone       `yaml:"floor_zones"`
	TargetPointRaw  *TargetPointConfig `yaml:"target_point"` // Optional YAML-supplied target point
	Player          *TargetPointConfig `yaml:"player,omitempty"`
	Weather         string             `yaml:"weather"`
	MapWidth        float64            `yaml:"-"` // Cartesian width
	MapHeight       float64            `yaml:"-"` // Cartesian height

	HeightZones     []*HeightZone      `yaml:"height_zones,omitempty"`

	TargetNPC      *EntityConfig `yaml:"-"`
	TargetObstacle *Obstacle     `yaml:"-"`
	TargetPoint    engine.Point  `yaml:"-"` // Resolved at loadMapLevel time
	StartTime      float64       `yaml:"-"`
	IsCompleted    bool          `yaml:"-"`
	Heightmap      map[string]float64 `yaml:"-"` // "x,y" -> z (Cartesian grid height)
	MineralMap     map[string]string `yaml:"-"` // "x,y" -> mineral type (gold, silver, etc)
}

func (m *MapType) SeedMinerals(seed int64) {
	m.MineralMap = make(map[string]string)
	r := rand.New(rand.NewSource(seed))
	
	// Define mineral types and their rarities
	minerals := []struct {
		id       string
		rarity   float64
		maxRadius float64
	}{
		{"gold_ore", 0.02, 3.0},
		{"silver_ore", 0.05, 4.0},
		{"iron_ore", 0.15, 6.0},
		{"copper_ore", 0.25, 8.0},
	}

	numVeins := 50 // Total possible vein centers per map
	for i := 0; i < numVeins; i++ {
		// Random center
		cx := r.Float64() * m.MapWidth - m.MapWidth/2
		cy := r.Float64() * m.MapHeight - m.MapHeight/2
		
		// Choose mineral type
		prob := r.Float64()
		var mType string
		var radius float64
		cumProb := 0.0
		for _, mDef := range minerals {
			cumProb += mDef.rarity
			if prob < cumProb {
				mType = mDef.id
				radius = 1.0 + r.Float64()*mDef.maxRadius
				break
			}
		}

		if mType == "" { continue }

		// Fill a circle around the center with this mineral
		for dx := -int(radius); dx <= int(radius); dx++ {
			for dy := -int(radius); dy <= int(radius); dy++ {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= radius {
					// Some randomness in density
					if r.Float64() < (1.0 - dist/radius) {
						tx, ty := int(math.Floor(cx))+dx, int(math.Floor(cy))+dy
						key := fmt.Sprintf("%d,%d", tx, ty)
						m.MineralMap[key] = mType
					}
				}
			}
		}
	}
}

// HeightZone defines an area with a specific elevation or a gradient
type HeightZone struct {
	Polygon       engine.Polygon `yaml:"polygon"`
	BaseZ         float64        `yaml:"base_z"`
	GradientPoint engine.Point   `yaml:"gradient_point,omitempty"`
	GradientZ     float64        `yaml:"gradient_z,omitempty"`
	Priority      int            `yaml:"priority,omitempty"`
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

func (m *MapType) GetElevationAt(x, y float64) float64 {
	// Discrete corner positions
	fx := math.Floor(x)
	fy := math.Floor(y)

	// Local offsets [0.0, 1.0]
	wx := x - fx
	wy := y - fy

	getV := func(gx, gy float64) float64 {
		gridX, gridY := int(gx), int(gy)
		key := fmt.Sprintf("%d,%d", gridX, gridY)
		if m.Heightmap != nil {
			if val, exists := m.Heightmap[key]; exists {
				return val
			}
		}
		// Fallback to zones for vertices
		z := 0.0
		hp := -1
		for _, zone := range m.HeightZones {
			if zone.Priority > hp && zone.Polygon.Contains(gx, gy) {
				z = zone.BaseZ
				hp = zone.Priority
			}
		}
		return z
	}

	// Bilinear interpolation between the 4 corners
	z00 := getV(fx, fy)
	z10 := getV(fx+1, fy)
	z01 := getV(fx, fy+1)
	z11 := getV(fx+1, fy+1)

	zLeft := z00*(1-wx) + z10*wx
	zRight := z01*(1-wx) + z11*wx

	return zLeft*(1-wy) + zRight*wy
}

// Dig modifies the heightmap at the given coordinates, lowering the elevation.
func (m *MapType) Dig(x, y float64, amount float64) {
	if m.Heightmap == nil {
		m.Heightmap = make(map[string]float64)
	}

	// For digging, we affect the vertex nearest to the target OR the 4 surrounding vertices
	// Let's affect the 4 grid corners around (x,y) with a gaussian-like falloff
	affectedRange := 1
	for dx := -affectedRange; dx <= affectedRange; dx++ {
		for dy := -affectedRange; dy <= affectedRange; dy++ {
			gridX, gridY := int(math.Floor(x))+dx, int(math.Floor(y))+dy
			key := fmt.Sprintf("%d,%d", gridX, gridY)
			
			// Calculate distance factor
			dist := math.Sqrt(math.Pow(x-float64(gridX), 2) + math.Pow(y-float64(gridY), 2))
			if dist > float64(affectedRange) {
				continue
			}
			weight := (float64(affectedRange) - dist) / float64(affectedRange)
			
			currentZ := m.GetElevationAt(float64(gridX), float64(gridY))
			m.Heightmap[key] = currentZ - (amount * weight)
		}
	}
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
			log.Printf("DEBUG: forEachYAML processing %s", normalizedPath)
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
