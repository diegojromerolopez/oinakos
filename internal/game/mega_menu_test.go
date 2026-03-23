package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestMenu_UpdateMega(t *testing.T) {
	input := NewMockInputManager()
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", input, &MockAudioManager{}, false, "0.1-test")
	mh := g.menuHandler
	
	g.width, g.height = 1200, 800
	
	// 1. Main Menu
	g.isMainMenu = true
	input.JustPressedKeys[engine.KeyDown] = true
	mh.Update()
	if g.mainMenuIndex != 1 { t.Error("Failed to navigate down in main menu") }
	input.JustPressedKeys[engine.KeyDown] = false
	
	input.JustPressedKeys[engine.KeyEnter] = true
	mh.Update() // Select LOAD (index 1)
	input.JustPressedKeys[engine.KeyEnter] = false
	
	// 2. Character Select
	g.isMainMenu = false
	g.isCharacterSelect = true
	g.characterRegistry.IDs = []string{"id1", "id2"}
	input.JustPressedKeys[engine.KeyDown] = true
	mh.Update()
	input.JustPressedKeys[engine.KeyDown] = false
	
	// 3. Settings
	g.isCharacterSelect = false
	g.isSettingsScreen = true
	input.JustPressedKeys[engine.KeyDown] = true
	mh.Update()
	input.JustPressedKeys[engine.KeyDown] = false
	
	// 4. About
	g.isSettingsScreen = false
	g.isAboutScreen = true
	input.JustPressedKeys[engine.KeyEscape] = true
	mh.Update()
	input.JustPressedKeys[engine.KeyEscape] = false
	
	// 5. Campaign Select
	g.isCampaignSelect = true
	g.campaignRegistry.IDs = []string{"cam1"}
	input.JustPressedKeys[engine.KeyEnter] = true
	mh.Update()
	input.JustPressedKeys[engine.KeyEnter] = false
	
	// 6. Pause Menu
	g.isMenuOpen = true
	input.JustPressedKeys[engine.KeyUp] = true
	mh.Update()
	input.JustPressedKeys[engine.KeyUp] = false
	
	// 7. Inventory
	g.isInventoryOpen = true
	input.JustPressedKeys[engine.KeyEscape] = true
	mh.Update()
	input.JustPressedKeys[engine.KeyEscape] = false

	// 8. Quit Confirmation
	g.isQuitConfirmationOpen = true
	input.JustPressedKeys[engine.KeyLeft] = true
	mh.Update()
	input.JustPressedKeys[engine.KeyLeft] = false
}
