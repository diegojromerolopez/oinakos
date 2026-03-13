package game

import (
	"math"
	"math/rand"
)

func (n *NPC) findTarget(playableCharacter *PlayableCharacter, allNPCs []*NPC, playerDist float64) (targetX, targetY float64, hasTarget, isTargetPlayer bool) {
	// Override behavior based on alignment
	if n.Alignment == AlignmentAlly {
		const followDistThreshold = 8.0
		const rejoinDistTarget = 3.0
		const enemyDetectionRange = 12.0

		if playableCharacter != nil && playerDist > followDistThreshold {
			n.TargetActor = &playableCharacter.Actor
			return playableCharacter.X, playableCharacter.Y, true, true
		}
		
		nearestEnemy := gNearestEnemy(n, playableCharacter, allNPCs, enemyDetectionRange)
		if nearestEnemy != nil {
			n.TargetActor = &nearestEnemy.Actor
			return nearestEnemy.X, nearestEnemy.Y, true, false
		} else if playableCharacter != nil && playerDist > rejoinDistTarget {
			n.TargetActor = &playableCharacter.Actor
			return playableCharacter.X, playableCharacter.Y, true, true
		}
		n.TargetActor = nil
		n.State = NPCIdle
		return 0, 0, false, false
	}

	if n.Alignment == AlignmentNeutral {
		if n.Behavior == BehaviorFlee {
			if n.TargetActor != nil && n.TargetActor.IsAlive() {
				if playableCharacter != nil && n.TargetActor == &playableCharacter.Actor {
					isTargetPlayer = true
				}
				dx := n.X - n.TargetActor.X
				dy := n.Y - n.TargetActor.Y
				mag := math.Sqrt(dx*dx + dy*dy)
				if mag > 0 {
					return n.X + (dx/mag)*5.0, n.Y + (dy/mag)*5.0, true, isTargetPlayer
				}
				return n.X + 1, n.Y + 1, true, isTargetPlayer
			}
			n.Behavior = BehaviorWander
			n.TargetActor = nil
		}
		if n.Tick%120 == 0 {
			n.WanderDirX = rand.Float64()*2 - 1
			n.WanderDirY = rand.Float64()*2 - 1
		}
		return n.X + n.WanderDirX, n.Y + n.WanderDirY, true, false
	}

	// Reassess target for certain behaviors
	if n.Behavior == BehaviorChaotic || n.Behavior == BehaviorNpcFighter {
		n.TargetActor = nil
	}

	if n.TargetActor != nil && n.TargetActor.IsAlive() {
		if playableCharacter != nil && n.TargetActor == &playableCharacter.Actor {
			return n.TargetActor.X, n.TargetActor.Y, true, true
		}
		return n.TargetActor.X, n.TargetActor.Y, true, false
	}

	switch n.Behavior {
	case BehaviorKnightHunter:
		if playableCharacter != nil && playableCharacter.IsAlive() {
			n.TargetActor = &playableCharacter.Actor
			return playableCharacter.X, playableCharacter.Y, true, true
		}
	case BehaviorNpcFighter:
		var minDist = 999.0
		var nearestNPC *NPC
		for _, other := range allNPCs {
			if other == n || !other.IsAlive() || !n.isEnemy(other, allNPCs) {
				continue
			}
			dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
			if dist < minDist {
				minDist = dist
				nearestNPC = other
			}
		}
		if nearestNPC != nil {
			n.TargetActor = &nearestNPC.Actor
			return nearestNPC.X, nearestNPC.Y, true, false
		}
	case BehaviorChaotic:
		var minDist = 999.0
		var nearestActor *Actor
		var pDist = 999.0
		if playableCharacter != nil {
			pDist = math.Sqrt(math.Pow(n.X-playableCharacter.X, 2) + math.Pow(n.Y-playableCharacter.Y, 2))
		}
		for _, other := range allNPCs {
			if other == n || !other.IsAlive() {
				continue
			}
			dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
			if dist < minDist {
				minDist = dist
				nearestActor = &other.Actor
			}
		}
		if playableCharacter != nil && playableCharacter.IsAlive() && pDist < minDist {
			n.TargetActor = &playableCharacter.Actor
			return playableCharacter.X, playableCharacter.Y, true, true
		} else if nearestActor != nil {
			n.TargetActor = nearestActor
			return nearestActor.X, nearestActor.Y, true, false
		}
	case BehaviorEscort:
		if playableCharacter != nil {
			const seekEnemyRange = 12.0
			nearestEnemy := gNearestEnemy(n, playableCharacter, allNPCs, seekEnemyRange)
			if nearestEnemy != nil {
				n.TargetActor = &nearestEnemy.Actor
				return nearestEnemy.X, nearestEnemy.Y, true, false
			}
			if playerDist > 3.0 {
				n.TargetActor = &playableCharacter.Actor
				return playableCharacter.X, playableCharacter.Y, true, true
			}
			n.State = NPCIdle
		}
	case BehaviorWander:
		if n.Tick%120 == 0 {
			n.WanderDirX = rand.Float64()*2 - 1
			n.WanderDirY = rand.Float64()*2 - 1
			if n.Alignment != AlignmentEnemy && playableCharacter != nil && playerDist > 15.0 {
				dx := playableCharacter.X - n.X
				dy := playableCharacter.Y - n.Y
				mag := math.Sqrt(dx*dx + dy*dy)
				if mag > 0 {
					n.WanderDirX = (n.WanderDirX + (dx/mag)*0.5)
					n.WanderDirY = (n.WanderDirY + (dy/mag)*0.5)
				}
			}
		}
		return n.X + n.WanderDirX, n.Y + n.WanderDirY, true, false
	case BehaviorPatrol:
		if n.PatrolHeading {
			if math.Sqrt(math.Pow(n.X-n.PatrolEndX, 2)+math.Pow(n.Y-n.PatrolEndY, 2)) < 0.5 {
				n.PatrolHeading = false
				if rand.Float64() < 0.3 {
					n.PatrolStartX = n.X + (rand.Float64()*10 - 5)
					n.PatrolStartY = n.Y + (rand.Float64()*10 - 5)
				}
			}
			return n.PatrolEndX, n.PatrolEndY, true, false
		} else {
			if math.Sqrt(math.Pow(n.X-n.PatrolStartX, 2)+math.Pow(n.Y-n.PatrolStartY, 2)) < 0.5 {
				n.PatrolHeading = true
				if rand.Float64() < 0.3 {
					n.PatrolEndX = n.X + (rand.Float64()*10 - 5)
					n.PatrolEndY = n.Y + (rand.Float64()*10 - 5)
				}
			}
			return n.PatrolStartX, n.PatrolStartY, true, false
		}
	default:
		if n.Alignment == AlignmentEnemy {
			if playerDist < 20.0 {
				return playableCharacter.X, playableCharacter.Y, true, true
			}
			if n.Tick%120 == 0 {
				n.WanderDirX = rand.Float64()*2 - 1
				n.WanderDirY = rand.Float64()*2 - 1
			}
			return n.X + n.WanderDirX, n.Y + n.WanderDirY, true, false
		}
	}
	return 0, 0, false, false
}


func (n *NPC) executeMovement(dx, dy float64, obstacles []*Obstacle, isKite bool) {
	mag := math.Sqrt(dx*dx + dy*dy)
	if mag > 0 {
		ndx := dx / mag
		ndy := dy / mag
		if isKite {
			ndx, ndy = -ndx, -ndy
		}
		moveX := ndx * n.Speed * n.GetSpeedModifier()
		moveY := ndy * n.Speed * n.GetSpeedModifier()

		if !n.checkCollisionAt(n.X+moveX, n.Y+moveY, obstacles) {
			n.X += moveX
			n.Y += moveY
		} else {
			if !n.checkCollisionAt(n.X+moveX, n.Y, obstacles) { n.X += moveX }
			if !n.checkCollisionAt(n.X, n.Y+moveY, obstacles) { n.Y += moveY }
		}

		if !isKite {
			if ndx > 0 {
				if ndy < 0 { n.Facing = DirNE } else { n.Facing = DirSE }
			} else if ndx < 0 {
				if ndy < 0 { n.Facing = DirNW } else { n.Facing = DirSW }
			}
		}
	}
	n.State = NPCWalking
}

func (n *NPC) isEnemy(other *NPC, allNPCs []*NPC) bool {
	if other.LeaderID != "" {
		for _, potentialLeader := range allNPCs {
			if potentialLeader.Archetype != nil && potentialLeader.Archetype.ID == other.LeaderID && potentialLeader.IsAlive() {
				if other.Alignment != potentialLeader.Alignment && n.Alignment == potentialLeader.Alignment {
					return true
				}
				break
			}
		}
	}
	if n.Alignment == AlignmentNeutral || other.Alignment == AlignmentNeutral {
		return false
	}
	return n.Alignment != other.Alignment
}

func gNearestEnemy(n *NPC, playableCharacter *PlayableCharacter, allNPCs []*NPC, maxRange float64) *NPC {
	var nearest *NPC
	minDist := maxRange
	for _, other := range allNPCs {
		if other == n || !other.IsAlive() { continue }
		if n.isEnemy(other, allNPCs) {
			dist := math.Sqrt(math.Pow(n.X-other.X, 2) + math.Pow(n.Y-other.Y, 2))
			if dist < minDist {
				minDist = dist
				nearest = other
			}
		}
	}
	return nearest
}
