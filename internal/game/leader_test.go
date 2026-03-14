package game

import (
	"testing"
)

func TestLeaderDeathConsequence(t *testing.T) {
	ctx := NewTestContext()
	// Setup
	leaderArch := &Archetype{ID: "queen_leader", Name: "Queen"}
	followerArch := &Archetype{ID: "guard_follower", Name: "Guard", LeaderID: "queen_leader"}
	
	leader := NewNPC(0, 0, leaderArch, 1)
	follower := NewNPC(1, 1, followerArch, 1)
	
	ctx.World.NPCs = []*NPC{leader, follower}
	mc := &PlayableCharacter{Actor: Actor{X: 10, Y: 10}} // Far away
	ctx.World.PlayableCharacter = mc
	
	// Initial state
	if follower.Alignment != AlignmentEnemy {
		t.Errorf("Follower should start as Enemy, got %v", follower.Alignment)
	}
	
	// Update follower while leader is alive
	follower.Update(ctx)
	if follower.Alignment != AlignmentEnemy {
		t.Errorf("Follower should stay Enemy while leader is alive, got %v", follower.Alignment)
	}
	
	// Kill leader
	leader.Health = 0
	leader.State = NPCDead
	
	// Update follower after leader death
	follower.Update(ctx)
	
	if follower.Alignment != AlignmentNeutral {
		t.Errorf("Follower should become Neutral after leader death, got %v", follower.Alignment)
	}
	
	if follower.Behavior != BehaviorWander {
		t.Errorf("Follower behavior should change to Wander, got %v", follower.Behavior)
	}
}

func TestTraitorTargeting(t *testing.T) {
	ctx := NewTestContext()
	// Setup: Leader (Enemy), Peer (Enemy), Traitor (Neutral)
	leaderArch := &Archetype{ID: "queen", Name: "Queen"}
	followerArch := &Archetype{ID: "guard", Name: "Guard", LeaderID: "queen"}

	leader := NewNPC(0, 0, leaderArch, 1)
	leader.Alignment = AlignmentEnemy

	peer := NewNPC(1, 1, followerArch, 1)
	peer.Alignment = AlignmentEnemy
	peer.Behavior = BehaviorNpcFighter

	traitor := NewNPC(2, 2, followerArch, 1)
	traitor.Alignment = AlignmentNeutral // Switched!

	ctx.World.NPCs = []*NPC{leader, peer, traitor}
	mc := &PlayableCharacter{Actor: Actor{X: 10, Y: 10}}
	ctx.World.PlayableCharacter = mc

	// Peer should normally ignore Neutral NPCs if they weren't traitors,
	// but because traitor has leader "queen" (Enemy), and Peer is Enemy,
	// Peer should target the traitor.
	peer.Update(ctx)

	if peer.TargetActor != &traitor.Actor {
		t.Errorf("Peer should have targeted the traitor, got %v", peer.TargetActor)
	}
}
