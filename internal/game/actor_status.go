package game

import (
	"fmt"
	"image/color"
	"math/rand"
)

func (a *Actor) updateSanity(ctx *SystemContext) {
	// Sanity drains from physical misery (now at 90+)
	if a.State.Hunger > 90 { a.State.Sanity -= 0.00001 }
	if a.State.Thirst > 90 { a.State.Sanity -= 0.00002 }
	if a.State.Fatigue > 90 { a.State.Sanity -= 0.00001 }
	if a.FluTicks > 0 { a.State.Sanity -= 0.00005 }
	
	// Thermal stress affects mood
	if a.BodyTemperature < 32.0 || a.BodyTemperature > 40.0 {
		a.State.Sanity -= 0.00002
	}

	// Sanity recovers during quality rest or leisure
	if a.ActionState == ActorResting {
		bonus := 0.005
		if a.State.Hunger < 50 && a.State.Thirst < 50 { bonus = 0.02 }
		a.State.Sanity += bonus
	}

	// Hard caps
	if a.State.Sanity < 0 { a.State.Sanity = 0 }
	if a.State.Sanity > 100 { a.State.Sanity = 100 }
	
	if a.State.Sanity < 10 && a.Tick%600 == 0 && ctx != nil && ctx.World != nil {
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: "Psychological Break!", X: a.X, Y: a.Y, Life: 60, Color: color.RGBA{255, 165, 0, 255},
		})
	}
}

func (a *Actor) updateArousal(ctx *SystemContext) {
	if a.SexualOrientation == "asexual" { a.State.Arousal = 0; return }

	// Monthly arousal cycle: 100 units / 1 Month (518,400 ticks) = 0.000192 / tick
	growth := 0.000192; if ctx != nil && ctx.World != nil && ctx.World.State.Season == SeasonSpring { growth *= 1.5 }
	a.State.Arousal += growth

	if a.State.Arousal >= 100 {
		a.ArousalTimer++
		if a.ArousalTimer > 21600 { // 1 hour at 60 TPS
			if a.Tick%600 == 0 {
				a.CausePain(0.1, ctx) 
				if ctx != nil && ctx.World != nil {
					ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
						Text: "Distracted...", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm,
					})
				}
			}
		}
	} else {
		a.ArousalTimer = 0
	}
}

func (a *Actor) updatePain(ctx *SystemContext) {
	if a.State.Pain > 0 {
		a.State.Pain -= 0.01 // Pain decays slowly
		if a.State.Pain < 0 { a.State.Pain = 0 }
	}

	// Pain effects
	if a.State.Pain >= 100 {
		if a.ActionState != ActorIncapacitated {
			a.UnconsciousTimer = 600 // 10 seconds of unconsciousness from extreme pain
			a.SyncLifeStatus()
			if ctx != nil && ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has collapsed from extreme pain!", a.Name), LogNPC) }
		}
	} else if a.State.Pain > 80 {
		if a.ActionState != ActorIncapacitated && a.IsAlive() {
			a.ActionState = ActorIncapacitated 
		}
	}
}

func (a *Actor) Say(text string, ctx *SystemContext) {
	if ctx == nil { return }
	a.LastReaction = text
	a.ThoughtTimer = 180 // Show reflection for 3 seconds
	
	// Add floating text over head
	if ctx.World != nil {
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text:  text,
			X:     a.X,
			Y:     a.Y - 1.2, // Above head
			Life:  180,
			Color: color.White,
		})
	}
	
	// Log to event log if it's significant
	if ctx.Log != nil {
		ctx.Log(fmt.Sprintf("%s: \"%s\"", a.Name, text), LogNPC)
	}
}

func (a *Actor) CausePain(amount float64, ctx *SystemContext) {
	a.State.Pain += amount
	if a.State.Pain > 100 { a.State.Pain = 100 }
	
	if amount >= 3.0 && ctx != nil && rand.Float64() < 0.4 {
		shouts := []string{"Ah!", "Ugh!", "Ouch!", "Gah!", "Pleas no!", "Don't hurt me!", "Stop!", "Ugh"}
		if a.State.Pain > 70 {
			shouts = []string{"AARGH!", "MAKE IT STOP!", "MERCY!", "PLEASE!", "GAHHHH!"}
		}
		a.Say(shouts[rand.Intn(len(shouts))], ctx)
	}
}
