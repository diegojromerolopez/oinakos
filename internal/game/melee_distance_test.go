package game

import (
	"math"
	"testing"
)

// Test melee distance boundaries to ensure "touching distance" is correctly enforced.
func TestMeleeTouchingDistance(t *testing.T) {
	ctx := NewTestContext()
	
	// Create attacker (player) at (0,0) with a melee weapon (default 2.5 distance)
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.BaseAttack = 9999 // Ensure hit if in range
	mc.Facing = DirSE    // Facing +X
	ctx.World.PlayableCharacter = mc

	// The hit calculation uses:
	// attackDist = 2.5
	// checkX = mc.X + attackDist*0.5 = 1.25
	// checkY = mc.Y = 0
	// minDist = attackDist * 0.75 = 1.875
	// Max Reach = 1.25 + 1.875 = 3.125
	// Min Reach = 1.25 - 1.875 = -0.625

	tests := []struct {
		name     string
		npcX     float64
		npcY     float64
		expectHit bool
	}{
		{
			name:      "Just Inside Range",
			npcX:      3.1, // Less than 3.125
			npcY:      0,
			expectHit: true,
		},
		{
			name:      "Just Outside Range",
			npcX:      3.15, // More than 3.125
			npcY:      0,
			expectHit: false,
		},
		{
			name:      "Behind Attacker (Inside small back-reach)",
			npcX:      -0.5, // Between -0.625 and 0
			npcY:      0,
			expectHit: true, 
		},
		{
			name:      "Behind Attacker (Outside back-reach)",
			npcX:      -0.7, // Less than -0.625
			npcY:      0,
			expectHit: false,
		},
		{
			name:      "Extreme Side (Outside Range)",
			npcX:      1.25,
			npcY:      1.9, // Radius is 1.875, so 1.9 is outside
			expectHit: false,
		},
		{
			name:      "Extreme Side (Inside Range)",
			npcX:      1.25,
			npcY:      1.8, // Inside radius
			expectHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := NewCharacter(tt.npcX, tt.npcY, &EntityConfig{ID: "target"}, 1, false, nil)
			npc.State.HealthPoints = 100
			npc.BaseDefense = 0
			ctx.World.Characters = []*Character{npc}

			// Run attack multiple times to overcome 5% random miss chance
			hitDetected := false
			for attempt := 0; attempt < 100; attempt++ {
				mc.CheckAttackHits(ctx, "")
				if npc.State.HealthPoints < 100 {
					hitDetected = true
					break
				}
			}

			if hitDetected != tt.expectHit {
				distToCenter := math.Sqrt(math.Pow(1.25-tt.npcX, 2) + math.Pow(0-tt.npcY, 2))
				t.Errorf("%s: at (%.2f, %.2f) expected hit:%v, got hit:%v (Dist to hit center: %.3f, Max allowed: 1.875)", 
					tt.name, tt.npcX, tt.npcY, tt.expectHit, hitDetected, distToCenter)
			}
		})
	}
}
