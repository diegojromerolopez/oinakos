package game

import (
	"math/rand"
	"testing"
)

func TestIllness_FatigueDecay(t *testing.T) {
	g := setupTestGame()
	ctx := g.GetContext()
	
	a := &Actor{
		State: State{
			Fatigue:      0.0,
			HealthPoints: 100,
		},
		FluTicks: 100,
	}
	
	// Two update cycles
	a.SharedUpdate(ctx)
	a.SharedUpdate(ctx)
	
	// Normal decay: 0.003
	// Flu decay: 0.02
	// Total decay: 0.023 per tick
	// Total for 2 ticks: 0.046
	if a.State.Fatigue <= 0 {
		t.Errorf("Expected Fatigue increase from illness, got %.3f", a.State.Fatigue)
	}
}

func TestIllness_CriticalImpact(t *testing.T) {
	g := setupTestGame()
	ctx := g.GetContext()
	
	a := &Actor{
		State: State{
			Fatigue:         96.0,
			MaxHealthPoints: 100,
		},
		FluTicks: 1000,
	}
	a.InitBodyStatus()
	
	// Should lose health after 180 ticks
	for i := 0; i < 181; i++ {
		a.Tick++
		a.SharedUpdate(ctx)
	}
	
	if a.State.HealthPoints >= 100 {
		t.Errorf("Expected Health loss from critical illness, got %d", a.State.HealthPoints)
	}
}

func TestIllness_Infection(t *testing.T) {
	rand.Seed(1) // Deterministic test
	g := setupTestGame()
	ctx := g.GetContext()
	
	sick := NewCharacter(1, 1, nil, 1, false, nil)
	sick.ID = "sick_actor"
	sick.FluTicks = 1000
	
	healthy := NewCharacter(1.5, 1.5, nil, 1, false, nil)
	healthy.ID = "healthy_actor"
	healthy.FluTicks = 0
	
	g.characters = []*Character{healthy}
	ctx.World.Characters = g.characters
	
	// Trigger contagion check (every 600 ticks)
	sick.Tick = 600
	sick.SharedUpdate(ctx)
	
	// Contagion prob: 0.12. With seed 1, this should eventually trigger if we repeat enough or fix seed.
	// Actually, let's force the spread for the test
	for i := 0; i < 100; i++ {
		sick.Tick = 600
		sick.SharedUpdate(ctx)
		if healthy.FluTicks > 0 {
			break
		}
	}
	
	if healthy.FluTicks == 0 {
		t.Error("Expected healthy NPC to be infected by proximity")
	}
}

func TestIllness_CureMedicine(t *testing.T) {
	g := setupTestGame()
	ctx := g.GetContext()
	
	a := &Actor{FluTicks: 1000}
	medicine := &ObjectConfig{ID: "medicinal_herbs", ClearSick: true}
	med_it := NewItemInstance("med_0", medicine, 0, 0)
	
	if !a.ConsumeItem(med_it, ctx) {
		t.Error("Expected ConsumeItem to return true for medicinal_herbs")
	}
	
	if a.FluTicks != 0 {
		t.Errorf("Expected FluTicks to be cleared, got %d", a.FluTicks)
	}
}
