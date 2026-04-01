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
	// Calibrated for ~2.5 meals per day (0.005 * 17280 = 86 units total decay/day)
	a.State.Hunger += 0.005 * decayMultiplier * weatherPenalty * pMult
	a.State.Thirst += 0.03 * decayMultiplier * weatherPenalty * pMult
	
	fMult := 1.0; if a.IsPregnant { fMult = 1.5 }
	a.State.Fatigue += 0.01 * decayMultiplier * weatherPenalty * fMult

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
		recoveryRate := 0.05; isComfy := false
		if ctx != nil && ctx.World != nil {
			for _, o := range ctx.World.Obstacles {
				id := strings.ToLower(o.ID)
				if o.Alive && (strings.Contains(id, "tavern") || strings.Contains(id, "house") || strings.Contains(id, "farm") || strings.Contains(id, "campfire")) {
					dist := math.Sqrt(math.Pow(a.X-o.X, 2) + math.Pow(a.Y-o.Y, 2))
					// Commercial Rest: Owner OR Guest OR Public Campfire
					isLodged := a.LodgingTicks > 0 || strings.Contains(id, "campfire") || a.OwnedChestID != ""
					if dist < 8.0 && isLodged { recoveryRate, isComfy = 0.25, true; break }
				}
			}
		}
		a.State.Fatigue -= recoveryRate
		if a.State.Fatigue <= 0 { a.State.Fatigue, a.ActionState = 0, ActorIdle }
		if a.Tick%60 == 0 {
			healthFactor := 0.20; if isComfy { healthFactor = 0.60 }
			if a.State.HealthPoints < a.GetTotalMaxHealth() {
				regen := int(float64(a.GetTotalMaxHealth()) * healthFactor / 33.0)
				if regen < 1 { regen = 1 }; a.Heal(regen)
			}
		}
	} else if a.ActionState == ActorDrinking {
		a.State.Thirst -= 2.0; a.State.Sanity += 0.05
		if a.State.Thirst <= 0 { a.State.Thirst, a.ActionState = 0, ActorIdle }
		if a.Tick%60 == 0 { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Refreshing...", X: a.X, Y: a.Y, Life: 45, Color: ColorHeal }) }
	} else if a.ActionState == ActorEating {
		a.State.Hunger -= 2.0
		if a.State.Hunger < 0 { a.State.Hunger = 0 }
	} else if a.ActionState == ActorBathing {
		a.State.Hygiene += 2.0
		if a.State.Hygiene >= 100 { a.State.Hygiene, a.ActionState = 100, ActorIdle }
	} else if a.ActionState == ActorAttacking || a.ActionState == ActorChopping || a.ActionState == ActorDigging {
		a.State.Hunger += 0.02; a.State.Thirst += 0.03; a.State.Fatigue += 0.08
	} else if a.ActionState == ActorWalking {
		a.State.Hunger += 0.002; a.State.Thirst += 0.004; a.State.Fatigue += 0.01
	}

	if a.State.Hunger >= 100 || a.State.Thirst >= 100 || a.State.Fatigue >= 100 {
		if a.Tick%TicksPerHour == 0 && a.IsAlive() { 
			a.State.HealthPoints -= 1 
		}
	} else if a.ActionState == ActorIncapacitated {
		if a.Tick%TicksPerHour == 0 && a.IsAlive() {
			a.State.HealthPoints -= 1
		}
	}

	a.SyncLifeStatus()
}

func (a *Actor) AlleviateProperly(ctx *SystemContext) { a.State.BladderLevel, a.State.BowelLevel, a.State.Pain = 0, 0, 0 }
func (a *Actor) TakeBath(ctx *SystemContext) { a.ActionState, a.Tick = ActorBathing, 0 }
