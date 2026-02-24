package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

func (ft *FloatingText) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	// Transition to iso for drawing
	// We'll just draw it above the entity it spawned from.
	// Since entities are at Cartesian (X, Y), we use that.
	// Convert to Iso
	isoX, isoY := CartesianToIso(ft.X, ft.Y)

	// Fade out based on life
	alpha := uint8(255)
	if ft.Life < 20 {
		alpha = uint8(float64(ft.Life) / 20.0 * 255.0)
	}

	c := ft.Color.(color.RGBA)
	c.A = alpha

	ebitenutil.DebugPrintAt(screen, ft.Text, int(isoX+offsetX), int(isoY+offsetY))
}

func CartesianToIso(x, y float64) (float64, float64) {
	return (x - y) * 32, (x + y) * 16
}
