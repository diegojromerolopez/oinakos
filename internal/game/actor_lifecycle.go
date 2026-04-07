package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync/atomic"
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

	// Optimization: Run biological/psychological simulation less frequently in large steps
	simStep := 10
	if ctx != nil && ctx.Settings != nil && ctx.Settings.SimStep > 0 { simStep = ctx.Settings.SimStep }
	
	if a.Tick % simStep == 0 {
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
	}

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
		if a.WorkTicks < 1800 { a.State.Sanity += 0.0005 } // Pride in work
		if a.WorkTicks > 7200 { a.State.Sanity -= 0.0001 } // Moderate stress (reduced from 0.01)
	} else if isLeisure {
		a.LeisureTicks++
		a.WorkTicks = 0
		if a.LeisureTicks < 3600 { a.State.Sanity += 0.001 } // Rest recovery
		// Removed long-leisure sanity penalty (0.005)

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
	
	// Natural Death Check: If state updates (needs, illness) caused death, trigger die()
	if !a.IsAlive() && a.ActionState != ActorDead {
		// This handles the transition from another state to Dead
		a.die(nil, ctx)
	} else if a.State.HealthPoints <= a.GetDeathThreshold() && a.ActionState != ActorDead {
		// Direct HP threshold check
		a.die(nil, ctx)
	}

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
					other.Actor.State.Sanity -= 0.005 // Reduced from 0.05
					other.Actor.State.Hygiene -= 0.05 // Reduced from 0.1
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
		simStep := 10
		if ctx != nil && ctx.Settings != nil && ctx.Settings.SimStep > 0 { simStep = ctx.Settings.SimStep }
		
		loss := 0.001 * float64(simStep)
		a.GriefTicks -= simStep
		if a.GriefTicks < 0 { a.GriefTicks = 0 }
		
		a.State.Sanity -= loss
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
			other.Actor.Lock()
			other.Actor.GriefTicks += TicksPerDay * 3 
			other.Actor.Unlock()
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

	// INFANT MORTALITY: One-time mortality roll during the first year
	if a.LifeStage == StageBaby && !a.MortalityChecked && a.AgeTicks > (TicksPerYear/2) {
		a.MortalityChecked = true
		// 1/3 die in the first year of life
		if rand.Float64() < 0.33 {
			a.ActionState = ActorDead
			a.TriggerSocialCascade(ctx)
			if ctx != nil && ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has passed away in infancy.", a.Name), LogNPC) }
			atomic.AddInt64(&ctx.World.Demographics.DeathsNatural, 1)
			return
		}
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
				atomic.AddInt64(&ctx.World.Demographics.DeathsNatural, 1)
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
	if ctx == nil { return }
	
	// Recovery logic for non-flu sickness
	if a.SicknessTicks > 0 {
		a.SicknessTicks--
		if a.Sickness == "stomach sickness" {
			a.State.Sanity -= 0.0001; a.State.Pain += 0.0002
			if a.Tick%600 == 0 && ctx.World != nil { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Stomach Cramps!", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm }) }
		}
		if a.SicknessTicks == 0 {
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s recovered from %s.", a.Name, a.Sickness), LogInfo) }
			a.Sickness = ""; if a.FluTicks <= 0 { a.State.IsSick = false }
		}
	}

	// Calculate Resistance Bonus
	// Users: adults/elders and kids 10+ are more resilient.
	resilienceBonus := 0
	isAdultAge := a.LifeStage == StageAdult || a.LifeStage == StageElder
	isResilientKid := a.LifeStage == StageKid && a.State.Age.Current >= 10.0
	if isAdultAge || isResilientKid {
		resilienceBonus = 25
	}

	if a.FluTicks > 0 {
		a.FluTicks--
		a.State.Fatigue += 0.03 // Increased fatigue drain
		
		damageInterval := 300
		if a.State.Fatigue > 80 { damageInterval = 150 }
		
		if a.Tick % damageInterval == 0 {
			// Health Roll: Can they resist the symptoms this interval?
			roll := rand.Intn(100)
			if roll > (a.PrimaryAttributes.Health + resilienceBonus) {
				a.State.HealthPoints -= 1
				if ctx.World != nil {
					color := ColorHarm
					if a.State.HealthPoints < 20 { color = ColorWarn }
					ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Fever Burn", X: a.X, Y: a.Y, Life: 60, Color: color })
				}
			}
		}

		// Contagion logic: Proximal spread
		if a.Tick%600 == 0 && ctx.World != nil {
			radius := 4.0
			for _, other := range ctx.World.Characters {
				if other.Actor.FluTicks > 0 || !other.IsAlive() || other.ID == a.ID { continue }
				dist := math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2))
				if dist < radius {
					exposureChance := 0.15
					if rand.Float64() < exposureChance {
						// Pathogen Lethality vs Actor Health + Resilience
						pathogenStrength := 45 + rand.Intn(20) 
						
						otherResilience := 0
						if other.Actor.LifeStage == StageAdult || other.Actor.LifeStage == StageElder || (other.Actor.LifeStage == StageKid && other.Actor.State.Age.Current >= 10.0) {
							otherResilience = 25
						}
						
						healthRoll := rand.Intn(other.Actor.PrimaryAttributes.Health + otherResilience + 20)
						
						if healthRoll > pathogenStrength {
							if rand.Float64() < 0.8 { // 80% chance of total immunity if roll passed
								continue 
							}
							other.Actor.FluTicks = 17280 // Mild, 1 day
						} else {
							other.Actor.FluTicks = (3 + rand.Intn(7)) * 17280 // 3-10 days
						}
						
						other.Actor.State.IsSick = true
						if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s caught fever from %s!", other.Name, a.Name), LogWarning) }
					}
				}
			}
		}
		if a.FluTicks <= 0 { a.State.IsSick = false }
	} else {
		// SPONTANEOUS OUTBREAKS: Rare and Seasonal (Autumn/Winter)
		if ctx.World != nil {
			season := ctx.World.State.Season
			chance := 0.000001 // Reduced base chance
			
			if season == SeasonAutumn {
				chance *= 5.0
			} else if season == SeasonWinter {
				chance *= 10.0
			} else {
				chance = 0.0 // Pathogens dormant in Spring/Summer
			}

			if rand.Float64() < chance {
				// Initial health roll for patient zero
				if rand.Intn(a.PrimaryAttributes.Health + resilienceBonus) < 30 {
					a.FluTicks = (5 + rand.Intn(10)) * 17280; a.State.IsSick = true
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has developed a SEVERE fever.", a.Name), LogWarning) }
				} else {
					a.FluTicks = (2 + rand.Intn(4)) * 17280; a.State.IsSick = true
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has developed a mild fever.", a.Name), LogWarning) }
				}
			}
		}
	}
}
