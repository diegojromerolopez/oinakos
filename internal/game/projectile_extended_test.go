package game

import (
	"testing"
)

// TestProjectileLifecycle_Range verifies that projectiles despawn after traveling their maximum range.
func TestProjectileLifecycle_Range(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	// NewProjectile(x, y, dx, dy, speed, damage, isPlayer, maxRange)
	p := NewProjectile(0, 0, 1, 0, 1.0, 10, true, 5.0)

	var fts []*FloatingText

	// After 4 updates, it has moved 4.0 units. Still alive.
	for i := 0; i < 4; i++ {
		p.Update(mc, nil, &fts, nil)
	}
	if !p.Alive {
		t.Errorf("Projectile should be alive after 4 units, distance is %v", p.DistanceTraveled)
	}

	// 5th update moves it to 5.0 units.
	p.Update(mc, nil, &fts, nil)

	// Since DistanceTraveled (5.0) >= MaxRange (5.0), it should be dead.
	if p.Alive {
		t.Error("Projectile should be dead after reaching its max range of 5.0")
	}
}

// TestProjectileCollision_NPC is not directly supported as Projectile.Update
// only checks mc for collision in the current code (aiming at player).
// Let's test the mc collision instead.

func TestProjectileCollision_MC(t *testing.T) {
	mc := NewMainCharacter(2, 0, nil)
	mc.Health = 100

	// Projectile at (0,0) moving East (+X) toward MC at (2,0)
	// It is fired by an NPC (!p.IsPlayer)
	p := NewProjectile(0, 0, 1, 0, 1.0, 50, false, 100.0)

	var fts []*FloatingText

	// Update 1: p moves to (1,0). No collision yet (MC is at 2,0, dist is 1.0, threshold is 0.6)
	p.Update(mc, nil, &fts, nil)
	if !p.Alive || mc.Health != 100 {
		t.Errorf("P should be alive at (1,0), mc health %d", mc.Health)
	}

	// Update 2: p moves to (2,0). Collides.
	p.Update(mc, nil, &fts, nil)

	if p.Alive {
		t.Error("Projectile should be dead after hitting MC")
	}
	if mc.Health >= 100 {
		t.Errorf("MC should have taken damage, health is %d", mc.Health)
	}
}

// TestProjectileCollision_Obstacle verifies that projectiles are blocked by solid obstacles.
func TestProjectileCollision_Obstacle(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	p := NewProjectile(0, 0, 1, 0, 1.0, 10, true, 100.0)

	// Rock obstacle at (3, 0)
	obs := NewObstacle("rock", 3, 0, &ObstacleArchetype{
		ID:        "rock",
		Footprint: []FootprintPoint{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}},
	})
	obstacles := []*Obstacle{obs}

	var fts []*FloatingText

	// Update 1: (1,0)
	// Update 2: (2,0)
	// Update 3: (3,0) -> collision
	for i := 0; i < 3; i++ {
		p.Update(mc, obstacles, &fts, nil)
	}

	if p.Alive {
		t.Error("Projectile should be destroyed upon hitting a rock at (3,0)")
	}
}
