package game

import (
	"math"
	"testing"
)

func TestCharacterAI_Targeting(t *testing.T) {
	ctx := setupTestGame()
	c := ctx.playableCharacter
	
	npc := NewCharacter(10.0, 0, nil, 1, false, nil)
	npc.Alignment = AlignmentEnemy
	ctx.World.Characters = []*Character{npc}
	
	// Test targeting player from far away
	tx, ty, has, isP := npc.findTarget(c, nil, 10.0)
	if !has || !isP || tx != c.X || ty != c.Y {
		t.Errorf("NPC failed to target player at range 10.0. has=%v, isP=%v", has, isP)
	}
	
	// Test targeting of another NPC
	ally := NewCharacter(5.0, 0, nil, 1, false, nil)
	ally.Alignment = AlignmentAlly
	tx, ty, has, isP = npc.findTarget(c, []*Character{ally}, 10.0)
	if !has || isP || tx != ally.X || ty != ally.Y {
		t.Errorf("NPC failed to prioritize closer ally NPC over further player. tx=%f, ty=%f, isP=%v", tx, ty, isP)
	}
}

func TestCharacterAI_Wander(t *testing.T) {
	setupTestGame()
	npc := NewCharacter(0, 0, nil, 1, false, nil)
	npc.Behavior = BehaviorWander
	npc.Speed = 0.5
	
	initialX, initialY := npc.X, npc.Y
	npc.updateAI(NewTestContext()) // Should wander since no targets
	
	if npc.X == initialX && npc.Y == initialY && npc.State != ActorWalking {
		t.Errorf("Wander behavior did not move NPC")
	}
}

func TestCharacterAI_Patrol(t *testing.T) {
	ctx := setupTestGame()
	ctx.World.PlayableCharacter = nil // Remove player to avoid targeting
	npc := NewCharacter(0, 0, nil, 1, false, nil)
	npc.Behavior = BehaviorPatrol
	npc.Speed = 1.0
	npc.PatrolStartX, npc.PatrolStartY = 0, 0
	npc.PatrolEndX, npc.PatrolEndY = 10, 10
	npc.PatrolHeading = true
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	for i := 0; i < 5; i++ {
		npc.updateAI(sysCtx)
	}
	
	if npc.X == 0 && npc.Y == 0 {
		t.Errorf("Patrol behavior did not move NPC after 5 ticks. Final Pos=(%f, %f)", npc.X, npc.Y)
	}
	
	// Snap to end and check heading flip
	npc.X, npc.Y = 10, 10
	npc.updateAI(sysCtx)
	if npc.PatrolHeading {
		t.Errorf("Patrol heading did not flip at target")
	}
}

func TestCharacterAI_Flee(t *testing.T) {
	ctx := setupTestGame()
	mc := ctx.playableCharacter
	mc.X, mc.Y = 5.0, 0.0
	mc.Health = 100
	mc.MaxHealth = 100
	
	npc := NewCharacter(0.0, 0.0, nil, 1, false, nil)
	npc.Alignment = AlignmentEnemy
	npc.Behavior = BehaviorFlee
	npc.Speed = 1.0
	npc.Health = 100
	npc.MaxHealth = 100
	npc.Speed = 1.0
	npc.TargetActor = &mc.Actor // Explicitly set target to flee from
	
	ctx.characters = []*Character{npc}
	ctx.World.Characters = ctx.characters
	ctx.World.PlayableCharacter = mc
	
	sysCtx := &SystemContext{
		World: ctx.World,
	}
	
	initialDist := math.Sqrt(math.Pow(npc.X-mc.X, 2) + math.Pow(npc.Y-mc.Y, 2))
	t.Logf("Initial dist: %f", initialDist)
	for i := 0; i < 5; i++ {
		npc.updateAI(sysCtx)
	}
	
	newDist := math.Sqrt(math.Pow(npc.X-mc.X, 2) + math.Pow(npc.Y-mc.Y, 2))
	if newDist <= initialDist {
		t.Errorf("Flee behavior did not increase distance from player after 5 ticks. initial=%f, new=%f", initialDist, newDist)
	}
}

func TestCharacterAI_Looting(t *testing.T) {
	ctx := setupTestGame()
	npc := NewCharacter(0, 0, nil, 1, false, nil)
	
	item := NewItemInstance("gold", &ObjectConfig{Name: "Gold"}, 5.0, 0)
	item.Pickable = true
	ctx.World.Items = []*ItemInstance{item}
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	
	// Update multiple times to move towards item
	for i := 0; i < 100; i++ {
		npc.updateAI(sysCtx)
		if len(ctx.World.Items) == 0 {
			break
		}
	}
	
	if len(ctx.World.Items) != 0 {
		t.Errorf("NPC failed to loot item. Pos: %f, %f", npc.X, npc.Y)
	}
}

func TestCharacterAI_Exhaustion(t *testing.T) {
	ctx := setupTestGame()
	npc := NewCharacter(0, 0, nil, 1, false, nil)
	npc.Energy = 5.0
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	
	npc.updateAI(sysCtx)
	if npc.State != ActorResting {
		t.Errorf("Low energy NPC did not enter Resting state. state=%v", npc.State)
	}
}

func TestCharacterAI_ApplyAIDecision(t *testing.T) {
	c := NewCharacter(0, 0, nil, 1, false, nil)
	
	dec := AIDecision{ChosenOption: "ATTACK the player", Reasoning: "Aggressiveness"}
	c.ApplyAIDecision(dec)
	if c.Behavior != BehaviorNpcFighter {
		t.Errorf("Expected BehaviorNpcFighter for attack, got %v", c.Behavior)
	}
	
	dec = AIDecision{ChosenOption: "FLEE please", Reasoning: "Fear"}
	c.ApplyAIDecision(dec)
	if c.Behavior != BehaviorFlee {
		t.Errorf("Expected BehaviorFlee for flee, got %v", c.Behavior)
	}
}
