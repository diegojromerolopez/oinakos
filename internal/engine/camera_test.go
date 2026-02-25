package engine

import (
	"math"
	"testing"
)

func TestNewCamera(t *testing.T) {
	c := NewCamera(10, 20)
	if c.X != 10 || c.Y != 20 {
		t.Errorf("NewCamera(10, 20): got (%v, %v), want (10, 20)", c.X, c.Y)
	}
}

func TestCameraFollow(t *testing.T) {
	c := NewCamera(0, 0)
	c.Follow(100, 100, 1.0) // lerp=1 → snaps immediately
	if math.Abs(c.X-100) > 0.001 || math.Abs(c.Y-100) > 0.001 {
		t.Errorf("Follow with lerp=1: got (%v, %v), want (100, 100)", c.X, c.Y)
	}

	// Lerp=0.5 should move halfway
	c2 := NewCamera(0, 0)
	c2.Follow(100, 100, 0.5)
	if math.Abs(c2.X-50) > 0.001 || math.Abs(c2.Y-50) > 0.001 {
		t.Errorf("Follow with lerp=0.5: got (%v, %v), want (50, 50)", c2.X, c2.Y)
	}

	// Lerp=0 should not move
	c3 := NewCamera(5, 7)
	c3.Follow(100, 100, 0)
	if c3.X != 5 || c3.Y != 7 {
		t.Errorf("Follow with lerp=0: expected no movement, got (%v, %v)", c3.X, c3.Y)
	}
}

func TestCameraSnapTo(t *testing.T) {
	c := NewCamera(0, 0)
	c.SnapTo(42, 99)
	if c.X != 42 || c.Y != 99 {
		t.Errorf("SnapTo(42, 99): got (%v, %v)", c.X, c.Y)
	}
}

func TestCameraGetOffsets(t *testing.T) {
	c := NewCamera(100, 50)
	offX, offY := c.GetOffsets(1280, 720)
	// offX = 1280/2 - 100 = 540
	// offY = 720/2  - 50  = 310
	if math.Abs(offX-540) > 0.001 {
		t.Errorf("GetOffsets offX: got %v, want 540", offX)
	}
	if math.Abs(offY-310) > 0.001 {
		t.Errorf("GetOffsets offY: got %v, want 310", offY)
	}

	// Camera at (0,0) — offsets equal half-screen
	c2 := NewCamera(0, 0)
	ox, oy := c2.GetOffsets(800, 600)
	if ox != 400 || oy != 300 {
		t.Errorf("GetOffsets at origin: got (%v, %v), want (400, 300)", ox, oy)
	}
}
