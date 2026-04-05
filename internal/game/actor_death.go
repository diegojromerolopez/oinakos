package game

import (
	"fmt"
	"math/rand"
	"strings"
)

func (a *Actor) die(attacker ActorInterface, ctx *SystemContext) {
	a.ActionState = ActorDead
	if ctx != nil && ctx.World != nil {
		var heir *Character
		for _, char := range ctx.World.Characters { if char.ParentID == a.Name && char.IsAlive() { heir = char; break } }
		if heir != nil {
			heir.Denarii += a.Denarii
			if a.OwnedChestID != "" && heir.OwnedChestID == "" { heir.OwnedChestID = a.OwnedChestID }
			if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s has inherited %d denarii from parent %s.", heir.Name, a.Denarii, a.Name), LogNPC) }
			a.Denarii = 0
		}
	}
	if ctx != nil && ctx.World != nil && ctx.World.Game != nil { 
		ctx.World.Game.DropAllItems(a) 
		// BUTCHERY PREPARATION (Elite Simulation: Resource Layer)
		if a.Config != nil && a.Config.IsAnimal {
			a.MeatQuantity = float64(a.Config.Meat)
			if a.MeatQuantity <= 0 { a.MeatQuantity = 1.0 } // Minimum meat if not specified
		}
	}
	prefix := "unknown"
	if a.Config != nil { 
		prefix = a.Config.SoundID; if prefix == "" { prefix = a.Config.ID }
		a.MeatQuantity = float64(a.Config.Meat)
	}
	if ctx != nil && ctx.Audio != nil { ctx.Audio.PlayRandomSound(prefix + "/death") }
	if attacker != nil {
		if act := attacker.GetActor(); act != nil {
			act.Kills++; if a.Config != nil {
				// Guilt Penalty: Killing humanoids (non-animals) harms the killer's sanity
				if !a.Config.IsAnimal { 
					penalty := 20.0
					if act.Behavior == BehaviorCriminal || act.ActionState == ActorBerserk {
						penalty = 5.0 // Hardened/Psychotic characters suffer less guilt
					}
					act.State.Sanity -= penalty 
				}
				act.MapKills[a.Config.ID]++; xp := a.Config.XP
				if xp <= 0 { xp = 1 }; act.AddXP(xp)
			}
			if act.Config != nil && act.Config.Actions != nil {
				for _, action := range act.Config.Actions.OnKill { if rand.Float64() < action.Probability { a.applyKillAction(action, attacker, ctx) } }
			}
		}
	}
}

func (a *Actor) applyKillAction(action KillAction, attacker ActorInterface, ctx *SystemContext) {
	if action.Type == "transform_victim" {
		e := action.Effect.Victim
		if e == nil { return }
		targetID := e.Transform
		if a.Config != nil { targetID = strings.ReplaceAll(targetID, "{gender}", a.Config.Gender) }
		var newConfig *EntityConfig; var ok bool
		if ctx != nil && ctx.Registries != nil {
			if ctx.Registries.Archetypes != nil { newConfig, ok = ctx.Registries.Archetypes.Archetypes[targetID] }
			if !ok && ctx.Registries.Characters != nil { newConfig, ok = ctx.Registries.Characters.Characters[targetID] }
		}
		if ok {
			a.Config, a.UnconsciousTimer, a.ActionState = newConfig, 0, ActorIdle
			a.State.HealthPoints = a.GetTotalMaxHealth(); a.InitBodyStatus()
			if e.Alignment == "inherit" { a.Alignment = attacker.GetActor().Alignment }
		}
	}
	if action.Type == "heal_attacker" || (action.Effect.Attacker != nil && action.Effect.Attacker.Heal > 0) {
		attk := attacker.GetActor()
		if action.Effect.Attacker != nil {
			attk.State.HealthPoints += action.Effect.Attacker.Heal
			if attk.State.HealthPoints > attk.GetTotalMaxHealth() { attk.State.HealthPoints = attk.GetTotalMaxHealth() }
		}
	}
}

func (a *Actor) SpawnDefecation(ctx *SystemContext) {
	if ctx == nil || ctx.World == nil || ctx.Registries.Obstacles == nil { return }
	config := ctx.Registries.Obstacles.Archetypes["defecation"]
	if config == nil { return }
	id := fmt.Sprintf("waste_%d_%d", ctx.World.DayTick, int(a.X*100))
	obs := NewObstacle(id, a.X, a.Y, config)
	ctx.World.Obstacles = append(ctx.World.Obstacles, obs)
}
