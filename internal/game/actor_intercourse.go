package game

import (
	"fmt"
	"math"
	"math/rand"
)

func (a *Actor) isWilling() bool {
	// Willing if alive, sane enough, and not on cooldown
	return a.IsAlive() && a.TemporalState.Sanity > 20 && a.MatingCooldown <= 0 && !a.IsPregnant
}

func (a *Actor) updateBreeding(ctx *SystemContext) {
	if !a.IsAlive() || a.Config == nil { return }
	
	adultMode := true
	if ctx != nil && ctx.Settings != nil {
		adultMode = ctx.Settings.AdultMode
	}

	// If adult mode is off, skip human breeding logic but keep animals
	if !adultMode && !a.Config.IsAnimal {
		return
	}

	// 1. Timer decrements
	if a.MatingCooldown > 0 { a.MatingCooldown-- }
	if a.IsPregnant {
		a.GestationTicks--
		if a.GestationTicks <= 0 {
			a.giveBirth(ctx)
		}
		return
	}

	// 2. Mating Logic: Only during Leisure shift
	if a.Shift != ShiftLeisure || a.IsPregnant { return }

	// Check for nearby mate
	var mate *Actor
	minDist := 2.0
	for _, char := range ctx.World.Characters {
		other := &char.Actor
		if other.Name == a.Name || !other.IsAlive() { continue }
		
		// Eligibility Check
		if a.isBioOpposite(other) || (adultMode && !a.Config.IsAnimal) {
			// Eligibility: (Both willing) OR (One willing AND other incapacitated/immobilized)
			canMate := false
			if a.isWilling() && other.isWilling() {
				canMate = true
			} else if a.isWilling() && (other.State == ActorIncapacitated) {
				canMate = true
			}

			if canMate {
				dist := math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2))
				if dist < minDist {
					mate = other
					break
				}
			}
		}
	}

	if mate != nil {
		if a.Config.IsAnimal {
			a.mate(ctx, mate, "vaginal")
		} else if adultMode {
			// Auto sex for humans in leisure
			practice := "vaginal"
			if rand.Float64() < 0.3 { practice = "anal" }
			a.mate(ctx, mate, practice)
		}
	}
}

func (a *Actor) isBioOpposite(other *Actor) bool {
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
	// GENDER ROLES: Only those with a penis (bio-males) can be on the giving end.
	// The initiator 'a' in mate(mate) is the giver.
	if a.GetBioSex() == "female" {
		return
	}

	// Set cooldowns
	a.MatingCooldown = 50000 
	mate.MatingCooldown = 50000

	if ctx.Log != nil {
		ctx.Log(fmt.Sprintf("[%s]: is mating with %s", a.Name, mate.Name), LogNPC)
	}

	// Handle effects
	receiving := mate
	
	// Arousal effects
	if practice == "vaginal" {
		female := a
		if a.GetBioSex() == "male" { female = mate }
		
		if female.TemporalState.Arousal < 30 && !a.Config.IsAnimal {
			female.CausePain(10.0, ctx)
			if ctx.Log != nil { 
				ctx.Log(fmt.Sprintf("[%s] receives pain from lack of arousal", female.Name), LogNPC)
			}
		}
	} else if practice == "anal" {
		receiving.CausePain(15.0, ctx)
		if ctx.Log != nil { 
			ctx.Log(fmt.Sprintf("[%s] receives discomfort from anal intercourse", receiving.Name), LogNPC)
		}
	}

	// Relieve arousal for both
	a.TemporalState.Arousal = 0
	mate.TemporalState.Arousal = 0

	// Hygiene decreases a lot during sex
	a.TemporalState.Hygiene -= 30
	mate.TemporalState.Hygiene -= 30
	if a.TemporalState.Hygiene < 0 { a.TemporalState.Hygiene = 0 }
	if mate.TemporalState.Hygiene < 0 { mate.TemporalState.Hygiene = 0 }

	// Pregnancy logic
	if practice != "vaginal" { return }

	// Determine mother/father
	var mother, father *Actor
	if a.GetBioSex() == "female" && mate.GetBioSex() == "male" {
		mother, father = a, mate
	} else if a.GetBioSex() == "male" && mate.GetBioSex() == "female" {
		mother, father = mate, a
	}

	if mother == nil || father == nil { return } // Same bio sex cannot breed

	// Chance of pregnancy
	chance := mother.GetAbilityYield("mate") // health * 0.01 (0.0 to 1.0)
	if mother.Config.ID == "courtesan_female" {
		chance *= 0.2 // Courtesans are more specialized
	}

	if rand.Float64() < chance {
		if !mother.IsPregnant {
			mother.IsPregnant = true
			mother.FatherID = father.Name
			mother.AddMemory(ctx.World.State.Ticks, "mated", father.Name, 0)
			
			mother.GestationTicks = 9 * TicksPerMonth
			if mother.Config.IsAnimal {
				mother.GestationTicks = TicksPerMonth
			}

			if ctx.Log != nil {
				ctx.Log(fmt.Sprintf("[%s] receives pregnant", mother.Name), LogNPC)
			}
		}
	}
}

func (a *Actor) giveBirth(ctx *SystemContext) {
	a.IsPregnant = false
	
	// Determine child archetype
	archID := a.Config.ID
	if a.Config.IsAnimal {
		if archID == "sheep" || archID == "ram" {
			archID = "lamb"
		} else if archID == "cow" || archID == "bull" {
			archID = "calf"
		}
	} else {
		// Default to peasant, but inherit group
		archID = "peasant_female" 
		if a.Config.Group == "Nobility" { archID = "noble_female" }
	}

	arch, ok := ctx.Registries.Archetypes.Archetypes[archID]
	if !ok {
		arch = a.Config
	}

	child := NewCharacter(a.X, a.Y, arch, 1, false, ctx.Registries.Objects)
	child.Alignment = a.Alignment
	child.ParentID = a.Name
	child.FatherID = a.FatherID
	
	// INHERITANCE: Blend parents' attributes
	// We find the father if possible
	var father *Actor
	if ctx.World.PlayableCharacter != nil && ctx.World.PlayableCharacter.Name == a.FatherID {
		father = &ctx.World.PlayableCharacter.Actor
	} else {
		for _, char := range ctx.World.Characters {
			if char.Name == a.FatherID {
				father = &char.Actor
				break
			}
		}
	}

	if father != nil {
		child.PrimaryAttributes.Strength = (a.PrimaryAttributes.Strength + father.PrimaryAttributes.Strength) / 2
		child.PrimaryAttributes.Dexterity = (a.PrimaryAttributes.Dexterity + father.PrimaryAttributes.Dexterity) / 2
		child.PrimaryAttributes.Health = (a.PrimaryAttributes.Health + father.PrimaryAttributes.Health) / 2
		child.PrimaryAttributes.Intellect = (a.PrimaryAttributes.Intellect + father.PrimaryAttributes.Intellect) / 2
		child.PrimaryAttributes.Wisdom = (a.PrimaryAttributes.Wisdom + father.PrimaryAttributes.Wisdom) / 2
	}

	// MUTATION (Item 2: Genetic Mutation)
	mutationChance := 0.05
	if rand.Float64() < mutationChance {
		mutationType := "Frail"
		bonus := -15
		if rand.Float64() < 0.2 { // Rare Heroic mutation (1% total)
			mutationType = "Heroic"
			bonus = 20
		}
		child.PrimaryAttributes.Strength += bonus
		child.PrimaryAttributes.Dexterity += bonus
		child.PrimaryAttributes.Health += bonus
		child.PrimaryAttributes.Intellect += bonus
		child.PrimaryAttributes.Wisdom += bonus
		if ctx.Log != nil {
			ctx.Log(fmt.Sprintf("%s's child was born with a %s mutation!", a.Name, mutationType), LogNPC)
		}
	}
	
	// Clamp attributes to 0-100
	clampAttr := func(v int) int { if v < 5 { return 5 }; if v > 100 { return 100 }; return v }
	child.PrimaryAttributes.Strength = clampAttr(child.PrimaryAttributes.Strength)
	child.PrimaryAttributes.Dexterity = clampAttr(child.PrimaryAttributes.Dexterity)
	child.PrimaryAttributes.Health = clampAttr(child.PrimaryAttributes.Health)
	child.PrimaryAttributes.Intellect = clampAttr(child.PrimaryAttributes.Intellect)
	child.PrimaryAttributes.Wisdom = clampAttr(child.PrimaryAttributes.Wisdom)
	
	child.SyncStats(ctx.Registries.Objects)

	// Scaled down (Child)
	child.TemporalState.MaxHealthPoints /= 2
	child.TemporalState.HealthPoints = child.TemporalState.MaxHealthPoints
	
	ctx.World.Characters = append(ctx.World.Characters, child)

	if ctx.Log != nil {
		ctx.Log(fmt.Sprintf("[%s]: has given birth to a child", a.Name), LogNPC)
	}
}
