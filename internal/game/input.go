package game

import (
	"oinakos/internal/engine"
)

// InputManager defines an interface for all input operations to allow mocking.
type InputManager interface {
	engine.Input
}

// MockInputManager can be used in headless tests.
type MockInputManager struct {
	PressedKeys     map[engine.Key]bool
	JustPressedKeys map[engine.Key]bool
	MouseX, MouseY  int
	PressedButtons  map[engine.MouseButton]bool
	JustPressedButtons map[engine.MouseButton]bool
	WheelX, WheelY  float64
}

func NewMockInputManager() *MockInputManager {
	return &MockInputManager{
		PressedKeys:        make(map[engine.Key]bool),
		JustPressedKeys:    make(map[engine.Key]bool),
		PressedButtons:     make(map[engine.MouseButton]bool),
		JustPressedButtons: make(map[engine.MouseButton]bool),
	}
}

func (m *MockInputManager) IsKeyPressed(key engine.Key) bool {
	return m.PressedKeys[key]
}

func (m *MockInputManager) IsKeyJustPressed(key engine.Key) bool {
	return m.JustPressedKeys[key]
}

func (m *MockInputManager) AppendJustPressedKeys(keys []engine.Key) []engine.Key {
	for k, v := range m.JustPressedKeys {
		if v {
			keys = append(keys, k)
		}
	}
	return keys
}

func (m *MockInputManager) MousePosition() (x, y int) {
	return m.MouseX, m.MouseY
}

func (m *MockInputManager) AppendInputChars(chars []rune) []rune {
	return chars
}

func (m *MockInputManager) IsMouseButtonPressed(button engine.MouseButton) bool {
	return m.PressedButtons[button]
}

func (m *MockInputManager) IsMouseButtonJustPressed(button engine.MouseButton) bool {
	return m.JustPressedButtons[button]
}

func (m *MockInputManager) Wheel() (x, y float64) {
	return m.WheelX, m.WheelY
}
