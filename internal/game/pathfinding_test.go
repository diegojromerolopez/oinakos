package game

import (
	"math"
	"testing"
)

func TestActor_FindPath(t *testing.T) {
	ctx := NewTestContext()
	a := NewCharacter(0, 0, nil, 1, true, ctx.Registries.Objects)
	
	t.Run("simple straight path", func(t *testing.T) {
		path := a.FindPath(5, 5, ctx)
		if len(path) == 0 {
			t.Fatal("expected path, got nil")
		}
		last := path[len(path)-1]
		if math.Abs(last.X-5.0) > 0.1 || math.Abs(last.Y-5.0) > 0.1 {
			t.Errorf("expected end at (5.0, 5.0), got %v", last)
		}
	})
	
	t.Run("path around obstacle", func(t *testing.T) {
		// Centered 1x1 footprint to block the cell origin effectively
		obsArch := &ObstacleArchetype{
			ID: "wall", 
			Passable: false,
			Footprint: []FootprintPoint{
				{X: -0.5, Y: -0.5}, {X: 0.5, Y: -0.5}, {X: 0.5, Y: 0.5}, {X: -0.5, Y: 0.5},
			},
		}
		// Block (2, 0)
		obs := NewObstacle("wall1", 2, 0, obsArch)
		ctx.World.Obstacles = []*Obstacle{obs}
		ctx.World.Game.obstacles = ctx.World.Obstacles 
		
		path := a.FindPath(4, 0, ctx)
		if len(path) == 0 {
			t.Fatal("expected path around wall, got nil")
		}
		
		for _, p := range path {
			// Center of cell (2,0)
			if math.Abs(p.X-2.0) < 0.1 && math.Abs(p.Y-0.0) < 0.1 {
				t.Errorf("path point %v is inside obstacle", p)
			}
		}
	})
	
	t.Run("too far", func(t *testing.T) {
		path := a.FindPath(1000, 1000, ctx)
		if path != nil {
			t.Errorf("expected nil path for extremely distant target, got %v", path)
		}
	})
}
