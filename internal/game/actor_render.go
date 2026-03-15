package game

import (
	"image/color"
	"math"
	"oinakos/internal/engine"
)

// DrawActor is the unified rendering function for any Actor (PlayableCharacter or NPC).
func DrawActor(a *Actor, screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64, isPlayableCharacter bool) {
	if screen == nil || a.Config == nil {
		return
	}

	isoX, isoY := engine.CartesianToIso(a.X, a.Y)

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
		if img := a.Config.CorpseImage; img != nil {
			drawSprite = img
		} else {
			// If no corpse image and dead, we might draw nothing
			if !isPlayableCharacter {
				return
			}
			// Fallback for MC if corpse is missing (should not happen with standard assets)
			if img := a.Config.StaticImage; img != nil {
				drawSprite = img
			}
		}
	} else if a.HitTimer > 0 {
		// Hit animation: toggle between hit frames
		if img := a.Config.PickHitImage(a.Tick / 15); img != nil {
			drawSprite = img
		}
	} else if a.State == ActorAttacking {
		// Attack animation
		cooldown := 30 // MC default
		if a.Weapon != nil && !isPlayableCharacter {
			// NPCs use their archetype's attack cooldown for animation speed
			// Wait, the NPC code used n.AttackCooldown which I didn't see in Actor.
			// Actually, NPC has it via its archetype stats.
			// For unity, let's assume a standard anim speed if not specified.
			cooldown = 30
		}
		if img := a.Config.PickAttackImage(a.Tick / cooldown); img != nil {
			drawSprite = img
		}
	}

	if drawSprite == nil {
		return
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
			// NPC lunge logic from before
			// phase := float64(n.AttackTimer) / float64(n.AttackCooldown)
			// But we don't have AttackTimer/Cooldown in Actor yet.
			// Let's use Tick for consistency if we want it simple.
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

	// Draw Alignment Ellipse
	DrawAlignmentIndicator(screen, vectorRenderer, a.X, a.Y, offsetX, offsetY, a.Alignment, a.IsAlive())

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

	// UI Elements (Health bar for NPCs, Names)
	if a.IsAlive() {
		if !isPlayableCharacter {
			// Health Bar for NPCs
			barWidth := 40.0
			barHeight := 4.0
			bx := isoX + offsetX - barWidth/2
			by := isoY + offsetY - float64(h)*scale*0.9

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
}
