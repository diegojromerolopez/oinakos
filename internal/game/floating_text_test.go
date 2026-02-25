package game

import (
	"image/color"
	"math"
	"oinakos/internal/engine"
	"testing"
)

func TestFloatingTextUpdate_Alive(t *testing.T) {
	ft := &FloatingText{Text: "5", X: 0, Y: 0, Life: 10, Color: color.RGBA{255, 0, 0, 255}}
	alive := ft.Update()
	if !alive {
		t.Error("FloatingText with Life=10 should still be alive after one Update")
	}
	if ft.Life != 9 {
		t.Errorf("Life after Update: got %d, want 9", ft.Life)
	}
	if ft.Y >= 0 {
		t.Errorf("Y should have decreased (drifted up): got %v", ft.Y)
	}
}

func TestFloatingTextUpdate_Dies(t *testing.T) {
	ft := &FloatingText{Text: "X", X: 0, Y: 0, Life: 1, Color: color.RGBA{255, 0, 0, 255}}
	alive := ft.Update() // Life goes from 1 to 0
	if alive {
		t.Error("FloatingText with Life=1 should be dead after Update")
	}
}

func TestFloatingTextUpdate_AlreadyDead(t *testing.T) {
	ft := &FloatingText{Text: "X", Life: 0, Color: color.RGBA{255, 0, 0, 255}}
	alive := ft.Update() // Life becomes -1 → still not > 0
	if alive {
		t.Error("FloatingText with Life≤0 should return false")
	}
}

func TestCartesianToIso_LocalFunction(t *testing.T) {
	// The game package has its own CartesianToIso (duplicated from engine)
	// (x-y)*32, (x+y)*16
	cases := []struct{ x, y, wantX, wantY float64 }{
		{0, 0, 0, 0},
		{1, 0, 32, 16},
		{0, 1, -32, 16},
		{1, 1, 0, 32},
	}
	for _, tc := range cases {
		gx, gy := engine.CartesianToIso(tc.x, tc.y)
		if math.Abs(gx-tc.wantX) > 0.001 || math.Abs(gy-tc.wantY) > 0.001 {
			t.Errorf("CartesianToIso(%v,%v): got (%v,%v), want (%v,%v)",
				tc.x, tc.y, gx, gy, tc.wantX, tc.wantY)
		}
	}
}
