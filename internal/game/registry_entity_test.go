package game

import (
	"oinakos/internal/engine"
	"testing"
	"testing/fstest"
)

func TestArchetypeRegistry_CreateLoadJobs(t *testing.T) {
	reg := NewArchetypeRegistry()
	reg.Archetypes["orc"] = &Archetype{
		ID:       "orc",
		AssetDir: "assets/images/archetypes/orc",
	}
	
	t.Run("no permit list", func(t *testing.T) {
		jobs := reg.createLoadJobs(nil)
		// 15 assets per archetype (static, back, corpse, attack, attack1, attack2, hit, hit1, hit2, crouch, chopping, digging, pregnant, cooking, resting)
		if len(jobs) != 15 {
			t.Errorf("expected 15 jobs, got %d", len(jobs))
		}
	})
	
	t.Run("permit list excludes", func(t *testing.T) {
		permit := map[string]bool{"human": true}
		jobs := reg.createLoadJobs(permit)
		if len(jobs) != 0 {
			t.Errorf("expected 0 jobs, got %d", len(jobs))
		}
	})
}

func TestCharacterRegistry_CreateLoadJobs(t *testing.T) {
	fsys := fstest.MapFS{
		"assets/audio/characters/hero/attack_1.wav": &fstest.MapFile{Data: []byte("wav")},
		"assets/images/characters/hero/static.png":  &fstest.MapFile{Data: []byte("png")},
		"assets/images/characters/hero/attack.png":  &fstest.MapFile{Data: []byte("png")},
		"assets/images/characters/hero/hit.png":     &fstest.MapFile{Data: []byte("png")},
	}
	
	archReg := NewArchetypeRegistry()
	archReg.Archetypes["man"] = &Archetype{
		ID:       "man",
		AssetDir: "assets/images/archetypes/man",
		Stats: EntityStatsConfig{HealthPoints: IntInterval{Min: 50, Max: 50}},
	}
	
	charReg := NewCharacterRegistry()
	charReg.Characters["hero"] = &EntityConfig{
		ID:          "hero",
		Archetype: "man",
		AssetDir:    "assets/images/characters/hero",
		AudioDir:    "assets/audio/characters/hero",
		Playable:    true,
	}
	
	charReg.ProcessInheritance(archReg)
	jobs := charReg.createLoadJobs(fsys, archReg, nil)
	
	// Hero has at least 1 local image (static.png) and several other jobs including attack.png
	foundAttack := false
	for _, job := range jobs {
		if job.Path == "assets/images/characters/hero/attack.png" {
			foundAttack = true; break
		}
	}
	if !foundAttack {
		t.Error("expected attack.png job for hero character")
	}
	if len(jobs) < 2 {
		t.Errorf("expected multiple load jobs, got %d", len(jobs))
	}
	
	hero := charReg.Characters["hero"]
	if hero.SoundID != "hero" {
		t.Errorf("expected SoundID hero, got %s", hero.SoundID)
	}
	// Verify stat fallback
	if hero.Stats.HealthPoints.Min != 50 {
		t.Errorf("expected HealthPoints 50 from archetype, got %v", hero.Stats.HealthPoints)
	}
}

func TestPickAttackImage(t *testing.T) {
	config := &EntityConfig{}
	
	img1 := &engine.MockImage{}
	img2 := &engine.MockImage{}
	imgDefault := &engine.MockImage{}

	config.AttackImage = imgDefault
	if got := config.PickAttackImage(0); got != imgDefault {
		t.Errorf("expected default attack image, got %v", got)
	}

	config.Attack1Image = img1
	if got := config.PickAttackImage(0); got != img1 {
		t.Errorf("expected attack1 image, got %v", got)
	}

	config.Attack2Image = img2
	if got := config.PickAttackImage(0); got != img1 {
		t.Errorf("expected attack1 image for seed 0, got %v", got)
	}
	if got := config.PickAttackImage(1); got != img2 {
		t.Errorf("expected attack2 image for seed 1, got %v", got)
	}
}

func TestPickHitImage(t *testing.T) {
	config := &EntityConfig{}
	
	img1 := &engine.MockImage{}
	img2 := &engine.MockImage{}
	imgDefault := &engine.MockImage{}

	config.HitImage = imgDefault
	if got := config.PickHitImage(0); got != imgDefault {
		t.Errorf("expected default hit image, got %v", got)
	}

	config.Hit1Image = img1
	if got := config.PickHitImage(0); got != img1 {
		t.Errorf("expected hit1 image, got %v", got)
	}

	config.Hit2Image = img2
	if got := config.PickHitImage(0); got != img1 {
		t.Errorf("expected hit1 image for seed 0, got %v", got)
	}
	if got := config.PickHitImage(1); got != img2 {
		t.Errorf("expected hit2 image for seed 1, got %v", got)
	}
}

func TestGetFootprint(t *testing.T) {
	config := &EntityConfig{}
	
	// Default footprint
	fp := config.GetFootprint()
	if len(fp.Points) != 4 {
		t.Errorf("expected 4 points in default footprint, got %d", len(fp.Points))
	}
	
	// Custom footprint
	config = &EntityConfig{
		Footprint: []FootprintPoint{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 0, Y: 1},
		},
	}
	fp = config.GetFootprint()
	if len(fp.Points) != 3 {
		t.Errorf("expected 3 points in custom footprint, got %d", len(fp.Points))
	}
	
	// Caching check
	fp2 := config.GetFootprint()
	if config.CachedBaseFootprint == nil || len(config.CachedBaseFootprint.Points) != 3 {
		t.Error("footprint should be cached")
	}
	if &fp2.Points[0] == &fp.Points[0] {
		// This might not be true if GetFootprint returns a copy, but CachedBaseFootprint should be set.
	}
}

func TestCharacterRegistry_PlayableIDs(t *testing.T) {
	reg := NewCharacterRegistry()
	reg.Characters["c1"] = &EntityConfig{ID: "c1", Playable: true}
	reg.Characters["c2"] = &EntityConfig{ID: "c2", Playable: false}
	reg.Characters["c3"] = &EntityConfig{ID: "c3", Playable: true}
	reg.IDs = []string{"c1", "c2", "c3"}
	
	playables := reg.PlayableIDs()
	if len(playables) != 2 {
		t.Errorf("expected 2 playable IDs, got %d", len(playables))
	}
	if playables[0] != "c1" || playables[1] != "c3" {
		t.Errorf("incorrect playable IDs: %v", playables)
	}
}
