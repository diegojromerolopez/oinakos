package game

import (
	"log"
	"testing"
)

func TestSimulation_SocialAndPsychosis(t *testing.T) {
	ctx := NewTestContext()
	
	p := NewCharacter(0, 0, nil, 1, true, nil)
	p.Name = "Player"
	ctx.World.PlayableCharacter = p
	
	npc := NewCharacter(2, 2, nil, 1, false, nil)
	npc.Name = "NPC"
	npc.State.HealthPoints = 100 // Ensure alive
	ctx.World.Characters = append(ctx.World.Characters, npc)

	// 1. Memory and Sentiment
	npc.AddMemory(0, "gift", "Player", 50.0)
	tier := npc.GetRelationshipTier("Player")
	if tier != "Acquaintance" {
		t.Errorf("Expected Acquaintance after gift, got %s", tier)
	}

	// 2. Love Relationships
	p.ModifyRomanticInterest("NPC", 50.0)
	p.ModifySentiment("NPC", 35.0)
	tier = p.GetRelationshipTier("NPC")
	if tier != "Romantic" {
		t.Errorf("Expected Romantic tier, got %s", tier)
	}

	// 3. Psychosis (Berserk)
	npc.State.Sanity = 0
	npc.State.Hunger = 100 // Ensure sanity doesn't recover during SharedUpdate
	log.Printf("NPC State before: %s, Sanity: %.2f", npc.ActionState.String(), npc.State.Sanity)
	npc.SharedUpdate(ctx) // This should trigger berserk
	log.Printf("NPC State after: %s", npc.ActionState.String())
	
	if npc.ActionState != ActorBerserk {
		t.Errorf("Expected ActorBerserk when sanity is 0, got %s", npc.ActionState.String())
	}
	if npc.Alignment != AlignmentEnemy {
		t.Errorf("Expected AlignmentEnemy when berserk, got %d", npc.Alignment)
	}

	// Recovery
	npc.State.Sanity = 50
	npc.SharedUpdate(ctx)
	if npc.ActionState == ActorBerserk {
		t.Errorf("Expected recovery from Berserk when sanity restored")
	}
}

func TestSimulation_Workshop(t *testing.T) {
	ctx := NewTestContext()
	
	c := NewCharacter(0, 0, nil, 1, false, nil)
	c.PrimaryAttributes = PrimaryAttributes{Intellect: 100, Strength: 100} // Ensure 100% success
	c.Behavior = BehaviorArtisan
	ctx.World.Characters = append(ctx.World.Characters, c)
	
	// Create a workbench
	bench := &Obstacle{
		ID: "workbench_01",
		Archetype: &ObstacleArchetype{ID: "workbench"},
		Alive: true,
		X: 0.5, Y: 0.5, // Nearby
	}
	ctx.World.Obstacles = append(ctx.World.Obstacles, bench)

	// Degrade gear
	sword := &ItemInstance{Config: &ObjectConfig{ID: "sword", Resistance: 100}, Resistance: 10}
	c.Slots["body"] = sword

	// Try workshop
	c.ActionState = ActorWorkshop
	c.Tick = 0
	
	// Update 481 times
	for i := 0; i < 482; i++ {
		c.Update(ctx)
	}

	if sword.Resistance < 100 {
		t.Errorf("Expected gear to be repaired after workshop cycle, got %d", sword.Resistance)
	}
}
