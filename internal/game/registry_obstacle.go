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

func (r *ObstacleRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, progress *int32) {
	var jobs []*SpriteLoadJob
	for _, config := range r.Archetypes {
		// Derive image path from ID: assets/images/obstacles/<id>.png
		imagePath := path.Join("assets/images/obstacles", config.ID+".png")
		jobs = append(jobs, &SpriteLoadJob{
			Path: imagePath,
			Dest: &config.Image,
		})
	}
	loadSpritesParallel(assets, jobs, graphics, progress)
}
