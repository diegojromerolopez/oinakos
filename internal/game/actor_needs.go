package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strings"
)

func (a *Actor) SharedUpdate(ctx *SystemContext) {
	if a.ThoughtTimer > 0 { a.ThoughtTimer-- }
	if a.UnconsciousTimer > 0 {
		a.UnconsciousTimer--
		a.SyncLifeStatus()
	}
	a.updateIllness(ctx)
	if !a.IsAlive() { return }

	a.updateNeeds(ctx)
	a.updateSanity(ctx)
	a.updateHusbandry(ctx)
	a.updateOwnership(ctx)
	a.updateMood(ctx)
	a.updateBreeding(ctx)
	a.updateMaintenance(ctx)
	a.updateArousal(ctx)
	a.updatePain(ctx)
	a.updateAge(ctx)

	// Psychotic Break (Psychosis)
	if a.TemporalState.Sanity <= 0 && a.IsAlive() && a.State != ActorIncapacitated {
		if a.State != ActorBerserk {
			a.State = ActorBerserk
			a.Alignment = AlignmentEnemy // Breakdowns make you hostile
			DebugLog("Actor [%s] has suffered a Psychotic Break!", a.Name)
		}
	} else if a.State == ActorBerserk && a.TemporalState.Sanity > 20 {
		a.State = ActorIdle // Recover from break
	}

	// Sanity Logic: work/leisure balance
	isWorking := a.State == ActorAttacking || a.State == ActorChopping || a.State == ActorDigging || a.State == ActorForaging
	isLeisure := a.State == ActorIdle || a.State == ActorResting || a.State == ActorCrouching || a.State == ActorDrinking || a.State == ActorEating

	if isWorking {
		a.WorkTicks++
		a.LeisureTicks = 0
		if a.WorkTicks < 1800 { a.TemporalState.Sanity += 0.001 } // Satisfaction from short work
		if a.WorkTicks > 3600 { a.TemporalState.Sanity -= 0.01  } // Burnout from long work
	} else if isLeisure {
		a.LeisureTicks++
		a.WorkTicks = 0
		if a.LeisureTicks < 3600 { a.TemporalState.Sanity += 0.005 } // Leisure improves sanity
		if a.LeisureTicks > 10800 { a.TemporalState.Sanity -= 0.005 } // Too much idleness (ennui) makes you crazy
	}

	if a.TemporalState.Sanity < 0 { a.TemporalState.Sanity = 0 }
	if a.TemporalState.Sanity > 100 { a.TemporalState.Sanity = 100 }

	a.UpdateEffects() // Refresh bonuses from items
}

func (a *Actor) updateMaintenance(ctx *SystemContext) {
	// Passive degradation: items lose 1 resistance every ~4 hours of game time
	if a.Tick%8640 != 0 { return } 

	for _, it := range a.Inventory {
		if it != nil && it.Resistance > 0 && it.Config != nil && it.Config.Resistance > 0 {
			it.Resistance--
		}
	}
	for _, it := range a.Slots {
		if it != nil && it.Resistance > 0 && it.Config != nil && it.Config.Resistance > 0 {
			it.Resistance--
		}
	}
}

func (a *Actor) updateNeeds(ctx *SystemContext) {
	// Percentile scaling: Health 0 -> 1.25x decay, Health 50 -> 0.75x decay, Health 100 -> 0.25x decay
	decayMultiplier := 1.25 - (float64(a.PrimaryAttributes.Health) * 0.01)
	if decayMultiplier < 0.25 { decayMultiplier = 0.25 }

	// Weather modifiers
	weatherPenalty := 1.0
	if ctx != nil && ctx.World != nil && ctx.World.State.Weather == WeatherRain {
		weatherPenalty += (ctx.World.State.Intensity * 0.5)
	}

	// Re-balanced rates for 5.1M ticks/day cycle (60 TPS)
	// Base rates to hit 100 in approx 1 simulated day (5,184,000 ticks)
	a.TemporalState.Hunger += 0.00008 * decayMultiplier * weatherPenalty
	a.TemporalState.Thirst += 0.00012 * decayMultiplier * weatherPenalty 
	a.TemporalState.Fatigue += 0.00004 * decayMultiplier * weatherPenalty

	// Passive excretion buildup (hits 100 in ~24h)
	a.TemporalState.Miccionate += 0.00006
	a.TemporalState.Defecate += 0.00004

	// Retentive Pain: Holding it above 80% causes discomfort and eventual agony
	if a.TemporalState.Miccionate > 80 {
		a.TemporalState.Pain += 0.0001 // Progressive ache
	}
	if a.TemporalState.Defecate > 80 {
		a.TemporalState.Pain += 0.00005
	}
	if a.TemporalState.Miccionate >= 100 || a.TemporalState.Defecate >= 100 {
		a.TemporalState.Pain += 0.001 // Intense cramp
	}

	// Hygiene decay (approx 3-4 days to hit 0)
	a.TemporalState.Hygiene -= 0.00002

	if a.TemporalState.Hunger > 100 { a.TemporalState.Hunger = 100 }
	if a.TemporalState.Thirst > 100 { a.TemporalState.Thirst = 100 }
	if a.TemporalState.Fatigue > 100 { a.TemporalState.Fatigue = 100 }
	if a.TemporalState.Miccionate > 100 { a.TemporalState.Miccionate = 100 }
	if a.TemporalState.Defecate > 100 { a.TemporalState.Defecate = 100 }
	if a.TemporalState.Hygiene < 0 { a.TemporalState.Hygiene = 0 }
	
	// Alcohol decay (approx 1 pint decays in 1 real hour = 3600 ticks)
	// 10 units per pint / 3600 = 0.0028 per tick
	if a.TemporalState.AlcoholLevel > 0 {
		a.TemporalState.AlcoholLevel -= 0.0028
		if a.TemporalState.AlcoholLevel <= 0 {
			a.TemporalState.AlcoholLevel = 0
			if a.TemporalState.IsDrunk {
				a.TemporalState.IsDrunk = false
				if ctx != nil && ctx.Log != nil {
					ctx.Log(fmt.Sprintf("[%s] has sobered up.", a.Name), LogInfo)
				}
			}
		}
	}

	// Environmental Soiling Hazards (Stepping on defecation, mud, etc)
	if ctx != nil && ctx.World != nil {
		for _, o := range ctx.World.Obstacles {
			if o.Alive && o.Archetype != nil && o.Archetype.IsHazard {
				dist := math.Sqrt(math.Pow(a.X-o.X, 2) + math.Pow(a.Y-o.Y, 2))
				if dist < 0.8 { // Proximity to hazard
					for _, action := range o.Archetype.Actions {
						if action.Type == ActionSoiling {
							a.TemporalState.Hygiene -= float64(action.Amount) * 0.01 // Slow progressive soiling
						}
					}
				}
			}
		}
	}

	// Hygiene Sickness logic: once hygiene hits 0, character risks getting sick
	if a.TemporalState.Hygiene <= 0 && a.Tick%86400 == 0 && a.IsAlive() { // Check once per day
		if rand.Float64() < 0.2 { // 20% daily chance
			a.TemporalState.IsSick = true
		}
	}
	// Weekly HP loss from extreme sickness (per user requirement: 0 hygiene leads to health loss)
	if a.TemporalState.IsSick && a.Tick%604800 == 0 && a.IsAlive() { 
		a.TemporalState.HealthPoints -= 1
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Sick (Sepsis)", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm })
		}
	}

	if ctx != nil && ctx.World != nil && ctx.World.CurrentMapType != nil {
		groundZ := ctx.World.CurrentMapType.GetElevationAt(a.X, a.Y)
		a.VerticalVelocity -= 0.05 // Gravity
		a.Z += a.VerticalVelocity
		if a.Z < groundZ {
			a.Z = groundZ
			a.VerticalVelocity = 0
		}
	}

		// Regeneration (1 unit per second = 1 unit every 60 ticks)
	if a.RegenPerSecond > 0 && a.TemporalState.HealthPoints < a.GetTotalMaxHealth() {
		if a.Tick%60 == 0 {
			a.Heal(a.RegenPerSecond)
		}
	}

		// Trauma: Continuous Pain from BurnedAlive (-1 HP every 600 ticks = 10s)
	if a.Trauma.BurnedAlive && a.Tick%600 == 0 {
		a.TemporalState.HealthPoints -= 1
	}

	if a.BodyStatus == nil { a.InitBodyStatus() }

	// Biological Needs Logic
	if a.State == ActorDrinking {
		a.TemporalState.Thirst -= 1.0 // Satiate thirst via drinking
		a.TemporalState.Miccionate += 0.5 // Miccionate increases when drinking
		if a.TemporalState.Thirst < 0 { a.TemporalState.Thirst = 0 }
		if a.TemporalState.Miccionate > 100 { a.TemporalState.Miccionate = 100 }
	}
	
	if a.State == ActorEating {
		a.TemporalState.Hunger -= 1.0 // Satiate hunger via eating
		a.TemporalState.Defecate += 0.5 // Defecate increases when eating
		if a.TemporalState.Hunger < 0 { a.TemporalState.Hunger = 0 }
		if a.TemporalState.Defecate > 100 { a.TemporalState.Defecate = 100 }
	}

	if a.State == ActorBathing {
		a.TemporalState.Hygiene += 2.0 // Fast increase when bathing
		if a.TemporalState.Hygiene >= 100 {
			a.TemporalState.Hygiene = 100
			a.State = ActorIdle
		}
	}
	
	if a.State == ActorRelieving {
		a.TemporalState.Miccionate -= 5.0
		a.TemporalState.Defecate -= 5.0
		if (a.TemporalState.Miccionate <= 0 && a.TemporalState.Defecate <= 0) || a.IsTrulyDead() {
			a.AlleviateProperly(ctx)
			if a.State != ActorDead { a.State = ActorIdle }
		}
	}

	if a.State == ActorResting {
		recoveryRate := 0.05
		isComfy := false
		if ctx != nil && ctx.World != nil {
			for _, o := range ctx.World.Obstacles {
				id := strings.ToLower(o.ID)
				if o.Alive && (strings.Contains(id, "tavern") || strings.Contains(id, "house") || strings.Contains(id, "farm") || strings.Contains(id, "campfire")) {
					dist := math.Sqrt(math.Pow(a.X-o.X, 2) + math.Pow(a.Y-o.Y, 2))
					if dist < 8.0 {
						recoveryRate = 0.25
						isComfy = true
						break
					}
				}
			}
		}
		a.TemporalState.Fatigue -= recoveryRate
		if a.TemporalState.Fatigue <= 0 {
			a.TemporalState.Fatigue = 0
			a.State = ActorIdle // Wake up fully rested
		}
		if a.Tick%60 == 0 {
			healthFactor := 0.20
			if isComfy { healthFactor = 0.60 }
			if a.TemporalState.HealthPoints < a.GetTotalMaxHealth() {
				regen := int(float64(a.GetTotalMaxHealth()) * healthFactor / 33.0)
				if regen < 1 { regen = 1 }
				a.Heal(regen)
			}
			if ctx != nil && ctx.World != nil && (a.Tick%120 == 0) {
				msg := "Zzz"
				if isComfy { msg = "Zzz (comfy rest)" }
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
					Text: msg, X: a.X, Y: a.Y, Life: 60, Color: ColorHeal,
				})
			}
		}
	} else if a.State == ActorAttacking || a.State == ActorChopping || a.State == ActorDigging {
		a.TemporalState.Hunger += 0.02
		a.TemporalState.Thirst += 0.03
		a.TemporalState.Fatigue += 0.08
		a.TemporalState.Hygiene -= 0.01 // Work makes you sweaty/dirty
	} else if a.State == ActorWalking {
		a.TemporalState.Hunger += 0.002
		a.TemporalState.Thirst += 0.004
		a.TemporalState.Fatigue += 0.01
		a.TemporalState.Hygiene -= 0.0005
	}

	// Thermodynamics
	if ctx != nil && ctx.World != nil {
		envTemp := ctx.World.State.Temperature
		if a.BodyTemperature == 0 { a.BodyTemperature = 36.5 }
		diff := envTemp - a.BodyTemperature
		a.BodyTemperature += diff * 0.0001

		if a.BodyTemperature > 38.5 { // Overheating
			a.TemporalState.Thirst += 0.005 // Sweat
			if a.Tick%300 == 0 {
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
					Text: "Too Hot!", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm,
				})
			}
		} else if a.BodyTemperature < 34.5 { // Freezing
			a.TemporalState.Hunger += 0.005 // Caloric expenditure
			a.TemporalState.Fatigue += 0.002
			if a.Tick%300 == 0 {
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
					Text: "Too Cold!", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm,
				})
			}
		}

				// Lethal Homeostasis Thresholds
		if a.BodyTemperature > 41.5 { // Deadly heat stroke
			if a.Tick%3600 == 0 {
				a.TemporalState.HealthPoints -= 1
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Heatstroke!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm })
			}
		} else if a.BodyTemperature < 29.5 { // Lethal hypothermia
			if a.Tick%3600 == 0 {
				a.TemporalState.HealthPoints -= 1
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Severe Hypothermia!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm })
			}
		}
	}

		// Critical Penalties (now at 100)
	isCritical := a.TemporalState.Hunger >= 100 || a.TemporalState.Thirst >= 100 || a.TemporalState.Fatigue >= 100
	if isCritical && a.Tick%3600 == 0 && a.IsAlive() {
		a.TemporalState.HealthPoints -= 1
		msg := "-Exhausted-"
		if a.TemporalState.Hunger >= 100 { msg = "-Starving-" }
		if a.TemporalState.Thirst >= 100 { msg = "-Dehydrated-" }
		
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: msg, X: a.X, Y: a.Y, Life: 45, Color: ColorHarm,
			})
		}
	}

	// Alleviation needs logic (Pain and Sickness)
	if (a.TemporalState.Miccionate > 80 || a.TemporalState.Defecate > 80) && a.Tick%300 == 0 {
		a.CausePain(5.0, ctx)
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: "Urgent Need!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm,
			})
		}
	}
	if (a.TemporalState.Miccionate >= 100 || a.TemporalState.Defecate >= 100) && a.Tick%600 == 0 {
		a.TemporalState.HealthPoints -= 1
		a.TemporalState.IsSick = true // Eventually sickness
		a.AlleviateOnSelf(ctx) 
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: "Pants Soiled!", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm,
			})
		}
	}

		// Incapacitated Bleed-out
	if a.IsIncapacitated() && a.Tick%3600 == 0 {
		a.TemporalState.HealthPoints -= 1
	}

	a.SyncLifeStatus()

	if a.CrouchTimer > 0 {
		a.CrouchTimer--
		if a.CrouchTimer == 0 && a.State == ActorCrouching {
			a.State = ActorIdle
		}
	}
}

func (a *Actor) updateSanity(ctx *SystemContext) {
	// Sanity drains from physical misery (now at 90+)
	if a.TemporalState.Hunger > 90 { a.TemporalState.Sanity -= 0.00001 }
	if a.TemporalState.Thirst > 90 { a.TemporalState.Sanity -= 0.00002 }
	if a.TemporalState.Fatigue > 90 { a.TemporalState.Sanity -= 0.00001 }
	if a.FluTicks > 0 { a.TemporalState.Sanity -= 0.00005 }
	
	// Thermal stress affects mood
	if a.BodyTemperature < 32.0 || a.BodyTemperature > 40.0 {
		a.TemporalState.Sanity -= 0.00002
	}

		// Sanity recovers during quality rest or leisure
	if a.State == ActorResting {
		bonus := 0.005
		if a.TemporalState.Hunger < 50 && a.TemporalState.Thirst < 50 { bonus = 0.02 }
		a.TemporalState.Sanity += bonus
	}

	// Hard caps
	if a.TemporalState.Sanity < 0 { a.TemporalState.Sanity = 0 }
	if a.TemporalState.Sanity > 100 { a.TemporalState.Sanity = 100 }
	
	if a.TemporalState.Sanity < 10 && a.Tick%600 == 0 {
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: "Psychological Break!", X: a.X, Y: a.Y, Life: 60, Color: color.RGBA{255, 165, 0, 255},
		})
	}
}

func (a *Actor) updateIllness(ctx *SystemContext) {
	if ctx == nil || ctx.World == nil { return }
	
	// Sepsis HP Drain (Item 4: Wound Infection)
	if a.TemporalState.IsSeptic && a.Tick%600 == 0 {
		a.TemporalState.HealthPoints -= 2
		if ctx.Log != nil && ctx.World.PlayableCharacter != nil && a.Name == ctx.World.PlayableCharacter.Name {
			ctx.Log("The sepsis is worsening...", LogCombatDamage)
		}
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Septic Pain!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm })
	}

	if a.FluTicks > 0 {
		a.FluTicks--
		// Massive Fatigue drain
		a.TemporalState.Fatigue += 0.02
		
		// If fatigue is depleted, start losing health
		// Correction: Fatigue hits 100 when depleted!
		if a.TemporalState.Fatigue > 95 && a.Tick%180 == 0 {
			a.TemporalState.HealthPoints -= 1
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: "-Exhausted-", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm,
			})
		}
		
		// Contagion spread (R coefficient logic)
		if a.ContagionTimer > 0 {
			a.ContagionTimer--
		} else {
			a.ContagionTimer = 600
			radius := 4.0
			for _, other := range ctx.World.Characters {
				if other.Actor.FluTicks > 0 || !other.IsAlive() { continue }
				dist := math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2))
				if dist < radius && rand.Float64() < 0.12 {
					other.Actor.FluTicks = 86400 * 3 // 3 days
					other.Actor.TemporalState.IsSick = true
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has caught the flu from %s!", other.Name, a.Name), LogWarning) }
				}
			}
		}

		if a.FluTicks <= 0 {
			if a.SicknessTicks <= 0 { a.TemporalState.IsSick = false }
		}
	} else {
		// Spontaneous infection chance
		chance := 0.000005 // Default low chance per tick
		if ctx.World.State.Season == SeasonAutumn {
			chance = 0.00005 // 10x higher in autumn
		}
		
		if rand.Float64() < chance {
			a.FluTicks = (3 + rand.Intn(5)) * 17280
			a.TemporalState.IsSick = true
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has developed a fever.", a.Name), LogWarning) }
		}
	}

	if a.SicknessTicks > 0 {
		a.SicknessTicks--
		if a.Sickness == "stomach sickness" {
			// Side effects: Sanity loss and increasing Pain
			a.TemporalState.Sanity -= 0.00005
			a.TemporalState.Pain += 0.0001

			if a.Tick%600 == 0 {
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
					Text: "Stomach Cramps!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm,
				})
			}
		}

		if a.SicknessTicks == 0 {
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has recovered from %s.", a.Name, a.Sickness), LogInfo) }
			a.Sickness = ""
			if a.FluTicks <= 0 { a.TemporalState.IsSick = false }
		}
	}
}

func (a *Actor) updateArousal(ctx *SystemContext) {
	if a.TemporalState.Arousal > 0 {
		a.TemporalState.Arousal -= 0.05 // Natural decay
		if a.TemporalState.Arousal < 0 { a.TemporalState.Arousal = 0 }
	}

	if a.TemporalState.Arousal >= 100 {
		a.ArousalTimer++
		if a.ArousalTimer > 21600 { // 1 hour at 60 TPS
			if a.Tick%600 == 0 {
				a.TemporalState.HealthPoints -= 1
				if ctx != nil && ctx.World != nil {
					ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
						Text: "Strain!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm,
					})
				}
			}
		}
	} else {
		a.ArousalTimer = 0
	}
}

func (a *Actor) updatePain(ctx *SystemContext) {
	if a.TemporalState.Pain > 0 {
		a.TemporalState.Pain -= 0.01 // Pain decays slowly
		if a.TemporalState.Pain < 0 { a.TemporalState.Pain = 0 }
	}

	// Pain effects
	if a.TemporalState.Pain >= 100 {
		if a.State != ActorIncapacitated {
			a.UnconsciousTimer = 600 // 10 seconds of unconsciousness from extreme pain
			a.SyncLifeStatus()
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has collapsed from extreme pain!", a.Name), LogNPC) }
		}
	} else if a.TemporalState.Pain > 80 {
		// Incapacitated - handled by State check in movement logic or here?
		// We'll set a state if they are not already incapacitated
		if a.State != ActorIncapacitated && a.IsAlive() {
			a.State = ActorIncapacitated 
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
	a.TemporalState.Pain += amount
	if a.TemporalState.Pain > 100 { a.TemporalState.Pain = 100 }
	
	if amount >= 3.0 && ctx != nil && rand.Float64() < 0.4 {
		shouts := []string{"Ah!", "Ugh!", "Ouch!", "Gah!", "Pleas no!", "Don't hurt me!", "Stop!", "Ugh"}
		if a.TemporalState.Pain > 70 {
			shouts = []string{"AARGH!", "MAKE IT STOP!", "MERCY!", "PLEASE!", "GAHHHH!"}
		}
		a.Say(shouts[rand.Intn(len(shouts))], ctx)
	}
}

func (a *Actor) updateAge(ctx *SystemContext) {
	if !a.IsAlive() { return }
	a.AgeTicks += a.TemporalState.Age.Rate
	a.TemporalState.Age.Current = a.AgeTicks / float64(TicksPerYear)

	// Max Age Enforcement (0 = immortal)
	if a.TemporalState.Age.Max > 0 && a.TemporalState.Age.Current >= a.TemporalState.Age.Max {
		a.DeadTimer = 0
		a.State = ActorDead
		if ctx != nil && ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has died of extreme old age.", a.Name), LogNPC) }
		return
	}

	nextStage := ""
	ticksPY := float64(TicksPerYear)
	switch a.LifeStage {
	case StageBaby:
		if a.AgeTicks >= 1.0*ticksPY { nextStage = StageKid }
	case StageKid:
		if a.AgeTicks >= 12.0*ticksPY { nextStage = StageTeenager }
	case StageTeenager:
		if a.AgeTicks >= 18.0*ticksPY { nextStage = StageAdult }
	case StageAdult:
		if a.AgeTicks >= 65.0*ticksPY { nextStage = StageElder }
	case StageElder:
		if a.AgeTicks > 85.0*ticksPY { 
			// Check for natural death (increasing chance)
			chance := (a.AgeTicks - 85.0*ticksPY) / (15.0*ticksPY)
			if rand.Float64() < (chance * 0.000001) { // Very low per-tick chance
				a.DeadTimer = 0
				a.State = ActorDead
				if ctx != nil && ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has passed away of old age.", a.Name), LogNPC) }
			}
		}
	}

	if nextStage != "" && ctx != nil && ctx.Registries != nil {
		a.LifeStage = nextStage
		// Find new archetype
		prefix := "archetypes/"
		if a.Config != nil && strings.Contains(a.Config.Archetype, "female") {
			prefix += nextStage + "/female"
		} else {
			prefix += nextStage + "/male"
		}
		
		if newArch, ok := ctx.Registries.Archetypes.Archetypes[prefix]; ok {
			a.Config = newArch
			// Re-roll stats and update visual (selected model)
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has grown into a %s!", a.Name, nextStage), LogNPC) }
			a.SyncStats(ctx.Registries.Objects)
			a.TemporalState.HealthPoints = a.TemporalState.MaxHealthPoints
		}
	}
}
