package game

import (
	"testing"
	"testing/fstest"
)

func TestObjectiveLogic(t *testing.T) {
	// 1. Test Reach Portal objective
	mockFS := fstest.MapFS{
		"data/map_types/portal.yaml": {
			Data: []byte(`id: "portal"
name: "Portal Reach"
type: "reach_portal"
difficulty: 1
width_px: 1000
height_px: 1000
target_point:
  x: 5.0
  y: 5.0
target_radius: 2.0
`),
		},
		"data/map_types/survive.yaml": {
			Data: []byte(`id: "survive"
name: "Survive Time"
type: "survive"
difficulty: 1
target_time: 10.0
`),
		},
		"data/map_types/building.yaml": {
			Data: []byte(`id: "building"
name: "Destroy Building"
type: "destroy_building"
difficulty: 1
obstacles: 
  - id: target_building
    archetype: house
    x: 0
    y: 0
`),
		},
	}

	// Reach Point — target_point is defined in the YAML, so no override needed
	g := NewGame(mockFS, "portal", "", "", NewMockInputManager(), NewMockAudioManager(), false)
	g.isMainMenu = false
	g.isCharacterSelect = false
	g.playableCharacter.X = 0
	g.playableCharacter.Y = 0
	g.Update()
	if g.isMapWon {
		t.Error("Portal objective should not be won at (0,0)")
	}
	g.playableCharacter.X = 5.1
	g.playableCharacter.Y = 5.1
	g.Update()
	if !g.isMapWon {
		t.Errorf("Portal win logic failed at (5.1,5.1), distance to (5,5) should be within 2.0 radius (actual dist = ~0.14)")
	}

	// Survive Time
	g = NewGame(mockFS, "survive", "", "", NewMockInputManager(), NewMockAudioManager(), false)
	g.isMainMenu = false
	g.isCharacterSelect = false
	g.playTime = 5.0
	g.Update()
	if g.isMapWon {
		t.Error("Survive objective should not be won at 5s (target 10s)")
	}
	g.playTime = 11.0
	g.Update()
	if !g.isMapWon {
		t.Error("Survive objective win logic failed")
	}

	// Destroy Building — test the win condition logic directly without relying on YAML loading
	g = NewGame(mockFS, "building", "", "", NewMockInputManager(), NewMockAudioManager(), false)
	g.isMainMenu = false
	g.isCharacterSelect = false
	// Inject the target obstacle directly into game state
	targetObstacle := NewObstacle("target_building", 0, 0, &ObstacleArchetype{ID: "house", Health: 500})
	targetObstacle.Alive = true
	g.obstacles = []*Obstacle{targetObstacle}
	g.currentMapType.TargetObstacle = targetObstacle

	g.Update()
	if g.isMapWon {
		t.Error("Destroy objective won but building is still alive")
	}
	targetObstacle.Alive = false
	g.Update()
	if !g.isMapWon {
		t.Errorf("Destroy objective failed even after building died")
	}
}
