package game

import (
	"math"
	"oinakos/internal/engine"
)

func DrawActorGetSprite(a *Actor, adultMode bool) engine.Image {
	if a.Config == nil { return nil }
	conf, mod := a.Config, (*ModelConfig)(nil)
	if a.SelectedModel != "" && conf.Models != nil { mod = conf.Models[a.SelectedModel] }
	var dS engine.Image
	if a.Facing == DirNE || a.Facing == DirNW {
		if mod != nil && mod.BackImage != nil { dS = mod.BackImage } else if conf.BackImage != nil { dS = conf.BackImage } else if mod != nil && mod.StaticImage != nil { dS = mod.StaticImage } else { dS = conf.StaticImage }
	} else {
		if a.IsPregnant && a.GestationTicks < 43200 { if mod != nil && mod.PregnantImage != nil { dS = mod.PregnantImage } else if conf.PregnantImage != nil { dS = conf.PregnantImage } }
		if dS == nil { if mod != nil && mod.StaticImage != nil { dS = mod.StaticImage } else { dS = conf.StaticImage } }
	}
	if a.ActionState == ActorDead || a.ActionState == ActorIncapacitated { if mod != nil && mod.CorpseImage != nil { return mod.CorpseImage }; return conf.CorpseImage }
	if a.HitTimer > 0 { if mod != nil && mod.HitImage != nil { return mod.HitImage }; if img := conf.PickHitImage(a.Tick / 15); img != nil { return img } }
	if a.ActionState == ActorAttacking { if mod != nil && mod.AttackImage != nil { return mod.AttackImage }; if img := conf.PickAttackImage(a.Tick / 30); img != nil { return img } }
	if a.ActionState == ActorChopping { if conf.ChoppingImage != nil { return conf.ChoppingImage }; if mod != nil && mod.AttackImage != nil { return mod.AttackImage }; if img := conf.PickAttackImage(a.Tick / 30); img != nil { return img } }
	if a.ActionState == ActorDigging { if conf.DiggingImage != nil { return conf.DiggingImage }; if mod != nil && mod.AttackImage != nil { return mod.AttackImage }; if img := conf.PickAttackImage(a.Tick / 30); img != nil { return img } }
	if a.ActionState == ActorResting {
		if adultMode {
			if mod != nil && mod.RestingAdultImage != nil { return mod.RestingAdultImage }
			if conf.RestingAdultImage != nil { return conf.RestingAdultImage }
		}
		if mod != nil && mod.RestingImage != nil { return mod.RestingImage }
		if img := conf.RestingImage; img != nil { return img }
		if mod != nil && mod.CrouchImage != nil { return mod.CrouchImage }
		if img := conf.CrouchImage; img != nil { return img }
		if mod != nil && mod.StaticImage != nil { return mod.StaticImage }
		return conf.StaticImage
	}
	if a.ActionState == ActorCrouching { if mod != nil && mod.CrouchImage != nil { return mod.CrouchImage }; if img := conf.CrouchImage; img != nil { return img }; if mod != nil && mod.StaticImage != nil { return mod.StaticImage }; return conf.StaticImage }
	if a.ActionState == ActorCooking { if mod != nil && mod.CookingImage != nil { return mod.CookingImage }; if conf.CookingImage != nil { return conf.CookingImage }; if mod != nil && mod.StaticImage != nil { return mod.StaticImage }; return conf.StaticImage }
	return dS
}

func DrawActorGetOptions(a *Actor, offsetX, offsetY float64, isPC bool, adultMode bool) *engine.DrawImageOptions {
	isoX, isoY := engine.CartesianToIsoZ(a.X, a.Y, a.Z); dS := DrawActorGetSprite(a, adultMode); if dS == nil { return engine.NewDrawImageOptions() }
	w, h := dS.Size(); flip, op := 1.0, engine.NewDrawImageOptions()
	if a.Facing == DirSE || a.Facing == DirNE { flip = -1.0 }; op.Scale(flip, 1.0)
	tx, ty := isoX+offsetX, isoY+offsetY-float64(h)*0.85; if flip < 0 { tx += float64(w) / 2 } else { tx -= float64(w) / 2 }
	if a.ActionState == ActorDead { ty = isoY + offsetY - float64(h)*0.5 } else if a.ActionState == ActorWalking { bobS, bobF := 2.0, 0.2; if isPC { bobS, bobF = 3.0, 0.3 }; ty += math.Sin(float64(a.Tick)*bobF) * bobS } else if a.ActionState == ActorAttacking || a.ActionState == ActorChopping || a.ActionState == ActorDigging {
		lAmt, tick := 0.0, a.Tick%30; if tick < 15 { lAmt = (float64(tick)/15.0)*5.0 } else { lAmt = (float64(30-tick)/15.0)*5.0 }; if flip < 0 { tx += lAmt } else { tx -= lAmt }
	}
	op.Translate(tx, ty); return op
}

func DrawActor(a *Actor, screen engine.Image, textRenderer engine.TextRenderer, vectorRenderer engine.VectorRenderer, paletteShader engine.Shader, offsetX, offsetY float64, isPlayableCharacter bool, adultMode bool) {
	if screen == nil || a.Config == nil { return }
	dS := DrawActorGetSprite(a, adultMode); if dS == nil { return }; op := DrawActorGetOptions(a, offsetX, offsetY, isPlayableCharacter, adultMode)
	if !a.IsOccluded { DrawAlignmentIndicator(screen, vectorRenderer, a, offsetX, offsetY, false) }
	hasP, hasT := a.Config.PrimaryColor != "" || a.Config.SecondaryColor != "", a.Trauma != (PhysicalTrauma{})
	if (hasP || hasT) && paletteShader != nil {
		pA, sA := HexToRGBA(a.Config.PrimaryColor), HexToRGBA(a.Config.SecondaryColor)
		uniforms := map[string]any{"PrimaryColor": pA[:], "SecondaryColor": sA[:], "LeftArmLost": toF(a.Trauma.LeftArmLost), "RightArmLost": toF(a.Trauma.RightArmLost), "LeftLegLost": toF(a.Trauma.LeftLegLost), "RightLegLost": toF(a.Trauma.RightLegLost), "BurnedAlive": toF(a.Trauma.BurnedAlive), "EyesLost": float32(a.Trauma.EyesLost), "StatusTint": []float32{0, 0, 0, 0}}
		if g, ok := vectorRenderer.(engine.Graphics); ok { g.DrawImageWithShader(screen, dS, paletteShader, uniforms, op) } else { screen.DrawImage(dS, op) }
	} else { screen.DrawImage(dS, op) }
}
func toF(b bool) float32 { if b { return 1.0 }; return 0.0 }
