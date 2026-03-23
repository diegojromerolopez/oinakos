package game

import (
	"math"
	"testing"
)

func TestCombatMechanics(t *testing.T) {
	ctx := NewTestContext()
	// Setup a controlled combat scenario
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.BaseAttack = 20
	mc.BaseDefense = 5
	mc.Health = 100
	mc.MaxHealth = 100
	ctx.World.PlayableCharacter = mc

	npc := NewCharacter(1, 0, nil, 1, false, nil)
	npc.BaseAttack = 15
	npc.BaseDefense = 2
	npc.Health = 50
	npc.MaxHealth = 50
	npc.Alignment = AlignmentEnemy
	ctx.World.Characters = []*Character{npc}

	// Player attacks NPC
	initialNpcHealth := npc.Health
	// For testing, we won't roll, just assume a hit and calculate damage
	rawDmg := 25 // mc.Weapon.Damage.Max is not used here to be safer
	protection := npc.GetTotalProtection()
	damage := int(math.Max(1, float64(rawDmg-protection)))
	npc.TakeDamage(damage, mc, ctx)

	// Ensure the expected damage is correct
	if npc.Health != initialNpcHealth-damage {
		t.Errorf("NPC health mismatch. Expected %d, got %d", initialNpcHealth-damage, npc.Health)
	}

	// 2. NPC attacks Player
	initialMcHealth := mc.Health
	nRawDmg := float64(npc.BaseAttack)
	nProtection := float64(mc.GetTotalProtection())
	npcDamage := int(math.Max(1, nRawDmg-nProtection))
	mc.TakeDamage(npcDamage, nil, ctx)

	// NPC damage: npc.BaseAttack=15, mc.BaseDefense=5, mc.Protection=0 → expect 15
	if mc.Health != initialMcHealth-npcDamage {
		t.Errorf("MC health mismatch. Expected %d, got %d", initialMcHealth-npcDamage, mc.Health)
	}

	// 3. Test XP reward on death — use a known archetype so XP logic fires
	npc2 := NewCharacter(1, 0, &EntityConfig{ID: "orc", XP: 10}, 1, false, nil)
	npc2.Health = 1
	mc.XP = 0
	mc.Kills = 0
	ctx.World.Characters = []*Character{npc2}
	npc2.TakeDamage(10, mc, ctx)
	if npc2.State != ActorDead {
		t.Fatalf("NPC should be dead")
	}
	if mc.XP <= 0 {
		t.Error("Player should gain XP from killing NPC")
	}
	if mc.Kills != 1 {
		t.Errorf("Player kills should be 1, got %d", mc.Kills)
	}
}

func TestProjectileCombat(t *testing.T) {
	ctx := NewTestContext()
	// NPC projectile fires at player (the actual path in Projectile.Update)
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.Health = 100
	ctx.World.PlayableCharacter = mc

	// NPC fires projectile at mc's position
	p := NewProjectile(5, 0, -1, 0, 0.15, 20, false, 100.0) // IsFriendly=false → targets mc
	ctx.World.Projectiles = []*Projectile{p}

	// Put the projectile right at the player
	p.X = mc.X
	p.Y = mc.Y
	p.Update(ctx)

	// Player should have taken damage
	if mc.Health >= 100 {
		t.Errorf("Player should have taken damage; health=%d", mc.Health)
	}
	if p.Alive {
		t.Error("Projectile should be dead after hitting player")
	}
	if len(ctx.World.FloatingTexts) == 0 {
		t.Error("Expected floating damage text")
	}
}
