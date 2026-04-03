package game

import (
	"math"
	"strings"
)

func (a *Actor) updateMaintenance(ctx *SystemContext) {
	if a.Tick%8640 != 0 { return } 
	for _, it := range a.Inventory { if it != nil && it.Resistance > 0 && it.Config != nil && it.Config.Resistance > 0 { it.Resistance-- } }
	for _, it := range a.Slots { if it != nil && it.Resistance > 0 && it.Config != nil && it.Config.Resistance > 0 { it.Resistance-- } }
}

func (a *Actor) updateNeeds(ctx *SystemContext) {
	decayMultiplier := 1.25 - (float64(a.PrimaryAttributes.Health) * 0.01)
	if decayMultiplier < 0.25 { decayMultiplier = 0.25 }
	if a.LodgingTicks > 0 { a.LodgingTicks-- }

	weatherPenalty := 1.0
	if ctx != nil && ctx.World != nil && ctx.World.State.Weather == WeatherRain {
		weatherPenalty += (ctx.World.State.Intensity * 0.5)
	}

	pMult := 1.0; if a.IsPregnant { pMult = 1.25 }
	// Rule of Threes - 100 units spanning 30 days (Hunger) and 3 days (Thirst)
	// (2 Liters/Day calibration: 100 units = lethality, requires ~0.0023/tick)
	a.State.Hunger += 0.000192 * decayMultiplier * weatherPenalty * pMult
	a.State.Thirst += 0.0023 * decayMultiplier * weatherPenalty * pMult
	
	fMult := 1.0; if a.IsPregnant { fMult = 1.5 }
	
	// Circadian Rhythm: Night (6 PM - 6 AM) vs Day (6 AM - 6 PM)
	isNight := false
	if ctx != nil && ctx.World != nil {
		hour := (ctx.World.DayTick / TicksPerHour) % 24
		if hour < 6 || hour >= 18 { isNight = true }
	}
	
	circadianMult := 0.5; if isNight { circadianMult = 2.0 }
	
	// Sanity Evolution: Loneliness and Cognitive Strain
	// Base decay if alone at night (Medieval fear of isolation)
	isAlone := true
	for _, other := range ctx.World.Characters {
		if other.GetActor() != a && other.IsAlive() {
			dist := math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2))
			if dist < 15.0 { isAlone = false; break }
		}
	}
	if isAlone && isNight { a.State.Sanity -= 0.005 }
	if a.State.Fatigue > 80 { a.State.Sanity -= 0.002 } // Cognitive exhaustion

	// Spiritual Healing (Proximity to Cleric/Monastery)
	if ctx != nil && ctx.World != nil {
		for _, other := range ctx.World.Characters {
			if other.IsAlive() && (strings.Contains(strings.ToLower(other.Config.ID), "cleric") || strings.Contains(strings.ToLower(other.Config.ID), "monk")) {
				dist := math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2))
				if dist < 4.0 { a.State.Sanity += 0.01; break } // Holy presence
			}
		}
	}

	// Low Sanity increases fatigue accumulation (Depression/Lack of will)
	sanityPenalty := 1.0; if a.State.Sanity < 30 { sanityPenalty = 1.2 }

	// Base Stamina: 72 hours awake at Day-rate, but forced to night-respect.
	a.State.Fatigue += 0.00192 * decayMultiplier * weatherPenalty * fMult * circadianMult * sanityPenalty

	// Forced Collapse (Natural Human Limit)
	if a.State.Fatigue >= 100 && a.ActionState != ActorResting {
		a.ActionState, a.Tick = ActorResting, 0
		a.State.Hygiene -= 20 
		a.Say("I... can't... stay... awake...", ctx)
	}

	a.State.BladderLevel += 0.015
	a.State.BowelLevel += 0.01
	a.State.Hygiene -= 0.005

	if a.State.BladderLevel > 80 || a.State.BowelLevel > 80 {
		a.CausePain(0.1, ctx)
	}

	if a.State.BladderLevel >= 100 || a.State.BowelLevel >= 100 {
		a.State.BladderLevel, a.State.BowelLevel = 0, 0
		a.State.Hygiene -= 50
		a.State.Pain = 0
		a.Say("Oh no...", ctx)
	}

	if a.State.Hunger > 100 { a.State.Hunger = 100 }
	if a.State.Thirst > 100 { a.State.Thirst = 100 }
	if a.State.Fatigue > 100 { a.State.Fatigue = 100 }
	if a.State.Hygiene < 0 { a.State.Hygiene = 0 }
	
	if a.State.AlcoholLevel > 0 {
		a.State.AlcoholLevel -= 0.0028
		if a.State.AlcoholLevel <= 0 {
			a.State.AlcoholLevel = 0
			if a.State.IsDrunk { a.State.IsDrunk = false }
		}
	}

	if a.ActionState == ActorResting {
		// Base recovery: 20% per 8 hours (8h = 5,760 ticks). 
		// Rate = 20 / 5,760 = 0.0034 unit/tick. 
		recoveryRate := 0.0034; isComfy := false
		if ctx != nil && ctx.World != nil {
			for _, o := range ctx.World.Obstacles {
				id := strings.ToLower(o.ID)
				if o.Alive && (strings.Contains(id, "tavern") || strings.Contains(id, "house") || strings.Contains(id, "farm") || strings.Contains(id, "campfire")) {
					dist := math.Sqrt(math.Pow(a.X-o.X, 2) + math.Pow(a.Y-o.Y, 2))
					// Comfy rest (Bed/Inn) doubles recovery to 40% per 8h.
					isLodged := a.LodgingTicks > 0 || strings.Contains(id, "campfire") || a.OwnedChestID != ""
					if dist < 8.0 && isLodged { recoveryRate, isComfy = 0.0068, true; break }
				}
			}
		}
		a.State.Fatigue -= recoveryRate
		if a.State.Fatigue <= 0 { a.State.Fatigue, a.ActionState = 0, ActorIdle }
		if a.Tick%60 == 0 {
			healthFactor := 0.20; if isComfy { 
				healthFactor = 0.60 
				a.State.Sanity += 0.1 // A good night's sleep clears the mind
			}
			if a.State.HealthPoints < a.GetTotalMaxHealth() {
				regen := int(float64(a.GetTotalMaxHealth()) * healthFactor / 33.0)
				if regen < 1 { regen = 1 }; a.Heal(regen)
			}
		}
	} else if a.ActionState == ActorDrinking {
		a.State.Thirst -= 2.0; a.State.Sanity += 0.05
		// Last Gasp Heal: Drinking while at death's door gives 1 HP back to break the 'Incapacitated' cycle
		if a.State.HealthPoints <= 0 && a.Tick%30 == 0 { a.State.HealthPoints = 1 }
		if a.State.Thirst <= 0 { a.State.Thirst, a.ActionState = 0, ActorIdle }
		if a.Tick%60 == 0 { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Refreshing...", X: a.X, Y: a.Y, Life: 45, Color: ColorHeal }) }
	} else if a.ActionState == ActorEating {
		a.State.Hunger -= 2.0
		if a.State.Hunger < 0 { a.State.Hunger = 0 }
	} else if a.ActionState == ActorBathing {
		a.State.Hygiene += 2.0
		if a.State.Hygiene >= 100 { a.State.Hygiene, a.ActionState = 100, ActorIdle }
	} else if a.ActionState == ActorAttacking || a.ActionState == ActorChopping || a.ActionState == ActorDigging {
		// Physical Efficiency: High Strength reduces the toll of heavy labor (up to 40% reduction).
		strEff := 1.0 - (float64(a.PrimaryAttributes.Strength) * 0.005); if strEff < 0.6 { strEff = 0.6 }
		a.State.Hunger += 0.01 * strEff; a.State.Thirst += 0.015 * strEff; a.State.Fatigue += 0.01 * strEff
	} else if a.ActionState == ActorWalking {
		// Persistence Hunter: High Strength improves walking gait efficiency.
		strEff := 1.0 - (float64(a.PrimaryAttributes.Strength) * 0.005); if strEff < 0.6 { strEff = 0.6 }
		a.State.Hunger += 0.0005 * strEff; a.State.Thirst += 0.001 * strEff; a.State.Fatigue += 0.00005 * strEff
	}

	if a.State.Hunger >= 100 || a.State.Thirst >= 100 {
		if a.Tick%TicksPerHour == 0 && a.IsAlive() { 
			a.State.HealthPoints -= 1 
		}
	} else if a.ActionState == ActorIncapacitated {
		// Only Wounds or Critical Needs kill; Fatigue simply forces Sleep.
		if a.Tick%TicksPerHour == 0 && a.IsAlive() {
			a.State.HealthPoints -= 1
		}
	}

	a.SyncLifeStatus()
}

func (a *Actor) AlleviateProperly(ctx *SystemContext) { a.State.BladderLevel, a.State.BowelLevel, a.State.Pain = 0, 0, 0 }
func (a *Actor) TakeBath(ctx *SystemContext) { a.ActionState, a.Tick = ActorBathing, 0 }
