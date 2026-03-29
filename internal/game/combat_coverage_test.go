package game

import (
	"testing"
)

func TestActor_CombatCoverage(t *testing.T) {
	ctx := NewTestContext()
	p := NewCharacter(0, 0, nil, 1, true, nil)
	target := NewCharacter(1, 1, nil, 1, false, nil)
	
	// Ensure registries and configs are set
	if p.Config == nil { p.Config = &EntityConfig{} }
	if p.Config.Abilities == nil { p.Config.Abilities = make(map[string]Ability) }
	
	// 1. GetAbilityDamage
	p.Config.Abilities["test_melee"] = Ability{Damage: "melee_attack * 2.0"}
	p.Config.Abilities["test_ranged"] = Ability{Damage: "ranged_attack * 1.5"}
	p.Config.Abilities["test_legacy"] = Ability{Damage: "attack * 1.2"}
	
	p.BaseAttack = 10
	p.RangedAttack = 20
	
	if d := p.GetAbilityDamage("test_melee"); d != 20 { t.Errorf("Expected 20, got %d", d) }
	if d := p.GetAbilityDamage("test_ranged"); d != 30 { t.Errorf("Expected 30, got %d", d) }
	if d := p.GetAbilityDamage("test_legacy"); d != 12 { t.Errorf("Expected 12, got %d", d) }
	if d := p.GetAbilityDamage("unknown"); d != 10 { t.Errorf("Expected 10, got %d", d) }

	// 2. ResolveAbilityEffects
	p.Config.Abilities["test_fx"] = Ability{
		Effects: []AbilityEffect{
			{StunChance: 1.0, Duration: 1.0},
			{KnockbackDistance: 2.0},
			{PoisonDamagePerSecond: 5},
		},
	}
	p.ResolveAbilityEffects("test_fx", &target.Actor, ctx)
	if target.Actor.ActionState != ActorIncapacitated { t.Error("Expected target to be stunned") }
	if !target.State.IsPoisoned { t.Error("Expected target to be poisoned") }
	
	// 3. TakeDamage with sepsis risk
	// Mock animal attacker
	attacker := NewCharacter(0, 0, nil, 1, false, nil)
	attacker.Config = &EntityConfig{IsAnimal: true}
	target.State.HealthPoints = 100
	target.TakeDamage(10, &attacker.Actor, ctx)
}
