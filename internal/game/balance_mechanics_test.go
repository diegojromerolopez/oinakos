package game

import (
	"testing"
	"math"
)

func TestBalance_ConsensualSexRewards(t *testing.T) {
	ctx := NewTestContext()
	
	p := NewCharacter(0, 0, nil, 1, false, nil)
	p.Name = "Romeo"; p.Config = &EntityConfig{Gender: "male"}
	
	mate := NewCharacter(0.1, 0.1, nil, 1, false, nil)
	mate.Name = "Juliet"; mate.Config = &EntityConfig{Gender: "female"}
	
	ctx.World.Characters = append(ctx.World.Characters, p, mate)

	// Set starting state
	p.State.Sanity = 50.0
	mate.State.Sanity = 50.0
	p.State.Arousal = 60.0
	mate.State.Arousal = 60.0

	// Consensual mating (Both willing)
	p.Actor.haveSex(ctx, &mate.Actor, "vaginal")

	if p.State.Sanity <= 50.0 {
		t.Errorf("Expected Sanity reward for Romeo, got %.2f", p.State.Sanity)
	}
	if mate.State.Sanity <= 50.0 {
		t.Errorf("Expected Sanity reward for Juliet, got %.2f", mate.State.Sanity)
	}
	if p.State.Arousal != 0 {
		t.Errorf("Expected Arousal reset for Romeo, got %.2f", p.State.Arousal)
	}
}

func TestBalance_CriminalPredationRewards(t *testing.T) {
	ctx := NewTestContext()
	ctx.Settings.AdultMode = true
	
	criminal := NewCharacter(0, 0, nil, 1, false, nil)
	criminal.Name = "Marca"; criminal.Behavior = BehaviorCriminal
	criminal.Config = &EntityConfig{Gender: "male", ID: "criminal"}
	
	victim := NewCharacter(0.1, 0.1, nil, 1, false, nil)
	victim.Name = "Peasant"; victim.Config = &EntityConfig{Gender: "female", ID: "peasant"}
	
	ctx.World.Characters = append(ctx.World.Characters, criminal, victim)

	// Set starting state
	criminal.State.Sanity = 50.0
	victim.State.Sanity = 80.0
	criminal.State.Fatigue = 50.0

	// Forced mating (Violent/Criminal)
	criminal.Actor.haveSex(ctx, &victim.Actor, "anal")

	if criminal.State.Sanity <= 50.0 {
		t.Errorf("Expected predator Sanity reward, got %.2f", criminal.State.Sanity)
	}
	if criminal.State.Fatigue != 50.0 {
		t.Errorf("Expected predator Fatigue to be 50.0, got %.2f", criminal.State.Fatigue)
	}
	if victim.State.Sanity >= 80.0 {
		t.Errorf("Expected victim Sanity loss, got %.2f", victim.State.Sanity)
	}
}

func TestBalance_MourningSustainability(t *testing.T) {
	ctx := NewTestContext()
	p := NewCharacter(0, 0, nil, 1, false, nil)
	p.State.Sanity = 100.0
	p.GriefTicks = TicksPerDay // 17,280 ticks of mourning
	
	// Set state to Walking to avoid Leisure sanity gain
	p.ActionState = ActorWalking
	
	// Update for 1 hour (720 ticks)
	for i := 0; i < 720; i++ {
		p.SharedUpdate(ctx)
		p.Tick++
	}

	// 0.001 * 720 = 0.72 Sanity loss
	expected := 100.0 - 0.72
	if p.State.Sanity >= 100.0 {
		t.Errorf("Sanity failed to decrease during mourning (still %.2f)", p.State.Sanity)
	}
	if math.Abs(p.State.Sanity-expected) > 0.1 {
		t.Errorf("Expected Sanity after 1hr mourning roughly %.2f, got %.2f", expected, p.State.Sanity)
	}
}

func TestBalance_ArousalCapAndDecay(t *testing.T) {
	ctx := NewTestContext()
	ctx.World.State.Season = SeasonWinter // Avoid spring bonus

	p := NewCharacter(0, 0, nil, 1, false, nil)
	
	// Test Cap
	p.State.Arousal = 150.0
	p.updateArousal(ctx)
	if p.State.Arousal > 100.0 {
		t.Errorf("Arousal failed to cap at 100, got %.2f", p.State.Arousal)
	}

	// Test Decay when miserable (Sanity 0)
	p.State.Sanity = 0.0
	p.State.Arousal = 50.0
	// 0.000192 growth, decay 0.00005 * 4 = 0.0002
	// Net: -0.000008 loss
	p.updateArousal(ctx)
	if p.State.Arousal >= 50.0 {
		t.Errorf("Arousal should decay when insane, but it is %.6f (Growth: %.6f)", p.State.Arousal, 0.000192)
	}
}

func TestBalance_CriminalRobbery(t *testing.T) {
	ctx := NewTestContext()
	
	criminal := NewCharacter(0, 0, nil, 1, false, nil)
	// IMPORTANT: Criminals suffer guilt (-20 Sanity) if their victim dies.
	// To test robbery satisfaction, we need a victim that SURVIVES.
	criminal.Behavior = BehaviorCriminal
	criminal.State.Sanity = 50.0
	criminal.Denarii = 0
	
	victim := NewCharacter(1, 1, nil, 1, false, nil)
	victim.Denarii = 500
	victim.State.HealthPoints = 5000 // Huge HP to survive the "beating"
	victim.State.MaxHealthPoints = 5000
	
	// Force a hit that triggers robbery (25% chance in hitCharacter)
	success := false
	for i := 0; i < 100; i++ {
		oldDenarii := criminal.Denarii
		criminal.hitCharacter(&victim.Actor, AttackPunch, ctx)
		if criminal.Denarii > oldDenarii { 
			success = true
			break 
		}
	}

	if !success {
		t.Errorf("Criminal failed to rob any denarii in 100 hits")
	}
	if criminal.State.Sanity <= 50.0 {
		t.Errorf("Criminal failed to gain sanity from robbery (Sanity: %.2f)", criminal.State.Sanity)
	}
}

func TestBalance_MetabolicDecay(t *testing.T) {
	ctx := NewTestContext()
	
	// Character with LOW Health (0)
	pLow := NewCharacter(0, 0, nil, 1, false, nil)
	pLow.PrimaryAttributes.Health = 0
	pLow.PrimaryAttributes.Strength = 0
	
	// Character with HIGH Health (100)
	pHigh := NewCharacter(0, 0, nil, 1, false, nil)
	pHigh.PrimaryAttributes.Health = 100
	pHigh.PrimaryAttributes.Strength = 100
	
	pLow.Actor.SyncStats(nil)
	pHigh.Actor.SyncStats(nil)

	// Update both for 1,000 ticks (~1.4 hours) - Fewer ticks to avoid hitting the cap
	for i := 0; i < 1000; i++ {
		pLow.updateNeeds(ctx)
		pHigh.updateNeeds(ctx)
	}

	// High Health should have significantly lower hunger/thirst/fatigue
	if pHigh.State.Hunger >= pLow.State.Hunger {
		t.Errorf("High Health should have lower hunger than Low Health (H: %.2f vs L: %.2f)", pHigh.State.Hunger, pLow.State.Hunger)
	}
	if pHigh.State.Thirst >= pLow.State.Thirst {
		t.Errorf("High Health should have lower thirst than Low Health (H: %.2f vs L: %.2f)", pHigh.State.Thirst, pLow.State.Thirst)
	}
	
	// Verify fatigue recovery in rest
	pLow.State.Fatigue = 100.0
	pLow.ActionState = ActorResting
	for i := 0; i < 1000; i++ {
		pLow.updateNeeds(ctx)
	}
	if pLow.State.Fatigue >= 100.0 {
		t.Errorf("Fatigue failed to recover during resting")
	}
}
