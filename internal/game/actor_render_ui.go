package game

import (
	"image/color"
	"oinakos/internal/engine"
)

func DrawActorUI(g *Game, a *Actor, screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64, isPlayableCharacter bool, debug bool) {
	if screen == nil || a.Config == nil || !a.IsAlive() { return }
	isoX, isoY := engine.CartesianToIsoZ(a.X, a.Y, a.Z)
	if a.IsOccluded { DrawAlignmentIndicator(screen, vectorRenderer, a, offsetX, offsetY, true)
		if sb := g.GetSilhouetteBuffer(); sb != nil {
			sb.Clear(); sOp := *DrawActorGetOptions(a, offsetX, offsetY, isPlayableCharacter); sOp.SetColorScale(0, 0, 0, 1)
			if sprite := DrawActorGetSprite(a); sprite != nil {
				sb.DrawImage(sprite, &sOp); aSortY := a.GetSortY()
				for _, o := range g.obstacles { if o.GetSortY() > aSortY { oOp := engine.NewDrawImageOptions(); ox, oy := o.GetIsoPos(); oOp.Translate(ox+offsetX, oy+offsetY); oOp.Blend = engine.BlendDestinationIn; if o.Archetype != nil && o.Archetype.Image != nil { sb.DrawImage(o.Archetype.Image, oOp) } } }
				screen.DrawImage(sb, engine.NewDrawImageOptions())
			}
		}
	}
	barW, barH, bx := 40.0, 3.0, isoX+offsetX-20.0
	h := 160.0; if a.Config.StaticImage != nil { _, ih := a.Config.StaticImage.Size(); h = float64(ih) }
	mult := 0.9; if a.Config.IsAnimal { mult = 0.35 }; by := isoY + offsetY - h*mult
	if vectorRenderer != nil {
		vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barW), float32(barH), color.RGBA{80, 0, 0, 255}, false)
		if a.State.MaxHealthPoints > 0 {
			hpF := float32(a.State.HealthPoints) / float32(a.State.MaxHealthPoints); if hpF > 1 { hpF = 1 }
			if hpF > 0 { vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barW)*hpF, float32(barH), color.RGBA{0, 255, 0, 255}, false) }
		}
		ebY := by + barH + 1
		needs := []struct { val float64; clr color.RGBA }{{a.State.Hunger, color.RGBA{210, 105, 30, 255}}, {a.State.Thirst, color.RGBA{0, 191, 255, 255}}, {a.State.Fatigue, color.RGBA{255, 215, 0, 255}}}
		for _, n := range needs { vectorRenderer.DrawFilledRect(screen, float32(bx), float32(ebY), float32(barW), 1, color.RGBA{30, 30, 30, 255}, false); vectorRenderer.DrawFilledRect(screen, float32(bx), float32(ebY), float32(barW)*float32(n.val/100.0), 1, n.clr, false); ebY += 2 }
	}
	if textRenderer != nil {
		name := a.Name; if name == "" { name = a.Config.Name }; if isPlayableCharacter && name == "" { name = "Player" }
		if name != "" {
			nX, nY, clr := int(isoX+offsetX-float64(len(name))*3.5), int(isoY+offsetY+5), color.Color(color.White)
			if !isPlayableCharacter && (a.IsTarget || a.MustSurvive) { clr = color.RGBA{218, 165, 32, 255} }
			textRenderer.DrawTextAt(screen, name, nX, nY, clr, 12)
			if !isPlayableCharacter && g.playableCharacter != nil {
				if tier := g.playableCharacter.GetRelationshipTier(a.Name); tier != "" && tier != "Neutral" { textRenderer.DrawTextAt(screen, tier, int(isoX+offsetX-float64(len(tier))*3.0), nY+12, color.RGBA{180, 180, 180, 255}, 10) }
			}
		}
	}
	icX, icY := int(isoX+offsetX+25), int(by)
	if a.State.Hunger < 30 { textRenderer.DrawTextAt(screen, "🍗", icX, icY, color.RGBA{255, 200, 0, 255}, 14); icX += 12 }
	if a.State.Thirst < 30 { textRenderer.DrawTextAt(screen, "💧", icX, icY, color.RGBA{0, 255, 255, 255}, 14); icX += 12 }
	if a.State.Fatigue < 30 { textRenderer.DrawTextAt(screen, "💤", icX, icY, color.RGBA{100, 100, 255, 255}, 14); icX += 12 }
	if a.State.IsSeptic { textRenderer.DrawTextAt(screen, "☣️", icX, icY, color.RGBA{150, 255, 0, 255}, 14); icX += 12 }
	if a.ActionState == ActorBerserk { textRenderer.DrawTextAt(screen, "💢", icX, icY, color.RGBA{255, 0, 0, 255}, 14); icX += 12 }
}
