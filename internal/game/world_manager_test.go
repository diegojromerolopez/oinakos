package game

import (
	"testing"
	"testing/fstest"
)

func TestWorldManager_LoadMapAssets(t *testing.T) {
	fsys := fstest.MapFS{
		"assets/audio/archetypes/orc/male/attack_1.wav": &fstest.MapFile{Data: []byte("wav")},
	}
	
	g := setupTestGame()
	g.assets = fsys
	wm := NewWorldManager(g)
	
	g.archetypeRegistry.Archetypes["orc"] = &Archetype{
		ID:       "orc",
		AudioDir: "assets/audio/archetypes/orc/male",
	}
	
	g.characters = []*Character{
		NewCharacter(0, 0, &EntityConfig{ID: "orc_warrior", Archetype: "orc"}, 1, false, nil),
	}
	
	// Should not panic and should at least process jobs
	wm.LoadMapAssets()
	
	if g.LoadingProgress < 0 {
		t.Errorf("LoadingProgress should be >= 0, got %d", g.LoadingProgress)
	}
}

func TestWorldManager_UpdateNPCSpawning(t *testing.T) {
	g := setupTestGame()
	wm := NewWorldManager(g)
	
	g.archetypeRegistry.Archetypes["orc"] = &Archetype{ID: "orc"}
	g.archetypeRegistry.IDs = []string{"orc"}
	
	x, y := 10.0, 10.0
	g.currentMapType = MapType{
		Spawns: []SpawnConfig{
			{
				Archetype:   "orc",
				Alignment:   AlignmentEnemy,
				Probability: 1.0,
				Frequency:   1.0 / 60.0, // Should spawn every tick? No, Frequency is in seconds.
				// Frequency 1.0/60.0 * 60 = 1 tick.
				X: &x,
				Y: &y,
			},
		},
	}
	
	g.characters = nil
	g.World.Characters = nil
	wm.UpdateNPCSpawning()
	
	if len(g.characters) != 1 {
		t.Errorf("expected 1 NPC to be spawned, got %d", len(g.characters))
	}
	
	npc := g.characters[0]
	if npc.Config.ID != "orc" {
		t.Errorf("expected orc to be spawned, got %s", npc.Config.ID)
	}
	if npc.Alignment != AlignmentEnemy {
		t.Errorf("expected enemy alignment, got %v", npc.Alignment)
	}
}

func TestWorldManager_SpawnMethods(t *testing.T) {
	g := setupTestGame()
	wm := NewWorldManager(g)
	
	g.archetypeRegistry.Archetypes["orc"] = &Archetype{ID: "orc"}
	g.archetypeRegistry.IDs = []string{"orc"}
	
	sc := &SpawnConfig{
		Archetype: "orc",
		Alignment: AlignmentEnemy,
	}
	
	t.Run("SpawnNearPosition", func(t *testing.T) {
		g.characters = nil
		wm.spawnNPCNearPosition(10, 10, sc)
		if len(g.characters) != 1 {
			t.Errorf("expected 1 character, got %d", len(g.characters))
		}
	})
	
	t.Run("SpawnAtMapEdges", func(t *testing.T) {
		g.characters = nil
		g.playableCharacter.X, g.playableCharacter.Y = 0, 0
		wm.spawnNPCAtMapEdges(sc)
		if len(g.characters) != 1 {
			t.Errorf("expected 1 character, got %d", len(g.characters))
		}
		// Should be roughly 30 units away
		c := g.characters[0]
		dist := c.X*c.X + c.Y*c.Y
		if dist < 25*25 {
			t.Errorf("spawned too close: dist squared %f", dist)
		}
	})
}

func TestWorldManager_NPCSpawnTimer_Cleanup(t *testing.T) {
	g := setupTestGame()
	wm := NewWorldManager(g)
	
	g.playableCharacter.X, g.playableCharacter.Y = 0, 0
	
	// Add an NPC far away
	farNpc := NewCharacter(1000, 1000, &EntityConfig{ID: "far"}, 1, false, nil)
	g.characters = []*Character{farNpc}
	
	g.npcSpawnTimer = 299
	wm.UpdateNPCSpawning()
	
	if len(g.characters) != 0 {
		t.Error("expected far NPC to be cleaned up")
	}
}
