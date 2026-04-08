package game

import (
	"image/color"
	"strings"
	"oinakos/internal/engine"
)

func (gr *GameRenderer) drawDialogueBox(screen engine.Image) {
	g := gr.game
	boxH := 180
	isDialogue := g.ActiveDialogue != nil
	if isDialogue {
		if g.ActiveDialogue.UIState == DialogueMaximized { boxH = 650 } else { boxH = 350 }
	} else {
		if g.LogUIState == DialogueMaximized { boxH = 300 } else { boxH = 85 }
	}
	boxW, bx, by := g.width-20, 10, g.height-boxH-10
	gr.graphics.DrawFilledRect(screen, float32(bx), float32(by), float32(boxW), float32(boxH), color.RGBA{0, 0, 0, 180}, false)
	rectPts := []engine.Point{{X: float64(bx), Y: float64(by)}, {X: float64(bx+boxW), Y: float64(by)}, {X: float64(bx+boxW), Y: float64(by+boxH)}, {X: float64(bx), Y: float64(by+boxH)}, {X: float64(bx), Y: float64(by)}}
	gr.graphics.DrawPolygon(screen, rectPts, color.RGBA{218, 165, 32, 255}, 1)
	gr.graphics.DrawTextAt(screen, "-- EVENT LOG (History) --", bx+10, by+12, color.RGBA{150, 150, 150, 255}, 10)
	logY, maxLogEntries := by+30, 3
	if isDialogue { if g.ActiveDialogue.UIState == DialogueMaximized { maxLogEntries = 12 } else { maxLogEntries = 5 } } else if g.LogUIState == DialogueMaximized { maxLogEntries = 15 }
	startIdx := len(g.EventLog) - maxLogEntries - g.LogScrollOffset; if startIdx < 0 { startIdx = 0 }
	endIdx := startIdx + maxLogEntries; if endIdx > len(g.EventLog) { endIdx = len(g.EventLog) }
	for i := startIdx; i < endIdx; i++ {
		entry, clr := g.EventLog[i], color.Color(color.White)
		switch entry.Category { case LogPlayer: clr = color.RGBA{0, 255, 255, 255}; case LogCombatDamage: clr = color.RGBA{220, 20, 60, 255}; case LogCombatRecovery: clr = color.RGBA{0, 255, 0, 255} }
		gr.graphics.DrawTextAt(screen, entry.Text, bx+10, logY, clr, 12); logY += 15
	}
	if len(g.EventLog) > maxLogEntries {
		sbW, logAreaH := float32(4), float32(maxLogEntries*15+10)
		sbX, sbTrackY, sbTrackH := float32(bx+boxW-10), float32(by+25), logAreaH
		gr.graphics.DrawFilledRect(screen, sbX, sbTrackY, sbW, sbTrackH, color.RGBA{50, 50, 50, 150}, false)
		visRatio := float32(maxLogEntries) / float32(len(g.EventLog))
		handleH := sbTrackH * visRatio; if handleH < 10 { handleH = 10 }
		maxOff, scrollRatio := float32(len(g.EventLog)-maxLogEntries), float32(0)
		if maxOff > 0 { scrollRatio = float32(g.LogScrollOffset) / maxOff }
		handleY := sbTrackY + (sbTrackH-handleH)*(1.0-scrollRatio)
		gr.graphics.DrawFilledRect(screen, sbX, handleY, sbW, handleH, color.RGBA{218, 165, 32, 200}, false)
	}
	diagYStart := float32(logY) + 10; if !isDialogue { diagYStart = float32(by) + float32(boxH) - 25 }
	if isDialogue {
		gr.graphics.DrawLine(screen, float32(bx+5), float32(diagYStart-5), float32(bx+boxW-5), float32(diagYStart-5), color.RGBA{136, 136, 136, 255}, 1)
		diagY := diagYStart + 10; gr.graphics.DrawTextAt(screen, g.ActiveDialogue.SpeakerNPC.Name+":", bx+10, int(diagY), color.RGBA{218, 165, 32, 255}, 16)
		diagY += 25; choiceYStart := float32(by) + float32(boxH) - 80 - float32(len(g.ActiveDialogue.Choices))*22
		gr.drawWrappedText(screen, g.ActiveDialogue.CurrentText, bx+20, int(diagY), boxW-40, color.White, 14, int(choiceYStart)-10)
		choiceY := choiceYStart
		for i, choice := range g.ActiveDialogue.Choices {
			clr, prefix := color.RGBA{200, 200, 200, 255}, "  "; if i == g.ActiveDialogue.SelectedChoice { clr, prefix = color.RGBA{255, 255, 0, 255}, "> " }
			gr.graphics.DrawTextAt(screen, prefix+choice.Text, bx+30, int(choiceY), clr, 14); choiceY += 22
		}
		gr.graphics.DrawTextAt(screen, "[ESC/BACKSPACE] Close", bx+boxW-195, by+boxH-60, color.RGBA{136, 136, 136, 255}, 11)
		btnTxt := "[+]"; if g.ActiveDialogue.UIState == DialogueMaximized { btnTxt = "[-]" }; gr.graphics.DrawTextAt(screen, btnTxt, bx+boxW-225, by+boxH-60, color.RGBA{218, 165, 32, 255}, 14)
	} else {
		hintTxt := "Click to Expand Log"; if g.LogUIState == DialogueMaximized { hintTxt = "Click to Shrink Log" }; gr.graphics.DrawTextAt(screen, hintTxt, bx+boxW-150, by+boxH-25, color.RGBA{136, 136, 136, 255}, 10)
	}
}

func (gr *GameRenderer) drawWrappedText(screen engine.Image, text string, x, y, maxWidth int, clr color.Color, size int, maxY int) int {
	words, line, currY := strings.Fields(text), "", y
	for _, w := range words {
		wW, _ := gr.graphics.MeasureText(line+w+" ", float64(size))
		if int(wW) > maxWidth {
			if currY <= maxY { gr.graphics.DrawTextAt(screen, line, x, currY, clr, float64(size)) }
			line, currY = w+" ", currY+size+6
		} else { line += w + " " }
	}
	if currY <= maxY { gr.graphics.DrawTextAt(screen, line, x, currY, clr, float64(size)) }
	return currY + size + 6
}
