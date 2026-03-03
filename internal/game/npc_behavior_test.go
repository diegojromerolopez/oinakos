package game

import (
	"testing"
)

// Tests for NPC behavior branches (wander, fighter, chaotic, neutral, ally)

func TestNPCBehavior_Wander_SetsDirection(t *testing.T) {
	mc := NewMainCharacter(100, 100, nil)
	npc := NewNPC(0, 0, &Archetype{ID: "test"}, 1)
	npc.Behavior = BehaviorWander
	npc.Alignment = AlignmentEnemy
	npc.Speed = 0.1 // must be non-zero
	// Pre-set direction so movement is predictable
	npc.WanderDirX = 1.0
	npc.WanderDirY = 0.0

	audio := NewMockAudioManager()
	var projs []*Projectile
	var fts []*FloatingText

	for i := 0; i < 5; i++ {
		npc.Update(mc, nil, nil, &projs, &fts, 1000, 1000, audio)
	}

	if npc.X <= 0 {
		t.Error("Wander NPC with WanderDirX=1 should have moved in +X direction")
	}
}

func TestNPCBehavior_Fighter_TargetsNearestNPC(t *testing.T) {
	mc := NewMainCharacter(100, 100, nil) // Far away
	fighter := NewNPC(0, 0, &Archetype{ID: "fighter"}, 1)
	fighter.Behavior = BehaviorNpcFighter
	fighter.Alignment = AlignmentEnemy

	target := NewNPC(2, 0, &Archetype{ID: "target"}, 1)
	target.Alignment = AlignmentAlly

	audio := NewMockAudioManager()
	var projs []*Projectile
	var fts []*FloatingText
	allNPCs := []*NPC{fighter, target}

	for i := 0; i < 10; i++ {
		fighter.Update(mc, nil, allNPCs, &projs, &fts, 1000, 1000, audio)
	}

	if fighter.TargetNPC == nil {
		t.Error("Fighter NPC should have acquired a target NPC")
	}
}

func TestNPCBehavior_Chaotic_TargetsNearestActor(t *testing.T) {
	mc := NewMainCharacter(3, 0, nil) // Closer than farNPC
	chaotic := NewNPC(0, 0, &Archetype{ID: "chaotic"}, 1)
	chaotic.Behavior = BehaviorChaotic
	chaotic.Alignment = AlignmentEnemy

	farNPC := NewNPC(20, 0, &Archetype{ID: "far"}, 1)
	farNPC.Alignment = AlignmentEnemy

	audio := NewMockAudioManager()
	var projs []*Projectile
	var fts []*FloatingText

	chaotic.Update(mc, nil, []*NPC{chaotic, farNPC}, &projs, &fts, 1000, 1000, audio)

	// Player at dist 3, farNPC at dist 20 → chaotic should target player
	if chaotic.TargetPlayer == nil {
		t.Error("Chaotic NPC should target the nearest actor (player at dist 3)")
	}
}

func TestNPCBehavior_Neutral_DoesNotTargetPlayer(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	npc := NewNPC(1, 0, nil, 1)
	npc.Alignment = AlignmentNeutral

	audio := NewMockAudioManager()
	var projs []*Projectile
	var fts []*FloatingText

	for i := 0; i < 5; i++ {
		npc.Update(mc, nil, nil, &projs, &fts, 1000, 1000, audio)
	}

	if npc.TargetPlayer != nil {
		t.Error("Neutral NPC should never target the player")
	}
}

func TestNPCBehavior_Ally_FollowsPlayerWhenNoEnemies(t *testing.T) {
	mc := NewMainCharacter(10, 10, nil)
	ally := NewNPC(0, 0, &Archetype{ID: "ally"}, 1)
	ally.Alignment = AlignmentAlly
	ally.Speed = 0.2 // must be non-zero

	audio := NewMockAudioManager()
	var projs []*Projectile
	var fts []*FloatingText

	for i := 0; i < 20; i++ {
		ally.Update(mc, nil, []*NPC{ally}, &projs, &fts, 1000, 1000, audio)
	}

	// Ally should have moved toward the player (closer than initial dist ~14)
	dist := (ally.X-mc.X)*(ally.X-mc.X) + (ally.Y-mc.Y)*(ally.Y-mc.Y)
	if dist >= 200 { // initial dist^2 ≈ 200
		t.Errorf("Ally NPC should be moving toward player; dist²=%.1f", dist)
	}
}
