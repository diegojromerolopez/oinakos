package game

import (
	"os"
	"strings"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	g := NewGame(nil, "data/maps/test_save.yaml", "", &MockInputManager{}, &MockAudioManager{}, false)
	g.mainCharacter.X = 123.45
	g.mainCharacter.Y = 67.89
	g.mainCharacter.Kills = 10
	g.mainCharacter.XP = 500
	g.mainCharacter.Health = 15
	g.playTime = 3600.0

	testPath := "test_save.yaml"
	defer os.Remove(testPath)

	if err := g.Save(testPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Create a new game and load
	g2 := NewGame(nil, "", "", NewMockInputManager(), NewMockAudioManager(), false)
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
