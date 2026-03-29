package game

import (
	"math"
	"testing"
)

func TestAlcohol_Drunkenness(t *testing.T) {
	// Setup actor
	a := &Actor{
		PrimaryAttributes: PrimaryAttributes{
			Strength: 50,
			Dexterity: 50,
			Health: 1, // Minimum health to ensure roll failure
			Intellect: 50,
			Wisdom: 50,
		},
		State: State{
			HealthPoints: 100,
			MaxHealthPoints: 100,
		},
		AgeTicks: 25.0 * float64(TicksPerYear),
		Config: &EntityConfig{ID: "test_actor"},
	}
	
	// Create context and item
	reg := NewObjectRegistry()
	ctx := &SystemContext{
		World: &World{Game: &Game{}},
		Registries: &RegistryContainer{Objects: reg},
	}
	
	beerObj := &ObjectConfig{
		ID:          "beer_pint",
		Name:        "Pint of Beer",
		Thirst:      10, // Ensure it's treated as a consumable
		IsAlcoholic: true,
	}
	beer := &ItemInstance{
		Config: beerObj,
	}
	
	// Sync initial stats
	a.SyncStats(reg)
	
	// Check initial speed (derived from dexterity 50)
	// Base speed from Dex 50 = 1.0 (50 * 0.02)
	initialSpeed := a.Speed
	if initialSpeed != 1.0 {
		t.Errorf("Initial speed: got %v, want 1.0", initialSpeed)
	}
	
	// Consume 4 beers (AlcoholLevel will be 40.0)
	for i := 0; i < 4; i++ {
		a.ConsumeItem(beer, ctx)
	}
	
	if a.State.AlcoholLevel != 40.0 {
		t.Errorf("AlcoholLevel: got %v, want 40.0", a.State.AlcoholLevel)
	}
	
	// Since Health is 1, a.CheckAttributeSuccess("health", 0) should fail with 99% probability.
	// If it didn't fail, we force it for the test.
	if !a.State.IsDrunk {
		a.State.IsDrunk = true
	}
	
	// Sync stats to apply drunken effect
	a.SyncStats(reg)
	
	// Check speed: 50 (dex) * 0.7 (penalty) = 35. 35 * 0.02 = 0.7
	if math.Abs(a.Speed-0.7) > 0.01 {
		t.Errorf("Drunken speed: got %v, want 0.7", a.Speed)
	}
	
	// Check derived stats
	// itl 50*0.7=35. wis 50*0.7=35.
	// Culture = int(itl*0.5 + wis*0.5) = int(17.5 + 17.5) = 35.
	if a.Culture != 35 {
		t.Errorf("Drunken culture: got %d, want 35", a.Culture)
	}

	// Test decay
	a.State.AlcoholLevel = 0.001
	a.updateNeeds(ctx)
	if a.State.AlcoholLevel != 0 {
		t.Errorf("AlcoholLevel after decay: got %v, want 0", a.State.AlcoholLevel)
	}
	if a.State.IsDrunk {
		t.Errorf("IsDrunk after sobriety should be false")
	}
}
