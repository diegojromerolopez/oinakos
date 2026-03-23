package game

import (
	"context"
	"math"
	"math/rand"
	"strings"
)

func (c *Character) updateAI(ctx *SystemContext) {
	worldObstacles, mapW, mapH := ctx.World.Obstacles, ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight
	if c.HitTimer > 0 { c.HitTimer-- }
	var playerDist float64
	playableCharacter := ctx.World.PlayableCharacter
	if playableCharacter != nil { playerDist = math.Sqrt(math.Pow(c.X-playableCharacter.X, 2) + math.Pow(c.Y-playableCharacter.Y, 2)) }

	if c.LeaderID != "" {
		leaderAlive := false
		for _, char := range ctx.World.Characters { if char.Config != nil && char.Config.ID == c.LeaderID && char.IsAlive() { leaderAlive = true; break } }
		if !leaderAlive { c.Alignment, c.Behavior, c.LeaderID = AlignmentNeutral, BehaviorWander, "" }
	}

	if c.Energy < 10 && c.State != ActorResting {
		isSafe := true
		for _, other := range ctx.World.Characters { if other.IsAlive() && other.Alignment != c.Alignment { if math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2)) < 10.0 { isSafe = false; break } } }
		if isSafe { c.State, c.Tick = ActorResting, 0 }
	}

	if c.State == ActorResting {
		wakeUp := c.Energy >= 100
		for _, other := range ctx.World.Characters { if other.IsAlive() && other.Alignment != c.Alignment { if math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2)) < 6.0 { wakeUp = true; break } } }
		if wakeUp { c.State = ActorIdle } else { return }
	}

	if c.State == ActorCrouching || c.IsTrulyDead() || c.IsIncapacitated() {
		if c.IsTrulyDead() { if c.DeadTimer == 0 { c.X, c.Y = findSafePosition(c.X, c.Y, c.GetCollisionCircle(), worldObstacles) }
			c.DeadTimer++
		}
		return
	}

	if c.State == ActorAttacking {
		if c.Tick == 15 { c.CheckAttackHits(ctx) }
		if c.Tick >= 30 { c.State = ActorIdle }
		return
	}
	if c.State == ActorChopping || c.State == ActorDigging {
		if c.Tick == 15 { c.CheckAttackHits(ctx) }
		if c.Tick >= 30 { c.State = ActorIdle }
		return
	}

	if ctx.AIManager != nil && !c.AIDecisionPending && c.needsAIDecision(playerDist) {
		worldCtx := BuildWorldContext(ctx.World.Game, c)
		options := []string{"attack", "flee", "wander", "patrol"}
		ctx.AIManager.RequestDecision(context.Background(), c.Config.ID, worldCtx, options)
		c.AIDecisionPending = true
	}

	if c.Tick%30 == 0 { c.TargetItem = c.findLootTarget(ctx.World.Items) }

	if c.TargetItem != nil && c.TargetItem.Pickable {
		dx, dy := c.TargetItem.X-c.X, c.TargetItem.Y-c.Y
		if math.Sqrt(dx*dx+dy*dy) < 1.5 {
			if ctx.World.Game != nil && ctx.World.Game.TryPickup(&c.Actor, c.TargetItem) {
				c.EquipItem(c.TargetItem)
				itList := []*ItemInstance{}
				for _, it := range ctx.World.Items { if it != c.TargetItem { itList = append(itList, it) } }
				ctx.World.Items, c.TargetItem = itList, nil
				return
			}
			c.TargetItem, c.State = nil, ActorIdle
		} else { c.executeMovement(ctx, dx, dy, worldObstacles, false); return }
	}

	targetX, targetY, hasTarget, isTargetPlayer := c.findTarget(playableCharacter, ctx.World.Characters, playerDist)
	if !hasTarget {
		if c.Behavior == BehaviorWander { c.updateWander(ctx, worldObstacles) } else if c.Behavior == BehaviorPatrol { c.updatePatrol(ctx, worldObstacles)
		} else if c.Alignment == AlignmentAlly && playableCharacter != nil && playerDist > 5.0 && playerDist < 20.0 { c.executeMovement(ctx, playableCharacter.X-c.X, playableCharacter.Y-c.Y, worldObstacles, false)
		} else { c.State = ActorIdle }
		c.clampToMap(mapW, mapH); return
	}

	dx, dy := targetX-c.X, targetY-c.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	attackRange := 1.4
	if c.Config != nil && c.Config.Stats.AttackRange > 0 { attackRange = c.Config.Stats.AttackRange }
	if c.Weapon != nil { attackRange = c.Weapon.GetMaxDistance() }

	canAttack := (c.TargetActor != nil && c.Alignment != c.TargetActor.Alignment) || c.Behavior == BehaviorChaotic
	tooClose := dist < attackRange*0.5 && c.Weapon != nil && c.Weapon.IsRanged()
	if tooClose && canAttack { c.executeMovement(ctx, dx, dy, worldObstacles, true) } else if dist < attackRange && canAttack { c.executeAttack(ctx, isTargetPlayer, dx, dy)
	} else { c.executeMovement(ctx, dx, dy, worldObstacles, c.Behavior == BehaviorFlee) }
	c.clampToMap(mapW, mapH)
}

func (c *Character) updateWander(ctx *SystemContext, obstacles []*Obstacle) {
	if c.Tick%120 == 0 || (c.WanderDirX == 0 && c.WanderDirY == 0) {
		angle := rand.Float64() * 2 * math.Pi
		c.WanderDirX, c.WanderDirY = math.Cos(angle), math.Sin(angle)
	}
	c.executeMovement(ctx, c.WanderDirX, c.WanderDirY, obstacles, false)
}

func (c *Character) updatePatrol(ctx *SystemContext, obstacles []*Obstacle) {
	targetX, targetY := c.PatrolEndX, c.PatrolEndY
	if !c.PatrolHeading { targetX, targetY = c.PatrolStartX, c.PatrolStartY }
	if math.Sqrt(math.Pow(c.X-targetX, 2) + math.Pow(c.Y-targetY, 2)) < 0.5 { c.PatrolHeading = !c.PatrolHeading } else { c.executeMovement(ctx, targetX-c.X, targetY-c.Y, obstacles, false) }
}

func (c *Character) findTarget(player *Character, others []*Character, playerDist float64) (float64, float64, bool, bool) {
	var bestX, bestY float64; var hasTarget, isTargetPlayer bool; minDist := 15.0
	isTargetValid := func(other *Character) bool {
		if c.Behavior == BehaviorChaotic { return true }
		if c.Alignment == AlignmentEnemy { return other.Alignment == AlignmentAlly || other.LeaderID != "" || other.Group != ""
		} else if c.Alignment == AlignmentAlly { return other.Alignment == AlignmentEnemy
		} else if c.Alignment == AlignmentNeutral { return c.TargetActor == &other.Actor }
		return false
	}
	if player != nil && player.IsAlive() && playerDist < minDist && isTargetValid(player) { minDist, bestX, bestY, hasTarget, isTargetPlayer, c.TargetActor = playerDist, player.X, player.Y, true, true, &player.Actor }
	for _, other := range others {
		if other == c || !other.IsAlive() { continue }
		dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
		if dist < minDist && isTargetValid(other) { minDist, bestX, bestY, hasTarget, isTargetPlayer, c.TargetActor = dist, other.X, other.Y, true, false, &other.Actor }
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
	if playerDist < 10.0 || (c.Health < c.MaxHealth/2 && playerDist < 20.0) {
		interval := 300
		if IsDebugEnabled() { interval = 60 }
		return (c.Tick - c.LastAIDecisionTick) >= interval
	}
	return false
}

func (c *Character) ApplyAIDecision(dec AIDecision) {
	c.AIDecisionPending, c.LastAIChoice, c.LastAIReasoning = false, dec.ChosenOption, dec.Reasoning
	choice := strings.ToLower(dec.ChosenOption)
	if strings.Contains(choice, "attack") { c.Behavior = BehaviorNpcFighter
	} else if strings.Contains(choice, "flee") { c.Behavior = BehaviorFlee
	} else if strings.Contains(choice, "wander") || strings.Contains(choice, "talk") { c.Behavior = BehaviorWander
	} else if strings.Contains(choice, "patrol") { c.Behavior = BehaviorPatrol }
}
