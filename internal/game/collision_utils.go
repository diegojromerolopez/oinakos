package game

import (
	"math"
	"oinakos/internal/engine"
)

// findSafePosition searches for a non-colliding position near (x, y) given a collision circle.
func findSafePosition(x, y float64, circle engine.Circle, obstacles []*Obstacle) (float64, float64) {
	if circle.Radius <= 0 {
		return x, y
	}
	// 1. Check if the current position is already safe.
	if !isPositionColliding(x, y, circle, obstacles) {
		return x, y
	}

	// 2. Search in expanding circles until a safe spot is found.
	// We check up to 5 world units (characters are ~1 unit wide).
	for dist := 0.2; dist <= 5.0; dist += 0.2 {
		// Increase number of samples as circle expands (8, 16, 24...)
		steps := 8 + int(dist*4)
		for i := 0; i < steps; i++ {
			angle := (float64(i) / float64(steps)) * 2 * math.Pi
			testX := x + math.Cos(angle)*dist
			testY := y + math.Sin(angle)*dist

			if !isPositionColliding(testX, testY, circle, obstacles) {
				return testX, testY
			}
		}
	}

	// Return original if no safe spot found (highly unlikely in open world)
	return x, y
}

// isPositionColliding checks if a collision circle at a specific (x, y) overlaps any obstacle.
func isPositionColliding(x, y float64, circle engine.Circle, obstacles []*Obstacle) bool {
	circle.X = x
	circle.Y = y
	for _, o := range obstacles {
		if !o.Alive {
			continue
		}
		if engine.CheckCirclePolygonCollision(circle, o.GetFootprint()) {
			return true
		}
	}
	return false
}
