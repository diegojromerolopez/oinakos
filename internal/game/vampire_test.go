package game

import (
	"testing"
)

func newBool(b bool) *bool { return &b }

func TestVampireConversion(t *testing.T) {
	ctx := NewTestContext()
	// Setup
	vampArch := &EntityConfig{
		ID: "vampire_male",
		Name: "Vampire Male",
		Gender: "male",
	}
	vampArch.Stats.HealthMin = 50
	vampArch.Stats.HealthMax = 50
	vampArch.Actions = &ActionConfig{
		OnKill: []KillAction{
			{
				Type:        "transform_victim",
				Probability: 1.0,
				Effect: ActionEffect{
					Victim: &VictimEffect{
						Transform:   "vampire_{gender}",
						Alignment:   "inherit",
						SpawnCorpse: newBool(false),
					},
				},
			},
		},
	}
	humanArch := &EntityConfig{
		ID:     "peasant_male",
		Name:   "Peasant Male",
		Gender: "male",
	}
	humanArch.Stats.HealthMin = 10
	humanArch.Stats.HealthMax = 10

	ctx.Registries.Archetypes.Archetypes["vampire_male"] = vampArch
	ctx.Registries.Archetypes.Archetypes["peasant_male"] = humanArch

	vampire := NewCharacter(0, 0, vampArch, 1, false)
	vampire.Alignment = AlignmentEnemy

	victim := NewCharacter(1, 1, humanArch, 1, false)
	victim.Alignment = AlignmentNeutral
	victim.Health = 1

	ctx.World.Characters = []*Character{vampire, victim}

	// Act: Victim takes lethal damage from vampire
	victim.TakeDamage(10, vampire, ctx)

	// Assert
	if victim.Config.ID != "vampire_male" {
		t.Errorf("Expected victim to be converted to vampire_male, got %s", victim.Config.ID)
	}
	if victim.Alignment != AlignmentEnemy {
		t.Errorf("Expected converted vampire to inherit alignment ENEMY, got %v", victim.Alignment)
	}
	if victim.State != ActorIdle {
		t.Errorf("Expected converted vampire to be Idle, got %v", victim.State)
	}
	if victim.Health <= 0 {
		t.Error("Expected converted vampire to have health restored")
	}
}
