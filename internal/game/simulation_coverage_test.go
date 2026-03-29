package game

import (
	"testing"
)

func TestActor_SimulationCoverage(t *testing.T) {
	ctx := NewTestContext()
	p := NewCharacter(0, 0, nil, 1, true, nil)
	p.Name = "Test Subject"
	p.State.MaxHealthPoints = 100
	p.State.HealthPoints = 100
	
	// Set some state to trigger branches
	p.State.Hunger = 50
	p.State.Thirst = 50
	p.State.Fatigue = 50
	p.State.Sanity = 50
	p.State.Arousal = 50
	p.State.Pain = 50
	p.State.Hygiene = 50
	p.State.IsSeptic = true
	
	p.State.Age.Current = 20
	p.State.Age.Max = 80
	p.State.Age.Rate = 1.0
	
	// Call SharedUpdate
	p.SharedUpdate(ctx)
	
	// Test specific internal methods
	p.updateNeeds(ctx)
	p.updateIllness(ctx)
	p.updateSanity(ctx)
	p.updateAge(ctx)
	p.updatePain(ctx)
	p.updateArousal(ctx)
	
	// Defecate/Miccionate
	p.State.BowelLevel = 99
	p.State.BladderLevel = 99
	p.updateNeeds(ctx) // This should trigger relief/soiling
}
