package game

import (
	"fmt"
	"image/color"
	"strings"
	"oinakos/internal/engine"
)

func (gr *GameRenderer) drawHUD(screen engine.Image) {
	g := gr.game
	// Top-Left World Time & Climate Panel
	sState := g.World.State
	gr.graphics.DrawFilledRect(screen, 10, 10, 220, 50, color.RGBA{0, 0, 0, 180}, false)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("%s, HOUR %02d:%02d", sState.Season, sState.Hour, sState.Ticks/12), 20, 25, color.RGBA{218, 165, 32, 255}, 14)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("TEMP: %.1fC  WEATHER: %s", sState.Temperature, sState.Weather), 20, 45, color.White, 11)

	// Primary Status Panel (shifted down)
	hudY := 70
	gr.graphics.DrawFilledRect(screen, 10, float32(hudY), 350, 220, color.RGBA{0, 0, 0, 180}, false)

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("HP: %d/%d", g.playableCharacter.TemporalState.HealthPoints, g.playableCharacter.TemporalState.MaxHealthPoints), 20, hudY+10, color.White, 16)
	gr.graphics.DrawFilledRect(screen, 160, float32(hudY+12), 200, 10, color.RGBA{100, 0, 0, 255}, false)

	healthPct := float64(g.playableCharacter.TemporalState.HealthPoints) / float64(g.playableCharacter.TemporalState.MaxHealthPoints)
	if healthPct > 0 {
		var healthColor color.RGBA
		if healthPct > 0.7 {
			healthColor = color.RGBA{0, 200, 0, 255}
		} else if healthPct > 0.5 {
			healthColor = color.RGBA{200, 200, 0, 255}
		} else {
			healthColor = color.RGBA{200, 0, 0, 255}
		}
		gr.graphics.DrawFilledRect(screen, 160, float32(hudY+12), float32(200*healthPct), 10, healthColor, false)
	}

	// --- Hunger, Thirst, Fatigue, Hygiene, Needs Bars ---
	barX, barW := 160, 200
	
	// Hunger (0=satiated, 100=famished)
	hungerColor := color.RGBA{210, 105, 30, 255} 
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("HUNGER: %d%%", int(g.playableCharacter.TemporalState.Hunger)), 20, hudY+35, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+37), float32(barW), 6, color.RGBA{50, 25, 0, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+37), float32(float64(barW)*(g.playableCharacter.TemporalState.Hunger/100.0)), 6, hungerColor, false)

	// Thirst (0=hydrated, 100=thirsty)
	thirstColor := color.RGBA{0, 191, 255, 255} 
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("THIRST: %d%%", int(g.playableCharacter.TemporalState.Thirst)), 20, hudY+48, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+50), float32(barW), 6, color.RGBA{0, 20, 50, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+50), float32(float64(barW)*(g.playableCharacter.TemporalState.Thirst/100.0)), 6, thirstColor, false)

	// Fatigue (0=rested, 100=exhausted)
	fatigueColor := color.RGBA{255, 215, 0, 255} 
	fatigueLabel := "FATIGUE"
	if g.playableCharacter.State == ActorResting { fatigueLabel = "RESTING" }
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("%s: %d%%", fatigueLabel, int(g.playableCharacter.TemporalState.Fatigue)), 20, hudY+61, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+63), float32(barW), 6, color.RGBA{50, 45, 0, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+63), float32(float64(barW)*(g.playableCharacter.TemporalState.Fatigue/100.0)), 6, fatigueColor, false)

	// Hygiene
	hygieneColor := color.RGBA{144, 238, 144, 255} // Light Green
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("HYGIENE: %d%%", int(g.playableCharacter.TemporalState.Hygiene)), 20, hudY+74, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+76), float32(barW), 6, color.RGBA{0, 50, 0, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+76), float32(float64(barW)*(g.playableCharacter.TemporalState.Hygiene/100.0)), 6, hygieneColor, false)

	// Miccionate
	miccColor := color.RGBA{255, 255, 0, 255} // Yellow
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("MICCIONATE: %d%%", int(g.playableCharacter.TemporalState.Miccionate)), 20, hudY+87, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+89), float32(barW), 6, color.RGBA{50, 50, 0, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+89), float32(float64(barW)*(g.playableCharacter.TemporalState.Miccionate/100.0)), 6, miccColor, false)

	// Defecate
	defColor := color.RGBA{139, 69, 19, 255} // Brown
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("DEFECATE: %d%%", int(g.playableCharacter.TemporalState.Defecate)), 20, hudY+100, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+102), float32(barW), 6, color.RGBA{40, 20, 0, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+102), float32(float64(barW)*(g.playableCharacter.TemporalState.Defecate/100.0)), 6, defColor, false)

	// Sanity
	sanityColor := color.RGBA{0, 255, 255, 255} // Cyan
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("SANITY: %d%%", int(g.playableCharacter.TemporalState.Sanity)), 20, hudY+113, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+115), float32(barW), 6, color.RGBA{0, 50, 50, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+115), float32(float64(barW)*(g.playableCharacter.TemporalState.Sanity/100.0)), 6, sanityColor, false)

	// Arousal
	arousalColor := color.RGBA{255, 105, 180, 255} // Pink
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("AROUSAL: %d%%", int(g.playableCharacter.TemporalState.Arousal)), 20, hudY+126, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+128), float32(barW), 6, color.RGBA{50, 0, 30, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+128), float32(float64(barW)*(g.playableCharacter.TemporalState.Arousal/100.0)), 6, arousalColor, false)

	// Pain
	painColor := color.RGBA{255, 0, 0, 255} // Red
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("PAIN: %d%%", int(g.playableCharacter.TemporalState.Pain)), 20, hudY+139, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+141), float32(barW), 6, color.RGBA{50, 0, 0, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+141), float32(float64(barW)*(g.playableCharacter.TemporalState.Pain/100.0)), 6, painColor, false)

	// Alcohol
	alcColor := color.RGBA{150, 50, 250, 255} // Purple
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("ALCOHOL: %d%%", int(g.playableCharacter.TemporalState.AlcoholLevel)), 20, hudY+152, color.White, 12)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+154), float32(barW), 6, color.RGBA{30, 0, 50, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(barX), float32(hudY+154), float32(float64(barW)*(g.playableCharacter.TemporalState.AlcoholLevel/100.0)), 6, alcColor, false)

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("LVL: %d  XP: %d  AGE: %.1f  GOLD: %d", g.playableCharacter.Level, g.playableCharacter.XP, g.playableCharacter.TemporalState.Age.Current, g.playableCharacter.Denarii), 20, hudY+168, color.White, 12)
	
	objText := fmt.Sprintf("OBJ: %s", g.currentMapType.Description)
	nextY := gr.drawWrappedText(screen, objText, 20, hudY+183, 310, color.White, 12, hudY+280)

	// Removed legacy time logic
	
	yPos := nextY
	if yPos < hudY+80 { yPos = hudY+80 }

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("POS %s, %s  KILLS: %d", g.settings.FormatDistance(g.playableCharacter.X), g.settings.FormatDistance(g.playableCharacter.Y), g.playableCharacter.Kills), 20, yPos, color.White, 12)
	yPos += 15
	
	// Primary Stats Grid
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("STR: %d  DEX: %d  HEA: %d", g.playableCharacter.PrimaryAttributes.Strength, g.playableCharacter.PrimaryAttributes.Dexterity, g.playableCharacter.PrimaryAttributes.Health), 20, yPos, color.RGBA{200, 200, 255, 255}, 12)
	yPos += 15
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("INT: %d  WIS: %d", g.playableCharacter.PrimaryAttributes.Intellect, g.playableCharacter.PrimaryAttributes.Wisdom), 20, yPos, color.RGBA{200, 200, 255, 255}, 12)
	yPos += 15

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("ATK: %d  DEF: %d  SHIELD: %d", g.playableCharacter.GetTotalAttack(), g.playableCharacter.GetTotalDefense(), g.playableCharacter.GetTotalProtection()), 20, yPos, color.White, 12)
	yPos += 15
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("WEIGHT: %s / %s", g.settings.FormatWeight(g.playableCharacter.GetTotalWeight()), g.settings.FormatWeight(g.playableCharacter.MaxWeight)), 20, yPos, color.White, 12)

	weaponText := "WEAPON: Unarmed (1-2)"
	if g.playableCharacter.Weapon != nil {
		weaponText = fmt.Sprintf("WEAPON: %s (%d-%d)", g.playableCharacter.Weapon.Name, g.playableCharacter.Weapon.Damage.Min, g.playableCharacter.Weapon.Damage.Max)
		if g.playableCharacter.Weapon.Bonus > 0 {
			weaponText += fmt.Sprintf(" +%d", g.playableCharacter.Weapon.Bonus)
		}
	}
	gr.graphics.DrawTextAt(screen, weaponText, 20, yPos+15, color.White, 12)
	
	// Complex State Indicators
	states := []string{}
	if g.playableCharacter.FluTicks > 0 { states = append(states, "INFECTED (FLU)") }
	if g.playableCharacter.TemporalState.IsSeptic { states = append(states, "SEPTIC") }
	if g.playableCharacter.TemporalState.IsSick { states = append(states, "SICK") }
	if g.playableCharacter.TemporalState.IsPoisoned { states = append(states, "POISONED") }
	if g.playableCharacter.TemporalState.IsDrunk { states = append(states, "DRUNK") }
	if g.playableCharacter.IsPregnant { states = append(states, "PREGNANT") }

	if len(states) > 0 {
		gr.graphics.DrawTextAt(screen, "STATE: "+strings.Join(states, ", "), 20, yPos+35, color.RGBA{255, 100, 100, 255}, 12)
	}

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
	gr.drawMinimap(screen)
}

func (gr *GameRenderer) drawMinimap(screen engine.Image) {
	g := gr.game
	miniW, miniH := 150.0, 150.0
	mx, my := float32(g.width-int(miniW)-10), float32(10)

	// Border & Background
	gr.graphics.DrawFilledRect(screen, mx-2, my-2, float32(miniW+4), float32(miniH+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, mx, my, float32(miniW), float32(miniH), color.RGBA{0, 0, 0, 180}, false)

	mapW, mapH := float64(g.currentMapType.MapWidth), float64(g.currentMapType.MapHeight)
	if mapW == 0 { mapW = 100 }
	if mapH == 0 { mapH = 100 }

	// Draw obstacles as gray dots
	for _, o := range g.obstacles {
		if !o.Alive { continue }
		px := mx + float32((o.X / mapW) * miniW)
		py := my + float32((o.Y / mapH) * miniH)
		gr.graphics.DrawFilledRect(screen, px, py, 1, 1, color.RGBA{100, 100, 100, 255}, false)
	}

	// Draw characters
	for _, n := range g.characters {
		if !n.IsAlive() { continue }
		px := mx + float32((n.X / mapW) * miniW)
		py := my + float32((n.Y / mapH) * miniH)
		dotColor := color.RGBA{255, 255, 255, 255}
		if n.Alignment == AlignmentEnemy { dotColor = color.RGBA{255, 0, 0, 255} }
		if n.Alignment == AlignmentAlly { dotColor = color.RGBA{0, 255, 0, 255} }
		gr.graphics.DrawFilledRect(screen, px-1, py-1, 2, 2, dotColor, false)
	}

	pxPlayer := mx + float32((g.playableCharacter.X / mapW) * miniW)
	pyPlayer := my + float32((g.playableCharacter.Y / mapH) * miniH)
	gr.graphics.DrawFilledRect(screen, pxPlayer-1, pyPlayer-1, 3, 3, color.Color(color.RGBA{0, 255, 255, 255}), false)
}

func (gr *GameRenderer) drawHoverInfo(screen engine.Image) {
	g := gr.game
	mx, my := g.input.MousePosition()
	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)

	pinnedDrawn := false
	if g.pinnedCharacter != nil && g.pinnedCharacter.IsAlive() {
		isoX, isoY := engine.CartesianToIso(g.pinnedCharacter.X, g.pinnedCharacter.Y)
		scrX, scrY := isoX+offsetX, isoY+offsetY
		gr.drawNPCStatusBox(screen, g.pinnedCharacter, int(scrX), int(scrY))
		pinnedDrawn = true
	}

	// Hover info for characters is now disabled, only click-to-pin is allowed.

	// Only check for items if we aren't hovering a character
	if g.World != nil && !pinnedDrawn {
		mouseX, mouseY := engine.IsoToCartesian(float64(mx)-offsetX, float64(my)-offsetY)
		for _, it := range g.World.Items {
			if it == nil || it.Config == nil {
				continue
			}
			poly := it.GetFootprint()
			if poly.Contains(mouseX, mouseY) {
				isoX, isoY := engine.CartesianToIso(it.X, it.Y)
				gr.graphics.DrawTextAt(screen, it.Config.Name, int(isoX+offsetX), int(isoY+offsetY+15), color.White, 12)
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

	gold := color.RGBA{218, 165, 32, 255}
	black := color.RGBA{0, 0, 0, 240}
	white := color.White

	gr.graphics.DrawFilledRect(screen, bx-2, by-2, float32(boxW)+4, float32(boxH)+4, gold, false)
	gr.graphics.DrawFilledRect(screen, bx, by, float32(boxW), float32(boxH), black, false)
	gr.graphics.DebugPrintAt(screen, title, int(bx)+10, int(by)+10, gold)

	words := strings.Fields(desc)
	line := ""
	lineNum := 0
	for _, w := range words {
		if len(line)+len(w) > 35 {
			gr.graphics.DebugPrintAt(screen, line, int(bx)+10, int(by)+35+lineNum*15, white)
			line = w + " "
			lineNum++
		} else {
			line += w + " "
		}
	}
	gr.graphics.DebugPrintAt(screen, line, int(bx)+10, int(by)+35+lineNum*15, white)
}

func (gr *GameRenderer) drawNPCStatusBox(screen engine.Image, n *Character, x, y int) {
	boxW, boxH := 320.0, 480.0
	bx, by := float32(x+20), float32(y+20)
	if float64(bx)+boxW > float64(gr.game.width) { bx = float32(float64(x) - boxW - 20) }
	if float64(by)+boxH > float64(gr.game.height) { by = float32(float64(y) - boxH - 20) }

	gold := color.RGBA{218, 165, 32, 255}
	black := color.RGBA{0, 0, 0, 240}
	gray := color.RGBA{136, 136, 136, 255}
	white := color.White
	mx, my := gr.game.input.MousePosition()

	gr.graphics.DrawFilledRect(screen, bx-2, by-2, float32(boxW)+4, float32(boxH)+4, gold, false)
	gr.graphics.DrawFilledRect(screen, bx, by, float32(boxW), float32(boxH), black, false)
	
	title := fmt.Sprintf("%s (%s)", n.Name, n.Alignment)
	gr.graphics.DrawTextAt(screen, title, int(bx)+10, int(by)+20, gold, 16)
	
	if n.LastReaction != "" {
		gr.graphics.DrawTextAt(screen, "REACTION: " + n.LastReaction, int(bx)+10, int(by)+38, color.RGBA{0, 255, 255, 255}, 11)
	}
	
	// Medical
	gr.graphics.DrawTextAt(screen, "-- MEDICAL STATUS --", int(bx)+10, int(by)+55, gray, 11)
	yMed := int(by) + 70
	for _, limb := range []string{"head", "torso", "l_arm", "r_arm", "l_leg", "r_leg"} {
		hp := n.BodyStatus[limb]
		lbl := fmt.Sprintf("%s: %d", strings.ToUpper(limb), hp)
		var clr color.Color = color.White
		if hp < 5 { clr = color.RGBA{255, 0, 0, 255} } else if hp < 15 { clr = color.RGBA{255, 255, 0, 255} }
		gr.graphics.DrawTextAt(screen, lbl, int(bx)+20, yMed, clr, 11)
		yMed += 14
	}

	// Sentiment Toward Player
	sent := 0.0
	if n.Relationships != nil { sent = n.Relationships[gr.game.playableCharacter.Name] }
	var sentColor color.Color = color.White
	if sent < -10 { sentColor = color.RGBA{255, 50, 50, 255} } else if sent > 10 { sentColor = color.RGBA{50, 255, 50, 255} }
	tier := n.GetRelationshipTier(gr.game.playableCharacter.Name)
	if tier == "Romantic" || tier == "Devoted" { sentColor = color.RGBA{255, 105, 180, 255} }
	
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Relationship: %s (%.1f)", tier, sent), int(bx)+10, yMed+10, sentColor, 12)

	// Primary Stats
	yMed += 25
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("STR:%d DEX:%d HEA:%d", n.PrimaryAttributes.Strength, n.PrimaryAttributes.Dexterity, n.PrimaryAttributes.Health), int(bx)+10, yMed, color.RGBA{200, 200, 255, 255}, 11)
	yMed += 14
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("INT:%d WIS:%d  AGE:%.1f", n.PrimaryAttributes.Intellect, n.PrimaryAttributes.Wisdom, n.TemporalState.Age.Current), int(bx)+10, yMed, color.RGBA{200, 200, 255, 255}, 11)
	
	love := 0.0
	if n.RomanticInterest != nil { love = n.RomanticInterest[gr.game.playableCharacter.Name] }
	if love > 0 {
		yMed += 14
		gr.graphics.DrawTextAt(screen, fmt.Sprintf("Passion: %.1f ❤", love), int(bx)+10, yMed, color.RGBA{255, 105, 180, 255}, 12)
	}

	// Biological Needs
	yMed += 25
	gr.graphics.DrawTextAt(screen, "-- BIOLOGICAL NEEDS --", int(bx)+10, yMed, gray, 11)
	yBio := yMed + 15
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("HUNGER: %d%%  THIRST: %d%%", int(n.TemporalState.Hunger), int(n.TemporalState.Thirst)), int(bx)+20, yBio, white, 10)
	yBio += 12
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("FATIGUE: %d%% HYGIENE: %d%%", int(n.TemporalState.Fatigue), int(n.TemporalState.Hygiene)), int(bx)+20, yBio, white, 10)
	yBio += 12
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("MICC: %d%%    DEF: %d%%    SAN: %d%%", int(n.TemporalState.Miccionate), int(n.TemporalState.Defecate), int(n.TemporalState.Sanity)), int(bx)+20, yBio, white, 10)
	yBio += 12
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("AROUSAL: %d%% PAIN: %d%%    GOLD: %d", int(n.TemporalState.Arousal), int(n.TemporalState.Pain), n.Denarii), int(bx)+20, yBio, white, 10)
	yBio += 12
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("ALCOHOL: %d%%", int(n.TemporalState.AlcoholLevel)), int(bx)+20, yBio, white, 10)
	
	// State Indicators
	states := []string{}
	if n.FluTicks > 0 { states = append(states, "FLU") }
	if n.TemporalState.IsSeptic { states = append(states, "SEPTIC") }
	if n.TemporalState.IsSick { states = append(states, "SICK") }
	if n.TemporalState.IsPoisoned { states = append(states, "POISONED") }
	if n.TemporalState.IsDrunk { states = append(states, "DRUNK") }
	if n.IsPregnant { states = append(states, "PREGNANT") }
	if n.IsIncapacitated() { states = append(states, "INCAPACITATED") }
	
	if len(states) > 0 {
		yBio += 14
		gr.graphics.DrawTextAt(screen, "STATE: "+strings.Join(states, ", "), int(bx)+10, yBio, color.RGBA{255, 100, 100, 255}, 10)
	}

	// Memory
	yMemHeader := yBio + 25
	gr.graphics.DrawTextAt(screen, "-- RECENT MEMORIES --", int(bx)+10, yMemHeader, color.RGBA{136, 136, 136, 255}, 11)
	yMem := yMemHeader + 15
	count := 0
	for i := len(n.Memories)-1; i >= 0 && count < 3; i-- {
		m := n.Memories[i]
		gr.graphics.DrawTextAt(screen, fmt.Sprintf("- %s: %s (%.1f)", m.Source, m.Type, m.Value), int(bx)+20, yMem, color.RGBA{200, 200, 200, 255}, 10)
		yMem += 13
		count++
	}

	if n.Config != nil && n.Config.Description != "" {
		gr.graphics.DrawTextAt(screen, "-- NOTES --", int(bx)+10, yMem+10, color.RGBA{136, 136, 136, 255}, 11)
		yMem = gr.drawWrappedText(screen, n.Config.Description, int(bx)+20, yMem+22, int(boxW)-40, color.RGBA{180, 180, 180, 255}, 10, int(by)+int(boxH)-120)
	}

	// --- COMMANDS SECTION ---
	yCmds := int(by) + int(boxH) - 120
	gr.graphics.DrawLine(screen, bx+5, float32(yCmds-5), bx+float32(boxW)-5, float32(yCmds-5), gray, 1)
	gr.graphics.DrawTextAt(screen, "-- COMMANDS --", int(bx)+10, yCmds, gray, 11)
	yCmds += 20
	
	commands := []string{"TALK", "ATTACK", "TRADE", "SEDUCE", "INTIMIDATE", "STEAL", "RESTRAIN", "HEAL", "GIVE ITEM", "SEX", "TORTURE"}
	for i, cmd := range commands {
		cx, cy := int(bx)+10 + (i%3)*100, yCmds + (i/3)*25
		var clr color.Color = color.RGBA{255, 255, 255, 255}
		if mx >= cx && mx <= cx+85 && my >= cy-12 && my <= cy+8 {
			clr = color.RGBA{255, 255, 0, 255}
		}
		gr.graphics.DrawTextAt(screen, cmd, cx, cy, clr, 11)
	}
}
func (gr *GameRenderer) drawDialogueBox(screen engine.Image) {
	g := gr.game
	// Position: Bottom of the window
	boxH := 180
	isDialogue := g.ActiveDialogue != nil
	if isDialogue {
		if g.ActiveDialogue.UIState == DialogueMaximized {
			boxH = 650
		} else {
			boxH = 350
		}
	} else {
		if g.LogUIState == DialogueMaximized {
			boxH = 300 // Expanded Log
		} else {
			boxH = 85 // Slim Log
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

	// Draw Event Log Header
	gr.graphics.DrawTextAt(screen, "-- EVENT LOG (History) --", bx+10, by+12, color.RGBA{150, 150, 150, 255}, 10)
	
	logY := by + 30
	var maxLogEntries int
	if isDialogue {
		if g.ActiveDialogue.UIState == DialogueMaximized {
			maxLogEntries = 12
		} else {
			maxLogEntries = 5
		}
	} else {
		if g.LogUIState == DialogueMaximized {
			maxLogEntries = 15
		} else {
			maxLogEntries = 3
		}
	}

	startIdx := len(g.EventLog) - maxLogEntries - g.LogScrollOffset
	if startIdx < 0 { startIdx = 0 }
	endIdx := startIdx + maxLogEntries
	if endIdx > len(g.EventLog) { endIdx = len(g.EventLog) }

	for i := startIdx; i < endIdx; i++ {
		entry := g.EventLog[i]
		var clr color.Color = color.White
		switch entry.Category {
		case LogPlayer: clr = color.RGBA{0, 255, 255, 255}
		case LogNPC: clr = color.RGBA{255, 255, 255, 255}
		case LogCombatDamage: clr = color.RGBA{220, 20, 60, 255}
		case LogCombatRecovery: clr = color.RGBA{0, 255, 0, 255}
		}
		gr.graphics.DrawTextAt(screen, entry.Text, bx+10, logY, clr, 12)
		logY += 15
	}

	// Draw Scrollbar (if there's enough history)
	if len(g.EventLog) > maxLogEntries {
		sbW := float32(4)
		logAreaH := float32(maxLogEntries * 15 + 10)
		sbX := float32(bx + boxW - 10)
		sbTrackY := float32(by + 25)
		sbTrackH := logAreaH
		
		// Track (darker)
		gr.graphics.DrawFilledRect(screen, sbX, sbTrackY, sbW, sbTrackH, color.RGBA{50, 50, 50, 150}, false)
		
		// Handle
		visibleRatio := float32(maxLogEntries) / float32(len(g.EventLog))
		handleH := sbTrackH * visibleRatio
		if handleH < 10 { handleH = 10 }
		
		maxOffset := float32(len(g.EventLog) - maxLogEntries)
		scrollRatio := float32(0)
		if maxOffset > 0 { scrollRatio = float32(g.LogScrollOffset) / maxOffset }
		handleY := sbTrackY + (sbTrackH - handleH) * (1.0 - scrollRatio)
		
		gr.graphics.DrawFilledRect(screen, sbX, handleY, sbW, handleH, color.RGBA{218, 165, 32, 200}, false)
	}

	diagYStart := float32(logY) + 10
	if !isDialogue { diagYStart = float32(by) + float32(boxH) - 25 }
	
	// Draw line separator if in dialogue
	if isDialogue {
		gr.graphics.DrawLine(screen, float32(bx+5), float32(diagYStart-5), float32(bx+boxW-5), float32(diagYStart-5), color.RGBA{136, 136, 136, 255}, 1)
	}

	// Draw Active Dialogue
	if isDialogue {
		diagY := diagYStart + 10
		speakerName := g.ActiveDialogue.SpeakerNPC.Name
		gr.graphics.DrawTextAt(screen, speakerName+":", bx+10, int(diagY), color.RGBA{218, 165, 32, 255}, 16)
		diagY += 25
		
		// Calculate choices area start to avoid overlap
		choiceYStart := float32(by) + float32(boxH) - 80 - float32(len(g.ActiveDialogue.Choices))*22
		
		gr.drawWrappedText(screen, g.ActiveDialogue.CurrentText, bx+20, int(diagY), boxW-40, color.White, 14, int(choiceYStart)-10)

		// Draw Choices
		choiceY := choiceYStart
		for i, choice := range g.ActiveDialogue.Choices {
			clr := color.RGBA{200, 200, 200, 255}
			prefix := "  "
			if i == g.ActiveDialogue.SelectedChoice {
				clr = color.RGBA{255, 255, 0, 255}
				prefix = "> "
			}
			gr.graphics.DrawTextAt(screen, prefix+choice.Text, bx+30, int(choiceY), clr, 14)
			choiceY += 22
		}
		
		gr.graphics.DrawTextAt(screen, "[ESC/BACKSPACE] Close", bx+boxW-195, by+boxH-60, color.RGBA{136, 136, 136, 255}, 11)
		
		if g.debug {
			debugInfo := fmt.Sprintf("AI DECISION: %s (%s)", g.ActiveDialogue.SpeakerNPC.LastAIChoice, g.ActiveDialogue.SpeakerNPC.LastAIReasoning)
			gr.graphics.DrawTextAt(screen, debugInfo, bx+10, by+boxH-25, color.RGBA{0, 255, 0, 255}, 10)
		}
		
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

func (gr *GameRenderer) drawWrappedText(screen engine.Image, text string, x, y, maxWidth int, clr color.Color, size int, maxY int) int {
	words := strings.Fields(text)
	line := ""
	currY := y
	for _, w := range words {
		wWidth, _ := gr.graphics.MeasureText(line+w+" ", float64(size))
		if int(wWidth) > maxWidth {
			if currY <= maxY {
				gr.graphics.DrawTextAt(screen, line, x, currY, clr, float64(size))
			}
			line = w + " "
			currY += size + 6
		} else {
			line += w + " "
		}
	}
	if currY <= maxY {
		gr.graphics.DrawTextAt(screen, line, x, currY, clr, float64(size))
	}
	return currY + size + 6
}
