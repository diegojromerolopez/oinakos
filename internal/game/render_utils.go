package game

import (
	"image/color"
	"oinakos/internal/engine"
)

// DrawAlignmentIndicator draws an isometric ellipse under the feet of an entity.
func DrawAlignmentIndicator(screen engine.Image, vectorRenderer engine.VectorRenderer, x, y, offsetX, offsetY float64, alignment Alignment, isAlive bool) {
	if !isAlive || vectorRenderer == nil {
		return
	}

	isoX, isoY := engine.CartesianToIso(x, y)

	var clr color.Color
	switch alignment {
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
	vectorRenderer.DrawEllipse(screen, float32(isoX+offsetX), float32(isoY+offsetY), 30, 15, clr, 1, true)
}
