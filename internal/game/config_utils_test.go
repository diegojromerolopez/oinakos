package game

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestForEachYAML(t *testing.T) {
	fsys := fstest.MapFS{
		"data/map_types/test.yaml": &fstest.MapFile{Data: []byte("id: test")},
		"data/map_types/other.yml":  &fstest.MapFile{Data: []byte("id: other")},
		"data/map_types/ignore.txt": &fstest.MapFile{Data: []byte("ignore")},
	}
	
	count := 0
	err := forEachYAML(fsys, "data/map_types", func(fpath string, data []byte) error {
		count++
		return nil
	})
	
	if err != nil {
		t.Fatalf("forEachYAML failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 YAML files, got %d", count)
	}
}

func TestLoadPlayableCharacterConfig(t *testing.T) {
	fsys := fstest.MapFS{
		"data/characters/oinakos.yaml": &fstest.MapFile{Data: []byte("id: oinakos\nstats:\n  health_min: 100")},
	}
	
	t.Run("read from embedded", func(t *testing.T) {
		conf, err := LoadPlayableCharacterConfig(fsys)
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}
		if conf.ID != "oinakos" || conf.Stats.HealthMin != 100 {
			t.Errorf("incorrect config: %+v", conf)
		}
	})
	
	t.Run("read from local override", func(t *testing.T) {
		// Mock local override
		tmpDir, _ := os.MkdirTemp("", "oinakos-test")
		defer os.RemoveAll(tmpDir)
		
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)
		
		localDir := filepath.Join("oinakos", "data", "characters")
		os.MkdirAll(localDir, 0755)
		localPath := filepath.Join(localDir, "oinakos.yaml")
		os.WriteFile(localPath, []byte("id: oinakos_override\nstats:\n  health_min: 200"), 0644)
		
		conf, err := LoadPlayableCharacterConfig(nil) // nil fsys to force check local/etc.
		// Wait, LoadPlayableCharacterConfig(nil) returns empty conf if no data found.
		// But in our case it should find local override if it exists.
		
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}
		if conf.ID != "oinakos_override" {
			t.Errorf("override not loaded, got %s", conf.ID)
		}
	})
}
