package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strings"
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

// TakeDamage reduces health and potentially applies trauma.
func (a *Actor) TakeDamage(amount int, attacker ActorInterface, ctx *SystemContext) {
	if a.State == ActorDead { return }
	if ctx.Log != nil {
		attackerName := "something"
		if attacker != nil && attacker.GetActor() != nil { attackerName = attacker.GetActor().Name }
		ctx.Log(fmt.Sprintf("[%s]: is damaged by %s for %d", a.Name, attackerName, amount), LogCombatDamage)
	}
	
	if a.Relationships == nil { a.Relationships = make(map[string]float64) }
	if a.RomanticInterest == nil { a.RomanticInterest = make(map[string]float64) }
	attackerID := "unknown"
	if attacker != nil && attacker.GetActor() != nil {
		attn := attacker.GetActor()
		attackerID = attn.Name 
		a.AddMemory(a.Tick, "attack", attackerID, -5.0)
		a.Relationships[attackerID] -= 2.0
		
		// Group Sentiment Ripple (Item 3)
		if ctx != nil && ctx.World != nil && ctx.World.State.GroupSentiment != nil {
			victimGroup := a.Group
			attackerGroup := attn.Group
			if victimGroup != "" && attackerGroup != "" {
				if ctx.World.State.GroupSentiment[victimGroup] == nil {
					ctx.World.State.GroupSentiment[victimGroup] = make(map[string]float64)
				}
				ctx.World.State.GroupSentiment[victimGroup][attackerGroup] -= 5.0
				if ctx.World.State.GroupSentiment[victimGroup][attackerGroup] < -100 {
					ctx.World.State.GroupSentiment[victimGroup][attackerGroup] = -100
				}
			}
		}
	}

	a.TemporalState.HealthPoints -= amount
	a.CausePain(float64(amount)*0.8, ctx) // High pain per HP loss
	a.HitTimer = 30
	a.DegradeArmor(ctx)
	
	// Distribute to Limb
	if a.BodyStatus == nil { a.InitBodyStatus() }
	limbs := []string{"head", "torso", "l_arm", "r_arm", "l_leg", "r_leg"}
	weights := []int{10, 40, 12, 12, 13, 13} 
	totalW := 100
	r := rand.Intn(totalW)
	chosenLimb := "torso"
	runningW := 0
	for i, w := range weights {
		runningW += w
		if r < runningW { chosenLimb = limbs[i]; break }
	}
	a.BodyStatus[chosenLimb] -= amount
	if a.BodyStatus[chosenLimb] < 0 { a.BodyStatus[chosenLimb] = 0 }

	// Permanent Trauma Acquisition via Limb Death
	if a.BodyStatus["l_leg"] <= 0 && !a.Trauma.LeftLegLost { a.Trauma.LeftLegLost = true; DebugLog("%s LOST LEFT LEG", a.Name) }
	if a.BodyStatus["r_leg"] <= 0 && !a.Trauma.RightLegLost { a.Trauma.RightLegLost = true; DebugLog("%s LOST RIGHT LEG", a.Name) }
	if a.BodyStatus["l_arm"] <= 0 && !a.Trauma.LeftArmLost { a.Trauma.LeftArmLost = true; DebugLog("%s LOST LEFT ARM", a.Name) }
	if a.BodyStatus["r_arm"] <= 0 && !a.Trauma.RightArmLost { a.Trauma.RightArmLost = true; DebugLog("%s LOST RIGHT ARM", a.Name) }
	if a.BodyStatus["head"] <= 0 && a.Trauma.EyesLost < 2 { a.Trauma.EyesLost = 2; DebugLog("%s IS NOW BLIND", a.Name) }

	// Sepsis Acquisition (Item 4: Wound Infection)
	if ctx != nil && ctx.World != nil && amount > 5 {
		risk := 0.05 // 5% base risk per heavy hit
		if attacker != nil && attacker.GetActor() != nil && attacker.GetActor().Config != nil && attacker.GetActor().Config.IsAnimal {
			risk = 0.20 // Animals/Zombies are filthier
		}
		if rand.Float64() < risk && !a.TemporalState.IsSeptic {
			a.TemporalState.IsSeptic = true
			if ctx.Log != nil {
				ctx.Log(fmt.Sprintf("%s's wounds have become septic!", a.Name), LogNPC)
			}
		}
	}

		// Desperation Trauma: Critical hits at low health bypass limb HP
	if a.TemporalState.HealthPoints < a.GetTotalMaxHealth()/10 && a.TemporalState.HealthPoints > -10 && amount > 0 {
		a.acquireRandomTrauma(attacker)
	}

	deathThreshold := a.GetDeathThreshold()
	if a.TemporalState.HealthPoints < deathThreshold {
		a.TemporalState.HealthPoints = deathThreshold
	}

	a.SyncLifeStatus()
	
	if !a.IsAlive() {
		a.die(attacker, ctx)
	}
}

func (a *Actor) acquireRandomTrauma(attacker ActorInterface) {
	r := rand.Intn(7)
	switch r {
	case 0:
		if !a.Trauma.LeftArmLost {
			a.Trauma.LeftArmLost = true
			DebugLog("Actor [%s] %s lost their LEFT ARM!", a.Alignment, a.Name)
		}
	case 1:
		if !a.Trauma.RightArmLost {
			a.Trauma.RightArmLost = true
			DebugLog("Actor [%s] %s lost their RIGHT ARM!", a.Alignment, a.Name)
		}
	case 2:
		if !a.Trauma.LeftLegLost {
			a.Trauma.LeftLegLost = true
			DebugLog("Actor [%s] %s lost their LEFT LEG!", a.Alignment, a.Name)
		}
	case 3:
		if !a.Trauma.RightLegLost {
			a.Trauma.RightLegLost = true
			DebugLog("Actor [%s] %s lost their RIGHT LEG!", a.Alignment, a.Name)
		}
	case 4:
		if a.Trauma.EyesLost < 2 {
			a.Trauma.EyesLost++
			DebugLog("Actor [%s] %s lost an EYE! (Total lost: %d)", a.Alignment, a.Name, a.Trauma.EyesLost)
		}
	case 5:
		if !a.Trauma.BurnedAlive {
			a.Trauma.BurnedAlive = true
			DebugLog("Actor [%s] %s was BURNED ALIVE and survived!", a.Alignment, a.Name)
		}
	case 6:
		if !a.Trauma.SpineBroken {
			a.Trauma.SpineBroken = true
			DebugLog("Actor [%s] %s suffered a BROKEN SPINE!", a.Alignment, a.Name)
		}
	}
}

func (a *Actor) die(attacker ActorInterface, ctx *SystemContext) {
	a.State = ActorDead
	
	// INHERITANCE (Item 3: Inheritance & Lineage)
	if ctx != nil && ctx.World != nil {
		// Find offspring
		var heir *Character
		for _, char := range ctx.World.Characters {
			if char.ParentID == a.Name && char.IsAlive() {
				heir = char; break
			}
		}
		if heir != nil {
			heir.Denarii += a.Denarii
			if a.OwnedChestID != "" && heir.OwnedChestID == "" {
				heir.OwnedChestID = a.OwnedChestID
			}
			if ctx.Log != nil {
				ctx.Log(fmt.Sprintf("%s has inherited %d denarii from their parent %s.", heir.Name, a.Denarii, a.Name), LogNPC)
			}
			a.Denarii = 0 // Transferred
		}
	}

	if ctx != nil && ctx.World != nil && ctx.World.Game != nil { 
		ctx.World.Game.DropAllItems(a) 
	}
	
	prefix := "unknown"
	if a.Config != nil { 
		prefix = a.Config.SoundID 
		if prefix == "" { prefix = a.Config.ID }
		a.MeatQuantity = float64(a.Config.Meat)
	}
	
	if ctx != nil && ctx.Audio != nil { 
		ctx.Audio.PlayRandomSound(prefix + "/death") 
	}
	
	if attacker != nil {
		if act := attacker.GetActor(); act != nil {
			act.Kills++
			if a.Config != nil {
				act.MapKills[a.Config.ID]++
				xp := a.Config.XP
				if xp <= 0 { xp = 1 }
				act.AddXP(xp)
			}
			
			if act.Config != nil && act.Config.Actions != nil {
				for _, action := range act.Config.Actions.OnKill {
					if rand.Float64() < action.Probability {
						a.applyKillAction(action, attacker, ctx)
					}
				}
			}
		}
	}
}

func (a *Actor) applyKillAction(action KillAction, attacker ActorInterface, ctx *SystemContext) {
	if IsDebugEnabled() { DebugLog("applyKillAction: %s on %s", action.Type, a.Name) }
	if action.Type == "transform_victim" {
		e := action.Effect.Victim
		if e == nil { return }
		
		targetID := e.Transform
		if a.Config != nil {
			targetID = strings.ReplaceAll(targetID, "{gender}", a.Config.Gender)
		}
		
		var newConfig *EntityConfig
		var ok bool
		if ctx != nil && ctx.Registries != nil {
			if ctx.Registries.Archetypes != nil {
				newConfig, ok = ctx.Registries.Archetypes.Archetypes[targetID]
			}
			if !ok && ctx.Registries.Characters != nil {
				newConfig, ok = ctx.Registries.Characters.Characters[targetID]
			}
		}
		
		if ok {
			a.Config = newConfig
			a.TemporalState.HealthPoints = a.GetTotalMaxHealth()
			a.InitBodyStatus()
			a.UnconsciousTimer = 0
			a.State = ActorIdle
			if e.Alignment == "inherit" {
				a.Alignment = attacker.GetActor().Alignment
			}
		}
	}
	
	if action.Type == "heal_attacker" || (action.Effect.Attacker != nil && action.Effect.Attacker.Heal > 0) {
		attk := attacker.GetActor()
		if action.Effect.Attacker != nil {
			attk.TemporalState.HealthPoints += action.Effect.Attacker.Heal
			if attk.TemporalState.HealthPoints > attk.GetTotalMaxHealth() {
				attk.TemporalState.HealthPoints = attk.GetTotalMaxHealth()
			}
		}
	}
}

func (a *Actor) Heal(amount int) {
	if a.IsTrulyDead() {
		return
	}
	oldHealth := a.TemporalState.HealthPoints
	a.TemporalState.HealthPoints += amount
	maxH := a.GetTotalMaxHealth()
	if a.TemporalState.HealthPoints > maxH {
		a.TemporalState.HealthPoints = maxH
	}
	
	a.SyncLifeStatus()

	if a.TemporalState.HealthPoints > oldHealth {
		DebugLog("Actor Healed [%s] %s! +%d | Health: %d -> %d", a.Alignment, a.Name, amount, oldHealth, a.TemporalState.HealthPoints)
	}
}

func (a *Actor) DegradeArmor(ctx *SystemContext) {
	for slot, it := range a.Slots {
		if it != nil && it.Config != nil && it.Config.Slot != "weapon" && it.Config.Resistance > 0 {
			it.Resistance--
			if it.Resistance <= 0 {
				delete(a.Slots, slot)
				newInv := []*ItemInstance{}
				for _, invIt := range a.Inventory { if invIt != it { newInv = append(newInv, invIt) } }
				a.Inventory = newInv
				if IsDebugEnabled() { DebugLog("%s: Armor %s has BROKEN!", a.Name, it.Config.Name) }
				a.UpdateEffects()
			}
		}
	}
}

func (a *Actor) DegradeWeapon(ctx *SystemContext) {
	it, ok := a.Slots["weapon"]
	if !ok || it == nil || it.Config == nil { return }
	if it.Resistance <= 0 { return }

	it.Resistance--
	if it.Resistance <= 0 {
		delete(a.Slots, "weapon")
		newInv := []*ItemInstance{}
		for _, invIt := range a.Inventory { if invIt != it { newInv = append(newInv, invIt) } }
		a.Inventory = newInv
		if IsDebugEnabled() { DebugLog("%s: Weapon %s has BROKEN!", a.Name, it.Config.Name) }
		a.UpdateEffects()
	}
}

func (a *Actor) rollDamage() int {
	dmg := a.BaseAttack
	if a.Weapon != nil { 
		dmg += a.Weapon.RollDamage() 
	} else {
		dmg += 1 + rand.Intn(2) // Fists variance
	}

	// Critical Strike check
	if rand.Float64() <= a.CriticalChance {
		dmg *= 2
		if IsDebugEnabled() { DebugLog("CRITICAL HIT by %s", a.Name) }
	}

	return dmg
}

func (a *Actor) Torture(target *Actor, ctx *SystemContext) {
	if !target.IsAlive() { return }
	
	// Can only torture if target is incapacitated or unconscious
	canTorture := target.IsIncapacitated() || target.UnconsciousTimer > 0 || target.Trauma.SpineBroken
	if !canTorture {
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s cannot be tortured while they can still resist!", target.Name), LogNPC) }
		return
	}

	// Torture causes extreme pain and potentially physical trauma
	painAmount := 20.0 + rand.Float64()*10.0
	target.CausePain(painAmount, ctx)
	
	// Sentiment impact
	target.Relationships[a.Name] -= 10.0
	target.AddMemory(a.Tick, "torture", a.Name, -20.0)

	if ctx.Log != nil {
		ctx.Log(fmt.Sprintf("[%s]: is torturing %s", a.Name, target.Name), LogNPC)
		ctx.Log(fmt.Sprintf("[%s] receives extreme pain", target.Name), LogNPC)
	}
	target.LastReaction = "MERCY! PLEASE! STOP!!"

	// 10% chance of random trauma per torture session
	if rand.Float64() < 0.10 {
		target.acquireRandomTrauma(a)
	}
}

func (c *Character) hitCharacter(target *Actor, skill string, ctx *SystemContext) {
	if !target.IsAlive() && c.State == ActorChopping {
		c.executeButchery(target, ctx)
		return
	}

	// Use ability-driven damage if a skill is provided
	dmg := c.rollDamage()
	if skill != "" {
		dmg = c.GetAbilityDamage(skill)
	}

	defense := target.BaseDefense // Dexterity*1.5 + Health*1.0
	protection := target.GetTotalProtection()
	
	finalDmg := dmg - (defense / 3) - protection // Modified mitigation factor for balance
	if finalDmg < 1 { finalDmg = 1 }

	target.TakeDamage(finalDmg, c, ctx)
	
	// Blunt attacks (hit/kick) cause more pain than health loss
	if skill == AttackPunch || skill == AttackKick || skill == AttackSlap {
		painBonus := float64(finalDmg) * 1.5 // 1.5x additional pain for blunt trauma
		target.CausePain(painBonus, ctx)
	}

	// Restrain Ability (Utility)
	if skill == "restrain" {
		if c.CompetitiveContest(target, "dexterity", "strength") {
			target.UnconsciousTimer = 300 // 5 seconds @ 60 TPS
			target.State = ActorIncapacitated
			if ctx.Log != nil { 
				ctx.Log(fmt.Sprintf("[%s]: has restrained %s", c.Name, target.Name), LogNPC)
				ctx.Log(fmt.Sprintf("[%s] receives restrained", target.Name), LogNPC)
			}
		} else {
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s tried to restrain %s but failed!", c.Name, target.Name), LogNPC) }
		}
	}

	// Resolve status effects and custom mechanics
	if skill != "" {
		c.ResolveAbilityEffects(skill, target, ctx)
	}

	// Legacy submission gain handles specific skills like SLAP
	submissionGain := 0.0
	if skill == AttackSlap {
		submissionGain = 3.0 + rand.Float64()*4.0
	} else if skill == AttackPunch {
		submissionGain = 1.0 + rand.Float64()*2.0
	}

	if submissionGain > 0 {
		if target.Submission == nil { target.Submission = make(map[string]float64) }
		target.Submission[c.Name] += submissionGain
		if target.Submission[c.Name] > 100 { target.Submission[c.Name] = 100 }
		
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: fmt.Sprintf("+%.1f submission", submissionGain), X: target.X, Y: target.Y - 1.5, Life: 60, Color: color.RGBA{200, 100, 255, 255},
			})
		}
	}
	
	if ctx != nil && ctx.World != nil {
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: fmt.Sprintf("%d", finalDmg), X: target.X, Y: target.Y - 1, Life: 45, Color: ColorHarm,
		})
	}
}

func (c *Character) executeButchery(target *Actor, ctx *SystemContext) {
	if target.MeatQuantity <= 0 {
		return
	}

	// 1 Meat per strike (roughly)
	yield := 1.0 + rand.Float64()*2.0
	if yield > target.MeatQuantity { yield = target.MeatQuantity }
	
	target.MeatQuantity -= yield
	
	// Feedback
	ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
		Text: "butchering...", X: target.X, Y: target.Y - 0.5, Life: 30, Color: ColorHeal,
	})

	if ctx.Log != nil {
		ctx.Log(fmt.Sprintf("[%s]: is butchering %s", c.Name, target.Name), LogNPC)
	}
	// Spawn meat if threshold reached
	if target.MeatQuantity <= 0 || int(target.MeatQuantity+yield)/5 > int(target.MeatQuantity)/5 {
		meatConfig := ctx.Registries.Objects.Get("raw_meat")
		if meatConfig != nil {
			dropX := target.X + (rand.Float64()*1.0 - 0.5)
			dropY := target.Y + (rand.Float64()*1.0 - 0.5)
			meat := NewItemInstance("raw_meat", meatConfig, dropX, dropY)
			ctx.World.Items = append(ctx.World.Items, meat)
			
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: "+Raw Meat", X: target.X, Y: target.Y - 1.2, Life: 60, Color: ColorHeal,
			})
		}
	}
}
func (a *Actor) GetAbilityDamage(abilityID string) int {
	if a.Config == nil || a.Config.Abilities == nil {
		return a.BaseAttack
	}
	ability, ok := a.Config.Abilities[abilityID]
	if !ok {
		return a.BaseAttack
	}

	formula := ability.Damage
	var multiplier float64

	switch {
	case strings.HasPrefix(formula, "melee_attack * "):
		if _, err := fmt.Sscanf(formula, "melee_attack * %f", &multiplier); err == nil {
			return int(float64(a.BaseAttack) * multiplier)
		}
	case strings.HasPrefix(formula, "ranged_attack * "):
		if _, err := fmt.Sscanf(formula, "ranged_attack * %f", &multiplier); err == nil {
			return int(float64(a.RangedAttack) * multiplier)
		}
	case strings.HasPrefix(formula, "attack * "):
		// Legacy format — treat as melee_attack for backward compat
		if _, err := fmt.Sscanf(formula, "attack * %f", &multiplier); err == nil {
			return int(float64(a.BaseAttack) * multiplier)
		}
	}

	return a.BaseAttack
}

func (a *Actor) ResolveAbilityEffects(abilityID string, target *Actor, ctx *SystemContext) {
	if a.Config == nil || a.Config.Abilities == nil { return }
	ability, ok := a.Config.Abilities[abilityID]
	if !ok { return }

	for _, effect := range ability.Effects {
		// Roll for probability if specified
		if effect.Probability > 0 && rand.Float64() > effect.Probability {
			continue
		}

		// Apply Stun
		if effect.StunChance > 0 && rand.Float64() <= effect.StunChance {
			target.UnconsciousTimer = int(effect.Duration * 60) // Convert seconds to ticks
			target.State = ActorIncapacitated
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives stunned", target.Name), LogNPC) }
			if IsDebugEnabled() { DebugLog("%s STUNNED by %s", target.Name, a.Name) }
		}

		// Apply Knockback
		if effect.KnockbackDistance > 0 {
			dx, dy := target.X-a.X, target.Y-a.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 {
				target.X += (dx / dist) * effect.KnockbackDistance
				target.Y += (dy / dist) * effect.KnockbackDistance
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives knockback", target.Name), LogNPC) }
			}
		}

		// Apply Poison (IsPoisoned flag for now, expanded in update loop)
		if effect.PoisonDamagePerSecond > 0 {
			target.TemporalState.IsPoisoned = true
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives poisoned", target.Name), LogNPC) }
			// In a real implementation, we'd add a status effect timer here
		}
	}
}
