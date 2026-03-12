package game

import (
	"log"
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updateMainMenu() error {
	g := mh.game
	nOptions := 4
	if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
		g.mainMenuIndex--
		if g.mainMenuIndex < 0 {
			g.mainMenuIndex = nOptions - 1
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
		g.mainMenuIndex++
		if g.mainMenuIndex >= nOptions {
			g.mainMenuIndex = 0
		}
	}

	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my

	hoverIndex := -1
	centerX := g.width / 2
	for i := 0; i < nOptions; i++ {
		itemY := 350 + i*60
		if mx >= centerX-200 && mx <= centerX+200 && my >= itemY-30 && my <= itemY+30 {
			hoverIndex = i
		}
	}
	if hoverIndex != -1 && mouseMoved {
		g.mainMenuIndex = hoverIndex
	}

	handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

	if handleSelect {
		switch g.mainMenuIndex {
		case 0: // New Game
			g.isMainMenu = false
			g.isCharacterSelect = true
		case 1: // Load Game
			g.loadDialogActive = true
		case 2: // Settings
			g.settings = LoadSettings()
			g.settingsFontIndex = 0
			for idx, val := range FontOptions {
				if val == g.settings.Font {
					g.settingsFontIndex = idx
					break
				}
			}
			g.settingsAudioIndex = 0
			for idx, val := range FrequencyOptions {
				if val == g.settings.SoundFrequency {
					g.settingsAudioIndex = idx
					break
				}
			}
			g.settingsFogIndex = 0
			for idx, val := range FogOfWarOptions {
				if val == g.settings.FogOfWar {
					g.settingsFogIndex = idx
					break
				}
			}
			g.settingsMenuIndex = 0
			g.isMainMenu = false
			g.isSettingsScreen = true
		case 3: // Quit
			if !g.isWasm() {
				g.isQuitConfirmationOpen = true
				g.quitConfirmationIndex = 1
			}
		}
	}

	if g.loadDialogActive {
		path := g.openFilePicker()
		if path != "" {
			if err := g.Load(path); err == nil {
				g.isMainMenu = false
				g.isCharacterSelect = false
			} else {
				log.Printf("Failed to load map: %v", err)
			}
		}
		g.loadDialogActive = false
	}

	return nil
}
