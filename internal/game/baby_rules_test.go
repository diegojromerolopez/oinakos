package game

import (
	"testing"
)

func TestBabyRules(t *testing.T) {
	ctx := NewTestContext()
	
	mother := NewCharacter(10, 10, nil, 1, false, ctx.Registries.Objects)
	mother.Name = "Mother"
	mother.LifeStage = StageAdult
	mother.Alignment = AlignmentFriendly
	
	baby := NewCharacter(20, 20, nil, 1, false, ctx.Registries.Objects)
	baby.Name = "Baby"
	baby.LifeStage = StageBaby
	baby.ParentID = "Mother"
	baby.Alignment = AlignmentFriendly
	
	ctx.World.Characters = append(ctx.World.Characters, mother, baby)
	
	t.Run("Baby sticks to mother", func(t *testing.T) {
		baby.updateAI(ctx)
		// Baby should have snapped to 10,10 (mother's position)
		if baby.X != 10 || baby.Y != 10 {
			t.Errorf("Baby did not snap to mother. Pos: %v, %v", baby.X, baby.Y)
		}
	})
	
	t.Run("Baby cannot move independently", func(t *testing.T) {
		baby.X, baby.Y = 20, 20
		baby.executeMovement(ctx, 10, 10, nil, false)
		if baby.X != 20 || baby.Y != 20 {
			t.Errorf("Baby moved independently when it should be immobile")
		}
	})

	t.Run("Baby cannot attack", func(t *testing.T) {
		enemy := NewCharacter(11, 11, nil, 1, false, ctx.Registries.Objects)
		enemy.Alignment = AlignmentEnemy
		baby.TargetActor = &enemy.Actor
		baby.executeAttack(ctx, false, -9, -9)
		if baby.ActionState == ActorAttacking {
			t.Errorf("Baby should not be able to attack")
		}
	})
	
	t.Run("Parent retaliates when baby is attacked", func(t *testing.T) {
		attacker := NewCharacter(21, 21, nil, 1, false, ctx.Registries.Objects)
		attacker.Name = "Bad Guy"
		attacker.Alignment = AlignmentEnemy
		
		baby.TakeDamage(10, attacker, ctx)
		
		if mother.TargetActor == nil || mother.TargetActor.Name != "Bad Guy" {
			t.Errorf("Mother did not retaliate against baby's attacker")
		}
		if mother.Alignment != AlignmentEnemy {
			t.Errorf("Mother alignment should be Enemy (Hostile) when retaliating")
		}
	})
}
