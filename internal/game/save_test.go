package game

import (
	"oinakos/internal/engine"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	g := NewGame(nil, &engine.MockGraphics{}, "data/maps/test_save.yaml", "", "", &MockInputManager{}, &MockAudioManager{}, false, "0.1-test")
	// Add NPC and Obstacle to test persistence
	n := NewCharacter(10, 20, &EntityConfig{ID: "orc", XP: 10, Stats: EntityStats{HealthMin: 100, HealthMax: 100}}, 1, false, nil)
	g.World.Characters = []*Character{n}
	g.characters = []*Character{n}
	
	obs := NewObstacle("test_building", 30, 40, &ObstacleArchetype{ID: "test_arch", Health: 100})
	g.World.Obstacles = []*Obstacle{obs}
	g.obstacles = []*Obstacle{obs}

	testPath := "test_save.yaml"
	defer os.Remove(testPath)

	if err := g.Save(testPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Create a new game and load
	g2 := NewGame(nil, &engine.MockGraphics{}, "", "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	// Mock registries for loading to work
	g2.characterRegistry.Characters["orc"] = &EntityConfig{ID: "orc"}
	g2.obstacleRegistry.Archetypes["test_arch"] = &ObstacleArchetype{ID: "test_arch"}

	if err := g2.Load(testPath); err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if g2.playableCharacter.X != g.playableCharacter.X || g2.playableCharacter.Y != g.playableCharacter.Y {
		t.Errorf("Position mismatch: expected (%f,%f), got (%f,%f)", g.playableCharacter.X, g.playableCharacter.Y, g2.playableCharacter.X, g2.playableCharacter.Y)
	}
	if g2.playableCharacter.Kills != g.playableCharacter.Kills {
		t.Errorf("Kills mismatch: expected %d, got %d", g.playableCharacter.Kills, g2.playableCharacter.Kills)
	}
	if g2.playableCharacter.XP != g.playableCharacter.XP {
		t.Errorf("XP mismatch: expected %d, got %d", g.playableCharacter.XP, g2.playableCharacter.XP)
	}
	if g2.playableCharacter.Health != g.playableCharacter.Health {
		t.Errorf("Health mismatch: expected %d, got %d", g.playableCharacter.Health, g2.playableCharacter.Health)
	}
	if g2.playTime != g.playTime {
		t.Errorf("PlayTime mismatch: expected %f, got %f", g.playTime, g2.playTime)
	}

	if len(g2.characters) != 1 {
		t.Errorf("NPCs mismatch: expected 1, got %d", len(g2.characters))
	} else if g2.characters[0].X != 10 || g2.characters[0].Y != 20 {
		t.Errorf("NPC pos mismatch: expected (10,20), got (%f,%f)", g2.characters[0].X, g2.characters[0].Y)
	}

	if len(g2.obstacles) != 1 {
		t.Errorf("Obstacles mismatch: expected 1, got %d", len(g2.obstacles))
	} else if g2.obstacles[0].ID != "test_building" {
		t.Errorf("Building ID mismatch: expected 'test_building', got '%s'", g2.obstacles[0].ID)
	}
}

func TestQuickSave(t *testing.T) {
	// Create a temporary directory for the test
	dir, err := os.MkdirTemp("", "test_quicksave")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir) // Clean up the directory

	// Change to the temporary directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Ensure the oinakos/data/maps directory exists for the map file
	mapDir := filepath.Join("oinakos", "data", "maps")
	if err := os.MkdirAll(mapDir, 0755); err != nil {
		t.Fatalf("Failed to create map directory: %v", err)
	}

	// Create a dummy map file
	dummyMapPath := filepath.Join(mapDir, "test_quicksave_map.yaml")
	if err := os.WriteFile(dummyMapPath, []byte("map_data: {}"), 0644); err != nil {
		t.Fatalf("Failed to create dummy map file: %v", err)
	}

	SetOinakosDir(dir)
	defer SetOinakosDir("") // Reset after test

	g := NewGame(nil, &engine.MockGraphics{}, dummyMapPath, "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	g.performQuicksave()

	// Check if 'saves' dir exists in the temp dir
	savesDir := filepath.Join(dir, "saves")
	if _, err := os.Stat(savesDir); os.IsNotExist(err) {
		t.Error("'saves' directory was not created in temp dir")
	}

	// Verify a .oinakos file was created in savesDir/
	files, _ := os.ReadDir(savesDir)
	found := false
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "quicksave-") && strings.HasSuffix(f.Name(), ".oinakos.yaml") {
			found = true
			os.Remove(filepath.Join(savesDir, f.Name()))
		}
	}
	if !found {
		t.Error("No .oinakos quicksave file found")
	}
}

func TestLoad_Errors(t *testing.T) {
	// Create a temporary directory for the test
	dir, err := os.MkdirTemp("", "test_load_errors")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir) // Clean up the directory

	// Change to the temporary directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Ensure the oinakos/data/maps directory exists for the map file
	mapDir := filepath.Join("oinakos", "data", "maps")
	if err := os.MkdirAll(mapDir, 0755); err != nil {
		t.Fatalf("Failed to create map directory: %v", err)
	}

	// Create a dummy map file
	dummyMapPath := filepath.Join(mapDir, "test_load_errors_map.yaml")
	if err := os.WriteFile(dummyMapPath, []byte("map_data: {}"), 0644); err != nil {
		t.Fatalf("Failed to create dummy map file: %v", err)
	}

	SetOinakosDir(dir)
	defer SetOinakosDir("")

	g := NewGame(nil, &engine.MockGraphics{}, dummyMapPath, "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")

	// Ensure the saves directory exists for the save files in the temp dir
	saveDir := filepath.Join(dir, "saves")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		t.Fatalf("Failed to create save directory: %v", err)
	}

	// 1. Non-existent file
	err = g.Load(filepath.Join(saveDir, "nonexistent_file.yaml"))
	if err == nil {
		t.Error("Expected error loading non-existent file")
	}

	// 2. Corrupted YAML
	corruptPath := filepath.Join(saveDir, "corrupt.yaml")
	os.WriteFile(corruptPath, []byte("invalid: yaml: {{"), 0644)
	defer os.Remove(corruptPath)

	err = g.Load(corruptPath)
	if err == nil {
		t.Error("Expected error loading corrupted YAML")
	}

	// 3. Empty file
	emptyPath := filepath.Join(saveDir, "empty.yaml")
	os.WriteFile(emptyPath, []byte(""), 0644)
	defer os.Remove(emptyPath)
	err = g.Load(emptyPath)
	if err != nil {
		t.Errorf("Loading empty file should not fail, got: %v", err)
	}

	// 4. Map template instead of save
	templatePath := filepath.Join(saveDir, "template.yaml")
	os.WriteFile(templatePath, []byte("floor_tile: grass.png"), 0644)
	defer os.Remove(templatePath)
	err = g.Load(templatePath)
	if err == nil || !strings.Contains(err.Error(), "map template") {
		t.Errorf("Expected error loading map template, got: %v", err)
	}
}

func TestSave_InvalidPath(t *testing.T) {
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	err := g.Save("/invalid/dir/save.yaml")
	if err == nil {
		t.Error("Expected error saving to invalid directory")
	}
}
