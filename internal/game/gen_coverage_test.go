package game

import (
	"testing"
)

func TestWorldGenerator_Coverage(t *testing.T) {
	g := setupTestGame()
	wg := NewWorldGenerator(g, 42)
	
	// 1. GenerateVillage
	// Need some archetypes in registries to avoid empty village
	g.obstacleRegistry.Archetypes["shop_tent"] = &ObstacleArchetype{ID: "shop_tent", Name: "Tent"}
	g.archetypeRegistry.Archetypes["vampire_male"] = &Archetype{ID: "vampire_male", Name: "Vampire"}
	g.obstacleRegistry.Archetypes["campfire"] = &ObstacleArchetype{ID: "campfire", Name: "Fire"}

	wg.GenerateVillage(50, 50, 20)
	
	// 2. SeedResources
	g.Registries.Objects.Objects["wood"] = &ObjectConfig{ID: "wood", Name: "Wood"}
	wg.SeedResources(100, 100, "wood", 5)
}
