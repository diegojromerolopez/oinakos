package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestWorldManager_SpawningAdvanced(t *testing.T) {
	graphics := &engine.MockGraphics{}
	g := NewGame(nil, graphics, "", "", "", NewMockInputManager(), &MockAudioManager{}, false, "0.1-test")
	wm := g.worldManager
	
	sc := &SpawnConfig{
		Archetype:   "orc",
		Probability: 1.0,
		Alignment:   AlignmentEnemy,
		Frequency:   1.0,
	}
	
	g.currentMapType = MapType{
		ID: "test",
		Spawns: []SpawnConfig{*sc},
	}
	g.archetypeRegistry.Archetypes["orc"] = &Archetype{ID: "orc"}
	g.archetypeRegistry.IDs = []string{"orc"}
	
	// 1. Force spawn near position
	wm.spawnNPCNearPosition(0, 0, sc)
	if len(g.characters) != 1 {
		t.Errorf("Expected 1 NPC, got %d", len(g.characters))
	}
	
	// 2. Force spawn at map edges
	g.currentMapType.MapWidth = 100
	g.currentMapType.MapHeight = 100
	wm.spawnNPCAtMapEdges(sc)
	if len(g.characters) != 2 {
		t.Errorf("Expected 2 NPCs, got %d", len(g.characters))
	}
	
	// 3. Update spawning logic
	wm.UpdateNPCSpawning()
}
