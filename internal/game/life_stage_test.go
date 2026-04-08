package game

import (
	"fmt"
	"testing"
)

func TestLifeStage_Transitions(t *testing.T) {
	setup := func() *SystemContext {
		ctx := NewTestContext()
		// Setup archetypes for transitions
		stages := []string{StageBaby, StageKid, StageTeenager, StageAdult, StageElder}
		genders := []string{"male", "female"}
		for _, stage := range stages {
			for _, gender := range genders {
				id := "archetypes/" + stage + "/" + gender
				ctx.Registries.Archetypes.Archetypes[id] = &Archetype{ID: id, Name: stage + " " + gender, Gender: gender}
			}
		}
		ctx.Registries.Archetypes.Archetypes["archetypes/man_at_arms"] = &Archetype{ID: "archetypes/man_at_arms", Name: "Man-at-Arms", Gender: "male"}
		ctx.Registries.Archetypes.Archetypes["archetypes/peasant"] = &Archetype{ID: "archetypes/peasant", Name: "Peasant", Gender: "female"}
		return ctx
	}

	t.Run("Dead Actor does not age", func(t *testing.T) {
		ctx := setup()
		a := &Actor{ActionState: ActorDead, AgeTicks: 10 * float64(TicksPerYear)}
		a.updateAge(ctx)
		if a.AgeTicks != 10 * float64(TicksPerYear) { t.Errorf("Dead actor should not age") }
	})

	t.Run("Max Age Death", func(t *testing.T) {
		ctx := setup()
		a := &Actor{Name: "Old Man", LifeStage: StageElder, ActionState: ActorIdle, State: State{Age: AgeState{Max: 100, Rate: 1.0}}, AgeTicks: 101 * float64(TicksPerYear)}
		a.updateAge(ctx)
		if a.ActionState != ActorDead { t.Errorf("Actor should be dead of old age") }
	})

	t.Run("Stage Transitions", func(t *testing.T) {
		scenarios := []struct {
			startStage string; startAge float64; endStage string; gender string
		}{
			{StageBaby, 1.1, StageKid, "male"},
			{StageKid, 12.1, StageTeenager, "female"},
			{StageTeenager, 18.1, StageAdult, "male"},
			{StageAdult, 65.1, StageElder, "female"},
		}

		for _, s := range scenarios {
			ctx := setup()
			a := &Actor{
				Name: "Test Subject", LifeStage: s.startStage, AgeTicks: s.startAge * float64(TicksPerYear),
				Config: &EntityConfig{Archetype: "some_id_" + s.gender, Gender: s.gender},
				State: State{Age: AgeState{Rate: 1.0}, HealthPoints: 1, MaxHealthPoints: 100},
				MortalityChecked: true,
			}
			if s.gender == "male" && s.endStage == StageAdult { a.PrimaryAttributes.Strength = 65 }
			if s.gender == "female" { a.Config.Archetype = "fake_female" }
			a.updateAge(ctx)
			if a.LifeStage != s.endStage { t.Errorf("Expected stage %s, got %s", s.endStage, a.LifeStage) }
			prefix := "archetypes/" + s.endStage + "/" + s.gender
			if s.endStage == StageAdult { if s.gender == "male" { prefix = "archetypes/man_at_arms" } else { prefix = "archetypes/peasant" } }
			if a.Config == nil || a.Config.ID != prefix { t.Errorf("Scenario %v -> %v: Archetype not swapped to %s, got %v", s.startStage, s.endStage, prefix, a.Config.ID) }
		}
	})

	t.Run("Lookup failure does not crash", func(t *testing.T) {
		ctx := setup()
		ctx.Registries.Archetypes.Archetypes = make(map[string]*Archetype)
		a := &Actor{LifeStage: StageBaby, AgeTicks: 1.5 * float64(TicksPerYear), Config: &EntityConfig{Archetype: "male"}, State: State{Age: AgeState{Rate: 1.0}}, MortalityChecked: true}
		a.updateAge(ctx)
		if a.LifeStage != StageKid { t.Errorf("Stage should still change even if lookup fails, got: %q, wanted: %q", a.LifeStage, StageKid) }
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
