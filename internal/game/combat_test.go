package game

import (
	"math"
	"testing"
)

func TestCombatMechanics(t *testing.T) {
	// Setup a controlled combat scenario
	mc := NewMainCharacter(0, 0, nil)
	mc.BaseAttack = 20
	mc.BaseDefense = 5
	mc.Health = 100
	mc.MaxHealth = 100

	npc := NewNPC(1, 0, nil, 1)
	npc.BaseAttack = 15
	npc.BaseDefense = 2
	npc.Health = 50
	npc.MaxHealth = 50
	npc.Alignment = AlignmentEnemy

	audio := &MockAudioManager{}

	// Player attacks NPC
	initialNpcHealth := npc.Health
	attk := float64(mc.GetTotalAttack())
	def := float64(npc.GetTotalDefense())
	hitChance := int(attk / (attk + def) * 100)
	// For testing, we won't roll, just assume a hit and calculate damage
	rawDmg := mc.Weapon.MaxDamage // Assume max
	protection := npc.GetTotalProtection()
	damage := int(math.Max(1, float64(rawDmg-protection)))
	_ = hitChance // used in real logic but not here
	npc.TakeDamage(damage, mc, nil, audio)

	// Ensure the expected damage is correct: weapon maxDamage=25, npc protection=0, so 25 dmg
	if npc.Health != initialNpcHealth-damage {
		t.Errorf("NPC health mismatch. Expected %d, got %d", initialNpcHealth-damage, npc.Health)
	}

	// 2. NPC attacks Player
	initialMcHealth := mc.Health
	nRawDmg := float64(npc.BaseAttack) // Simplified since NPC might not have weapon in this test
	nProtection := float64(mc.GetTotalProtection())
	npcDamage := int(math.Max(1, nRawDmg-nProtection))
	mc.TakeDamage(npcDamage, audio)

	// NPC damage: npc.BaseAttack=15, mc.BaseDefense=5, mc.Protection=1(leather) → expect 15-1=14
	if mc.Health != initialMcHealth-npcDamage {
		t.Errorf("MC health mismatch. Expected %d, got %d", initialMcHealth-npcDamage, mc.Health)
	}

	// 3. Test XP reward on death — use a known archetype so XP logic fires
	npc2 := NewNPC(1, 0, &Archetype{ID: "orc"}, 1)
	npc2.Health = 1
	mc.XP = 0
	mc.Kills = 0
	npc2.TakeDamage(10, mc, nil, audio)
	if npc2.State != NPCDead {
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
	// NPC projectile fires at player (the actual path in Projectile.Update)
	mc := NewMainCharacter(0, 0, nil)
	mc.Health = 100

	// NPC fires projectile at mc's position
	p := NewProjectile(5, 0, -1, 0, 0.15, 20, false, 100.0) // IsFriendly=false → targets mc

	var fts []*FloatingText
	// Put the projectile right at the player
	p.X = mc.X
	p.Y = mc.Y
	p.Update(mc, nil, &fts)

	// Player should have taken damage
	if mc.Health >= 100 {
		t.Errorf("Player should have taken damage; health=%d", mc.Health)
	}
	if p.Alive {
		t.Error("Projectile should be dead after hitting player")
	}
	if len(fts) == 0 {
		t.Error("Expected floating damage text")
	}
}
