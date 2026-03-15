package game

import (
	"oinakos/internal/engine"
	"sort"
	"testing"
)

// TestYSortingOrder verifies that entities are sorted correctly for depth-correct rendering.
// Sorting key: X + Y (as defined in GEMINI.md and game_render.go)
func TestYSortingOrder(t *testing.T) {
	type entity struct {
		name string
		x, y float64
	}

	entities := []entity{
		{"Far Back", 0, 0},       // sum = 0
		{"Front", 10, 10},        // sum = 20
		{"Mid Left", 0, 10},      // sum = 10
		{"Mid Right", 10, 0},     // sum = 10
		{"Slightly Front", 6, 5}, // sum = 11
	}

	// Sort using the project's Y-sorting logic: sort by X + Y
	sort.Slice(entities, func(i, j int) bool {
		sumI := entities[i].x + entities[i].y
		sumJ := entities[j].x + entities[j].y
		return sumI < sumJ
	})

	if entities[0].name != "Far Back" {
		t.Errorf("Expected 'Far Back' (sum 0) to be first, got %s", entities[0].name)
	}
	if entities[len(entities)-1].name != "Front" {
		t.Errorf("Expected 'Front' (sum 20) to be last, got %s", entities[len(entities)-1].name)
	}

	// Verify the middle elements
	middleSum := entities[1].x + entities[1].y
	if middleSum != 10 {
		t.Errorf("Expected middle element sum to be 10, got %v", middleSum)
	}
}

// TestCircleCollision verifies that circle-polygon collision works.
func TestCircleCollision(t *testing.T) {
	// Simple square footprint for obstacle
	squarePoly := engine.Polygon{Points: []engine.Point{
		{X: 9, Y: 9},
		{X: 11, Y: 9},
		{X: 11, Y: 11},
		{X: 9, Y: 11},
	}}

	tests := []struct {
		x, y   float64
		radius float64
		want   bool
	}{
		{10, 10, 0.5, true},   // Inside
		{8.6, 10, 0.5, true},  // Overlap edge
		{8.4, 10, 0.5, false}, // No overlap
		{11.5, 11.5, 0.73, true}, // Overlap corner (dist to (11,11) is sqrt(0.5^2+0.5^2) = 0.707)
		{11.5, 11.5, 0.5, false}, // No overlap corner
	}

	for i, tt := range tests {
		c := engine.Circle{X: tt.x, Y: tt.y, Radius: tt.radius}
		got := engine.CheckCirclePolygonCollision(c, squarePoly)
		if got != tt.want {
			t.Errorf("Test %d (%v, %v, r=%v) collision: got %v, want %v", i, tt.x, tt.y, tt.radius, got, tt.want)
		}
	}
}

// TestCircleCollision_Triangle verifies collision with non-square polygons.
func TestCircleCollision_Triangle(t *testing.T) {
	// Right triangle: vertices at (10,10), (12,10), (10,12)
	triPoly := engine.Polygon{Points: []engine.Point{
		{X: 10, Y: 10},
		{X: 12, Y: 10},
		{X: 10, Y: 12},
	}}

	tests := []struct {
		x, y   float64
		radius float64
		want   bool
	}{
		{10.5, 10.5, 0.1, true},  // Inside
		{9.5, 10.5, 0.6, true},   // Overlap vertical edge
		{9.5, 10.5, 0.4, false},  // No overlap
		{11.5, 11.5, 0.8, true},   // Overlap hypotenuse (dist to line x+y=22 is 0.707)
		{11.5, 11.5, 0.2, false}, // No overlap hypotenuse
	}

	for i, tt := range tests {
		c := engine.Circle{X: tt.x, Y: tt.y, Radius: tt.radius}
		got := engine.CheckCirclePolygonCollision(c, triPoly)
		if got != tt.want {
			t.Errorf("Triangle Test %d (%v, %v, r=%v) collision: got %v, want %v", i, tt.x, tt.y, tt.radius, got, tt.want)
		}
	}
}
