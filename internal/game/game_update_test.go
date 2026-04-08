package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestGameUpdate_HUDButtons(t *testing.T) {
	input := NewMockInputManager()
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", input, &MockAudioManager{}, false, "0.1-test")
	
	g.width, g.height = 1280, 720
	g.LoadingProgress = 1000 // Finished loading
	g.isMainMenu = false     // Disable main menu to reach HUD click logic

	// 1. Menu Button Click
	// Coordinates: (g.width-110, 20) to (g.width-10, 50) => (1170, 20) to (1270, 50)
	input.MouseX, input.MouseY = 1200, 30
	input.JustPressedButtons[engine.MouseButtonLeft] = true
	g.Update()
	if !g.isMenuOpen {
		t.Error("Expected Menu to open after clicking Menu button")
	}
	g.isMenuOpen = false
	input.JustPressedButtons[engine.MouseButtonLeft] = false

	// 2. Map Selection Click (in Main Menu)
	g.isMainMenu = true
	// Main menu has choices. "Select Character" (0) and "New Sandbox" (1) and "Settings" (2) etc.
	// Click "New Sandbox" (index 1)
	// menu_handler.go usually handles this.
	// Let's check menu_handler.go instead.

	// 3. Tab Key (Debug Toggle)
	input.JustPressedKeys[engine.KeyTab] = true
	g.Update()
	if !g.showBoundaries {
		t.Error("Expected showBoundaries to be true after pressing Tab")
	}
	input.JustPressedKeys[engine.KeyTab] = false

	// 4. I Key (Inventory Toggle)
	g.isMainMenu = false
	input.JustPressedKeys[engine.KeyI] = true
	g.Update()
	if !g.isInventoryOpen {
		t.Error("Expected Inventory to open after pressing I")
	}
	input.JustPressedKeys[engine.KeyI] = false
}

func TestGameRenderer_BasicDraw(t *testing.T) {
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "conde_olinos", NewMockInputManager(), &MockAudioManager{}, true, "0.1-test")
	g.LoadingProgress = 1000
	g.isMainMenu = false
	
	// Setup dialogue
	g.ActiveDialogue = &DialogueState{IsActive: true, UIState: DialogueMaximized, CurrentText: "Hello world"}
	
	gr := NewGameRenderer(g, nil, &engine.MockGraphics{})
	screen := engine.NewMockImage(1280, 720)
	
	// Test different states for coverage
	states := []func(){
		func() { g.isMainMenu = true },
		func() { g.isMainMenu = false; g.isCharacterSelect = true },
		func() { g.isCharacterSelect = false; g.isMapWon = true },
		func() { g.isMapWon = false; g.isGameOver = true },
		func() { g.isGameOver = false; g.isInventoryOpen = true },
		func() { g.isInventoryOpen = false; g.isPaused = true },
	}

	for _, stateFn := range states {
		stateFn()
		gr.Draw(screen)
	}

	// Weather
	g.World.State.Weather = WeatherRain
	g.World.State.Intensity = 0.5
	gr.Draw(screen)
}
