package game

import (
	"image/color"
	"oinakos/internal/engine"
)

func (ft *FloatingText) Draw(screen engine.Image, textRenderer engine.TextRenderer, offsetX, offsetY float64) {
	if screen == nil {
		return
	}
	// Convert to Iso
	isoX, isoY := engine.CartesianToIso(ft.X, ft.Y)

	// Fade out based on life
	alpha := uint8(255)
	if ft.Life < 20 {
		alpha = uint8(float64(ft.Life) / 20.0 * 255.0)
	}

	c := color.RGBAModel.Convert(ft.Color).(color.RGBA)
	c.A = alpha

	if textRenderer != nil {
		textRenderer.DrawTextAt(screen, ft.Text, int(isoX+offsetX), int(isoY+offsetY), c, 16)
	}
}
