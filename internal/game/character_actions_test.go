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
	mc.ActionState = ActorChopping
	
	// Create a tree
	treeArch := &ObstacleArchetype{
		ID: "tree_oak",
		Type: "tree",
		Destructible: true,
		HealthPoints: 100,
	}
	ctx.Registries.Obstacles.Archetypes["tree_oak"] = treeArch
	tree := &Obstacle{
		X: 1.0, Y: 0,
		Archetype: treeArch,
		Alive: true,
		WeightLeft: 100,
	}
	ctx.World.Obstacles = []*Obstacle{tree}
	
	// Create wood object config for drops
	ctx.Registries.Objects.Objects["wood"] = &ObjectConfig{ID: "wood", Name: "Wood"}
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	sysCtx.Registries = ctx.Registries
	
	mc.CheckAttackHits(sysCtx, "")
	
	if tree.WeightLeft == 100 && tree.HealthPoints == 100 {
		t.Errorf("Tree did not take damage or mass reduction from chopping")
	}
	
	if len(ctx.World.Items) == 0 && len(mc.Inventory) == 0 {
		t.Errorf("Wood was not dropped from chopping tree (not in world nor in inventory)")
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
	mc.ActionState = ActorDigging
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	sysCtx.Registries = ctx.Registries
	
	// Set target direction for digging
	mc.CheckAttackHits(sysCtx, "")
	
	// Range is now 5.0 for harvesting actions
	targetX := mc.X + 5.0
	elevation := ctx.World.CurrentMapType.GetElevationAt(targetX, mc.Y)
	if elevation >= 0 {
		t.Errorf("Digging did not reduce elevation at target. got %f", elevation)
	}
}

func TestCharacterActions_GenerousWoodcutting(t *testing.T) {
	ctx := setupTestGame()
	mc := ctx.playableCharacter
	mc.X, mc.Y = 0, 0
	mc.Facing = DirSE // +X
	
	// Create an axe
	axe := &ObjectConfig{
		ID: "axe",
		Name: "Bronze Axe",
		Combat: &Weapon{Name: "Bronze Axe", Damage: Damage{Min: 10, Max: 10}},
	}
	ctx.Registries.Objects.Objects["axe"] = axe
	mc.Weapon = axe.Combat
	mc.ActionState = ActorChopping
	
	// Create wood object config for drops
	ctx.Registries.Objects.Objects["wood"] = &ObjectConfig{ID: "wood", Name: "Wood"}
	
	// Create a tree FURTHER AWAY (e.g. 4.5 units)
	// With range 5.0, hitCircle center is at 2.5, radius is 3.75.
	// It covers up to 2.5 + 3.75 = 6.25 units.
	treeDist := 4.5
	treeArch := &ObstacleArchetype{
		ID: "tree_oak",
		Type: "tree",
		Destructible: true,
	}
	tree := &Obstacle{
		X: treeDist, Y: 0,
		Archetype: treeArch,
		Alive: true,
		WeightLeft: 100,
	}
	ctx.World.Obstacles = []*Obstacle{tree}
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	sysCtx.Registries = ctx.Registries
	
	mc.CheckAttackHits(sysCtx, "")
	
	if tree.WeightLeft == 100 && tree.HealthPoints == 100 {
		t.Errorf("Tree at distance %.1f was NOT hit despite generous range limits", treeDist)
	}
	
	if len(ctx.World.Items) == 0 && len(mc.Inventory) == 0 {
		t.Errorf("No wood generated (neither dropped nor in inventory) for tree at distance %.1f", treeDist)
	}
}

func TestCharacterActions_GenerousDigging(t *testing.T) {
	ctx := setupTestGame()
	mc := ctx.playableCharacter
	mc.X, mc.Y = 0, 0
	mc.Facing = DirSW // +Y
	
	pike := &ObjectConfig{
		ID: "pike",
		Name: "Iron Pike",
		Combat: &Weapon{Name: "Iron Pike"},
	}
	ctx.Registries.Objects.Objects["pike"] = pike
	mc.Weapon = pike.Combat
	mc.ActionState = ActorDigging
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	sysCtx.Registries = ctx.Registries
	
	// CheckAttackHits should now use a range of 5.0
	mc.CheckAttackHits(sysCtx, "")
	
	// DirSW is +Y, so it should dig at (0, 5)
	targetY := mc.Y + 5.0
	elevation := ctx.World.CurrentMapType.GetElevationAt(mc.X, targetY)
	if elevation >= 0 {
		t.Errorf("Digging at distance 5.0 did not reduce elevation. got %f", elevation)
	}
}

func TestCharacterActions_CaveIn(t *testing.T) {
	ctx := setupTestGame()
	mc := ctx.playableCharacter
	mc.X, mc.Y = 0, 0
	mc.Facing = DirSE // dig target → gridX=5, gridY=0
	mc.State.HealthPoints = 100
	mc.State.MaxHealthPoints = 100

	pike := &ObjectConfig{
		ID: "pike",
		Name: "Iron Pike",
		Combat: &Weapon{Name: "Iron Pike"},
	}
	ctx.Registries.Objects.Objects["pike"] = pike
	mc.Weapon = pike.Combat
	mc.ActionState = ActorDigging
	
	// Place a neighbour of the dig target (5,0) that is 10.0 above the new dug level (-0.5).
	// Difference = 10.0 - (-0.5) = 10.5 ≥ 6.0 → triggers cave-in.
	ctx.World.CurrentMapType.Heightmap = make(map[string]float64)
	ctx.World.CurrentMapType.Heightmap["6,0"] = 10.0 // direct east neighbour of dig target (5,0)
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	sysCtx.Registries = ctx.Registries
	
	mc.CheckAttackHits(sysCtx, "")
	
	if mc.State.HealthPoints > 0 {
		t.Errorf("Cave-in did not kill character")
	}
}

func TestCharacterActions_AIReaction(t *testing.T) {
	ctx := setupTestGame()
	npc := NewCharacter(5.0, 0, nil, 1, false, nil)
	npc.State.HealthPoints = 100
	npc.State.MaxHealthPoints = 100
	npc.Alignment = AlignmentNeutral
	ctx.World.Characters = []*Character{npc}
	
	attacker := NewCharacter(0, 0, nil, 1, true, nil)
	attacker.State.HealthPoints = 100
	attacker.State.MaxHealthPoints = 100
	attacker.State.HealthPoints = 100 // High health to trigger fleeing in NPC
	attacker.State.MaxHealthPoints = 100
	
	sysCtx := NewTestContext()
	sysCtx.World = ctx.World
	
	// Light damage -> becomes enemy
	npc.TakeDamage(10, attacker, sysCtx)
	if npc.Alignment != AlignmentEnemy {
		t.Errorf("NPC did not become enemy after taking damage. alignment=%v", npc.Alignment)
	}
	
	// Heavy damage (low health) -> flees
	npc.State.HealthPoints = 5
	npc.TakeDamage(1, attacker, sysCtx) // Triggers handleAIReaction internally
	if npc.Behavior != BehaviorFlee {
		t.Errorf("Low health NPC did not switch to Flee behavior. behavior=%v", npc.Behavior)
	}
}
