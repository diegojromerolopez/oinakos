package game

import (
	"testing"
	"testing/fstest"
)

func TestObstacleRegistry_LoadAll(t *testing.T) {
	fs := fstest.MapFS{
		"data/obstacles/tree_oak.yaml": {
			Data: []byte(`
id: tree_oak
name: Oak Tree
type: tree
destructible: true
health: 500
passable: false
timber: 100
`),
		},
		"data/obstacles/well.yaml": {
			Data: []byte(`
id: well
name: Town Well
type: well
cooldown_time: 15.0
passable: false
`),
		},
		"data/obstacles/crop_wheat.yaml": {
			Data: []byte(`
id: crop_wheat
name: Wheat Field
plant_season: SPRING
growth_duration: 1000
`),
		},
	}

	r := NewObstacleRegistry()
	err := r.LoadAll(fs)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(r.Archetypes) < 3 {
		t.Errorf("expected 3 archetypes, got %d", len(r.Archetypes))
	}

	tree := r.Archetypes["tree_oak"]
	if tree == nil || tree.HealthPoints != 500 {
		t.Errorf("failed to load oak tree correctly: %v", tree)
	}

	well := r.Archetypes["well"]
	if well == nil || !well.IsWell() {
		t.Error("well was not correctly identified as a well")
	}

	crop := r.Archetypes["crop_wheat"]
	if crop == nil || !crop.IsCrop {
		t.Error("crop_wheat was not correctly identified as a crop (ID prefix rule failed)")
	}
}

func TestObstacleRegistry_CreateLoadJobs(t *testing.T) {
	r := NewObstacleRegistry()
	r.Archetypes["tree_oak"] = &ObstacleArchetype{ID: "tree_oak"}
	r.Archetypes["well"] = &ObstacleArchetype{ID: "well", CooldownTime: 1.0} // IsWell=true
	
	// Test without permit list
	jobs := r.createLoadJobs(fstest.MapFS{}, nil)
	if len(jobs) != 2 {
		t.Errorf("expected 2 load jobs, got %d", len(jobs))
	}
	
	// wells are often loaded even if not in permit list? 
	// (Logic: config.IsWell() bypasses permit check)
	permit := map[string]bool{"tree_oak": false} // Only well should be allowed if well bypasses
	
	jobs = r.createLoadJobs(fstest.MapFS{}, permit)
	// Check if well bypassed the filter
	foundWell := false
	for _, j := range jobs {
		if j.Path == "assets/images/obstacles/well.png" {
			foundWell = true
		}
	}
	if !foundWell {
		t.Error("Expected well to bypass permit list filter as it provides critical functionality")
	}
}
