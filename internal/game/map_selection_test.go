package game

import (
	"testing"
	"testing/fstest"
)

func TestMapSelection(t *testing.T) {
	// Create a mock filesystem
	mockFS := fstest.MapFS{
		"data/map_types/type1.yaml": {
			Data: []byte(`id: "type1"
name: "Type One"
type: "kill_vip"
difficulty: 1
width_px: 1000
height_px: 1000
`),
		},
		"data/maps/default1.yaml": {
			Data: []byte(`map:
  id: "type1"
  level: 1
player:
  x: 10
  y: 20
  health: 100
  max_health: 100
`),
		},
		"data/maps/priority.yaml": {
			Data: []byte(`map:
  id: "type1"
  level: 99
player:
  x: 99
  y: 99
`),
		},
		"priority.yaml": {
			Data: []byte(`map:
  id: "type1"
  level: 1
player:
  x: 1
  y: 1
`),
		},
		"external/ext1.yaml": {
			Data: []byte(`map:
  id: "type1"
  level: 5
player:
  x: 100
  y: 200
  health: 50
  max_health: 100
`),
		},
		"data/maps/test.yml": {
			Data: []byte(`map:
  id: "type1"
  level: 7
player:
  x: 7
  y: 7
`),
		},
	}

	t.Run("LoneString_DefaultMap", func(t *testing.T) {
		g := NewGame(mockFS, "default1", "", NewMockInputManager(), NewMockAudioManager())
		if g.mainCharacter.X != 10 || g.mainCharacter.Y != 20 {
			t.Errorf("Expected player at (10, 20), got (%f, %f)", g.mainCharacter.X, g.mainCharacter.Y)
		}
	})

	t.Run("ImplicitExtension_DataMaps", func(t *testing.T) {
		g := NewGame(mockFS, "default1", "", NewMockInputManager(), NewMockAudioManager())
		if g.currentMapType.ID != "type1" {
			t.Errorf("Expected map type type1, got %s", g.currentMapType.ID)
		}
	})

	t.Run("ImplicitExtension_Path", func(t *testing.T) {
		g := NewGame(mockFS, "external/ext1", "", NewMockInputManager(), NewMockAudioManager())
		if g.mapLevel != 5 {
			t.Errorf("Expected map level 5 (loaded via implicit ext), got %d", g.mapLevel)
		}
	})

	t.Run("YmlExtension", func(t *testing.T) {
		g := NewGame(mockFS, "test", "", NewMockInputManager(), NewMockAudioManager())
		if g.mapLevel != 7 {
			t.Errorf("Expected map level 7 (loaded via .yml), got %d", g.mapLevel)
		}
	})

	t.Run("PrioritySearch", func(t *testing.T) {
		// priority.yaml exists in root and data/maps/. Root should win.
		g := NewGame(mockFS, "priority.yaml", "", NewMockInputManager(), NewMockAudioManager())
		if g.mapLevel != 1 {
			t.Errorf("Priority mismatch: expected level 1 (root), got %d", g.mapLevel)
		}
	})

	t.Run("LoneString_FallbackToMapType", func(t *testing.T) {
		g := NewGame(mockFS, "type1", "", NewMockInputManager(), NewMockAudioManager())
		if g.mainCharacter.X != 0 || g.mainCharacter.Y != 0 {
			t.Errorf("Expected player at (0, 0), got (%f, %f)", g.mainCharacter.X, g.mainCharacter.Y)
		}
	})

	t.Run("Path_ExternalMap", func(t *testing.T) {
		g := NewGame(mockFS, "external/ext1.yaml", "", NewMockInputManager(), NewMockAudioManager())
		if g.mainCharacter.X != 100 || g.mainCharacter.Y != 200 {
			t.Errorf("Expected player at (100, 200), got (%f, %f)", g.mainCharacter.X, g.mainCharacter.Y)
		}
	})

	t.Run("NonExistent_LoneString", func(t *testing.T) {
		g := NewGame(mockFS, "nonexistent", "", NewMockInputManager(), NewMockAudioManager())
		// Warning will be logged, but it shouldn't crash
		t.Log("Map type fallback:", g.currentMapType.ID)
	})
}
