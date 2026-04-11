package game

import (
	"fmt"
	"image/color"
	"oinakos/internal/engine"
	"strings"
)

func (gr *GameRenderer) drawSettingsScreen(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)
	title := "SETTINGS"
	tw, _ := gr.graphics.MeasureText(title, 40)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 100, color.RGBA{218, 165, 32, 255}, 40)

	rows := []string{"Font Style", "Sound Effects", "Fog of War", "AI Provider"}
	if g.settings.AIProvider != "none" { rows = append(rows, "AI Model") }
	rows = append(rows, "Time Pace", "Simulation Mode", "Talking Frequency", "Measurement Units", "Adult Mode", "Keymap", "Save and Back")

	for i, row := range rows {
		var clr color.Color = color.White; prefix := "  "
		if g.settingsMenuIndex == i { clr, prefix = color.RGBA{255, 255, 0, 255}, "> " }
		label := prefix + row
		switch row {
		case "Font Style": label += fmt.Sprintf(": [%s]", strings.ToUpper(FontOptions[g.settingsFontIndex]))
		case "Sound Effects": label += fmt.Sprintf(": [%s]", strings.ToUpper(FrequencyOptions[g.settingsAudioIndex]))
		case "Fog of War": label += fmt.Sprintf(": [%s]", strings.ToUpper(FogOfWarOptions[g.settingsFogIndex]))
		case "AI Provider": label += fmt.Sprintf(": [%s]", strings.ToUpper(g.settings.AIProvider))
		case "AI Model": model := g.getAIModelForCurrentProvider(); if model == "" { model = "AUTO" }
			if g.isFetchingModels { label += ": [FETCHING...]" } else { label += fmt.Sprintf(": [%s]", strings.ToUpper(model)) }
		case "Time Pace": label += fmt.Sprintf(": [%s]", strings.ToUpper(TimePaceLabels[g.settings.TimePace]))
		case "Simulation Mode": val := "OFF"; if g.settings.AISimulationMode { val = "ON" }; label += fmt.Sprintf(": [%s]", val)
		case "Talking Frequency": label += fmt.Sprintf(": [%s]", strings.ToUpper(g.settings.TalkingFrequency))
		case "Measurement Units": label += fmt.Sprintf(": [%s]", strings.ToUpper(g.settings.Units))
		case "Adult Mode": val := "OFF"; if g.settings.AdultMode { val = "ON" }; label += fmt.Sprintf(": [%s]", val)
		}
		lw, _ := gr.graphics.MeasureText(label, 18)
		gr.graphics.DrawTextAt(screen, label, (g.width-int(lw))/2, 200+i*40, clr, 18)
	}
}

func (gr *GameRenderer) drawKeymapScreen(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)
	title := "CHARACTER KEYMAP"
	tw, _ := gr.graphics.MeasureText(title, 40)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 100, color.RGBA{218, 165, 32, 255}, 40)

	startY, rowHeight, listX := 200, 30, (g.width-450)/2
	for i, action := range RemappableActions {
		curY := startY + i*rowHeight
		var labelColor color.Color = color.RGBA{200, 200, 200, 255}
		if i == g.keymapSelectedIndex { labelColor = color.RGBA{255, 255, 100, 255}; gr.graphics.DrawFilledRect(screen, float32(listX-10), float32(curY-5), 470, float32(rowHeight), color.RGBA{50, 50, 50, 255}, false) }
		keyName := g.settings.Keymap[action.ID]; if g.remappingAction == action.ID { keyName = "[Press ANY Key...]" }
		gr.graphics.DrawTextAt(screen, action.Name, listX, curY+rowHeight/2, labelColor, 20)
		gr.graphics.DrawTextAt(screen, keyName, listX+300, curY+rowHeight/2, labelColor, 20)
	}
	gr.graphics.DrawTextAt(screen, "[ BACK ]", (g.width-80)/2, startY+(len(RemappableActions)+1)*rowHeight+20, color.White, 20)
}
