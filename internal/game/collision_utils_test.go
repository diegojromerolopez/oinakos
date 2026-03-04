package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestFindSafePosition_EmptyFootprint(t *testing.T) {
	fps := engine.Polygon{}
	x, y := findSafePosition(10, 20, fps, nil)
	if x != 10 || y != 20 {
		t.Errorf("Expected unchanged position for empty footprint, got %v, %v", x, y)
	}
}

func TestFindSafePosition_AlreadySafe(t *testing.T) {
	fp := engine.Polygon{Points: []engine.Point{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}}}

	// No obstacles
	x, y := findSafePosition(5, 5, fp, nil)
	if x != 5 || y != 5 {
		t.Errorf("Expected unchanged position when no obstacles are present, got %v, %v", x, y)
	}

	// Uncolliding obstacle
	obs := NewObstacle("box", 10, 10, &ObstacleArchetype{
		Footprint: []FootprintPoint{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}},
	})
	x, y = findSafePosition(5, 5, fp, []*Obstacle{obs})
	if x != 5 || y != 5 {
		t.Errorf("Expected unchanged position when far from obstacles, got %v, %v", x, y)
	}
}

func TestFindSafePosition_CollisionMove(t *testing.T) {
	fp := engine.Polygon{Points: []engine.Point{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}}}

	// Colliding obstacle exactly at center
	obs := NewObstacle("box", 5, 5, &ObstacleArchetype{
		Footprint: []FootprintPoint{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}},
	})

	// Should push outwards.
	x, y := findSafePosition(5, 5, fp, []*Obstacle{obs})

	if x == 5 && y == 5 {
		t.Errorf("Expected position to change to avoid collision")
	}

	// We can't easily predict exact coordinates due to floating math & circling,
	// but we can check if the final position is safe.
	if isPositionColliding(x, y, fp, []*Obstacle{obs}) {
		t.Errorf("New position %v, %v still colliding!", x, y)
	}
}

func TestFindSafePosition_Trapped(t *testing.T) {
	// A huge footprint that covers everything up to dist 10
	fp := engine.Polygon{Points: []engine.Point{{X: -10, Y: -10}, {X: 10, Y: -10}, {X: 10, Y: 10}, {X: -10, Y: 10}}}

	// Colling
	obs := NewObstacle("box", 0, 0, &ObstacleArchetype{
		Footprint: []FootprintPoint{{X: -10, Y: -10}, {X: 10, Y: -10}, {X: 10, Y: 10}, {X: -10, Y: 10}},
	})

	x, y := findSafePosition(0, 0, fp, []*Obstacle{obs})

	// The function only checks up to 5 units, so this will never find a safe spot for such huge blocks.
	// It should return the original coordinates.
	if x != 0 || y != 0 {
		t.Errorf("Expected to return to origin if trapped, got %v, %v", x, y)
	}
}

func TestIsPositionColliding_DeadObstacleIgnored(t *testing.T) {
	fp := engine.Polygon{Points: []engine.Point{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}}}

	obs := NewObstacle("box", 5, 5, &ObstacleArchetype{
		Footprint: []FootprintPoint{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}},
	})
	obs.Alive = false

	if isPositionColliding(5, 5, fp, []*Obstacle{obs}) {
		t.Errorf("Dead obstacles should be ignored in collision")
	}
}
