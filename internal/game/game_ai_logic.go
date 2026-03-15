package game

import "strings"

func (g *Game) applyPlayerAIDecision(dec AIDecision) {
	g.playableCharacter.AIDecisionPending = false
	choice := strings.ToLower(dec.ChosenOption)
	
	// Example mapping - simpler for now
	if strings.Contains(choice, "attack") {
		// Logic to find nearest enemy and move/attack
	} else if strings.Contains(choice, "wander") {
		// Logic to move randomly
	}
}
