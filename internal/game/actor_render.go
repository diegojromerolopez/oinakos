package game

import (
	"image/color"
	"math"
	"strings"
	"oinakos/internal/engine"
)

// DrawActorGetSprite returns the sprite to draw for an actor based on its state and facing.
func DrawActorGetSprite(a *Actor) engine.Image {
	if a.Config == nil {
		return nil
	}

	conf := a.Config
	var mod *ModelConfig
	if a.SelectedModel != "" && conf.Models != nil {
		mod = conf.Models[a.SelectedModel]
	}

	var drawSprite engine.Image

	// 1. Determine base orientation/state image
	if a.Facing == DirNE || a.Facing == DirNW {
		// Back facing
		if mod != nil && mod.BackImage != nil {
			drawSprite = mod.BackImage
		} else if conf.BackImage != nil {
			drawSprite = conf.BackImage
		} else {
			// Fallback to static
			if mod != nil && mod.StaticImage != nil {
				drawSprite = mod.StaticImage
			} else {
				drawSprite = conf.StaticImage
			}
		}
	} else {
		// Front facing
		if a.IsPregnant && a.GestationTicks < 43200 {
			if mod != nil && mod.PregnantImage != nil {
				drawSprite = mod.PregnantImage
			} else if conf.PregnantImage != nil {
				drawSprite = conf.PregnantImage
			}
		}
		
		if drawSprite == nil {
			if mod != nil && mod.StaticImage != nil {
				drawSprite = mod.StaticImage
			} else {
				drawSprite = conf.StaticImage
			}
		}
	}

	// 2. State-based priority overrides
	if a.State == ActorDead || a.State == ActorIncapacitated {
		if mod != nil && mod.CorpseImage != nil {
			return mod.CorpseImage
		}
		return conf.CorpseImage
	}

	if a.HitTimer > 0 {
		if mod != nil && mod.HitImage != nil {
			return mod.HitImage
		}
		if img := conf.PickHitImage(a.Tick / 15); img != nil {
			return img
		}
	}

	if a.State == ActorAttacking {
		if mod != nil && mod.AttackImage != nil {
			return mod.AttackImage
		}
		cooldown := 30
		if img := conf.PickAttackImage(a.Tick / cooldown); img != nil {
			return img
		}
	}

	if a.State == ActorChopping {
		if conf.ChoppingImage != nil {
			drawSprite = conf.ChoppingImage
		} else {
			if mod != nil && mod.AttackImage != nil {
				return mod.AttackImage
			}
			cooldown := 30
			if img := conf.PickAttackImage(a.Tick / cooldown); img != nil {
				return img
			}
		}
	}

	if a.State == ActorDigging {
		if conf.DiggingImage != nil {
			drawSprite = conf.DiggingImage
		} else {
			if mod != nil && mod.AttackImage != nil {
				return mod.AttackImage
			}
			cooldown := 30
			if img := conf.PickAttackImage(a.Tick / cooldown); img != nil {
				return img
			}
		}
	} else if a.State == ActorCrouching || a.State == ActorResting {
		if mod != nil && mod.CrouchImage != nil {
			return mod.CrouchImage
		}
		if img := conf.CrouchImage; img != nil {
			return img
		}
		// Fallback to static
		if mod != nil && mod.StaticImage != nil {
			return mod.StaticImage
		}
		return conf.StaticImage
	}
	return drawSprite
}

// DrawActorGetOptions returns the DrawImageOptions for an actor.
func DrawActorGetOptions(a *Actor, offsetX, offsetY float64, isPlayableCharacter bool) *engine.DrawImageOptions {
	isoX, isoY := engine.CartesianToIsoZ(a.X, a.Y, a.Z)
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
	} else if a.State == ActorChopping || a.State == ActorDigging {
		// Similar to attack lunge
		lungeAmt := 0.0
		if a.Tick < 15 {
			lungeAmt = (float64(a.Tick) / 15.0) * 5.0
		} else {
			lungeAmt = (float64(30-a.Tick) / 15.0) * 5.0
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
func DrawActor(a *Actor, screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64, isPlayableCharacter bool, adultMode bool) {
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
		DrawAlignmentIndicator(screen, vectorRenderer, a, offsetX, offsetY, false)
	}

	// Palette Swapping and Trauma (Shader)
	hasPalette := a.Config.PrimaryColor != "" || a.Config.SecondaryColor != ""
	hasTrauma := a.Trauma != (PhysicalTrauma{}) // Check if any trauma is present

	if (hasPalette || hasTrauma) && paletteShader != nil {
		uniforms := make(map[string]any)

		// Palette Colors
		pArr := HexToRGBA(a.Config.PrimaryColor)
		sArr := HexToRGBA(a.Config.SecondaryColor)
		uniforms["PrimaryColor"] = pArr[:]
		uniforms["SecondaryColor"] = sArr[:]

		// Trauma Flags (1.0 = true, 0.0 = false)
		toF := func(b bool) float32 {
			if b {
				return 1.0
			}
			return 0.0
		}
		uniforms["LeftArmLost"] = toF(a.Trauma.LeftArmLost)
		uniforms["RightArmLost"] = toF(a.Trauma.RightArmLost)
		uniforms["LeftLegLost"] = toF(a.Trauma.LeftLegLost)
		uniforms["RightLegLost"] = toF(a.Trauma.RightLegLost)
		uniforms["BurnedAlive"] = toF(a.Trauma.BurnedAlive)
		uniforms["EyesLost"] = float32(a.Trauma.EyesLost)

		// Status Tint (Placeholder for future effects)
		uniforms["StatusTint"] = []float32{0, 0, 0, 0}

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

	isoX, isoY := engine.CartesianToIsoZ(a.X, a.Y, a.Z)

	// If occluded, draw the alignment indicator (solid) on top
	if a.IsOccluded {
		DrawAlignmentIndicator(screen, vectorRenderer, a, offsetX, offsetY, true)
		
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

	// UI Elements (Health and Energy bars for characters, Names)
	barWidth := 40.0
	barHeight := 3.0
	bx := isoX + offsetX - barWidth/2
	
	// Use sprite height if available for positioning
	h := 160.0 // Default 160x160
	if a.Config != nil && a.Config.StaticImage != nil {
		_, ih := a.Config.StaticImage.Size()
		h = float64(ih)
	}
	
	multiplier := 0.9
	if a.Config != nil && a.Config.IsAnimal {
		multiplier = 0.35 // Lower for animals
	}
	by := isoY + offsetY - h*multiplier

	if vectorRenderer != nil {
		// 1. Health Bar
		vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth), float32(barHeight), color.RGBA{80, 0, 0, 255}, false)
		if a.TemporalState.MaxHealthPoints > 0 {
			hpFrac := float32(a.TemporalState.HealthPoints) / float32(a.TemporalState.MaxHealthPoints)
			if hpFrac > 1 { hpFrac = 1 }
			if hpFrac > 0 {
				vectorRenderer.DrawFilledRect(screen, float32(bx), float32(by), float32(barWidth)*hpFrac, float32(barHeight), color.RGBA{0, 255, 0, 255}, false)
			}
		}

		// 2. Needs Bars (3 thin ones below health)
		ebY := by + barHeight + 1
		// Hunger (Brown)
		vectorRenderer.DrawFilledRect(screen, float32(bx), float32(ebY), float32(barWidth), 1, color.RGBA{50, 25, 0, 255}, false)
		vectorRenderer.DrawFilledRect(screen, float32(bx), float32(ebY), float32(barWidth)*float32(a.TemporalState.Hunger/100.0), 1, color.RGBA{210, 105, 30, 255}, false)
		
		ebY += 2
		// Thirst (Blue)
		vectorRenderer.DrawFilledRect(screen, float32(bx), float32(ebY), float32(barWidth), 1, color.RGBA{0, 20, 50, 255}, false)
		vectorRenderer.DrawFilledRect(screen, float32(bx), float32(ebY), float32(barWidth)*float32(a.TemporalState.Thirst/100.0), 1, color.RGBA{0, 191, 255, 255}, false)

		ebY += 2
		// Fatigue (Gold)
		vectorRenderer.DrawFilledRect(screen, float32(bx), float32(ebY), float32(barWidth), 1, color.RGBA{50, 45, 0, 255}, false)
		vectorRenderer.DrawFilledRect(screen, float32(bx), float32(ebY), float32(barWidth)*float32(a.TemporalState.Fatigue/100.0), 1, color.RGBA{255, 215, 0, 255}, false)
	}

	// Name Tag & Relationship Tier
	if textRenderer != nil {
		name := a.Name
		if name == "" && a.Config != nil { name = a.Config.Name }
		if isPlayableCharacter && name == "" { name = "Player" }
		
		tier := ""
		if !isPlayableCharacter && g.playableCharacter != nil {
			tier = g.playableCharacter.GetRelationshipTier(a.Name)
		}

		if name != "" {
			nameX := int(isoX + offsetX - float64(len(name))*3.5)
			nameY := int(isoY + offsetY + 5)
			var textColor color.Color = color.White
			if !isPlayableCharacter && (a.IsTarget || a.MustSurvive) {
				textColor = color.RGBA{218, 165, 32, 255} // Golden
			}
			textRenderer.DrawTextAt(screen, name, nameX, nameY, textColor, 12)
			
			// Trauma Icon next to name (ONLY if AdultMode is ON)
			adultMode := true
			if g.settings != nil { adultMode = g.settings.AdultMode }
			
			hasTrauma := a.Trauma.BurnedAlive || a.Trauma.LeftArmLost || a.Trauma.RightArmLost || a.Trauma.LeftLegLost || a.Trauma.RightLegLost || a.Trauma.EyesLost > 0 || a.Trauma.SpineBroken
			if hasTrauma && adultMode {
				nameWidth, _ := textRenderer.MeasureText(name, 12)
				textRenderer.DrawTextAt(screen, "🏥", nameX + int(nameWidth) + 5, nameY, color.RGBA{255, 50, 50, 255}, 10)
				
				// Hover tooltip
				mouseX, mouseY := g.input.MousePosition()
				if float64(mouseX) >= float64(nameX) && float64(mouseX) <= float64(nameX)+nameWidth+20 && float64(mouseY) >= float64(nameY)-12 && float64(mouseY) <= float64(nameY)+4 {
					desc := a.GetTraumaDescription()
					tw, _ := textRenderer.MeasureText(desc, 10)
					boxW := int(tw) + 12
					tx := mouseX + 10
					ty := mouseY + 10
					// Draw tooltip box and text
					if g.Graphics != nil {
						if vr, ok := g.Graphics.(engine.VectorRenderer); ok {
							vr.DrawFilledRect(screen, float32(tx), float32(ty), float32(boxW), 18, color.RGBA{0, 0, 0, 220}, false)
						}
					}
					textRenderer.DrawTextAt(screen, desc, tx+6, ty+13, color.White, 10)
				}
			}

			if tier != "" && tier != "Neutral" {
				tierX := int(isoX + offsetX - float64(len(tier))*3.0)
				tierY := nameY + 12
				textRenderer.DrawTextAt(screen, tier, tierX, tierY, color.RGBA{180, 180, 180, 255}, 10)
			}
		}
	}

	// Status Icons (Now for everyone!)
	iconX := int(isoX + offsetX + barWidth/2 + 5)
	iconY := int(by)
	
	// 1. Needs status icons
	if a.TemporalState.Hunger < 30 {
		textRenderer.DrawTextAt(screen, "🍗", iconX, iconY, color.RGBA{255, 200, 0, 255}, 14)
		iconX += 12
	}
	if a.TemporalState.Thirst < 30 {
		textRenderer.DrawTextAt(screen, "💧", iconX, iconY, color.RGBA{0, 255, 255, 255}, 14)
		iconX += 12
	}
	if a.TemporalState.Fatigue < 30 {
		textRenderer.DrawTextAt(screen, "💤", iconX, iconY, color.RGBA{100, 100, 255, 255}, 14)
		iconX += 12
	}
	
	// 2. State / Love
	if a.TemporalState.IsSeptic {
		textRenderer.DrawTextAt(screen, "☣️", iconX, iconY, color.RGBA{150, 255, 0, 255}, 14) // Green infection icon
		iconX += 12
	}
	if a.State == ActorBerserk {
		textRenderer.DrawTextAt(screen, "💢", iconX, iconY, color.RGBA{255, 0, 0, 255}, 14)
		iconX += 12
	}
	if g.playableCharacter != nil && !isPlayableCharacter {
		passion := g.playableCharacter.RomanticInterest[a.Name]
		if passion > 40 {
			textRenderer.DrawTextAt(screen, "❤️", iconX, iconY, color.RGBA{255, 100, 100, 255}, 14)
			iconX += 12
		}
	}

	if !isPlayableCharacter && a.ThoughtTimer > 0 && a.LastAIReasoning != "" {
		reason := a.LastAIReasoning
		// Cap the text length for the overlay
		if len(reason) > 100 { reason = reason[:97] + "..." }
		
		maxWidth := 140
		textSize := 10.0
		tw, _ := textRenderer.MeasureText(reason, textSize)
		boxW := int(tw) + 12
		if boxW > maxWidth { boxW = maxWidth }
		
		tx := int(isoX + offsetX) - boxW/2
		ty := int(by) - 45
		
		// Draw background box with alpha
		vectorRenderer.DrawFilledRect(screen, float32(tx), float32(ty), float32(boxW), 32, color.RGBA{0, 0, 0, 160}, false)
		// Draw yellow border
		borderPts := []engine.Point{
			{X: float64(tx), Y: float64(ty)},
			{X: float64(tx+boxW), Y: float64(ty)},
			{X: float64(tx+boxW), Y: float64(ty+32)},
			{X: float64(tx), Y: float64(ty+32)},
			{X: float64(tx), Y: float64(ty)},
		}
		vectorRenderer.DrawPolygon(screen, borderPts, color.RGBA{218, 165, 32, 255}, 1)
		
		// Draw reasoning text wrapped
		words := strings.Fields(reason)
		line := ""
		currY := ty + 12
		for _, w := range words {
			wWidth, _ := textRenderer.MeasureText(line+w+" ", textSize)
			if int(wWidth) > boxW - 10 {
				if currY < ty + 28 {
					textRenderer.DrawTextAt(screen, line, tx+6, currY, color.White, textSize)
				}
				line = w + " "
				currY += 12
			} else {
				line += w + " "
			}
		}
		if currY < ty + 28 {
			textRenderer.DrawTextAt(screen, line, tx+6, currY, color.White, textSize)
		}
	}

	traumaY := iconY + 15
	traumaX := int(isoX + offsetX - barWidth/2)
	
	adultMode := true
	if g.settings != nil { adultMode = g.settings.AdultMode }
	
	if adultMode {
		if a.Trauma.BurnedAlive { textRenderer.DrawTextAt(screen, "🔥", traumaX, traumaY, color.White, 12); traumaX += 15 }
		if a.Trauma.LeftArmLost || a.Trauma.RightArmLost { textRenderer.DrawTextAt(screen, "🦾", traumaX, traumaY, color.White, 12); traumaX += 15 }
		if a.Trauma.LeftLegLost || a.Trauma.RightLegLost { textRenderer.DrawTextAt(screen, "🦿", traumaX, traumaY, color.White, 12); traumaX += 15 }
		if a.Trauma.EyesLost > 0 { textRenderer.DrawTextAt(screen, "🕶", traumaX, traumaY, color.White, 12); traumaX += 15 }
		if a.Trauma.SpineBroken { textRenderer.DrawTextAt(screen, "♿", traumaX, traumaY, color.White, 12); traumaX += 15 }
	}
}
