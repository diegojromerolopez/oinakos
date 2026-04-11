package game

import (
	"testing"
)

func TestTrauma_IncapacitationThreshold(t *testing.T) {
	ctx := NewTestContext()
	ctx.Settings.AdultMode = true
	c := &Character{Actor: Actor{
		State: State{
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
		t.Errorf("Expected actor to be active at 5 HP, got %v", c.ActionState)
	}

	// Damage to 0 HP (should be incapacitated)
	c.TakeDamage(5, nil, ctx)
	if !c.IsIncapacitated() {
		t.Errorf("Expected actor to be incapacitated at 0 HP, got %v", c.ActionState)
	}
	if !c.IsAlive() {
		t.Error("Actor should still be 'alive' (not truly dead) at 0 HP")
	}
}

func TestTrauma_BleedOutAndDeath(t *testing.T) {
	ctx := NewTestContext()
	c := &Character{Actor: Actor{
		State: State{
			MaxHealthPoints: 100,
			Hunger:          0.0,
			Thirst:          0.0,
			Fatigue:         0.0,
		},
		ActionState: ActorIncapacitated,
		Config:      &EntityConfig{ID: "hero"},
	}}

	// Hour simulation (TicksPerHour ticks)
	for i := 0; i < TicksPerHour; i++ {
		c.Tick++
		c.SharedUpdate(ctx)
	}

	if c.State.HealthPoints != -1 {
		t.Errorf("Expected -1 HP after one game hour, got %d", c.State.HealthPoints)
	}

	// Fast forward to death (-10 HP)
	c.State.HealthPoints = -10
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
	ctx.Settings.AdultMode = true
	c := &Character{Actor: Actor{
		State: State{
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
	ctx.Settings.AdultMode = true
	c := &Character{Actor: Actor{
		State: State{
			MaxHealthPoints: 100,
			Hunger:          100.0,
			Thirst:          100.0,
			Fatigue:         100.0,
		},
		ActionState: ActorIncapacitated,
		Config:      &EntityConfig{ID: "hero"},
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
		State: State{
			MaxHealthPoints: 100,
			Hunger:          100.0,
			Thirst:          100.0,
			Fatigue:         100.0,
		},
		ActionState: ActorIncapacitated,
		Config:      &EntityConfig{ID: "hero"},
	}}

	// Heal to 10 HP (should stand up)
	c.Heal(15)

	if c.IsIncapacitated() {
		t.Error("Expected actor to NOT be incapacitated after healing to positive HP")
	}
	if c.ActionState != ActorIdle {
		t.Errorf("Expected state to be ActorIdle after recovery, got %v", c.ActionState)
	}
}
