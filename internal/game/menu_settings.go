package game

import (
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updateSettingsScreen() error {
	g := mh.game
	nRows := 6 // 0: Font, 1: Audio, 2: Fog of War, 3: AI Provider, 4: AI Sim, 5: Save & Back
	if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
		g.settingsMenuIndex--
		if g.settingsMenuIndex < 0 {
			g.settingsMenuIndex = nRows - 1
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
		g.settingsMenuIndex++
		if g.settingsMenuIndex >= nRows {
			g.settingsMenuIndex = 0
		}
	}

	if g.settingsMenuIndex == 0 { // Font
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			g.settingsFontIndex--
			if g.settingsFontIndex < 0 {
				g.settingsFontIndex = len(FontOptions) - 1
			}
			g.settings.Font = FontOptions[g.settingsFontIndex]
			g.UpdateFont()
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
			g.settingsFontIndex++
			if g.settingsFontIndex >= len(FontOptions) {
				g.settingsFontIndex = 0
			}
			g.settings.Font = FontOptions[g.settingsFontIndex]
			g.UpdateFont()
		}
	} else if g.settingsMenuIndex == 1 { // Audio
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			g.settingsAudioIndex--
			if g.settingsAudioIndex < 0 {
				g.settingsAudioIndex = len(FrequencyOptions) - 1
			}
			g.settings.SoundFrequency = FrequencyOptions[g.settingsAudioIndex]
			if g.audio != nil {
				g.audio.SetProbability(g.settings.GetSoundProbability())
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
			g.settingsAudioIndex++
			if g.settingsAudioIndex >= len(FrequencyOptions) {
				g.settingsAudioIndex = 0
			}
			g.settings.SoundFrequency = FrequencyOptions[g.settingsAudioIndex]
			if g.audio != nil {
				g.audio.SetProbability(g.settings.GetSoundProbability())
			}
		}
	} else if g.settingsMenuIndex == 2 { // Fog of War
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			g.settingsFogIndex--
			if g.settingsFogIndex < 0 {
				g.settingsFogIndex = len(FogOfWarOptions) - 1
			}
			g.settings.FogOfWar = FogOfWarOptions[g.settingsFogIndex]
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
			g.settingsFogIndex++
			if g.settingsFogIndex >= len(FogOfWarOptions) {
				g.settingsFogIndex = 0
			}
			g.settings.FogOfWar = FogOfWarOptions[g.settingsFogIndex]
		}
	} else if g.settingsMenuIndex == 3 { // AI Provider
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			found := -1
			for i, opt := range AIProviderOptions {
				if opt == g.settings.AIProvider {
					found = i
					break
				}
			}
			if found == -1 { found = 0 }
			found--
			if found < 0 { found = len(AIProviderOptions) - 1 }
			g.settings.AIProvider = AIProviderOptions[found]
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
			found := -1
			for i, opt := range AIProviderOptions {
				if opt == g.settings.AIProvider {
					found = i
					break
				}
			}
			if found == -1 { found = 0 }
			found++
			if found >= len(AIProviderOptions) { found = 0 }
			g.settings.AIProvider = AIProviderOptions[found]
		}
	} else if g.settingsMenuIndex == 4 { // AI Simulation Mode
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) ||
			g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
			g.settings.AISimulationMode = !g.settings.AISimulationMode
		}
	}

	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my

	hoverIdx := -1
	centerX := g.width / 2
	for i := 0; i < nRows; i++ {
		itemY := 250 + i*60
		if mx >= centerX-250 && mx <= centerX+250 && my >= itemY-30 && my <= itemY+30 {
			hoverIdx = i
		}
	}
	if hoverIdx != -1 && mouseMoved {
		g.settingsMenuIndex = hoverIdx
	}

	if g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIdx == 5 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft)) {
		g.settings.Save()
		g.UpdateFont()
		if g.audio != nil {
			g.audio.SetProbability(g.settings.GetSoundProbability())
		}
		g.isSettingsScreen = false
		if g.isSettingsFromPause {
			g.isMenuOpen = true
			g.isSettingsFromPause = false
		} else {
			g.isMainMenu = true
		}
	}

	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isSettingsScreen = false
		if g.isSettingsFromPause {
			g.isMenuOpen = true
			g.isSettingsFromPause = false
		} else {
			g.isMainMenu = true
		}
	}
	return nil
}
