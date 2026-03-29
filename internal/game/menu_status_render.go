package game

import (
	"fmt"
	"image/color"
	"oinakos/internal/engine"
	"sync/atomic"
)

func (gr *GameRenderer) drawPauseMenu(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)
	title := "GAME PAUSED"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-50, color.White, 32)
	msg1, msg2 := "Press S to SAVE and QUIT", "Press any other key to RESUME"
	mw1, _ := gr.graphics.MeasureText(msg1, 18); gr.graphics.DrawTextAt(screen, msg1, (g.width-int(mw1))/2, g.height/2, color.White, 18)
	mw2, _ := gr.graphics.MeasureText(msg2, 18); gr.graphics.DrawTextAt(screen, msg2, (g.width-int(mw2))/2, g.height/2+30, color.White, 18)
}

func (gr *GameRenderer) drawQuitConfirmation(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 200}, false)
	pw, ph := 400, 200; px, py := (g.width-pw)/2, (g.height-ph)/2
	gr.graphics.DrawFilledRect(screen, float32(px), float32(py), float32(pw), float32(ph), color.RGBA{30, 30, 30, 255}, false)
	msg := "Really quit?"
	tw, _ := gr.graphics.MeasureText(msg, 24); gr.graphics.DrawTextAt(screen, msg, px+(pw-int(tw))/2, py+50, color.White, 24)
	options := []string{"Yes, quit", "No, stay here"}
	if !g.isMainMenu { options = []string{"Quit to menu", "Cancel"} }
	for i, opt := range options {
		var clr color.Color = color.White; if i == g.quitConfirmationIndex { clr = color.RGBA{255, 255, 0, 255} }
		gr.graphics.DrawTextAt(screen, opt, px+100, py+100+i*40, clr, 20)
	}
}

func (gr *GameRenderer) drawGameOver(screen engine.Image) {
	g := gr.game; gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)
	title := "GAME OVER"; tw, _ := gr.graphics.MeasureText(title, 48); gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-110, color.White, 48)

	reason := fmt.Sprintf("Your character %s %s.", g.playableCharacter.Name, g.deathReason)
	rtw, _ := gr.graphics.MeasureText(reason, 22)
	gr.graphics.DrawTextAt(screen, reason, (g.width-int(rtw))/2, g.height/2-40, color.RGBA{200, 50, 50, 255}, 22)

	kills := fmt.Sprintf("Kills: %d", g.playableCharacter.Kills); time := fmt.Sprintf("Time: %02d:%02d", int(g.playTime)/60, int(g.playTime)%60)
	gr.graphics.DrawTextAt(screen, kills, (g.width-120)/2, g.height/2, color.White, 20)
	gr.graphics.DrawTextAt(screen, time, (g.width-120)/2, g.height/2+35, color.White, 20)
	gr.graphics.DrawTextAt(screen, "Press ESC to exit, or CLICK/ENTER to restart", (g.width-400)/2, g.height/2+80, color.White, 16)
}

func (gr *GameRenderer) drawMapWon(screen engine.Image) {
	g := gr.game; gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{20, 60, 20, 200}, false)
	mk := 0; for _, k := range g.playableCharacter.MapKills { mk += k }
	title := "MAP WON!"; tw, _ := gr.graphics.MeasureText(title, 48); gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-80, color.White, 48)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Map Kills: %d", mk), (g.width-120)/2, g.height/2-15, color.White, 20)
	options := []string{"Continue", "Quit"}
	for i, opt := range options {
		var clr color.Color = color.White; if g.mapWonMenuIndex == i { clr = color.RGBA{255, 255, 0, 255} }
		gr.graphics.DrawTextAt(screen, "  "+opt, g.width/2-50, g.height/2+60+i*35, clr, 18)
	}
}

func (gr *GameRenderer) drawGameWon(screen engine.Image) {
	g := gr.game; gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 200}, false)
	title := "YOU WIN!"
	if g.isCampaign { title = "CAMPAIGN COMPLETED: YOU WIN!" }
	tw, _ := gr.graphics.MeasureText(title, 40); gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 100, color.RGBA{218, 165, 32, 255}, 40)
	options := []string{"Replay", "Quit"}
	for i, opt := range options {
		var clr color.Color = color.White; if g.mapWonMenuIndex == i { clr = color.RGBA{255, 255, 0, 255} }
		gr.graphics.DrawTextAt(screen, "  "+opt, g.width/2-50, 200+i*40, clr, 20)
	}
}

func (gr *GameRenderer) drawLoadingProgress(screen engine.Image) {
	g := gr.game; gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)
	msg := g.LoadingMessage; if msg == "" { msg = "LOADING OINAKOS..." }
	tw, _ := gr.graphics.MeasureText(msg, 32); gr.graphics.DrawTextAt(screen, msg, (g.width-int(tw))/2, g.height/2-20, color.RGBA{218, 165, 32, 255}, 32)
	prog := atomic.LoadInt32(&g.LoadingProgress)
	barW, barH, barX, barY := 400, 10, (g.width-400)/2, g.height/2+30
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(barY), float32(barW), float32(barH), color.RGBA{40, 40, 40, 255}, false)
	if prog > 0 { gr.graphics.DrawFilledRect(screen, float32(barX), float32(barY), float32(barW)*(float32(prog)/1000.0), float32(barH), color.RGBA{218, 165, 32, 255}, false) }
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("LOAD: %d%%", int(float64(prog)/10.0)), g.width-120, g.height-30, color.RGBA{100, 100, 100, 255}, 12)
}
