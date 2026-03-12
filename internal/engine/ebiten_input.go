//go:build !test

package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// EbitenInput implements engine.Input using the actual ebiten package.
type EbitenInput struct{}

func (e *EbitenInput) IsKeyPressed(key Key) bool {
	return ebiten.IsKeyPressed(toEbitenKey(key))
}

func (e *EbitenInput) IsKeyJustPressed(key Key) bool {
	return inpututil.IsKeyJustPressed(toEbitenKey(key))
}

func (e *EbitenInput) AppendJustPressedKeys(keys []Key) []Key {
	ebKeys := inpututil.AppendJustPressedKeys(nil)
	for _, k := range ebKeys {
		ek := fromEbitenKey(k)
		if ek != -1 {
			keys = append(keys, ek)
		}
	}
	return keys
}

func (e *EbitenInput) AppendInputChars(chars []rune) []rune {
	return ebiten.AppendInputChars(chars)
}

func (e *EbitenInput) MousePosition() (x, y int) {
	return ebiten.CursorPosition()
}

func (e *EbitenInput) IsMouseButtonPressed(button MouseButton) bool {
	return ebiten.IsMouseButtonPressed(toEbitenMouseButton(button))
}

func (e *EbitenInput) IsMouseButtonJustPressed(button MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(toEbitenMouseButton(button))
}

func (e *EbitenInput) Wheel() (x, y float64) {
	return ebiten.Wheel()
}

func toEbitenMouseButton(button MouseButton) ebiten.MouseButton {
	switch button {
	case MouseButtonLeft:
		return ebiten.MouseButtonLeft
	case MouseButtonRight:
		return ebiten.MouseButtonRight
	}
	return ebiten.MouseButtonLeft
}

func NewEbitenInput() *EbitenInput {
	return &EbitenInput{}
}

func toEbitenKey(key Key) ebiten.Key {
	switch key {
	case KeyW: return ebiten.KeyW
	case KeyA: return ebiten.KeyA
	case KeyS: return ebiten.KeyS
	case KeyD: return ebiten.KeyD
	case KeySpace: return ebiten.KeySpace
	case KeyEscape: return ebiten.KeyEscape
	case KeyEnter: return ebiten.KeyEnter
	case KeyUp: return ebiten.KeyUp
	case KeyDown: return ebiten.KeyDown
	case KeyLeft: return ebiten.KeyLeft
	case KeyRight: return ebiten.KeyRight
	case KeyF9: return ebiten.KeyF9
	case KeyTab: return ebiten.KeyTab
	case KeyQ: return ebiten.KeyQ
	case KeyControl: return ebiten.KeyControl
	case KeyMeta: return ebiten.KeyMeta
	case KeyShift: return ebiten.KeyShift
	case KeyBackspace: return ebiten.KeyBackspace
	case KeyDelete: return ebiten.KeyDelete
	}
	return -1
}

func fromEbitenKey(key ebiten.Key) Key {
	switch key {
	case ebiten.KeyW: return KeyW
	case ebiten.KeyA: return KeyA
	case ebiten.KeyS: return KeyS
	case ebiten.KeyD: return KeyD
	case ebiten.KeySpace: return KeySpace
	case ebiten.KeyEscape: return KeyEscape
	case ebiten.KeyEnter: return KeyEnter
	case ebiten.KeyUp: return KeyUp
	case ebiten.KeyDown: return KeyDown
	case ebiten.KeyLeft: return KeyLeft
	case ebiten.KeyRight: return KeyRight
	case ebiten.KeyF9: return KeyF9
	case ebiten.KeyTab: return KeyTab
	case ebiten.KeyQ: return KeyQ
	case ebiten.KeyControl: return KeyControl
	case ebiten.KeyMeta: return KeyMeta
	case ebiten.KeyShift: return KeyShift
	case ebiten.KeyBackspace: return KeyBackspace
	case ebiten.KeyDelete: return KeyDelete
	}
	return -1
}
