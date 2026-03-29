package game

import (
	"fmt"
	"math"
	"strings"
	"oinakos/internal/engine"
)

func (c *Character) CheckAttackHits(ctx *SystemContext, skill string) {
	attackDist, hitSomething := 2.5, false
	if c.RawStats.AttackRange > 0 { attackDist = c.RawStats.AttackRange }
	if c.Weapon != nil { attackDist = c.Weapon.GetMaxDistance() }
	if c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging { attackDist = 5.0 }
	atX, atY := c.X, c.Y
	switch c.Facing { case DirSE: atX += attackDist; case DirSW: atY += attackDist; case DirNE: atY -= attackDist; case DirNW: atX -= attackDist }
	var hitCircle engine.Circle; var avgX, avgY float64
	if c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging { hitCircle = engine.Circle{X: c.X, Y: c.Y, Radius: attackDist}
	} else { avgX, avgY = (c.X+atX)*0.5, (c.Y+atY)*0.5; hitCircle = engine.Circle{X: avgX, Y: avgY, Radius: attackDist * 0.75} }
	
	targets := ctx.World.Characters; if ctx.World.PlayableCharacter != nil {
		found := false; for _, t := range targets { if t == ctx.World.PlayableCharacter { found = true; break } }
		if !found { targets = append([]*Character{ctx.World.PlayableCharacter}, targets...) }
	}
	for _, target := range targets {
		if target == c || (target.Alignment == c.Alignment && !c.IsPlayerControlled) || (!target.IsAlive() && c.ActionState != ActorChopping) { continue }
		checkX, checkY := avgX, avgY; if c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging { checkX, checkY = c.X, c.Y }
		if math.Sqrt(math.Pow(checkX-target.X, 2)+math.Pow(checkY-target.Y, 2)) < hitCircle.Radius { c.hitCharacter(&target.Actor, skill, ctx); hitSomething = true }
	}
	
	var bestTarget *Obstacle; bestDist := 999.0
	for _, o := range ctx.World.Obstacles {
		if !o.Alive || o.Archetype == nil || !o.Archetype.Destructible { continue }
		// Removed extra .Transformed(o.X, o.Y) because o.GetFootprint() is already transformed
		if !engine.CheckCirclePolygonCollision(hitCircle, o.GetFootprint()) && math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)) > attackDist { continue }
		
		if h := (c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging); h {
			typeID := strings.ToLower(o.Archetype.ID)
			isT, isB, isC := strings.Contains(typeID, "tree"), strings.Contains(typeID, "bush"), o.Archetype.IsCrop
			ok := (c.ActionState == ActorChopping && (isT || isC)) || (c.ActionState == ActorDigging && !isT && !isC) || (c.ActionState == ActorForaging && (isT || isB))
			if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); ok && d < bestDist { bestDist, bestTarget = d, o }
		} else if power := c.rollDamage(); power > 0 { o.TakeDamage(power); c.DegradeWeapon(ctx); hitSomething = true }
	}
	
	if o := bestTarget; o != nil {
		p := c.rollDamage()
		typeID := strings.ToLower(o.Archetype.ID)
		isT := strings.Contains(typeID, "tree")
		
		if isT && c.Weapon != nil && strings.Contains(strings.ToLower(c.Weapon.Name), "axe") && (c.ActionState == ActorChopping || c.ActionState == ActorAttacking) {
			if !c.CheckAbilitySuccess("chop", 0) { return }
			yield := c.GetAbilityYield("chop"); pEff := p + int(yield*0.5); if c.ActionState == ActorChopping { pEff *= 5 }
			o.WeightLeft -= float64(pEff); if o.WeightLeft < 0 { o.WeightLeft = 0 }
			
			// Drop wood item
			if wood := ctx.Registries.Objects.Objects["wood"]; wood != nil {
				it := NewItemInstance(wood.ID, wood, o.X, o.Y)
				ctx.World.Items = append(ctx.World.Items, it)
			}
			
			if o.WeightLeft <= 0 { o.Alive = false }
			hitSomething = true
		} else if o.Archetype.IsCrop && c.ActionState == ActorChopping && o.GrowthStage >= 2 {
			o.Alive = false
			itID := o.Archetype.Yield; if itID == "" { itID = "wheat" }
			if yObj := ctx.Registries.Objects.Objects[itID]; yObj != nil {
				it := NewItemInstance(yObj.ID, yObj, o.X, o.Y); ctx.World.Items = append(ctx.World.Items, it)
			}
			hitSomething = true
		} else if c.ActionState == ActorDigging && c.CheckAbilitySuccess("dig", 0) {
			o.TakeDamage(p + int(c.GetAbilityYield("dig")*0.4))
			if ctx.World.CurrentMapType != nil { ctx.World.CurrentMapType.Dig(o.X, o.Y, 0.5) }
			c.DegradeWeapon(ctx); hitSomething = true
		} else if c.ActionState == ActorForaging && c.CheckAbilitySuccess("forage", 0) {
			if item := ctx.Registries.Objects.Get(o.Archetype.Yield); item != nil {
				it := NewItemInstance(item.ID, item, o.X, o.Y); ctx.World.Items = append(ctx.World.Items, it); o.CooldownTicks, hitSomething = 600, true
			}
		}
	}
	
	if bestTarget == nil && c.ActionState == ActorDigging && ctx.World.CurrentMapType != nil {
		// Nothing hit, but we are digging - dig the ground itself
		ctx.World.CurrentMapType.Dig(atX, atY, 0.5)
		
		// Check for cave-in: if any neighbor vertex is > 6.0 units higher than current Z
		currentZ := ctx.World.CurrentMapType.GetElevationAt(atX, atY)
		neighbors := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
		for _, n := range neighbors {
			nx, ny := float64(int(atX)+n[0]), float64(int(atY)+n[1])
			nz := ctx.World.CurrentMapType.GetElevationAt(nx, ny)
			if nz-currentZ >= 6.0 {
				// Fatal cave-in!
				c.TakeDamage(9999, nil, ctx)
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("CAVE-IN! %s was buried alive.", c.Name), LogNPC) }
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "💀 CAVE-IN", X: c.X, Y: c.Y - 1, Life: 90, Color: ColorHarm })
				break
			}
		}
	}
	_ = hitSomething
}
