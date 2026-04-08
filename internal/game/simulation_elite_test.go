package game

import (
	"testing"
)

func TestPregnancyPenalties(t *testing.T) {
	mother := &Character{
		Actor: Actor{
			Name: "Expectant",
			PrimaryAttributes: PrimaryAttributes{ Strength: 100, Dexterity: 100, Intellect: 100, Wisdom: 100 },
			AgeTicks: 25.0 * float64(TicksPerYear), 
			Config: &EntityConfig{ 
				Name: "Human", 
				Stats: EntityStatsConfig{ Age: AgeConfig{ Rate: FloatInterval{Min: 1.0, Max: 1.0} } },
			},
		},
	}
	
	mother.SyncStats(nil)
	baseSpeed := mother.Speed
	
	// Apply Pregnancy
	mother.IsPregnant = true
	mother.SyncStats(nil)
	
	// Verify Physical Penalties (Speed reduction)
	if mother.Speed >= baseSpeed { t.Errorf("Expected speed reduction for pregnant character.") }
	
	yield := mother.GetAbilityYield("craft")
	if yield == 0 { t.Errorf("Expected non-zero craft yield") }
}

func TestBreedingConstraints(t *testing.T) {
	ctx := &SystemContext{
		World: &World{ Characters: []*Character{} },
		Registries: &RegistryContainer{ Archetypes: &ArchetypeRegistry{ Archetypes: make(map[string]*EntityConfig) }, Objects: &ObjectRegistry{} },
		Settings: &Settings{ AdultMode: true },
	}
	
	m := &Character{ Actor: Actor{ Name: "Male", Config: &EntityConfig{ Gender: "male" } } }
	f := &Character{ Actor: Actor{ Name: "Female", Config: &EntityConfig{ Gender: "female" } } }
	t1 := &Character{ Actor: Actor{ Name: "TransFemale", Config: &EntityConfig{ Gender: "female" }, IsTransexual: true } }
	
	ctx.World.Characters = append(ctx.World.Characters, m, f, t1)
	
	// Case 1: Trans mother (sterile)
	m.Actor.haveSex(ctx, &t1.Actor, "vaginal")
	if t1.IsPregnant { t.Errorf("Transgender characters should be sterile in the biological procreation engine.") }
}

func TestGriefCascade(t *testing.T) {
	ctx := &SystemContext{
		World: &World{ Characters: []*Character{} },
	}
	
	a1 := &Character{ Actor: Actor{ ID: "A1", Name: "Partner 1", Relationships: make(map[string]float64) } }
	a2 := &Character{ Actor: Actor{ ID: "A2", Name: "Partner 2", Relationships: make(map[string]float64) } }
	
	a2.Actor.Relationships["A1"] = 100.0 // Strong bond
	ctx.World.Characters = append(ctx.World.Characters, a1, a2)
	
	a1.ActionState = ActorDead
	a1.TriggerSocialCascade(ctx)
	
	if a2.GriefTicks == 0 {
		t.Errorf("Partner should experience grief cascade upon death of bonded partner.")
	}
}

func TestDecayAndMiasma(t *testing.T) {
	ctx := &SystemContext{
		World: &World{ Characters: []*Character{} },
	}
	
	corpse := &Character{ Actor: Actor{ ID: "C1", ActionState: ActorDead, X: 0, Y: 0 } }
	survivor := &Character{ Actor: Actor{ ID: "S1", X: 1, Y: 1, State: State{ Sanity: 100, Hygiene: 100 } } }
	
	ctx.World.Characters = append(ctx.World.Characters, corpse, survivor)
	
	// Skip day to rot
	corpse.RotTicks = TicksPerDay + 1
	corpse.updateDecay(ctx)
	
	if survivor.State.Sanity >= 100 {
		t.Errorf("Survivor should suffer sanity loss from nearby rotting miasma.")
	}
}

func TestAgeTransitions(t *testing.T) {
	ctx := &SystemContext{
		Registries: &RegistryContainer{ 
			Archetypes: &ArchetypeRegistry{ Archetypes: make(map[string]*EntityConfig) },
			Objects: &ObjectRegistry{},
		},
	}
	ctx.Registries.Archetypes.Archetypes["archetypes/kid/male"] = &EntityConfig{ID: "kid"}
	
	child := &Character{
		Actor: Actor{
			LifeStage: StageBaby,
			AgeTicks: 0,
			MortalityChecked: true, // skip infant mortality roll — we're testing stage transitions only
			State: State{ Age: AgeState{ Rate: 1.0 } },
			Config: &EntityConfig{ ID: "baby", Archetype: "baby/male" },
		},
	}
	
	// Advance 1 year
	child.AgeTicks = float64(TicksPerYear) + 100
	child.updateAge(ctx)
	
	if child.LifeStage != StageKid {
		t.Errorf("Child should transition to Kid stage after 1 year.")
	}
}
