package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func (c *Character) updateAI(ctx *SystemContext) {
	if !ctx.Settings.AdultMode {
		if c.Behavior == BehaviorSlave || c.Behavior == BehaviorSlaver {
			c.Behavior = BehaviorWander
			c.MasterID = ""
			c.Alignment = AlignmentNeutral
		}
	}

	// DEBT DEFAULT LEDGER: If any independent debt is due and unpaid, become a slave to the specific lender
	if ctx.Settings.AdultMode && len(c.Debts) > 0 {
		stillInDebt := make([]Loan, 0)
		enslaved := false
		for _, loan := range c.Debts {
			if ctx.World.State.Ticks > loan.Deadline && !enslaved {
				if c.Denarii >= loan.Amount {
					// Auto-pay if possible
					c.Denarii -= loan.Amount
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s auto-paid a loan of %d denarii.", c.Name, loan.Amount), LogInfo) }
				} else {
					// Default: Enslavement to the SPECIFIC lender
					c.Behavior = BehaviorSlave
					c.MasterID = loan.LenderUID
					if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s was enslaved to UID %s for defaulting on %d debt.", c.Name, loan.LenderUID, loan.Amount), LogWarning) }
					enslaved = true
				}
			} else if !enslaved {
				stillInDebt = append(stillInDebt, loan)
			}
		}
		if enslaved {
			c.Debts = nil // Clear all debts as they are now a slave
		} else {
			c.Debts = stillInDebt
		}
	}

	// Emergency Hydration: If incapacitated but at a source, drink anyway (Last Gasp)
	if c.IsAlive() && c.State.Thirst > 80 && c.ActionState == ActorIncapacitated {
		atSource := false
		for _, o := range ctx.World.Obstacles {
			if o.Alive && strings.Contains(strings.ToLower(o.ID), "well") && math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)) < 3.0 { atSource = true; break }
		}
		if !atSource && (c.CurrentTile == "water.png" || strings.Contains(c.CurrentTile, "water")) { atSource = true }
		if atSource { c.ActionState, c.Tick = ActorDrinking, 0; return }
	}

	if c.LifeStage == StageBaby {
		c.updateBabyAI(ctx, ctx.World.Obstacles)
		return
	}

	if c.TargetActor != nil && !c.TargetActor.IsAlive() { c.TargetActor = nil }
	
	// CRITICAL FIX: If we are already EATING or DRINKING, we must FINISH the action
	// even if the character is technically incapacitated. This prevents the rapid-switch loop
	// that causes thirst to stay at 100%.
	isConsuming := c.ActionState == ActorDrinking || c.ActionState == ActorEating
	if !c.IsAlive() || (c.IsIncapacitated() && c.ActionState != ActorBerserk && !isConsuming) { 
		return 
	}

	// Sanity Check
	// Leader death consequence check
	if c.LeaderID != "" {
		leaderDead := true
		for _, other := range ctx.World.Characters {
			if other.ID == c.LeaderID && other.IsAlive() {
				leaderDead = false
				break
			}
		}
		if leaderDead && c.Alignment == AlignmentEnemy {
			c.Alignment = AlignmentNeutral
			c.Behavior = BehaviorWander 
		}
	}

	c.EvaluateBerserk(ctx)

	// Action State progression (for all AI characters, including simulated player)
	// We exclude long-running actions (Cooking, Workshop, Intercourse, Milking) from this early timeout
	// because they have their own internal timer/completion logic (400+ ticks).
	isLongRunning := c.ActionState == ActorCooking || c.ActionState == ActorWorkshop || c.ActionState == ActorIntercourse || c.ActionState == ActorMilking
	if isLongRunning {
		// Only interrupt if absolutely dying
		if c.State.Thirst > 95 || c.State.Hunger > 95 || c.State.HealthPoints < 20 {
			// Falls through to handleSurvivalNeeds
		} else {
			return
		}
	}

	if c.ActionState == ActorDrinking || c.ActionState == ActorEating || c.ActionState == ActorRelieving || c.ActionState == ActorChopping || c.ActionState == ActorDigging || c.ActionState == ActorForaging || c.ActionState == ActorFeeding || c.ActionState == ActorCrouching {
		if c.Tick >= 60 { 
			if c.ActionState == ActorRelieving {
				c.AlleviateProperly(ctx)
			}
			c.ActionState = ActorIdle 
		}
		return
	}
	if c.ActionState == ActorAttacking {
		if c.Tick == 15 { c.CheckAttackHits(ctx, c.PendingSkill) }
		if c.Tick >= 30 { c.ActionState = ActorIdle; c.PendingSkill = "" }
		return
	}

	simMode := false
	if ctx.World != nil && ctx.World.Game != nil && ctx.World.Game.settings != nil {
		simMode = ctx.World.Game.settings.AISimulationMode
	}
	
	isAIPlayer := (c.IsPlayerControlled && simMode) || (c.ActionState == ActorBerserk && c.IsPlayerControlled)

	// 1. Survival Layer (Priority)
	if (isAIPlayer || !c.IsPlayerControlled) && !c.AIDecisionPending && c.handleSurvivalNeeds(ctx) {
		return
	}

	// 1a. Medical Support (Doctor/Cleric Specialization)
	if !c.IsPlayerControlled && c.Config != nil && (strings.Contains(strings.ToLower(c.Config.ID), "doctor") || strings.Contains(strings.ToLower(c.Config.ID), "cleric")) {
		var patient *Character; minDist := 15.0
		for _, other := range ctx.World.Characters {
			if other == c || !other.IsAlive() || other.Alignment == AlignmentEnemy { continue }
			if other.IsIncapacitated() {
				dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
				if dist < minDist { minDist, patient = dist, other }
			}
		}
		if patient != nil {
			if minDist < 1.5 {
				patient.Heal(25); patient.ActionState, patient.Tick = ActorIdle, 0
				if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has treated %s.", c.Name, patient.Name), LogNPC) }
				c.ActionState, c.Tick = ActorIdle, 0; return 
			}
			c.MoveTo(ctx, patient.X, patient.Y); return
		}
	}

	// 2. A* Path Following (Priority over WanderDir/Behaviors)
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

	// 2.5 Targeting and Looting Layer
	playableCharacter := ctx.World.PlayableCharacter
	playerDist := 999.0
	if playableCharacter != nil {
		pDx, pDy := playableCharacter.X-c.X, playableCharacter.Y-c.Y
		playerDist = math.Sqrt(pDx*pDx + pDy*pDy)
	}

	canLoot := c.MaxWeight > 0 && (c.Config == nil || !c.Config.IsAnimal)
	if c.Tick%30 == 0 && canLoot { c.TargetItem = c.findLootTarget(ctx.World.Items) }
	if c.TargetItem != nil && canLoot && c.TargetItem.Pickable {
		dx, dy := c.TargetItem.X-c.X, c.TargetItem.Y-c.Y
		if math.Sqrt(dx*dx+dy*dy) < 1.5 {
			if ctx.World.Game != nil && ctx.World.Game.TryPickup(&c.Actor, c.TargetItem) {
				c.EquipItem(c.TargetItem)
				itList := []*ItemInstance{}
				for _, it := range ctx.World.Items { if it != c.TargetItem { itList = append(itList, it) } }
				ctx.World.Items, c.TargetItem = itList, nil
				return
			}
			c.TargetItem, c.ActionState = nil, ActorIdle
		} else {
			c.executeMovement(ctx, dx, dy, ctx.World.Obstacles, false)
			return
		}
	}

	targetX, targetY, hasTarget, isTargetPlayer := c.findTarget(playableCharacter, ctx.World.Characters, playerDist)
	if hasTarget {
		dx, dy := targetX-c.X, targetY-c.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		attackRange := 1.4
		if c.RawStats.AttackRange > 0 { attackRange = c.RawStats.AttackRange }
		if c.Weapon != nil { attackRange = c.Weapon.GetMaxDistance() }
		canAttack := (c.TargetActor != nil && c.Alignment != c.TargetActor.Alignment) || c.Behavior == BehaviorChaotic
		if dist < attackRange && canAttack { 
			if dist < attackRange*0.5 && c.Weapon != nil && c.Weapon.IsRanged() { 
				c.executeMovement(ctx, dx, dy, ctx.World.Obstacles, true) 
			} else { 
				c.executeAttack(ctx, isTargetPlayer, dx, dy) 
			} 
		} else { 
			c.executeMovement(ctx, dx, dy, ctx.World.Obstacles, c.Behavior == BehaviorFlee) 
		}
		c.clampToMap(ctx.World.CurrentMapType.MapWidth, ctx.World.CurrentMapType.MapHeight)
		return
	}

	// 3. Shift-Based Behavior Layer (Map Goals / Roles)
	// PROTECTED STATES: Do not interrupt long-running tasks or survival relief unless shift change is urgent
	isBusy := c.ActionState != ActorIdle && c.ActionState != ActorWalking && c.ActionState != ActorBerserk && !isLongRunning
	if isBusy { return }

	// SHIFT OVERRIDE: Leisure and Rest shifts override standard labor
	if c.Shift == ShiftLeisure {
		// SOCIAL MAGNET: Graviate toward Tavern/Market during leisure
		var hub *Obstacle; minD := 300.0
		for _, o := range ctx.World.Obstacles {
			if o.Alive && (strings.Contains(strings.ToLower(o.ID), "tavern") || strings.Contains(strings.ToLower(o.ID), "market") || strings.Contains(strings.ToLower(o.ID), "inn")) {
				d := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
				if d < minD { minD, hub = d, o }
			}
		}
		if hub != nil && minD > 4.0 {
			c.executeMovement(ctx, hub.X-c.X, hub.Y-c.Y, ctx.World.Obstacles, false)
			return
		}
		// If already at hub or no hub, just wander/idle
		c.updateWander(ctx, ctx.World.Obstacles)
		return
	}

	if c.Shift == ShiftRest {
		// SLEEP OVERRIDE: Head to nearest bed or campfire
		var nRest *Obstacle; minDist := 400.0
		for _, o := range ctx.World.Obstacles {
			id := strings.ToLower(o.ID)
			if o.Alive && (strings.Contains(id, "campfire") || strings.Contains(id, "bed") || strings.Contains(id, "house") || strings.Contains(id, "tavern")) {
				if d := math.Sqrt(math.Pow(c.X-o.X, 2)+math.Pow(c.Y-o.Y, 2)); d < minDist { minDist, nRest = d, o }
			}
		}
		if nRest != nil {
			if minDist < 2.0 { 
				c.ActionState, c.Tick = ActorResting, 0; return 
			}
			c.MoveTo(ctx, nRest.X, nRest.Y); return
		}
		// If no bed found, just rest here
		c.ActionState, c.Tick = ActorResting, 0; return
	}

	// Standard Labor/Behavior (Only during ShiftWork or fallback)
	if isAIPlayer {
		if c.ActionState == ActorIdle || c.ActionState == ActorWalking {
			if c.WanderDirX != 0 || c.WanderDirY != 0 {
				c.executeMovement(ctx, c.WanderDirX, c.WanderDirY, ctx.World.Obstacles, false)
			} else {
				c.ActionState = ActorIdle
			}
		}
	}
	
	obstacles := ctx.World.Obstacles
	if !c.IsPlayerControlled && c.Behavior == BehaviorWander && c.checkSlaverySeeking(ctx, obstacles) {
		return
	}

	switch c.Behavior {
	case BehaviorHauler:
		c.updateHauler(ctx, obstacles)
	case BehaviorLumberjack:
		c.updateLumberjack(ctx, obstacles)
	case BehaviorFarmer:
		c.updateFarmer(ctx, obstacles)
	case BehaviorArtisan:
		c.updateArtisan(ctx, obstacles)
	case BehaviorPatrol:
		c.updatePatrol(ctx, obstacles)
	case BehaviorChaos:
		c.updateChaotic(ctx, obstacles)
	case BehaviorTrader:
		c.updateWander(ctx, obstacles)
	case BehaviorCriminal:
		c.updateCriminal(ctx, obstacles)
	case BehaviorSlave:
		c.updateSlave(ctx, obstacles)
	case BehaviorSlaver:
		c.updateSlaver(ctx, obstacles)
	case BehaviorHunter:
		c.updateChaotic(ctx, obstacles)
	case BehaviorWander:
		if c.Alignment == AlignmentAlly && ctx.World.PlayableCharacter != nil {
			pc := ctx.World.PlayableCharacter
			dx, dy := pc.X-c.X, pc.Y-c.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 8.0 {
				c.TargetActor = &pc.Actor
				c.executeMovement(ctx, dx, dy, obstacles, false)
				return
			}
		}
		c.updateWander(ctx, obstacles)
	case BehaviorNpcFighter, BehaviorFlee, BehaviorEscort:
		c.updateChaotic(ctx, obstacles)
	default:
		if c.Alignment == AlignmentAlly && ctx.World.PlayableCharacter != nil {
			pc := ctx.World.PlayableCharacter
			dx, dy := pc.X-c.X, pc.Y-c.Y
			if math.Sqrt(dx*dx+dy*dy) > 8.0 && c.TargetActor == nil {
				c.TargetActor = &pc.Actor
				c.executeMovement(ctx, dx, dy, obstacles, false)
			} else {
				c.updateWander(ctx, obstacles)
			}
		} else if rand.Float64() < 0.05 && c.checkEconomicSeeking(ctx) {
			// Handled by economic seeking
		} else {
			c.updateWander(ctx, obstacles)
		}
	}

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
	case strings.Contains(choice, "eat"):
		c.ActionState, c.Tick = ActorEating, 0
	case strings.Contains(choice, "drink"):
		c.ActionState, c.Tick = ActorDrinking, 0
	case strings.Contains(choice, "feed"):
		c.ActionState, c.Tick = ActorFeeding, 0
	case strings.Contains(choice, "hunt"):
		c.Behavior = BehaviorHunter
	case strings.Contains(choice, "forage"):
		c.ActionState, c.Tick = ActorForaging, 0
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
