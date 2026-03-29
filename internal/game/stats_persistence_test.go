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
		Attributes: PrimaryAttributeConfig{ Health: IntInterval{Min: 10, Max: 10} }, // Will result in 100HP base
		Stats: EntityStatsConfig{
			HealthMin: IntInterval{Min: 50, Max: 50},
			HealthMax: IntInterval{Min: 100, Max: 100},
			HungerMax: FloatInterval{Min: 80.0, Max: 80.0},
			ThirstMax: FloatInterval{Min: 80.0, Max: 80.0},
			FatigueMax: FloatInterval{Min: 80.0, Max: 80.0},
			Age: AgeConfig{Current: FloatInterval{Min: 25, Max: 25}},
		},
	}

	// Create 100 characters to verify we get values within range
	for i := 0; i < 100; i++ {
		c := NewCharacter(0, 0, configRange, 1, false, objReg)
		if c.State.HealthPoints < 50 || c.State.HealthPoints > 100 {
			t.Errorf("NPC Health %d out of range [50, 100]", c.State.HealthPoints)
		}
		if c.State.Hunger > 80.0 || c.State.Thirst > 80.0 || c.State.Fatigue > 80.0 {
			t.Errorf("NPC Stats out of range")
		}
	}

	// 2. Test Playable character specific initialization
	configPlayer := &EntityConfig{
		ID:     "hero",
		Attributes: PrimaryAttributeConfig{ Health: IntInterval{Min: 25, Max: 25} }, // 250 HP
		State: State{
			Hunger: 75.5,
			Thirst: 75.5,
			Fatigue: 75.5,
			Age: AgeState{Current: 25.0},
		},
	}
	p := NewCharacter(0, 0, configPlayer, 1, true, objReg)
	if p.State.HealthPoints != 250 {
		t.Errorf("Player HealthPoints: got %d, want 250", p.State.HealthPoints)
	}
	if p.State.Hunger != 75.5 {
		t.Errorf("Player Hunger: got %f, want 75.5", p.State.Hunger)
	}

	// 3. Test defaults
	configEmpty := &EntityConfig{ID: "nothing"}
	empty := NewCharacter(0, 0, configEmpty, 1, false, objReg)
	if empty.State.Hunger != 0.0 {
		t.Errorf("Default Hunger: got %f, want 0.0", empty.State.Hunger)
	}
}

func TestStatsPersistence(t *testing.T) {
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", NewMockInputManager(), NewMockAudioManager(), false, "0.1-test")
	
	// Setup player
	g.playableCharacter.PrimaryAttributes.Health = 50
	g.playableCharacter.State.HealthPoints = 123
	g.playableCharacter.State.MaxHealthPoints = 500
	g.playableCharacter.State.Hunger = 45.6
	g.playableCharacter.State.Thirst = 45.6
	g.playableCharacter.State.Fatigue = 45.6
	
	// Setup NPC
	npc := NewCharacter(10, 10, &EntityConfig{ID: "orc"}, 1, false, g.Registries.Objects)
	npc.PrimaryAttributes.Health = 77
	npc.State.HealthPoints = 77
	npc.State.MaxHealthPoints = 770
	npc.State.Hunger = 12.3
	npc.State.Thirst = 12.3
	npc.State.Fatigue = 12.3
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
	if g2.playableCharacter.State.HealthPoints != 123 {
		t.Errorf("Loaded Player HealthPoints: got %d, want 123", g2.playableCharacter.State.HealthPoints)
	}
	if g2.playableCharacter.State.Hunger != 45.6 {
		t.Errorf("Loaded Player Hunger: got %f, want 45.6", g2.playableCharacter.State.Hunger)
	}

	// Verify NPC
	if len(g2.characters) != 1 {
		t.Fatalf("Expected 1 NPC, got %d", len(g2.characters))
	}
	if g2.characters[0].State.HealthPoints != 77 {
		t.Errorf("Loaded NPC HealthPoints: got %d, want 77", g2.characters[0].State.HealthPoints)
	}
	if g2.characters[0].State.Hunger != 12.3 {
		t.Errorf("Loaded NPC Hunger: got %f, want 12.3", g2.characters[0].State.Hunger)
	}
}
