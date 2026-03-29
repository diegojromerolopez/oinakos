package game

import (
	"testing"
)

func TestLeaderDeathConsequence(t *testing.T) {
	ctx := NewTestContext()
	// Setup
	leaderArch := &EntityConfig{ID: "queen_leader", Name: "Queen"}
	followerArch := &EntityConfig{ID: "guard_follower", Name: "Guard", LeaderID: "queen_leader"}
	
	leader := NewCharacter(0, 0, leaderArch, 1, false, nil)
	follower := NewCharacter(1, 1, followerArch, 1, false, nil)
	leader.TemporalState.HealthPoints = 100
	leader.TemporalState.MaxHealthPoints = 100
	follower.TemporalState.HealthPoints = 100
	follower.TemporalState.MaxHealthPoints = 100
	
	ctx.World.Characters = []*Character{leader, follower}
	mc := NewCharacter(10, 10, nil, 1, true, nil)
	mc.TemporalState.HealthPoints = 100
	mc.TemporalState.MaxHealthPoints = 100
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
	
	// Kill leader (irremediably)
	leader.TemporalState.HealthPoints = -10
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

	leader := NewCharacter(0, 0, leaderArch, 1, false, nil)
	leader.Alignment = AlignmentEnemy

	peer := NewCharacter(1, 1, followerArch, 1, false, nil)
	peer.Alignment = AlignmentEnemy
	peer.Behavior = BehaviorNpcFighter

	traitor := NewCharacter(2, 2, followerArch, 1, false, nil)
	traitor.Alignment = AlignmentNeutral // Switched!

	leader.TemporalState.HealthPoints = 100
	leader.TemporalState.MaxHealthPoints = 100
	peer.TemporalState.HealthPoints = 100
	peer.TemporalState.MaxHealthPoints = 100
	traitor.TemporalState.HealthPoints = 100
	traitor.TemporalState.MaxHealthPoints = 100

	ctx.World.Characters = []*Character{leader, peer, traitor}
	mc := NewCharacter(10, 10, nil, 1, true, nil)
	mc.TemporalState.HealthPoints = 100
	mc.TemporalState.MaxHealthPoints = 100
	ctx.World.PlayableCharacter = mc

	// Peer should normally ignore Neutral NPCs if they weren't traitors,
	// but because traitor has leader "queen" (Enemy), and Peer is Enemy,
	// Peer should target the traitor.
	peer.Update(ctx)

	if peer.TargetActor != &traitor.Actor {
		t.Errorf("Peer should have targeted the traitor, got %v", peer.TargetActor)
	}
}
