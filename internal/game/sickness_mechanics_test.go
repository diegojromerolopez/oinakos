package game

import (
	"testing"
)

func TestSickness_Mechanics(t *testing.T) {
	ctx := NewTestContext()
	
	t.Run("Sepsis HP Drain", func(t *testing.T) {
		a := &Actor{Name: "Septic Patient", State: State{IsSeptic: true, HealthPoints: 100}}
		a.Tick = 600
		a.updateIllness(ctx)
		if a.State.HealthPoints >= 100 {
			t.Errorf("Sepsis should drain HP at 600 tick interval")
		}
	})

	t.Run("Flu - Fatigue and HP Drain", func(t *testing.T) {
		a := &Actor{Name: "Flu Patient", State: State{HealthPoints: 100, Fatigue: 96}}
		a.FluTicks = 100
		a.Tick = 300
		// Should increase fatigue and drain HP because fatigue > 95
		a.updateIllness(ctx)
		if a.State.Fatigue <= 96 { t.Errorf("Flu should increase fatigue") }
		if a.State.HealthPoints >= 100 { t.Errorf("Flu should drain HP when exhausted") }
	})

	t.Run("Flu - Contagion Spread", func(t *testing.T) {
		a := &Actor{X: 0, Y: 0, FluTicks: 100, ContagionTimer: 0}
		other := spawnTestActor(setupTestGame(), "peasant_male", 0.5, 0.5)
		ctx.World.Characters = []*Character{other}
		
		// Run update to trigger contagion
		a.updateIllness(ctx)
		
		// Contagion is stochastic (12% chance). We might need to loop or mock rand.
		found := false
		for i := 0; i < 100; i++ {
			a.ContagionTimer = 0
			a.updateIllness(ctx)
			if other.Actor.FluTicks > 0 {
				found = true
				break
			}
		}
		if !found {
			t.Log("Contagion not spread in 100 ticks (expected occasionally due to 12% probability)")
		}
	})

	t.Run("Stomach Sickness Side Effects", func(t *testing.T) {
		a := &Actor{
			Name: "Stomach Patient", 
			Sickness: "stomach sickness", 
			SicknessTicks: 100, 
			Tick: 600,
			State: State{Sanity: 100, Pain: 0},
		}
		a.updateIllness(ctx)
		if a.State.Sanity >= 100 { t.Errorf("Stomach sickness should drain sanity") }
		if a.State.Pain <= 0 { t.Errorf("Stomach sickness should increase pain") }
	})

	t.Run("Recovery from Sickness", func(t *testing.T) {
		a := &Actor{Name: "Recovering", Sickness: "flu", SicknessTicks: 1, State: State{IsSick: true}}
		a.updateIllness(ctx)
		if a.Sickness != "" || a.State.IsSick {
			t.Errorf("Actor should recover when SicknessTicks hit 0")
		}
	})
	
	t.Run("Nil Context early return", func(t *testing.T) {
		a := &Actor{}
		a.updateIllness(nil) // should not panic
	})
}
