package game

import (
	"fmt"
	"image/color"
	"oinakos/internal/engine"
	"strings"
	"sync/atomic"
)

func (gr *GameRenderer) drawMainMenu(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	title := "OINAKOS"
	tw, _ := gr.graphics.MeasureText(title, 60)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 150, color.RGBA{218, 165, 32, 255}, 60)

	subtitle := "A KNIGHT'S PATH"
	stw, _ := gr.graphics.MeasureText(subtitle, 24)
	gr.graphics.DrawTextAt(screen, subtitle, (g.width-int(stw))/2, 220, color.RGBA{150, 150, 150, 255}, 24)

	options := []string{"NEW GAME", "LOAD GAME", "SETTINGS", "ABOUT", "QUIT"}

	for i, opt := range options {
		var clr color.Color = color.White
		prefix := "  "
		if g.mainMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}
		label := prefix + opt
		lw, _ := gr.graphics.MeasureText(label, 32)
		gr.graphics.DrawTextAt(screen, label, (g.width-int(lw))/2, 350+i*60, clr, 32)
	}

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("v%s", g.Version), 20, g.height-30, color.RGBA{80, 80, 80, 255}, 14)
}

func (gr *GameRenderer) drawAboutScreen(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	title := "THE STORY OF OINAKOS"
	tw, _ := gr.graphics.MeasureText(title, 40)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 80, color.RGBA{218, 165, 32, 255}, 40)

	story := []string{
		"The man stirred, a searing, lacerating pain radiating from his left arm.",
		"He had been lost to unconsciousness for the better part of the day.",
		"Through the haze, fragments of the previous night flickered: a wagon laden",
		"with timber for the market, the biting chill of the evening wind against",
		"his skin, and then... the howl.",
		"",
		"It had happened where the forest pressed tight against the road, a place of",
		"deep shadows and treacherous silence. A young wolf had lunged first,",
		"snapping at the frantic horses. Another followed, jaws clamping near his",
		"boot. Finally, a grizzled alpha leapt into the wagon, its teeth sinking",
		"deep into his arm. In the ensuing frenzy, the horses bolted, the wagon",
		"overturned, and the pack vanished into the gloom.",
		"",
		"When the chaos stilled, he was left broken in the freezing mud, his",
		"livelihood shattered.",
		"",
		"HE HATED THE WOLVES.",
	}

	for i, line := range story {
		var clr color.Color = color.White
		size := 16
		if strings.Contains(line, "HE HATED THE WOLVES") {
			clr = color.RGBA{255, 50, 50, 255}
			size = 20
		}
		lw, _ := gr.graphics.MeasureText(line, float64(size))
		gr.graphics.DrawTextAt(screen, line, (g.width-int(lw))/2, 180+i*28, clr, float64(size))
	}

	prompt := "Press ESC to return"
	pw, _ := gr.graphics.MeasureText(prompt, 14)
	gr.graphics.DrawTextAt(screen, prompt, (g.width-int(pw))/2, g.height-80, color.RGBA{150, 150, 150, 255}, 14)
}

func (gr *GameRenderer) drawCharacterSelect(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	title := "OINAKOS: CHOOSE YOUR HERO"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 50, color.RGBA{218, 165, 32, 255}, 32)

	mx, my := g.input.MousePosition()
	playableIDs := g.characterRegistry.PlayableIDs()

	for i, id := range playableIDs {
		char := g.characterRegistry.Characters[id]
		var clr color.Color = color.White
		prefix := "  "

		// Mouse hover detection (simplified, actual interaction handled in menu_handler)
		if mx >= 100 && mx <= 400 && my >= 130+i*30-5 && my <= 130+i*30+25 {
			// hover
		}

		if g.characterMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
			gr.drawHeroPreview(screen, char, g.width/2+50, 130)
		}
		gr.graphics.DrawTextAt(screen, prefix+char.Name, 100, 130+i*35, clr, 18)
	}
	nPlayable := len(playableIDs)
	// Back button
	var clrBack color.Color = color.White
	prefixBack := "  "
	if g.characterMenuIndex == nPlayable {
		clrBack = color.RGBA{255, 255, 0, 255}
		prefixBack = "> "
	}
	gr.graphics.DrawTextAt(screen, prefixBack+"BACK", 100, 130+nPlayable*35+20, clrBack, 18)

	msg := "Press UP/DOWN to navigate, ENTER to select hero."
	mw, _ := gr.graphics.MeasureText(msg, 14)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(mw))/2, g.height-50, color.RGBA{136, 136, 136, 255}, 14)
}

func (gr *GameRenderer) drawHeroPreview(screen engine.Image, char *EntityConfig, x, y int) {
	gr.graphics.DrawTextAt(screen, "--- HERO PROFILE ---", x, y, color.RGBA{218, 165, 32, 255}, 20)

	if char.StaticImage != nil {
		img := char.StaticImage
		op := engine.NewDrawImageOptions()
		op.Scale(1.5, 1.5)
		op.Translate(float64(x), float64(y+30))
		screen.DrawImage(img, op)
	}

	statsX := x + 180
	statsY := y + 40
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Health:  %d", char.Stats.HealthMin), statsX, statsY, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Attack:  %d", char.Stats.BaseAttack), statsX, statsY+25, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Defense: %d", char.Stats.BaseDefense), statsX, statsY+50, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Speed:   %.2f", char.Stats.Speed), statsX, statsY+75, color.White, 16)
	weaponName := "None"
	if !char.Weapon.IsEmpty() {
		w := char.Weapon.Resolve(gr.game.Registries.Objects)
		if w != nil {
			weaponName = w.Name
		}
	}
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Weapon:  %s", weaponName), statsX, statsY+100, color.White, 16)

	gr.graphics.DrawTextAt(screen, "--- BIOGRAPHY ---", x, y+230, color.RGBA{218, 165, 32, 255}, 20)
	words := strings.Fields(char.Description)
	line := ""
	lineNum := 0
	for _, w := range words {
		if len(line)+len(w) > 40 {
			gr.graphics.DrawTextAt(screen, line, x, y+260+lineNum*20, color.White, 14)
			line = w + " "
			lineNum++
		} else {
			line += w + " "
		}
	}
	gr.graphics.DrawTextAt(screen, line, x, y+260+lineNum*20, color.White, 14)
}

func (gr *GameRenderer) drawCampaignSelect(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	title := "OINAKOS: SELECT YOUR JOURNEY"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 50, color.RGBA{218, 165, 32, 255}, 32)

	col1X := 100
	col2X := g.width / 2

	gr.graphics.DrawTextAt(screen, "--- CAMPAIGNS ---", col1X-20, 100, color.RGBA{150, 150, 150, 255}, 18)
	gr.graphics.DrawTextAt(screen, "--- MAPS ---", col2X-20, 100, color.RGBA{150, 150, 150, 255}, 18)

	nC := len(g.campaignRegistry.IDs)
	nM := len(g.mapTypeRegistry.IDs)
	y := 130

	for i, id := range g.campaignRegistry.IDs {
		camp := g.campaignRegistry.Campaigns[id]
		var clr color.Color = color.White
		prefix := "  "
		if g.campaignMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}
		gr.graphics.DrawTextAt(screen, prefix+camp.Name, col1X, y+i*30, clr, 16)
	}

	for i, id := range g.mapTypeRegistry.IDs {
		m := g.mapTypeRegistry.Types[id]
		var clr color.Color = color.White
		prefix := "  "
		idx := nC + i
		if g.campaignMenuIndex == idx {
			clr = color.RGBA{150, 255, 150, 255}
			prefix = "> "
		}

		colOffset := col2X
		rowOffset := i
		if i > 15 {
			colOffset += 250
			rowOffset = i - 16
		}

		gr.graphics.DrawTextAt(screen, prefix+m.Name, colOffset, y+rowOffset*30, clr, 16)
	}

	var clr color.Color = color.White
	prefix := "  "
	if g.campaignMenuIndex == nC+nM {
		clr = color.RGBA{255, 0, 0, 255}
		prefix = "> "
	}
	quitText := prefix + "QUIT"
	qw, _ := gr.graphics.MeasureText(quitText, 24)
	gr.graphics.DrawTextAt(screen, quitText, (g.width-int(qw))/2, g.height-90, clr, 24)

	msg := "Press UP/DOWN to navigate, ENTER to begin."
	mw, _ := gr.graphics.MeasureText(msg, 14)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(mw))/2, g.height-50, color.RGBA{136, 136, 136, 255}, 14)
}

func (gr *GameRenderer) drawSettingsScreen(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	title := "SETTINGS"
	tw, _ := gr.graphics.MeasureText(title, 40)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 100, color.RGBA{218, 165, 32, 255}, 40)

	rows := []string{"Font Style", "Sound Effects", "Fog of War", "AI Provider"}
	showAIModel := g.settings.AIProvider != "none"
	if showAIModel {
		rows = append(rows, "AI Model")
	}
	rows = append(rows, "Simulation Mode", "Talking Frequency", "Measurement Units", "Keymap", "Save and Back")

	for i, row := range rows {
		var clr color.Color = color.White
		prefix := "  "
		if g.settingsMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}

		label := prefix + row
		switch row {
		case "Font Style":
			label += fmt.Sprintf(": [%s]", strings.ToUpper(FontOptions[g.settingsFontIndex]))
		case "Sound Effects":
			label += fmt.Sprintf(": [%s]", strings.ToUpper(FrequencyOptions[g.settingsAudioIndex]))
		case "Fog of War":
			label += fmt.Sprintf(": [%s]", strings.ToUpper(FogOfWarOptions[g.settingsFogIndex]))
		case "AI Provider":
			label += fmt.Sprintf(": [%s]", strings.ToUpper(g.settings.AIProvider))
		case "AI Model":
			model := g.getAIModelForCurrentProvider()
			if model == "" {
				model = "AUTO"
			}
			if g.isFetchingModels {
				label += ": [FETCHING...]"
			} else {
				label += fmt.Sprintf(": [%s]", strings.ToUpper(model))
			}
		case "Simulation Mode":
			val := "OFF"
			if g.settings.AISimulationMode {
				val = "ON"
			}
			label += fmt.Sprintf(": [%s]", val)
		case "Talking Frequency":
			label += fmt.Sprintf(": [%s]", strings.ToUpper(g.settings.TalkingFrequency))
		case "Measurement Units":
			label += fmt.Sprintf(": [%s]", strings.ToUpper(g.settings.Units))
		}

		lw, _ := gr.graphics.MeasureText(label, 18)
		gr.graphics.DrawTextAt(screen, label, (g.width-int(lw))/2, 200+i*40, clr, 18)
	}

	hint := "UP/DOWN to navigate, LEFT/RIGHT to change value, ENTER to confirm."
	hw, _ := gr.graphics.MeasureText(hint, 14)
	gr.graphics.DrawTextAt(screen, hint, (g.width-int(hw))/2, g.height-120, color.RGBA{136, 136, 136, 255}, 14)

	yamlHint := "Edit API keys in ~/.oinakos/settings.yml"
	yhw, _ := gr.graphics.MeasureText(yamlHint, 12)
	gr.graphics.DrawTextAt(screen, yamlHint, (g.width-int(yhw))/2, g.height-90, color.RGBA{150, 150, 150, 255}, 12)
}

func (gr *GameRenderer) drawKeymapScreen(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	title := "CHARACTER KEYMAP"
	tw, _ := gr.graphics.MeasureText(title, 40)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 100, color.RGBA{218, 165, 32, 255}, 40)

	keys := []struct {
		Key    string
		Action string
	}{
		{"SPACE", "Attack enemies / Interact with Wells"},
		{"C", "Chop Trees (Auto-equips Axe from inventory)"},
		{"V", "Dig Ground (Auto-equips Pike from inventory)"},
		{"I", "Toggle Inventory & Equipment Menu"},
		{"R", "Toggle Resting State (Regain stamina/energy)"},
		{"TAB", "Toggle Debug Boundaries (Collision View)"},
		{"ESC", "Menu / Back / Pause"},
	}

	for i, k := range keys {
		gr.graphics.DrawTextAt(screen, k.Key, 150, 220+i*50, color.RGBA{255, 255, 0, 255}, 20)
		gr.graphics.DrawTextAt(screen, ": "+k.Action, 250, 220+i*50, color.White, 18)
	}

	hint := "Press ESC or ENTER to return to Settings"
	hw, _ := gr.graphics.MeasureText(hint, 16)
	gr.graphics.DrawTextAt(screen, hint, (g.width-int(hw))/2, g.height-100, color.RGBA{136, 136, 136, 255}, 16)
}

func (gr *GameRenderer) drawPauseMenu(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)

	title := "GAME PAUSED"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-50, color.White, 32)

	msg1 := "Press S to SAVE and QUIT"
	msg2 := "Press any other key to RESUME"
	mw1, _ := gr.graphics.MeasureText(msg1, 18)
	mw2, _ := gr.graphics.MeasureText(msg2, 18)
	gr.graphics.DrawTextAt(screen, msg1, (g.width-int(mw1))/2, g.height/2, color.White, 18)
	gr.graphics.DrawTextAt(screen, msg2, (g.width-int(mw2))/2, g.height/2+30, color.White, 18)
}

func (gr *GameRenderer) drawQuitConfirmation(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 200}, false)

	pw, ph := 400, 200
	px, py := (g.width-pw)/2, (g.height-ph)/2
	gr.graphics.DrawFilledRect(screen, float32(px), float32(py), float32(pw), float32(ph), color.RGBA{30, 30, 30, 255}, false)

	gr.graphics.DrawLine(screen, float32(px), float32(py), float32(px+pw), float32(py), color.White, 2)
	gr.graphics.DrawLine(screen, float32(px+pw), float32(py), float32(px+pw), float32(py+ph), color.White, 2)
	gr.graphics.DrawLine(screen, float32(px+pw), float32(py+ph), float32(px), float32(py+ph), color.White, 2)
	gr.graphics.DrawLine(screen, float32(px), float32(py+ph), float32(px), float32(py), color.White, 2)

	msg := "Really quit?"
	tw, _ := gr.graphics.MeasureText(msg, 24)
	gr.graphics.DrawTextAt(screen, msg, px+(pw-int(tw))/2, py+50, color.White, 24)

	options := []string{"Yes, quit", "No, stay here"}
	if !g.isMainMenu {
		options = []string{"Quit to menu", "Cancel"}
	}
	for i, opt := range options {
		var clr color.Color = color.White
		if i == g.quitConfirmationIndex {
			clr = color.RGBA{255, 255, 0, 255}
		}
		gr.graphics.DrawTextAt(screen, opt, px+100, py+100+i*40, clr, 20)
	}
}

func (gr *GameRenderer) drawGameOver(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)
	minutes := int(g.playTime) / 60
	seconds := int(g.playTime) % 60

	title := "GAME OVER"
	tw, _ := gr.graphics.MeasureText(title, 48)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-80, color.White, 48)

	kills := fmt.Sprintf("Kills: %d", g.playableCharacter.Kills)
	time := fmt.Sprintf("Time: %02d:%02d", minutes, seconds)
	msg := "Press ESC to exit, or CLICK/ENTER to restart"

	kw, _ := gr.graphics.MeasureText(kills, 20)
	tmw, _ := gr.graphics.MeasureText(time, 20)
	mw, _ := gr.graphics.MeasureText(msg, 16)

	gr.graphics.DrawTextAt(screen, kills, (g.width-int(kw))/2, g.height/2-15, color.White, 20)
	gr.graphics.DrawTextAt(screen, time, (g.width-int(tmw))/2, g.height/2+20, color.White, 20)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(mw))/2, g.height/2+60, color.White, 16)
}

func (gr *GameRenderer) drawMapWon(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{20, 60, 20, 200}, false)
	mapKillTotal := 0
	for _, k := range g.playableCharacter.MapKills {
		mapKillTotal += k
	}

	title := "MAP WON!"
	tw, _ := gr.graphics.MeasureText(title, 48)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-80, color.White, 48)

	kills := fmt.Sprintf("Map Kills: %d", mapKillTotal)
	kw, _ := gr.graphics.MeasureText(kills, 20)
	gr.graphics.DrawTextAt(screen, kills, (g.width-int(kw))/2, g.height/2-15, color.White, 20)

	options := []string{"Continue", "Quit"}
	for i, opt := range options {
		var clr color.Color = color.White
		prefix := "  "
		if g.mapWonMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}
		gr.graphics.DrawTextAt(screen, prefix+opt, g.width/2-50, g.height/2+60+i*35, clr, 18)
	}
}

func (gr *GameRenderer) drawGameWon(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 200}, false)

	title := "YOU WIN!"
	if g.isCampaign {
		title = "CAMPAIGN COMPLETED: YOU WIN!"
	}
	tw, _ := gr.graphics.MeasureText(title, 40)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 100, color.RGBA{218, 165, 32, 255}, 40)

	options := []string{"Replay", "Quit"}
	for i, opt := range options {
		var clr color.Color = color.White
		prefix := "  "
		if g.mapWonMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}
		gr.graphics.DrawTextAt(screen, prefix+opt, g.width/2-50, 200+i*40, clr, 20)
	}
}
func (gr *GameRenderer) drawInventoryScreen(screen engine.Image) {
	g := gr.game
	
	// Transparent background (very faint)
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 50}, false)
	
	dialogW, dialogH := 900, 600
	dialogX := (g.width - dialogW) / 2
	dialogY := (g.height - dialogH) / 2
	
	// Dialog background
	gr.graphics.DrawFilledRect(screen, float32(dialogX-2), float32(dialogY-2), float32(dialogW+4), float32(dialogH+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(dialogX), float32(dialogY), float32(dialogW), float32(dialogH), color.RGBA{15, 15, 15, 245}, false)
	
	title := "INVENTORY & EQUIPMENT"
	tw, _ := gr.graphics.MeasureText(title, 36)
	gr.graphics.DrawTextAt(screen, title, dialogX+(dialogW-int(tw))/2, dialogY+45, color.RGBA{218, 165, 32, 255}, 36)
	
	// Weight
	pc := g.playableCharacter
	weightStr := fmt.Sprintf("Weight: %s / %s", g.settings.FormatWeight(pc.GetTotalWeight()), g.settings.FormatWeight(pc.MaxWeight))
	ww, _ := gr.graphics.MeasureText(weightStr, 14)
	gr.graphics.DrawTextAt(screen, weightStr, dialogX+dialogW-int(ww)-40, dialogY+45, color.White, 14)

	// Traumas (Listed vertically below weight if any exist)
	traumas := pc.GetActiveTraumas()
	if len(traumas) > 0 {
		gr.graphics.DrawTextAt(screen, "TRAUMAS:", dialogX+dialogW-int(ww)-30, dialogY+70, color.RGBA{200, 50, 50, 255}, 12)
		for i, t := range traumas {
			gr.graphics.DrawTextAt(screen, "- "+t, dialogX+dialogW-int(ww)-30, dialogY+90+i*18, color.RGBA{220, 100, 100, 255}, 11)
		}
	}
	
	// Paper Doll Layout (Left side of the dialog)
	dollCenterX := dialogX + 220
	dollCenterY := dialogY + 300
	
	slots := []struct {
		id    string
		label string
		x     int
		y     int
	}{
		{"head", "Head", dollCenterX, dollCenterY - 140},
		{"shield", "Left Arm", dollCenterX - 110, dollCenterY - 40},
		{"body", "Torso", dollCenterX, dollCenterY - 40},
		{"weapon", "Right Arm", dollCenterX + 110, dollCenterY - 40},
		{"ring1", "Left Ring", dollCenterX - 110, dollCenterY + 50},
		{"ring2", "Right Ring", dollCenterX + 110, dollCenterY + 50},
		{"legs", "Legs/Feet", dollCenterX, dollCenterY + 110},
	}
	
	for _, slot := range slots {
		sx, sy := slot.x, slot.y
		
		// Slot Box
		gr.graphics.DrawTextAt(screen, slot.label, sx-40, sy-35, color.Gray16{0xAAAA}, 14)
		
		it, hasIt := pc.Slots[slot.id]
		if hasIt && it != nil && it.Config != nil {
			// Draw Item Sprite
			if it.Config.Sprite != nil {
				sw, sh := it.Config.Sprite.Size()
				op := engine.NewDrawImageOptions()
				
				// Scale to fit roughly 32x32 within the box
				scale := 32.0 / float64(sh)
				if float64(sw) > float64(sh) {
					scale = 32.0 / float64(sw)
				}
				op.GeoM.Scale(scale, scale)
				
				// Center in the box (80x50)
				tx := float64(sx) - (float64(sw)*scale)/2
				ty := float64(sy) - (float64(sh)*scale)/2 - 5
				op.GeoM.Translate(tx, ty)
				screen.DrawImage(it.Config.Sprite, op)
			}

			// Ensure names fit
			nameStr := it.Config.Name
			if len(nameStr) > 15 {
				nameStr = nameStr[:13] + ".."
			}
			nw, _ := gr.graphics.MeasureText(nameStr, 14)
			gr.graphics.DrawTextAt(screen, nameStr, sx-int(nw)/2, sy+20, color.RGBA{218, 165, 32, 255}, 14)
			
			// Draw Resistance if applicable
			if it.Config.Resistance > 0 {
				resStr := fmt.Sprintf("%d/%d", it.Resistance, it.Config.Resistance)
				rw, _ := gr.graphics.MeasureText(resStr, 12)
				gr.graphics.DrawTextAt(screen, resStr, sx-int(rw)/2, sy+38, color.RGBA{200, 200, 100, 255}, 12)
			}
			
			// Drop 'X' Label for equipment (No red square)
			gr.graphics.DrawTextAt(screen, "[X]", sx+30, sy-15, color.RGBA{200, 50, 50, 255}, 14)
		} else {
			em, _ := gr.graphics.MeasureText("Empty", 14)
			gr.graphics.DrawTextAt(screen, "Empty", sx-int(em)/2, sy+5, color.RGBA{100, 100, 100, 255}, 14)
		}
	}
	
	// Inventory List (Right side)
	listStartX := dialogX + 400
	listStartY := dialogY + 80
	listW := dialogW - 420
	
	gr.graphics.DrawTextAt(screen, "Backpack", listStartX, listStartY-10, color.RGBA{218, 165, 32, 255}, 20)
	
	if len(pc.Inventory) == 0 {
		emptyMsg := "Backpack is empty."
		ew, _ := gr.graphics.MeasureText(emptyMsg, 16)
		gr.graphics.DrawTextAt(screen, emptyMsg, listStartX+(listW-int(ew))/2, listStartY+50, color.Gray16{0x8888}, 16)
	} else {
		for i, it := range pc.Inventory {
			if it == nil || it.Config == nil { continue }
			itemY := listStartY + 20 + i*40
			
			
			// Draw Item Icon
			if it.Config.Sprite != nil {
				sw, sh := it.Config.Sprite.Size()
				op := engine.NewDrawImageOptions()
				scale := 24.0 / float64(sh)
				if float64(sw) > float64(sh) {
					scale = 24.0 / float64(sw)
				}
				op.GeoM.Scale(scale, scale)
				tx := float64(listStartX) + 5 + (24.0 - float64(sw)*scale)/2
				ty := float64(itemY) + 5 + (24.0 - float64(sh)*scale)/2
				op.GeoM.Translate(tx, ty)
				screen.DrawImage(it.Config.Sprite, op)
			}

			nameStr := it.Config.Name
			gr.graphics.DrawTextAt(screen, nameStr, listStartX+50, itemY+22, color.RGBA{218, 165, 32, 255}, 18)
			
			descStr := it.Config.Description
			if len(descStr) > 35 {
				descStr = descStr[:32] + "..."
			}
			gr.graphics.DrawTextAt(screen, descStr, listStartX+220, itemY+23, color.RGBA{180, 180, 180, 255}, 13)
			
			// Show durability next to description
			if it.Config.Resistance > 0 {
				durStr := fmt.Sprintf("Dur: %d/%d", it.Resistance, it.Config.Resistance)
				gr.graphics.DrawTextAt(screen, durStr, listStartX+450, itemY+23, color.RGBA{200, 200, 100, 255}, 13)
			}
			
			// Drop 'X' Button for inventory (No red square)
			gr.graphics.DrawTextAt(screen, "[X]", listStartX+listW-35, itemY+25, color.RGBA{200, 50, 50, 255}, 16)

			// Read 'R' Button for readable items
			if it.Config.Content != "" {
				gr.graphics.DrawTextAt(screen, "[R]", listStartX+listW-75, itemY+25, color.RGBA{0, 150, 255, 255}, 16)
			}
		}
	}
	
	// Close Button
	closeW := 100
	closeX := dialogX + (dialogW - closeW) / 2
	closeY := dialogY + dialogH - 50
	
	cw, _ := gr.graphics.MeasureText("[CLOSE]", 20)
	gr.graphics.DrawTextAt(screen, "[CLOSE]", closeX+(closeW-int(cw))/2, closeY+22, color.RGBA{218, 165, 32, 255}, 20)

	if g.ActiveBook != nil {
		gr.drawBookOverlay(screen)
	}
}

func (gr *GameRenderer) drawBookOverlay(screen engine.Image) {
	g := gr.game
	book := g.ActiveBook

	// Dim background
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 200}, false)

	// Book Dialog
	dialogW, dialogH := 600, 400
	dialogX := (g.width - dialogW) / 2
	dialogY := (g.height - dialogH) / 2

	gr.graphics.DrawFilledRect(screen, float32(dialogX-2), float32(dialogY-2), float32(dialogW+4), float32(dialogH+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(dialogX), float32(dialogY), float32(dialogW), float32(dialogH), color.RGBA{20, 20, 25, 255}, false)

	// Title
	tw, _ := gr.graphics.MeasureText(book.Config.Name, 24)
	gr.graphics.DrawTextAt(screen, book.Config.Name, dialogX+(dialogW-int(tw))/2, dialogY+40, color.RGBA{218, 165, 32, 255}, 24)

	// Content
	gr.drawWrappedText(screen, book.Config.Content, dialogX+40, dialogY+80, dialogW-80, color.White, 16, dialogY+dialogH-60)

	// Exit Hint
	exitMsg := "[Click or ESC to close]"
	ew, _ := gr.graphics.MeasureText(exitMsg, 14)
	gr.graphics.DrawTextAt(screen, exitMsg, dialogX+(dialogW-int(ew))/2, dialogY+dialogH-20, color.RGBA{150, 150, 150, 255}, 14)
}

func (gr *GameRenderer) drawLoadingProgress(screen engine.Image) {
	g := gr.game
	// Reverted to black background as requested
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	msg := g.LoadingMessage
	if msg == "" {
		msg = "LOADING OINAKOS..."
	}

	tw, _ := gr.graphics.MeasureText(msg, 32)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(tw))/2, g.height/2-20, color.RGBA{218, 165, 32, 255}, 32)

	hint := "Please wait while assets are being prepared"
	hw, _ := gr.graphics.MeasureText(hint, 14)
	gr.graphics.DrawTextAt(screen, hint, (g.width-int(hw))/2, g.height/2+80, color.RGBA{180, 180, 180, 255}, 14)

	// Progress Bar
	prog := atomic.LoadInt32(&g.LoadingProgress)
	barW := 400
	barH := 10
	barX := (g.width - barW) / 2
	barY := g.height/2 + 30
	
	// Background of the bar
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(barY), float32(barW), float32(barH), color.RGBA{40, 40, 40, 255}, false)
	// Foreground (the actual progress)
	if prog > 0 {
		fillW := float32(barW) * (float32(prog) / 1000.0)
		gr.graphics.DrawFilledRect(screen, float32(barX), float32(barY), fillW, float32(barH), color.RGBA{218, 165, 32, 255}, false)
	}

	// Minimal tech indicator at the bottom
	percent := fmt.Sprintf("LOADING PROGRESS: %d%%", int(float64(prog)/10.0))
	pw, _ := gr.graphics.MeasureText(percent, 12)
	gr.graphics.DrawTextAt(screen, percent, g.width-int(pw)-20, g.height-30, color.RGBA{100, 100, 100, 255}, 12)
}
