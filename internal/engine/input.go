package engine

import "github.com/hajimehoshi/ebiten/v2"

// Input defines an interface for all input operations to allow mocking.
type Input interface {
	IsKeyPressed(key ebiten.Key) bool
	IsKeyJustPressed(key ebiten.Key) bool
	AppendJustPressedKeys(keys []ebiten.Key) []ebiten.Key
}
