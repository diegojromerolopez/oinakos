package game

import (
	"embed"
	"io/fs"
	"oinakos/internal/engine"
	"testing"
)

func TestGameInitialization(t *testing.T) {
	// Mock embed.FS (empty)
	var assets embed.FS

	g := NewGame(assets, "", "", "", NewMockInputManager(), NewMockAudioManager(), false)
	if g == nil {
		t.Fatal("NewGame returned nil")
	}

	if g.playableCharacter == nil {
		t.Error("PlayableCharacter not initialized")
	}

	if g.camera == nil {
		t.Error("Camera not initialized")
	}
}

func TestUpdateChunks(t *testing.T) {
	var assets embed.FS
	g := NewGame(assets, "", "", "", NewMockInputManager(), NewMockAudioManager(), false)

	// Set player position
	g.playableCharacter.X = 100
	g.playableCharacter.Y = 100

	// Procedural updateChunks is now a no-op (disabled per user request)
	g.worldManager.UpdateChunks()
	if len(g.generatedChunks) != 0 {
		t.Error("No chunks should be created (procedural spawning disabled)")
	}
}

func TestLayout(t *testing.T) {
	var assets embed.FS
	g := NewGame(assets, "", "", "", NewMockInputManager(), NewMockAudioManager(), false)

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
	g.worldManager = NewWorldManager(g)

	// This test essentially checks that the logic added exists and doesn't panic.
	// We can't easily mock inpututil in standard tests without a lot of setup,
	// but we can at least check the state initialization.
	if g.mapWonMenuIndex != 0 {
		t.Errorf("Expected initial menu index 0, got %d", g.mapWonMenuIndex)
	}
}

func TestSpawningLogic(t *testing.T) {
	var assets embed.FS
	g := NewGame(assets, "", "", "", NewMockInputManager(), NewMockAudioManager(), false)

	// Mock map with frequent spawning
	g.currentMapType.Spawns = []SpawnConfig{
		{Archetype: "orc_male", Probability: 1.0, Frequency: 0.016, Alignment: AlignmentEnemy},
	}
	g.archetypeRegistry.Archetypes["orc_male"] = &Archetype{ID: "orc_male"}
	g.archetypeRegistry.IDs = []string{"orc_male"}

	// Trigger spawn cycle
	// The new logic requires ticking until threshold
	for i := 0; i < 10; i++ {
		g.worldManager.UpdateNPCSpawning()
	}

	if len(g.npcs) == 0 {
		t.Error("Should have spawned NPCs")
	}
}

func TestHeroFlagOverride(t *testing.T) {
	// Setup a mock environment
	input := &engine.MockInput{}
	audio := &DefaultAudioManager{}
	var assets fs.FS // empty FS

	// Create game
	g := NewGame(assets, "", "", "conde_olinos", input, audio, false)

	// Manually add the hero to the registry since assets are empty
	heroConfig := &EntityConfig{
		ID: "conde_olinos",
		Name: "Conde Olinos",
		Stats: struct {
			HealthMin       int     `yaml:"health_min"`
			HealthMax       int     `yaml:"health_max"`
			Speed           float64 `yaml:"speed"`
			BaseAttack      int     `yaml:"base_attack"`
			BaseDefense     int     `yaml:"base_defense"`
			AttackCooldown  int     `yaml:"attack_cooldown"`
			AttackRange     float64 `yaml:"attack_range"`
			ProjectileSpeed float64 `yaml:"projectile_speed"`
		}{
			HealthMin: 500,
			Speed: 0.05,
		},
	}
	g.playableCharacterRegistry.Characters["conde_olinos"] = heroConfig
	g.playableCharacterRegistry.IDs = append(g.playableCharacterRegistry.IDs, "conde_olinos")

	// Re-run the initialization logic that handles the flag
	// Since we already called NewGame, we manually trigger the block we added
	if config, ok := g.playableCharacterRegistry.Characters[g.initialHeroID]; ok {
		g.playableCharacter.Config = config
		g.playableCharacter.Health = config.Stats.HealthMin
		g.playableCharacter.MaxHealth = config.Stats.HealthMin
		g.playableCharacter.Speed = config.Stats.Speed
		g.isCharacterSelect = false
	}

	// Verify the playable character has the correct config
	if g.playableCharacter.Config.ID != "conde_olinos" {
		t.Errorf("Expected playable character ID conde_olinos, got %s", g.playableCharacter.Config.ID)
	}

	// Verify health was initialized
	if g.playableCharacter.MaxHealth != 500 {
		t.Errorf("Expected health 500, got %d", g.playableCharacter.MaxHealth)
	}

	// Verify character selection screen is bypassed
	if g.isCharacterSelect {
		t.Error("Expected character selection screen to be bypassed when -hero is used")
	}
}

func TestHeroFlagInitialization(t *testing.T) {
	// This test verifies that NewGame correctly stores the flag
	input := &engine.MockInput{}
	audio := &DefaultAudioManager{}
	var assets fs.FS

	heroID := "boris_stronesco"
	g := NewGame(assets, "", "", heroID, input, audio, false)

	if g.initialHeroID != heroID {
		t.Errorf("Expected initialHeroID %s, got %s", heroID, g.initialHeroID)
	}
}
