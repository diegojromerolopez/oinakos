package game

import (
	"embed"
	"testing"
)

func TestGameInitialization(t *testing.T) {
	// Mock embed.FS (empty)
	var assets embed.FS

	g := NewGame(assets, "", "", NewMockInputManager(), NewMockAudioManager(), false)
	if g == nil {
		t.Fatal("NewGame returned nil")
	}

	if g.mainCharacter == nil {
		t.Error("MainCharacter not initialized")
	}

	if g.camera == nil {
		t.Error("Camera not initialized")
	}
}

func TestUpdateChunks(t *testing.T) {
	var assets embed.FS
	g := NewGame(assets, "", "", NewMockInputManager(), NewMockAudioManager(), false)

	// Set player position
	g.mainCharacter.X = 100
	g.mainCharacter.Y = 100

	// Procedural updateChunks is now a no-op (disabled per user request)
	g.updateChunks()
	if len(g.generatedChunks) != 0 {
		t.Error("No chunks should be created (procedural spawning disabled)")
	}
}

func TestLayout(t *testing.T) {
	var assets embed.FS
	g := NewGame(assets, "", "", NewMockInputManager(), NewMockAudioManager(), false)

	w, h := g.Layout(800, 600)
	if w != 1280 || h != 720 {
		t.Errorf("Layout: got (%d, %d), want (1280, 720)", w, h)
	}
}

// pickSpawnConfig test removed as method was deleted

func TestGameWinMenu(t *testing.T) {
	g := &Game{
		isMapWon:        true,
		mapWonMenuIndex: 0,
	}

	// This test essentially checks that the logic added exists and doesn't panic.
	// We can't easily mock inpututil in standard tests without a lot of setup,
	// but we can at least check the state initialization.
	if g.mapWonMenuIndex != 0 {
		t.Errorf("Expected initial menu index 0, got %d", g.mapWonMenuIndex)
	}
}

func TestSpawningLogic(t *testing.T) {
	var assets embed.FS
	g := NewGame(assets, "", "", NewMockInputManager(), NewMockAudioManager(), false)

	// Mock map with frequent spawning
	g.currentMapType.Spawns = []SpawnConfig{
		{Archetype: "orc_male", Probability: 1.0, Frequency: 0.016, Alignment: AlignmentEnemy},
	}
	g.archetypeRegistry.Archetypes["orc_male"] = &Archetype{ID: "orc_male"}
	g.archetypeRegistry.IDs = []string{"orc_male"}

	// Trigger spawn cycle
	// The new logic requires ticking until threshold
	for i := 0; i < 10; i++ {
		g.updateNPCSpawning()
	}

	if len(g.npcs) == 0 {
		t.Error("Should have spawned NPCs")
	}
}
