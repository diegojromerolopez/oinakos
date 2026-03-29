package game

import (
	"fmt"
	"math/rand"
)

// ApplyPermanentEffects permanently modifies the actor's base stats based on the object's effects.
func (a *Actor) ApplyPermanentEffects(obj *ObjectConfig) {
	if obj == nil || obj.Effects == nil {
		return
	}
	for stat, effect := range obj.Effects {
		switch stat {
		case "attack":
			a.BaseAttack += int(effect.Increase)
		case "defense":
			a.BaseDefense += int(effect.Increase)
		case "speed":
			a.Speed += effect.Increase
		case "max_health":
			a.State.MaxHealthPoints += int(effect.Increase)
			a.State.HealthPoints += int(effect.Increase)
		case "xp":
			a.AddXP(int(effect.Increase))
		}
	}
}

// ConsumeItem applies one-time effects from a consumable object and returns true if used.
func (a *Actor) ConsumeItem(item *ItemInstance, ctx *SystemContext) bool {
	if item == nil || item.Config == nil {
		return false
	}
	obj := item.Config
	
	// Expiration check
	isSpoiled := false
	if obj.MaxHours > 0 && item.HoursLeft <= 0 {
		isSpoiled = true
		if ctx != nil && ctx.Log != nil {
			ctx.Log(fmt.Sprintf("[%s]: %s is spoiled! This caused stomach sickness.", a.Name, obj.Name), LogWarning)
		}
		// Effects: reduces health_points a 25%
		damage := int(float64(a.State.HealthPoints) * 0.25)
		if damage < 5 { damage = 5 } // Minimum impact
		a.TakeDamage(damage, nil, ctx)
		
		// stomach sickness that lasts 1-2 days
		// 1 day = 17280 ticks. 2 days = 34560.
		a.SicknessTicks = 17280 + rand.Intn(17280)
		a.Sickness = "stomach sickness"
		a.State.IsSick = true
	}

	if ctx != nil && ctx.Log != nil && !isSpoiled {
		ctx.Log(fmt.Sprintf("[%s]: is consuming %s", a.Name, obj.Name), LogNPC)
	}
	if obj.Effects == nil && !obj.ClearSick && obj.Hunger == 0 && obj.Thirst == 0 && obj.Fatigue == 0 && obj.Energy == 0 {
		return false
	}
	
	used := false

	// Specific need restoration fields
	if obj.Hunger > 0 { a.State.Hunger -= obj.Hunger; used = true; a.State.BowelLevel += float64(obj.Hunger) * 0.1 }
	if obj.Thirst > 0 { a.State.Thirst -= obj.Thirst; used = true; a.State.BladderLevel += float64(obj.Thirst) * 0.1 }
	if obj.Fatigue > 0 { a.State.Fatigue -= obj.Fatigue; used = true }
	if obj.Energy > 0 { 
		// Legacy: energy restores everything
		a.State.Hunger -= obj.Energy
		a.State.Thirst -= obj.Energy
		a.State.Fatigue -= obj.Energy
		used = true 
	}
	if obj.ClearSick {
		if a.FluTicks > 0 {
			a.FluTicks = 0
			used = true
		}
	}

	// Dynamic effects map
	for stat, effect := range obj.Effects {
		switch stat {
		case "hunger":
			a.State.Hunger -= effect.Increase; used = true
		case "thirst":
			a.State.Thirst -= effect.Increase; used = true
		case "fatigue":
			a.State.Fatigue -= effect.Increase; used = true
		case "energy":
			a.State.Hunger -= effect.Increase; a.State.Thirst -= effect.Increase; a.State.Fatigue -= effect.Increase; used = true
		case "health":
			a.Heal(int(effect.Increase))
			if ctx != nil && ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] receives healing", a.Name), LogNPC) }
			used = true
		case "attack", "defense", "speed", "max_health", "xp":
			a.ApplyPermanentEffects(obj)
			used = true
		}
	}

	if obj.IsAlcoholic {
		a.State.AlcoholLevel += 10.0
		used = true
		if a.State.AlcoholLevel > 30.0 {
			// Health roll to resist drunkenness
			if !a.CheckAttributeSuccess("health", 0) {
				a.State.IsDrunk = true
				if ctx != nil && ctx.Log != nil {
					ctx.Log(fmt.Sprintf("[%s] is now drunk!", a.Name), LogWarning)
				}
			}
		}
	}

	if a.State.Hunger > 100 { a.State.Hunger = 100 }
	if a.State.Thirst > 100 { a.State.Thirst = 100 }
	if a.State.Fatigue > 100 { a.State.Fatigue = 100 }
	
	if used && ctx != nil && ctx.World != nil {
		a.ActionState, a.Tick = ActorEating, 0
		msg := fmt.Sprintf("+%s", obj.Name)
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: msg, X: a.X, Y: a.Y - 1, Life: 45, Color: ColorHeal,
		})
	}
	
	return used
}

func (a *Actor) UpdateEffects() {
	a.AttackBonus = 0
	a.DefenseBonus = 0
	a.ProtectionBonus = 0
	a.SpeedBonus = 0
	a.MaxHealthBonus = 0
	a.RegenPerSecond = 0

	// Apply effects from equipped slots
	for _, it := range a.Slots {
		if it == nil || it.Config == nil {
			continue
		}
		for stat, effect := range it.Config.Effects {
			switch stat {
			case "attack":
				a.AttackBonus += int(effect.Increase)
			case "defense":
				a.DefenseBonus += int(effect.Increase)
			case "protection":
				a.ProtectionBonus += int(effect.Increase)
			case "speed":
				a.SpeedBonus += effect.Increase
			case "max_health":
				a.MaxHealthBonus += int(effect.Increase)
			case "regen":
				a.RegenPerSecond += int(effect.Increase)
			}
		}
	}

	// Apply effects from Trauma (Body Status)
	if a.BodyStatus != nil {
		maxH := float64(a.GetTotalMaxHealth())
		// Leg penalties
		legHP := float64(a.BodyStatus["l_leg"] + a.BodyStatus["r_leg"])
		if legHP < maxH * 0.25 { a.SpeedBonus -= 0.6 } else if legHP < maxH * 0.4 { a.SpeedBonus -= 0.3 }
		
		// Arm penalties
		armHP := float64(a.BodyStatus["l_arm"] + a.BodyStatus["r_arm"])
		if armHP < maxH * 0.2 { a.AttackBonus -= 8 }
		
		// Head penalties
		if float64(a.BodyStatus["head"]) < maxH * 0.1 { a.DefenseBonus -= 10; a.AttackBonus -= 5 }
	}

	if a.Trauma.LeftArmLost { a.AttackBonus -= 5 }
	if a.Trauma.RightArmLost { a.AttackBonus -= 5 }
	if a.Trauma.EyesLost > 0 { a.AttackBonus -= 5 * a.Trauma.EyesLost }
	if a.Trauma.BurnedAlive { a.MaxHealthBonus -= 30 }
	if a.Trauma.SpineBroken { a.MaxHealthBonus -= 20 }

	// Sync active weapon from "weapon" slot
	if weaponItem, ok := a.Slots["weapon"]; ok && weaponItem != nil {
		if weaponItem.Config != nil && weaponItem.Config.Combat != nil {
			a.Weapon = weaponItem.Config.Combat
		}
	} else {
		a.Weapon = a.BaseWeapon
	}
}

// EquipItem tries to equip the given object into its slot.
func (a *Actor) EquipItem(it *ItemInstance) bool {
	if it == nil || it.Config == nil || it.Config.Slot == "" {
		return false
	}

	current := a.Slots[it.Config.Slot]
	shouldEquip := false

	if current == nil {
		shouldEquip = true
	} else if it.Config.Type == "weapon" && current.Config.Type == "weapon" {
		curDmg := current.Config.Combat.Damage.Average()
		newDmg := it.Config.Combat.Damage.Average()
		if newDmg > curDmg { shouldEquip = true }
	} else {
		// Compare stat totals
		curStats := 0.0
		newStats := 0.0
		for _, e := range current.Config.Effects { curStats += e.Increase }
		for _, e := range it.Config.Effects { newStats += e.Increase }
		if newStats > curStats { shouldEquip = true }
	}

	if shouldEquip {
		if current != nil { a.Inventory = append(a.Inventory, current) }
		a.Slots[it.Config.Slot] = it

		// Remove from inventory
		for i, item := range a.Inventory {
			if item == it {
				a.Inventory = append(a.Inventory[:i], a.Inventory[i+1:]...)
				break
			}
		}

		a.UpdateEffects()
		return true
	}

	return false
}

// EvaluateUpgrade checks if the item is better than what the actor has equipped.
func (a *Actor) EvaluateUpgrade(it *ItemInstance) bool {
	if it == nil || it.Config == nil || it.Config.Slot == "" { return false }

	current := a.Slots[it.Config.Slot]
	if current == nil { return true }

	if it.Config.Type == "weapon" && current.Config.Type == "weapon" {
		curDmg := current.Config.Combat.Damage.Average()
		newDmg := it.Config.Combat.Damage.Average()
		return newDmg > curDmg
	}

	curStats := 0.0
	newStats := 0.0
	for _, e := range current.Config.Effects { curStats += e.Increase }
	for _, e := range it.Config.Effects { newStats += e.Increase }
	return newStats > curStats
}

// GetTotalWeight returns the total weight of everything carried and equipped.
func (a *Actor) GetTotalWeight() float64 {
	total := 0.0
	for _, item := range a.Inventory { if item != nil { total += item.Weight } }
	for _, item := range a.Slots { if item != nil { total += item.Weight } }
	return total
}

func (a *Actor) CanCarry(weight float64) bool {
	return a.GetTotalWeight()+weight <= a.MaxWeight
}

// LoadEquipment loads items from Config.Equipment map into Slots and Config.Inventory array into Inventory.
func (a *Actor) LoadEquipment(objRegistry *ObjectRegistry) {
	if a.Config == nil || objRegistry == nil { return }
	a.Inventory = nil
	if a.Slots == nil { a.Slots = make(map[string]*ItemInstance) }

	for slotName, objID := range a.Config.Equipment {
		if obj, ok := objRegistry.Objects[objID]; ok {
			a.Slots[slotName] = NewItemInstance(obj.ID, obj, a.X, a.Y)
		}
	}

	for _, objID := range a.Config.Inventory {
		if obj, ok := objRegistry.Objects[objID]; ok {
			a.Inventory = append(a.Inventory, NewItemInstance(obj.ID, obj, a.X, a.Y))
		}
	}

	a.UpdateEffects()
}
