package game

import (
	"testing"
)

func TestGame_Occlusion(t *testing.T) {
	g := setupTestGame()
	mc := g.playableCharacter
	mc.X, mc.Y = 0, 0
	
	// Large tree north of player
	treeArch := &ObstacleArchetype{
		ID: "tree",
		Type: "tree",
		Footprint: []FootprintPoint{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}},
	}
	tree := &Obstacle{
		X: -1, Y: -1, // Cartesian North (Isometric Up)
		Archetype: treeArch,
		Alive: true,
	}
	g.obstacles = []*Obstacle{tree}
	
	g.updateOcclusion()
	
	// Player at 0,0, tree at -1,-1. SortY of player is 0, sortY of tree is -2.
	// Player is SOUTH of tree.
	if mc.IsOccluded {
		t.Errorf("Player should NOT be occluded by a tree to their north. sortY player: %f, tree: %f", GetActorSortY(&mc.Actor), GetObstacleSortY(tree))
	}
	
	// Tree at 2,2 (Cartesian South)
	tree.X, tree.Y = 2, 2
	g.updateOcclusion()
	
	// Player 0,0, tree 2,2. SortY player 0, tree 4.
	// Player is NORTH of tree. Tree is closer to screen.
	// But it only occludes if current position IS COVERED.
	// Isometric 0,0 -> 0,0
	// Obstacle at 2,2 (Cartesian) -> Isometric (2-2), (2+2)*0.5 -> 0, 2
	// Point 0,0 is covered if within obstacle bounds?
	// Obstacles render from their base up.
}

func TestGame_Unloading(t *testing.T) {
	g := setupTestGame()
	mc := g.playableCharacter
	mc.Health = 100
	mc.MaxHealth = 100
	mc.X, mc.Y = 0, 0
	
	// Set dummy inventory
	it1 := NewItemInstance("wood", &ObjectConfig{ID: "wood", Name: "Wood"}, 0, 0)
	it2 := NewItemInstance("iron_ore", &ObjectConfig{ID: "iron_ore", Name: "Iron Ore"}, 0, 0)
	it3 := NewItemInstance("sword", &ObjectConfig{ID: "sword", Name: "Sword"}, 0, 0)
	mc.Inventory = []*ItemInstance{it1, it2, it3}
	
	// Test without warehouse
	g.tryUnloading()
	if len(mc.Inventory) != 3 {
		t.Errorf("Should not unload without warehouse. got %d", len(mc.Inventory))
	}
	
	// Add warehouse
	whArch := &ObstacleArchetype{ID: "warehouse", Name: "Warehouse"}
	warehouse := &Obstacle{X: 0.5, Y: 0.5, Archetype: whArch, Alive: true}
	g.obstacles = []*Obstacle{warehouse}
	g.World.Obstacles = g.obstacles
	
	g.tryUnloading()
	
	if len(mc.Inventory) != 1 {
		for i, item := range mc.Inventory {
			t.Logf("Inventory[%d]: %s", i, item.Config.ID)
		}
		t.Errorf("Inventory not unloaded correctly. Expected 1 (sword), got %d", len(mc.Inventory))
	}
	
	if mc.Inventory[0].Config.ID != "sword" {
		t.Errorf("Wrong item left in inventory. got %s", mc.Inventory[0].Config.ID)
	}
}

func TestGame_StuckRecovery(t *testing.T) {
	g := setupTestGame()
	mc := g.playableCharacter
	mc.X, mc.Y = 0, 0
	
	// Put character inside building
	bArch := &ObstacleArchetype{
		ID: "building",
		Footprint: []FootprintPoint{{X: -0.1, Y: -0.1}, {X: 0.1, Y: -0.1}, {X: 0.1, Y: 0.1}, {X: -0.1, Y: 0.1}},
	}
	g.obstacles = []*Obstacle{{X: 0, Y: 0, Archetype: bArch, Alive: true}}
	
	g.ensurePlayerNotStuck()
	
	// Should have moved out
	inside := mc.checkCollisionAt(mc.X, mc.Y, g.obstacles)
	if inside {
		t.Errorf("ensurePlayerNotStuck failed to move player out of obstacle")
	}
}
