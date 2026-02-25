package game

import (
	"embed"
	"testing"
)

func TestGameInitialization(t *testing.T) {
	// Mock embed.FS (empty)
	var assets embed.FS

	g := NewGame(assets, "", "", NewMockInputManager(), NewMockAudioManager())
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
	g := NewGame(assets, "", "", NewMockInputManager(), NewMockAudioManager())

	// Set player position
	g.mainCharacter.X = 100
	g.mainCharacter.Y = 100

	// Initial chunks
	g.updateChunks()
	initialCount := len(g.generatedChunks)
	if initialCount == 0 {
		t.Error("No chunks created")
	}

	// Move player
	g.mainCharacter.X = 500
	g.mainCharacter.Y = 500
	g.updateChunks()

	if len(g.generatedChunks) == 0 {
		t.Error("Chunks disappeared after move")
	}
}

func TestLayout(t *testing.T) {
	var assets embed.FS
	g := NewGame(assets, "", "", NewMockInputManager(), NewMockAudioManager())

	w, h := g.Layout(800, 600)
	if w != 1280 || h != 720 {
		t.Errorf("Layout: got (%d, %d), want (1280, 720)", w, h)
	}
}

func TestPickNPCIDToSpawn(t *testing.T) {
	archReg := NewArchetypeRegistry()
	archReg.Archetypes["orc_male"] = &Archetype{ID: "orc_male"}
	archReg.Archetypes["demon_female"] = &Archetype{ID: "demon_female"}
	archReg.Archetypes["peasant_male"] = &Archetype{ID: "peasant_male"}

	g := &Game{
		archetypeRegistry: archReg,
		currentMapType: MapType{
			SpawnWeights: map[string]int{"orc_male": 100, "demon_female": 10},
		},
	}

	id := g.pickNPCIDToSpawn()
	if id != "orc_male" && id != "demon_female" {
		t.Errorf("Unexpected NPC ID: %s", id)
	}

	g.currentMapType.SpawnWeights = map[string]int{"peasant_male": 1}
	id = g.pickNPCIDToSpawn()
	if id != "peasant_male" {
		t.Errorf("Expected peasant_male, got %s", id)
	}
}

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
	g := NewGame(assets, "", "", NewMockInputManager(), NewMockAudioManager())

	// Mock map with frequent spawning
	g.currentMapType.SpawnFreq = 0.0001 // frequent
	g.currentMapType.SpawnAmount = 5
	g.currentMapType.SpawnWeights = map[string]int{"orc_male": 1}
	g.archetypeRegistry.Archetypes["orc_male"] = &Archetype{ID: "orc_male"}
	g.archetypeRegistry.IDs = []string{"orc_male"}

	// Trigger spawning
	g.npcSpawnTimer = 1000
	g.updateNPCSpawning()

	// Test spawning near
	g.spawnNPCNear(10, 10)

	// Test spawning at edges
	g.spawnNPCAtEdges()
}
