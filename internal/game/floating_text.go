package game

import (
	"image/color"
)

var (
	ColorHarm = color.RGBA{220, 20, 60, 255}   // Crimson
	ColorHeal = color.RGBA{0, 255, 0, 255}     // Green
	ColorMiss = color.RGBA{200, 200, 200, 255} // Gray
)

type FloatingText struct {
	Text  string
	X, Y  float64 // Cartesian coordinates
	Life  int     // Frames remaining
	Color color.Color
}

func (ft *FloatingText) Update() bool {
	ft.Life--
	// Move strictly upwards in Isometric space (requires decreasing both X and Y in Cartesian)
	ft.X -= 0.005
	ft.Y -= 0.005
	return ft.Life > 0
}
