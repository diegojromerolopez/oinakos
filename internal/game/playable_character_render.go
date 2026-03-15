package game

import (
	"oinakos/internal/engine"
)

func (pc *PlayableCharacter) Draw(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64) {
	// DrawActor takes care of sprite selection and procedural animation.
	DrawActor(&pc.Actor, screen, textRenderer, vectorRenderer, nil, offsetX, offsetY, true)
}

func (pc *PlayableCharacter) DrawUI(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64) {
	DrawActorUI(&pc.Actor, screen, textRenderer, vectorRenderer, offsetX, offsetY, true)
}
