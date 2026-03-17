package game

import (
	"image/color"
	"oinakos/internal/engine"
	"testing"
)

type SolidMockImage struct {
	engine.MockImage
}

func (m *SolidMockImage) At(x, y int) color.Color {
	return color.RGBA{255, 255, 255, 255}
}

func TestOcclusionFeedback(t *testing.T) {
	// Setup a mock game
	input := engine.NewMockInput()
	g := NewGame(nil, "", "", "", input, nil, true, "0.1-test")
	
	// Create a mock actor
	mcConfig := &EntityConfig{ID: "player", Name: "Hero"}
	g.playableCharacter.Config = mcConfig
	g.playableCharacter.X = 0
	g.playableCharacter.Y = 0
	g.playableCharacter.IsOccluded = false

	// Create an obstacle in front of the actor
	// Isometric: x+y for sorting
	// Actor at (0,0) -> sortY = 0
	// Obstacle at (1,1) -> sortY = 2
	obsConfig := &ObstacleArchetype{
		ID:   "large_building",
		Type: "static",
		Image: &SolidMockImage{MockImage: engine.MockImage{W: 200, H: 200}},
	}
	
	// Position obstacle such that its sprite covers (0,0) in iso space
	// oIsoX, oIsoY = (1-1), (1+1)*0.5 = 0, 1
	// Actor iso: (0-0), (0+0)*0.5 = 0, 0
	obstacle := NewObstacle("obs1", 1, 1, obsConfig)
	g.obstacles = []*Obstacle{obstacle}

	// First update: should detect occlusion and log it
	g.updateOcclusion()

	if !g.playableCharacter.IsOccluded {
		t.Errorf("Expected player to be occluded")
	}

	found := false
	for _, event := range g.EventLog {
		if event.Text == "Hero is hidden behind an obstacle." {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected occlusion message in event log")
	}

	// Second update: should still be occluded but NOT log again
	initialLogCount := len(g.EventLog)
	g.updateOcclusion()
	if len(g.EventLog) > initialLogCount {
		t.Errorf("Should not log occlusion feedback repeatedly")
	}

	// Move actor in front of obstacle
	// Actor at (2,2) -> sortY = 4
	g.playableCharacter.X = 2
	g.playableCharacter.Y = 2
	g.updateOcclusion()

	if g.playableCharacter.IsOccluded {
		t.Errorf("Expected player to NOT be occluded after moving in front")
	}
}
