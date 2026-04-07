package game

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type CharacterRegistry struct { Characters map[string]*EntityConfig; IDs []string }
func NewCharacterRegistry() *CharacterRegistry { return &CharacterRegistry{Characters: make(map[string]*EntityConfig), IDs: make([]string, 0)} }
func (r *CharacterRegistry) PlayableIDs() []string {
	var ids []string
	for _, id := range r.IDs { if c, ok := r.Characters[id]; ok && c.Playable { ids = append(ids, id) } }; return ids
}

func (r *CharacterRegistry) LoadAll(assets fs.FS) error {
	return forEachYAML(assets, "data/characters", func(fpath string, data []byte) error {
		var cfg EntityConfig; if err := yaml.Unmarshal(data, &cfg); err != nil { return nil }
		if cfg.ID == "" { cfg.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath)) }
		if _, exists := r.Characters[cfg.ID]; exists { return nil }
		sanitizeEntityConfig(&cfg, fpath); cfg.AssetDir, cfg.AudioDir, cfg.SoundID = path.Join("assets/images/characters", cfg.ID), path.Join("assets/audio/characters", cfg.ID), cfg.ID
		if cfg.Playable { cfg.PlayableCharacter = cfg.ID }; r.Characters[cfg.ID], r.IDs = &cfg, append(r.IDs, cfg.ID); return nil
	})
}

func (r *CharacterRegistry) LoadAssets(assets fs.FS, graphics engine.Graphics, archs *ArchetypeRegistry, permit map[string]bool, ls *LoadingState) {
	if jobs := r.createLoadJobs(assets, archs, permit); len(jobs) > 0 { loadSpritesParallel(assets, jobs, graphics, ls) }
}

func (r *CharacterRegistry) createLoadJobs(assets fs.FS, archs *ArchetypeRegistry, permit map[string]bool) []*SpriteLoadJob {
	if assets == nil { return nil }
	var jobs []*SpriteLoadJob
	for _, c := range r.Characters {
		if c.StaticImage != nil || (permit != nil && !permit[c.ID] && !c.Playable) { continue }
		lookup, hasAud := c.Archetype, false
		if c.Gender != "" && !strings.Contains(c.Archetype, c.Gender) { if _, ok := archs.Archetypes[c.Archetype+"_"+c.Gender]; ok { lookup = c.Archetype + "_" + c.Gender } }
		arch, _ := archs.Archetypes[lookup]
		if assets != nil && c.AudioDir != "" { if entries, err := fs.ReadDir(assets, c.AudioDir); err == nil && len(entries) > 0 { hasAud = true } }
		if hasAud { c.SoundID = c.ID } else if arch != nil { c.SoundID = lookup } else { c.SoundID = c.ID }
		if c.AssetDir != "" {
			add := func(f string, t *engine.Image, fb engine.Image) {
				if _, err := fs.Stat(assets, path.Join(c.AssetDir, f)); err == nil { jobs = append(jobs, &SpriteLoadJob{Path: path.Join(c.AssetDir, f), Dest: t}) } else if fb != nil { *t = fb }
			}
			fArch, bArch, crArch, cArch, pArch, ckArch, rArch := engine.Image(nil), engine.Image(nil), engine.Image(nil), engine.Image(nil), engine.Image(nil), engine.Image(nil), engine.Image(nil)
			if arch != nil { fArch, bArch, crArch, cArch, pArch, ckArch, rArch = arch.StaticImage, arch.BackImage, arch.CorpseImage, arch.CrouchImage, arch.PregnantImage, arch.CookingImage, arch.RestingImage }
			add("static.png", &c.StaticImage, fArch); add("back.png", &c.BackImage, bArch); add("corpse.png", &c.CorpseImage, crArch); add("crouch.png", &c.CrouchImage, cArch); add("pregnant.png", &c.PregnantImage, pArch); add("cooking.png", &c.CookingImage, ckArch); add("resting.png", &c.RestingImage, rArch)
			
			// Combat and interaction images with archetypal fallbacks
			cFiles := []string{"attack.png", "attack1.png", "attack2.png", "hit.png", "hit1.png", "hit2.png", "chopping.png", "digging.png"}
			cDests := []*engine.Image{&c.AttackImage, &c.Attack1Image, &c.Attack2Image, &c.HitImage, &c.Hit1Image, &c.Hit2Image, &c.ChoppingImage, &c.DiggingImage}
			var cArchs []engine.Image
			if arch != nil {
				cArchs = []engine.Image{arch.AttackImage, arch.Attack1Image, arch.Attack2Image, arch.HitImage, arch.Hit1Image, arch.Hit2Image, arch.ChoppingImage, arch.DiggingImage}
			} else {
				cArchs = make([]engine.Image, len(cFiles))
			}
			for i, f := range cFiles {
				add(f, cDests[i], cArchs[i])
			}
		}
	}
	return jobs
}

func (r *CharacterRegistry) ProcessInheritance(archs *ArchetypeRegistry) {
	for _, c := range r.Characters {
		lookup := c.Archetype; if c.Gender != "" && !strings.Contains(c.Archetype, c.Gender) { if _, ok := archs.Archetypes[c.Archetype+"_"+c.Gender]; ok { lookup = c.Archetype + "_" + c.Gender } }
		if arch, ok := archs.Archetypes[lookup]; ok {
			if c.Stats.HealthPoints.IsZero() {
		c.Stats.HealthPoints = arch.Stats.HealthPoints
	}
			if c.Stats.Speed.IsZero() { c.Stats.Speed = arch.Stats.Speed }
			if c.Stats.BaseAttack.IsZero() { c.Stats.BaseAttack = arch.Stats.BaseAttack }
			if c.Stats.BaseDefense.IsZero() { c.Stats.BaseDefense = arch.Stats.BaseDefense }
			if c.Attributes.Strength.IsZero() { c.Attributes.Strength = arch.Attributes.Strength }
			if c.Attributes.Dexterity.IsZero() { c.Attributes.Dexterity = arch.Attributes.Dexterity }
			if c.Attributes.Health.IsZero() { c.Attributes.Health = arch.Attributes.Health }; if c.PrimaryColor == "" { c.PrimaryColor = arch.PrimaryColor }
			if c.SecondaryColor == "" { c.SecondaryColor = arch.SecondaryColor }; if len(c.Footprint) == 0 { c.Footprint = arch.Footprint }
		}; sanitizeEntityConfig(c, c.ID)
	}
}
