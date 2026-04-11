package game

import (
	"testing"
)

func TestAdultModeEnforcement(t *testing.T) {
	// Test 1: Adult Mode is OFF by default
	settings := DefaultSettings()
	if settings.AdultMode {
		t.Errorf("Expected AdultMode to be disabled by default")
	}

	// Test 2: Adult Mode OFF - Consensual sex SHOULD work now (for population growth)
	ctx := NewTestContext()
	ctx.Settings.AdultMode = false

	fConf := &EntityConfig{ID: "f", Gender: "female"}; fConf.Stats.Age.Current = FloatInterval{Min: 25, Max: 25}
	mConf := &EntityConfig{ID: "m", Gender: "male"}; mConf.Stats.Age.Current = FloatInterval{Min: 25, Max: 25}
	male := NewCharacter(10, 10, mConf, 25, false, nil)
	female := NewCharacter(11, 11, fConf, 25, false, nil)

	// Ensure success
	pregnant := false
	for i := 0; i < 100; i++ {
		female.haveSex(ctx, &male.Actor, "vaginal")
		if female.IsPregnant {
			pregnant = true
			break
		}
		female.MatingCooldown = 0
		male.MatingCooldown = 0
	}
	if !pregnant {
		t.Errorf("Consensual pregnancy should be possible even when AdultMode is OFF")
	}

	// Test 3: Adult Mode OFF - Incest SHOULD NOT work
	female.IsPregnant = false
	male.Name = "Father"
	female.FatherID = "Father" // Daughter
	
	for i := 0; i < 50; i++ {
		female.haveSex(ctx, &male.Actor, "vaginal")
		if female.IsPregnant {
			t.Errorf("INCESTuous pregnancy occurred while AdultMode was OFF")
		}
		female.MatingCooldown = 0
	}

	// Test 4: Adult Mode OFF - Bestiality SHOULD NOT work
	female.FatherID = "" // Clear incest
	animal := NewCharacter(12, 12, &EntityConfig{ID: "sheep", IsAnimal: true}, 5, false, nil)
	
	for i := 0; i < 50; i++ {
		female.haveSex(ctx, &animal.Actor, "bestiality")
		if female.IsPregnant {
			t.Errorf("BESTIALITY pregnancy occurred while AdultMode was OFF")
		}
		female.MatingCooldown = 0
	}

	// Test 5: Adult Mode OFF - Forced sex (Berserk) SHOULD NOT work
	male.ActionState = ActorBerserk
	for i := 0; i < 50; i++ {
		female.haveSex(ctx, &male.Actor, "vaginal")
		if female.IsPregnant {
			t.Errorf("FORCED pregnancy occurred while AdultMode was OFF")
		}
		female.MatingCooldown = 0
	}
}

func TestAdultModeEnabled_AllowTaboos(t *testing.T) {
	ctx := NewTestContext()
	ctx.Settings.AdultMode = true

	fConf := &EntityConfig{ID: "f", Gender: "female"}; fConf.Stats.Age.Current = FloatInterval{Min: 25, Max: 25}
	mConf := &EntityConfig{ID: "m", Name: "Father", Gender: "male"}; mConf.Stats.Age.Current = FloatInterval{Min: 25, Max: 25}
	male := NewCharacter(10, 10, mConf, 25, false, nil)
	female := NewCharacter(11, 11, fConf, 25, false, nil)
	female.FatherID = "Father"

	// Test Incest Allowed
	pregnant := false
	for i := 0; i < 100; i++ {
		female.haveSex(ctx, &male.Actor, "vaginal")
		if female.IsPregnant {
			pregnant = true
			break
		}
		female.MatingCooldown = 0
		male.MatingCooldown = 0
	}
	if !pregnant {
		t.Errorf("Expected INCEST pregnancy to be possible when AdultMode is ON")
	}
}
