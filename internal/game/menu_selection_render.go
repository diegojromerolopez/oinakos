package game

import (
	"fmt"
	"image/color"
	"oinakos/internal/engine"
	"strings"
)

func (gr *GameRenderer) drawCharacterSelect(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)
	title := "OINAKOS: CHOOSE YOUR HERO"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 50, color.RGBA{218, 165, 32, 255}, 32)

	playableIDs := g.characterRegistry.PlayableIDs()
	for i, id := range playableIDs {
		char := g.characterRegistry.Characters[id]
		var clr color.Color = color.White; prefix := "  "
		if g.characterMenuIndex == i { clr, prefix = color.RGBA{255, 255, 0, 255}, "> "; gr.drawHeroPreview(screen, char, g.width/2+50, 130) }
		gr.graphics.DrawTextAt(screen, prefix+char.Name, 100, 130+i*35, clr, 18)
	}
	nPlayable := len(playableIDs)
	var clrBack color.Color = color.White; prefixBack := "  "
	if g.characterMenuIndex == nPlayable { clrBack, prefixBack = color.RGBA{255, 255, 0, 255}, "> " }
	gr.graphics.DrawTextAt(screen, prefixBack+"BACK", 100, 130+nPlayable*35+20, clrBack, 18)
	gr.graphics.DrawTextAt(screen, "Press UP/DOWN to navigate, ENTER to select hero.", g.width/2-200, g.height-50, color.RGBA{136, 136, 136, 255}, 14)
}

func (gr *GameRenderer) drawHeroPreview(screen engine.Image, char *EntityConfig, x, y int) {
	gr.graphics.DrawTextAt(screen, "--- HERO PROFILE ---", x, y, color.RGBA{218, 165, 32, 255}, 20)
	if char.StaticImage != nil {
		op := engine.NewDrawImageOptions(); op.Scale(1.5, 1.5); op.Translate(float64(x), float64(y+30)); screen.DrawImage(char.StaticImage, op)
	}
	stX, stY := x+180, y+40
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Strength:  %s", char.Attributes.Strength.String()), stX, stY, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Health:    %s", char.Attributes.Health.String()), stX, stY+25, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Dexterity: %s", char.Attributes.Dexterity.String()), stX, stY+50, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Intellect: %s", char.Attributes.Intellect.String()), stX, stY+75, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Wisdom:    %s", char.Attributes.Wisdom.String()), stX, stY+100, color.White, 16)
	gr.graphics.DrawTextAt(screen, "--- BIOGRAPHY ---", x, y+230, color.RGBA{218, 165, 32, 255}, 20)
	words, line, lineNum := strings.Fields(char.Description), "", 0
	for _, w := range words {
		if len(line)+len(w) > 40 { gr.graphics.DrawTextAt(screen, line, x, y+260+lineNum*20, color.White, 14); line, lineNum = w+" ", lineNum+1
		} else { line += w + " " }
	}
	gr.graphics.DrawTextAt(screen, line, x, y+260+lineNum*20, color.White, 14)
}

func (gr *GameRenderer) drawCampaignSelect(screen engine.Image) {
	g := gr.game; gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)
	title := "OINAKOS: SELECT YOUR JOURNEY"; tw, _ := gr.graphics.MeasureText(title, 32); gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 50, color.RGBA{218, 165, 32, 255}, 32)
	col1X, col2X, y := 100, g.width/2, 130
	gr.graphics.DrawTextAt(screen, "--- CAMPAIGNS ---", col1X-20, 100, color.RGBA{150, 150, 150, 255}, 18)
	gr.graphics.DrawTextAt(screen, "--- MAPS ---", col2X-20, 100, color.RGBA{150, 150, 150, 255}, 18)
	nC, nM := len(g.campaignRegistry.IDs), len(g.mapTypeRegistry.IDs)
	for i, id := range g.campaignRegistry.IDs {
		camp := g.campaignRegistry.Campaigns[id]; var clr color.Color = color.White; prefix := "  "
		if g.campaignMenuIndex == i { clr, prefix = color.RGBA{255, 255, 0, 255}, "> " }
		gr.graphics.DrawTextAt(screen, prefix+camp.Name, col1X, y+i*30, clr, 16)
	}
	for i, id := range g.mapTypeRegistry.IDs {
		m := g.mapTypeRegistry.Types[id]; var clr color.Color = color.White; prefix := "  "
		if g.campaignMenuIndex == nC+i { clr, prefix = color.RGBA{150, 255, 150, 255}, "> " }
		cX, rX := col2X, i; if i > 15 { cX, rX = col2X+250, i-16 }
		gr.graphics.DrawTextAt(screen, prefix+m.Name, cX, y+rX*30, clr, 16)
	}
	var clr color.Color = color.White; prefix := "  "; if g.campaignMenuIndex == nC+nM { clr, prefix = color.RGBA{255, 0, 0, 255}, "> " }
	gr.graphics.DrawTextAt(screen, prefix+"QUIT", (g.width-80)/2, g.height-90, clr, 24)
}
