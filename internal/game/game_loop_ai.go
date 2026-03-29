package game

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
)

func (g *Game) updateAI() {
	if g.aiManager == nil { return }
	applied := g.aiManager.Poll()
	for _, a := range applied {
		if a.Decision.Err == nil {
			log.Printf("[AI] Decision for %s: %s", a.NPCID, a.Decision.ChosenOption)
			if a.NPCID == "PLAYER" { g.applyPlayerAIDecision(a.Decision); g.playableCharacter.LastAIDecisionTick = g.Tick
			} else {
				for _, n := range g.characters {
					if n.Name == a.NPCID || (n.Config != nil && n.Config.ID == a.NPCID) {
						if choice := strings.ToLower(a.Decision.ChosenOption); strings.Contains(choice, "talk") || strings.Contains(choice, "say") { g.LogEvent(fmt.Sprintf("%s: %s", n.Name, a.Decision.Reasoning), LogNPC) }
						n.ApplyAIDecision(g.GetContext(), a.Decision); n.LastAIDecisionTick = g.Tick; break
					}
				}
			}
		} else {
			if a.NPCID == "PLAYER" { g.playableCharacter.AIDecisionPending = false } else {
				for _, n := range g.characters { if n.Config.ID == a.NPCID || n.Name == a.NPCID { n.AIDecisionPending = false; break } }
			}
		}
	}
	interval := 300; if IsDebugEnabled() { interval = 60 }
	if g.settings.AISimulationMode && !g.playableCharacter.AIDecisionPending && (g.Tick-g.playableCharacter.LastAIDecisionTick) >= interval {
		g.aiManager.RequestDecision(context.Background(), "PLAYER", BuildWorldContext(g, nil), []string{"wander", "attack_nearest", "defend", "flee", "move_to_objective"})
		g.playableCharacter.AIDecisionPending, g.playableCharacter.LastAIDecisionTick = true, g.Tick
	}
	prob := g.settings.GetTalkingProbability()
	if prob > 0 && g.Tick%600 == 0 {
		for _, n := range g.characters {
			if !n.IsAlive() || n.AIDecisionPending { continue }
			if math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2)+math.Pow(n.Y-g.playableCharacter.Y, 2)) < 12.0 && rand.Float64() < prob {
				if g.aiManager != nil && g.settings.AIProvider != "none" {
					g.aiManager.RequestDecision(context.Background(), n.Name, BuildWorldContext(g, n), []string{"wander", "talk_to_player", "mutter_to_self"})
					n.AIDecisionPending, n.LastAIDecisionTick = true, g.Tick
				} else if n.Config != nil && n.Config.Dialogues != nil {
					if bark := n.PickIdleBark(); bark != "" { g.LogEvent(fmt.Sprintf("%s: %s", n.Name, bark), LogNPC) }
				}
				break
			}
		}
	}
}
