package game

import (
	"fmt"
	"image/color"
	"oinakos/internal/engine"
)

func (gr *GameRenderer) drawInventoryScreen(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 80}, false)
	dialogW, dialogH := 900, 600
	dialogX, dialogY := (g.width-dialogW)/2, (g.height-dialogH)/2
	
	gr.graphics.DrawFilledRect(screen, float32(dialogX-2), float32(dialogY-2), float32(dialogW+4), float32(dialogH+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(dialogX), float32(dialogY), float32(dialogW), float32(dialogH), color.RGBA{15, 15, 15, 245}, false)
	
	title := "INVENTORY & EQUIPMENT"
	tw, _ := gr.graphics.MeasureText(title, 36)
	gr.graphics.DrawTextAt(screen, title, dialogX+(dialogW-int(tw))/2, dialogY+45, color.RGBA{218, 165, 32, 255}, 36)
	
	pc := g.playableCharacter
	weightStr := fmt.Sprintf("Weight: %s / %s", g.settings.FormatWeight(pc.GetTotalWeight()), g.settings.FormatWeight(pc.MaxWeight))
	ww, _ := gr.graphics.MeasureText(weightStr, 14)
	gr.graphics.DrawTextAt(screen, weightStr, dialogX+dialogW-int(ww)-40, dialogY+45, color.White, 14)

	traumas := pc.GetActiveTraumas()
	if len(traumas) > 0 {
		gr.graphics.DrawTextAt(screen, "TRAUMAS:", dialogX+dialogW-int(ww)-30, dialogY+70, color.RGBA{200, 50, 50, 255}, 12)
		for i, t := range traumas { gr.graphics.DrawTextAt(screen, "- "+t, dialogX+dialogW-int(ww)-30, dialogY+90+i*18, color.RGBA{220, 100, 100, 255}, 11) }
	}
	
	dollCenterX, dollCenterY := dialogX+220, dialogY+300
	slots := []struct { id, label string; x, y int }{
		{"head", "Head", dollCenterX, dollCenterY - 140}, {"shield", "L Arm", dollCenterX - 110, dollCenterY - 40},
		{"body", "Torso", dollCenterX, dollCenterY - 40}, {"weapon", "R Arm", dollCenterX + 110, dollCenterY - 40},
		{"ring1", "L Ring", dollCenterX - 110, dollCenterY + 50}, {"ring2", "R Ring", dollCenterX + 110, dollCenterY + 50},
		{"legs", "Legs/Feet", dollCenterX, dollCenterY + 110},
	}
	for _, slot := range slots {
		sx, sy := slot.x, slot.y
		gr.graphics.DrawTextAt(screen, slot.label, sx-40, sy-35, color.RGBA{170, 170, 170, 255}, 14)
		if it, hasIt := pc.Slots[slot.id]; hasIt && it != nil && it.Config != nil {
			if it.Config.Sprite != nil {
				_, sh := it.Config.Sprite.Size()
				scale := 32.0 / float64(sh)
				sw, _ := it.Config.Sprite.Size() // Re-declare sw for use below
				if float64(sw) > float64(sh) { scale = 32.0 / float64(sw) }
				op := engine.NewDrawImageOptions()
				op.GeoM.Scale(scale, scale); op.GeoM.Translate(float64(sx)-float64(sw)*scale/2, float64(sy)-float64(sh)*scale/2-5)
				screen.DrawImage(it.Config.Sprite, op)
			}
			name := it.Config.Name; if len(name) > 15 { name = name[:13] + ".." }; nw, _ := gr.graphics.MeasureText(name, 14)
			gr.graphics.DrawTextAt(screen, name, sx-int(nw)/2, sy+20, color.RGBA{218, 165, 32, 255}, 14)
			if it.Config.Resistance > 0 {
				res := fmt.Sprintf("%d/%d", it.Resistance, it.Config.Resistance); rw, _ := gr.graphics.MeasureText(res, 12)
				gr.graphics.DrawTextAt(screen, res, sx-int(rw)/2, sy+38, color.RGBA{200, 200, 100, 255}, 12)
			}
			gr.graphics.DrawTextAt(screen, "[X]", sx+30, sy-15, color.RGBA{200, 50, 50, 255}, 14)
		} else { em, _ := gr.graphics.MeasureText("Empty", 14); gr.graphics.DrawTextAt(screen, "Empty", sx-int(em)/2, sy+5, color.RGBA{100, 100, 100, 255}, 14) }
	}
	
	listStartX, listStartY, listW := dialogX+400, dialogY+80, dialogW-420
	gr.graphics.DrawTextAt(screen, "Backpack", listStartX, listStartY-10, color.RGBA{218, 165, 32, 255}, 20)
	if len(pc.Inventory) == 0 { gr.graphics.DrawTextAt(screen, "Backpack is empty.", listStartX+50, listStartY+50, color.RGBA{136, 136, 136, 255}, 16)
	} else {
		for i, it := range pc.Inventory {
			if it == nil || it.Config == nil { continue }
			itemY := listStartY + 20 + i*40
			if it.Config.Sprite != nil {
				sw, sh := it.Config.Sprite.Size(); scale := 24.0 / float64(sh); if float64(sw) > float64(sh) { scale = 24.0 / float64(sw) }
				op := engine.NewDrawImageOptions(); op.GeoM.Scale(scale, scale); op.GeoM.Translate(float64(listStartX)+5, float64(itemY)+5)
				screen.DrawImage(it.Config.Sprite, op)
			}
			gr.graphics.DrawTextAt(screen, it.Config.Name, listStartX+50, itemY+22, color.RGBA{218, 165, 32, 255}, 18)
			desc := it.Config.Description; if len(desc) > 35 { desc = desc[:32] + "..." }; gr.graphics.DrawTextAt(screen, desc, listStartX+220, itemY+23, color.RGBA{180, 180, 180, 255}, 13)
			gr.graphics.DrawTextAt(screen, "[X]", listStartX+listW-35, itemY+25, color.RGBA{200, 50, 50, 255}, 16)
			if it.Config.Content != "" { gr.graphics.DrawTextAt(screen, "[R]", listStartX+listW-75, itemY+25, color.RGBA{0, 150, 255, 255}, 16) }
			if it.Config.Type == "consumable" { gr.graphics.DrawTextAt(screen, "[E]", listStartX+listW-115, itemY+25, color.RGBA{0, 255, 150, 255}, 16) }
		}
	}
	gr.graphics.DrawTextAt(screen, "[CLOSE]", dialogX+400, dialogY+dialogH-30, color.RGBA{218, 165, 32, 255}, 20)
	if g.ActiveBook != nil { gr.drawBookOverlay(screen) }
}
