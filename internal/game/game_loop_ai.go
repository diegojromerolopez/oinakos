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
		if a.NPCID == "PLAYER" { g.playableCharacter.AIDecisionPending = false } else {
			for _, n := range g.characters { if n.Config != nil && (n.Config.ID == a.NPCID || n.Name == a.NPCID) { n.AIDecisionPending = false; break } }
		}

		if a.Decision.Err == nil {
			if a.NPCID == "PLAYER" {
				p := g.playableCharacter
				isVampire := p.State.Age.Rate == 0
				hungerLabel := "Hunger"
				if isVampire { hungerLabel = "Bloodlust" }
				
				log.Printf("[AI] Decision for PLAYER: %s | Reasoning: %s | Loc: (%.1f, %.1f) | HP: %d/%d | %s(H): %.1f | Age: %.1f", 
					a.Decision.ChosenOption, a.Decision.Reasoning, p.X, p.Y,
					p.State.HealthPoints, p.State.MaxHealthPoints,
					hungerLabel, p.State.Hunger, p.State.Age.Current)
				g.applyPlayerAIDecision(a.Decision)
				p.LastAIDecisionTick = g.CurrentTick()
			} else {
				for _, n := range g.characters {
					if n.Name == a.NPCID || (n.Config != nil && n.Config.ID == a.NPCID) {
						n.ApplyAIDecision(g.GetContext(), a.Decision); n.LastAIDecisionTick = g.CurrentTick(); break
					}
				}
			}
		}
	}
	
	interval := 60 
	p := g.playableCharacter
	isVampire := p.State.Age.Rate == 0
	
	// RICHEST FRONTIER LIFESTYLE 
	currentTick := g.CurrentTick()
	dayTick := currentTick % TicksPerDay
	isNight := dayTick > 12000 || dayTick < 4000 // Roman Night (Approx 10 PM to 6 AM)
	needsCompanion := (currentTick - p.LastCompanionTick) > (7 * TicksPerDay)
	wantsToGamble := (currentTick - p.LastGamblingTick) > (3 * TicksPerDay)

	isLongRunning := p.ActionState == ActorCooking || p.ActionState == ActorWorkshop || p.ActionState == ActorIntercourse || p.ActionState == ActorMilking
	if g.settings.AISimulationMode && !p.AIDecisionPending && (currentTick-p.LastAIDecisionTick) >= interval && !isLongRunning {
		var options []string
		g.updatePlayerSpatialMemory()

		// 1. ABSOLUTE CRITICAL SURVIVAL
		if p.State.Thirst > 90 || p.State.Hunger > 90 || p.State.Fatigue > 95 {
			if isVampire { 
				options = append(options, "feed") 
			} else {
				if p.State.Thirst > 90 { options = append(options, "drink") }
				if p.State.Hunger > 90 { options = append(options, "eat") }
			}
			if p.State.Fatigue > 95 { options = append(options, "rest") }
		} else {
			// 2. NOCTURNAL REST (Elite Simulation: Nightly Lodging)
			if isNight && p.State.Fatigue > 40 {
				options = []string{"rest", "wander"}
			} else {
				// 3. COMBAT AWARENESS & HUNTING (Awareness radius 50ft)
				var nearestTarget *Character; minTDist := 50.0
				for _, n := range g.characters {
					if n == nil || n == p || !n.IsAlive() { continue }
					isHuntingTarget := (p.GetAbilityYield("hunt") > 40 && n.Config != nil && n.Config.IsAnimal)
					isEnemy := n.Alignment != p.Alignment && n.Alignment != AlignmentNeutral
					if isEnemy || isHuntingTarget {
						dist := math.Sqrt(math.Pow(n.X-p.X, 2) + math.Pow(n.Y-p.Y, 2))
						if dist < minTDist { minTDist, nearestTarget = dist, n }
					}
				}

				if nearestTarget != nil {
					if p.State.HealthPoints < 35 { options = []string{"flee", "defend"} } else {
						options = []string{"hunt", "attack_nearest", "defend", "loot_nearest"}
						if isVampire && (p.State.Hunger > 40 || p.State.Thirst > 40) { options = append(options, "feed") }
					}
				} else {
					// 4. SOCIAL & LEISURE (Weekly Companion & Daily Chat)
					var courtesan *Character
					var neighbor *Character
					isNearFortuneHome := g.IsNearFortuneHome(p)

					for _, n := range g.characters {
						if n == nil || n == p || !n.IsAlive() { continue }
						dist := math.Sqrt(math.Pow(n.X-p.X, 2) + math.Pow(n.Y-p.Y, 2))
						if dist < 30.0 {
							if strings.Contains(strings.ToLower(n.Config.ID), "courtesan") { courtesan = n } else { neighbor = n }
						}
					}

					if needsCompanion && courtesan != nil && p.Denarii >= 25 {
						options = []string{"mate", "trade", "wander"} // Professional companionship
					} else if wantsToGamble && isNearFortuneHome && p.Denarii >= 2 {
						options = []string{"gamble", "wander"} 
					} else if neighbor != nil && (currentTick - p.LastTalkTick) > 600 {
						options = []string{"greet", "wander"} // Daily social chatting
					} else if p.State.Thirst > 60 || p.State.Hunger > 60 || p.State.Fatigue > 60 {
						if isVampire { options = []string{"feed", "rest"} } else { options = []string{"drink", "eat", "rest"} }
					} else {
						// 5. SCAVENGING & MISSIONS
						minLDist := 15.0
						var bestItem *ItemInstance
						if g.World != nil {
							for _, it := range g.World.Items {
								if it.Pickable {
									if d := math.Sqrt(math.Pow(it.X-p.X, 2)+math.Pow(it.Y-p.Y, 2)); d < minLDist { minLDist, bestItem = d, it }
								}
							}
						}
						
						hasMeat := false
						for _, it := range p.Inventory { if it != nil && it.Config != nil && strings.Contains(strings.ToLower(it.Config.Name), "meat") { hasMeat = true; break } }

						if bestItem != nil { 
							options = []string{"loot_nearest", "hunt"} 
						} else if hasMeat && p.Denarii < 100 {
							options = []string{"trade", "hunt", "wander"}
						} else if g.currentMapType.Type == ObjSurvive {
							options = []string{"hunt", "trade", "wander"}
						} else { 
							options = []string{"hunt", "forage", "trade", "wander"} 
						}
						
						if p.State.Thirst > 40 { options = append(options, "drink") }
						if p.State.Hunger > 40 { options = append(options, "eat") }
						if p.State.Fatigue > 40 { options = append(options, "rest") }
						options = append(options, "wander")
					}
				}
			}
		}

		g.aiManager.RequestDecision(context.Background(), "PLAYER", BuildWorldContext(g, nil), options)
		p.AIDecisionPending, p.LastAIDecisionTick = true, currentTick
	}
}

func (g *Game) CurrentTick() int {
	return g.Tick
}

func (g *Game) updatePlayerSpatialMemory() {
	if g.World == nil || g.playableCharacter == nil { return }
	for _, o := range g.obstacles {
		if !o.Alive || o.ID == "" { continue }
		lID := strings.ToLower(o.ID)
		if strings.Contains(lID, "well") || strings.Contains(lID, "tree") || strings.Contains(lID, "portal") || strings.Contains(lID, "tavern") || strings.Contains(lID, "house") || strings.Contains(lID, "fortune_home") {
			if math.Sqrt(math.Pow(g.playableCharacter.X-o.X, 2)+math.Pow(g.playableCharacter.Y-o.Y, 2)) < 30.0 { 
				g.playableCharacter.AddMemory(g.Tick, "location", o.ID, 1.0)
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
