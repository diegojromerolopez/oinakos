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
	if a.CrouchTimer > 0 {
		a.CrouchTimer--
		if a.CrouchTimer == 0 && a.ActionState == ActorCrouching {
			a.ActionState = ActorIdle
		}
	}
	a.updateIllness(ctx)
	
	if a.ActionState == ActorDead {
		a.updateDecay(ctx)
		return 
	}

	a.updateNeeds(ctx)
	a.updateSanity(ctx)
	a.updateHusbandry(ctx)
	a.updateOwnership(ctx)
	a.updateMood(ctx)
	a.updateBreeding(ctx)
	a.updateMaintenance(ctx)
	a.updateArousal(ctx)
	a.updatePain(ctx)
	a.updateScale(ctx)
	a.updateAge(ctx)
	a.updateGrief(ctx)

	// Psychotic Break (Psychosis)
	if a.State.Sanity <= 0 && a.IsAlive() && a.ActionState != ActorIncapacitated {
		if a.ActionState != ActorBerserk {
			a.ActionState = ActorBerserk
			a.Alignment = AlignmentEnemy 
			DebugLog("Actor [%s] has suffered a Psychotic Break!", a.Name)
		}
	} else if a.ActionState == ActorBerserk && a.State.Sanity > 20 {
		a.ActionState = ActorIdle 
	}

	// Sanity Logic: work/leisure balance
	isWorking := a.ActionState == ActorAttacking || a.ActionState == ActorChopping || a.ActionState == ActorDigging || a.ActionState == ActorForaging
	isLeisure := a.ActionState == ActorIdle || a.ActionState == ActorResting || a.ActionState == ActorCrouching || a.ActionState == ActorDrinking || a.ActionState == ActorEating

	if isWorking {
		a.WorkTicks++
		a.LeisureTicks = 0
		if a.WorkTicks < 1800 { a.State.Sanity += 0.001 } 
		if a.WorkTicks > 3600 { a.State.Sanity -= 0.01  } 
	} else if isLeisure {
		a.LeisureTicks++
		a.WorkTicks = 0
		if a.LeisureTicks < 3600 { a.State.Sanity += 0.005 } 
		if a.LeisureTicks > 10800 { a.State.Sanity -= 0.005 } 

		// DRUNK BEHAVIORAL SHIFTS
		if a.State.IsDrunk && a.Tick%TicksPerHour == 0 {
			r := rand.Float64()
			if r < 0.10 { // Aggressive
				a.ActionState = ActorBerserk
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has become a mean drunk!", a.Name), LogWarning) }
			} else if r < 0.30 { // Submissive/Weepy
				a.State.Sanity -= 5.0
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s is feeling weepy and vulnerable while drunk.", a.Name), LogNPC) }
			}
		}
	}

	a.SyncState()
	a.UpdateEffects() 
}

func (a *Actor) updateDecay(ctx *SystemContext) {
	if a.ActionState != ActorDead { return }
	a.RotTicks++
	
	if a.RotTicks > TicksPerDay {
		if a.Tick%600 == 0 { a.LastReaction = "ROTTEN" }
		
		if ctx.World != nil {
			radius := 4.0
			for _, other := range ctx.World.Characters {
				if !other.IsAlive() || other.Actor.ID == a.ID { continue }
				dist := math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2))
				if dist < radius {
					other.Actor.State.Sanity -= 0.05
					other.Actor.State.Hygiene -= 0.1
					if rand.Float64() < 0.0005 {
						other.Actor.State.IsSick = true
						if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has been nauseated by the miasma of %s.", other.Name, a.Name), LogWarning) }
					}
				}
			}
		}
	}
}

func (a *Actor) updateGrief(ctx *SystemContext) {
	if a.GriefTicks > 0 {
		a.GriefTicks--
		a.State.Sanity -= 0.05 
		if a.GriefTicks % 600 == 0 && a.IsAlive() && ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Mourning...", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm })
		}
	}
}

func (a *Actor) TriggerSocialCascade(ctx *SystemContext) {
	if ctx == nil || ctx.World == nil { return }
	for _, other := range ctx.World.Characters {
		if other.ID == a.ID || !other.IsAlive() { continue }
		sentiment := other.Actor.Relationships[a.ID]
		if sentiment > 30.0 || other.Actor.ParentID == a.ID {
			other.Actor.GriefTicks += TicksPerDay * 3 
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s is devastated by the death of %s.", other.Name, a.Name), LogWarning) }
		}
	}
}

func (a *Actor) updateAge(ctx *SystemContext) {
	if !a.IsAlive() { return }
	a.AgeTicks += a.State.Age.Rate
	a.State.Age.Current = a.AgeTicks / float64(TicksPerYear)

	if a.State.Age.Max > 0 && a.State.Age.Current >= a.State.Age.Max {
		a.ActionState = ActorDead
		a.TriggerSocialCascade(ctx)
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
			chance := (a.AgeTicks - 85.0*ticksPY) / (15.0*ticksPY)
			if rand.Float64() < (chance * 0.000001) { 
				a.ActionState = ActorDead
				a.TriggerSocialCascade(ctx)
				if ctx != nil && ctx.Log != nil { ctx.Log(fmt.Sprintf("%s passed away of old age.", a.Name), LogNPC) }
			}
		}
	}

	if nextStage != "" && ctx != nil && ctx.Registries != nil {
		a.LifeStage = nextStage
		if nextStage == StageAdult {
			archID := "peasant"
			if a.PrimaryAttributes.Strength > 70 && a.PrimaryAttributes.Intellect < 40 {
				archID = "criminal"
			} else if a.PrimaryAttributes.Strength > 60 {
				archID = "man_at_arms"
			}
			if a.PrimaryAttributes.Wisdom > 50 { archID = "hermit" }
			if newArch, ok := ctx.Registries.Archetypes.Archetypes["archetypes/"+archID]; ok { a.Config = newArch }
		} else {
			prefix := "archetypes/" + nextStage + "/male"
			if a.Config != nil && strings.Contains(a.Config.Archetype, "female") { prefix = "archetypes/" + nextStage + "/female" }
			if newArch, ok := ctx.Registries.Archetypes.Archetypes[prefix]; ok { a.Config = newArch }
		}
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s grown into %s!", a.Name, nextStage), LogNPC) }
		a.SyncStats(ctx.Registries.Objects)
		a.State.HealthPoints = a.State.MaxHealthPoints
	}
}

func (a *Actor) updateScale(ctx *SystemContext) {
	if a.ActionState == ActorDead { return }
	targetScale := 1.0
	switch a.LifeStage {
	case StageBaby:     targetScale = 0.3
	case StageKid:      targetScale = 0.5
	case StageTeenager: targetScale = 0.8
	}
	
	// Smooth transition
	if a.Scale < 0.1 { a.Scale = targetScale }
	a.Scale = a.Scale*0.999 + targetScale*0.001
}

func (a *Actor) updateIllness(ctx *SystemContext) {
	if ctx == nil || ctx.World == nil { return }
	
	if a.State.IsSeptic && a.Tick%600 == 0 {
		a.State.HealthPoints -= 2
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Septic Pain!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm })
	}

	if a.FluTicks > 0 {
		a.FluTicks--; a.State.Fatigue += 0.02
		if a.State.Fatigue > 95 && a.Tick%180 == 0 {
			a.State.HealthPoints -= 1
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "-Exhausted-", X: a.X, Y: a.Y, Life: 60, Color: ColorHarm })
		}
		if a.Tick%600 == 0 {
			radius := 4.0
			for _, other := range ctx.World.Characters {
				if other.Actor.FluTicks > 0 || !other.IsAlive() || other.ID == a.ID { continue }
				dist := math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2))
				if dist < radius && rand.Float64() < 0.12 {
					if IsDebugEnabled() { DebugLog("Contagion TRIGGERED from %s to %s", a.Name, other.Name) }
					other.Actor.FluTicks = 86400 * 3; other.Actor.State.IsSick = true
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s caught fever from %s!", other.Name, a.Name), LogWarning) }
				}
			}
		}
		if a.FluTicks <= 0 && a.SicknessTicks <= 0 { a.State.IsSick = false }
	} else {
		chance := 0.000005
		if ctx.World.State.Season == SeasonAutumn { chance *= 10 }
		if rand.Float64() < chance {
			a.FluTicks = (3 + rand.Intn(5)) * 17280; a.State.IsSick = true
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has developed a fever.", a.Name), LogWarning) }
		}
	}

	if a.SicknessTicks > 0 {
		a.SicknessTicks--
		if a.Sickness == "stomach sickness" {
			a.State.Sanity -= 0.00005; a.State.Pain += 0.0001
			if a.Tick%600 == 0 { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Stomach Cramps!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm }) }
		}
		if a.SicknessTicks == 0 {
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s recovered from %s.", a.Name, a.Sickness), LogInfo) }
			a.Sickness = ""; if a.FluTicks <= 0 { a.State.IsSick = false }
		}
	}
}
