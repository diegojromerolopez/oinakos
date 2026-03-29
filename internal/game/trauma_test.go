package game

import (
	"testing"
)

func TestTrauma_IncapacitationThreshold(t *testing.T) {
	ctx := NewTestContext()
	c := &Character{Actor: Actor{
		TemporalState: TemporalState{
			HealthPoints:    100,
			MaxHealthPoints: 100,
			Hunger:          100.0,
			Thirst:          100.0,
			Fatigue:         100.0,
		},
		Config: &EntityConfig{ID: "hero"},
	}}

	// Damage to 5 HP (should still be active)
	c.TakeDamage(95, nil, ctx)
	if c.IsIncapacitated() {
		t.Errorf("Expected actor to be active at 5 HP, got %v", c.State)
	}

	// Damage to 0 HP (should be incapacitated)
	c.TakeDamage(5, nil, ctx)
	if !c.IsIncapacitated() {
		t.Errorf("Expected actor to be incapacitated at 0 HP, got %v", c.State)
	}
	if !c.IsAlive() {
		t.Error("Actor should still be 'alive' (not truly dead) at 0 HP")
	}
}

func TestTrauma_BleedOutAndDeath(t *testing.T) {
	ctx := NewTestContext()
	c := &Character{Actor: Actor{
		TemporalState: TemporalState{
			HealthPoints:    0,
			MaxHealthPoints: 100,
			Hunger:          0.0,
			Thirst:          0.0,
			Fatigue:         0.0,
		},
		State:  ActorIncapacitated,
		Config: &EntityConfig{ID: "hero"},
	}}

	// Hour simulation (3600 ticks)
	for i := 0; i < 3600; i++ {
		c.Tick++
		c.SharedUpdate(ctx)
	}

	if c.TemporalState.HealthPoints != -1 {
		t.Errorf("Expected -1 HP after one game hour, got %d", c.TemporalState.HealthPoints)
	}

	// Fast forward to death (-10 HP)
	c.TemporalState.HealthPoints = -10
	c.SharedUpdate(ctx)

	if !c.IsTrulyDead() {
		t.Error("Expected actor to be truly dead at -10% health")
	}
	if c.IsAlive() {
		t.Error("IsAlive should be false when Truly Dead")
	}
}

func TestTrauma_DeterministicLimbLoss(t *testing.T) {
	ctx := NewTestContext()
	c := &Character{Actor: Actor{
		TemporalState: TemporalState{
			HealthPoints:    9, // < 10% of 100
			MaxHealthPoints: 100,
			Hunger:          100.0,
			Thirst:          100.0,
			Fatigue:         100.0,
		},
		Config: &EntityConfig{ID: "hero"},
	}}

	// Any hit at < 10% health should cause a trauma
	c.TakeDamage(1, nil, ctx)

	hasTrauma := c.Trauma.LeftArmLost || c.Trauma.RightArmLost ||
		c.Trauma.LeftLegLost || c.Trauma.RightLegLost ||
		c.Trauma.EyesLost > 0 || c.Trauma.BurnedAlive ||
		c.Trauma.SpineBroken

	if !hasTrauma {
		t.Error("Expected guaranteed trauma when hit at < 10% health")
	}
}

func TestTrauma_CumulativeTrauma(t *testing.T) {
	ctx := NewTestContext()
	c := &Character{Actor: Actor{
		TemporalState: TemporalState{
			HealthPoints:    0,
			MaxHealthPoints: 100,
			Hunger:          100.0,
			Thirst:          100.0,
			Fatigue:         100.0,
		},
		State:  ActorIncapacitated,
		Config: &EntityConfig{ID: "hero"},
	}}

	// Multiple hits while incapacitated should cause multiple traumas
	for i := 0; i < 5; i++ {
		c.TakeDamage(1, nil, ctx)
	}

	traumaCount := 0
	if c.Trauma.LeftArmLost {
		traumaCount++
	}
	if c.Trauma.RightArmLost {
		traumaCount++
	}
	if c.Trauma.LeftLegLost {
		traumaCount++
	}
	if c.Trauma.RightLegLost {
		traumaCount++
	}
	if c.Trauma.EyesLost > 0 {
		traumaCount += c.Trauma.EyesLost
	}
	if c.Trauma.BurnedAlive {
		traumaCount++
	}
	if c.Trauma.SpineBroken {
		traumaCount++
	}

	if traumaCount < 2 {
		t.Errorf("Expected multiple traumas from multiple hits, got count: %d", traumaCount)
	}
}

func TestTrauma_Recovery(t *testing.T) {
	c := &Character{Actor: Actor{
		TemporalState: TemporalState{
			HealthPoints:    -5,
			MaxHealthPoints: 100,
			Hunger:          100.0,
			Thirst:          100.0,
			Fatigue:         100.0,
		},
		State:  ActorIncapacitated,
		Config: &EntityConfig{ID: "hero"},
	}}

	// Heal to 10 HP (should stand up)
	c.Heal(15)

	if c.IsIncapacitated() {
		t.Error("Expected actor to NOT be incapacitated after healing to positive HP")
	}
	if c.State != ActorIdle {
		t.Errorf("Expected state to be ActorIdle after recovery, got %v", c.State)
	}
}
