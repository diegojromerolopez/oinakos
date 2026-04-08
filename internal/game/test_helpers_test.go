package game

import (
	"image"
	"oinakos/internal/engine"
	"os"
)

// NewTestContext creates a SystemContext populated with mocks for testing.
func NewTestContext() *SystemContext {
	g := setupTestGame()
	world := g.World
	// Initialize with dummy values
	world.CurrentMapType = &MapType{MapWidth: 1000, MapHeight: 1000}
	
	return &SystemContext{
		World:      world,
		Input:      g.input,
		Audio:      g.audio,
		Registries: g.Registries,
		Log:        func(s string, c LogCategory) {},
		Settings:   g.settings,
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
func (m *GenericMockInput) SetCursorMode(mode engine.CursorMode) {}

func setupTestGame() *Game {
	archetypeRegistry := NewArchetypeRegistry()
	characterRegistry := NewCharacterRegistry()
	obstacleRegistry := NewObstacleRegistry()
	objectRegistry := NewObjectRegistry()

	g := &Game{
		World: &World{
			Items:         []*ItemInstance{},
			Characters:    []*Character{},
			Obstacles:     []*Obstacle{},
			FloatingTexts: []*FloatingText{},
			ExploredTiles: make(map[image.Point]bool),
		},
		ExploredTiles: make(map[image.Point]bool),
		currentMapType: MapType{MapWidth: 500, MapHeight: 500},
		camera: engine.NewCamera(800, 600),
		mapTypeRegistry: NewMapTypeRegistry(),
		archetypeRegistry: archetypeRegistry,
		characterRegistry: characterRegistry,
		obstacleRegistry: obstacleRegistry,
		Registries: &RegistryContainer{
			Archetypes: archetypeRegistry,
			Characters: characterRegistry,
			Obstacles:  obstacleRegistry,
			Objects:    objectRegistry,
		},
		settings: DefaultSettings(),
	}
	g.World.Game = g
	g.World.CurrentMapType = &g.currentMapType
	g.input = NewMockInputManager()
	g.audio = NewMockAudioManager()
	g.mechanicsManager = NewMechanicsManager(g)
	g.worldManager = NewWorldManager(g)
	g.menuHandler = NewMenuHandler(g)
	g.obstacles = g.World.Obstacles
	pConfig := &EntityConfig{
		MaxItems: 10,
		Stats: EntityStatsConfig{
			HealthPoints: IntInterval{Min: 100, Max: 100},
			Speed: FloatInterval{Min: 0.1, Max: 0.1}, BaseAttack: IntInterval{Min: 10, Max: 10}, BaseDefense: IntInterval{Min: 5, Max: 5},
		},
		Attributes: PrimaryAttributeConfig{
			Strength: IntInterval{Min: 100, Max: 100}, Dexterity: IntInterval{Min: 100, Max: 100}, Health: IntInterval{Min: 100, Max: 100}, Intellect: IntInterval{Min: 100, Max: 100}, Wisdom: IntInterval{Min: 100, Max: 100},
		},
	}
	g.playableCharacter = NewCharacter(0, 0, pConfig, 1, true, objectRegistry)
	g.World.PlayableCharacter = g.playableCharacter
	g.characters = []*Character{g.playableCharacter}
	g.World.Characters = g.characters

	return g
}

func init() {
	InTestMode = true
	isTestingEnvironment = true
	tmpDir, _ := os.MkdirTemp("", "oinakos-test")
	SetOinakosDir(tmpDir)
}
