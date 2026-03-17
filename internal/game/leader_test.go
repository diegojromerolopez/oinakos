package game

import (
	"testing"
)

func TestLeaderDeathConsequence(t *testing.T) {
	ctx := NewTestContext()
	// Setup
	leaderArch := &EntityConfig{ID: "queen_leader", Name: "Queen"}
	followerArch := &EntityConfig{ID: "guard_follower", Name: "Guard", LeaderID: "queen_leader"}
	
	leader := NewCharacter(0, 0, leaderArch, 1, false)
	follower := NewCharacter(1, 1, followerArch, 1, false)
	
	ctx.World.Characters = []*Character{leader, follower}
	mc := NewCharacter(10, 10, nil, 1, true)
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
	leader.State = ActorDead
	
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
	leaderArch := &EntityConfig{ID: "queen", Name: "Queen"}
	followerArch := &EntityConfig{ID: "guard", Name: "Guard", LeaderID: "queen"}

	leader := NewCharacter(0, 0, leaderArch, 1, false)
	leader.Alignment = AlignmentEnemy

	peer := NewCharacter(1, 1, followerArch, 1, false)
	peer.Alignment = AlignmentEnemy
	peer.Behavior = BehaviorNpcFighter

	traitor := NewCharacter(2, 2, followerArch, 1, false)
	traitor.Alignment = AlignmentNeutral // Switched!

	ctx.World.Characters = []*Character{leader, peer, traitor}
	mc := NewCharacter(10, 10, nil, 1, true)
	ctx.World.PlayableCharacter = mc

	// Peer should normally ignore Neutral NPCs if they weren't traitors,
	// but because traitor has leader "queen" (Enemy), and Peer is Enemy,
	// Peer should target the traitor.
	peer.Update(ctx)

	if peer.TargetActor != &traitor.Actor {
		t.Errorf("Peer should have targeted the traitor, got %v", peer.TargetActor)
	}
}
