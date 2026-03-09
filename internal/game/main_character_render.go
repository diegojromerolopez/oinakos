package game

import (
	"image/color"
	"math"
	"oinakos/internal/engine"
	"strings"
)

func (mc *MainCharacter) Draw(screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64) {
	if screen == nil {
		return
	}
	isoX, isoY := engine.CartesianToIso(mc.X, mc.Y)

	var drawSprite engine.Image
	if mc.Config != nil {
		if mc.Facing == DirNE || mc.Facing == DirNW {
			if img, ok := mc.Config.BackImage.(engine.Image); ok {
				drawSprite = img
			} else if img, ok := mc.Config.StaticImage.(engine.Image); ok {
				drawSprite = img
			}
		} else {
			if img, ok := mc.Config.StaticImage.(engine.Image); ok {
				drawSprite = img
			}
		}
		if mc.State == StateDead {
			if img, ok := mc.Config.CorpseImage.(engine.Image); ok {
				drawSprite = img
			}
		} else if mc.HitTimer > 0 {
			if img := mc.Config.PickHitImage(mc.Tick / 15); img != nil {
				drawSprite = img
			}
		} else if mc.State == StateAttacking {
			if img := mc.Config.PickAttackImage(mc.Tick / 30); img != nil {
				drawSprite = img
			}
		}
	}

	scale := 1.0
	frameTick := mc.Tick

	// Paperdoll Logic
	if mc.Config != nil && mc.Config.Engine == "paperdoll" && mc.Config.Paperdoll != nil {
		mc.drawPlayerLayersStrict(screen, isoX+offsetX, isoY+offsetY, scale, vectorRenderer, paletteShader, frameTick)
	} else if drawSprite != nil {
		w, h := drawSprite.Size()
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

		if mc.State == StateDead {
			ty = isoY + offsetY - float64(h)*scale*0.5
		} else if mc.State == StateWalking {
			bob := math.Sin(float64(mc.Tick)*0.3) * 3.0
			ty += bob
		} else if mc.State == StateAttacking {
			lungeAmt := 0.0
			if mc.Tick < 15 {
				lungeAmt = (float64(mc.Tick) / 15.0) * 5.0
			} else {
				lungeAmt = (float64(30-mc.Tick) / 15.0) * 5.0
			}
			if flip < 0 {
				tx += lungeAmt
			} else {
				tx -= lungeAmt
			}
		}

		op.Translate(tx, ty)

		hasPalette := mc.Config.PrimaryColor != "" || mc.Config.SecondaryColor != ""
		if hasPalette && paletteShader != nil {
			uniforms := make(map[string]interface{})
			pArr := HexToRGBA(mc.Config.PrimaryColor)
			sArr := HexToRGBA(mc.Config.SecondaryColor)
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

	// UI
	if mc.State != StateDead && vectorRenderer != nil {
		vectorRenderer.DrawEllipse(screen, float32(isoX+offsetX), float32(isoY+offsetY), 30, 15, ColorMainCharacter, 1, true)
	}

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

func (mc *MainCharacter) drawPlayerLayersStrict(screen engine.Image, isoX, isoY, scale float64, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, frameTick int) {
	if mc.Config == nil || mc.Config.Paperdoll == nil {
		return
	}

	actionName := "static"
	switch mc.State {
	case StateWalking:
		walkFrames := []string{"walk1", "walk2", "walk3", "walk4"}
		frameIdx := (frameTick / 10) % 4
		actionName = walkFrames[frameIdx]
	case StateAttacking:
		actionName = "attack"
	case StateDead:
		actionName = "static" // fallback
	}

	if mc.HitTimer > 0 {
		actionName = "hit"
	}

	layers := []string{"body", "head_details", "tunic", "cape", "armor", "weapon_r"}

	for _, layerName := range layers {
		layerActions, ok := mc.Config.Paperdoll.Layers[layerName]
		if !ok {
			// log.Printf("DEBUG: Player layer %s not found in config", layerName)
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

		facing := mc.Facing
		flip := 1.0

		// Horizontal symmetry mapping:
		// We want the character to look in the direction of movement.
		// If the base render ('e') looks right, we flip for west.
		// If the user says it's opposite, then 'e' must look left, so we flip for east.
		switch facing {
		case DirSE, DirE, DirNE:
			// Inverting current logic: flip for the right-facing directions
			if facing == DirSE {
				facing = DirSE
			} // no change
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
			// fallback to DirSE if specific direction missing
			img, ok = layer[DirSE]
			if !ok {
				continue
			}
		}

		w, h := img.Size()
		op := engine.NewDrawImageOptions()

		// 1. Move pivot (bottom center of footprint) to origin
		op.Translate(-float64(w)/2, -float64(h)*0.85)

		// 2. Apply Scale and Flip (around origin/pivot)
		op.Scale(scale*flip, scale)

		// 3. Procedural animations (Rotate around pivot)
		finalTx := isoX
		finalTy := isoY
		if mc.State == StateWalking {
			// Bobbing and sway are now integrated into the paperdoll poses
		}

		// 4. Translate to final world position (offset by camera/scrolling)
		op.Translate(finalTx, finalTy)

		if (layerName == "tunic" || layerName == "cape") && paletteShader != nil && (mc.Config.PrimaryColor != "" || mc.Config.SecondaryColor != "") {
			uniforms := make(map[string]interface{})
			pArr := HexToRGBA(mc.Config.PrimaryColor)
			sArr := HexToRGBA(mc.Config.SecondaryColor)
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
