package game

import (
	"fmt"
	"math"
	"math/rand"
)

func (a *Actor) isWilling() bool {
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

	if a.Shift != ShiftLeisure { return }

	var mate *Actor
	minDist := 2.0
	for _, char := range ctx.World.Characters {
		other := &char.Actor
		if other.Name == a.Name || !other.IsAlive() { continue }
		
		if a.isBioOpposite(other) || (adultMode && !a.Config.IsAnimal) {
			sentiment := a.Relationships[other.ID]
			canMate := false
			
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

			if canMate {
				if d := math.Sqrt(math.Pow(a.X-other.X, 2)+math.Pow(a.Y-other.Y, 2)); d < minDist { mate = other; break }
			}
		}
	}

	if mate != nil {
		if a.Config.IsAnimal { a.mate(ctx, mate, "vaginal") } else if adultMode {
			a.mate(ctx, mate, "vaginal")
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
	if a.GetBioSex() == "female" { return }

	a.MatingCooldown, mate.MatingCooldown = 50000, 50000
	if ctx.Log != nil { 
		relStr := "BONDED"
		if a.Relationships[mate.Name] <= 40 { relStr = "CASUAL" }
		if a.State.IsDrunk || mate.State.IsDrunk { relStr += " UNINHIBITED" }
		ctx.Log(fmt.Sprintf("[%s]: initiates %s mating with %s", a.Name, relStr, mate.Name), LogNPC) 
	}

	if practice == "vaginal" {
		f := a; if a.GetBioSex() == "male" { f = mate }
		if f.State.Arousal < 30 && !a.Config.IsAnimal {
			f.CausePain(10.0, ctx)
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives pain (un-aroused)", f.Name), LogNPC) }
		}
	}

	a.State.Arousal, mate.State.Arousal = 0, 0
	a.State.Hygiene -= 30; mate.State.Hygiene -= 30
	if a.State.Hygiene < 0 { a.State.Hygiene = 0 }; if mate.State.Hygiene < 0 { mate.State.Hygiene = 0 }

	if practice != "vaginal" { return }

	var mother, father *Actor
	if a.GetBioSex() == "female" && mate.GetBioSex() == "male" { mother, father = a, mate } else if a.GetBioSex() == "male" && mate.GetBioSex() == "female" { mother, father = mate, a }

	if mother == nil || father == nil { return } 

	if mother.IsTransexual || father.IsTransexual {
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] and [%s] biologically sterile pair.", mother.Name, father.Name), LogNPC) }
		return
	}

	chance := mother.GetAbilityYield("mate")
	if mother.Config.ID == "courtesan_female" { chance *= 0.2 }

	if rand.Float64() < chance && !mother.IsPregnant {
		mother.IsPregnant, mother.FatherID = true, father.Name
		mother.GestationTicks = 9 * TicksPerMonth
		if mother.Config.IsAnimal { mother.GestationTicks = TicksPerMonth }
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] is pregnant (F:%s)", mother.Name, father.Name), LogNPC) }
	}
}

func (a *Actor) giveBirth(ctx *SystemContext) {
	a.IsPregnant = false
	archID := a.Config.ID
	if a.Config.IsAnimal {
		if archID == "sheep" || archID == "ram" { archID = "lamb" } else if archID == "cow" || archID == "bull" { archID = "calf" }
	} else {
		archID = "peasant_female"; if a.Config.Group == "Nobility" { archID = "noble_female" }
	}

	arch, ok := ctx.Registries.Archetypes.Archetypes[archID]; if !ok { arch = a.Config }
	child := NewCharacter(a.X, a.Y, arch, 1, false, ctx.Registries.Objects)
	child.Alignment, child.ParentID, child.FatherID = a.Alignment, a.Name, a.FatherID
	
	// Reset to infancy
	child.LifeStage = StageBaby
	child.AgeTicks = 0
	child.State.Age.Current = 0
	child.State.Age.Rate = 1.0 

	var father *Actor
	if ctx.World.PlayableCharacter != nil && ctx.World.PlayableCharacter.Name == a.FatherID { father = &ctx.World.PlayableCharacter.Actor } else {
		for _, char := range ctx.World.Characters { if char.Name == a.FatherID { father = &char.Actor; break } }
	}

	if father != nil {
		child.PrimaryAttributes.Strength = (a.PrimaryAttributes.Strength + father.PrimaryAttributes.Strength) / 2
		child.PrimaryAttributes.Dexterity = (a.PrimaryAttributes.Dexterity + father.PrimaryAttributes.Dexterity) / 2
		child.PrimaryAttributes.Health = (a.PrimaryAttributes.Health + father.PrimaryAttributes.Health) / 2
		child.PrimaryAttributes.Intellect = (a.PrimaryAttributes.Intellect + father.PrimaryAttributes.Intellect) / 2
		child.PrimaryAttributes.Wisdom = (a.PrimaryAttributes.Wisdom + father.PrimaryAttributes.Wisdom) / 2
	}

	if rand.Float64() < 0.05 {
		mod := -15; if rand.Float64() < 0.2 { mod = 20 }
		child.PrimaryAttributes.Strength += mod; child.PrimaryAttributes.Dexterity += mod
		child.PrimaryAttributes.Health += mod; child.PrimaryAttributes.Intellect += mod; child.PrimaryAttributes.Wisdom += mod
	}
	
	clampAttr := func(v int) int { if v < 5 { return 5 }; if v > 100 { return 100 }; return v }
	child.PrimaryAttributes.Strength, child.PrimaryAttributes.Dexterity = clampAttr(child.PrimaryAttributes.Strength), clampAttr(child.PrimaryAttributes.Dexterity)
	child.PrimaryAttributes.Health, child.PrimaryAttributes.Intellect = clampAttr(child.PrimaryAttributes.Health), clampAttr(child.PrimaryAttributes.Intellect)
	child.PrimaryAttributes.Wisdom = clampAttr(child.PrimaryAttributes.Wisdom)
	
	child.SyncStats(ctx.Registries.Objects)
	child.State.MaxHealthPoints /= 2; child.State.HealthPoints = child.State.MaxHealthPoints
	ctx.World.Characters = append(ctx.World.Characters, child)
	if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s]: has given birth to a child", a.Name), LogNPC) }
}
