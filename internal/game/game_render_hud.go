package game

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"oinakos/internal/engine"
)

func (gr *GameRenderer) drawHUD(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 10, 10, 350, 150, color.RGBA{0, 0, 0, 180}, false)

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("HP: %d/%d", g.playableCharacter.Health, g.playableCharacter.MaxHealth), 20, 20, color.White, 16)

	gr.graphics.DrawFilledRect(screen, 160, 22, 200, 10, color.RGBA{100, 0, 0, 255}, false)

	healthPct := float64(g.playableCharacter.Health) / float64(g.playableCharacter.MaxHealth)
	if healthPct > 0 {
		var healthColor color.RGBA
		if healthPct > 0.7 {
			healthColor = color.RGBA{0, 200, 0, 255}
		} else if healthPct > 0.5 {
			healthColor = color.RGBA{200, 200, 0, 255}
		} else {
			healthColor = color.RGBA{200, 0, 0, 255}
		}
		gr.graphics.DrawFilledRect(screen, 160, 22, float32(200*healthPct), 10, healthColor, false)
	}

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("LVL: %d  XP: %d", g.playableCharacter.Level, g.playableCharacter.XP), 20, 45, color.White, 14)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("OBJ: %s", g.currentMapType.Description), 20, 60, color.White, 12)

	minutes := int(g.playTime) / 60
	seconds := int(g.playTime) % 60
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("POS %.1f,%.1f  KILLS: %d  XP: %d  LVL: %d", g.playableCharacter.X, g.playableCharacter.Y, g.playableCharacter.Kills, g.playableCharacter.XP, g.playableCharacter.Level), 20, 80, color.White, 12)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("ATK: %d  DEF: %d  SHIELD: %d", g.playableCharacter.GetTotalAttack(), g.playableCharacter.GetTotalDefense(), g.playableCharacter.GetTotalProtection()), 20, 95, color.White, 12)

	weaponText := fmt.Sprintf("WEAPON: %s (%d-%d)", g.playableCharacter.Weapon.Name, g.playableCharacter.Weapon.MinDamage, g.playableCharacter.Weapon.MaxDamage)
	if g.playableCharacter.Weapon.Bonus > 0 {
		weaponText += fmt.Sprintf(" +%d", g.playableCharacter.Weapon.Bonus)
	}
	gr.graphics.DrawTextAt(screen, weaponText, 20, 110, color.White, 12)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("TIME: %02d:%02d", minutes, seconds), 20, 125, color.White, 12)

	gr.graphics.DrawFilledRect(screen, float32(g.width-110), 20, 100, 30, color.RGBA{50, 50, 50, 200}, false)
	gr.graphics.DrawTextAt(screen, "MENU", g.width-85, 28, color.White, 16)

	mapTitle := strings.ToUpper(g.currentMapType.Name)
	if g.isCampaign && g.currentCampaign != nil {
		mapTitle = strings.ToUpper(g.currentCampaign.Name)
	}
	mtw, _ := gr.graphics.MeasureText(mapTitle, 16)
	gr.graphics.DrawTextAt(screen, mapTitle, g.width-int(mtw)-20, 60, color.RGBA{218, 165, 32, 255}, 16)

	if g.isMenuOpen {
		mw_box, mh_box := 400, 280
		mx, my := (g.width-mw_box)/2, (g.height-mh_box)/2
		gr.graphics.DrawFilledRect(screen, float32(mx-2), float32(my-2), float32(mw_box+4), float32(mh_box+4), color.RGBA{218, 165, 32, 255}, false)
		gr.graphics.DrawFilledRect(screen, float32(mx), float32(my), float32(mw_box), float32(mh_box), color.RGBA{0, 0, 0, 240}, false)

		menuTitle := "GAME MENU"
		mtw2, _ := gr.graphics.MeasureText(menuTitle, 24)
		gr.graphics.DrawTextAt(screen, menuTitle, mx+(mw_box-int(mtw2))/2, my+30, color.RGBA{218, 165, 32, 255}, 24)

		options := []string{"Resume", "Quicksave (Q)", "Load", "Settings", "Quit"}
		for i, opt := range options {
			var clr color.Color = color.White
			prefix := "  "
			if g.menuIndex == i {
				clr = color.RGBA{255, 255, 0, 255}
				prefix = "> "
			}
			gr.graphics.DrawTextAt(screen, prefix+opt, mx+100, my+80+i*35, clr, 18)
		}
		instr := "Press ENTER to select"
		itw, _ := gr.graphics.MeasureText(instr, 14)
		gr.graphics.DrawTextAt(screen, instr, mx+(mw_box-int(itw))/2, my+mh_box-30, color.RGBA{136, 136, 136, 255}, 14)
	}

	if g.saveMessageTimer > 0 {
		msg := g.saveMessage
		sttw, _ := gr.graphics.MeasureText(msg, 18)
		gr.graphics.DrawTextAt(screen, msg, (g.width-int(sttw))/2, g.height-40, color.RGBA{218, 165, 32, 255}, 18)
	}

	gr.drawObjectiveArrow(screen)
}

func (gr *GameRenderer) drawHoverInfo(screen engine.Image) {
	g := gr.game
	mx, my := g.input.MousePosition()
	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)

	for _, n := range g.npcs {
		if !n.IsAlive() {
			continue
		}
		isoX, isoY := engine.CartesianToIso(n.X, n.Y)
		scrX, scrY := isoX+offsetX, isoY+offsetY

		dist := math.Sqrt(math.Pow(float64(mx)-scrX, 2) + math.Pow(float64(my)-scrY+40, 2))
		if dist < 40 {
			if n.Archetype != nil && n.Archetype.Description != "" {
				gr.drawInfoBox(screen, n.Name, n.Archetype.Description, mx, my)
				return
			}
		}
	}
}

func (gr *GameRenderer) drawInfoBox(screen engine.Image, title, desc string, x, y int) {
	boxW, boxH := 300.0, 160.0
	bx, by := float32(x+20), float32(y+20)

	if float64(bx)+boxW > float64(gr.game.width) {
		bx = float32(float64(x) - boxW - 20)
	}
	if float64(by)+boxH > float64(gr.game.height) {
		by = float32(float64(y) - boxH - 20)
	}

	gr.graphics.DrawFilledRect(screen, bx-2, by-2, float32(boxW+4), float32(boxH+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, bx, by, float32(boxW), float32(boxH), color.RGBA{0, 0, 0, 240}, false)
	gr.graphics.DebugPrintAt(screen, title, int(bx)+10, int(by)+10, color.RGBA{218, 165, 32, 255})

	words := strings.Fields(desc)
	line := ""
	lineNum := 0
	for _, w := range words {
		if len(line)+len(w) > 35 {
			gr.graphics.DebugPrintAt(screen, line, int(bx)+10, int(by)+35+lineNum*15, color.White)
			line = w + " "
			lineNum++
		} else {
			line += w + " "
		}
	}
	gr.graphics.DebugPrintAt(screen, line, int(bx)+10, int(by)+35+lineNum*15, color.White)
}
func (gr *GameRenderer) drawDialogueBox(screen engine.Image) {
	g := gr.game
	// Position: Bottom of the window
	boxH := 180
	isDialogue := g.ActiveDialogue != nil
	if isDialogue {
		if g.ActiveDialogue.UIState == DialogueMaximized {
			boxH = 600
		} else {
			boxH = 300
		}
	} else {
		if g.LogUIState == DialogueMaximized {
			boxH = 240 // Expanded Log
		} else {
			boxH = 60 // Slim Log
		}
	}

	boxW := g.width - 20
	bx, by := 10, g.height-boxH-10

	// Draw box background
	gr.graphics.DrawFilledRect(screen, float32(bx), float32(by), float32(boxW), float32(boxH), color.RGBA{0, 0, 0, 180}, false)
	// Hollow rect via polygon
	rectPts := []engine.Point{
		{X: float64(bx), Y: float64(by)},
		{X: float64(bx + boxW), Y: float64(by)},
		{X: float64(bx + boxW), Y: float64(by + boxH)},
		{X: float64(bx), Y: float64(by + boxH)},
		{X: float64(bx), Y: float64(by)},
	}
	gr.graphics.DrawPolygon(screen, rectPts, color.RGBA{218, 165, 32, 255}, 1)

	// Draw Event Log
	logY := by + 10
	var maxLogEntries int
	if isDialogue {
		maxLogEntries = 1
	} else {
		if g.LogUIState == DialogueMaximized {
			maxLogEntries = 4
		} else {
			maxLogEntries = 2
		}
	}

	startIdx := len(g.EventLog) - maxLogEntries - g.LogScrollOffset
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + maxLogEntries
	if endIdx > len(g.EventLog) {
		endIdx = len(g.EventLog)
	}

	for i := startIdx; i < endIdx; i++ {
		entry := g.EventLog[i]
		var clr color.Color = color.White
		switch entry.Category {
		case LogPlayer:
			clr = color.RGBA{0, 255, 255, 255}
		case LogNPC:
			clr = color.RGBA{255, 255, 255, 255}
		case LogCombatDamage:
			clr = color.RGBA{220, 20, 60, 255}
		case LogCombatRecovery:
			clr = color.RGBA{0, 255, 0, 255}
		}
		gr.graphics.DrawTextAt(screen, entry.Text, bx+10, logY, clr, 12)
		logY += 15
	}

	// Draw Scrollbar (if there's enough history)
	if len(g.EventLog) > maxLogEntries {
		sbW := float32(4)
		sbX := float32(bx + boxW - 10)
		sbTrackY := float32(by + 5)
		sbTrackH := float32(boxH - 10)
		
		// Track (darker)
		gr.graphics.DrawFilledRect(screen, sbX, sbTrackY, sbW, sbTrackH, color.RGBA{50, 50, 50, 150}, false)
		
		// Handle
		visibleRatio := float32(maxLogEntries) / float32(len(g.EventLog))
		handleH := sbTrackH * visibleRatio
		if handleH < 10 { handleH = 10 }
		
		maxOffset := float32(len(g.EventLog) - maxLogEntries)
		scrollRatio := float32(0)
		if maxOffset > 0 {
			scrollRatio = float32(g.LogScrollOffset) / maxOffset
		}
		// Offset 0 is BOTTOM (newest), maxOffset is TOP (oldest)
		handleY := sbTrackY + (sbTrackH - handleH) * (1.0 - scrollRatio)
		
		gr.graphics.DrawFilledRect(screen, sbX, handleY, sbW, handleH, color.RGBA{218, 165, 32, 200}, false)
	}

	// Draw Active Dialogue
	if isDialogue {
		diagY := by + 40
		gr.graphics.DrawLine(screen, float32(bx+5), float32(diagY), float32(bx+boxW-5), float32(diagY), color.RGBA{136, 136, 136, 255}, 1)
		diagY += 15

		speakerName := g.ActiveDialogue.SpeakerNPC.Name
		gr.graphics.DrawTextAt(screen, speakerName+":", bx+10, diagY, color.RGBA{218, 165, 32, 255}, 16)
		diagY += 25

		// Calculate choices area start to avoid overlap
		choiceYStart := by + boxH - 80 - len(g.ActiveDialogue.Choices)*22
		
		gr.drawWrappedText(screen, g.ActiveDialogue.CurrentText, bx+20, diagY, boxW-40, color.White, 14, choiceYStart-10)

		// Draw Choices
		choiceY := choiceYStart
		for i, choice := range g.ActiveDialogue.Choices {
			clr := color.RGBA{200, 200, 200, 255}
			prefix := "  "
			if i == g.ActiveDialogue.SelectedChoice {
				clr = color.RGBA{255, 255, 0, 255}
				prefix = "> "
			}
			gr.graphics.DrawTextAt(screen, prefix+choice.Text, bx+30, choiceY, clr, 14)
			choiceY += 22
		}
		
		gr.graphics.DrawTextAt(screen, "[ESC/BACKSPACE] Close", bx+boxW-195, by+boxH-60, color.RGBA{136, 136, 136, 255}, 11)
		
		// Draw maximize/minimize button hint
		btnTxt := "[+]"
		if g.ActiveDialogue.UIState == DialogueMaximized {
			btnTxt = "[-]"
		}
		gr.graphics.DrawTextAt(screen, btnTxt, bx+boxW-225, by+boxH-60, color.RGBA{218, 165, 32, 255}, 14)
	} else {
		// Hint for log expansion
		hintTxt := "Click to Expand Log"
		if g.LogUIState == DialogueMaximized {
			hintTxt = "Click to Shrink Log"
		}
		gr.graphics.DrawTextAt(screen, hintTxt, bx+boxW-150, by+boxH-25, color.RGBA{136, 136, 136, 255}, 10)
	}
}

func (gr *GameRenderer) drawWrappedText(screen engine.Image, text string, x, y, maxWidth int, clr color.Color, size int, maxY int) {
	words := strings.Fields(text)
	line := ""
	currY := y
	for _, w := range words {
		wWidth, _ := gr.graphics.MeasureText(line+w+" ", float64(size))
		if int(wWidth) > maxWidth {
			if currY > maxY { return }
			gr.graphics.DrawTextAt(screen, line, x, currY, clr, float64(size))
			line = w + " "
			currY += size + 6
		} else {
			line += w + " "
		}
	}
	if currY <= maxY {
		gr.graphics.DrawTextAt(screen, line, x, currY, clr, float64(size))
	}
}
