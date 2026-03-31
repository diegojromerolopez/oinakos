package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

const (
	MemoryTypeLocation = "location"
)

func (g *Game) applyPlayerAIDecision(dec AIDecision) {
	if g.playableCharacter == nil { return }
	p := g.playableCharacter
	choice := strings.ToLower(dec.ChosenOption)
	p.LastAIReasoning = dec.Reasoning
	p.TargetActorID, p.TargetObstacle, p.TargetItem = "", nil, nil
	
	if dec.Reasoning != "" { g.LogEvent(fmt.Sprintf("%s: %s", p.Name, dec.Reasoning), LogNPC) }

	// Sanity Check
	if p.State.Sanity < 15.0 && rand.Float64() < 0.1 {
		p.ActionState = ActorBerserk; g.LogEvent(fmt.Sprintf("%s has lost their mind!", p.Name), LogWarning); return
	}

	switch {
	case strings.Contains(choice, "move_to_objective"):
		g.handleEliteNavigation(g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y)
		
	case strings.Contains(choice, "attack"), strings.Contains(choice, "hunt"), strings.Contains(choice, "defend"):
		g.handleEliteCombat()
	
	case strings.Contains(choice, "feed"):
		g.handleEliteFeeding()
		
	case strings.Contains(choice, "flee"):
		g.handleEliteFlee()
		
	case strings.Contains(choice, "drink"):
		if !g.tryConsumingFromInventory("water", "canteen") { g.handleEliteResourceSeeking("well", "water") }
		
	case strings.Contains(choice, "eat"), strings.Contains(choice, "forage"):
		if !g.tryConsumingFromInventory("cooked_meat", "bread", "apple") {
			if g.hasItemInInventory("raw_meat") {
				if campfire := g.findNearestObstacle("campfire", "fire", 15.0); campfire != nil {
					g.handleEliteNavigation(campfire.X, campfire.Y)
					if math.Abs(p.X-campfire.X) < 1.5 { p.ActionState = ActorCooking }
					return
				}
			}
			g.handleEliteResourceSeeking("tree", "food")
		}
		
	case strings.Contains(choice, "loot"):
		g.handleEliteLooting()
		
	case strings.Contains(choice, "rest"):
		if !g.tryConsumingFromInventory("sleeping_bag") { g.handleEliteResourceSeeking("tavern", "house") } else { p.ActionState, p.Tick = ActorResting, 0 }
		
	default: 
		p.WanderDirX, p.WanderDirY, p.Path = rand.Float64()*2 - 1, rand.Float64()*2 - 1, nil
	}
}

func (g *Game) handleEliteNavigation(tx, ty float64) {
	p := g.playableCharacter
	if len(p.Path) == 0 || g.Tick % 120 == 0 { p.Path = g.FindAStarPath(p.X, p.Y, tx, ty) }
	if len(p.Path) == 0 { 
		dx, dy := tx-p.X, ty-p.Y; mag := math.Sqrt(dx*dx + dy*dy)
		if mag > 0 { p.WanderDirX, p.WanderDirY = (dx/mag)*5.0, (dy/mag)*5.0 }
	}
}

func (g *Game) handleEliteFeeding() {
	p := g.playableCharacter
	// Priority 1: Incapacitated or Unconscious victims
	var victims []*Character
	for _, n := range g.characters {
		if n == nil || n == p || !n.IsAlive() { continue }
		if n.ActionState == ActorIncapacitated || n.UnconsciousTimer > 0 { victims = append(victims, n) }
	}
	
	var best *Character; minDist := 50.0
	for _, v := range victims {
		if d := math.Sqrt(math.Pow(p.X-v.X, 2)+math.Pow(p.Y-v.Y, 2)); d < minDist { minDist, best = d, v }
	}
	
	// Priority 2: Any Enemy (must subdue first)
	if best == nil {
		for _, n := range g.characters {
			if n == nil || n == p || !n.IsAlive() || n.Alignment != AlignmentEnemy { continue }
			if d := math.Sqrt(math.Pow(p.X-n.X, 2)+math.Pow(p.Y-n.Y, 2)); d < minDist { minDist, best = d, n }
		}
	}

	if best != nil {
		p.TargetActorID = best.ID
		g.handleEliteNavigation(best.X, best.Y)
		if minDist < 1.0 {
			p.ActionState, p.Tick = ActorFeeding, 0
			p.LastReaction = "BLOOD! I NEED BLOOD!"
			g.LogEvent(fmt.Sprintf("%s is feeding on %s!", p.Name, best.Name), LogCombatDamage)
		}
	} else {
		// No victims? Wander in frustration
		p.WanderDirX, p.WanderDirY = rand.Float64()*2 - 1, rand.Float64()*2 - 1
	}
}

func (g *Game) handleEliteCombat() {
	p := g.playableCharacter
	var vips []*Character
	for _, n := range g.characters { if n != nil && n.MustSurvive && n.IsAlive() && n != p { vips = append(vips, n) } }
	var threatToVip *Character
	for _, v := range vips {
		for _, n := range g.characters {
			if n == nil || !n.IsAlive() || n.Alignment != AlignmentEnemy { continue }
			if math.Sqrt(math.Pow(n.X-v.X, 2)+math.Pow(n.Y-v.Y, 2)) < 5.0 { threatToVip = n; break }
		}
		if threatToVip != nil { break }
	}
	if threatToVip != nil {
		g.LogEvent(fmt.Sprintf("%s protecting ally!", p.Name), LogNPC); g.handleEliteNavigation(threatToVip.X, threatToVip.Y); p.TargetActorID = threatToVip.ID
	} else {
		var nearest *Character; minDist := 50.0
		for _, n := range g.characters {
			if n != nil && n != p && n.IsAlive() && n.Alignment == AlignmentEnemy {
				if d := math.Sqrt(math.Pow(p.X-n.X, 2)+math.Pow(p.Y-n.Y, 2)); d < minDist { minDist, nearest = d, n }
			}
		}
		if nearest != nil {
			p.TargetActorID = nearest.ID; g.handleEliteNavigation(nearest.X, nearest.Y)
			if minDist < 1.5 { p.WanderDirX, p.WanderDirY, p.ActionState, p.Tick = 0, 0, ActorAttacking, 0 }
		}
	}
}

func (g *Game) handleEliteFlee() {
	p := g.playableCharacter
	var enemy *Character; minDist := 15.0
	for _, n := range g.characters {
		if n != nil && n != p && n.IsAlive() && n.Alignment == AlignmentEnemy {
			if d := math.Sqrt(math.Pow(p.X-n.X, 2)+math.Pow(p.Y-n.Y, 2)); d < minDist { minDist, enemy = d, n }
		}
	}
	if enemy != nil {
		dx, dy := p.X-enemy.X, p.Y-enemy.Y; mag := math.Sqrt(dx*dx + dy*dy)
		p.WanderDirX, p.WanderDirY, p.Path = (dx/mag)*5.0, (dy/mag)*5.0, nil
	}
}

func (g *Game) handleEliteLooting() {
	p := g.playableCharacter
	var best *ItemInstance; minDist := 15.0
	if g.World == nil { return }
	for _, it := range g.World.Items {
		if it.Pickable {
			if d := math.Sqrt(math.Pow(p.X-it.X, 2)+math.Pow(p.Y-it.Y, 2)); d < minDist {
				if (it.Config != nil && it.Config.Type == "weapon" && p.EvaluateUpgrade(it)) || strings.Contains(it.Config.Name, "meat") { minDist, best = d, it }
			}
		}
	}
	if best != nil {
		p.TargetItem = best; g.handleEliteNavigation(best.X, best.Y)
		if minDist < 1.5 { if g.TryPickup(&p.Actor, best) { p.EquipItem(best) }; p.WanderDirX, p.WanderDirY = 0, 0 }
	}
}

func (g *Game) handleEliteResourceSeeking(id, family string) {
	p := g.playableCharacter
	minDist := 40.0; if p.State.Hunger > 80 { minDist = 1000.0 }
	var target *Obstacle
	for _, o := range g.obstacles {
		if o.Alive {
			lID := strings.ToLower(o.ID)
			if strings.Contains(lID, id) || strings.Contains(lID, family) {
				if d := math.Sqrt(math.Pow(p.X-o.X, 2)+math.Pow(p.Y-o.Y, 2)); d < minDist { minDist, target = d, o }
			}
		}
	}
	if target != nil {
		g.handleEliteNavigation(target.X, target.Y)
		if minDist < 2.0 {
			p.ActionState, p.Tick = ActorIdle, 0
			if strings.Contains(id, "well") { p.ActionState = ActorDrinking }
			if strings.Contains(id, "tree") { p.ActionState = ActorForaging }
			if strings.Contains(id, "tavern") || strings.Contains(id, "house") { p.ActionState = ActorResting }
		}
	}
}

func (g *Game) hasItemInInventory(name string) bool {
	for _, it := range g.playableCharacter.Inventory {
		if it != nil && it.Config != nil && strings.Contains(strings.ToLower(it.Config.Name), strings.ToLower(name)) { return true }
	}
	return false
}

func (g *Game) findNearestObstacle(id, family string, dist float64) *Obstacle {
	var best *Obstacle; minD := dist
	for _, o := range g.obstacles {
		if o.Alive {
			lID := strings.ToLower(o.ID)
			if strings.Contains(lID, id) || strings.Contains(lID, family) {
				if d := math.Sqrt(math.Pow(g.playableCharacter.X-o.X, 2)+math.Pow(g.playableCharacter.Y-o.Y, 2)); d < minD { minD, best = d, o }
			}
		}
	}
	return best
}

func (g *Game) tryConsumingFromInventory(names ...string) bool {
	p := g.playableCharacter
	for _, name := range names {
		for i, item := range p.Inventory {
			if item != nil && item.Config != nil && strings.Contains(strings.ToLower(item.Config.Name), name) {
				if p.ConsumeItem(item, g.GetContext()) {
					p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
					return true
				}
			}
		}
	}
	return false
}
