package game

import (
	"testing"
)

func TestGlobalAgingFix(t *testing.T) {
	// 1. Create a config without age defined
	config := &EntityConfig{ID: "test_peasant", Name: "Test Peasant"}
	
	// 2. Create character
	c := NewCharacter(0, 0, config, 1, false, nil)
	
	// 3. Verify aging rate is 1.0 (default) and not 0.0
	if c.State.Age.Rate != 1.0 {
		t.Errorf("Expected default Age.Rate 1.0, got %.2f", c.State.Age.Rate)
	}
}

func TestGestationFix(t *testing.T) {
	// Setup world
	arch := &EntityConfig{ID: "parent", Gender: "female"}
	mate := &EntityConfig{ID: "mate", Gender: "male"}
	
	mother := NewCharacter(0, 0, arch, 1, false, nil)
	mother.AgeTicks = 25.0 * float64(TicksPerYear)
	mother.State.Age.Current = 25.0
	
	father := NewCharacter(0, 0, mate, 1, false, nil)
	father.AgeTicks = 25.0 * float64(TicksPerYear)
	father.State.Age.Current = 25.0

	ctx := NewTestContext()

	// Human pregnancy
	mother.mate(ctx, &father.Actor, "vaginal")
	
	if mother.IsPregnant {
		expectedGestation := int(TicksPerMonth * 9)
		if mother.GestationTicks != expectedGestation {
			t.Errorf("Expected human gestation %d ticks, got %d", expectedGestation, mother.GestationTicks)
		}
	} else {
		// Try again if random chance failed
		for i := 0; i < 100 && !mother.IsPregnant; i++ {
			mother.mate(ctx, &father.Actor, "vaginal")
		}
		if !mother.IsPregnant {
			t.Fatalf("Failed to trigger pregnancy in 100 attempts")
		}
		expectedGestation := int(TicksPerMonth * 9)
		if mother.GestationTicks != expectedGestation {
			t.Errorf("Expected human gestation %d ticks, got %d", expectedGestation, mother.GestationTicks)
		}
	}
}

func TestInfantMortalityRoll(t *testing.T) {
	// Create a baby
	arch := &EntityConfig{ID: "baby", Archetype: "baby"}
	baby := NewCharacter(0, 0, arch, 1, false, nil)
	baby.LifeStage = StageBaby
	
	ctx := NewTestContext()
	
	// 1. Before 6 months (AgeTicks < TicksPerYear/2)
	baby.AgeTicks = (float64(TicksPerYear) / 2.0) - 100
	baby.updateAge(ctx)
	
	if baby.MortalityChecked {
		t.Errorf("Mortality checked too early")
	}
	
	// 2. After 6 months
	baby.AgeTicks = (float64(TicksPerYear) / 2.0) + 100
	
	// We might need multiple attempts to see a 'death' or 'pass' if we want to verify the roll
	// But just checking it 'checks' is enough for logic verification
	baby.updateAge(ctx)
	
	if !baby.MortalityChecked {
		t.Errorf("Mortality check failed to trigger at 6 months")
	}
}

func TestDeathTrackingDemographics(t *testing.T) {
	ctx := NewTestContext()
	
	// Natural death (hunger)
	victim1 := NewCharacter(0, 0, nil, 1, false, nil)
	victim1.die(nil, ctx)
	
	if ctx.World.Demographics.DeathsNatural != 1 {
		t.Errorf("Expected 1 natural death, got %d", ctx.World.Demographics.DeathsNatural)
	}
	
	// Violent death (attacker)
	victim2 := NewCharacter(0, 0, nil, 1, false, nil)
	attacker := NewCharacter(0, 0, nil, 1, false, nil)
	victim2.die(&attacker.Actor, ctx)
	
	if ctx.World.Demographics.DeathsViolent != 1 {
		t.Errorf("Expected 1 violent death, got %d", ctx.World.Demographics.DeathsViolent)
	}
}
