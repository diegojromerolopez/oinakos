package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func (a *Actor) GetAbilityDamage(abilityID string) int {
	if a.Config == nil || a.Config.Abilities == nil { return a.BaseAttack }
	ability, ok := a.Config.Abilities[abilityID]
	if !ok { return a.BaseAttack }
	formula, m := ability.Damage, 0.0
	switch {
	case strings.HasPrefix(formula, "melee_attack * "): if _, err := fmt.Sscanf(formula, "melee_attack * %f", &m); err == nil { return int(float64(a.BaseAttack) * m) }
	case strings.HasPrefix(formula, "ranged_attack * "): if _, err := fmt.Sscanf(formula, "ranged_attack * %f", &m); err == nil { return int(float64(a.RangedAttack) * m) }
	case strings.HasPrefix(formula, "attack * "): if _, err := fmt.Sscanf(formula, "attack * %f", &m); err == nil { return int(float64(a.BaseAttack) * m) }
	}
	return a.BaseAttack
}

func (a *Actor) ResolveAbilityEffects(abilityID string, target *Actor, ctx *SystemContext) {
	if a.Config == nil || a.Config.Abilities == nil { return }
	ability, ok := a.Config.Abilities[abilityID]; if !ok { return }
	for _, effect := range ability.Effects {
		if effect.Probability > 0 && rand.Float64() > effect.Probability { continue }
		if effect.StunChance > 0 && rand.Float64() <= effect.StunChance {
			target.UnconsciousTimer, target.ActionState = int(effect.Duration*60), ActorIncapacitated
		}
		if effect.KnockbackDistance > 0 {
			dx, dy := target.X-a.X, target.Y-a.Y; dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 { target.X, target.Y = target.X+(dx/dist)*effect.KnockbackDistance, target.Y+(dy/dist)*effect.KnockbackDistance }
		}
		if effect.PoisonDamagePerSecond > 0 { target.State.IsPoisoned = true }
	}
}
