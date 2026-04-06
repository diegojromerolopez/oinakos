package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync/atomic"
	"unsafe"
)

func (a *Actor) isWilling() bool {
	if a.SexualOrientation == "asexual" { return false }
	return a.IsAlive() && a.State.Sanity >= 0 && a.MatingCooldown <= 0 && !a.IsPregnant
}

func (a *Actor) updateBreeding(ctx *SystemContext) {
	if !a.IsAlive() || a.Config == nil { return }
	
	adultMode := true
	if ctx != nil && ctx.Settings != nil { adultMode = ctx.Settings.AdultMode }
	if !adultMode && !a.Config.IsAnimal { return }

	if a.IsPregnant {
		a.GestationTicks--
		if a.GestationTicks <= 0 { a.giveBirth(ctx) }
		return
	}
	if a.MatingCooldown > 0 { 
		a.MatingCooldown-- 
		return 
	}

	if !a.Config.IsAnimal && a.Shift != ShiftLeisure && a.ActionState != ActorBerserk { return }

	var mate *Actor
	minDist := 10.0 // Increased interaction range for social density
	if IsFastMode() { minDist = 15.0 } // Scale to match frame-skipping velocity
	if IsDebugEnabled() && a.Tick%TicksPerHour == 0 { DebugLog("BREEDING-UPDATE: %s is searching (Shift:%d, Sanity:%.1f, Arousal:%.1f)", a.Name, a.Shift, a.State.Sanity, a.State.Arousal) }

	for _, char := range ctx.World.Characters {
		other := &char.Actor
		if other.Name == a.Name || !other.IsAlive() { continue }
		
		isPreferred := false
		if a.SexualOrientation == "" || a.SexualOrientation == "heterosexual" {
			if a.isBioOpposite(other) { isPreferred = true }
		} else if a.SexualOrientation == "homosexual" {
			if !a.isBioOpposite(other) { isPreferred = true }
		}

		if isPreferred || (adultMode && !a.Config.IsAnimal) {
			sentiment := 0.0
			if a.Relationships != nil { sentiment = a.Relationships[other.ID] }

			canMate := false
			
			// PROFESSIONAL GATE (Elite Simulation: Economic Requirement)
			isService := other.Config != nil && strings.Contains(other.Config.ID, "courtesan")
			isViolent := a.ActionState == ActorBerserk || a.Behavior == BehaviorCriminal

			if isService && !isViolent {
				// Non-violent mating with courtesans REQUIRES money.
				if a.Denarii >= 2 && other.isWilling() { canMate = true }
			} else {
				// 1. Arousal or Alcohol driven (Casual/Uninhibited)
				isUninhibited := a.State.Arousal > 50 || other.State.Arousal > 50 || a.State.IsDrunk || other.State.IsDrunk
				if a.isWilling() && other.isWilling() && isUninhibited {
					canMate = true
				} 
				// 2. Emotional driven (Romantic)
				if !canMate && a.isWilling() && other.isWilling() && sentiment > 40 {
					canMate = true
				}
				// 3. Incapacitated (if allowed)
				if !canMate && a.isWilling() && (other.ActionState == ActorIncapacitated) {
					canMate = true
				}
				// 4. Hostile/Violent (Psychotic break bypasses social/economic rules)
				if !canMate && isViolent && other.IsAlive() {
					canMate = true
				}
				// 5. BLACKOUT MATING (Uninhibited drunk logic)
				if !canMate && a.State.IsDrunk && other.State.IsDrunk && a.isBioOpposite(other) && sentiment > 10 {
					if rand.Float64() < 0.5 { canMate = true }
				}
				// 6. CRIMINAL PREDATION
				if !canMate && a.Behavior == BehaviorCriminal && other.IsAlive() && a.Name != other.Name {
					if a.Config.IsAnimal == other.Config.IsAnimal { canMate = true }
				}
				// 7. HYPERSEXUAL COMPULSION (Mental Break result)
				if !canMate && a.State.IsHypersexual && other.IsAlive() && a.Name != other.Name {
					if a.Config.IsAnimal == other.Config.IsAnimal { canMate = true }
				}
			}

			if canMate {
				dist := math.Sqrt(math.Pow(a.X-other.X, 2)+math.Pow(a.Y-other.Y, 2))
				if IsDebugEnabled() && a.Tick%600 == 0 { DebugLog("BREEDING-CANDIDATE: %s found %s at %.1f pedes", a.Name, other.Name, dist) }
				if dist < minDist { mate = other; break }
				
				// IMPROVEMENT: If searching and found a mate, move towards them!
				if a.Config.IsAnimal && dist < 100.0 && a.ActionState == ActorIdle {
					a.TargetActor = other
					// Setting target to move towards
				}
			}
		}
	}

	if mate == nil && a.TargetActor != nil && a.TargetActor.IsAlive() && a.Config.IsAnimal {
		// Just setting the TargetActor is enough; the Character AI will 
		// handle movement in findTarget() -> executeMovement() during the next tick.
		// We set ActionState to ActorIdle to ensure we aren't 'busy' so movement can trigger.
		if a.ActionState == ActorIdle || a.ActionState == ActorWalking {
			a.ActionState = ActorIdle
		}
		return
	}

	if mate != nil {
		practice := "vaginal"
		if a.Behavior == BehaviorCriminal && rand.Float64() < 0.5 { practice = "anal" }
		
		if a.Config.IsAnimal { a.mate(ctx, mate, "vaginal") } else if adultMode {
			a.mate(ctx, mate, practice)
		}
	}
}

func (a *Actor) isBioOpposite(other *Actor) bool {
	if a.Config == nil || other.Config == nil { return false }
	// Humans / NPC can mate within the same broad species (non-animals)
	if !a.Config.IsAnimal && !other.Config.IsAnimal {
		return a.GetBioSex() != other.GetBioSex()
	}
	// Animals must be of the same species ID
	if a.Config.ID != other.Config.ID { return false }
	return a.GetBioSex() != other.GetBioSex()
}

func (a *Actor) GetBioSex() string {
	if a.Config == nil { return "unknown" }
	if a.IsTransexual {
		if a.Config.Gender == "female" { return "male" }
		return "female"
	}
	return a.Config.Gender
}

func (a *Actor) mate(ctx *SystemContext, mate *Actor, practice string) {
	if a == nil || mate == nil { return }

	// Deadlock prevention: Lock in order of memory address
	if uintptr(unsafe.Pointer(a)) < uintptr(unsafe.Pointer(mate)) {
		a.Lock(); defer a.Unlock()
		mate.Lock(); defer mate.Unlock()
	} else {
		mate.Lock(); defer mate.Unlock()
		a.Lock(); defer a.Unlock()
	}

	// SERVICE TRANSACTION (Elite Simulation: Economic Layer)
	isService := mate.Config != nil && strings.Contains(mate.Config.ID, "courtesan")
	isViolent := a.ActionState == ActorBerserk || a.Behavior == BehaviorCriminal
	
	if isService && !isViolent {
		if a.Denarii < 2 {
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] cannot afford professional courtesan services (%d/2 Denarii)", a.Name, a.Denarii), LogNPC) }
			return
		}
		a.Denarii -= 2; mate.Denarii += 2
	}

	a.MatingCooldown, mate.MatingCooldown = 5000, 5000
	if ctx.Log != nil { 
		relStr := "BONDED"
		sentiment := 0.0
		if a.Relationships != nil { sentiment = a.Relationships[mate.ID] }
		if sentiment <= 40 { relStr = "CASUAL" }
		if a.State.IsDrunk || mate.State.IsDrunk { relStr += " UNINHIBITED" }
		if isService { relStr = "PROFESSIONAL-SERVICE" }
		if isViolent { relStr = "HOSTILE-FORCED" }
		ctx.Log(fmt.Sprintf("[%s]: initiates %s mating with %s", a.Name, relStr, mate.Name), LogNPC) 
	}

	if practice == "vaginal" {
		f := a; if a.GetBioSex() == "male" { f = mate }
		// Only cause pain/trauma if the participant was UNWILLING.
		if f.State.Arousal < 30 && !a.Config.IsAnimal && !f.isWilling() {
			f.CausePain(40.0, ctx)
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives severe physical trauma (assault)", f.Name), LogNPC) }
		}
	} else if practice == "anal" {
		// Only cause pain if unwilling or specifically hostile
		if !mate.isWilling() || isViolent {
			mate.CausePain(30.0, ctx)
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives physical trauma from forced intercourse", mate.Name), LogNPC) }
		}
	}

	// BLACKOUT MEMORY LOSS: If both were drunk, they gain NO relationship from the act
	if a.State.IsDrunk && mate.State.IsDrunk {
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] and [%s] have no memory of their drunken encounter.", a.Name, mate.Name), LogNPC) }
	} else if !isViolent {
		a.ModifySentiment(mate.ID, 5.0)
		mate.ModifySentiment(a.ID, 5.0)
	}

	if isViolent {
		if mate.Relationships == nil { mate.Relationships = make(map[string]float64) }
		mate.Relationships[a.ID] -= 50.0 // Massive relationship damage
		mate.State.Sanity -= 10.0        // Emotional trauma (reduced from 20)
		mate.ModifySubmission(a.ID, 25.0) // Submission as a trauma response
		mate.AddMemory(mate.Tick, "trauma", a.Name, -50.0)
		
		// Probability of acute mental break leading to depression
		if rand.Float64() < 0.15 {
			mate.State.Sanity -= 15.0
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] has suffered a complete mental breakdown after the assault.", mate.Name), LogWarning) }
		}

		// PREDATOR REWARD: Criminal gain satisfaction/dominance from the act
		a.State.Sanity += 30.0
		a.State.Fatigue -= 20.0
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s feels empowered by their dominance over %s.", a.Name, mate.Name), LogNPC) }
	} else {
		// Consensual mating REWARDS sanity
		a.State.Sanity += 15.0
		mate.State.Sanity += 15.0
	}

	a.State.Arousal, mate.State.Arousal = 0, 0
	
	// TAVERN AS HYDRATION HUB: Successful social mating grants a hydration buffer (frozen thirst)
	if !isViolent {
		a.State.HydrationBuffer = 3000
		mate.State.HydrationBuffer = 3000
	}
	a.State.Hygiene -= 10; mate.State.Hygiene -= 10
	if a.State.Hygiene < 0 { a.State.Hygiene = 0 }; if mate.State.Hygiene < 0 { mate.State.Hygiene = 0 }

	// POST-COITAL EXHAUSTION: Physical toll of the act
	a.State.Fatigue += 40.0; mate.State.Fatigue += 40.0
	if isViolent { mate.State.Fatigue += 30.0 } // Trauma-induced exhaustion
	
	// Force characters into Resting state to recover
	a.ActionState = ActorResting
	mate.ActionState = ActorResting
	a.Tick, mate.Tick = 0, 0 // Reset animation/action timers

	if practice != "vaginal" {
		atomic.AddInt64(&ctx.World.Demographics.MatingActs, 1)
		return
	}

	var mother, father *Actor
	if a.GetBioSex() == "female" && mate.GetBioSex() == "male" { mother, father = a, mate } else if a.GetBioSex() == "male" && mate.GetBioSex() == "female" { mother, father = mate, a }
	if mother == nil || father == nil { 
		atomic.AddInt64(&ctx.World.Demographics.MatingActs, 1)
		return 
	} 
	if mother.IsTransexual || father.IsTransexual { 
		atomic.AddInt64(&ctx.World.Demographics.MatingActs, 1)
		return 
	}

	// Actual pregnancy check
	fertility := mother.GetFertilityMultiplier()
	chance := 0.3 * fertility
	if mother.Config.IsAnimal { 
		chance = 0.8 * fertility // Animals conceive more reliably
	}

	if rand.Float64() < chance {
		mother.IsPregnant = true
		mother.GestationTicks = 0
		if mother.Config.IsAnimal { 
			mother.State.Age.Max = 1.0 * float64(TicksPerMonth) // Set gestation for animals
		}
		mother.FatherID = father.Name
		atomic.AddInt64(&ctx.World.Demographics.MatingPregancies, 1)
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] is now PREGNANT with %s's child", mother.Name, father.Name), LogNPC) }
	}
	
	atomic.AddInt64(&ctx.World.Demographics.MatingActs, 1)
}

func (a *Actor) GetFertilityMultiplier() float64 {
	if a.GetBioSex() != "female" { return 1.0 }
	age := a.State.Age.Current
	if age < 18 { return 0.0 }
	if age < 35 { return 1.0 }
	if age < 45 {
		// Linear decline 1.0 -> 0.2
		return 1.0 - (age - 35) * (0.8 / 10.0)
	}
	return 0.0 // Menopause
}

func (a *Actor) giveBirth(ctx *SystemContext) {
	archID := a.Config.ID
	if a.Config.IsAnimal {
		if archID == "sheep" || archID == "ram" { archID = "lamb" } else if archID == "cow" || archID == "bull" { archID = "calf" } else if archID == "pig" || archID == "boar" { archID = "piglet" }
	} else {
		// Default to peasant child if not specified otherwise
		if strings.Contains(strings.ToLower(a.Config.Group), "noble") {
			archID = "noble_female" 
		} else {
			archID = "peasant_female"
		}
	}

	arch, ok := ctx.Registries.Archetypes.Archetypes[archID]; if !ok { arch = a.Config }
	child := NewCharacter(a.X, a.Y, arch, 1, false, ctx.Registries.Objects)
	child.Alignment, child.ParentID, child.FatherID = a.Alignment, a.Name, a.FatherID
	child.LifeStage, child.AgeTicks = StageBaby, 0
	child.State.Age.Rate = 1.0 

	var father *Actor
	for _, char := range ctx.World.Characters { if char.Name == a.FatherID { father = &char.Actor; break } }
	if father != nil {
		child.PrimaryAttributes.Strength = (a.PrimaryAttributes.Strength + father.PrimaryAttributes.Strength) / 2
		child.PrimaryAttributes.Dexterity = (a.PrimaryAttributes.Dexterity + father.PrimaryAttributes.Dexterity) / 2
		child.PrimaryAttributes.Health = (a.PrimaryAttributes.Health + father.PrimaryAttributes.Health) / 2
		child.PrimaryAttributes.Intellect = (a.PrimaryAttributes.Intellect + father.PrimaryAttributes.Intellect) / 2
		child.PrimaryAttributes.Wisdom = (a.PrimaryAttributes.Wisdom + father.PrimaryAttributes.Wisdom) / 2
	}
	child.SyncStats(ctx.Registries.Objects)
	child.State.MaxHealthPoints /= 2; child.State.HealthPoints = child.State.MaxHealthPoints
	ctx.World.Characters = append(ctx.World.Characters, child)
	
	a.IsPregnant = false
	a.FatherID = ""
	a.MatingCooldown = 6000 // Short recovery

	if a.Config.IsAnimal {
		atomic.AddInt64(&ctx.World.Demographics.BirthsAnimals, 1)
	} else {
		atomic.AddInt64(&ctx.World.Demographics.BirthsHumans, 1)
	}

	if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s]: has given birth to a child", a.Name), LogNPC) }
}
