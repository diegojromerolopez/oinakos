package game

import "math"

func (c *Character) findTarget(player *Character, others []*Character, playerDist float64) (float64, float64, bool, bool) {
	var bestX, bestY float64; var hasTarget, isTargetPlayer bool; minDist := 15.0
	
	isTargetValid := func(other *Character) bool {
		if c.Relationships != nil {
			if sentiment, ok := c.Relationships[other.ID]; ok && sentiment < -20.0 {
				return true // Grudge!
			}
		}
		if c.Behavior == BehaviorChaotic { return true }
		
		// Hunting Logic: Hunters target animals regardless of alignment
		// We use "hunt" yield which is mapped correctly in GetAbilityYield
		if (c.Behavior == BehaviorKnightHunter || c.GetAbilityYield("hunt") > 40) && other.Config != nil && other.Config.IsAnimal {
			return true
		}
		
		if c.Behavior == BehaviorCriminal && other.IsAlive() {
			isVulnerable := other.State.Hunger > 60 || other.State.Thirst > 60 || other.State.Fatigue > 60 || other.State.IsDrunk
			isAlone := true
			for _, n := range others {
				if n != other && n != c && n.IsAlive() && math.Sqrt(math.Pow(other.X-n.X, 2)+math.Pow(other.Y-n.Y, 2)) < 10.0 { isAlone = false; break }
			}
			if isVulnerable || isAlone { return true }
		}

		if c.Alignment == AlignmentEnemy { return other.Alignment == AlignmentAlly || other.LeaderID != "" || other.Group != ""
		} else if c.Alignment == AlignmentAlly { return other.Alignment == AlignmentEnemy
		} else if c.Alignment == AlignmentNeutral { return c.TargetActor == &other.Actor }
		return false
	}
	
	if player != nil && player.IsAlive() && playerDist < minDist && isTargetValid(player) {
		minDist, bestX, bestY, hasTarget, isTargetPlayer, c.TargetActor = playerDist, player.X, player.Y, true, true, &player.Actor
	}
	for _, other := range others {
		if other == c || !other.IsAlive() { continue }
		dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
		if dist < minDist && isTargetValid(other) {
			minDist, bestX, bestY, hasTarget, isTargetPlayer, c.TargetActor = dist, other.X, other.Y, true, false, &other.Actor
		}
	}
	return bestX, bestY, hasTarget, isTargetPlayer
}

func (c *Character) findLootTarget(items []*ItemInstance) *ItemInstance {
	var best *ItemInstance; minDist := 10.0
	for _, it := range items {
		if !it.Pickable { continue }
		dist := math.Sqrt(math.Pow(c.X-it.X, 2) + math.Pow(c.Y-it.Y, 2))
		if dist < minDist { minDist, best = dist, it }
	}
	return best
}

func (c *Character) needsAIDecision(playerDist float64) bool {
	if playerDist < 10.0 || (c.State.HealthPoints < c.State.MaxHealthPoints/2 && playerDist < 20.0) {
		interval := 300
		if IsDebugEnabled() { interval = 60 }
		return (c.Tick - c.LastAIDecisionTick) >= interval
	}
	return false
}
