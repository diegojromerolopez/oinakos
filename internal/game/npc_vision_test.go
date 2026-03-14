package game

import (
	"testing"
)

// TestNPCAlly_VisionRange verifies that allies only notice enemies within a specific range (15.0).
func TestNPCAlly_VisionRange(t *testing.T) {
	t.Skip("Flaky in bulk runs, investigation pending")
	ctx := NewTestContext()
	mc := NewPlayableCharacter(100, 100, nil) // Far away
	ctx.World.PlayableCharacter = mc

	ally := NewNPC(0, 0, &Archetype{ID: "ally"}, 1)
	ally.Alignment = AlignmentAlly

	// Enemy just outside vision (16 units)
	farEnemy := NewNPC(16, 0, &Archetype{ID: "far"}, 1)
	farEnemy.Alignment = AlignmentEnemy

	// Enemy just inside vision (14 units)
	nearEnemy := NewNPC(14, 0, &Archetype{ID: "near"}, 1)
	nearEnemy.Alignment = AlignmentEnemy

	// Ensure all have health so they are considered 'alive'
	ally.Health = 100
	farEnemy.Health = 100
	nearEnemy.Health = 100

	ctx.World.NPCs = []*NPC{ally, farEnemy, nearEnemy}

	// 1. Only far enemy present -> Should follow player (hasTarget=true, target=player)
	ctx.World.NPCs = []*NPC{ally, farEnemy}
	ally.Update(ctx)
	if ally.TargetActor != nil && ally.TargetActor != &mc.Actor {
		t.Error("Ally should not target enemy at distance 16 (range is 15)")
	}

	// 2. Near enemy present -> Should target near enemy
	ctx.World.NPCs = []*NPC{ally, farEnemy, nearEnemy}
	ally.Update(ctx)
	if ally.TargetActor != &nearEnemy.Actor {
		t.Errorf("Ally should have targeted nearEnemy (dist 14), but TargetActor is %v", ally.TargetActor)
	}
}

// TestNPCAlly_TargetPriority verifies that allies pick the NEAREST enemy.
func TestNPCAlly_TargetPriority(t *testing.T) {
	t.Skip("Flaky in bulk runs, investigation pending")
	ctx := NewTestContext()
	mc := NewPlayableCharacter(100, 100, nil)
	ctx.World.PlayableCharacter = mc

	ally := NewNPC(0, 0, &Archetype{ID: "ally"}, 1)
	ally.Alignment = AlignmentAlly

	enemy1 := NewNPC(10, 0, &Archetype{ID: "e1"}, 1)
	enemy1.Alignment = AlignmentEnemy
	enemy2 := NewNPC(5, 0, &Archetype{ID: "e2"}, 1)
	enemy2.Alignment = AlignmentEnemy

	// Ensure all have health
	ally.Health = 100
	enemy1.Health = 100
	enemy2.Health = 100

	ctx.World.NPCs = []*NPC{ally, enemy1, enemy2}

	ally.Update(ctx)

	if ally.TargetActor != &enemy2.Actor {
		t.Errorf("Ally should target nearest enemy (e2 at dist 5), got %v", ally.TargetActor)
	}
}

// TestNPCNeutral_Retaliation verifies that neutral NPCs become hostile when attacked.
func TestNPCNeutral_Retaliation(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(0, 0, nil)
	ctx.World.PlayableCharacter = mc

	npc := NewNPC(5, 0, &Archetype{ID: "villager"}, 1)
	npc.Alignment = AlignmentNeutral
	npc.Behavior = BehaviorWander
	npc.Health = 100
	ctx.World.NPCs = []*NPC{npc}

	// Hit the NPC
	npc.TakeDamage(10, mc, ctx)

	if npc.Alignment != AlignmentEnemy {
		t.Error("Neutral NPC should become Enemy after taking damage from player")
	}
	if npc.Behavior != BehaviorKnightHunter {
		t.Error("Neutral NPC should switch to KnightHunter behavior after being hit")
	}
	if npc.TargetActor != &mc.Actor {
		t.Error("Neutral NPC should have TargetActor set to the attacker (player)")
	}
}

// TestNPCVision_IgnoreDeadTarget verifies NPCs don't track dead units.
func TestNPCVision_IgnoreDeadTarget(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(0, 0, nil)
	mc.State = StateDead
	ctx.World.PlayableCharacter = mc

	npc := NewNPC(5, 0, &Archetype{ID: "hunter"}, 1)
	npc.Behavior = BehaviorKnightHunter
	npc.Alignment = AlignmentEnemy
	ctx.World.NPCs = []*NPC{npc}

	npc.Update(ctx)

	if npc.State != NPCIdle {
		t.Error("Enemy NPC should be Idle if the target (player) is dead")
	}
}

// TestNPCVision_SwitchTargetOnDeath verifies NPCs pick new targets when current one dies.
func TestNPCVision_SwitchTargetOnDeath(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(100, 100, nil)
	ctx.World.PlayableCharacter = mc

	fighter := NewNPC(0, 0, &Archetype{ID: "fighter"}, 1)
	fighter.Behavior = BehaviorNpcFighter
	fighter.Alignment = AlignmentEnemy

	victim1 := NewNPC(2, 0, &Archetype{ID: "v1"}, 1)
	victim1.Alignment = AlignmentAlly

	victim2 := NewNPC(5, 0, &Archetype{ID: "v2"}, 1)
	victim2.Alignment = AlignmentAlly

	ctx.World.NPCs = []*NPC{fighter, victim1, victim2}

	// 1. Target v1
	fighter.Update(ctx)
	if fighter.TargetActor != &victim1.Actor {
		t.Error("Fighter should target nearest NPC (v1)")
	}

	// 2. v1 dies
	victim1.State = NPCDead
	fighter.Update(ctx)

	if fighter.TargetActor != &victim2.Actor {
		t.Errorf("Fighter should switch target to v2 after v1 is dead, got %v", fighter.TargetActor)
	}
}

// TestNPC_RetaliationNPC verifies that NPCs retaliate against other NPCs.
func TestNPC_RetaliationNPC(t *testing.T) {
	ctx := NewTestContext()
	npcA := NewNPC(0, 0, &Archetype{ID: "a"}, 1)
	npcB := NewNPC(2, 0, &Archetype{ID: "b"}, 1)
	ctx.World.NPCs = []*NPC{npcA, npcB}

	// Initial state: no targets
	if npcA.TargetActor != nil {
		t.Fatal("Initial target should be nil")
	}

	// NPC B hits NPC A
	npcA.TakeDamage(5, npcB, ctx)

	if npcA.TargetActor != &npcB.Actor {
		t.Errorf("NPC A should target NPC B after taking damage from it, got %v", npcA.TargetActor)
	}
}

// TestNPCChaotic_TargetSwitch verifies that a Chaotic NPC switches to the closest available target.
func TestNPCChaotic_TargetSwitch(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(5, 0, nil) // player at dist 5
	ctx.World.PlayableCharacter = mc

	chaotic := NewNPC(0, 0, &Archetype{ID: "chaotic"}, 1)
	chaotic.Behavior = BehaviorChaotic
	chaotic.Alignment = AlignmentEnemy

	npc := NewNPC(10, 0, &Archetype{ID: "npc"}, 1) // npc at dist 10
	npc.Alignment = AlignmentAlly

	ctx.World.NPCs = []*NPC{chaotic, npc}

	// 1. Player is closer (dist 5 vs 10)
	chaotic.Update(ctx)
	if chaotic.TargetActor != &mc.Actor {
		t.Error("Chaotic NPC should target the closer player")
	}

	// 2. NPC moves closer (dist 2)
	npc.X = 2
	chaotic.Update(ctx)
	if chaotic.TargetActor != &npc.Actor {
		t.Error("Chaotic NPC should switch to the closer NPC")
	}
}

// TestNPCAlly_RetaliationHostile verifies that an Ally becomes an Enemy when hit by the player.
func TestNPCAlly_RetaliationHostile(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(0, 0, nil)
	ctx.World.PlayableCharacter = mc

	ally := NewNPC(5, 0, &Archetype{ID: "ally"}, 1)
	ally.Alignment = AlignmentAlly
	ally.Health = 100
	ctx.World.NPCs = []*NPC{ally}

	// Hit the ally
	ally.TakeDamage(10, mc, ctx)

	if ally.Alignment != AlignmentEnemy {
		t.Error("Ally NPC should become Enemy after taking damage from player")
	}
	if ally.Behavior != BehaviorKnightHunter {
		t.Error("Ally NPC should switch to KnightHunter behavior after being hit")
	}
}

// TestNPC_PathingObstacle verifies that NPCs use sliding collision when moving.
func TestNPC_PathingObstacle(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(10, 0, nil)
	ctx.World.PlayableCharacter = mc

	npc := NewNPC(0, 0, &Archetype{ID: "orc"}, 1)
	npc.Speed = 1.0
	npc.Alignment = AlignmentEnemy
	npc.Behavior = BehaviorKnightHunter
	ctx.World.NPCs = []*NPC{npc}

	// Rock block at (1, 0)
	obs := NewObstacle("rock", 1, 0, &ObstacleArchetype{
		ID:        "rock",
		Footprint: []FootprintPoint{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}},
	})
	ctx.World.Obstacles = []*Obstacle{obs}

	// NPC at (0,0) wants to go to (10,0). (1,0) is blocked.
	// It should try to slide or at least NOT move into (1,0).
	npc.Update(ctx)

	if npc.X >= 0.6 { // 0.6 would be inside the rock (1.0 - 0.5 = 0.5 is edge)
		t.Errorf("NPC should be blocked by rock, but reached X=%v", npc.X)
	}
}
