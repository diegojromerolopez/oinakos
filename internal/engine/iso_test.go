package engine

import (
	"math"
	"testing"
)

func TestCartesianToIso(t *testing.T) {
	tests := []struct {
		x, y   float64
		wantX, wantY float64
	}{
		{0, 0, 0, 0},
		{1, 0, 32, 16},
		{0, 1, -32, 16},
		{1, 1, 0, 32},
	}

	for _, tt := range tests {
		gotX, gotY := CartesianToIso(tt.x, tt.y)
		if math.Abs(gotX-tt.wantX) > 0.001 || math.Abs(gotY-tt.wantY) > 0.001 {
			t.Errorf("CartesianToIso(%v, %v) = (%v, %v); want (%v, %v)", tt.x, tt.y, gotX, gotY, tt.wantX, tt.wantY)
		}
	}
}

func TestIsoToCartesian(t *testing.T) {
	tests := []struct {
		screenX, screenY float64
		wantX, wantY     float64
	}{
		{0, 0, 0, 0},
		{32, 16, 1, 0},
		{-32, 16, 0, 1},
		{0, 32, 1, 1},
	}

	for _, tt := range tests {
		gotX, gotY := IsoToCartesian(tt.screenX, tt.screenY)
		if math.Abs(gotX-tt.wantX) > 0.001 || math.Abs(gotY-tt.wantY) > 0.001 {
			t.Errorf("IsoToCartesian(%v, %v) = (%v, %v); want (%v, %v)", tt.screenX, tt.screenY, gotX, gotY, tt.wantX, tt.wantY)
		}
	}
}
