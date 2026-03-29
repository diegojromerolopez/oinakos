package game

import "fmt"

func (c *Character) Rest(ctx *SystemContext) {
	if !c.CheckAbilitySuccess("rest", 0) { return }
	if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s]: is resting", c.Name), LogNPC) }
	c.State.Fatigue += 10.0; if c.State.Fatigue > 100 { c.State.Fatigue = 100 }
	if ctx.World != nil { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Resting...", X: c.X, Y: c.Y - 1, Life: 60, Color: ColorHeal }) }
}

func (c *Character) Milk(target *Actor, ctx *SystemContext) {
	if target == nil || !target.Config.Stats.IsMilkable { return }
	if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s]: is milking %s", c.Name, target.Name), LogNPC) }
	if target.MilkCooldownTicks > 0 { return }
	if !c.CheckAbilitySuccess("milk", 0) {
		if ctx.World != nil { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Spilled!", X: target.X, Y: target.Y - 1, Life: 60, Color: ColorHarm }) }
		target.MilkCooldownTicks = 300; return
	}
	if cfg := ctx.Registries.Objects.Objects["milk"]; cfg != nil {
		it := NewItemInstance("milk", cfg, target.X, target.Y); ctx.World.Items = append(ctx.World.Items, it)
		target.MilkCooldownTicks = target.RawStats.MilkCooldown
		if ctx.World != nil { ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "+Milk", X: target.X, Y: target.Y - 1, Life: 60, Color: ColorHeal }) }
	}
}
