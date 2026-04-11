package game

import (
	"testing"
)

// Test ranged hit scenarios including distance limits and alignment misses.
func TestRangedHitScenarios(t *testing.T) {
	ctx := NewTestContext()

	// Setup world
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	ctx.World.PlayableCharacter = mc

	tests := []struct {
		name      string
		npcX      float64
		npcY      float64
		maxRange  float64
		expectHit bool
	}{
		{
			name:      "Ranged Hit - Well within range",
			npcX:      5.0,
			npcY:      0.0,
			maxRange:  10.0,
			expectHit: true,
		},
		{
			name:      "Ranged Miss - Beyond MaxRange",
			npcX:      12.0,
			npcY:      0.0,
			maxRange:  10.0,
			expectHit: false,
		},
		{
			name:      "Ranged Miss - Off-Axis",
			npcX:      5.0,
			npcY:      2.0, // Projectile travels along Y=0, NPC hit box is 0.8 radius
			maxRange:  10.0,
			expectHit: false,
		},
		{
			name:      "Ranged Hit - Diagonal",
			npcX:      5.0,
			npcY:      5.0,
			maxRange:  15.0,
			expectHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy := tt.npcX, tt.npcY
			if dx == 0 && dy == 0 {
				dx = 1
			} // prevent zero vector

			// Fire projectile from (0,0) towards NPC
			pX, pY := dx, dy
			if tt.name == "Ranged Miss - Off-Axis" {
				pX, pY = 1.0, 0.0 // Fire along X axis, NPC is at (5,2), should miss
			}
			p := NewProjectile(0, 0, pX, pY, 1.0, 10, true, tt.maxRange)

			npc := NewCharacter(tt.npcX, tt.npcY, &EntityConfig{ID: "target"}, 1, false, nil)
			npc.State.HealthPoints = 100
			ctx.World.Characters = []*Character{npc}
			ctx.World.Projectiles = []*Projectile{p}

			// Simulate travel
			ticks := int(tt.maxRange) + 5
			hitDetected := false
			for i := 0; i < ticks; i++ {
				p.Update(ctx)
				if npc.State.HealthPoints < 100 {
					hitDetected = true
					break
				}
				if !p.Alive {
					break
				}
			}

			if hitDetected != tt.expectHit {
				t.Errorf("%s: at (%.1f, %.1f) expected hit:%v, got hit:%v",
					tt.name, tt.npcX, tt.npcY, tt.expectHit, hitDetected)
			}
		})
	}
}

// Verifies that CheckAttackHits correctly spawns a projectile when using a ranged weapon.
func TestCheckAttackHits_SpawnsProjectileForRanged(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.Weapon = &Weapon{Type: "ranged", Name: "Test Bow", MaxDistance: "10"}
	mc.Facing = DirSE // This maps to dx=1, dy=0.5 in our fixed logic
	ctx.World.PlayableCharacter = mc

	// Clear any existing projectiles
	ctx.World.Projectiles = nil

	// Target in range but we don't care about the hit here, just the spawn
	mc.CheckAttackHits(ctx, "")

	if len(ctx.World.Projectiles) != 1 {
		t.Errorf("Expected 1 projectile to be spawned, got %d", len(ctx.World.Projectiles))
	} else {
		p := ctx.World.Projectiles[0]
		if !p.IsPlayer {
			t.Error("Spawned projectile should be marked as player-fired (IsPlayer=true)")
		}
		if p.X != mc.X || p.Y != mc.Y {
			t.Errorf("Projectile should spawn at character coordinates (%.1f, %.1f), got (%.1f, %.1f)", mc.X, mc.Y, p.X, p.Y)
		}
	}
}

// Verifies that ranged weapons DO NOT do instant hit-scan damage.
func TestCheckAttackHits_RangedNoInstantDamage(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.Weapon = &Weapon{Type: "ranged", Name: "Test Bow", MaxDistance: "10"}
	mc.Facing = DirSE
	ctx.World.PlayableCharacter = mc

	// NPC right in front of player
	npc := NewCharacter(2, 0, &EntityConfig{ID: "target"}, 1, false, nil)
	npc.State.HealthPoints = 100
	ctx.World.Characters = []*Character{npc}

	mc.CheckAttackHits(ctx, "")

	if npc.State.HealthPoints != 100 {
		t.Error("Ranged weapon should NOT do instant damage; it should fire a projectile first")
	}
}
