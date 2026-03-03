package game

import (
	"image/color"
	"math"
	"oinakos/internal/engine"
)

func (mc *MainCharacter) Draw(screen engine.Image, textRenderer engine.TextRenderer, offsetX, offsetY float64) {
	if screen == nil {
		return
	}
	isoX, isoY := engine.CartesianToIso(mc.X, mc.Y)

	var drawSprite engine.Image
	if mc.Config != nil {
		if img, ok := mc.Config.StaticImage.(engine.Image); ok {
			drawSprite = img
		}
		if mc.State == StateDead {
			if img, ok := mc.Config.CorpseImage.(engine.Image); ok {
				drawSprite = img
			}
		} else if mc.HitTimer > 0 {
			// Hit frame is 15 ticks
			if img := mc.Config.PickHitImage(mc.Tick / 15); img != nil {
				drawSprite = img
			}
		} else if mc.State == StateAttacking {
			// Attack animation is 30 ticks
			if img := mc.Config.PickAttackImage(mc.Tick / 30); img != nil {
				drawSprite = img
			}
		}
	}

	if drawSprite == nil {
		return
	}

	w, h := drawSprite.Size()

	scale := 1.0

	flip := 1.0
	if mc.Facing == DirSE || mc.Facing == DirNE {
		flip = -1.0
	}

	op := engine.NewDrawImageOptions()
	op.Scale(scale*flip, scale)

	tx := isoX + offsetX
	if flip < 0 {
		tx += float64(w) * scale / 2
	} else {
		tx -= float64(w) * scale / 2
	}

	ty := isoY + offsetY - float64(h)*scale*0.85

	// Procedural Animation Overrides
	if mc.State == StateDead {
		ty = isoY + offsetY - float64(h)*scale*0.5
	} else if mc.State == StateWalking {
		// Bobbing effect
		bob := math.Sin(float64(mc.Tick)*0.3) * 3.0
		ty += bob
	} else if mc.State == StateAttacking {
		// Lunge effect (move forward then back)
		lungeAmt := 0.0
		if mc.Tick < 15 {
			lungeAmt = (float64(mc.Tick) / 15.0) * 5.0
		} else {
			lungeAmt = (float64(30-mc.Tick) / 15.0) * 5.0
		}

		if flip < 0 {
			tx += lungeAmt // Lunge right
		} else {
			tx -= lungeAmt // Lunge left
		}
	}

	op.Translate(tx, ty)
	screen.DrawImage(drawSprite, op)

	// Draw Name below feet
	if textRenderer != nil && mc.State != StateDead {
		name := "Player"
		if mc.Config != nil && mc.Config.Name != "" {
			name = mc.Config.Name
		}
		nameX := int(isoX + offsetX - float64(len(name))*3.5)
		nameY := int(isoY + offsetY + 5)
		textRenderer.DebugPrintAt(screen, name, nameX, nameY, color.White)
	}
}
