package game

import (
	"fmt"
	"testing"
)

func TestLifeStage_Transitions(t *testing.T) {
	ctx := NewTestContext()
	
	// Setup archetypes for transitions
	stages := []string{StageBaby, StageKid, StageTeenager, StageAdult, StageElder}
	genders := []string{"male", "female"}
	
	for _, stage := range stages {
		for _, gender := range genders {
			id := fmt.Sprintf("archetypes/%s/%s", stage, gender)
			ctx.Registries.Archetypes.Archetypes[id] = &Archetype{
				ID: id,
				Name: fmt.Sprintf("%s %s", stage, gender),
				Gender: gender,
			}
		}
	}
	// Special cases for Adult professions
	ctx.Registries.Archetypes.Archetypes["archetypes/man_at_arms"] = &Archetype{ID: "archetypes/man_at_arms", Name: "Man-at-Arms", Gender: "male"}
	ctx.Registries.Archetypes.Archetypes["archetypes/peasant"] = &Archetype{ID: "archetypes/peasant", Name: "Peasant", Gender: "female"}

	ticksPY := float64(TicksPerYear)

	t.Run("Dead Actor does not age", func(t *testing.T) {
		a := &Actor{ActionState: ActorDead, AgeTicks: 10 * ticksPY}
		a.updateAge(ctx)
		// AgeTicks should remain same (ActionState checked at beginning)
		if a.AgeTicks != 10 * ticksPY {
			t.Errorf("Dead actor should not age")
		}
	})

	t.Run("Max Age Death", func(t *testing.T) {
		a := &Actor{
			Name: "Old Man",
			LifeStage: StageElder,
			ActionState: ActorIdle,
			State: State{
				Age: AgeState{Max: 100, Rate: 1.0},
			},
			AgeTicks: 101 * ticksPY,
		}
		a.updateAge(ctx)
		if a.ActionState != ActorDead {
			t.Errorf("Actor should be dead of old age")
		}
	})

	t.Run("Stage Transitions", func(t *testing.T) {
		scenarios := []struct {
			startStage string
			startAge   float64
			endStage   string
			gender     string
		}{
			{StageBaby, 1.1, StageKid, "male"},
			{StageKid, 12.1, StageTeenager, "female"},
			{StageTeenager, 18.1, StageAdult, "male"},
			{StageAdult, 65.1, StageElder, "female"},
		}

		for _, s := range scenarios {
			a := &Actor{
				Name: "Test Subject",
				LifeStage: s.startStage,
				AgeTicks: s.startAge * ticksPY,
				Config: &EntityConfig{Archetype: "some_id_" + s.gender, Gender: s.gender},
				State: State{Age: AgeState{Rate: 1.0}, HealthPoints: 1, MaxHealthPoints: 100},
			}
			if s.gender == "female" {
				a.Config.Archetype = "fake_female" // must contain female
			}

			a.updateAge(ctx)
			if a.LifeStage != s.endStage {
				t.Errorf("Expected stage %s, got %s for age %v", s.endStage, a.LifeStage, s.startAge)
			}
			
			// Verify archetype swap if it succeeded
			prefix := "archetypes/" + s.endStage + "/" + s.gender
			if s.endStage == StageAdult {
				if s.gender == "male" { prefix = "archetypes/man_at_arms" } else { prefix = "archetypes/peasant" }
			}
			if a.Config.ID != prefix {
				t.Errorf("Archetype not swapped to %s, got %s", prefix, a.Config.ID)
			}
			
			// HP should be reset to max
			if a.State.HealthPoints != a.State.MaxHealthPoints {
				t.Errorf("HP not reset to max after stage transition")
			}
		}
	})

	t.Run("Elder Natural Death Chance", func(t *testing.T) {
		a := &Actor{
			LifeStage: StageElder,
			ActionState: ActorIdle,
			AgeTicks: 86 * ticksPY,
			State: State{Age: AgeState{Rate: 1.0}},
		}
		// Chance increases after 85. 
		// (86-85)/15 = 1/15.
		// per-tick chance is very low. Just checking coverage.
		a.updateAge(ctx)
	})

	t.Run("Lookup failure does not crash", func(t *testing.T) {
		// Temporarily wipe registries
		oldArchs := ctx.Registries.Archetypes.Archetypes
		ctx.Registries.Archetypes.Archetypes = make(map[string]*Archetype)
		
		a := &Actor{
			LifeStage: StageBaby,
			AgeTicks: 1.5 * ticksPY,
			Config: &EntityConfig{Archetype: "male"},
			State: State{Age: AgeState{Rate: 1.0}},
		}
		a.updateAge(ctx)
		// Stage should change, but Config should remain same if lookup fails
		if a.LifeStage != StageKid {
			t.Errorf("Stage should still change even if lookup fails")
		}
		
		// Restore
		ctx.Registries.Archetypes.Archetypes = oldArchs
	})
}

func TestAging_UntilDeath(t *testing.T) {
	ctx := NewTestContext()
	// TicksPerYear is 829,440. 
	// To avoid 800k iterations, we'll set the aging rate to reach 100 quickly.
	a := &Actor{
		Name: "Mortal Man",
		LifeStage: StageAdult,
		ActionState: ActorIdle,
		State: State{
			Age: AgeState{Max: 100, Rate: float64(TicksPerYear) / 10.0},
		},
		AgeTicks: 99.0 * float64(TicksPerYear),
	}
	
	initialAge := a.AgeTicks / float64(TicksPerYear)
	
	// Simulate 11 steps. Each step is 1/10th of a year.
	// 99.0 + 11*0.1 = 100.1
	for i := 0; i < 15 && a.IsAlive(); i++ {
		a.updateAge(ctx)
	}
	
	finalAge := a.AgeTicks / float64(TicksPerYear)
	if a.IsAlive() {
		t.Errorf("Mortal Man should have died of old age! Started: %.1f, Ended: %.1f, Max: %.1f", 
			initialAge, finalAge, a.State.Age.Max)
	} else {
		fmt.Printf("Mortal Man died correctly at age %.2f\n", finalAge)
	}
}
