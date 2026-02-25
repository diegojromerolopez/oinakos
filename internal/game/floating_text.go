package game

import (
	"image/color"
)

type FloatingText struct {
	Text  string
	X, Y  float64 // Cartesian coordinates
	Life  int     // Frames remaining
	Color color.Color
}

func (ft *FloatingText) Update() bool {
	ft.Life--
	ft.Y -= 0.01 // Float upwards in Cartesian (which translates to up-ish in Iso)
	return ft.Life > 0
}
