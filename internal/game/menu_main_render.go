package game

import (
	"fmt"
	"image/color"
	"oinakos/internal/engine"
	"strings"
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
