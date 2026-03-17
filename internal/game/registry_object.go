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

func (r *ObjectRegistry) LoadAll(assets fs.FS) error {
	if assets == nil {
		return nil
	}
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
		
		r.Objects[config.ID] = &config
		r.IDs = append(r.IDs, config.ID)
		return nil
	})
}

func (r *ObjectRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, progress *int32) {
	var jobs []*SpriteLoadJob
	for _, config := range r.Objects {
		// Each object has its own image file named after its ID
		filename := config.ID + ".png"
		jobs = append(jobs, &SpriteLoadJob{
			Path: path.Join("assets/images/objects", filename),
			Dest: &config.Sprite,
		})
	}
	loadSpritesParallel(assets, jobs, graphics, progress)
}
