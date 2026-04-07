package game

import (
	"fmt"
	"math"
	"math/rand"
)

func (c *Character) TakeDamage(amount int, attacker ActorInterface, ctx *SystemContext) {
	oldCfg := c.Config
	c.Actor.TakeDamage(amount, attacker, ctx)
	if c.IsAlive() && !c.IsPlayerControlled && c.Config == oldCfg {
		c.handleAIReaction(attacker, ctx)
	}
}

func (c *Character) handleAIReaction(attacker ActorInterface, ctx *SystemContext) {
	if attacker == nil || c.State.Thirst > 90 || c.State.Hunger > 90 { return }
	act := attacker.GetActor(); fmt.Printf("DEBUG: handleAIReaction() start. Current Alignment = %v\n", c.Alignment)
	c.TargetActor = act
	// Flee if HP is critically low (50) or if significantly outmatched
	if c.State.HealthPoints < 50 || float64(c.State.HealthPoints) < float64(act.State.HealthPoints)*0.2 { 
		fmt.Printf("DEBUG: handleAIReaction() FLEE triggered for %s. HP = %d, Attacker HP = %d\n", c.Name, c.State.HealthPoints, act.State.HealthPoints)
		c.Alignment, c.Behavior = AlignmentNeutral, BehaviorFlee
	} else {
		fmt.Printf("DEBUG: handleAIReaction() FIGHT triggered for %s. HP = %d, Attacker HP = %d\n", c.Name, c.State.HealthPoints, act.State.HealthPoints)
		c.Alignment, c.Behavior = AlignmentEnemy, BehaviorKnightHunter
		if c.Group != "" {
			for _, other := range ctx.World.Characters {
				if other == c || other.Alignment == AlignmentEnemy || !other.IsAlive() || other.Group != c.Group { continue }
				if math.Sqrt(math.Pow(c.X-other.X, 2)+math.Pow(c.Y-other.Y, 2)) < 20.0 { other.Alignment, other.Behavior, other.TargetActor = AlignmentEnemy, BehaviorKnightHunter, act }
			}
		}
	}
}

func (c *Character) executeAttack(ctx *SystemContext, isTargetPlayer bool, dx, dy float64) {
	if IsDebugEnabled() { DebugLog("NPC [%s] EXECUTE ATTACK on Target (Player: %v)", c.Name, isTargetPlayer) }
	if c.ActionState == ActorIncapacitated { return }
	if c.ActionState != ActorAttacking {
		if isTargetPlayer {
			if c.CheckAttributeSuccess("wisdom", 0) && c.Config != nil && c.Config.Dialogues != nil {
				if bark := c.Config.Dialogues.PickCombatBark(); bark != "" && ctx.Log != nil { gLog(ctx, fmt.Sprintf("%s: %s", c.Name, bark)) }
			}
			if rand.Float64() < 0.3 && ctx.Audio != nil && c.Config != nil { ctx.Audio.PlayRandomSound(c.Config.SoundID + "/attack") }
		}
		c.ActionState, c.Tick = ActorAttacking, 0
	}
	if c.AttackTimer >= c.AttackCooldown {
		c.AttackTimer = 0
		skill := c.PendingSkill
		if !c.IsPlayerControlled && c.TargetActor != nil {
			if c.TargetActor.IsIncapacitated() { skill = "torture" } else if isTargetPlayer && rand.Float64() < 0.15 { skill = "restrain" }
		}
		if c.Weapon != nil && c.Weapon.IsRanged() && skill == "" {
			if mag := math.Sqrt(dx*dx + dy*dy); mag > 0 {
				pSpd := c.RawStats.ProjectileSpeed; if pSpd <= 0 { pSpd = 0.5 }
				ctx.World.AddProjectile(NewProjectile(c.X, c.Y, dx/mag, dy/mag, pSpd, c.GetTotalAttack(), false, 100.0))
			}
		} else { c.CheckAttackHits(ctx, skill) }
	}
}

func gLog(ctx *SystemContext, msg string) { if ctx.Log != nil { ctx.Log(msg, LogNPC) } }
