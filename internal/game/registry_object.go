package game

import (
	"io/fs"
	"log"
	"math/rand"
	"path"
	"path/filepath"
	"strings"

	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type ObjectRegistry struct {
	Objects map[string]*ObjectConfig
	IDs     []string
}

func NewObjectRegistry() *ObjectRegistry {
	return &ObjectRegistry{
		Objects: make(map[string]*ObjectConfig),
		IDs:     make([]string, 0),
	}
}

func (r *ObjectRegistry) Get(id string) *ObjectConfig {
	return r.Objects[id]
}

func (r *ObjectRegistry) LoadAll(assets fs.FS) error {
	const baseDir = "data/objects"
	return forEachYAML(assets, baseDir, func(fpath string, data []byte) error {
		var config ObjectConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", fpath, err)
			return nil
		}

		if config.ID == "" {
			config.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
		}

		config.AssetDir = path.Join("assets/images/objects")
		if _, exists := r.Objects[config.ID]; exists { return nil }
		r.Objects[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		return nil
	})
}

func (r *ObjectRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, permitList map[string]bool, ls *LoadingState) {
	jobs := r.createLoadJobs(permitList)
	if len(jobs) > 0 {
		loadSpritesParallel(assets, jobs, graphics, ls)
	}
}

func (r *ObjectRegistry) createLoadJobs(permitList map[string]bool) []*SpriteLoadJob {
	var jobs []*SpriteLoadJob
	for _, config := range r.Objects {
		if config.Sprite != nil {
			continue
		}
		if permitList != nil && !permitList[config.ID] {
			continue
		}
		// Each object has its own image file named after its ID
		filename := config.ID + ".png"
		jobs = append(jobs, &SpriteLoadJob{
			Path: path.Join("assets/images/objects", filename),
			Dest: &config.Sprite,
		})
	}
	return jobs
}

func (r *ObjectRegistry) CountAssets(permitList map[string]bool) int {
	return len(r.createLoadJobs(permitList))
}

// RandomVariantID returns a random variant config for a base ID (e.g. "raw_meat").
// It looks for IDs like "raw_meat_1", "raw_meat_2", etc. in the registry.
// If no variants are found, it falls back to the base ID itself.
func (r *ObjectRegistry) RandomVariantID(baseID string) (string, *ObjectConfig) {
	var variants []*ObjectConfig
	prefix := baseID + "_"
	for _, cfg := range r.Objects {
		if strings.HasPrefix(cfg.ID, prefix) {
			// Ensure the suffix is purely numeric
			suffix := cfg.ID[len(prefix):]
			allDigits := len(suffix) > 0
			for _, ch := range suffix {
				if ch < '0' || ch > '9' { allDigits = false; break }
			}
			if allDigits {
				variants = append(variants, cfg)
			}
		}
	}
	if len(variants) == 0 {
		return baseID, r.Objects[baseID]
	}
	chosen := variants[rand.Intn(len(variants))]
	return chosen.ID, chosen
}
