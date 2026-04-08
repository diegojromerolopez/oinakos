package game

import (
	"image/color"
	"oinakos/internal/engine"
)

// DrawAlignmentIndicator draws an isometric ellipse under the feet of an entity.
func DrawAlignmentIndicator(screen engine.Image, vectorRenderer engine.VectorRenderer, a *Actor, offsetX, offsetY float64, isOccluded bool) {
	if a == nil || !a.IsAlive() || vectorRenderer == nil {
		return
	}

	isoX, isoY := engine.CartesianToIso(a.X, a.Y)

	// Draw universal shadow first (black sub-ellipse)
	sw, sh := 25.0, 12.0
	iw, ih := 30.0, 15.0
	
	// Small animals get smaller shadows
	if a.Config != nil && a.Config.IsAnimal {
		sw, sh = 15.0, 8.0
		iw, ih = 20.0, 10.0
	}

	vectorRenderer.DrawEllipse(screen, float32(isoX+offsetX), float32(isoY+offsetY), float32(sw), float32(sh), color.RGBA{0, 0, 0, 80}, 1, true)

	var clr color.Color
	switch a.Alignment {
	case AlignmentAlly:
		clr = ColorAlly
	case AlignmentEnemy:
		clr = ColorEnemy
	case AlignmentNeutral:
		clr = ColorNeutral
	default:
		clr = color.RGBA{150, 150, 150, 150}
	}

	// Vertical radius is half of horizontal to match isometric perspective
	vectorRenderer.DrawEllipse(screen, float32(isoX+offsetX), float32(isoY+offsetY), float32(iw), float32(ih), clr, 1, true)
}
