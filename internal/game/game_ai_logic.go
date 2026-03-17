package game

import (
	"math"
	"math/rand"
	"strings"
)

func (g *Game) applyPlayerAIDecision(dec AIDecision) {
	g.playableCharacter.AIDecisionPending = false
	g.playableCharacter.LastAIChoice = dec.ChosenOption
	g.playableCharacter.LastAIReasoning = dec.Reasoning
	choice := strings.ToLower(dec.ChosenOption)

	if strings.Contains(choice, "attack") {
		// Find target mentioned in reasoning or just nearest enemy
		var nearest *Character
		minDist := 999.0
		for _, n := range g.characters {
			if !n.IsAlive() || n.Alignment == AlignmentAlly || n.Alignment == AlignmentNeutral {
				continue
			}
			dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
			if dist < minDist {
				minDist = dist
				nearest = n
			}
		}
		if nearest != nil {
			g.playableCharacter.TargetActor = &nearest.Actor
		} else {
			g.playableCharacter.TargetActor = nil
			// Fallback to wander if no enemies
			g.playableCharacter.WanderDirX = rand.Float64()*2 - 1
			g.playableCharacter.WanderDirY = rand.Float64()*2 - 1
		}
	} else if strings.Contains(choice, "wander") {
		g.playableCharacter.TargetActor = nil
		g.playableCharacter.WanderDirX = rand.Float64()*2 - 1
		g.playableCharacter.WanderDirY = rand.Float64()*2 - 1
	} else if strings.Contains(choice, "defend") || strings.Contains(choice, "idle") {
		g.playableCharacter.TargetActor = nil
		g.playableCharacter.WanderDirX = 0
		g.playableCharacter.WanderDirY = 0
	} else if strings.Contains(choice, "flee") {
		// Move away from nearest enemy
		var nearest *Character
		minDist := 999.0
		for _, n := range g.characters {
			if !n.IsAlive() || n.Alignment == AlignmentAlly || n.Alignment == AlignmentNeutral {
				continue
			}
			dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
			if dist < minDist {
				minDist = dist
				nearest = n
			}
		}
		if nearest != nil {
			dx := g.playableCharacter.X - nearest.X
			dy := g.playableCharacter.Y - nearest.Y
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag > 0 {
				g.playableCharacter.TargetActor = nil
				g.playableCharacter.WanderDirX = (dx / mag) * 5.0
				g.playableCharacter.WanderDirY = (dy / mag) * 5.0
			}
		}
	} else if strings.Contains(choice, "move_to_objective") || strings.Contains(choice, "goal") || strings.Contains(choice, "portal") {
		g.playableCharacter.TargetActor = nil
		if g.currentMapType.TargetPoint.X != 0 || g.currentMapType.TargetPoint.Y != 0 {
			dx := g.currentMapType.TargetPoint.X - g.playableCharacter.X
			dy := g.currentMapType.TargetPoint.Y - g.playableCharacter.Y
			mag := math.Sqrt(dx*dx + dy*dy)
			if mag > 0 {
				g.playableCharacter.WanderDirX = (dx / mag) * 5.0
				g.playableCharacter.WanderDirY = (dy / mag) * 5.0
			}
		}
	}
}
