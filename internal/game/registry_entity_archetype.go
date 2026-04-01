package game

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type Archetype = EntityConfig
type ArchetypeRegistry struct { Archetypes map[string]*Archetype; IDs []string }

func NewArchetypeRegistry() *ArchetypeRegistry { return &ArchetypeRegistry{Archetypes: make(map[string]*Archetype), IDs: make([]string, 0)} }

func (r *ArchetypeRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, permitList map[string]bool, ls *LoadingState) {
	for _, config := range r.Archetypes {
		if config.AssetDir == "" { continue }
		if entries, err := fs.ReadDir(assets, path.Join(config.AssetDir, "models")); err == nil {
			config.Models = make(map[string]*ModelConfig)
			for _, entry := range entries { if entry.IsDir() { config.Models[entry.Name()] = &ModelConfig{ID: entry.Name()} } }
		}
	}
	if jobs := r.createLoadJobs(permitList); len(jobs) > 0 { loadSpritesParallel(assets, jobs, graphics, ls) }
}

func (r *ArchetypeRegistry) createLoadJobs(permitList map[string]bool) []*SpriteLoadJob {
	var jobs []*SpriteLoadJob
	for _, c := range r.Archetypes {
		if (permitList != nil && !permitList[c.ID]) || c.AssetDir == "" { continue }
		add := func(p string, t *engine.Image) { jobs = append(jobs, &SpriteLoadJob{Path: p, Dest: t}) }
		files := []string{"static.png", "back.png", "corpse.png", "attack.png", "attack1.png", "attack2.png", "hit.png", "hit1.png", "hit2.png", "crouch.png", "chopping.png", "digging.png", "pregnant.png", "cooking.png"}
		dests := []*engine.Image{&c.StaticImage, &c.BackImage, &c.CorpseImage, &c.AttackImage, &c.Attack1Image, &c.Attack2Image, &c.HitImage, &c.Hit1Image, &c.Hit2Image, &c.CrouchImage, &c.ChoppingImage, &c.DiggingImage, &c.PregnantImage, &c.CookingImage}
		for i, f := range files { add(path.Join(c.AssetDir, f), dests[i]) }
		for mID, mod := range c.Models {
			mDir := path.Join(c.AssetDir, "models", mID)
			mFiles := []string{"static.png", "back.png", "corpse.png", "attack.png", "hit.png", "crouch.png", "pregnant.png", "cooking.png"}
			mDests := []*engine.Image{&mod.StaticImage, &mod.BackImage, &mod.CorpseImage, &mod.AttackImage, &mod.HitImage, &mod.CrouchImage, &mod.PregnantImage, &mod.CookingImage}
			for i, f := range mFiles { add(path.Join(mDir, f), mDests[i]) }
		}
	}
	return jobs
}

func (r *ArchetypeRegistry) LoadAll(assets fs.FS) error {
	baseDirs := []string{"data/archetypes", "data/animals"}
	for _, baseDir := range baseDirs {
		_ = forEachYAML(assets, baseDir, func(fpath string, data []byte) error {
			relP, _ := filepath.Rel(baseDir, fpath); subDir, varN := filepath.Dir(relP), strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
			if subDir == "." { subDir = "" }
			var cfg Archetype; if err := yaml.Unmarshal(data, &cfg); err != nil { return nil }
			if cfg.ID == "" { cfg.ID = varN }; sanitizeEntityConfig(&cfg, fpath)
			cat := "archetypes"; if baseDir == "data/animals" { cat, cfg.IsAnimal = "animals", true }
			cfg.AssetDir, cfg.AudioDir, cfg.SoundID = path.Join("assets/images", cat, subDir, varN), path.Join("assets/audio", cat, subDir, varN), cfg.ID
			r.Archetypes[cfg.ID], r.IDs = &cfg, append(r.IDs, cfg.ID); return nil
		})
	}
	return nil
}
func (r *ArchetypeRegistry) CountAssets(p map[string]bool) int { return len(r.createLoadJobs(p)) }
