package game

import (
	"fmt"
	"testing"
)

func TestIntegration_VampireAging(t *testing.T) {
	g := setupTestGame()
	
	// Create a vampire character manually or from config
	// Here we use Boris who is a vampire
	p := g.playableCharacter
	p.Name = "Boris"
	p.State.Age = AgeState{Current: 100.0, Rate: 0.0}
	p.AgeTicks = 100.0 * float64(TicksPerYear)
	p.SyncStats(nil)

	initialAgeTicks := p.AgeTicks
	initialYears := p.State.Age.Current

	fmt.Printf("Initial Boris Age: %.2f years (Ticks: %.0f), Rate: %.1f\n", 
		p.State.Age.Current, p.AgeTicks, p.State.Age.Rate)

	// Simulate 1 day
	for i := 0; i < TicksPerDay; i++ {
		p.updateAge(nil)
	}

	fmt.Printf("After 1 day Boris Age: %.2f years (Ticks: %.0f)\n", 
		p.State.Age.Current, p.AgeTicks)

	if p.AgeTicks != initialAgeTicks {
		t.Errorf("Vampire aged! Expected ticks %.0f, got %.0f", initialAgeTicks, p.AgeTicks)
	}
	if p.State.Age.Current != initialYears {
		t.Errorf("Vampire age years changed! Expected %.2f, got %.2f", initialYears, p.State.Age.Current)
	}

	// Now check a normal peasant
	peasant := &Actor{
		State: State{
			Age: AgeState{Current: 20.0, Rate: 1.0},
		},
		AgeTicks: 20.0 * float64(TicksPerYear),
	}
	peasant.SyncStats(nil)
	
	initialPeasantTicks := peasant.AgeTicks
	for i := 0; i < TicksPerDay; i++ {
		peasant.updateAge(nil)
	}
	
	fmt.Printf("After 1 day Peasant Age: %.5f years (Ticks: %.0f)\n", 
		peasant.State.Age.Current, peasant.AgeTicks)

	if peasant.AgeTicks <= initialPeasantTicks {
		t.Errorf("Peasant did not age! Expected ticks > %.0f, got %.0f", initialPeasantTicks, peasant.AgeTicks)
	}
}
