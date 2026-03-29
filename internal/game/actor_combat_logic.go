package game

import "math/rand"

func (a *Actor) CheckAbilitySuccess(abilityID string, modifier int) bool {
	if a.SkillValues != nil { if val, exists := a.SkillValues[abilityID]; exists { return a.checkThreshold(val, modifier) } }
	if a.Config != nil && a.Config.Abilities != nil {
		if ability, exists := a.Config.Abilities[abilityID]; exists && ability.ParentAttribute != "" {
			return a.CheckAttributeSuccess(ability.ParentAttribute, modifier)
		}
	}
	switch abilityID {
	case "punch", "kick", "heavy_strike", "chop", "dig", "build", "butcher", "throw", "knockout", "grapple": return a.CheckAttributeSuccess("strength", modifier)
	case "slap", "slash", "shoot_arrow", "milk", "shear", "sneak", "steal", "seduce", "weave": return a.CheckAttributeSuccess("dexterity", modifier)
	case "rest", "eat", "drink", "mate": return a.CheckAttributeSuccess("health", modifier)
	case "cook", "craft", "repair", "brew", "trade", "smelt", "read", "appraise", "intimidate", "tan", "lie": return a.CheckAttributeSuccess("intellect", modifier)
	case "forage", "plant", "harvest_crop", "tame", "fish", "pray", "guard", "hunt", "trap", "tend_animal", "breed", "water_crops", "bury", "recruit", "teach": return a.CheckAttributeSuccess("wisdom", modifier)
	}
	return false
}

func (a *Actor) CheckAttributeSuccess(attr string, modifier int) bool {
	val := a.getAttrValue(attr)
	return a.checkThreshold(val, modifier)
}

func (a *Actor) checkThreshold(val, modifier int) bool {
	effective := clampInt(val+modifier, 0, 100)
	return rand.Intn(101) <= effective
}

func (a *Actor) CompetitiveAttributeRoll(other *Actor, attr string) bool {
	return a.CompetitiveContest(other, attr, attr)
}

func (a *Actor) CompetitiveContest(other *Actor, myAttr, theirAttr string) bool {
	valA, valB := a.getAttrValue(myAttr), other.getAttrValue(theirAttr)
	rollA, rollB := rand.Intn(101), rand.Intn(101)
	successA, successB := rollA <= valA, rollB <= valB
	if successA && !successB { return true }
	if !successA && successB { return false }
	if successA && successB { return rollA < rollB }
	return false
}
