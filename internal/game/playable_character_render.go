package game

import (
	"oinakos/internal/engine"
)

func (pc *PlayableCharacter) Draw(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64) {
	// DrawActor takes care of alignment indicators, names, and sprite selection.
	// Palette shaders can be passed if supported.
	DrawActor(&pc.Actor, screen, textRenderer, vectorRenderer, nil, offsetX, offsetY, true)
}
