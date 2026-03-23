package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestObjectRegistry_CreateLoadJobs(t *testing.T) {
	reg := NewObjectRegistry()
	reg.Objects["sword"] = &ObjectConfig{ID: "sword"}
	reg.Objects["shield"] = &ObjectConfig{ID: "shield"}
	reg.IDs = []string{"sword", "shield"}
	
	t.Run("no permit list", func(t *testing.T) {
		jobs := reg.createLoadJobs(nil)
		if len(jobs) != 2 {
			t.Errorf("expected 2 jobs, got %d", len(jobs))
		}
	})
	
	t.Run("with permit list", func(t *testing.T) {
		permit := map[string]bool{"sword": true}
		jobs := reg.createLoadJobs(permit)
		if len(jobs) != 1 {
			t.Fatalf("expected 1 job, got %d", len(jobs))
		}
		if jobs[0].Path != "assets/images/objects/sword.png" {
			t.Errorf("incorrect job path: %s", jobs[0].Path)
		}
	})
	
	t.Run("skip already loaded", func(t *testing.T) {
		reg.Objects["sword"].Sprite = &engine.MockImage{}
		jobs := reg.createLoadJobs(nil)
		if len(jobs) != 1 {
			t.Errorf("expected 1 job (shield), got %d", len(jobs))
		}
	})
}

func TestObjectRegistry_CountAssets(t *testing.T) {
	reg := NewObjectRegistry()
	reg.Objects["sword"] = &ObjectConfig{ID: "sword"}
	reg.IDs = []string{"sword"}
	
	count := reg.CountAssets(nil)
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}
