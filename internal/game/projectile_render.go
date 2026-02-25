package game

import (
	"image/color"
	"oinakos/internal/engine"
)

func (p *Projectile) Draw(screen engine.Image, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64) {
	if screen == nil || !p.Alive {
		return
	}
	isoX, isoY := engine.CartesianToIso(p.X, p.Y)

	// Since we don't have a global asset manager for projectiles yet,
	// we'll use procedural drawing or a placeholder for now.
	// In a real game, this would be loaded by GameRenderer.
	if isTestingEnvironment {
		return
	}

	// Fallback simple red circle using vectorRenderer
	if vectorRenderer != nil {
		vectorRenderer.DrawFilledCircle(screen, float32(isoX+offsetX), float32(isoY+offsetY-10), 3, color.RGBA{255, 50, 50, 255}, true)
	}
}

var isTestingEnvironment = false
