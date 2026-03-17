package game

import (
	"image/color"
	"math"
	"oinakos/internal/engine"
)

// DrawActorGetSprite returns the sprite to draw for an actor based on its state and facing.
func DrawActorGetSprite(a *Actor) engine.Image {
	if a.Config == nil {
		return nil
	}
	var drawSprite engine.Image
	// Facing back (North)
	if a.Facing == DirNE || a.Facing == DirNW {
		if img := a.Config.BackImage; img != nil {
			drawSprite = img
		} else if img := a.Config.StaticImage; img != nil {
			drawSprite = img
		}
	} else {
		// Facing front (South)
		if img := a.Config.StaticImage; img != nil {
			drawSprite = img
		}
	}

	// State-based overrides
	if a.State == ActorDead {
		return a.Config.CorpseImage
	} else if a.HitTimer > 0 {
		// Hit animation: toggle between hit frames
		if img := a.Config.PickHitImage(a.Tick / 15); img != nil {
			drawSprite = img
		}
	} else if a.State == ActorAttacking {
		// Attack animation
		cooldown := 30 // Default
		if img := a.Config.PickAttackImage(a.Tick / cooldown); img != nil {
			drawSprite = img
		}
	}
	return drawSprite
}

// DrawActorGetOptions returns the DrawImageOptions for an actor.
func DrawActorGetOptions(a *Actor, offsetX, offsetY float64, isPlayableCharacter bool) *engine.DrawImageOptions {
	isoX, isoY := engine.CartesianToIso(a.X, a.Y)
	drawSprite := DrawActorGetSprite(a)
	if drawSprite == nil {
		return engine.NewDrawImageOptions()
	}

	w, h := drawSprite.Size()
	scale := 1.0
	flip := 1.0
	if a.Facing == DirSE || a.Facing == DirNE {
		flip = -1.0
	}

	op := engine.NewDrawImageOptions()
	op.Scale(scale*flip, scale)

	// Anchoring logic
	tx := isoX + offsetX
	if flip < 0 {
		tx += float64(w) * scale / 2
	} else {
		tx -= float64(w) * scale / 2
	}

	ty := isoY + offsetY - float64(h)*scale*0.85

	// Procedural Animation Overrides
	if a.State == ActorDead {
		ty = isoY + offsetY - float64(h)*scale*0.5
	} else if a.State == ActorWalking {
		// Bobbing effect
		bobScale := 2.0
		bobFreq := 0.2
		if isPlayableCharacter {
			bobScale = 3.0
			bobFreq = 0.3
		}
		bob := math.Sin(float64(a.Tick)*bobFreq) * bobScale
		ty += bob
	} else if a.State == ActorAttacking {
		// Lunge effect
		lungeAmt := 0.0
		if isPlayableCharacter {
			if a.Tick < 15 {
				lungeAmt = (float64(a.Tick) / 15.0) * 5.0
			} else {
				lungeAmt = (float64(30-a.Tick) / 15.0) * 5.0
			}
		} else {
			if a.Tick%60 < 15 {
				lungeAmt = (float64(a.Tick%60) / 15.0) * 5.0
			} else if a.Tick%60 < 30 {
				lungeAmt = 5.0 - (float64(a.Tick%60-15) / 15.0) * 5.0
			}
		}

		if flip < 0 {
			tx += lungeAmt
		} else {
			tx -= lungeAmt
		}
	}

	op.Translate(tx, ty)
	return op
}

// DrawActor is the unified rendering function for any Actor (PlayableCharacter or NPC).
func DrawActor(a *Actor, screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64, isPlayableCharacter bool) {
	if screen == nil || a.Config == nil {
		return
	}

	drawSprite := DrawActorGetSprite(a)
	if drawSprite == nil {
		return
	}

	op := DrawActorGetOptions(a, offsetX, offsetY, isPlayableCharacter)

	// Draw Alignment Indicator - Render BEFORE sprite to be behind the feet
	// Only draw here if NOT occluded. If occluded, it will be drawn in the UI pass on top.
	if !a.IsOccluded {
		DrawAlignmentIndicator(screen, vectorRenderer, a.X, a.Y, offsetX, offsetY, a.Alignment, a.IsAlive(), false)
	}

	// Palette Swapping (Shader)
	hasPalette := a.Config.PrimaryColor != "" || a.Config.SecondaryColor != ""
	if hasPalette && paletteShader != nil {
		uniforms := make(map[string]any)
		pArr := HexToRGBA(a.Config.PrimaryColor)
		sArr := HexToRGBA(a.Config.SecondaryColor)
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

// DrawActorUI draws the UI elements for an actor (alignment indicator, health bar, name tag).
func DrawActorUI(g *Game, a *Actor, screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, offsetX, offsetY float64, isPlayableCharacter bool, debug bool) {
	if screen == nil || a.Config == nil || !a.IsAlive() {
		return
	}

	isoX, isoY := engine.CartesianToIso(a.X, a.Y)

	// If occluded, draw the alignment indicator (solid) on top
	if a.IsOccluded {
		DrawAlignmentIndicator(screen, vectorRenderer, a.X, a.Y, offsetX, offsetY, a.Alignment, a.IsAlive(), true)
		
		// Draw black silhouette of the part that is BEHIND obstacles
		sb := g.GetSilhouetteBuffer()
		if sb != nil {
			sb.Clear()
			
			// 1. Draw actor silhouette (solid black) to buffer
			sOp := *DrawActorGetOptions(a, offsetX, offsetY, isPlayableCharacter)
			sOp.SetColorScale(0, 0, 0, 1) // Pure black
			
			sprite := DrawActorGetSprite(a)
			if sprite != nil {
				sb.DrawImage(sprite, &sOp)
				
				// 2. Multiply with obstacle masks to only keep the "behind" parts
				actorSortY := a.GetSortY()
				for _, o := range g.obstacles {
					if o.GetSortY() > actorSortY {
						oOp := engine.NewDrawImageOptions()
						ox, oy := o.GetIsoPos()
						oOp.Translate(ox+offsetX, oy+offsetY)
						oOp.Blend = engine.BlendDestinationIn
						
						if o.Archetype != nil && o.Archetype.Image != nil {
							sb.DrawImage(o.Archetype.Image, oOp)
						}
					}
				}
				
				// 3. Draw the final silhouette buffer to the screen
				screen.DrawImage(sb, engine.NewDrawImageOptions())
			}
		}
	}

	// UI Elements (Health bar for NPCs, Names)
	if !isPlayableCharacter {
		// Health Bar for NPCs
		barWidth := 40.0
		barHeight := 4.0
		bx := isoX + offsetX - barWidth/2
		
		// Use sprite height if available for positioning
		h := 160.0 // Default 160x160
		if img := a.Config.StaticImage; img != nil {
			_, ih := img.Size()
			h = float64(ih)
		}
		by := isoY + offsetY - h*0.9

		if vectorRenderer != nil {
			vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth), float32(barHeight), color.RGBA{100, 0, 0, 255}, false)
			hpFrac := float32(a.Health) / float32(a.MaxHealth)
			if hpFrac > 0 {
				vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth)*hpFrac, float32(barHeight), color.RGBA{0, 255, 0, 255}, false)
			}
		}
	}

	// Name Tag
	if textRenderer != nil {
		name := a.Name
		if name == "" && a.Config != nil {
			name = a.Config.Name
		}
		if isPlayableCharacter && name == "" {
			name = "Player"
		}
		if name != "" {
			nameX := int(isoX + offsetX - float64(len(name))*3.5)
			nameY := int(isoY + offsetY + 5)
			var textColor color.Color = color.White
			if !isPlayableCharacter && a.Config.Unique {
				textColor = color.RGBA{218, 165, 32, 255} // Golden
			}
			textRenderer.DrawTextAt(screen, name, nameX, nameY, textColor, 12)
		}
	}
}
