package game

import (
	"testing"
)

func TestCharacterActions_Woodcutting(t *testing.T) {
	ctx := setupTestGame()
	mc := ctx.playableCharacter
	mc.X, mc.Y = 0, 0
	
	// Create an axe
	axe := &ObjectConfig{
		ID: "axe",
		Name: "Bronze Axe",
		Combat: &Weapon{Name: "Bronze Axe", Damage: Damage{Min: 10, Max: 10}},
	}
	ctx.Registries.Objects.Objects["axe"] = axe
	mc.Weapon = axe.Combat
	mc.State = ActorChopping
	
	// Create a tree
	treeArch := &ObstacleArchetype{
		ID: "tree_oak",
		Type: "tree",
		Destructible: true,
		Health: 100,
	}
	ctx.Registries.Obstacles.Archetypes["tree_oak"] = treeArch
	tree := &Obstacle{
		X: 1.0, Y: 0,
		Archetype: treeArch,
		Alive: true,
		Health: 100,
	}
	tree.Health = 100
	ctx.World.Obstacles = []*Obstacle{tree}
	
	// Create wood object config for drops
	ctx.Registries.Objects.Objects["wood"] = &ObjectConfig{ID: "wood", Name: "Wood"}
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	sysCtx.Registries = ctx.Registries
	
	mc.CheckAttackHits(sysCtx)
	
	if tree.Health == 100 {
		t.Errorf("Tree did not take damage from chopping")
	}
	
	if len(ctx.World.Items) == 0 {
		t.Errorf("Wood was not dropped from chopping tree")
	}
}

func TestCharacterActions_Digging(t *testing.T) {
	ctx := setupTestGame()
	mc := ctx.playableCharacter
	mc.X, mc.Y = 0, 0
	mc.Facing = DirSE
	
	pike := &ObjectConfig{
		ID: "pike",
		Name: "Iron Pike",
		Combat: &Weapon{Name: "Iron Pike"},
	}
	ctx.Registries.Objects.Objects["pike"] = pike
	mc.Weapon = pike.Combat
	mc.State = ActorDigging
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	sysCtx.Registries = ctx.Registries
	
	// Set target direction for digging
	mc.CheckAttackHits(sysCtx)
	
	elevation := ctx.World.CurrentMapType.GetElevationAt(mc.X+1, mc.Y)
	if elevation >= 0 {
		t.Errorf("Digging did not reduce elevation. got %f", elevation)
	}
}

func TestCharacterActions_CaveIn(t *testing.T) {
	ctx := setupTestGame()
	mc := ctx.playableCharacter
	mc.X, mc.Y = 0, 0
	mc.Facing = DirSE
	
	pike := &ObjectConfig{
		ID: "pike",
		Name: "Iron Pike",
		Combat: &Weapon{Name: "Iron Pike"},
	}
	ctx.Registries.Objects.Objects["pike"] = pike
	mc.Weapon = pike.Combat
	mc.State = ActorDigging
	
	// Create a steep slope to trigger cave-in
	ctx.World.CurrentMapType.Heightmap = make(map[string]float64)
	ctx.World.CurrentMapType.Heightmap["1,1"] = 10.0
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	sysCtx.Registries = ctx.Registries
	
	mc.CheckAttackHits(sysCtx)
	
	if mc.Health > 0 {
		t.Errorf("Cave-in did not kill character")
	}
}

func TestCharacterActions_AIReaction(t *testing.T) {
	ctx := setupTestGame()
	npc := NewCharacter(5.0, 0, nil, 1, false, nil)
	npc.Health = 100
	npc.MaxHealth = 100
	npc.Alignment = AlignmentNeutral
	ctx.World.Characters = []*Character{npc}
	
	attacker := NewCharacter(0, 0, nil, 1, true, nil)
	attacker.Health = 100
	attacker.MaxHealth = 100
	attacker.Health = 100 // High health to trigger fleeing in NPC
	attacker.MaxHealth = 100
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	
	// Light damage -> becomes enemy
	npc.TakeDamage(10, attacker, sysCtx)
	if npc.Alignment != AlignmentEnemy {
		t.Errorf("NPC did not become enemy after taking damage. alignment=%v", npc.Alignment)
	}
	
	// Heavy damage (low health) -> flees
	npc.Health = 5
	npc.handleAIReaction(attacker, sysCtx)
	if npc.Behavior != BehaviorFlee {
		t.Errorf("Low health NPC did not switch to Flee behavior. behavior=%v", npc.Behavior)
	}
}
