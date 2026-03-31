package game

import (
	"math"
	"strings"
)

func (a *Actor) GetTotalProtection() int { return a.BaseProtection + a.ProtectionBonus }
func (a *Actor) GetTotalAttack() int { return a.calculateStat(a.BaseAttack, a.Level) + a.AttackBonus }
func (a *Actor) GetTotalDefense() int { return a.calculateStat(a.BaseDefense, a.Level) + a.DefenseBonus }

func (a *Actor) SyncStats(objReg *ObjectRegistry) {
	if a.Config == nil { return }

	a.PrimaryAttributes.Strength = clampInt(a.PrimaryAttributes.Strength, 0, 100)
	a.PrimaryAttributes.Dexterity = clampInt(a.PrimaryAttributes.Dexterity, 0, 100)
	a.PrimaryAttributes.Health = clampInt(a.PrimaryAttributes.Health, 0, 100)
	a.PrimaryAttributes.Intellect = clampInt(a.PrimaryAttributes.Intellect, 0, 100)
	a.PrimaryAttributes.Wisdom = clampInt(a.PrimaryAttributes.Wisdom, 0, 100)

	a.SyncState()

	age := float64(a.AgeTicks) / float64(TicksPerYear)
	pMult, mMult := 1.0, 1.0
	if age < 25 {
		penaltyPrc := (25.0 - age) / 25.0
		pMult, mMult = 1.0-(0.25*penaltyPrc), 1.0-(0.30*penaltyPrc)
	} else if age > 40 {
		mMult = 1.0 + (0.05 * math.Floor((age-40.0)/10.0))
		pPenalty := 0.25 * (age - 40.0) / (85.0 - 40.0)
		if pPenalty > 0.25 { pPenalty = 0.25 }
		pMult = 1.0 - pPenalty
	}

	str, dex, hlt, itl, wis := float64(a.PrimaryAttributes.Strength)*pMult, float64(a.PrimaryAttributes.Dexterity)*pMult, float64(a.PrimaryAttributes.Health)*pMult, float64(a.PrimaryAttributes.Intellect)*mMult, float64(a.PrimaryAttributes.Wisdom)*mMult

	// PREGNANCY PENALTIES (Biology Item 5: Maternal Burden)
	if a.IsPregnant {
		str, dex, itl, wis = str*0.6, dex*0.5, itl*0.7, wis*0.8
	}

	if a.State.Arousal > 10 { itl, wis = itl - a.State.Arousal*0.5, wis - a.State.Arousal*0.5 }
	if a.State.IsDrunk { dex, itl, wis = dex*0.7, itl*0.7, wis*0.7 }
	if itl < 1 { itl = 1 }; if wis < 1 { wis = 1 }

	if a.RawStats.BaseAttack > 0 { a.BaseAttack = int(float64(a.RawStats.BaseAttack) * pMult) } else { a.BaseAttack = int(str * 2) }
	if a.RawStats.BaseDefense > 0 { a.BaseDefense = int(float64(a.RawStats.BaseDefense) * pMult) } else { a.BaseDefense = int(dex*1.5 + hlt*1.0) }
	
	a.RangedAttack, a.CriticalChance = int(dex * 2), str * 0.005
	a.Speed = dex * 0.02
	if a.RawStats.Speed > 0 { a.Speed = a.RawStats.Speed * pMult }
	if a.IsPregnant { a.Speed *= 0.7 } // Move 30% slower
	if a.Speed <= 0 { a.Speed = 0.01 }

	a.Nourishment, a.Survivalism, a.Mate = int(hlt * 2), int(str*0.5 + hlt*0.5), hlt * 0.01
	a.Crafting, a.Herbalism, a.Trading, a.Harvesting, a.Husbandry, a.Art, a.Culture = int(itl*1.2 + str*0.3), int(wis*1.0 + itl*0.5), int(itl*1.2 + wis*0.3), int(wis*1.2 + dex*0.3), int(wis*1.0 + dex*0.5), int(dex*0.5 + itl*0.5), int(itl*0.5 + wis*0.5)

	if a.RawStats.HealthMin > 0 { a.State.MaxHealthPoints = int(float64(a.RawStats.HealthMin) * pMult) } else { a.State.MaxHealthPoints = int(hlt * 10) }
	if a.State.MaxHealthPoints < 10 { a.State.MaxHealthPoints = 10 }
	if a.State.HealthPoints > a.State.MaxHealthPoints { a.State.HealthPoints = a.State.MaxHealthPoints }

	cooldownMult := 1.5 - (dex * 0.01)
	baseCD := a.RawStats.AttackCooldown; if baseCD == 0 { baseCD = 60 }
	a.BaseAttackCooldown = int(float64(baseCD) * cooldownMult)
	if a.BaseAttackCooldown < 10 { a.BaseAttackCooldown = 10 }

	if a.RawStats.MaxWeight > 0 { a.MaxWeight = a.RawStats.MaxWeight } else { a.MaxWeight = (str*1.5 + hlt*0.5) / 0.329 }
	a.BaseWeapon = a.Config.Weapon.Resolve(objReg); if a.BaseWeapon == nil { a.BaseWeapon = WeaponFists }
	a.Weapon, a.BaseProtection = a.BaseWeapon, a.calculateStat(a.RawStats.BaseProtection, a.Level)
}

func (a *Actor) calculateStat(base int, level int) int {
	if level <= 1 { return base }
	return int(float64(base) * math.Pow(1.15, float64(level-1)))
}

func (a *Actor) getAttrValue(attr string) int {
	attr, val := strings.ToLower(attr), 0
	switch attr {
	case "strength":  val = a.PrimaryAttributes.Strength
	case "dexterity": val = a.PrimaryAttributes.Dexterity
	case "health":    val = a.PrimaryAttributes.Health
	case "intellect": val = a.PrimaryAttributes.Intellect
	case "wisdom":    val = a.PrimaryAttributes.Wisdom
	}
	if (attr == "intellect" || attr == "wisdom") && a.State.Arousal > 10 {
		val -= int(a.State.Arousal * 0.5)
		if val < 1 { val = 1 }
	}
	return val
}

func (a *Actor) GetAbilityYield(abilityID string) float64 {
	switch abilityID {
	case "milk": return float64(a.Husbandry) * 1.0
	case "shear": return float64(a.Husbandry) * 0.5
	case "forage": return float64(a.Survivalism) * 0.3
	case "cook", "brew", "heal": return float64(a.Herbalism) * 1.0
	case "rest": return float64(a.Nourishment) * 0.25
	case "eat": return float64(a.Nourishment) * 1.0
	case "drink": return float64(a.Nourishment) * 0.8
	case "mate": return a.Mate * 1.0
	case "plant": return float64(a.Harvesting) * 0.8
	case "harvest_crop": return float64(a.Harvesting) * 1.5
	case "water_crops": return float64(a.Harvesting) * 0.5
	case "fish", "butcher", "guard", "stash": return float64(a.Survivalism) * 0.5
	case "hunt": return float64(a.Survivalism) * 1.0
	case "trap": return float64(a.Survivalism) * 0.8
	case "craft", "build": return float64(a.Crafting) * 1.0
	case "repair": return float64(a.Crafting) * 0.5
	case "smelt", "tan": return float64(a.Crafting) * 0.8
	case "trade": return float64(a.Trading) * 1.0
	case "appraise": return float64(a.Trading) * 0.5
	case "haul": return float64(a.PrimaryAttributes.Strength) * 0.01
	case "sneak": return float64(a.PrimaryAttributes.Dexterity) * 1.0
	case "steal": return float64(a.PrimaryAttributes.Dexterity) * 0.5
	case "pray", "bury": return float64(a.Culture) * 0.3
	case "teach": return float64(a.Culture) * 0.5
	case "intimidate", "recruit", "lie", "seduce", "perform": return float64(a.Culture) * 1.0
	case "compose", "read": return float64(a.Culture) * 0.5
	case "paint", "sculpt": return float64(a.Art) * 0.8
	case "weave": return float64(a.Crafting) * 0.5
	}
	return 0.0
}
