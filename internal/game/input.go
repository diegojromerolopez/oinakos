package game

import (
	"oinakos/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
)

// InputManager defines an interface for all input operations to allow mocking.
type InputManager interface {
	engine.Input
}

// MockInputManager can be used in headless tests.
type MockInputManager struct {
	PressedKeys     map[ebiten.Key]bool
	JustPressedKeys map[ebiten.Key]bool
}

func NewMockInputManager() *MockInputManager {
	return &MockInputManager{
		PressedKeys:     make(map[ebiten.Key]bool),
		JustPressedKeys: make(map[ebiten.Key]bool),
	}
}

func (m *MockInputManager) IsKeyPressed(key ebiten.Key) bool {
	return m.PressedKeys[key]
}

func (m *MockInputManager) IsKeyJustPressed(key ebiten.Key) bool {
	return m.JustPressedKeys[key]
}

func (m *MockInputManager) AppendJustPressedKeys(keys []ebiten.Key) []ebiten.Key {
	for k, v := range m.JustPressedKeys {
		if v {
			keys = append(keys, k)
		}
	}
	return keys
}
