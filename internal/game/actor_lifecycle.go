package game

import (
	"fmt"
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
	if a.State.Sanity <= 0 && a.IsAlive() && a.ActionState != ActorIncapacitated {
		if a.ActionState != ActorBerserk {
			a.ActionState = ActorBerserk
			a.Alignment = AlignmentEnemy // Breakdowns make you hostile
			DebugLog("Actor [%s] has suffered a Psychotic Break!", a.Name)
		}
	} else if a.ActionState == ActorBerserk && a.State.Sanity > 20 {
		a.ActionState = ActorIdle // Recover from break
	}

	// Sanity Logic: work/leisure balance
	isWorking := a.ActionState == ActorAttacking || a.ActionState == ActorChopping || a.ActionState == ActorDigging || a.ActionState == ActorForaging
	isLeisure := a.ActionState == ActorIdle || a.ActionState == ActorResting || a.ActionState == ActorCrouching || a.ActionState == ActorDrinking || a.ActionState == ActorEating

	if isWorking {
		a.WorkTicks++
		a.LeisureTicks = 0
		if a.WorkTicks < 1800 { a.State.Sanity += 0.001 } // Satisfaction from short work
		if a.WorkTicks > 3600 { a.State.Sanity -= 0.01  } // Burnout from long work
	} else if isLeisure {
		a.LeisureTicks++
		a.WorkTicks = 0
		if a.LeisureTicks < 3600 { a.State.Sanity += 0.005 } // Leisure improves sanity
		if a.LeisureTicks > 10800 { a.State.Sanity -= 0.005 } // Too much idleness (ennui) makes you crazy
	}

	a.SyncState()
	a.UpdateEffects() // Refresh bonuses from items
}

func (a *Actor) updateAge(ctx *SystemContext) {
	if !a.IsAlive() { return }
	a.AgeTicks += a.State.Age.Rate
	a.State.Age.Current = a.AgeTicks / float64(TicksPerYear)

	// Max Age Enforcement (0 = immortal)
	if a.State.Age.Max > 0 && a.State.Age.Current >= a.State.Age.Max {
		a.DeadTimer = 0
		a.ActionState = ActorDead
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
				a.ActionState = ActorDead
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
			a.State.HealthPoints = a.State.MaxHealthPoints
		}
	}
}

func (a *Actor) updateIllness(ctx *SystemContext) {
	if ctx == nil || ctx.World == nil { return }
	
	// Sepsis HP Drain (Item 4: Wound Infection)
	if a.State.IsSeptic && a.Tick%600 == 0 {
		a.State.HealthPoints -= 2
		if ctx.Log != nil && ctx.World.PlayableCharacter != nil && a.Name == ctx.World.PlayableCharacter.Name {
			ctx.Log("The sepsis is worsening...", LogCombatDamage)
		}
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Septic Pain!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm })
	}

	if a.FluTicks > 0 {
		a.FluTicks--
		// Massive Fatigue drain
		a.State.Fatigue += 0.02
		
		// If fatigue is depleted, start losing health
		if a.State.Fatigue > 95 && a.Tick%180 == 0 {
			a.State.HealthPoints -= 1
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
					other.Actor.State.IsSick = true
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has caught the flu from %s!", other.Name, a.Name), LogWarning) }
				}
			}
		}

		if a.FluTicks <= 0 {
			if a.SicknessTicks <= 0 { a.State.IsSick = false }
		}
	} else {
		// Spontaneous infection chance
		chance := 0.000005 // Default low chance per tick
		if ctx.World.State.Season == SeasonAutumn {
			chance = 0.000005 * 10 // 10x higher in autumn
		}
		
		if rand.Float64() < chance {
			a.FluTicks = (3 + rand.Intn(5)) * 17280
			a.State.IsSick = true
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has developed a fever.", a.Name), LogWarning) }
		}
	}

	if a.SicknessTicks > 0 {
		a.SicknessTicks--
		if a.Sickness == "stomach sickness" {
			// Side effects: Sanity loss and increasing Pain
			a.State.Sanity -= 0.00005
			a.State.Pain += 0.0001

			if a.Tick%600 == 0 {
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
					Text: "Stomach Cramps!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm,
				})
			}
		}

		if a.SicknessTicks == 0 {
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has recovered from %s.", a.Name, a.Sickness), LogInfo) }
			a.Sickness = ""
			if a.FluTicks <= 0 { a.State.IsSick = false }
		}
	}
}
