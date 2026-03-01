package game

import "testing"

func makeObstacleArchetype(health int) *ObstacleArchetype {
	return &ObstacleArchetype{
		ID:           "test",
		Name:         "Test",
		Health:       health,
		Destructible: true,
		Footprint: []FootprintPoint{
			{X: -0.25, Y: -0.25}, {X: 0.25, Y: -0.25},
			{X: 0.25, Y: 0.25}, {X: -0.25, Y: 0.25},
		},
	}
}

func TestNewObstacle_NilArchetype(t *testing.T) {
	o := NewObstacle(1, 2, nil)
	if o.X != 1 || o.Y != 2 {
		t.Errorf("Position: got (%v,%v), want (1,2)", o.X, o.Y)
	}
	if o.Health != 0 {
		t.Errorf("Health should be 0 for nil archetype, got %d", o.Health)
	}
	if !o.Alive {
		t.Error("New obstacle should be alive")
	}
}

func TestNewObstacle_WithArchetype(t *testing.T) {
	arch := makeObstacleArchetype(100)
	o := NewObstacle(3, 4, arch)
	if o.Health != 100 {
		t.Errorf("Health: got %d, want 100", o.Health)
	}
}

func TestObstacleUpdate_CooldownDecrement(t *testing.T) {
	o := NewObstacle(0, 0, makeObstacleArchetype(10))
	o.CooldownTicks = 5
	o.Update()
	if o.CooldownTicks != 4 {
		t.Errorf("CooldownTicks after Update: got %d, want 4", o.CooldownTicks)
	}
}

func TestObstacleUpdate_NoCooldownBelowZero(t *testing.T) {
	o := NewObstacle(0, 0, makeObstacleArchetype(10))
	o.CooldownTicks = 0
	o.Update()
	if o.CooldownTicks != 0 {
		t.Errorf("CooldownTicks should stay at 0, got %d", o.CooldownTicks)
	}
}

func TestObstacleUpdate_Dead(t *testing.T) {
	o := NewObstacle(0, 0, makeObstacleArchetype(10))
	o.Alive = false
	o.CooldownTicks = 5
	o.Update() // should return early
	if o.CooldownTicks != 5 {
		t.Error("Dead obstacle should not decrement cooldown")
	}
}

func TestObstacleTakeDamage_Normal(t *testing.T) {
	o := NewObstacle(0, 0, makeObstacleArchetype(100))
	o.TakeDamage(30)
	if o.Health != 70 {
		t.Errorf("Health after 30 damage: got %d, want 70", o.Health)
	}
	if !o.Alive {
		t.Error("Should still be alive")
	}
}

func TestObstacleTakeDamage_Lethal(t *testing.T) {
	o := NewObstacle(0, 0, makeObstacleArchetype(50))
	o.TakeDamage(100)
	if o.Alive {
		t.Error("Should be dead after lethal damage")
	}
}

func TestObstacleTakeDamage_Indestructible(t *testing.T) {
	arch := makeObstacleArchetype(1000)
	arch.Destructible = false // explicitly mark as indestructible
	o := NewObstacle(0, 0, arch)
	o.TakeDamage(9999)
	if !o.Alive || o.Health != 1000 {
		t.Error("Indestructible obstacle should take no damage")
	}
}

func TestObstacleTakeDamage_AlreadyDead(t *testing.T) {
	o := NewObstacle(0, 0, makeObstacleArchetype(100))
	o.Alive = false
	o.TakeDamage(10)
	if o.Health != 100 {
		t.Error("Dead obstacle health should not change")
	}
}

func TestObstacleTakeDamage_NilArchetype(t *testing.T) {
	o := NewObstacle(0, 0, nil)
	o.TakeDamage(10) // should not panic
}

func TestObstacleGetFootprint_DefaultSize(t *testing.T) {
	o := NewObstacle(0, 0, nil)
	fp := o.GetFootprint()
	if len(fp.Points) != 4 {
		t.Errorf("Footprint should have 4 points, got %d", len(fp.Points))
	}
}

func TestObstacleGetFootprint_WithArchetype(t *testing.T) {
	arch := makeObstacleArchetype(10)
	o := NewObstacle(2, 3, arch)
	fp := o.GetFootprint()
	if len(fp.Points) != 4 {
		t.Errorf("Footprint should have 4 points, got %d", len(fp.Points))
	}
	// Footprint from archetype is ±0.25 centered at (2, 3)
	for _, pt := range fp.Points {
		if pt.X < 1.74 || pt.X > 2.26 {
			t.Errorf("Footprint X out of expected range: %v", pt.X)
		}
		if pt.Y < 2.74 || pt.Y > 3.26 {
			t.Errorf("Footprint Y out of expected range: %v", pt.Y)
		}
	}
}
