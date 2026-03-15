package game

import (
	"oinakos/internal/engine"
)

func (n *NPC) Draw(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64) {
	// DrawActor handles sprite selection, procedural animation.
	DrawActor(&n.Actor, screen, textRenderer, vectorRenderer, paletteShader, offsetX, offsetY, false)
}

func (n *NPC) DrawUI(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64) {
	DrawActorUI(&n.Actor, screen, textRenderer, vectorRenderer, offsetX, offsetY, false)
}
