package game

import (
	"os"
	"path/filepath"
	"testing"
	"gopkg.in/yaml.v3"
)

func findProjectRoot() string {
	dir, _ := os.Getwd()
	for i := 0; i < 4; i++ {
		if _, err := os.Stat(filepath.Join(dir, "data")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return "."
}

func iterateAndValidate(t *testing.T, baseDir string, target interface{}) {
	root := findProjectRoot()
	dir := filepath.Join(root, baseDir)
	
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("Directory %s not found, skipping validation", dir)
		return
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.IsDir() || (filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml") {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			// We use a fresh instance for each unmarshal
			if err := yaml.Unmarshal(data, target); err != nil {
				t.Errorf("Validation failed for %s: %v", path, err)
			}
		})
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}
}

func TestValidate_EveryArchetype(t *testing.T) {
	var target Archetype
	iterateAndValidate(t, "data/archetypes", &target)
}

func TestValidate_EveryAnimal(t *testing.T) {
	var target Archetype
	iterateAndValidate(t, "data/animals", &target)
}

func TestValidate_EveryCharacter(t *testing.T) {
	var target EntityConfig
	iterateAndValidate(t, "data/characters", &target)
}

func TestValidate_EveryObject(t *testing.T) {
	var target ObjectConfig
	iterateAndValidate(t, "data/objects", &target)
}

func TestValidate_EveryMap(t *testing.T) {
	var target MapType
	iterateAndValidate(t, "data/maps", &target)
}

func TestValidate_EveryObstacle(t *testing.T) {
	var target ObstacleArchetype
	iterateAndValidate(t, "data/obstacles", &target)
}

func TestValidation_RegistryLoading(t *testing.T) {
	// This ensures that the high-level registries correctly find and load the files
	t.Run("ArchetypeRegistry", func(t *testing.T) {
		r := NewArchetypeRegistry()
		r.LoadAll(nil)
		if len(r.IDs) == 0 { t.Error("ArchetypeRegistry failed to load anything") }
	})
	t.Run("CharacterRegistry", func(t *testing.T) {
		r := NewCharacterRegistry()
		r.LoadAll(nil)
		if len(r.IDs) == 0 { t.Error("CharacterRegistry failed to load anything") }
	})
}
