package game

import (
	"image/color"
	"math"
	"oinakos/internal/engine"
	"strings"
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
	if n.Facing == DirNE || n.Facing == DirNW {
		if img, ok := n.Archetype.BackImage.(engine.Image); ok {
			drawSprite = img
		} else if img, ok := n.Archetype.StaticImage.(engine.Image); ok {
			drawSprite = img
		}
	} else {
		if img, ok := n.Archetype.StaticImage.(engine.Image); ok {
			drawSprite = img
		}
	}
	if n.State == NPCDead {
		if img, ok := n.Archetype.CorpseImage.(engine.Image); ok {
			drawSprite = img
		} else {
			drawSprite = nil
		}
	} else if n.BloodTimer > 0 {
		if img := n.Archetype.PickHitImage(n.Tick / 15); img != nil {
			drawSprite = img
		}
	} else if n.State == NPCAttacking {
		cooldown := n.AttackCooldown
		if cooldown <= 0 {
			cooldown = 1
		}
		if img := n.Archetype.PickAttackImage(n.Tick / cooldown); img != nil {
			drawSprite = img
		}
	}

	scale := 1.0
	frameTick := n.Tick

	if n.Archetype.Engine == "paperdoll" && n.Archetype.Paperdoll != nil {
		n.drawNPCLayersStrict(screen, isoX+offsetX, isoY+offsetY, scale, paletteShader, vectorRenderer, frameTick)
	} else if drawSprite != nil {
		w, h := drawSprite.Size()
		op := engine.NewDrawImageOptions()

		flip := 1.0
		if n.Facing == DirSE || n.Facing == DirNE {
			flip = -1.0
		}

		op.Scale(scale*flip, scale)

		tx := isoX + offsetX
		if flip < 0 {
			tx += float64(w) * scale / 2
		} else {
			tx -= float64(w) * scale / 2
		}

		ty := isoY + offsetY - float64(h)*scale*0.85

		if n.State == NPCDead {
			ty = isoY + offsetY - float64(h)*scale*0.5
		} else if n.State == NPCWalking {
			bob := math.Sin(float64(n.Tick)*0.2) * 2.0
			ty += bob
		} else if n.State == NPCAttacking {
			lungeAmt := 0.0
			attackPhase := float64(n.AttackTimer) / float64(n.AttackCooldown)
			if attackPhase < 0.2 {
				lungeAmt = (attackPhase / 0.2) * 5.0
			} else if attackPhase < 0.5 {
				lungeAmt = 5.0 - ((attackPhase-0.2)/0.3)*5.0
			}
			if flip < 0 {
				tx += lungeAmt
			} else {
				tx -= lungeAmt
			}
		}

		op.Translate(tx, ty)

		// Draw Alignment Ellipse under feet
		if n.IsAlive() && vectorRenderer != nil {
			var clr color.Color
			switch n.Alignment {
			case AlignmentAlly:
				clr = ColorAlly
			case AlignmentEnemy:
				clr = ColorEnemy
			case AlignmentNeutral:
				clr = ColorNeutral
			default:
				clr = color.RGBA{150, 150, 150, 150}
			}
			vectorRenderer.DrawEllipse(screen, float32(isoX+offsetX), float32(isoY+offsetY), 30, 15, clr, 1, true)
		}

		hasPalette := n.Archetype.PrimaryColor != "" || n.Archetype.SecondaryColor != ""
		if hasPalette && paletteShader != nil {
			uniforms := make(map[string]interface{})
			pArr := HexToRGBA(n.Archetype.PrimaryColor)
			sArr := HexToRGBA(n.Archetype.SecondaryColor)
			uniforms["PrimaryColor"] = pArr[:]
			uniforms["SecondaryColor"] = sArr[:]

			if g, ok := vectorRenderer.(engine.Graphics); ok {
				g.DrawImageWithShader(screen, drawSprite, paletteShader, uniforms, op)
			} else {
				screen.DrawImage(drawSprite, op)
			}
		} else {
			screen.DrawImage(drawSprite, op)
		}
	}

	// Only draw UI for living NPCs
	if n.IsAlive() {
		h := 160.0
		if drawSprite != nil {
			_, hInt := drawSprite.Size()
			h = float64(hInt)
		}
		// Draw Health Bar above NPC
		barWidth := 40.0
		barHeight := 4.0
		bx := isoX + offsetX - barWidth/2
		by := isoY + offsetY - h*scale*0.9

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

func (n *NPC) drawNPCLayersStrict(screen engine.Image, isoX, isoY, scale float64, paletteShader engine.Shader, vectorRenderer engine.VectorRenderer, frameTick int) {
	if n.Archetype == nil || n.Archetype.Paperdoll == nil {
		return
	}
	layers := []string{"body", "head_details", "tunic", "cape", "armor", "weapon_r"}

	actionName := "static"
	switch n.State {
	case NPCWalking:
		walkFrames := []string{"walk1", "walk2", "walk3", "walk4"}
		frameIdx := (frameTick / 10) % 4
		actionName = walkFrames[frameIdx]
	case NPCAttacking:
		actionName = "attack"
	case NPCDead:
		actionName = "static" // fallback
	}

	if n.BloodTimer > 0 {
		actionName = "hit"
	}

	for _, layerName := range layers {
		layerActions, ok := n.Archetype.Paperdoll.Layers[layerName]
		if !ok {
			continue
		}

		layer, ok := layerActions[actionName]
		if !ok {
			if strings.HasPrefix(actionName, "walk") {
				layer, ok = layerActions["walk1"]
			}
			if !ok {
				layer, ok = layerActions["static"]
			}
			if !ok {
				continue
			}
		}

		facing := n.Facing
		flip := 1.0

		// Horizontal symmetry mapping:
		// We flip based on the direction he is looking to match movement.
		switch facing {
		case DirSE, DirE, DirNE:
			flip = -1.0
		case DirSW:
			facing = DirSE
		case DirW:
			facing = DirE
		case DirNW:
			facing = DirNE
		}

		img, ok := layer[facing]
		if !ok {
			// fallback to DirSE
			img, ok = layer[DirSE]
			if !ok {
				continue
			}
		}

		w, h := img.Size()
		op := engine.NewDrawImageOptions()

		// 1. Move pivot (bottom center of footprint) to origin
		op.Translate(-float64(w)/2, -float64(h)*0.85)

		// 2. Scale and Flip
		op.Scale(scale*flip, scale)

		// 3. Procedural animations (Rotate/Bob around pivot)
		finalTx := isoX
		finalTy := isoY
		if n.State == NPCWalking {
			// Bobbing and sway are handled by paperdoll poses
		}

		// 4. Translate to final world position
		op.Translate(finalTx, finalTy)

		// Only tunic and cape get palette swapping
		if (layerName == "tunic" || layerName == "cape") && paletteShader != nil && (n.Archetype.PrimaryColor != "" || n.Archetype.SecondaryColor != "") {
			uniforms := make(map[string]interface{})
			pArr := HexToRGBA(n.Archetype.PrimaryColor)
			sArr := HexToRGBA(n.Archetype.SecondaryColor)
			uniforms["PrimaryColor"] = pArr[:]
			uniforms["SecondaryColor"] = sArr[:]

			if g, ok := vectorRenderer.(engine.Graphics); ok {
				g.DrawImageWithShader(screen, img, paletteShader, uniforms, op)
			} else {
				screen.DrawImage(img, op)
			}
		} else {
			screen.DrawImage(img, op)
		}
	}
}

func dirToSuffix(d Direction) string {
	switch d {
	case DirN:
		return "n"
	case DirNE:
		return "ne"
	case DirE:
		return "e"
	case DirSE:
		return "se"
	case DirS:
		return "s"
	case DirSW:
		return "sw"
	case DirW:
		return "w"
	case DirNW:
		return "nw"
	}
	return "se"
}
