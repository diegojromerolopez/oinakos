package engine

// Key represents a keyboard key in an engine-neutral way.
type Key int

// Keyboard keys
const (
	KeyW Key = iota
	KeyA
	KeyS
	KeyD
	KeySpace
	KeyEscape
	KeyEnter
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyF9
	KeyTab
	KeyQ
	KeyControl
	KeyMeta
	KeyShift
	KeyBackspace
	KeyDelete
)

// Input defines an interface for all input operations to allow mocking.
type Input interface {
	IsKeyPressed(key Key) bool
	IsKeyJustPressed(key Key) bool
	AppendJustPressedKeys(keys []Key) []Key
	AppendInputChars(chars []rune) []rune
	MousePosition() (x, y int)
	IsMouseButtonPressed(button MouseButton) bool
	IsMouseButtonJustPressed(button MouseButton) bool
	Wheel() (x, y float64)
}

type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
)
