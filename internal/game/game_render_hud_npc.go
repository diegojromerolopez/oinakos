package game

import (
	"fmt"
	"image/color"
	"oinakos/internal/engine"
	"strings"
)

func (gr *GameRenderer) drawHoverInfo(screen engine.Image) {
	g := gr.game
	mx, my := g.input.MousePosition()
	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)
	if g.pinnedCharacter != nil && g.pinnedCharacter.IsAlive() { gr.drawNPCStatusBox(screen, g.pinnedCharacter, g.pinnedUIX, g.pinnedUIY); return }
	if g.World == nil { return }
	mouseX, mouseY := engine.IsoToCartesian(float64(mx)-offsetX, float64(my)-offsetY)
	for _, it := range g.World.Items {
		if it == nil || it.Config == nil { continue }
		if it.GetFootprint().Contains(mouseX, mouseY) {
			isoX, isoY := engine.CartesianToIso(it.X, it.Y)
			gr.graphics.DrawTextAt(screen, it.Config.Name, int(isoX+offsetX), int(isoY+offsetY+15), color.White, 12)
			return
		}
	}
}

func (gr *GameRenderer) drawNPCStatusBox(screen engine.Image, n *Character, x, y int) {
	boxW, boxH, bx, by := 320.0, 480.0, float32(x), float32(y)
	gold, black, gray, white := color.RGBA{218, 165, 32, 255}, color.RGBA{0, 0, 0, 240}, color.RGBA{136, 136, 136, 255}, color.White
	mx, my := gr.game.input.MousePosition()
	gr.graphics.DrawFilledRect(screen, bx-2, by-2, float32(boxW)+4, float32(boxH)+4, gold, false)
	gr.graphics.DrawFilledRect(screen, bx, by, float32(boxW), float32(boxH), black, false)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("%s (%s)", n.Name, n.Alignment), int(bx)+10, int(by)+20, gold, 16)
	if n.LastReaction != "" { gr.graphics.DrawTextAt(screen, "REACTION: "+n.LastReaction, int(bx)+10, int(by)+38, color.RGBA{0, 255, 255, 255}, 11) }
	yMed, sent := int(by)+55, 0.0
	if n.Relationships != nil { sent = n.Relationships[gr.game.playableCharacter.Name] }
	sentColor := color.Color(white); if sent < -10 { sentColor = color.RGBA{255, 50, 50, 255} } else if sent > 10 { sentColor = color.RGBA{50, 255, 50, 255} }
	tier := n.GetRelationshipTier(gr.game.playableCharacter.Name)
	if tier == "Romantic" || tier == "Devoted" { sentColor = color.RGBA{255, 105, 180, 255} }
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Rel: %s (%.1f)", tier, sent), int(bx)+10, yMed+10, sentColor, 12)
	yMed += 35
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("STR:%d DEX:%d HEA:%d", n.PrimaryAttributes.Strength, n.PrimaryAttributes.Dexterity, n.PrimaryAttributes.Health), int(bx)+10, yMed, color.RGBA{200, 200, 255, 255}, 11)
	yMed += 14; gr.graphics.DrawTextAt(screen, fmt.Sprintf("INT:%d WIS:%d AGE:%.1f", n.PrimaryAttributes.Intellect, n.PrimaryAttributes.Wisdom, n.State.Age.Current), int(bx)+10, yMed, color.RGBA{200, 200, 255, 255}, 11)
	yBio := yMed + 25; gr.graphics.DrawTextAt(screen, "-- BIOLOGICAL NEEDS --", int(bx)+10, yBio, gray, 11)
	yBio += 15
	
	isVampire := n.State.Age.Rate == 0
	if isVampire {
		gr.graphics.DrawTextAt(screen, fmt.Sprintf("BLOODLUST:%d%% FATIGUE:%d%%", int(n.State.Hunger), int(n.State.Fatigue)), int(bx)+20, yBio, color.RGBA{255, 150, 150, 255}, 10)
	} else {
		gr.graphics.DrawTextAt(screen, fmt.Sprintf("HUNGER:%d%% THIRST:%d%% FATIGUE:%d%%", int(n.State.Hunger), int(n.State.Thirst), int(n.State.Fatigue)), int(bx)+20, yBio, white, 10)
	}
	
	yBio += 12; gr.graphics.DrawTextAt(screen, fmt.Sprintf("BLAD:%d%% BOWL:%d%% SAN:%d%% PAIN:%d%%", int(n.State.BladderLevel), int(n.State.BowelLevel), int(n.State.Sanity), int(n.State.Pain)), int(bx)+20, yBio, white, 10)
	yMem := yBio + 25; gr.graphics.DrawTextAt(screen, "-- RECENT MEMORIES --", int(bx)+10, yMem, gray, 11); yMem += 15
	for i, count := len(n.Memories)-1, 0; i >= 0 && count < 6; i, count = i-1, count+1 {
		m := n.Memories[i]; desc := fmt.Sprintf("%s by %s", strings.ToUpper(m.Type), m.Source)
		gr.graphics.DrawTextAt(screen, fmt.Sprintf("- %s (%.1f)", desc, m.Value), int(bx)+20, yMem, color.RGBA{200, 200, 200, 255}, 10); yMem += 12
	}
	yCmds := int(by) + int(boxH) - 120; gr.graphics.DrawLine(screen, bx+5, float32(yCmds-5), bx+float32(boxW)-5, float32(yCmds-5), gray, 1)
	gr.graphics.DrawTextAt(screen, "-- COMMANDS --", int(bx)+10, yCmds, gray, 11); yCmds += 20
	commands := []string{"TALK", "ATTACK", "TRADE", "SEDUCE", "INTIMIDATE", "STEAL", "RESTRAIN", "HEAL", "GIVE"}
	if gr.game.settings.AdultMode {
		commands = append(commands, "SEX", "TORTURE")
	}
	for i, cmd := range commands {
		cx, cy, clr := int(bx)+10+(i%3)*100, yCmds+(i/3)*25, color.RGBA{255, 255, 255, 255}
		if mx >= cx && mx <= cx+85 && my >= cy-12 && my <= cy+8 { clr = color.RGBA{255, 255, 0, 255} }
		gr.graphics.DrawTextAt(screen, cmd, cx, cy, clr, 11)
	}
}
