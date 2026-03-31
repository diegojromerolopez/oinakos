package game

import (
	"fmt"
	"math"
	"math/rand"
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

	weatherPenalty := 1.0
	if ctx != nil && ctx.World != nil && ctx.World.State.Weather == WeatherRain {
		weatherPenalty += (ctx.World.State.Intensity * 0.5)
	}

	pMult := 1.0; if a.IsPregnant { pMult = 1.25 }
	a.State.Hunger += 0.02 * decayMultiplier * weatherPenalty * pMult
	a.State.Thirst += 0.03 * decayMultiplier * weatherPenalty * pMult
	
	fMult := 1.0; if a.IsPregnant { fMult = 1.5 }
	a.State.Fatigue += 0.01 * decayMultiplier * weatherPenalty * fMult

	a.State.BladderLevel += 0.015
	a.State.BowelLevel += 0.01

	if a.State.BladderLevel > 80 { a.State.Pain += 0.0001 }
	if a.State.BowelLevel > 80 { a.State.Pain += 0.00005 }
	if a.State.BladderLevel >= 100 || a.State.BowelLevel >= 100 { a.State.Pain += 0.001 }

	a.State.Hygiene -= 0.005

	if a.State.Hunger > 100 { a.State.Hunger = 100 }
	if a.State.Thirst > 100 { a.State.Thirst = 100 }
	if a.State.Fatigue > 100 { a.State.Fatigue = 100 }
	if a.State.BladderLevel > 100 { a.State.BladderLevel = 100 }
	if a.State.BowelLevel > 100 { a.State.BowelLevel = 100 }
	if a.State.Hygiene < 0 { a.State.Hygiene = 0 }
	
	if a.State.AlcoholLevel > 0 {
		a.State.AlcoholLevel -= 0.0028
		if a.State.AlcoholLevel <= 0 {
			a.State.AlcoholLevel = 0
			if a.State.IsDrunk {
				a.State.IsDrunk = false
				if ctx != nil && ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] has sobered up.", a.Name), LogInfo) }
			}
		}
	}

	if ctx != nil && ctx.World != nil {
		for _, o := range ctx.World.Obstacles {
			if o.Alive && o.Archetype != nil && o.Archetype.IsHazard {
				dist := math.Sqrt(math.Pow(a.X-o.X, 2) + math.Pow(a.Y-o.Y, 2))
				if dist < 0.8 {
					for _, action := range o.Archetype.Actions { if action.Type == ActionSoiling { a.State.Hygiene -= float64(action.Amount) * 0.01 } }
				}
			}
		}
	}

	if a.State.Hygiene <= 0 && a.Tick%TicksPerDay == 0 && a.IsAlive() {
		if rand.Float64() < 0.2 { a.State.IsSick = true }
	}
	if a.State.IsSick && a.Tick%TicksPerMonth == 0 && a.IsAlive() { 
		a.State.HealthPoints -= 1
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Sick (Sepsis)", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm })
		}
	}

	if ctx != nil && ctx.World != nil && ctx.World.CurrentMapType != nil {
		groundZ := ctx.World.CurrentMapType.GetElevationAt(a.X, a.Y)
		a.VerticalVelocity -= 0.05 
		a.Z += a.VerticalVelocity
		if a.Z < groundZ { a.Z, a.VerticalVelocity = groundZ, 0 }
	}

	if a.RegenPerSecond > 0 && a.State.HealthPoints < a.GetTotalMaxHealth() {
		if a.Tick%60 == 0 { a.Heal(a.RegenPerSecond) }
	}

	if a.Trauma.BurnedAlive && a.Tick%600 == 0 { a.State.HealthPoints -= 1 }
	bleedingRate := 0
	if a.Trauma.LeftArmLost { bleedingRate++ }
	if a.Trauma.RightArmLost { bleedingRate++ }
	if a.Trauma.LeftLegLost { bleedingRate++ }
	if a.Trauma.RightLegLost { bleedingRate++ }
	if bleedingRate > 0 && a.Tick%300 == 0 {
		a.State.HealthPoints -= bleedingRate
		if ctx != nil && ctx.World != nil && ctx.World.PlayableCharacter != nil && a.Name == ctx.World.PlayableCharacter.Name { ctx.Log("Bleeding heavily...", LogCombatDamage) }
	}
	
	if a.ActionState == ActorIncapacitated && a.Tick % TicksPerHour == 0 { a.State.HealthPoints -= 1 }

	if a.ActionState == ActorBathing {
		a.State.Hygiene += 2.0; if a.State.Hygiene >= 100 { a.State.Hygiene, a.ActionState = 100, ActorIdle }
	}
	
	if a.ActionState == ActorRelieving {
		a.State.BladderLevel -= 5.0; a.State.BowelLevel -= 5.0
		if (a.State.BladderLevel <= 0 && a.State.BowelLevel <= 0) || a.IsTrulyDead() {
			a.AlleviateProperly(ctx)
			if a.ActionState != ActorDead { a.ActionState = ActorIdle }
		}
	}

	if a.ActionState == ActorResting {
		recoveryRate := 0.05; isComfy := false
		if ctx != nil && ctx.World != nil {
			for _, o := range ctx.World.Obstacles {
				id := strings.ToLower(o.ID)
				if o.Alive && (strings.Contains(id, "tavern") || strings.Contains(id, "house") || strings.Contains(id, "farm") || strings.Contains(id, "campfire")) {
					dist := math.Sqrt(math.Pow(a.X-o.X, 2) + math.Pow(a.Y-o.Y, 2))
					if dist < 8.0 { recoveryRate, isComfy = 0.25, true; break }
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
	} else if a.ActionState == ActorAttacking || a.ActionState == ActorChopping || a.ActionState == ActorDigging {
		a.State.Hunger += 0.02; a.State.Thirst += 0.03; a.State.Fatigue += 0.08; a.State.Hygiene -= 0.01
	} else if a.ActionState == ActorWalking {
		a.State.Hunger += 0.002; a.State.Thirst += 0.004; a.State.Fatigue += 0.01; a.State.Hygiene -= 0.0005
	}

	isCritical := a.State.Hunger >= 100 || a.State.Thirst >= 100 || a.State.Fatigue >= 100
	if isCritical && a.Tick%TicksPerHour == 0 && a.IsAlive() {
		a.State.HealthPoints -= 1
		if ctx != nil && ctx.World != nil {
			msg := "-Exhausted-"
			if a.State.Hunger >= 100 { msg = "-Starving-" }
			if a.State.Thirst >= 100 { msg = "-Dehydrated-" }
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: msg, X: a.X, Y: a.Y, Life: 45, Color: ColorHarm })
		}
	}

	if (a.State.BladderLevel > 80 || a.State.BowelLevel > 80) && a.Tick%(TicksPerSecond*5) == 0 { a.CausePain(5.0, ctx) }
	if (a.State.BladderLevel >= 100 || a.State.BowelLevel >= 100) && a.Tick%(TicksPerSecond*10) == 0 {
		a.State.HealthPoints -= 1; a.State.IsSick = true; a.AlleviateOnSelf(ctx) 
		if ctx != nil && ctx.World != nil { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Pants Soiled!", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm }) }
	}

	a.SyncLifeStatus()
}

func (a *Actor) AlleviateProperly(ctx *SystemContext) { a.State.BladderLevel, a.State.BowelLevel, a.State.Pain = 0, 0, 0 }
func (a *Actor) AlleviateOnSelf(ctx *SystemContext) {
	a.State.BladderLevel, a.State.BowelLevel = 0, 0
	a.State.Hygiene -= 50.0; if a.State.Hygiene < 0 { a.State.Hygiene = 0 }; a.State.Pain = 0
}

func (c *Character) TakeBath(ctx *SystemContext) { c.ActionState, c.Tick = ActorBathing, 0 }

func (c *Character) TransferSoilingToVictims(ctx *SystemContext) {
	if c.State.Hygiene > 30 { return }
	for _, other := range ctx.World.Characters {
		if other == c { continue }
		dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
		if dist < 0.8 { other.State.Hygiene -= 10.0; if other.State.Hygiene < 0 { other.State.Hygiene = 0 } }
	}
}
