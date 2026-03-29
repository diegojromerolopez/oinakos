package game

import (
	"context"
	"math"
	"strings"
)

func (c *Character) updateAI(ctx *SystemContext) {
	if c.TargetActor != nil && !c.TargetActor.IsAlive() {
		c.TargetActor = nil
	}
	if !c.IsAlive() || c.IsIncapacitated() { return }
	sanityResilience := 1.5 - (float64(c.PrimaryAttributes.Wisdom) * 0.01); if sanityResilience < 0.5 { sanityResilience = 0.5 }
	if c.State.Sanity < (10.0 * sanityResilience) { c.updateChaotic(ctx, ctx.World.Obstacles); return }
	if c.HitTimer > 0 { c.HitTimer-- }
	effectiveShift := c.Shift; if c.LeaderID != "" {
		for _, char := range ctx.World.Characters { if char.Name == c.LeaderID && char.IsAlive() && char.Shift == ShiftWork && c.State.Fatigue > 15 { effectiveShift = ShiftWork; break } }
	}
	c.HandleSocial(ctx); playerDist := 999.0; playableCharacter := ctx.World.PlayableCharacter
	if playableCharacter != nil { playerDist = math.Sqrt(math.Pow(c.X-playableCharacter.X, 2) + math.Pow(c.Y-playableCharacter.Y, 2)) }
	if c.LeaderID != "" { leaderAlive := false; for _, char := range ctx.World.Characters { if char.Config != nil && char.Config.ID == c.LeaderID && char.IsAlive() { leaderAlive = true; break } }; if !leaderAlive { c.Alignment, c.Behavior, c.LeaderID = AlignmentNeutral, BehaviorWander, "" } }
	if effectiveShift == ShiftSleep && (c.State.Hunger <= 80 && c.State.Thirst <= 80) { c.updateSleepCycle(ctx); return }
	if c.handleSurvivalNeeds(ctx) { return }
	// Relieving needs seeking
	isUrgent := c.State.BladderLevel > 75 || c.State.BowelLevel > 75
	if isUrgent && c.ActionState != ActorRelieving {
		var nLatrine *Obstacle; minLDist := 30.0; for _, o := range ctx.World.Obstacles { if o.Alive && o.Archetype != nil && strings.Contains(strings.ToLower(o.ID), "latrine") { if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minLDist { minLDist, nLatrine = d, o } } }
		if nLatrine != nil { if minLDist < 1.5 { c.ActionState = ActorRelieving } else { c.MoveTo(ctx, nLatrine.X, nLatrine.Y) }; return }
	}
	if c.ActionState == ActorResting || c.ActionState == ActorDrinking { if c.State.Fatigue >= 100 || c.State.Thirst >= 100 { c.ActionState = ActorIdle } else { return } }
	if c.ActionState == ActorAttacking || c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging {
		if (c.ActionState == ActorAttacking && c.TargetActor == nil) || ((c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging) && c.TargetObstacle == nil && (c.ActionState != ActorDigging || c.X == 0)) {
			c.ActionState = ActorIdle
		} else {
			if c.Tick == 15 { c.CheckAttackHits(ctx, "") }
			if c.Tick >= 30 { c.ActionState = ActorIdle }
			return
		}
	}
	if ctx.AIManager != nil && !c.AIDecisionPending && c.needsAIDecision(playerDist) { ctx.AIManager.RequestDecision(context.Background(), c.Config.ID, BuildWorldContext(ctx.World.Game, c), []string{"attack", "flee", "wander", "patrol", "trade", "forage", "cook", "rest"}); c.AIDecisionPending = true }
	canLoot := c.MaxWeight > 0 && (c.Config == nil || !c.Config.IsAnimal); if c.Tick%30 == 0 && canLoot { c.TargetItem = c.findLootTarget(ctx.World.Items) }
	if c.TargetItem != nil && canLoot && c.TargetItem.Pickable {
		dx, dy := c.TargetItem.X-c.X, c.TargetItem.Y-c.Y; if math.Sqrt(dx*dx+dy*dy) < 1.5 { if ctx.World.Game != nil && ctx.World.Game.TryPickup(&c.Actor, c.TargetItem) { c.EquipItem(c.TargetItem); itList := []*ItemInstance{}; for _, it := range ctx.World.Items { if it != c.TargetItem { itList = append(itList, it) } }; ctx.World.Items, c.TargetItem = itList, nil; return }; c.TargetItem, c.ActionState = nil, ActorIdle } else { c.executeMovement(ctx, dx, dy, ctx.World.Obstacles, false); return }
	}
	targetX, targetY, hasTarget, isTargetPlayer := c.findTarget(playableCharacter, ctx.World.Characters, playerDist)
	if !hasTarget {
		switch c.Behavior { case BehaviorHauler: c.updateHauler(ctx, ctx.World.Obstacles); case BehaviorLumberjack: c.updateLumberjack(ctx, ctx.World.Obstacles); case BehaviorFarmer: c.updateFarmer(ctx, ctx.World.Obstacles); case BehaviorArtisan: c.updateArtisan(ctx, ctx.World.Obstacles); case BehaviorWander: c.updateWander(ctx, ctx.World.Obstacles); case BehaviorPatrol: c.updatePatrol(ctx, ctx.World.Obstacles); case BehaviorTrader: break; default: if c.Alignment == AlignmentAlly && playableCharacter != nil && playerDist > 5.0 && playerDist < 20.0 { c.executeMovement(ctx, playableCharacter.X-c.X, playableCharacter.Y-c.Y, ctx.World.Obstacles, false) } else { c.ActionState = ActorIdle } }
		c.clampToMap(ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight); return
	}
	dx, dy := targetX-c.X, targetY-c.Y; dist := math.Sqrt(dx*dx + dy*dy); attackRange := 1.4; if c.RawStats.AttackRange > 0 { attackRange = c.RawStats.AttackRange }; if c.Weapon != nil { attackRange = c.Weapon.GetMaxDistance() }
	canAttack := (c.TargetActor != nil && c.Alignment != c.TargetActor.Alignment) || c.Behavior == BehaviorChaotic
	if dist < attackRange && canAttack { if dist < attackRange*0.5 && c.Weapon != nil && c.Weapon.IsRanged() { c.executeMovement(ctx, dx, dy, ctx.World.Obstacles, true) } else { c.executeAttack(ctx, isTargetPlayer, dx, dy) } } else { c.executeMovement(ctx, dx, dy, ctx.World.Obstacles, c.Behavior == BehaviorFlee) }
	c.clampToMap(ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight)
}

func (c *Character) ApplyAIDecision(ctx *SystemContext, dec AIDecision) {
	c.AIDecisionPending, c.LastAIChoice, c.LastAIReasoning, c.ThoughtTimer = false, dec.ChosenOption, dec.Reasoning, 180
	choice := strings.ToLower(dec.ChosenOption)
	switch {
	case strings.Contains(choice, "attack"): c.Behavior = BehaviorNpcFighter
	case strings.Contains(choice, "flee"): c.Behavior = BehaviorFlee
	case strings.Contains(choice, "wander"), strings.Contains(choice, "talk"): c.Behavior = BehaviorWander
	case strings.Contains(choice, "patrol"): c.Behavior = BehaviorPatrol
	case strings.Contains(choice, "trade"): c.Behavior = BehaviorTrader
	case strings.Contains(choice, "forage"): c.ActionState, c.Tick = ActorForaging, 0
	case strings.Contains(choice, "cook"): c.ActionState, c.Tick = ActorCooking, 0
	case strings.Contains(choice, "rest"): c.ActionState, c.Tick = ActorResting, 0
	}
}
