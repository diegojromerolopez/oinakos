package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func (c *Character) updateAI(ctx *SystemContext) {
	if c.TargetActor != nil && !c.TargetActor.IsAlive() { c.TargetActor = nil }
	if !c.IsAlive() || (c.IsIncapacitated() && c.ActionState != ActorBerserk) { return }

	// Sanity Check
	c.EvaluateBerserk(ctx)

	// 1. External AI Override (e.g. Simulation Player AI)
	simMode := false
	if ctx.World != nil && ctx.World.Game != nil && ctx.World.Game.settings != nil {
		simMode = ctx.World.Game.settings.AISimulationMode
	}
	
	if (c.IsPlayerControlled && simMode) || (c.ActionState == ActorBerserk && c.IsPlayerControlled) {
		// A* Path Following (Priority over WanderDir)
		if len(c.Path) > 0 {
			target := c.Path[0]
			dx, dy := target.X-c.X, target.Y-c.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 0.5 {
				c.Path = c.Path[1:]
			} else {
				c.executeMovement(ctx, dx, dy, ctx.World.Obstacles, false)
			}
			return
		}

		// Biological/Action State progression
		if c.ActionState == ActorDrinking || c.ActionState == ActorEating || c.ActionState == ActorResting || c.ActionState == ActorRelieving || c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging || c.ActionState == ActorCooking {
			if c.Tick >= 30 { c.ActionState = ActorIdle }
			return
		}
		if c.ActionState == ActorAttacking {
			if c.Tick == 15 { c.CheckAttackHits(ctx, c.PendingSkill) }
			if c.Tick >= 30 { c.ActionState = ActorIdle; c.PendingSkill = "" }
			return
		}
		
		// Fallback to simpler movement vectors (WanderDir)
		if c.WanderDirX != 0 || c.WanderDirY != 0 {
			c.executeMovement(ctx, c.WanderDirX, c.WanderDirY, ctx.World.Obstacles, false)
		} else {
			c.ActionState = ActorIdle
		}
		return
	}

	// 2. Default NPC AI Logic
	c.clampToMap(ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight)
}

func (c *Character) ApplyAIDecision(ctx *SystemContext, dec AIDecision) {
	c.AIDecisionPending, c.LastAIChoice, c.LastAIReasoning, c.ThoughtTimer = false, dec.ChosenOption, dec.Reasoning, 180
	choice := strings.ToLower(dec.ChosenOption)
	
	switch {
	case strings.Contains(choice, "attack"): 
		c.Behavior = BehaviorNpcFighter
	case strings.Contains(choice, "flee"): 
		c.Behavior = BehaviorFlee
	case strings.Contains(choice, "wander"): 
		c.Behavior = BehaviorWander
		c.WanderDirX, c.WanderDirY = rand.Float64()*2 - 1, rand.Float64()*2 - 1
	case strings.Contains(choice, "rest"): 
		c.ActionState, c.Tick = ActorResting, 0
	}
}

func (c *Character) EvaluateBerserk(ctx *SystemContext) {
	if c.IsPlayerControlled && c.State.Sanity < 10.0 {
		if c.ActionState != ActorBerserk && rand.Float64() < 0.05 {
			c.ActionState = ActorBerserk
			ctx.World.Game.LogEvent(fmt.Sprintf("%s has suffered a mental breakdown!", c.Name), LogWarning)
		}
	}
}
