package game

import (
	"fmt"
	"image/color"
	"oinakos/internal/engine"
)

func (gr *GameRenderer) drawHUD(screen engine.Image) {
	g := gr.game; sState := g.World.State
	gr.graphics.DrawFilledRect(screen, 10, 10, 220, 50, color.RGBA{0, 0, 0, 180}, false)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("%s, HOUR %02d:%02d", sState.Season, sState.Hour, sState.Ticks/12), 20, 25, color.RGBA{218, 165, 32, 255}, 14)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("TEMP: %.1fC WEATHER: %s", sState.Temperature, sState.Weather), 20, 45, color.White, 11)
	
	hudY, barX, barW := 70, 160, 200
	if !g.isHUDVisible {
		gr.graphics.DrawFilledRect(screen, 10, float32(hudY), 25, 25, color.RGBA{0, 0, 0, 180}, false)
		gr.graphics.DrawTextAt(screen, "->", 15, hudY+18, color.RGBA{218, 165, 32, 255}, 14)
		if g.isMenuOpen { gr.drawMenu(screen) }
		if g.saveMessageTimer > 0 { sttw, _ := gr.graphics.MeasureText(g.saveMessage, 18); gr.graphics.DrawTextAt(screen, g.saveMessage, (g.width-int(sttw))/2, g.height-40, color.RGBA{218, 165, 32, 255}, 18) }
		gr.drawMinimap(screen); gr.drawObjectiveArrow(screen)
		return
	}

	gr.graphics.DrawFilledRect(screen, 10, float32(hudY), 350, 410, color.RGBA{0, 0, 0, 180}, false)
	gr.graphics.DrawTextAt(screen, "<-", 15, hudY+12, color.RGBA{218, 165, 32, 255}, 14) // Move arrow to top-LEFT
	hp, maxHP := g.playableCharacter.State.HealthPoints, g.playableCharacter.State.MaxHealthPoints
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("HP: %d/%d", hp, maxHP), 20, hudY+35, color.White, 16) // Shift content further DOWN
	gr.graphics.DrawFilledRect(screen, 160, float32(hudY+37), 200, 10, color.RGBA{100, 0, 0, 255}, false)
	if healthPct := float64(hp) / float64(maxHP); healthPct > 0 {
		healthColor := color.RGBA{0, 200, 0, 255}
		if healthPct <= 0.5 { healthColor = color.RGBA{200, 0, 0, 255} } else if healthPct <= 0.7 { healthColor = color.RGBA{200, 200, 0, 255} }
		gr.graphics.DrawFilledRect(screen, 160, float32(hudY+12), float32(200*healthPct), 10, healthColor, false)
	}

	p := g.playableCharacter
	isVampire := p.State.Age.Rate == 0
	type needDef struct { label string; val float64; clr color.RGBA }
	var needs []needDef
	if isVampire { needs = append(needs, needDef{"BLOODLUST", p.State.Hunger, color.RGBA{180, 0, 0, 255}})
	} else { needs = append(needs, needDef{"HUNGER", p.State.Hunger, color.RGBA{210, 105, 30, 255}}, needDef{"THIRST", p.State.Thirst, color.RGBA{0, 191, 255, 255}}) }
	needs = append(needs, []needDef{
		{"FATIGUE", p.State.Fatigue, color.RGBA{255, 215, 0, 255}}, {"HYGIENE", p.State.Hygiene, color.RGBA{144, 238, 144, 255}},
		{"BLADDER", p.State.BladderLevel, color.RGBA{255, 255, 0, 255}}, {"BOWEL", p.State.BowelLevel, color.RGBA{139, 69, 19, 255}},
		{"SANITY", p.State.Sanity, color.RGBA{0, 255, 255, 255}}, {"AROUSAL", p.State.Arousal, color.RGBA{255, 105, 180, 255}},
		{"PAIN", p.State.Pain, color.RGBA{255, 0, 0, 255}}, {"ALCOHOL", p.State.AlcoholLevel, color.RGBA{150, 50, 250, 255}},
	}...)

	for i, n := range needs {
		gr.graphics.DrawTextAt(screen, fmt.Sprintf("%s:%d%%", n.label, int(n.val)), 20, hudY+50+i*13, color.White, 12)
		gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+52+i*13), float32(barW), 6, color.RGBA{30, 30, 30, 255}, false)
		gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+52+i*13), float32(float64(barW)*(n.val/100.0)), 6, n.clr, false)
	}
	yPos := hudY + 183 + 15; if isVampire { yPos -= 13 }
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("LV:%d XP:%d AGE:%.1f GLD:%d", p.Level, p.XP, p.State.Age.Current, p.Denarii), 20, yPos, color.White, 12)
	yPos = gr.drawWrappedText(screen, "OBJ: "+g.currentMapType.Description, 20, yPos+15, 310, color.White, 12, hudY+315); yPos += 15
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("POS:%s,%s KILLS:%d", g.settings.FormatDistance(p.X), g.settings.FormatDistance(p.Y), p.Kills), 20, yPos, color.White, 12); yPos += 15
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("STR:%d DEX:%d HEA:%d INT:%d WIS:%d", p.PrimaryAttributes.Strength, p.PrimaryAttributes.Dexterity, p.PrimaryAttributes.Health, p.PrimaryAttributes.Intellect, p.PrimaryAttributes.Wisdom), 20, yPos, color.RGBA{200, 200, 255, 255}, 12); yPos += 15
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("ATK:%d DEF:%d SHD:%d WT:%s/%s", p.GetTotalAttack(), p.GetTotalDefense(), p.GetTotalProtection(), g.settings.FormatWeight(p.GetTotalWeight()), g.settings.FormatWeight(p.MaxWeight)), 20, yPos, color.White, 12)
	
	if g.isMenuOpen { gr.drawMenu(screen) }
	if g.saveMessageTimer > 0 { sttw, _ := gr.graphics.MeasureText(g.saveMessage, 18); gr.graphics.DrawTextAt(screen, g.saveMessage, (g.width-int(sttw))/2, g.height-40, color.RGBA{218, 165, 32, 255}, 18) }
	gr.drawMinimap(screen); gr.drawObjectiveArrow(screen)
}

func (gr *GameRenderer) drawMenu(screen engine.Image) {
	g, mw_box, mh_box := gr.game, 400, 280
	mx, my := (g.width-mw_box)/2, (g.height-mh_box)/2
	gr.graphics.DrawFilledRect(screen, float32(mx-2), float32(my-2), float32(mw_box+4), float32(mh_box+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(mx), float32(my), float32(mw_box), float32(mh_box), color.RGBA{0, 0, 0, 240}, false)
	options := []string{"Resume", "Quicksave (Q)", "Load", "Settings", "Quit"}
	for i, opt := range options {
		clr, pfx := color.RGBA{255, 255, 255, 255}, "  "; if g.menuIndex == i { clr, pfx = color.RGBA{255, 255, 0, 255}, "> " }
		gr.graphics.DrawTextAt(screen, pfx+opt, mx+100, my+80+i*35, clr, 18)
	}
}

func (gr *GameRenderer) drawMinimap(screen engine.Image) {
	g, mW, mH := gr.game, 150.0, 150.0; mx, my := float32(g.width-int(mW)-10), float32(10)
	gr.graphics.DrawFilledRect(screen, mx-2, my-2, float32(mW+4), float32(mH+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, mx, my, float32(mW), float32(mH), color.RGBA{0, 0, 0, 180}, false)
	mapW, mapH := float64(g.currentMapType.MapWidth), float64(g.currentMapType.MapHeight); if mapW <= 0 { mapW = 100 }; if mapH <= 0 { mapH = 100 }
	
	clamp := func(v, max float64) float32 { if v < 0 { return 0 } else if v >= max { return float32(max - 0.1) } else { return float32(v) } }
	drawDot := func(x, y float64, clr color.Color, sz float32) {
		// Minimap coordinates: assume center of map is (0,0) and spans [-W/2, W/2]
		nX, nY := (x + mapW*0.5) / mapW, (y + mapH*0.5) / mapH
		gr.graphics.DrawFilledRect(screen, mx+clamp(nX*mW, mW)-sz/2, my+clamp(nY*mH, mH)-sz/2, sz, sz, clr, false)
	}

	for _, n := range g.characters {
		if !n.IsAlive() { continue }
		dotColor := color.RGBA{255, 255, 255, 255}
		if n.Alignment == AlignmentEnemy { dotColor = color.RGBA{255, 0, 0, 255} } else if n.Alignment == AlignmentAlly { dotColor = color.RGBA{0, 255, 0, 255} }
		drawDot(n.X, n.Y, dotColor, 2)
	}
	drawDot(g.playableCharacter.X, g.playableCharacter.Y, color.RGBA{0, 255, 255, 255}, 3)
}
