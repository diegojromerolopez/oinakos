package engine

import (
	"testing"
)

func TestCheckCollision(t *testing.T) {
	rect1 := Polygon{Points: []Point{
		{0, 0}, {2, 0}, {2, 2}, {0, 2},
	}}

	rect2 := Polygon{Points: []Point{
		{1, 1}, {3, 1}, {3, 3}, {1, 3},
	}}

	rect3 := Polygon{Points: []Point{
		{4, 4}, {6, 4}, {6, 6}, {4, 6},
	}}

	tests := []struct {
		name string
		p1   Polygon
		p2   Polygon
		want bool
	}{
		{"overlapping rectangles", rect1, rect2, true},
		{"separated rectangles", rect1, rect3, false},
		{"touching but not overlapping", rect1, Polygon{Points: []Point{{2, 0}, {4, 0}, {4, 2}, {2, 2}}}, true}, // SAT usually considers touching as colliding if we use >= or <=
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckCollision(tt.p1, tt.p2); got != tt.want {
				t.Errorf("CheckCollision() = %v, want %v", got, tt.want)
			}
		})
	}
}
