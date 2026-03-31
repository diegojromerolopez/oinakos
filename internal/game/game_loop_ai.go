package game

import (
	"context"
	"log"
	"math"
	"strings"
)

func (g *Game) updateAI() {
	if g.aiManager == nil || g.playableCharacter == nil { return }
	
	applied := g.aiManager.Poll()
	for _, a := range applied {
		if a.Decision.Err == nil {
			if a.NPCID == "PLAYER" {
				p := g.playableCharacter
				isVampire := p.State.Age.Rate == 0
				hungerLabel := "Hunger"
				if isVampire { hungerLabel = "Bloodlust" }
				
				log.Printf("[AI] Decision for PLAYER: %s | Reasoning: %s | Loc: (%.1f, %.1f) | HP: %d/%d | %s(H): %.1f", 
					a.Decision.ChosenOption, a.Decision.Reasoning, p.X, p.Y,
					p.State.HealthPoints, p.State.MaxHealthPoints,
					hungerLabel, p.State.Hunger)
				g.applyPlayerAIDecision(a.Decision)
				p.LastAIDecisionTick = g.Tick
			} else {
				for _, n := range g.characters {
					if n.Name == a.NPCID || (n.Config != nil && n.Config.ID == a.NPCID) {
						n.ApplyAIDecision(g.GetContext(), a.Decision)
						n.LastAIDecisionTick = g.Tick
						break
					}
				}
			}
		} else {
			if a.NPCID == "PLAYER" { g.playableCharacter.AIDecisionPending = false } else {
				for _, n := range g.characters { if n.Config != nil && (n.Config.ID == a.NPCID || n.Name == a.NPCID) { n.AIDecisionPending = false; break } }
			}
		}
	}
	
	interval := 60 
	p := g.playableCharacter
	isVampire := p.State.Age.Rate == 0
	
	if g.settings.AISimulationMode && !p.AIDecisionPending && (g.Tick-p.LastAIDecisionTick) >= interval {
		var options []string
		g.updatePlayerSpatialMemory()

		// 1. ABSOLUTE CRITICAL SURVIVAL
		if p.State.Thirst > 85 || p.State.Hunger > 85 || p.State.Fatigue > 90 {
			if p.State.Fatigue > 90 { options = append(options, "rest") }
			if p.State.Thirst > 85 || p.State.Hunger > 85 {
				if isVampire { options = append(options, "feed") } else {
					if p.State.Thirst > 85 { options = append(options, "drink") }
					if p.State.Hunger > 85 { options = append(options, "eat") }
				}
			}
		} else {
			// 2. COMBAT AWARENESS
			var nearestEnemy *Character; minEDist := 15.0
			for _, n := range g.characters {
				if n == nil || n == p || !n.IsAlive() { continue }
				if n.Alignment != p.Alignment && n.Alignment != AlignmentNeutral {
					dist := math.Sqrt(math.Pow(n.X-p.X, 2) + math.Pow(n.Y-p.Y, 2))
					if dist < minEDist { minEDist, nearestEnemy = dist, n }
				}
			}

			if nearestEnemy != nil {
				if p.State.HealthPoints < 35 { options = []string{"flee", "defend"} } else {
					options = []string{"attack_nearest", "defend", "loot_nearest"}
					if isVampire && (p.State.Hunger > 40 || p.State.Thirst > 40) { options = append(options, "feed") }
				}
			} else {
				// 3. SECONDARY NEEDS & OBJECTIVES
				targetsMet := g.areMissionGoalsMet()
				
				if p.State.Thirst > 60 || p.State.Hunger > 60 || p.State.Fatigue > 60 {
					if isVampire { options = []string{"feed", "rest"} } else { options = []string{"drink", "eat", "rest"} }
				} else if p.State.HealthPoints < 60 {
					options = []string{"rest", "eat"}; if isVampire { options = []string{"rest", "feed"} }
				} else {
					// 4. SCAVENGING & MISSIONS
					minLDist := 12.0
					if g.World != nil {
						for _, item := range g.World.Items {
							if item.Pickable {
								if d := math.Sqrt(math.Pow(item.X-p.X, 2)+math.Pow(item.Y-p.Y, 2)); d < minLDist { minLDist = d }
							}
						}
					}
					
					if minLDist < 12.0 { options = []string{"loot_nearest", "hunt"} } else if targetsMet { options = []string{"move_to_objective"} } else { options = []string{"hunt", "forage", "wander"} }
					
					if isVampire { if p.State.Hunger > 25 { options = append(options, "feed") } } else {
						if p.State.Thirst > 25 { options = append(options, "drink") }
						if p.State.Hunger > 25 { options = append(options, "eat") }
					}
					if p.State.Fatigue > 25 { options = append(options, "rest") }
					options = append(options, "wander")
				}
			}
		}

		g.aiManager.RequestDecision(context.Background(), "PLAYER", BuildWorldContext(g, nil), options)
		p.AIDecisionPending, p.LastAIDecisionTick = true, g.Tick
	}
}

func (g *Game) updatePlayerSpatialMemory() {
	if g.World == nil || g.playableCharacter == nil { return }
	for _, o := range g.obstacles {
		if !o.Alive { continue }
		lID := strings.ToLower(o.ID)
		if strings.Contains(lID, "well") || strings.Contains(lID, "tree") || strings.Contains(lID, "portal") || strings.Contains(lID, "tavern") || strings.Contains(lID, "house") {
			if math.Sqrt(math.Pow(g.playableCharacter.X-o.X, 2)+math.Pow(g.playableCharacter.Y-o.Y, 2)) < 30.0 { 
				g.playableCharacter.AddMemory(g.Tick, MemoryTypeLocation, o.ID, 1.0)
			}
		}
	}
}

func (g *Game) areMissionGoalsMet() bool {
	if g.playableCharacter == nil { return true }
	enemyCount := 0
	for _, n := range g.characters {
		if n != nil && n.Alignment == AlignmentEnemy && n.IsAlive() { enemyCount++ }
	}
	return enemyCount == 0
}
