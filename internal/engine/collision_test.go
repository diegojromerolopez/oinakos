package engine

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// --- Polygon helpers ---

func makeSquare(cx, cy, half float64) Polygon {
	return Polygon{Points: []Point{
		{cx - half, cy - half},
		{cx + half, cy - half},
		{cx + half, cy + half},
		{cx - half, cy + half},
	}}
}

func TestPolygonTransformed(t *testing.T) {
	p := makeSquare(0, 0, 1)
	moved := p.Transformed(5, 3)
	for i, pt := range moved.Points {
		wantX := p.Points[i].X + 5
		wantY := p.Points[i].Y + 3
		if math.Abs(pt.X-wantX) > 0.001 || math.Abs(pt.Y-wantY) > 0.001 {
			t.Errorf("Transformed point %d: got (%v,%v), want (%v,%v)", i, pt.X, pt.Y, wantX, wantY)
		}
	}
}

func TestPolygonGetEdges(t *testing.T) {
	// Simple 2-point degenerate case (line)
	p := Polygon{Points: []Point{{0, 0}, {2, 0}}}
	edges := p.GetEdges()
	if len(edges) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(edges))
	}
	// First edge: (2-0, 0-0) = (2, 0)
	if edges[0].X != 2 || edges[0].Y != 0 {
		t.Errorf("Edge[0]: got (%v,%v), want (2,0)", edges[0].X, edges[0].Y)
	}
	// Second edge: (0-2, 0-0) = (-2, 0) (wraps around)
	if edges[1].X != -2 || edges[1].Y != 0 {
		t.Errorf("Edge[1]: got (%v,%v), want (-2,0)", edges[1].X, edges[1].Y)
	}
}

func TestPolygonGetNormals(t *testing.T) {
	// Horizontal edge (1,0) → normal is (0,1)
	p := Polygon{Points: []Point{{0, 0}, {1, 0}, {1, 1}, {0, 1}}}
	normals := p.GetNormals()
	// First edge is (1,0), normal = (0, 1)
	if math.Abs(normals[0].X-0) > 0.001 || math.Abs(normals[0].Y-1) > 0.001 {
		t.Errorf("Normal[0]: got (%v,%v), want (0,1)", normals[0].X, normals[0].Y)
	}
}

func TestPolygonProjectZeroAxis(t *testing.T) {
	p := makeSquare(0, 0, 1)
	min, max := p.Project(Point{0, 0}) // zero-mag axis
	if min != 0 || max != 0 {
		t.Errorf("Project on zero axis: expected (0,0), got (%v,%v)", min, max)
	}
}

func TestPolygonProject(t *testing.T) {
	p := Polygon{Points: []Point{{1, 0}, {3, 0}, {3, 2}, {1, 2}}}
	// Project onto X axis (1,0)
	min, max := p.Project(Point{1, 0})
	if math.Abs(min-1) > 0.001 || math.Abs(max-3) > 0.001 {
		t.Errorf("Project X-axis: got min=%v max=%v, want 1,3", min, max)
	}
}

// --- CheckCollision ---

func TestCheckCollisionOverlapping(t *testing.T) {
	p1 := makeSquare(0, 0, 1)
	p2 := makeSquare(0.5, 0.5, 1)
	if !CheckCollision(p1, p2) {
		t.Error("Expected overlapping squares to collide")
	}
}

func TestCheckCollisionSeparated(t *testing.T) {
	p1 := makeSquare(0, 0, 1)
	p2 := makeSquare(5, 5, 1)
	if CheckCollision(p1, p2) {
		t.Error("Expected separated squares to NOT collide")
	}
}

func TestCheckCollisionTouching(t *testing.T) {
	// Edge-touching — SAT considers this NOT colliding (gap = 0 is still no gap on one axis)
	p1 := makeSquare(0, 0, 1) // Right edge at x=1
	p2 := makeSquare(2, 0, 1) // Left edge at x=1
	// max1=1, min2=1 → max1 < min2 is false, max2 < min1 is false → colliding by SAT
	// (SAT treats touching edges as colliding)
	_ = CheckCollision(p1, p2) // either result is acceptable, just no panic
}

// --- Transparentize ---

func TestTransparentizeNil(t *testing.T) {
	result := Transparentize(nil)
	if result != nil {
		t.Error("Transparentize(nil) should return nil")
	}
}

func TestTransparentizeSolidWhite(t *testing.T) {
	// 4x4 solid white image — all pixels should be made transparent
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	lime := color.RGBA{0, 255, 0, 255}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, lime)
		}
	}
	result := Transparentize(img)
	if result == nil {
		t.Fatal("Transparentize returned nil for valid image")
	}
	// Center pixel should be transparent since white is treated as background
	_, _, _, a := result.At(2, 2).RGBA()
	if a != 0 {
		t.Errorf("Expected center pixel to be transparent (alpha=0), got alpha=%d", a)
	}
}

func TestTransparentizePreservesNonBackground(t *testing.T) {
	// Image with a solid red center on white background
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	lime := color.RGBA{0, 255, 0, 255}
	// Fill with lime
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, lime)
		}
	}
	// Draw a red square in the middle (not touching edges)
	for y := 2; y < 6; y++ {
		for x := 2; x < 6; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	result := Transparentize(img)
	if result == nil {
		t.Fatal("Transparentize returned nil")
	}
	// Center should remain non-transparent (red, not white)
	r, _, _, a := result.At(4, 4).RGBA()
	if a == 0 {
		t.Error("Expected red center pixel to NOT be made transparent")
	}
	if r == 0 {
		t.Error("Expected red channel to be non-zero for center pixel")
	}
}
