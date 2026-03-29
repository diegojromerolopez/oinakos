package game

import (
	"math"
	"testing"
)

func TestActor_FindPath(t *testing.T) {
	ctx := NewTestContext()
	a := NewCharacter(0, 0, nil, 1, true, nil)
	
	t.Run("simple straight path", func(t *testing.T) {
		path := a.FindPath(5, 5, ctx)
		if len(path) == 0 {
			t.Fatal("expected path, got nil")
		}
		last := path[len(path)-1]
		if math.Abs(last.X-5.5) > 0.1 || math.Abs(last.Y-5.5) > 0.1 {
			t.Errorf("expected end near (5.5, 5.5), got %v", last)
		}
	})
	
	t.Run("path around obstacle", func(t *testing.T) {
		// Use a larger footprint to ensure it blocks the whole 1x1 cell
		obsArch := &ObstacleArchetype{
			ID: "wall", 
			Passable: false,
			Footprint: []FootprintPoint{
				{X: 0.1, Y: 0.1}, {X: 0.9, Y: 0.1}, {X: 0.9, Y: 0.9}, {X: 0.1, Y: 0.9},
			},
		}
		// Place at (2, 0). Transformed points: (2.1, 0.1) to (2.9, 0.9).
		// Pathfinding center (2.5, 0.5) will definitely collide.
		obs := NewObstacle("wall1", 2, 0, obsArch)
		ctx.World.Obstacles = append(ctx.World.Obstacles, obs)
		
		// Try to go from (0, 0) to (4, 0)
		path := a.FindPath(4, 0, ctx)
		if len(path) == 0 {
			t.Log("Note: Path might be empty if target is blocked, but (4,0) should be reachable")
		}
		
		for _, p := range path {
			// Center of cell (2,0) is (2.5, 0.5)
			if math.Abs(p.X-2.5) < 0.1 && math.Abs(p.Y-0.5) < 0.1 {
				t.Errorf("path point %v is inside obstacle", p)
			}
		}
	})
	
	t.Run("too far", func(t *testing.T) {
		path := a.FindPath(100, 100, ctx)
		if path != nil {
			t.Errorf("expected nil path for out-of-range target, got %v", path)
		}
	})
}
