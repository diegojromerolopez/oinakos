package game

import (
	"oinakos/internal/engine"
	"sync/atomic"
	"testing"
)

func TestGameRenderer_Mega(t *testing.T) {
	graphics := &engine.MockGraphics{}
	input := NewMockInputManager()
	g := NewGame(nil, graphics, "", "", "", input, &MockAudioManager{}, true, "0.1-test")
	
	// Create characters and objects to trigger more draw logic
	mc := NewCharacter(0, 0, &EntityConfig{Name: "Hero", Description: "A brave hero"}, 1, true, g.Registries.Objects)
	g.playableCharacter = mc
	
	npc := NewCharacter(5, 5, &EntityConfig{Name: "Villager", Description: "A nice villager"}, 1, false, g.Registries.Objects)
	g.characters = []*Character{npc}
	
	obs := NewObstacle("tree1", 2, 2, &ObstacleArchetype{ID: "tree", Name: "Oak", Type: TypeTree, HealthPoints: 100})
	g.obstacles = []*Obstacle{obs}
	
	item := NewItemInstance("sword", &ObjectConfig{Name: "Iron Sword", Description: "Sharp"}, 3, 3)
	g.World.Items = []*ItemInstance{item}
	
	g.currentMapType = MapType{Name: "Test Map", Description: "A place for testing"}
	
	gr := NewGameRenderer(g, nil, graphics)
	screen := graphics.NewImage(1280, 720)

	// 1. Loading screen
	atomic.StoreInt32(&g.LoadingProgress, 500)
	gr.Draw(screen)
	atomic.StoreInt32(&g.LoadingProgress, 1000)

	// 2. Main Menu
	g.isMainMenu = true
	gr.Draw(screen)
	g.isMainMenu = false

	// 3. Character Select
	g.isCharacterSelect = true
	g.characterRegistry.Characters["hero1"] = &EntityConfig{Name: "Hero 1", Description: "First hero", ID: "hero1"}
	g.characterRegistry.IDs = []string{"hero1"}
	gr.Draw(screen)
	g.isCharacterSelect = false

	// 4. About Screen
	g.isAboutScreen = true
	gr.Draw(screen)
	g.isAboutScreen = false

	// 5. Settings Screen
	g.isSettingsScreen = true
	gr.Draw(screen)
	g.isSettingsScreen = false

	// 6. Keymap Screen
	g.isKeymapScreen = true
	gr.Draw(screen)
	g.isKeymapScreen = false

	// 7. Inventory
	g.isInventoryOpen = true
	mc.Inventory = []*ItemInstance{item}
	mc.Slots = map[string]*ItemInstance{"weapon": item}
	gr.Draw(screen)
	g.isInventoryOpen = false

	// 8. Dialogue
	g.ActiveDialogue = &DialogueState{
		SpeakerNPC:  npc, 
		CurrentText: "Hello there!", 
		Choices: []Choice{{Text: "Hi"}},
		UIState: DialogueMaximized,
	}
	gr.Draw(screen)
	g.ActiveDialogue.UIState = DialogueMinimized
	gr.Draw(screen)
	g.ActiveDialogue = nil

	// 9. Map Won / Game Over / Game Won
	g.isMapWon = true
	gr.Draw(screen)
	g.isMapWon = false
	
	g.isGameOver = true
	gr.Draw(screen)
	g.isGameOver = false
	
	g.isGameWon = true
	gr.Draw(screen)
	g.isGameWon = false

	// 10. Pause Menu
	g.isPaused = true
	gr.Draw(screen)
	g.isPaused = false

	// 11. Quit Confirmation
	g.isQuitConfirmationOpen = true
	gr.Draw(screen)
	g.isQuitConfirmationOpen = false

	// 12. Book Overlay
	g.isInventoryOpen = true
	g.ActiveBook = item
	gr.Draw(screen)
	g.ActiveBook = nil
	g.isInventoryOpen = false

	// 13. Weather & Lighting
	g.World.State.Weather = WeatherStorm
	g.Tick = 0 // Flash frame
	gr.Draw(screen)
	
	// 14. Hover Info
	g.camera.SnapTo(0, 0)
	npc.X, npc.Y = 0, 0
	// Iso(0,0) is center of screen if camera is at (0,0)
	input.MouseX, input.MouseY = 640, 360-40 // -40 for anchor offset in drawHoverInfo
	gr.Draw(screen)

	// 15. Dialogue Log Scroll
	g.EventLog = []LogEntry{{Text: "Log 1"}, {Text: "Log 2"}, {Text: "Log 3"}}
	g.LogUIState = DialogueMaximized
	gr.Draw(screen)

	// 16. Load Assets (dummy)
	gr.LoadAssets(nil)
}
