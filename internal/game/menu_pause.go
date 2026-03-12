package game

import (
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updatePauseMenu() error {
	g := mh.game
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isMenuOpen = false
		return nil
	}
	if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
		g.menuIndex--
		if g.menuIndex < 0 {
			g.menuIndex = 4
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
		g.menuIndex++
		if g.menuIndex > 4 {
			g.menuIndex = 0
		}
	}

	mw, mh_val := 400, 280
	bx, by := (g.width-mw)/2, (g.height-mh_val)/2
	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my

	hoverIndex := -1
	for i := 0; i < 5; i++ {
		itemY := by + 70 + i*35
		if mx >= bx+100 && mx <= bx+250 && my >= itemY-10 && my <= itemY+30 {
			hoverIndex = i
		}
	}

	if hoverIndex != -1 && mouseMoved {
		g.menuIndex = hoverIndex
	}

	handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

	if handleSelect {
		switch g.menuIndex {
		case 0: // Resume
			g.isMenuOpen = false
		case 1: // Quicksave
			mh.game.performQuicksave()
			g.isMenuOpen = false
		case 2: // Load
			g.loadDialogActive = true
			g.isMenuOpen = false
		case 3: // Settings
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
			g.isSettingsFromPause = true
			g.isMenuOpen = false
			g.isSettingsScreen = true
		case 4: // Quit
			g.isQuitConfirmationOpen = true
			g.quitConfirmationIndex = 1
		}
	}
	return nil
}
