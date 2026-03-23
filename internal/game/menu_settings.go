package game

import (
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updateSettingsScreen() error {
	g := mh.game
	nRows := 9 // Font, Audio, Fog, AI Provider, AI Sim, Talking, Units, Keymap, Save & Back

	// 1. Mouse Detection (Prioritize mouse for selection/hover)
	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my
	mouseClicked := g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft)

	hoverIdx := -1
	centerX := g.width / 2
	for i := 0; i < nRows; i++ {
		itemY := 200 + i*40
		if mx >= centerX-250 && mx <= centerX+250 && my >= itemY-20 && my <= itemY+20 {
			hoverIdx = i
		}
	}
	if hoverIdx != -1 && (mouseMoved || mouseClicked) {
		g.settingsMenuIndex = hoverIdx
	}

	// 2. Keyboard Navigation
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

	// 3. Row-Specific Logic
	if g.settingsMenuIndex == 0 { // Font
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			g.settingsFontIndex--
			if g.settingsFontIndex < 0 {
				g.settingsFontIndex = len(FontOptions) - 1
			}
			g.settings.Font = FontOptions[g.settingsFontIndex]
			g.UpdateFont()
		}
		// Cycle on Right Arrow OR Click
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == 0) {
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
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == 1) {
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
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == 2) {
			g.settingsFogIndex++
			if g.settingsFogIndex >= len(FogOfWarOptions) {
				g.settingsFogIndex = 0
			}
			g.settings.FogOfWar = FogOfWarOptions[g.settingsFogIndex]
		}
	} else if g.settingsMenuIndex == 3 { // AI Provider
		options := AIProviderOptions
		if g.isWasm() {
			options = []string{"none", "webgpu"}
		}
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			found := -1
			for i, opt := range options {
				if opt == g.settings.AIProvider {
					found = i
					break
				}
			}
			if found == -1 { found = 0 }
			found--
			if found < 0 { found = len(options) - 1 }
			g.settings.AIProvider = options[found]
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == 3) {
			found := -1
			for i, opt := range options {
				if opt == g.settings.AIProvider {
					found = i
					break
				}
			}
			if found == -1 { found = 0 }
			found++
			if found >= len(options) { found = 0 }
			g.settings.AIProvider = options[found]
		}
	} else if g.settingsMenuIndex == 4 { // AI Simulation Mode
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) ||
			g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == 4) {
			g.settings.AISimulationMode = !g.settings.AISimulationMode
		}
	} else if g.settingsMenuIndex == 5 { // Talking Frequency
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			found := -1
			for i, opt := range FrequencyOptions {
				if opt == g.settings.TalkingFrequency {
					found = i
					break
				}
			}
			if found == -1 { found = 0 }
			found--
			if found < 0 { found = len(FrequencyOptions) - 1 }
			g.settings.TalkingFrequency = FrequencyOptions[found]
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == 5) {
			found := -1
			for i, opt := range FrequencyOptions {
				if opt == g.settings.TalkingFrequency {
					found = i
					break
				}
			}
			if found == -1 { found = 0 }
			found++
			if found >= len(FrequencyOptions) { found = 0 }
			g.settings.TalkingFrequency = FrequencyOptions[found]
		}
	} else if g.settingsMenuIndex == 6 { // Units
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			found := -1
			for i, opt := range UnitsOptions {
				if opt == g.settings.Units {
					found = i
					break
				}
			}
			if found == -1 { found = 0 }
			found--
			if found < 0 { found = len(UnitsOptions) - 1 }
			g.settings.Units = UnitsOptions[found]
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == 6) {
			found := -1
			for i, opt := range UnitsOptions {
				if opt == g.settings.Units {
					found = i
					break
				}
			}
			if found == -1 { found = 0 }
			found++
			if found >= len(UnitsOptions) { found = 0 }
			g.settings.Units = UnitsOptions[found]
		}
	} else if g.settingsMenuIndex == 7 { // Keymap
		if g.input.IsKeyJustPressed(engine.KeyEnter) || (mouseClicked && hoverIdx == 7) {
			g.isKeymapScreen = true
			g.isSettingsScreen = false
			return nil
		}
	}

	// 4. Global Confirmation / Back
	if g.input.IsKeyJustPressed(engine.KeyEnter) || (mouseClicked && hoverIdx == 8) {
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
