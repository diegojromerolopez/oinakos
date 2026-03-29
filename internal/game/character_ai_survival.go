package game

import (
	"math"
	"strings"
)

func (c *Character) handleSurvivalNeeds(ctx *SystemContext) bool {
	isHungry, isThirsty, isExhausted := c.State.Hunger > 70, c.State.Thirst > 70, c.State.Fatigue > 70
	if !(isHungry || isThirsty || isExhausted) || c.ActionState == ActorResting || c.ActionState == ActorForaging || c.ActionState == ActorEating || c.ActionState == ActorDrinking { return false }
	for i, item := range c.Inventory {
		if item != nil && item.Config != nil && item.Config.Type == "consumable" {
			if (isHungry && item.Config.Hunger > 0) || (isThirsty && item.Config.Thirst > 0) || (isExhausted && item.Config.Fatigue > 0) {
				if c.ConsumeItem(item, ctx) { c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...); return true }
			}
		}
	}
	safetyRadius := 5.0 + (float64(c.PrimaryAttributes.Intellect) * 0.1)
	for _, other := range ctx.World.Characters {
		if other.IsAlive() && other.Alignment != c.Alignment && math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)) < safetyRadius { return false }
	}
	searchRadius := 30.0 + (float64(c.PrimaryAttributes.Intellect) * 0.1)
	if isHungry || isThirsty {
		var nItem *ItemInstance; minIDist := searchRadius
		for _, item := range ctx.World.Items {
			if item.Config == nil || item.Config.Type != "consumable" || (item.Config.MaxHours > 0 && item.HoursLeft <= 0) { continue }
			if (isHungry && item.Config.Hunger <= 0) || (isThirsty && item.Config.Thirst <= 0) { continue }
			if d := math.Sqrt(math.Pow(c.X-item.X, 2)+math.Pow(c.Y-item.Y, 2)); d < minIDist { minIDist, nItem = d, item }
		}
		if nItem != nil {
			if minIDist < 1.0 { if ctx.World.Game != nil && ctx.World.Game.TryPickup(&c.Actor, nItem) { return true } } else { c.executeMovement(ctx, nItem.X-c.X, nItem.Y-c.Y, ctx.World.Obstacles, false); return true }
		}
	}
	if isHungry {
		var nForage *Obstacle; minFDist := searchRadius
		for _, o := range ctx.World.Obstacles {
			if !o.Alive { continue }
			if (strings.Contains(strings.ToLower(o.ID), "tree") || strings.Contains(strings.ToLower(o.ID), "bush")) && o.CooldownTicks <= 0 {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minFDist { minFDist, nForage = d, o }
			}
		}
		if nForage != nil {
			if minFDist < 1.5 { c.ActionState = ActorForaging; return true } else { c.executeMovement(ctx, nForage.X-c.X, nForage.Y-c.Y, ctx.World.Obstacles, false); return true }
		}
	}
	if isThirsty {
		var nWell *Obstacle; minWDist := 12.0
		for _, o := range ctx.World.Obstacles {
			if o.Alive && strings.Contains(strings.ToLower(o.ID), "well") {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minWDist { minWDist, nWell = d, o }
			}
		}
		if nWell != nil {
			if minWDist < 2.0 { c.ActionState = ActorDrinking; return true } else { c.MoveTo(ctx, nWell.X, nWell.Y); return true }
		}
	}
	if isExhausted {
		c.ActionState = ActorResting
		return true
	}
	return false
}
