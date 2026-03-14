package game

import (
	"testing"
)

// Tests for NPC behavior branches (wander, fighter, chaotic, neutral, ally)

func TestNPCBehavior_Wander_SetsDirection(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(100, 100, nil)
	ctx.World.PlayableCharacter = mc

	npc := NewNPC(0, 0, &Archetype{ID: "test"}, 1)
	npc.Behavior = BehaviorWander
	npc.Alignment = AlignmentEnemy
	npc.Speed = 0.1 // must be non-zero
	// Pre-set direction so movement is predictable
	npc.WanderDirX = 1.0
	npc.WanderDirY = 0.0
	ctx.World.NPCs = []*NPC{npc}

	for i := 0; i < 5; i++ {
		npc.Update(ctx)
	}

	if npc.X <= 0 {
		t.Error("Wander NPC with WanderDirX=1 should have moved in +X direction")
	}
}

func TestNPCBehavior_Fighter_TargetsNearestNPC(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(100, 100, nil) // Far away
	ctx.World.PlayableCharacter = mc

	fighter := NewNPC(0, 0, &Archetype{ID: "fighter"}, 1)
	fighter.Behavior = BehaviorNpcFighter
	fighter.Alignment = AlignmentEnemy

	target := NewNPC(2, 0, &Archetype{ID: "target"}, 1)
	target.Alignment = AlignmentAlly
	ctx.World.NPCs = []*NPC{fighter, target}

	for i := 0; i < 10; i++ {
		fighter.Update(ctx)
	}

	if fighter.TargetActor == nil {
		t.Error("Fighter NPC should have acquired a target NPC")
	}
}

func TestNPCBehavior_Chaotic_TargetsNearestActor(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(3, 0, nil) // Closer than farNPC
	ctx.World.PlayableCharacter = mc

	chaotic := NewNPC(0, 0, &Archetype{ID: "chaotic"}, 1)
	chaotic.Behavior = BehaviorChaotic
	chaotic.Alignment = AlignmentEnemy

	farNPC := NewNPC(20, 0, &Archetype{ID: "far"}, 1)
	farNPC.Alignment = AlignmentEnemy
	ctx.World.NPCs = []*NPC{chaotic, farNPC}

	chaotic.Update(ctx)

	// Player at dist 3, farNPC at dist 20 → chaotic should target player
	if chaotic.TargetActor != &mc.Actor {
		t.Error("Chaotic NPC should target the nearest actor (player at dist 3)")
	}
}

func TestNPCBehavior_Neutral_DoesNotTargetActor(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(0, 0, nil)
	ctx.World.PlayableCharacter = mc

	npc := NewNPC(1, 0, nil, 1)
	npc.Alignment = AlignmentNeutral
	ctx.World.NPCs = []*NPC{npc}

	for i := 0; i < 5; i++ {
		npc.Update(ctx)
	}

	if npc.TargetActor != nil && npc.TargetActor == &mc.Actor {
		t.Error("Neutral NPC should never target the player")
	}
}

func TestNPCBehavior_Ally_FollowsPlayerWhenNoEnemies(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(10, 10, nil)
	ctx.World.PlayableCharacter = mc

	ally := NewNPC(0, 0, &Archetype{ID: "ally"}, 1)
	ally.Alignment = AlignmentAlly
	ally.Speed = 0.2 // must be non-zero
	ctx.World.NPCs = []*NPC{ally}

	for i := 0; i < 20; i++ {
		ally.Update(ctx)
	}

	// Ally should have moved toward the player (closer than initial dist ~14)
	dist := (ally.X-mc.X)*(ally.X-mc.X) + (ally.Y-mc.Y)*(ally.Y-mc.Y)
	if dist >= 200 { // initial dist^2 ≈ 200
		t.Errorf("Ally NPC should be moving toward player; dist²=%.1f", dist)
	}
}
