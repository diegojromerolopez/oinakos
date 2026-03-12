package game

import (
	"fmt"
	"image/color"
	"strings"
	"sync/atomic"
	"oinakos/internal/engine"
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

	options := []string{"NEW GAME", "LOAD GAME", "SETTINGS", "QUIT"}

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

	gr.graphics.DrawTextAt(screen, "v0.9.0-alpha", 20, g.height-30, color.RGBA{80, 80, 80, 255}, 14)
}

func (gr *GameRenderer) drawCharacterSelect(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	title := "OINAKOS: CHOOSE YOUR HERO"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 50, color.RGBA{218, 165, 32, 255}, 32)

	mx, my := g.input.MousePosition()

	for i, id := range g.playableCharacterRegistry.IDs {
		char := g.playableCharacterRegistry.Characters[id]
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
	// Back button
	var clrBack color.Color = color.White
	prefixBack := "  "
	if g.characterMenuIndex == len(g.playableCharacterRegistry.IDs) {
		clrBack = color.RGBA{255, 255, 0, 255}
		prefixBack = "> "
	}
	gr.graphics.DrawTextAt(screen, prefixBack+"BACK", 100, 130+len(g.playableCharacterRegistry.IDs)*35+20, clrBack, 18)

	msg := "Press UP/DOWN to navigate, ENTER to select hero."
	mw, _ := gr.graphics.MeasureText(msg, 14)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(mw))/2, g.height-50, color.RGBA{136, 136, 136, 255}, 14)
}

func (gr *GameRenderer) drawHeroPreview(screen engine.Image, char *EntityConfig, x, y int) {
	gr.graphics.DrawTextAt(screen, "--- HERO PROFILE ---", x, y, color.RGBA{218, 165, 32, 255}, 20)

	if char.StaticImage != nil {
		img := char.StaticImage.(engine.Image)
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
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Weapon:  %s", char.WeaponName), statsX, statsY+100, color.White, 16)

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

	rows := []string{"Font Style", "Sound Effects", "Fog of War", "Save and Back"}
	for i, row := range rows {
		var clr color.Color = color.White
		prefix := "  "
		if g.settingsMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}

		label := prefix + row
		if i == 0 {
			label += fmt.Sprintf(": [%s]", strings.ToUpper(FontOptions[g.settingsFontIndex]))
		} else if i == 1 {
			label += fmt.Sprintf(": [%s]", strings.ToUpper(FrequencyOptions[g.settingsAudioIndex]))
		} else if i == 2 {
			label += fmt.Sprintf(": [%s]", strings.ToUpper(FogOfWarOptions[g.settingsFogIndex]))
		}

		lw, _ := gr.graphics.MeasureText(label, 24)
		gr.graphics.DrawTextAt(screen, label, (g.width-int(lw))/2, 250+i*60, clr, 24)
	}

	hint := "UP/DOWN to navigate, LEFT/RIGHT to change value, ENTER to confirm."
	hw, _ := gr.graphics.MeasureText(hint, 14)
	gr.graphics.DrawTextAt(screen, hint, (g.width-int(hw))/2, g.height-100, color.RGBA{136, 136, 136, 255}, 14)
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
func (gr *GameRenderer) drawLoadingProgress(screen engine.Image) {
	g := gr.game
	// Reverted to black background as requested
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	msg := g.LoadingMessage
	if msg == "" {
		msg = "LOADING OINAKOS..."
	}
	
	tw, _ := gr.graphics.MeasureText(msg, 32)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(tw))/2, g.height/2, color.RGBA{218, 165, 32, 255}, 32)

	hint := "Please wait while assets are being prepared"
	hw, _ := gr.graphics.MeasureText(hint, 14)
	gr.graphics.DrawTextAt(screen, hint, (g.width-int(hw))/2, g.height/2+50, color.RGBA{180, 180, 180, 255}, 14)

	// Minimal tech indicator at the bottom
	prog := atomic.LoadInt32(&g.LoadingProgress)
	percent := fmt.Sprintf("LOADING PROGRESS: %d%%", int(float64(prog)/10.0))
	pw, _ := gr.graphics.MeasureText(percent, 12)
	gr.graphics.DrawTextAt(screen, percent, g.width-int(pw)-20, g.height-30, color.RGBA{100, 100, 100, 255}, 12)
}
