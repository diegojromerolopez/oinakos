package game

import (
	"math"
	"testing"
)

func TestNewProjectile_Normalization(t *testing.T) {
	p := NewProjectile(0, 0, 3, 4, 1.0, 5, true, 100.0)
	mag := math.Sqrt(p.Dx*p.Dx + p.Dy*p.Dy)
	if math.Abs(mag-1.0) > 0.001 {
		t.Errorf("Direction should be normalized, got mag=%v", mag)
	}
}

func TestNewProjectile_ZeroDirection(t *testing.T) {
	// Zero direction should not panic (mag=0 guard)
	p := NewProjectile(0, 0, 0, 0, 1.0, 5, true, 100.0)
	if p == nil {
		t.Fatal("NewProjectile returned nil")
	}
	if !p.Alive {
		t.Error("New projectile should be alive")
	}
}

func TestProjectileUpdate_Moves(t *testing.T) {
	ctx := NewTestContext()
	p := NewProjectile(0, 0, 1, 0, 0.5, 5, true, 100.0)
	p.Update(ctx)
	if math.Abs(p.X-0.5) > 0.001 {
		t.Errorf("X after update: got %v, want 0.5", p.X)
	}
}

func TestProjectileUpdate_SkipsIfDead(t *testing.T) {
	ctx := NewTestContext()
	p := NewProjectile(0, 0, 1, 0, 1.0, 5, true, 100.0)
	p.Alive = false
	p.Update(ctx)
	// X should not have changed
	if p.X != 0 {
		t.Errorf("Dead projectile should not move, X=%v", p.X)
	}
}

func TestProjectileUpdate_HitsObstacle(t *testing.T) {
	ctx := NewTestContext()
	// Place projectile and obstacle at same location
	arch := makeObstacleArchetype(100)
	arch.Footprint = []FootprintPoint{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}}
	obs := NewObstacle("test_obs_a", 0, 0, arch)
	p := NewProjectile(0, 0, 1, 0, 0.0, 5, true, 100.0) // speed=0, won't move past
	ctx.World.Obstacles = []*Obstacle{obs}
	p.Update(ctx)
	if p.Alive {
		t.Error("Projectile should be killed when overlapping an obstacle")
	}
}

func TestProjectileUpdate_ObstacleDead(t *testing.T) {
	ctx := NewTestContext()
	arch := makeObstacleArchetype(100)
	arch.Footprint = []FootprintPoint{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}}
	obs := NewObstacle("test_obs_b", 0, 0, arch)
	obs.Alive = false
	p := NewProjectile(0, 0, 1, 0, 0.0, 5, true, 100.0)
	ctx.World.Obstacles = []*Obstacle{obs}
	p.Update(ctx)
	// Dead obstacle should not kill the projectile
	if !p.Alive {
		t.Error("Dead obstacle should not stop the projectile")
	}
}
