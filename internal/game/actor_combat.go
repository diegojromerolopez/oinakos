package game

import (
	"fmt"
	"math/rand"
)

const (
	AttackNormal = ""
	AttackPunch  = "punch"
	AttackSlap   = "slap"
	AttackKick   = "kick"
	AttackSlash  = "slash"
	AttackHeavy  = "heavy_strike"
	AttackArrow  = "shoot_arrow"
	AttackPower  = "power_shot"
	AttackBite   = "infect_bite"
)

func (a *Actor) TakeDamage(amount int, attacker ActorInterface, ctx *SystemContext) {
	if a.ActionState == ActorDead { return }
	if ctx.Log != nil {
		attnName := "something"; if attacker != nil && attacker.GetActor() != nil { attnName = attacker.GetActor().Name }
		ctx.Log(fmt.Sprintf("[%s]: damaged by %s for %d", a.Name, attnName, amount), LogCombatDamage)
	}
	if a.Relationships == nil { a.Relationships = make(map[string]float64) }
	if attacker != nil && attacker.GetActor() != nil {
		attn := attacker.GetActor(); a.AddMemory(a.Tick, "attack", attn.Name, -5.0)
		a.Relationships[attn.Name] -= 2.0; a.ModifyGroupSentiment(ctx, attn.Group, -5.0)
	}
	a.State.HealthPoints -= amount; a.CausePain(float64(amount)*0.8, ctx); a.HitTimer = 30; a.DegradeArmor(ctx)
	if a.BodyStatus == nil { a.InitBodyStatus() }
	limbs, weights, chosenLimb := []string{"head", "torso", "l_arm", "r_arm", "l_leg", "r_leg"}, []int{10, 40, 12, 12, 13, 13}, "torso"
	r, runningW := rand.Intn(100), 0
	for i, w := range weights { runningW += w; if r < runningW { chosenLimb = limbs[i]; break } }
	a.BodyStatus[chosenLimb] -= amount; if a.BodyStatus[chosenLimb] < 0 { a.BodyStatus[chosenLimb] = 0 }
	if a.BodyStatus["l_leg"] <= 0 && !a.Trauma.LeftLegLost { a.Trauma.LeftLegLost = true }
	if a.BodyStatus["r_leg"] <= 0 && !a.Trauma.RightLegLost { a.Trauma.RightLegLost = true }
	if a.BodyStatus["l_arm"] <= 0 && !a.Trauma.LeftArmLost { a.Trauma.LeftArmLost = true }
	if a.BodyStatus["r_arm"] <= 0 && !a.Trauma.RightArmLost { a.Trauma.RightArmLost = true }
	if a.BodyStatus["head"] <= 0 && a.Trauma.EyesLost < 2 { a.Trauma.EyesLost = 2 }
	if ctx != nil && ctx.World != nil && amount > 5 {
		risk := 0.05; if attacker != nil && attacker.GetActor() != nil && attacker.GetActor().Config != nil && attacker.GetActor().Config.IsAnimal { risk = 0.20 }
		if rand.Float64() < risk && !a.State.IsSeptic { a.State.IsSeptic = true }
	}
	if a.State.HealthPoints < a.GetTotalMaxHealth()/10 && amount > 0 { a.acquireRandomTrauma(attacker) }
	if a.State.HealthPoints < a.GetDeathThreshold() { a.State.HealthPoints = a.GetDeathThreshold() }
	a.SyncLifeStatus(); if !a.IsAlive() { a.die(attacker, ctx) }
}

func (a *Actor) Heal(amount int) {
	if a.IsTrulyDead() { return }
	a.State.HealthPoints += amount; maxH := a.GetTotalMaxHealth()
	if a.State.HealthPoints > maxH { a.State.HealthPoints = maxH }
	a.SyncLifeStatus()
}

func (a *Actor) DegradeArmor(ctx *SystemContext) {
	for slot, it := range a.Slots {
		if it != nil && it.Config != nil && it.Config.Slot != "weapon" && it.Resistance > 0 {
			it.Resistance--
			if it.Resistance <= 0 {
				delete(a.Slots, slot)
				// Thoroughly remove from inventory as well
				for i, invItem := range a.Inventory {
					if invItem == it { a.Inventory = append(a.Inventory[:i], a.Inventory[i+1:]...); break }
				}
				a.UpdateEffects()
			}
		}
	}
}

func (a *Actor) DegradeWeapon(ctx *SystemContext) {
	if it := a.Slots["weapon"]; it != nil && it.Resistance > 0 {
		it.Resistance--
		if it.Resistance <= 0 {
			delete(a.Slots, "weapon")
			// Thoroughly remove from inventory as well
			for i, invItem := range a.Inventory {
				if invItem == it { a.Inventory = append(a.Inventory[:i], a.Inventory[i+1:]...); break }
			}
			a.UpdateEffects()
		}
	}
}

func (a *Actor) ModifyGroupSentiment(ctx *SystemContext, group string, delta float64) {
	if group == "" { return }
	if a.GroupSentiment == nil { a.GroupSentiment = make(map[string]float64) }
	a.GroupSentiment[group] += delta
}

func (a *Actor) rollDamage() int {
	dmg := a.BaseAttack
	if a.Weapon != nil { dmg += a.Weapon.RollDamage() } else { dmg += 1 + rand.Intn(2) }
	if rand.Float64() <= a.CriticalChance { dmg *= 2 }
	return dmg
}

func (a *Actor) Torture(target *Actor, ctx *SystemContext) {
	if !target.IsAlive() || (!target.IsIncapacitated() && target.UnconsciousTimer <= 0 && !target.Trauma.SpineBroken) { return }
	target.CausePain(20.0+rand.Float64()*10.0, ctx); target.Relationships[a.Name] -= 10.0; target.AddMemory(a.Tick, "torture", a.Name, -20.0)
	target.LastReaction = "MERCY! PLEASE! STOP!!"
	if rand.Float64() < 0.10 { target.acquireRandomTrauma(a) }
}

func (c *Character) hitCharacter(target *Actor, skill string, ctx *SystemContext) {
	if !target.IsAlive() && c.ActionState == ActorChopping { c.executeButchery(target, ctx); return }
	dmg := c.rollDamage(); if skill != "" { dmg = c.GetAbilityDamage(skill) }
	finalDmg := dmg - (target.BaseDefense/3) - target.GetTotalProtection()
	if finalDmg < 1 { finalDmg = 1 }
	target.TakeDamage(finalDmg, c, ctx)
	if skill == AttackPunch || skill == AttackKick || skill == AttackSlap { target.CausePain(float64(finalDmg)*1.5, ctx) }
	if skill == "restrain" && c.CompetitiveContest(target, "dexterity", "strength") { target.UnconsciousTimer, target.ActionState = 300, ActorIncapacitated }
	gain := 0.0; if skill == AttackSlap { gain = 3.0 + rand.Float64()*4.0 } else if skill == AttackPunch { gain = 1.0 + rand.Float64()*2.0 }
	if gain > 0 {
		if target.Submission == nil { target.Submission = make(map[string]float64) }
		target.Submission[c.Name] += gain; if target.Submission[c.Name] > 100 { target.Submission[c.Name] = 100 }
	}
	if c.Behavior == BehaviorCriminal && target.Denarii > 0 && rand.Float64() < 0.25 {
		stolen := 1 + rand.Intn(3)
		if stolen > target.Denarii { stolen = target.Denarii }
		target.Denarii -= stolen; c.Denarii += stolen
		c.State.Sanity += 10.0 // Criminal satisfaction gain
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s robbed %d denarii from %s!", c.Name, stolen, target.Name), LogNPC) }
	}
	if ctx != nil && ctx.World != nil { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: fmt.Sprintf("%d", finalDmg), X: target.X, Y: target.Y - 1, Life: 45, Color: ColorHarm }) }
}

func (c *Character) executeButchery(target *Actor, ctx *SystemContext) {
	if target.MeatQuantity <= 0 { return }
	yield := c.GetAbilityYield("butcher"); if yield > target.MeatQuantity { yield = target.MeatQuantity }
	if yield <= 0 { yield = 1.0 } // Minimum yield if success checked elsewhere
	target.MeatQuantity -= yield
	if target.MeatQuantity <= 0 || int(target.MeatQuantity+yield)/5 > int(target.MeatQuantity)/5 {
		_, meatCfg := ctx.Registries.Objects.RandomVariantID("raw_meat")
		if meatCfg != nil {
			meat := NewItemInstance(meatCfg.ID, meatCfg, target.X+rand.Float64()-0.5, target.Y+rand.Float64()-0.5)
			ctx.World.Items = append(ctx.World.Items, meat)
		}
	}
}
