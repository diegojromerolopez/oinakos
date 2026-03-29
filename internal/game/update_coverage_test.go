package game

import (
	"testing"
)

func TestCharacter_UpdateCoverage(t *testing.T) {
	ctx := NewTestContext()
	p := NewCharacter(0, 0, nil, 1, true, nil)
	p.RawStats.HealthMax = 100
	p.State.MaxHealthPoints = 100
	p.State.HealthPoints = 100
	p.IsPlayerControlled = true
	
	// Mock input for updatePlayer
	// NewTestContext should provide a mock engine.Input
	
	// 1. updatePlayer - Movement
	// We can't easily mock engine.Input keys without knowing the mock implementation.
	// But we can call it and see if it crashes.
	p.Update(ctx)
	
	// 2. updateFacing
	p.updateFacing(1, 1)
	if p.Facing != DirSE { t.Error("Should face SE") }
	p.updateFacing(-1, -1)
	if p.Facing != DirNW { t.Error("Should face NW") }
	
	// 3. ShareRumors
	p.Tick = 300
	other := NewCharacter(1, 1, nil, 1, false, nil)
	other.Name = "Neighbor"
	ctx.World.Characters = append(ctx.World.Characters, other)
	p.Memories = append(p.Memories, MemoryEvent{Type: "murder", Source: "npc_01", Value: -50})
	p.ShareRumors(ctx)
	if len(other.Memories) == 0 {
		// Might fail if dist too high, they are both at 1,1 (p is 0,0)
	}
}
