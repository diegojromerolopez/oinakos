package game

import (
	"fmt"
	"sync/atomic"
	"testing"
)

func TestIntegration_BiologicalCycle(t *testing.T) {
	g := setupTestGame()
	g.World.State.Temperature = 25.0
	
	// Create a test character "Stultus"
	p := g.playableCharacter
	p.Name = "Stultus"
	p.PrimaryAttributes = PrimaryAttributes{Strength: 50, Dexterity: 50, Health: 50, Intellect: 50, Wisdom: 50}
	p.TemporalState = TemporalState{
		HealthPoints: 1000, MaxHealthPoints: 1000,
		Hunger: 0, Thirst: 0, Fatigue: 0,
		Hygiene: 100,
		Miccionate: 0, Defecate: 0,
		Pain: 0,
	}
	p.SyncStats(nil)

	// Create a "latrine" obstacle nearby
	latrineConfig := &ObstacleArchetype{
		ID: "latrine", Name: "Latrine",
		Type: TypeBuilding,
		Actions: []ObstacleActionConfig{
			{Type: ActionAlleviate, Amount: 100, RequiresInteraction: true},
		},
	}
	latrine := NewObstacle("latrine_0", 1.0, 1.0, latrineConfig)
	g.World.Obstacles = append(g.World.Obstacles, latrine)
	g.obstacles = g.World.Obstacles

	atomic.StoreInt32(&g.LoadingProgress, 1000)

	// Phase 1: Eat and Drink until needs are high
	fmt.Println("--- Phase 1: Consumption ---")
	for i := 0; i < 500; i++ {
		p.State = ActorEating
		g.Update()
		if p.TemporalState.Defecate > 50 {
			break
		}
	}
	fmt.Printf("After eating: Hunger=%.1f, Defecate=%.1f, Pain=%.1f\n", 
		p.TemporalState.Hunger, p.TemporalState.Defecate, p.TemporalState.Pain)

	// Phase 2: Wait until Defecate hits 100 and Causes Pain
	fmt.Println("--- Phase 2: Waiting for urgency ---")
	for i := 0; i < 200000; i++ { 
		p.TemporalState.Defecate += 0.01 // Fast-forward for test
		p.TemporalState.Miccionate += 0.005
		g.Update()
		if p.TemporalState.Defecate >= 90 {
			break
		}
	}
	p.TemporalState.Defecate = 90
	p.TemporalState.Miccionate = 90

	
	// Hold it for a while to increase pain
	fmt.Println("--- Phase 3: Holding it (Pain should increase) ---")
	for i := 0; i < 5000; i++ {
		g.Update()
	}
	
	fmt.Printf("Urgent state: Defecate=%.1f, Miccionate=%.1f, Pain=%.1f\n", 
		p.TemporalState.Defecate, p.TemporalState.Miccionate, p.TemporalState.Pain)
	
	if p.TemporalState.Pain <= 0 {
		t.Errorf("Expected pain from urgent need, got %.1f", p.TemporalState.Pain)
	}

	// Phase 4: Use the latrine
	fmt.Println("--- Phase 4: Using Latrine ---")
	p.X, p.Y = 1.0, 1.0 // Move to latrine
	
	for i := 0; i < 1000; i++ {
		p.State = ActorRelieving // IMPORTANT: Set state EVERY tick to avoid player-input Reset to Idle
		g.Update()
		if p.TemporalState.Defecate <= 0 && p.TemporalState.Miccionate <= 0 {
			break
		}
	}

	fmt.Printf("After Relief: Defecate=%.1f, Miccionate=%.1f, Pain=%.1f, Hygiene=%.1f\n", 
		p.TemporalState.Defecate, p.TemporalState.Miccionate, p.TemporalState.Pain, p.TemporalState.Hygiene)

	if p.TemporalState.Defecate > 0.1 || p.TemporalState.Miccionate > 0.1 {
		t.Errorf("Expected ~0 defecate/miccionate, got D=%.1f, M=%.1f", p.TemporalState.Defecate, p.TemporalState.Miccionate)
	}
	if p.TemporalState.Pain > 0 {
		t.Errorf("Expected 0 pain after relief, got %.1f", p.TemporalState.Pain)
	}
}
