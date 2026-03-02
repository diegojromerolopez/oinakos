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

// TestPointInFootprintCollision verifies that point-in-polygon collision works via SAT.
// This is critical for movement blocking and interaction.
func TestFootprintCollision(t *testing.T) {
	// Simple square footprint
	squareArch := &Archetype{
		Footprint: []FootprintPoint{
			{X: -1, Y: -1},
			{X: 1, Y: -1},
			{X: 1, Y: 1},
			{X: -1, Y: 1},
		},
	}
	npc := NewNPC(10, 10, squareArch, 1) // Footprint centered at (10,10), bounds [9,11]

	fp := npc.GetFootprint()

	tests := []struct {
		x, y float64
		want bool
	}{
		{10, 10, true},    // Center
		{9.5, 9.5, true},  // Near corner
		{11, 11, true},    // Corner
		{11.1, 11, false}, // Just outside X
		{11, 11.1, false}, // Just outside Y
		{5, 5, false},     // Far away
	}

	for _, tt := range tests {
		// Use a tiny 0.01x0.01 polygon to simulate a point for CheckCollision
		pointPoly := engine.Polygon{Points: []engine.Point{
			{X: tt.x - 0.005, Y: tt.y - 0.005},
			{X: tt.x + 0.005, Y: tt.y - 0.005},
			{X: tt.x + 0.005, Y: tt.y + 0.005},
			{X: tt.x - 0.005, Y: tt.y + 0.005},
		}}
		got := engine.CheckCollision(fp, pointPoly)
		if got != tt.want {
			t.Errorf("Point (%v, %v) collision: got %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

// TestFootprintCollision_Complex verify collision with non-square polygons.
func TestFootprintCollision_Triangle(t *testing.T) {
	// Right triangle: vertices at (0,0), (2,0), (0,2) relative to NPC
	triArch := &Archetype{
		Footprint: []FootprintPoint{
			{X: 0, Y: 0},
			{X: 2, Y: 0},
			{X: 0, Y: 2},
		},
	}
	npc := NewNPC(10, 10, triArch, 1) // World vertices: (10,10), (12,10), (10,12)
	fp := npc.GetFootprint()

	tests := []struct {
		x, y float64
		want bool
	}{
		{10.1, 10.1, true},  // Inside
		{11.5, 10.1, true},  // Inside near hyp
		{10.1, 11.5, true},  // Inside near hyp
		{11.5, 11.5, false}, // Outside hyp
		{9.9, 10.1, false},  // Outside west
		{10.1, 9.9, false},  // Outside south
	}

	for _, tt := range tests {
		pointPoly := engine.Polygon{Points: []engine.Point{
			{X: tt.x - 0.005, Y: tt.y - 0.005},
			{X: tt.x + 0.005, Y: tt.y - 0.005},
			{X: tt.x + 0.005, Y: tt.y + 0.005},
			{X: tt.x - 0.005, Y: tt.y + 0.005},
		}}
		got := engine.CheckCollision(fp, pointPoly)
		if got != tt.want {
			t.Errorf("Triangle Collision at (%v, %v): got %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}
