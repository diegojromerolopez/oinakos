package game

import (
	"image/color"
	"oinakos/internal/engine"
)

func (o *Obstacle) Draw(screen engine.Image, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64) {
	if screen == nil || o.Archetype == nil || o.Archetype.Image == nil || !o.Alive {
		return
	}

	isoX, isoY := engine.CartesianToIso(o.X, o.Y)

	op := engine.NewDrawImageOptions()
	scale := 1.0

	img, ok := o.Archetype.Image.(engine.Image)
	if !ok || img == nil {
		return
	}
	sw, sh := img.Size()

	// Pivot point for isometric depth
	pivotX := float64(sw) * scale / 2
	pivotY := float64(sh) * scale * 0.85

	op.Scale(scale, scale)
	op.Translate(isoX+offsetX-pivotX, isoY+offsetY-pivotY)

	screen.DrawImage(img, op)

	// Draw Cooldown Bar for wells
	if o.Archetype.Type == "well" && o.CooldownTicks > 0 {
		maxTicks := float64(o.Archetype.CooldownTime * 60 * 60) // CooldownTime is in minutes
		if maxTicks > 0 {
			barWidth := 30.0
			barHeight := 4.0
			bx := isoX + offsetX - barWidth/2
			by := isoY + offsetY + 15

			if vectorRenderer != nil {
				// Background
				vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth), float32(barHeight), color.RGBA{50, 50, 50, 200}, true)

				// Fill (inverse, builds up as it replies)
				remainingRatio := float64(o.CooldownTicks) / maxTicks
				fillRatio := 1.0 - remainingRatio
				vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth*fillRatio), float32(barHeight), color.RGBA{0, 200, 255, 200}, true)
			}
		}
	}
}
