package game

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGame_SaveAndSerialize(t *testing.T) {
	g := setupTestGame()
	g.currentMapType = MapType{ID: "test_map", WidthPixels: 640, HeightPixels: 320}
	
	// Add some data to serialize
	g.playableCharacter.Inventory = []*ItemInstance{
		{Config: &ObjectConfig{ID: "sword"}, Resistance: 10},
	}
	
	g.characters = []*Character{
		NewCharacter(10, 10, &EntityConfig{ID: "orc", Unique: false}, 1, false, nil),
	}
	
	g.obstacles = []*Obstacle{
		NewObstacle("tree", 5, 5, &ObstacleArchetype{ID: "oak"}),
	}
	
	g.World.Items = []*ItemInstance{
		{ID: "loot_1", X: 2, Y: 2, Resistance: 100},
	}
	
	// Test serialize
	data, err := g.serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Serialized data is empty")
	}
	
	// Test Save to file
	tmpFile := filepath.Join(t.TempDir(), "save.yaml")
	err = g.Save(tmpFile)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Save file was not created")
	}
}

func TestGame_PerformQuicksave(t *testing.T) {
	g := setupTestGame()
	g.currentMapType = MapType{ID: "test_map"}
	
	tmpDir := t.TempDir()
	SetOinakosDir(tmpDir)
	defer SetOinakosDir("")
	
	g.performQuicksave()
	
	saveDir := filepath.Join(tmpDir, "saves")
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		t.Fatalf("Failed to read saves dir: %v", err)
	}
	
	if len(entries) == 0 {
		t.Error("Quicksave file not found in saves directory")
	}
	
	if g.saveMessage == "" {
		t.Error("saveMessage should be set after quicksave")
	}
}

func TestGame_SerializeDetailed(t *testing.T) {
	g := setupTestGame()
	g.currentMapType = MapType{ID: "complex_save", WidthPixels: 1000, HeightPixels: 1000}
	
	npcs := []*Character{
		NewCharacter(10, 10, &EntityConfig{ID: "wanderer", Unique: false}, 1, false, nil),
		NewCharacter(20, 20, &EntityConfig{ID: "patrolman", Unique: false}, 2, false, nil),
		NewCharacter(30, 30, &EntityConfig{ID: "hunter", Unique: false}, 3, false, nil),
		NewCharacter(40, 40, &EntityConfig{ID: "fighter", Unique: false}, 4, false, nil),
		NewCharacter(50, 50, &EntityConfig{ID: "chaotic", Unique: false}, 5, false, nil),
		NewCharacter(60, 60, &EntityConfig{ID: "unique_npc", Unique: true}, 6, false, nil),
	}
	// Initializing behaviors to ensure switch coverage
	npcs[0].Behavior = BehaviorWander
	npcs[1].Behavior = BehaviorPatrol
	npcs[2].Behavior = BehaviorKnightHunter
	npcs[3].Behavior = BehaviorNpcFighter
	npcs[4].Behavior = BehaviorChaotic
	
	// Add inventory and slots to NPCs
	npcs[0].Inventory = []*ItemInstance{
		{Config: &ObjectConfig{ID: "wood"}, Resistance: 5},
	}
	if npcs[5].Slots == nil { npcs[5].Slots = make(map[string]*ItemInstance) }
	npcs[5].Slots["head"] = &ItemInstance{Config: &ObjectConfig{ID: "iron_helmet"}, Resistance: 100}

	g.characters = npcs
	
	// Player with complicated state
	if g.playableCharacter.Slots == nil { g.playableCharacter.Slots = make(map[string]*ItemInstance) }
	g.playableCharacter.Slots["body"] = &ItemInstance{Config: &ObjectConfig{ID: "leather_armor"}, Resistance: 50}
	g.playableCharacter.Weapon = &Weapon{Name: "Epic Sword", Damage: Damage{Min: 10, Max: 20}}
	
	g.ExploredTiles[image.Pt(1, 1)] = true
	g.ExploredTiles[image.Pt(2, 2)] = true
	
	data, err := g.serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Serialized data is empty")
	}
	
	var sd SaveData
	if err := yaml.Unmarshal(data, &sd); err != nil {
		t.Fatalf("YAML Unmarshal failed: %v", err)
	}
	
	if len(sd.Characters) != 6 {
		t.Errorf("Expected 6 characters, got %d", len(sd.Characters))
	}
	
	// Verify behavior mapping
	behaviors := make(map[string]bool)
	for _, c := range sd.Characters {
		if c.Behavior != "" {
			behaviors[c.Behavior] = true
		}
	}
	expectedBehaviors := []string{"wander", "patrol", "hunter", "fighter", "chaotic"}
	for _, b := range expectedBehaviors {
		if !behaviors[b] {
			t.Errorf("Behavior %s not found in serialized data", b)
		}
	}
}

func TestGame_SaveFailure(t *testing.T) {
	g := setupTestGame()
	g.currentMapType = MapType{ID: "test"}
	
	// Attempt to save to an invalid path
	err := g.Save("/this/path/should/not/exist/save.yaml")
	if err == nil {
		t.Error("expected error when saving to invalid path, got nil")
	}
}

func TestGame_PerformQuicksave_MkdirFail(t *testing.T) {
	g := setupTestGame()
	g.currentMapType = MapType{ID: "test"}
	
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "not_a_dir")
	os.WriteFile(tmpFile, []byte("file"), 0644)
	
	// Create a situation where os.MkdirAll fails
	// Point oinakos dir to a sub-path of the file
	SetOinakosDir(filepath.Join(tmpFile, "subdir"))
	defer SetOinakosDir("")
	
	g.performQuicksave()
	// Should log error but not panic
}
func TestGame_LoadAdvanced(t *testing.T) {
	g := setupTestGame()
	tmpFile := filepath.Join(t.TempDir(), "save_adv.yaml")
	
	// Prepare save data with many different configurations
	data := SaveData{}
	data.Map.ID = "test_map"
	data.Map.Level = 5
	data.Map.PlayTime = 3600
	data.Map.ExploredTiles = []image.Point{{X: 1, Y: 1}}
	data.Map.Overrides.Name = "Overridden Name"
	data.Map.Overrides.TargetTime = 120.0
	
	data.Player = PlayerSaveData{
		Archetype: "conde_olinos",
		X: 50, Y: 50, 
		State: State{HealthPoints: 80, MaxHealthPoints: 100},
		Level: 1,
		Inventory: []ItemInstanceSaveData{{ID: "sword_iron", Resistance: 50, X: 0, Y: 0}},
		Slots: map[string]ItemInstanceSaveData{"weapon": {ID: "sword_iron", Resistance: 50, X: 0, Y: 0}},
	}
	data.Characters = []NPCSaveData{
		{Archetype: "orc_male", X: 10, Y: 10, State: State{HealthPoints: 50, MaxHealthPoints: 50}, Level: 1, Behavior: "wander"},
		{NPCID: "unique_npc", X: 20, Y: 20, State: State{HealthPoints: 100, MaxHealthPoints: 100}, Level: 1, Behavior: "hunter", Alignment: AlignmentEnemy},
		{Archetype: "orc_male", X: 30, Y: 30, State: State{HealthPoints: 0, MaxHealthPoints: 50}, Level: 1, Behavior: "patrol"}, // Dead NPC
	}
	data.Obstacles = []ObstacleSaveData{
		{ID: "tree_1", Archetype: "tree_oak", X: ptr(5.0), Y: ptr(5.0), HealthPoints: 100},
		{ID: "broken_tree", Archetype: "tree_oak", HealthPoints: 0}, // Broken obstacle
	}
	data.Items = []ItemInstanceSaveData{
		{ID: "gold_ore_1", X: 2, Y: 2},
	}
	
	bytes, _ := yaml.Marshal(data)
	os.WriteFile(tmpFile, bytes, 0644)
	
	// Setup registries to allow successful load
	g.mapTypeRegistry.Types["test_map"] = &MapType{ID: "test_map"}
	g.characterRegistry.Characters["conde_olinos"] = &EntityConfig{ID: "conde_olinos", Name: "Conde"}
	g.characterRegistry.Characters["unique_npc"] = &EntityConfig{ID: "unique_npc", Name: "Unique"}
	g.archetypeRegistry.Archetypes["orc_male"] = &Archetype{ID: "orc_male"}
	g.obstacleRegistry.Archetypes["tree_oak"] = &ObstacleArchetype{ID: "tree_oak", HealthPoints: 100}
	g.Registries.Objects.Objects["sword_iron"] = &ObjectConfig{ID: "sword_iron"}
	g.Registries.Objects.Objects["gold_ore_1"] = &ObjectConfig{ID: "gold_ore_1"}
	
	err := g.Load(tmpFile)
	if err != nil {
		t.Fatalf("Advanced Load failed: %v", err)
	}
	
	if g.mapLevel != 5 { t.Errorf("Expected map level 5, got %d", g.mapLevel) }
	if len(g.characters) != 3 { t.Errorf("Expected 3 characters, got %d", len(g.characters)) }
	if len(g.obstacles) != 2 { t.Errorf("Expected 2 obstacles, got %d", len(g.obstacles)) }
}

func ptr(f float64) *float64 { return &f }
