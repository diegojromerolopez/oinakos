package game

import (
	"oinakos/internal/engine"
)

// NewTestContext creates a SystemContext populated with mocks for testing.
func NewTestContext() *SystemContext {
	world := NewWorld()
	// Initialize with dummy values
	world.CurrentMapType = &MapType{MapWidth: 1000, MapHeight: 1000}
	
	return &SystemContext{
		World:      world,
		Input:      NewMockInputManager(),
		Audio:      NewMockAudioManager(),
		Registries: &RegistryContainer{
			Archetypes: NewArchetypeRegistry(),
			NPCs:       NewNPCRegistry(),
			Obstacles:  NewObstacleRegistry(),
		},
		Log: func(s string, c LogCategory) {},
	}
}

// MockInput is sometimes used in specific tests, let's keep it here if needed
// but favor MockInputManager from input.go
type GenericMockInput struct {
	engine.Input
	PressedKeys []engine.Key
}

func (m *GenericMockInput) IsKeyPressed(k engine.Key) bool {
	for _, pk := range m.PressedKeys {
		if pk == k { return true }
	}
	return false
}
func (m *GenericMockInput) IsKeyJustPressed(k engine.Key) bool { return false }
func (m *GenericMockInput) MousePosition() (int, int) { return 0, 0 }
func (m *GenericMockInput) IsMouseButtonPressed(engine.MouseButton) bool { return false }
func (m *GenericMockInput) IsMouseButtonJustPressed(engine.MouseButton) bool { return false }
func (m *GenericMockInput) Wheel() (float64, float64) { return 0, 0 }
func (m *GenericMockInput) AppendInputChars(r []rune) []rune { return r }
func (m *GenericMockInput) AppendJustPressedKeys(k []engine.Key) []engine.Key { return k }

func init() {
	isTestingEnvironment = true
}
