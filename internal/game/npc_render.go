package game

import (
	"image/color"
	"log"
	"math"
	"oinakos/internal/engine"
)

func (n *NPC) Draw(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64) {
	if screen == nil {
		return
	}
	if n.Archetype == nil {
		return
	}

	isoX, isoY := engine.CartesianToIso(n.X, n.Y)

	var drawSprite engine.Image
	if img, ok := n.Archetype.StaticImage.(engine.Image); ok {
		drawSprite = img
	}
	if n.State == NPCDead {
		if img, ok := n.Archetype.CorpseImage.(engine.Image); ok {
			drawSprite = img
		} else {
			drawSprite = nil // Complete removal (invisibility) if no corpse
		}
	} else if n.BloodTimer > 0 {
		if img := n.Archetype.PickHitImage(); img != nil {
			drawSprite = img
		}
	} else if n.State == NPCAttacking {
		if img := n.Archetype.PickAttackImage(); img != nil {
			drawSprite = img
		}
	}

	if drawSprite == nil {
		log.Printf("NPC %s: drawSprite is nil (state=%d, hasStatic=%v)", n.Name, n.State, n.Archetype.StaticImage != nil)

		return
	}

	w, h := drawSprite.Size()
	op := engine.NewDrawImageOptions()
	scale := 1.0

	flip := 1.0
	if n.Facing == DirSE || n.Facing == DirNE {
		flip = -1.0
	}

	op.Scale(scale*flip, scale)

	// Anchoring: if flipped, we need to translate differently
	tx := isoX + offsetX
	if flip < 0 {
		tx += float64(w) * scale / 2
	} else {
		tx -= float64(w) * scale / 2
	}

	ty := isoY + offsetY - float64(h)*scale*0.85

	// Procedural Animation Overrides
	if n.State == NPCDead {
		// Lie flat on the ground
		ty = isoY + offsetY - float64(h)*scale*0.5
	} else if n.State == NPCWalking {
		// Bobbing effect
		bob := math.Sin(float64(n.Tick)*0.2) * 2.0
		ty += bob
	} else if n.State == NPCAttacking {
		// Lunge effect
		lungeAmt := 0.0
		attackPhase := float64(n.AttackTimer) / float64(n.AttackCooldown)
		if attackPhase < 0.2 { // Quick forward lunge
			lungeAmt = (attackPhase / 0.2) * 5.0
		} else if attackPhase < 0.5 { // Hold slightly, then pull back
			lungeAmt = 5.0 - ((attackPhase-0.2)/0.3)*5.0
		}

		if flip < 0 {
			tx += lungeAmt // Lunge right
		} else {
			tx -= lungeAmt // Lunge left
		}
	}

	op.Translate(tx, ty)

	// Palette Swapping
	hasPalette := n.Archetype.PrimaryColor != "" || n.Archetype.SecondaryColor != ""
	if hasPalette && paletteShader != nil {
		uniforms := make(map[string]interface{})
		pArr := HexToRGBA(n.Archetype.PrimaryColor)
		sArr := HexToRGBA(n.Archetype.SecondaryColor)
		uniforms["PrimaryColor"] = pArr[:] // Convert to slice
		uniforms["SecondaryColor"] = sArr[:]

		if g, ok := vectorRenderer.(engine.Graphics); ok {
			g.DrawImageWithShader(screen, drawSprite, paletteShader, uniforms, op)
		} else {
			screen.DrawImage(drawSprite, op)
		}
	} else {
		screen.DrawImage(drawSprite, op)
	}

	// Only draw UI for living NPCs
	if n.IsAlive() {
		// Draw Health Bar above NPC
		barWidth := 40.0
		barHeight := 4.0
		bx := isoX + offsetX - barWidth/2
		by := isoY + offsetY - float64(h)*scale*0.9 // Floating above head area

		if vectorRenderer != nil {
			vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth), float32(barHeight), color.RGBA{100, 0, 0, 255}, false)
			hpFrac := float32(n.Health) / float32(n.MaxHealth)
			if hpFrac > 0 {
				vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth)*hpFrac, float32(barHeight), color.RGBA{0, 255, 0, 255}, false)
			}
		}

		// Draw Name below feet
		if textRenderer != nil {
			nameX := int(isoX + offsetX - float64(len(n.Name))*3.5)
			nameY := int(isoY + offsetY + 5)
			if n.Archetype.Unique {
				textRenderer.DebugPrintAt(screen, n.Name, nameX, nameY, color.RGBA{218, 165, 32, 255})
			} else {
				textRenderer.DebugPrintAt(screen, n.Name, nameX, nameY, color.White)
			}
		}
	}
}
