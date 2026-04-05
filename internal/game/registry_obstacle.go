package game

import (
	"log"
	"path"
	"path/filepath"
	"strings"
	"io/fs"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type ObstacleArchetype struct {
	ID             string                 `yaml:"id"`
	Name           string                 `yaml:"name"`
	Type           ObstacleType           `yaml:"type"`
	Destructible   bool                   `yaml:"destructible"` // If false, cannot be damaged
	Description    string                 `yaml:"description"`
	HealthPoints   int                    `yaml:"health"`        // Base health (ignored if Destructible is false)
	Timber         int                    `yaml:"timber"`        // Available timber resources for harvesting
	Weight         float64                `yaml:"weight"`        // Total resource weight (game units)
	CooldownTime   float64                `yaml:"cooldown_time"` // Base cooldown in minutes
	Footprint      []FootprintPoint       `yaml:"footprint"`
	FrameCount     int                    `yaml:"frame_count"`     // Total number of frames
	FramesPerRow   int                    `yaml:"frames_per_row"`  // For grid-based spritesheets (default 0 = single row)
	AnimationSpeed int                    `yaml:"animation_speed"` // Ticks per frame
	Actions        []ObstacleActionConfig `yaml:"actions,omitempty"`

	IsCrop         bool   `yaml:"is_crop"`
	PlantSeason    string `yaml:"plant_season"`   // e.g. "SPRING"
	HarvestSeason  string `yaml:"harvest_season"` // e.g. "AUTUMN"
	GrowthDuration int    `yaml:"growth_duration"`
	Yield          string `yaml:"yield"`           // Object ID to drop when harvested

	// Container & Ownership System
	MaxCapacity    float64 `yaml:"max_capacity"` // Max weight it can hold
	OwnerID        string  `yaml:"owner_id"`     // Optional default owner
	LockResistance int     `yaml:"lock_resistance"`

	// Environmental Hazard System
	Passable       bool   `yaml:"passable"`  // If true, characters can walk over it
	IsHazard       bool   `yaml:"is_hazard"` // If true, it can affect actor states (hygiene, health)

	Image       engine.Image `yaml:"-"`
	OpenImage   engine.Image `yaml:"-"`
	ClosedImage engine.Image `yaml:"-"`
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

		if config.ID == "" {
			config.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
		}

		sanitizeObstacleArchetype(&config, fpath)
		if _, exists := r.Archetypes[config.ID]; exists { return nil }
		r.Archetypes[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		return nil
	})
}

func (r *ObstacleRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, permitList map[string]bool, ls *LoadingState) {
	jobs := r.createLoadJobs(permitList)
	if len(jobs) > 0 {
		loadSpritesParallel(assets, jobs, graphics, ls)
	}
}

func (r *ObstacleRegistry) createLoadJobs(permitList map[string]bool) []*SpriteLoadJob {
	var jobs []*SpriteLoadJob
	for _, config := range r.Archetypes {
		if config.Image != nil {
			continue
		}
		if permitList != nil && !permitList[config.ID] && !config.IsWell() {
			continue
		}
		if config.ID == "chest" {
			jobs = append(jobs, &SpriteLoadJob{
				Path: path.Join("assets/images/obstacles/chest", "open.png"),
				Dest: &config.OpenImage,
			})
			jobs = append(jobs, &SpriteLoadJob{
				Path: path.Join("assets/images/obstacles/chest", "closed.png"),
				Dest: &config.ClosedImage,
			})
			// Still load the default Image as fallback or for shared logic
			jobs = append(jobs, &SpriteLoadJob{
				Path: path.Join("assets/images/obstacles/chest", "closed.png"),
				Dest: &config.Image,
			})
			continue
		}
		imagePath := path.Join("assets/images/obstacles", config.ID+".png")
		jobs = append(jobs, &SpriteLoadJob{
			Path: imagePath,
			Dest: &config.Image,
		})
	}
	return jobs
}

func (r *ObstacleRegistry) CountAssets(permitList map[string]bool) int {
	return len(r.createLoadJobs(permitList))
}
