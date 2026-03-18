package game

import (
	"testing"
)

// Tests for NPC behavior branches (wander, fighter, chaotic, neutral, ally)

func TestNPCBehavior_Wander_SetsDirection(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(100, 100, nil, 1, true)
	ctx.World.PlayableCharacter = mc

	npc := NewCharacter(0, 0, &EntityConfig{ID: "test"}, 1, false)
	npc.Health = 100
	npc.MaxHealth = 100
	npc.Behavior = BehaviorWander
	npc.Alignment = AlignmentEnemy
	npc.Speed = 0.1 // must be non-zero
	// Pre-set direction so movement is predictable
	npc.WanderDirX = 1.0
	npc.WanderDirY = 0.0
	ctx.World.Characters = []*Character{npc}

	for i := 0; i < 5; i++ {
		npc.Update(ctx)
	}

	if npc.X <= 0 {
		t.Error("Wander NPC with WanderDirX=1 should have moved in +X direction")
	}
}

func TestNPCBehavior_Fighter_TargetsNearestNPC(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(100, 100, nil, 1, true) // Far away
	ctx.World.PlayableCharacter = mc

	fighter := NewCharacter(0, 0, &EntityConfig{ID: "fighter"}, 1, false)
	fighter.Behavior = BehaviorNpcFighter
	fighter.Alignment = AlignmentEnemy

	target := NewCharacter(2, 0, &EntityConfig{ID: "target"}, 1, false)
	target.Alignment = AlignmentAlly
	fighter.Health = 100
	fighter.MaxHealth = 100
	target.Health = 100
	target.MaxHealth = 100
	ctx.World.Characters = []*Character{fighter, target}

	for i := 0; i < 10; i++ {
		fighter.Update(ctx)
	}

	if fighter.TargetActor == nil {
		t.Error("Fighter NPC should have acquired a target NPC")
	}
}

func TestNPCBehavior_Chaotic_TargetsNearestActor(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(3, 0, nil, 1, true) // Closer than farNPC
	ctx.World.PlayableCharacter = mc

	chaotic := NewCharacter(0, 0, &EntityConfig{ID: "chaotic"}, 1, false)
	chaotic.Behavior = BehaviorChaotic
	chaotic.Alignment = AlignmentEnemy

	farNPC := NewCharacter(20, 0, &EntityConfig{ID: "far"}, 1, false)
	farNPC.Alignment = AlignmentEnemy
	chaotic.Health = 100
	chaotic.MaxHealth = 100
	farNPC.Health = 100
	farNPC.MaxHealth = 100
	ctx.World.Characters = []*Character{chaotic, farNPC}

	chaotic.Update(ctx)

	// Player at dist 3, farNPC at dist 20 → chaotic should target player
	if chaotic.TargetActor != &mc.Actor {
		t.Error("Chaotic NPC should target the nearest actor (player at dist 3)")
	}
}

func TestNPCBehavior_Neutral_DoesNotTargetActor(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true)
	ctx.World.PlayableCharacter = mc

	npc := NewCharacter(1, 0, nil, 1, false)
	npc.Alignment = AlignmentNeutral
	ctx.World.Characters = []*Character{npc}

	for i := 0; i < 5; i++ {
		npc.Update(ctx)
	}

	if npc.TargetActor != nil && npc.TargetActor == &mc.Actor {
		t.Error("Neutral NPC should never target the player")
	}
}

func TestNPCBehavior_Ally_FollowsPlayerWhenNoEnemies(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(10, 10, nil, 1, true)
	ctx.World.PlayableCharacter = mc

	ally := NewCharacter(0, 0, &EntityConfig{ID: "ally"}, 1, false)
	ally.Alignment = AlignmentAlly
	ally.Speed = 0.2 // must be non-zero
	ally.Health = 100
	ally.MaxHealth = 100
	ctx.World.Characters = []*Character{ally}

	for i := 0; i < 20; i++ {
		ally.Update(ctx)
	}

	// Ally should have moved toward the player (closer than initial dist ~14)
	dist := (ally.X-mc.X)*(ally.X-mc.X) + (ally.Y-mc.Y)*(ally.Y-mc.Y)
	if dist >= 200 { // initial dist^2 ≈ 200
		t.Errorf("Ally NPC should be moving toward player; dist²=%.1f", dist)
	}
}

func TestNPCBehavior_ScavengeUpgrade(t *testing.T) {
	ctx := NewTestContext()
	g := &Game{}
	ctx.World.Game = g
	
	// Create NPC with a weak weapon
	npc := NewCharacter(0, 0, &EntityConfig{ID: "scavenger", MaxWeight: 10.0}, 1, false)
	npc.Health = 100
	npc.MaxHealth = 100
	npc.Speed = 0.2
	weakWeapon := &ObjectConfig{
		ID: "weak_weapon", Type: "weapon", Slot: "weapon", Weight: 2.0,
		Combat: &Weapon{Damage: Damage{Min: 1, Max: 2}},
	}
	npc.EquipItem(weakWeapon)
	
	ctx.World.Characters = []*Character{npc}
	
	// Create a stronger weapon on the ground nearby
	strongWeapon := &ObjectConfig{
		ID: "strong_weapon", Type: "weapon", Slot: "weapon", Weight: 3.0,
		Combat: &Weapon{Damage: Damage{Min: 10, Max: 20}},
	}
	item := NewItemInstance(strongWeapon.ID, strongWeapon, 2.0, 0.0)
	ctx.World.Items = []*ItemInstance{item}
	
	// Let the NPC update for enough ticks to trigger the loot scan (Tick % 30 == 0) and walk to it
	for i := 0; i < 150; i++ {
		npc.Update(ctx)
	}
	
	// NPC should have picked up the strong weapon and equipped it, dropping the weak weapon to inventory.
	if npc.Weapon.Damage.Max != 20 {
		t.Errorf("NPC did not equip the upgrade weapon. Current max damage: %d", npc.Weapon.Damage.Max)
	}
}

