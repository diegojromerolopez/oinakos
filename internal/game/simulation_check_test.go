package game

import (
	"fmt"
	"testing"
)

func TestSimulation_CharacterStability(t *testing.T) {
	g := setupTestGame()
	g.settings.AISimulationMode = true
	g.settings.AIProvider = "native" // Use the rules-based provider
	
	// Ensure Oinakos has a reasonable starting state
	g.playableCharacter.X, g.playableCharacter.Y = 100, 100
	g.playableCharacter.State.HealthPoints = 1000
	g.playableCharacter.State.MaxHealthPoints = 1000
	
	// Add a food source nearby
	g.World.Obstacles = append(g.World.Obstacles, &Obstacle{
		ID: "apple_tree", X: 102, Y: 102, Alive: true,
		Archetype: &ObstacleArchetype{ID: "tree_apple", Type: TypeTree, Yield: "apple", Weight: 100.0, Destructible: true},
	})
	g.Registries.Objects.Objects["apple"] = &ObjectConfig{
		ID: "apple", Name: "Apple", Type: "consumable",
		Hunger: 50, Thirst: 10, Weight: 0.1,
	}

	// Mock AI initialization manually since initAIManager might be skipped in some test setups
	g.aiManager = NewAIManager(&NativeAIProvider{})

	fmt.Println("Running stability simulation for 10000 ticks...")
	for i := 0; i < 10000; i++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Game update failed at tick %d: %v", i, err)
		}
		
		if i % 1000 == 0 {
			pc := g.playableCharacter
			fmt.Printf("Tick %d: HP=%d, Hunger=%.1f, Thirst=%.1f, State=%s\n", 
				i, pc.State.HealthPoints, pc.State.Hunger, pc.State.Thirst, pc.ActionState.String())
		}
		
		if !g.playableCharacter.IsAlive() {
			t.Fatalf("Player character died at tick %d! Reason: %s", i, g.playableCharacter.GetDeathReason())
		}
	}
	fmt.Println("Stability simulation PASSED.")
}
