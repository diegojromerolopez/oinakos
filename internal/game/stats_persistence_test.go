package game

import (
	"oinakos/internal/engine"
	"os"
	"testing"
)

func TestCharacterStatsInitialization(t *testing.T) {
	objReg := NewObjectRegistry()

	// 1. Test NPC range initialization
	configRange := &EntityConfig{
		ID: "orc",
		Stats: EntityStats{
			HealthMin: 50,
			HealthMax: 100,
			EnergyMin: 20.0,
			EnergyMax: 80.0,
		},
	}

	// Create 100 characters to verify we get values within range
	for i := 0; i < 100; i++ {
		c := NewCharacter(0, 0, configRange, 1, false, objReg)
		if c.Health < 50 || c.Health > 100 {
			t.Errorf("NPC Health %d out of range [50, 100]", c.Health)
		}
		if c.Energy < 20.0 || c.Energy > 80.0 {
			t.Errorf("NPC Energy %f out of range [20.0, 80.0]", c.Energy)
		}
	}

	// 2. Test Playable character specific initialization
	configPlayer := &EntityConfig{
		ID:     "hero",
		Health: 250,
		Energy: 75.5,
	}
	p := NewCharacter(0, 0, configPlayer, 1, true, objReg)
	if p.Health != 250 {
		t.Errorf("Player Health: got %d, want 250", p.Health)
	}
	if p.Energy != 75.5 {
		t.Errorf("Player Energy: got %f, want 75.5", p.Energy)
	}

	// 3. Test defaults
	configEmpty := &EntityConfig{ID: "nothing"}
	empty := NewCharacter(0, 0, configEmpty, 1, false, objReg)
	if empty.Energy != 100.0 {
		t.Errorf("Default Energy: got %f, want 100.0", empty.Energy)
	}
}

func TestStatsPersistence(t *testing.T) {
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	
	// Setup player
	g.playableCharacter.Health = 123
	g.playableCharacter.MaxHealth = 123
	g.playableCharacter.Energy = 45.6
	
	// Setup NPC
	npc := NewCharacter(10, 10, &EntityConfig{ID: "orc"}, 1, false, g.Registries.Objects)
	npc.Health = 77
	npc.MaxHealth = 77
	npc.Energy = 12.3
	g.characters = []*Character{npc}
	g.World.Characters = []*Character{npc}

	testPath := "stats_persistence_test.yaml"
	defer os.Remove(testPath)

	if err := g.Save(testPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into new game
	g2 := NewGame(nil, &engine.MockGraphics{}, "", "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	g2.characterRegistry.Characters["orc"] = &EntityConfig{ID: "orc"}

	if err := g2.Load(testPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify player
	if g2.playableCharacter.Health != 123 {
		t.Errorf("Loaded Player Health: got %d, want 123", g2.playableCharacter.Health)
	}
	if g2.playableCharacter.Energy != 45.6 {
		t.Errorf("Loaded Player Energy: got %f, want 45.6", g2.playableCharacter.Energy)
	}

	// Verify NPC
	if len(g2.characters) != 1 {
		t.Fatalf("Expected 1 NPC, got %d", len(g2.characters))
	}
	if g2.characters[0].Health != 77 {
		t.Errorf("Loaded NPC Health: got %d, want 77", g2.characters[0].Health)
	}
	if g2.characters[0].Energy != 12.3 {
		t.Errorf("Loaded NPC Energy: got %f, want 12.3", g2.characters[0].Energy)
	}
}
