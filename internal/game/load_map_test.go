package game

import (
	"fmt"
	"testing"
	"testing/fstest"
)

func TestLoadMapLevel_AllObjectiveTypes(t *testing.T) {
	mockFS := fstest.MapFS{
		"data/archetypes/orc.yaml":         {Data: []byte("id: orc\nname: Orc\n")},
		"data/archetypes/magi_male.yaml":   {Data: []byte("id: magi_male\nname: Magi\n")},
		"data/obstacles/warehouse.yaml":    {Data: []byte("id: warehouse\nname: Warehouse\ncooldown_time: 1.0\n")},
		"data/obstacles/house_burned.yaml": {Data: []byte("id: house_burned\nname: Burned House\n")},
	}

	g := NewGame(mockFS, "", "", "", NewMockInputManager(), NewMockAudioManager(), false)
	g.archetypeRegistry.IDs = []string{"orc", "magi_male"}
	g.archetypeRegistry.Archetypes["orc"] = &EntityConfig{ID: "orc"}
	g.archetypeRegistry.Archetypes["magi_male"] = &EntityConfig{ID: "magi_male"}

	g.obstacleRegistry.IDs = []string{"warehouse", "house_burned"}
	g.obstacleRegistry.Archetypes["warehouse"] = &ObstacleArchetype{ID: "warehouse", CooldownTime: 1.0}
	g.obstacleRegistry.Archetypes["house_burned"] = &ObstacleArchetype{ID: "house_burned"}

	objectives := []ObjectiveType{
		ObjKillVIP,
		ObjReachPortal,
		ObjReachBuilding,
		ObjProtectNPC,
		ObjDestroyBuilding,
		ObjKillCount,
	}

	for _, obj := range objectives {
		t.Run(fmt.Sprintf("Objective_%d", obj), func(t *testing.T) {
			g.currentMapType = MapType{
				ID:   "test_map",
				Type: obj,
			}
			g.loadMapLevel()

			// Basic checks depending on objective
			switch obj {
			case ObjKillVIP:
				if len(g.npcs) == 0 {
					t.Error("VIP NPC not spawned")
				}
			case ObjReachBuilding:
				found := false
				for _, o := range g.obstacles {
					if o.ID == "target_warehouse" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Target building not spawned")
				}
			case ObjProtectNPC:
				if len(g.npcs) == 0 || g.npcs[0].Archetype.ID != "magi_male" {
					t.Errorf("Escort not spawned correctly: %+v", g.npcs)
				}
			case ObjDestroyBuilding:
				if g.currentMapType.TargetObstacle == nil {
					t.Error("Target obstacle not set")
				}
			}
		})
	}
}

func TestLoadMapLevel_InhabitantsAndObstacles(t *testing.T) {
	mockFS := fstest.MapFS{}
	g := NewGame(mockFS, "", "", "", NewMockInputManager(), NewMockAudioManager(), false)

	g.archetypeRegistry.Archetypes["orc"] = &EntityConfig{ID: "orc"}
	g.npcRegistry.NPCs["unique_orc"] = &EntityConfig{ID: "unique_orc", ArchetypeID: "orc"}
	g.obstacleRegistry.Archetypes["rock"] = &ObstacleArchetype{ID: "rock"}

	ten := 10.0
	g.currentMapType = MapType{
		Inhabitants: []Inhabitant{
			{Archetype: "orc", X: 5, Y: 5, State: "dead"},
			{NPC: "unique_orc", X: 15, Y: 15, Name: "Grimgor"},
		},
		Obstacles: []PreSpawnObstacle{
			{ID: "obs1", Archetype: "rock", X: &ten, Y: &ten},
			{ID: "obs_disabled", Archetype: "rock", Disabled: true},
		},
	}

	g.loadMapLevel()

	if len(g.npcs) != 2 {
		t.Errorf("Expected 2 NPCs, got %d", len(g.npcs))
	}
	if g.npcs[0].State != NPCDead {
		t.Errorf("Expected first NPC to be dead, got %v", g.npcs[0].State)
	}
	if g.npcs[1].Name != "Grimgor" {
		t.Errorf("Expected second NPC name to be Grimgor, got %s", g.npcs[1].Name)
	}

	foundObs := false
	for _, o := range g.obstacles {
		if o.ID == "obs1" {
			foundObs = true
			if o.X != 10 || o.Y != 10 {
				t.Errorf("Obstacle pos mismatch: got (%f,%f), want (10,10)", o.X, o.Y)
			}
		}
		if o.ID == "obs_disabled" {
			t.Error("Disabled obstacle should not be spawned")
		}
	}
	if !foundObs {
		t.Error("PreSpawn obstacle not found")
	}
}

func TestLoadMapLevel_Campaign(t *testing.T) {
	g := NewGame(nil, "", "", "", NewMockInputManager(), NewMockAudioManager(), false)
	g.mapTypeRegistry.Types["m1"] = &MapType{ID: "m1", Name: "Map 1"}
	g.mapTypeRegistry.Types["m2"] = &MapType{ID: "m2", Name: "Map 2"}

	g.currentCampaign = &Campaign{
		ID:   "c1",
		Maps: []string{"m1", "m2"},
	}
	g.isCampaign = true
	g.campaignIndex = 1

	g.loadMapLevel()

	if g.currentMapType.ID != "m2" {
		t.Errorf("Expected campaign map m2, got %s", g.currentMapType.ID)
	}
}

func TestLoadMapLevel_InitialMapID(t *testing.T) {
	g := NewGame(nil, "", "", "", NewMockInputManager(), NewMockAudioManager(), false)
	g.mapTypeRegistry.Types["init_map"] = &MapType{ID: "init_map", Name: "Initial Map"}
	g.initialMapID = "init_map"
	g.mapLevel = 1
	g.isCampaign = false

	g.loadMapLevel()

	if g.currentMapType.ID != "init_map" {
		t.Errorf("Expected initial map init_map, got %s", g.currentMapType.ID)
	}
}
