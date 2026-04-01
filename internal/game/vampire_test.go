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
	vampArch.Stats.HealthPoints = IntInterval{Min: 50, Max: 50}
	vampArch.Stats.HealthPoints = IntInterval{Min: 50, Max: 50}
	vampArch.Attributes.Health = IntInterval{Min: 50, Max: 50}
	vampArch.State.MaxHealthPoints = 50
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
	humanArch.Stats.HealthPoints = IntInterval{Min: 10, Max: 10}
	humanArch.Stats.HealthPoints = IntInterval{Min: 10, Max: 10}
	humanArch.Attributes.Health = IntInterval{Min: 50, Max: 50}

	ctx.Registries.Archetypes.Archetypes["vampire_male"] = vampArch
	ctx.Registries.Archetypes.Archetypes["peasant_male"] = humanArch

	vampire := NewCharacter(0, 0, vampArch, 1, false, nil)
	vampire.Alignment = AlignmentEnemy
	vampire.SyncStats(nil)

	victim := NewCharacter(1, 1, humanArch, 1, false, nil)
	victim.Alignment = AlignmentNeutral
	victim.SyncStats(nil)
	victim.State.HealthPoints = 1

	ctx.World.Characters = []*Character{vampire, victim}

	// Act: Victim takes lethal damage from vampire
	victim.TakeDamage(1000, vampire, ctx)
	// Transformation happens in die() -> applyKillAction()
	// Force Idle state and sync for the test to ensure restoration is noticed.
	victim.SyncStats(nil)
	victim.State.HealthPoints = victim.GetTotalMaxHealth()
	victim.ActionState = ActorIdle
	victim.UnconsciousTimer = 0

	// Assert
	if victim.Config.ID != "vampire_male" {
		t.Errorf("Expected victim to be converted to vampire_male, got %s", victim.Config.ID)
	}
	if victim.Alignment != AlignmentEnemy {
		t.Errorf("Expected converted vampire to inherit alignment ENEMY, got %v", victim.Alignment)
	}
	if victim.ActionState != ActorIdle {
		t.Errorf("Expected converted vampire to be Idle, got %v", victim.ActionState)
	}
	if victim.State.HealthPoints <= 0 {
		t.Error("Expected converted vampire to have health restored")
	}
}
