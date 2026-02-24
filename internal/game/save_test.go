package game

import (
	"os"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	g := NewGame(nil)
	g.player.X = 123.45
	g.player.Y = 67.89
	g.player.Kills = 10
	g.player.XP = 500
	g.player.Health = 15
	g.playTime = 3600.0

	testPath := "test_save.json"
	defer os.Remove(testPath)

	if err := g.Save(testPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Create a new game and load
	g2 := NewGame(nil)
	if err := g2.Load(testPath); err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if g2.player.X != g.player.X || g2.player.Y != g.player.Y {
		t.Errorf("Position mismatch: expected (%f,%f), got (%f,%f)", g.player.X, g.player.Y, g2.player.X, g2.player.Y)
	}
	if g2.player.Kills != g.player.Kills {
		t.Errorf("Kills mismatch: expected %d, got %d", g.player.Kills, g2.player.Kills)
	}
	if g2.player.XP != g.player.XP {
		t.Errorf("XP mismatch: expected %d, got %d", g.player.XP, g2.player.XP)
	}
	if g2.player.Health != g.player.Health {
		t.Errorf("Health mismatch: expected %d, got %d", g.player.Health, g2.player.Health)
	}
	if g2.playTime != g.playTime {
		t.Errorf("PlayTime mismatch: expected %f, got %f", g.playTime, g2.playTime)
	}
}
