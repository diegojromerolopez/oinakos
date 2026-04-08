package game

import (
	"context"
	"oinakos/internal/engine"
	"testing"
)

func TestNPC_BehaviorMega(t *testing.T) {
	ctx := NewTestContext()
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", nil, nil, false, "0.1-test")
	ctx.World.Game = g
	
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	ctx.World.PlayableCharacter = mc
	
	// Create different NPCs with different behaviors
	hunter := NewCharacter(10, 10, &EntityConfig{ID: "hunter"}, 1, false, nil)
	hunter.Behavior = BehaviorKnightHunter
	ctx.World.Characters = append(ctx.World.Characters, hunter)
	
	patroller := NewCharacter(20, 20, &EntityConfig{ID: "patrol"}, 1, false, nil)
	patroller.Behavior = BehaviorPatrol
	patroller.PatrolStartX, patroller.PatrolStartY = 20, 20
	patroller.PatrolEndX, patroller.PatrolEndY = 25, 25
	ctx.World.Characters = append(ctx.World.Characters, patroller)
	
	escort := NewCharacter(5, 5, &EntityConfig{ID: "escort"}, 1, false, nil)
	escort.Behavior = BehaviorEscort
	escort.LeaderID = "mc"
	ctx.World.Characters = append(ctx.World.Characters, escort)

	flee := NewCharacter(1, 1, &EntityConfig{ID: "flee"}, 1, false, nil)
	flee.Behavior = BehaviorFlee
	ctx.World.Characters = append(ctx.World.Characters, flee)

	// Tick them
	for i := 0; i < 60; i++ {
		for _, npc := range ctx.World.Characters {
			npc.Update(ctx)
		}
	}

	// 1. AI Manager Poll
	m := NewAIManager(&NoopAIProvider{})
	m.RequestDecision(context.Background(), "npc1", "world state", []string{"A", "B"})
	m.Poll()

	// 2. Alignment checks
	hunter.Alignment = AlignmentEnemy
	mc.Alignment = AlignmentAlly
	if hunter.Alignment == mc.Alignment { t.Error("Hunter should be enemy") }
}

func TestNPC_CombatMega(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	ctx.World.PlayableCharacter = mc
	
	orc := NewCharacter(1, 1, &EntityConfig{ID: "orc"}, 1, false, nil)
	ctx.World.Characters = []*Character{orc}
	
	// Combat mechanics
	mc.CheckAttackHits(ctx, "")
	orc.TakeDamage(10, &mc.Actor, ctx)
	
	// Death
	orc.TakeDamage(1000, &mc.Actor, ctx)
	if orc.IsAlive() { t.Error("Orc should be dead") }
	
	// Level up
	mc.AddXP(500)
}
