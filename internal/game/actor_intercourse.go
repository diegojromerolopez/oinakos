package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func (a *Actor) isWilling() bool {
	if a.SexualOrientation == "asexual" { return false }
	return a.IsAlive() && a.State.Sanity > 20 && a.MatingCooldown <= 0 && !a.IsPregnant
}

func (a *Actor) updateBreeding(ctx *SystemContext) {
	if !a.IsAlive() || a.Config == nil { return }
	
	adultMode := true
	if ctx != nil && ctx.Settings != nil { adultMode = ctx.Settings.AdultMode }
	if !adultMode && !a.Config.IsAnimal { return }

	if a.MatingCooldown > 0 { a.MatingCooldown-- }
	if a.IsPregnant {
		a.GestationTicks--
		if a.GestationTicks <= 0 { a.giveBirth(ctx) }
		return
	}

	if !a.Config.IsAnimal && a.Shift != ShiftLeisure && a.ActionState != ActorBerserk { return }

	var mate *Actor
	minDist := 4.0 // Increased interaction range for social density
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
			isViolent := a.ActionState == ActorBerserk

			if isService && !isViolent {
				// Non-violent mating with courtesans REQUIRES money.
				if a.Denarii >= 2 { canMate = true }
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
					canMate = true
				}
				// 7. HYPERSEXUAL COMPULSION (Mental Break result)
				if !canMate && a.State.IsHypersexual && other.IsAlive() && a.Name != other.Name {
					canMate = true
				}
			}

			if canMate {
				if d := math.Sqrt(math.Pow(a.X-other.X, 2)+math.Pow(a.Y-other.Y, 2)); d < minDist { mate = other; break }
			}
		}
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
	if a.GetBioSex() == "female" { return }

	// SERVICE TRANSACTION (Elite Simulation: Economic Layer)
	isService := mate.Config != nil && strings.Contains(mate.Config.ID, "courtesan")
	isViolent := a.ActionState == ActorBerserk
	
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
		if f.State.Arousal < 30 && !a.Config.IsAnimal {
			f.CausePain(40.0, ctx)
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives severe physical trauma (assault)", f.Name), LogNPC) }
		}
	} else if practice == "anal" {
		mate.CausePain(30.0, ctx)
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives severe physical trauma from anal assault", mate.Name), LogNPC) }
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
		mate.State.Sanity -= 20.0        // Deep emotional trauma
		mate.ModifySubmission(a.ID, 25.0) // Submission as a trauma response
		mate.AddMemory(mate.Tick, "trauma", a.Name, -50.0)
		
		// Probability of acute mental break leading to depression
		if rand.Float64() < 0.25 {
			mate.State.Sanity -= 25.0
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] has suffered a complete mental breakdown after the assault.", mate.Name), LogWarning) }
		}
	}


	a.State.Arousal, mate.State.Arousal = 0, 0
	
	// TAVERN AS HYDRATION HUB: Successful social mating grants a hydration buffer (frozen thirst)
	if !isViolent {
		a.State.HydrationBuffer = 3000
		mate.State.HydrationBuffer = 3000
	}
	a.State.Hygiene -= 10; mate.State.Hygiene -= 10
	if a.State.Hygiene < 0 { a.State.Hygiene = 0 }; if mate.State.Hygiene < 0 { mate.State.Hygiene = 0 }

	if practice != "vaginal" { return }

	var mother, father *Actor
	if a.GetBioSex() == "female" && mate.GetBioSex() == "male" { mother, father = a, mate } else if a.GetBioSex() == "male" && mate.GetBioSex() == "female" { mother, father = mate, a }
	if mother == nil || father == nil { return } 
	if mother.IsTransexual || father.IsTransexual { return }

	chance := mother.GetAbilityYield("mate")
	if mother.Config != nil && strings.Contains(mother.Config.ID, "courtesan") { chance *= 0.2 }

	if rand.Float64() < chance && !mother.IsPregnant {
		mother.IsPregnant, mother.FatherID = true, father.Name
		mother.GestationTicks = 9 * TicksPerMonth
		if mother.Config.IsAnimal { mother.GestationTicks = TicksPerMonth }
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] is pregnant (F:%s)", mother.Name, father.Name), LogNPC) }
	}
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
	if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s]: has given birth to a child", a.Name), LogNPC) }
}
