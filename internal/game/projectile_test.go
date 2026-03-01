package game

import (
	"math"
	"testing"
)

func TestNewProjectile_Normalization(t *testing.T) {
	p := NewProjectile(0, 0, 3, 4, 1.0, 5, true)
	mag := math.Sqrt(p.Dx*p.Dx + p.Dy*p.Dy)
	if math.Abs(mag-1.0) > 0.001 {
		t.Errorf("Direction should be normalized, got mag=%v", mag)
	}
}

func TestNewProjectile_ZeroDirection(t *testing.T) {
	// Zero direction should not panic (mag=0 guard)
	p := NewProjectile(0, 0, 0, 0, 1.0, 5, true)
	if p == nil {
		t.Fatal("NewProjectile returned nil")
	}
	if !p.Alive {
		t.Error("New projectile should be alive")
	}
}

func TestProjectileUpdate_Moves(t *testing.T) {
	p := NewProjectile(0, 0, 1, 0, 0.5, 5, true)
	var fts []*FloatingText
	p.Update(nil, nil, &fts)
	if math.Abs(p.X-0.5) > 0.001 {
		t.Errorf("X after update: got %v, want 0.5", p.X)
	}
}

func TestProjectileUpdate_SkipsIfDead(t *testing.T) {
	p := NewProjectile(0, 0, 1, 0, 1.0, 5, true)
	p.Alive = false
	var fts []*FloatingText
	p.Update(nil, nil, &fts)
	// X should not have changed
	if p.X != 0 {
		t.Errorf("Dead projectile should not move, X=%v", p.X)
	}
}

func TestProjectileUpdate_HitsObstacle(t *testing.T) {
	// Place projectile and obstacle at same location
	arch := makeObstacleArchetype(100)
	arch.Footprint = []FootprintPoint{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}}
	obs := NewObstacle(0, 0, arch)
	p := NewProjectile(0, 0, 1, 0, 0.0, 5, true) // speed=0, won't move past
	var fts []*FloatingText
	p.Update(nil, []*Obstacle{obs}, &fts)
	if p.Alive {
		t.Error("Projectile should be killed when overlapping an obstacle")
	}
}

func TestProjectileUpdate_ObstacleDead(t *testing.T) {
	arch := makeObstacleArchetype(100)
	arch.Footprint = []FootprintPoint{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}}
	obs := NewObstacle(0, 0, arch)
	obs.Alive = false
	p := NewProjectile(0, 0, 1, 0, 0.0, 5, true)
	var fts []*FloatingText
	p.Update(nil, []*Obstacle{obs}, &fts)
	// Dead obstacle should not kill the projectile
	if !p.Alive {
		t.Error("Dead obstacle should not stop the projectile")
	}
}
