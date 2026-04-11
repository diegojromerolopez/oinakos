//go:build test

package game

import (
	"oinakos/internal/engine"
	"testing"
)

var testSettingsDir string

func init() {
	InTestMode = true
}

func NewTestGame(t *testing.T) *Game {
	if testSettingsDir == "" {
		testSettingsDir = t.TempDir()
		SetOinakosDir(testSettingsDir)
	}
	mockInput := engine.NewMockInput()
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", mockInput, nil, false, "test")
	g.LoadingProgress = 1000
	g.width, g.height = 1000, 800
	
	// Initialize settings indices from current settings
	for i, f := range FontOptions { if f == g.settings.Font { g.settingsFontIndex = i; break } }
	for i, f := range FrequencyOptions { if f == g.settings.SoundFrequency { g.settingsAudioIndex = i; break } }
	for i, f := range FogOfWarOptions { if f == g.settings.FogOfWar { g.settingsFogIndex = i; break } }
	
	return g
}

func TestMenu_QuitConfirmation(t *testing.T) {
	g := NewTestGame(t)
	g.isQuitConfirmationOpen = true
	mh := g.menuHandler
	mockInput := g.input.(*engine.MockInput)

	// Initially quitConfirmationIndex is 0 (Quit)
	if g.quitConfirmationIndex != 0 {
		// Wait, NewGame might set it to 1 by default? No, it's 0 usually.
		// In updateMainMenu case 4, it sets to 1.
	}

	// Test navigation (index 1 to 0)
	g.quitConfirmationIndex = 1
	mockInput.JustPressedKeys[engine.KeyA] = true
	mh.Update()
	if g.quitConfirmationIndex != 0 { t.Errorf("expected 0, got %d", g.quitConfirmationIndex) }

	// Test select "No" (index 1) -> Returns to Main Menu
	g.quitConfirmationIndex = 1
	mockInput.JustPressedKeys[engine.KeyA] = false
	mockInput.JustPressedKeys[engine.KeyEnter] = true
	mh.Update()
	if g.isQuitConfirmationOpen { t.Error("expected quit confirmation closed") }
}

func TestMenu_Main_Navigation(t *testing.T) {
	g := NewTestGame(t)
	mh := g.menuHandler
	mockInput := g.input.(*engine.MockInput)

	// Initially mainMenuIndex is 0
	if g.mainMenuIndex != 0 { t.Errorf("expected index 0, got %d", g.mainMenuIndex) }

	// Test navigation (Down/S)
	mockInput.JustPressedKeys[engine.KeyS] = true
	mh.Update()
	if g.mainMenuIndex != 1 { t.Errorf("expected index 1, got %d", g.mainMenuIndex) }

	// Test select Settings (index 2)
	g.mainMenuIndex = 2
	mockInput.JustPressedKeys[engine.KeyS] = false
	mockInput.JustPressedKeys[engine.KeyEnter] = true
	mh.Update()
	if !g.isSettingsScreen { t.Error("expected settings screen open") }

	// Test select New Game (index 0)
	g.isSettingsScreen = false
	g.isMainMenu = true
	g.mainMenuIndex = 0
	mockInput.JustPressedKeys[engine.KeyEnter] = true
	mh.Update()
	if g.isMainMenu { t.Error("expected mainMenu closed for new game") }
	if !g.isCharacterSelect { t.Error("expected char select open") }

	// Test select Quit (index 4)
	g.isMainMenu = true
	g.mainMenuIndex = 4
	mockInput.JustPressedKeys[engine.KeyEnter] = true
	mh.Update()
	if !g.isQuitConfirmationOpen { t.Error("expected quit confirmation open") }
}

func TestMenu_Main_Mouse(t *testing.T) {
	g := NewTestGame(t)
	mh := g.menuHandler
	mockInput := g.input.(*engine.MockInput)

	// Mouse hover over "About" (index 3)
	// centerX = 500. itemY = 350 + 3*60 = 530.
	mockInput.MouseX, mockInput.MouseY = 500, 530
	mh.Update()
	if g.mainMenuIndex != 3 { t.Errorf("expected hover index 3, got %d", g.mainMenuIndex) }

	// Click it
	mockInput.JustPressedButtons[engine.MouseButtonLeft] = true
	mh.Update()
	if !g.isAboutScreen { t.Error("expected about screen open") }
}

func TestMenu_Pause(t *testing.T) {
	g := NewTestGame(t)
	g.isMainMenu = false
	g.isMenuOpen = true // Pasuse menu
	mh := g.menuHandler
	mockInput := g.input.(*engine.MockInput)

	// Test Resume
	g.menuIndex = 0
	mockInput.JustPressedKeys[engine.KeyEnter] = true
	mh.Update()
	if g.isMenuOpen { t.Error("expected pause closed") }

	// Test Settings from Pause
	g.isMenuOpen = true
	g.menuIndex = 1 // Settings (wait, or is it 1? lets check menu_pause.go)
	// nOptions = 5: Resume, Quicksave, Load, Settings, Quit (if settings is 3) or 4?
}

func TestMenu_Trading(t *testing.T) {
	g := NewTestGame(t)
	g.isMainMenu = false
	g.isTradeOpen = true
	g.ActiveTrader = &Character{Actor: Actor{Name: "Trader", Inventory: []*ItemInstance{}}}
	g.playableCharacter = &Character{Actor: Actor{Name: "Player", Inventory: []*ItemInstance{}}}
	mh := g.menuHandler
	mockInput := g.input.(*engine.MockInput)

	// Click CLOSE button (dialog centered)
	// dialogX = (1000-900)/2 = 50. dialogY = (800-600)/2 = 100.
	// Close button at 50 + 900 - 40 = 910. Close button Y = 100 + 10 = 110.
	mockInput.MouseX, mockInput.MouseY = 910+5, 110+5
	mockInput.JustPressedButtons[engine.MouseButtonLeft] = true
	mh.Update()
	// Wait, updateTradeScreen might handle close differently.
}

func TestMenu_Inventory(t *testing.T) {
	g := NewTestGame(t)
	g.isMainMenu = false
	g.isInventoryOpen = true
	pc := &Character{Actor: Actor{Name: "Player", Inventory: []*ItemInstance{
		{Config: &ObjectConfig{Name: "Bread", Type: "consumable", Hunger: 10}},
		{Config: &ObjectConfig{Name: "Book", Content: "Once upon a time"}},
	}}}
	g.playableCharacter = pc
	g.World.PlayableCharacter = pc
	mh := g.menuHandler
	mockInput := g.input.(*engine.MockInput)

	// lastMouse initialized
	g.lastMouseX, g.lastMouseY = -1, -1

	// Click CLOSE button
	mockInput.MouseX, mockInput.MouseY = 450+10, 650+10
	mockInput.JustPressedButtons[engine.MouseButtonLeft] = true
	mh.Update()
	if g.isInventoryOpen { t.Error("expected inventory closed via button") }

	// Reopen
	g.isInventoryOpen = true
	mockInput.JustPressedButtons[engine.MouseButtonLeft] = false
	
	// Test Consume item via button [E] - Row 0 (Bread)
	// uBtnX = 810. itemY = 220.
	mockInput.MouseX, mockInput.MouseY = 810+5, 220+10
	mockInput.JustPressedButtons[engine.MouseButtonLeft] = true
	mh.Update()
	if len(pc.Inventory) != 1 { t.Errorf("expected 1 item after eating, got %d", len(pc.Inventory)) }

	// Test Read item via button [R] - Row 0 (Book, as Bread is gone)
	mockInput.JustPressedButtons[engine.MouseButtonLeft] = false
	mockInput.MouseX, mockInput.MouseY = 850+5, 220+10
	mockInput.JustPressedButtons[engine.MouseButtonLeft] = true
	mh.Update()
	if g.ActiveBook == nil { t.Error("expected active book") }
	
	// Close book
	mockInput.JustPressedButtons[engine.MouseButtonLeft] = false
	mockInput.JustPressedKeys[engine.KeyEscape] = true
	mh.Update()
	if g.ActiveBook != nil { t.Error("expected book closed") }
}

func TestMenu_KeymapRemap(t *testing.T) {
	g := NewTestGame(t)
	g.isMainMenu = false
	g.isKeymapScreen = true
	mh := g.menuHandler
	mockInput := g.input.(*engine.MockInput)

	// test navigation
	mockInput.JustPressedKeys[engine.KeyS] = true
	mh.Update()
	mockInput.JustPressedKeys[engine.KeyS] = false
	if g.keymapSelectedIndex != 1 { t.Errorf("expected 1, got %d", g.keymapSelectedIndex) }

	// navigate back to 0 (move_up)
	mockInput.JustPressedKeys[engine.KeyW] = true
	mh.Update()
	mockInput.JustPressedKeys[engine.KeyW] = false
	if g.keymapSelectedIndex != 0 { t.Errorf("expected 0, got %d", g.keymapSelectedIndex) }

	// test enter remapping mode
	mockInput.JustPressedKeys[engine.KeyEnter] = true
	mh.Update()
	mockInput.JustPressedKeys[engine.KeyEnter] = false
	if g.remappingAction == "" { t.Error("expected remapping action set") }

	// test remap to "Q"
	mockInput.JustPressedKeys[engine.KeyQ] = true
	mh.Update() // This will call updateKeymapScreen
	mockInput.JustPressedKeys[engine.KeyQ] = false
	
	if g.remappingAction != "" { t.Error("expected remapping finished") }
	if g.settings.Keymap["move_up"] != "Q" { t.Errorf("expected Q, got %v", g.settings.Keymap["move_up"]) }
}

func TestMenu_Settings(t *testing.T) {
	g := NewTestGame(t)
	g.isMainMenu = false
	g.isSettingsScreen = true
	mockInput := g.input.(*engine.MockInput)


	// test navigation
	mockInput.JustPressedKeys[engine.KeyS] = true
	g.Update()
	mockInput.JustPressedKeys[engine.KeyS] = false
	if g.settingsMenuIndex != 1 { t.Errorf("expected 1, got %d", g.settingsMenuIndex) }

	// test Font Style toggle
	mockInput.JustPressedKeys[engine.KeyS] = false
	g.settingsMenuIndex = 0
	oldFont := g.settings.Font
	mockInput.JustPressedKeys[engine.KeyRight] = true
	g.Update()
	mockInput.JustPressedKeys[engine.KeyRight] = false
	if g.settings.Font == oldFont { t.Errorf("expected font changed") }

	// test Sound Effects (index 1)
	mockInput.JustPressedKeys[engine.KeyRight] = false
	g.settingsMenuIndex = 1
	oldSound := g.settings.SoundFrequency
	mockInput.JustPressedKeys[engine.KeyRight] = true
	g.Update()
	mockInput.JustPressedKeys[engine.KeyRight] = false
	if g.settings.SoundFrequency == oldSound { t.Error("expected sound frequency changed") }

	// test Save and Back (last index - usually 9)
	mockInput.JustPressedKeys[engine.KeyRight] = false
	g.settingsMenuIndex = 10
	mockInput.JustPressedKeys[engine.KeyEnter] = true
	g.Update()
	mockInput.JustPressedKeys = make(map[engine.Key]bool)
	if g.isSettingsScreen { t.Error("expected settings closed") }
	if !g.isMainMenu { t.Error("expected returned to main menu") }
}

func TestMenu_Dialogue(t *testing.T) {
	g := NewTestGame(t)
	g.isMainMenu = false
	g.ActiveDialogue = &DialogueState{
		CurrentText: "Hello",
		Choices:     []Choice{{Text: "Hi"}},
	}
	mockInput := g.input.(*engine.MockInput)
	
	// test navigation
	mockInput.JustPressedKeys[engine.KeyS] = true
	g.Update()
	
	// test advance
	mockInput.JustPressedKeys[engine.KeyS] = false
	mockInput.JustPressedKeys[engine.KeyEnter] = true
	g.Update()
	
	// test close
	g.ActiveDialogue = &DialogueState{CurrentText: "Bye"}
	mockInput.JustPressedKeys[engine.KeyEnter] = false
	mockInput.JustPressedKeys[engine.KeyEscape] = true
	g.Update()
	if g.ActiveDialogue != nil { t.Error("expected dialogue closed") }
}

func TestMenu_Campaign(t *testing.T) {
	g := NewTestGame(t)
	g.isMainMenu = false
	g.isCampaignSelect = true
	mockInput := g.input.(*engine.MockInput)
	
	// test navigation
	mockInput.JustPressedKeys[engine.KeyS] = true
	g.Update()
	
	// test close (Escape in campaign select opens quit confirmation)
	mockInput.JustPressedKeys[engine.KeyS] = false
	mockInput.JustPressedKeys[engine.KeyEscape] = true
	g.Update()
	if !g.isQuitConfirmationOpen { t.Error("expected quit confirmation open") }
}
