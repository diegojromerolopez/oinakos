package game

import (
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updateSettingsScreen() error {
	g := mh.game

	showAIModel := g.settings.AIProvider != "none"
	rows := []string{"Font Style", "Sound Effects", "Fog of War", "AI Provider"}
	if showAIModel {
		rows = append(rows, "AI Model")
	}
	rows = append(rows, "Time Pace", "Simulation Mode", "Talking Frequency", "Measurement Units", "Adult Mode", "Keymap", "Save and Back")
	nRows := len(rows)

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
	currentRow := rows[g.settingsMenuIndex]
	switch currentRow {
	case "Font Style":
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			g.settingsFontIndex--
			if g.settingsFontIndex < 0 {
				g.settingsFontIndex = len(FontOptions) - 1
			}
			g.settings.Font = FontOptions[g.settingsFontIndex]
			g.UpdateFont()
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
			g.settingsFontIndex++
			if g.settingsFontIndex >= len(FontOptions) {
				g.settingsFontIndex = 0
			}
			g.settings.Font = FontOptions[g.settingsFontIndex]
			g.UpdateFont()
		}
	case "Sound Effects":
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
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
			g.settingsAudioIndex++
			if g.settingsAudioIndex >= len(FrequencyOptions) {
				g.settingsAudioIndex = 0
			}
			g.settings.SoundFrequency = FrequencyOptions[g.settingsAudioIndex]
			if g.audio != nil {
				g.audio.SetProbability(g.settings.GetSoundProbability())
			}
		}
	case "Fog of War":
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			g.settingsFogIndex--
			if g.settingsFogIndex < 0 {
				g.settingsFogIndex = len(FogOfWarOptions) - 1
			}
			g.settings.FogOfWar = FogOfWarOptions[g.settingsFogIndex]
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
			g.settingsFogIndex++
			if g.settingsFogIndex >= len(FogOfWarOptions) {
				g.settingsFogIndex = 0
			}
			g.settings.FogOfWar = FogOfWarOptions[g.settingsFogIndex]
		}
	case "AI Provider":
		options := AIProviderOptions
		if g.isWasm() {
			options = []string{"none", "webgpu"}
		}
		oldProvider := g.settings.AIProvider
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
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
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
		if g.settings.AIProvider != oldProvider {
			g.initAIManager() // Re-fetch models
		}
	case "AI Model":
		if len(g.availableModels) > 0 {
			currentModel := g.getAIModelForCurrentProvider()
			found := -1
			for i, m := range g.availableModels {
				if m == currentModel {
					found = i
					break
				}
			}
			if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
				if found == -1 { found = 0 }
				found--
				if found < 0 { found = len(g.availableModels) - 1 }
				g.setAIModelForCurrentProvider(g.availableModels[found])
			}
			if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
				if found == -1 { found = 0 }
				found++
				if found >= len(g.availableModels) { found = 0 }
				g.setAIModelForCurrentProvider(g.availableModels[found])
			}
		}
	case "Time Pace":
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
			found := -1
			for i, opt := range TimePaceOptions {
				if opt == g.settings.TimePace { found = i; break }
			}
			if found == -1 { found = 0 }
			found--
			if found < 0 { found = len(TimePaceOptions) - 1 }
			g.settings.TimePace = TimePaceOptions[found]
		}
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
			found := -1
			for i, opt := range TimePaceOptions {
				if opt == g.settings.TimePace { found = i; break }
			}
			if found == -1 { found = 0 }
			found++
			if found >= len(TimePaceOptions) { found = 0 }
			g.settings.TimePace = TimePaceOptions[found]
		}
	case "Simulation Mode":
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) ||
			g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
			g.settings.AISimulationMode = !g.settings.AISimulationMode
		}
	case "Talking Frequency":
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
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
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
	case "Measurement Units":
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
		if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
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
	case "Adult Mode":
		if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) ||
			g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
			g.settings.AdultMode = !g.settings.AdultMode
		}
	case "Keymap":
		if g.input.IsKeyJustPressed(engine.KeyEnter) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
			g.isKeymapScreen = true
			g.isSettingsScreen = false
			return nil
		}
	case "Save and Back":
		if g.input.IsKeyJustPressed(engine.KeyEnter) || (mouseClicked && hoverIdx == g.settingsMenuIndex) {
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
