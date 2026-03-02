package game

import (
	"os"
	"strings"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	g := NewGame(nil, "data/maps/test_save.yaml", "", &MockInputManager{}, &MockAudioManager{}, false)
	// Add NPC and Obstacle to test persistence
	g.npcs = []*NPC{NewNPC(10, 20, &Archetype{ID: "test_npc"}, 1)}
	g.npcs[0].Health = 5
	g.obstacles = []*Obstacle{NewObstacle("test_building", 30, 40, &ObstacleArchetype{ID: "test_arch", Health: 100})}

	testPath := "test_save.yaml"
	defer os.Remove(testPath)

	if err := g.Save(testPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Create a new game and load
	g2 := NewGame(nil, "", "", NewMockInputManager(), NewMockAudioManager(), false)
	// Mock registries for loading to work
	g2.npcRegistry.IDs = []string{"test_npc"}
	g2.npcRegistry.NPCs["test_npc"] = &EntityConfig{ArchetypeID: "test_npc"}
	g2.archetypeRegistry.Archetypes["test_npc"] = &Archetype{ID: "test_npc"}
	g2.obstacleRegistry.IDs = []string{"test_arch"}
	g2.obstacleRegistry.Archetypes["test_arch"] = &ObstacleArchetype{ID: "test_arch", Health: 100}

	if err := g2.Load(testPath); err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if g2.mainCharacter.X != g.mainCharacter.X || g2.mainCharacter.Y != g.mainCharacter.Y {
		t.Errorf("Position mismatch: expected (%f,%f), got (%f,%f)", g.mainCharacter.X, g.mainCharacter.Y, g2.mainCharacter.X, g2.mainCharacter.Y)
	}
	if g2.mainCharacter.Kills != g.mainCharacter.Kills {
		t.Errorf("Kills mismatch: expected %d, got %d", g.mainCharacter.Kills, g2.mainCharacter.Kills)
	}
	if g2.mainCharacter.XP != g.mainCharacter.XP {
		t.Errorf("XP mismatch: expected %d, got %d", g.mainCharacter.XP, g2.mainCharacter.XP)
	}
	if g2.mainCharacter.Health != g.mainCharacter.Health {
		t.Errorf("Health mismatch: expected %d, got %d", g.mainCharacter.Health, g2.mainCharacter.Health)
	}
	if g2.playTime != g.playTime {
		t.Errorf("PlayTime mismatch: expected %f, got %f", g.playTime, g2.playTime)
	}

	if len(g2.npcs) != 1 {
		t.Errorf("NPCs mismatch: expected 1, got %d", len(g2.npcs))
	} else if g2.npcs[0].X != 10 || g2.npcs[0].Y != 20 {
		t.Errorf("NPC pos mismatch: expected (10,20), got (%f,%f)", g2.npcs[0].X, g2.npcs[0].Y)
	}

	if len(g2.obstacles) != 1 {
		t.Errorf("Obstacles mismatch: expected 1, got %d", len(g2.obstacles))
	} else if g2.obstacles[0].ID != "test_building" {
		t.Errorf("Building ID mismatch: expected 'test_building', got '%s'", g2.obstacles[0].ID)
	}
}

func TestQuickSave(t *testing.T) {
	g := NewGame(nil, "", "", NewMockInputManager(), NewMockAudioManager(), false)
	g.performQuicksave()

	// Check if 'saves' dir exists
	if _, err := os.Stat("saves"); os.IsNotExist(err) {
		t.Error("'saves' directory was not created")
	}

	// Verify a .oinakos file was created in saves/
	files, _ := os.ReadDir("saves")
	found := false
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "quicksave-") && strings.HasSuffix(f.Name(), ".oinakos.yaml") {
			found = true
			os.Remove("saves/" + f.Name())
		}
	}
	if !found {
		t.Error("No .oinakos quicksave file found")
	}
	os.Remove("saves") // Clean up if empty
}

func TestLoad_Errors(t *testing.T) {
	g := NewGame(nil, "", "", NewMockInputManager(), NewMockAudioManager(), false)

	// 1. Non-existent file
	err := g.Load("nonexistent_file.yaml")
	if err == nil {
		t.Error("Expected error loading non-existent file")
	}

	// 2. Corrupted YAML
	corruptPath := "corrupt.yaml"
	os.WriteFile(corruptPath, []byte("invalid: yaml: {{"), 0644)
	defer os.Remove(corruptPath)

	err = g.Load(corruptPath)
	if err == nil {
		t.Error("Expected error loading corrupted YAML")
	}

	// 3. Empty file
	emptyPath := "empty.yaml"
	os.WriteFile(emptyPath, []byte(""), 0644)
	defer os.Remove(emptyPath)
	err = g.Load(emptyPath)
	if err != nil {
		t.Errorf("Loading empty file should not fail, got: %v", err)
	}
}

func TestSave_InvalidPath(t *testing.T) {
	g := NewGame(nil, "", "", NewMockInputManager(), NewMockAudioManager(), false)
	err := g.Save("/invalid/dir/save.yaml")
	if err == nil {
		t.Error("Expected error saving to invalid directory")
	}
}
