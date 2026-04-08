package game

import (
	"image"
	"oinakos/internal/engine"
	"testing"
)

func TestMechanics_WinConditions(t *testing.T) {
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", nil, nil, false, "0.1-test")
	mm := NewMechanicsManager(g)
	g.World.CurrentMapType = &g.currentMapType
	ctx := &SystemContext{World: g.World}

	// 1. ObjKillCount
	g.currentMapType.Type = ObjKillCount
	g.currentMapType.TargetKillCount = 5
	g.playableCharacter.Kills = 4
	g.playableCharacter.MapKills = map[string]int{"orc": 4}
	if mm.CheckWinConditions(ctx) {
		t.Error("Should not win with 4 kills when 5 are needed")
	}
	g.playableCharacter.MapKills["orc"] = 5
	if !mm.CheckWinConditions(ctx) {
		t.Error("Should win with 5 kills")
	}

	// 2. ObjSurvive
	g.currentMapType.Type = ObjSurvive
	g.currentMapType.TargetTime = 60.0
	g.playTime = 59.0
	g.World.PlayTime = 59.0
	if mm.CheckWinConditions(ctx) {
		t.Error("Should not win before target time")
	}
	g.playTime = 60.0
	g.World.PlayTime = 60.0
	if !mm.CheckWinConditions(ctx) {
		t.Error("Should win at target time")
	}

	// 3. ObjReachPortal
	g.currentMapType.Type = ObjReachPortal
	g.currentMapType.TargetPoint = engine.Point{X: 10, Y: 10}
	g.currentMapType.TargetRadius = 2.0
	g.playableCharacter.X, g.playableCharacter.Y = 5, 5
	if mm.CheckWinConditions(ctx) {
		t.Error("Should not win far from portal")
	}
	g.playableCharacter.X, g.playableCharacter.Y = 9, 9
	if !mm.CheckWinConditions(ctx) {
		t.Error("Should win near portal")
	}
}

func TestMechanics_ProximityEffects(t *testing.T) {
	ctx := NewTestContext()
	g := NewGame(nil, &engine.MockGraphics{}, "", "", "", nil, nil, false, "0.1-test")
	ctx.World.Game = g
	mm := NewMechanicsManager(g)
	
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	ctx.World.PlayableCharacter = mc
	mc.X, mc.Y = 0, 0
	mc.State.MaxHealthPoints = 100
	mc.State.HealthPoints = 50
	
	// 1. Harmful Aura
	spikeArch := &ObstacleArchetype{
		ID: "spikes",
		Actions: []ObstacleActionConfig{
			{Type: ActionHarm, Amount: 10, Aura: 2.0},
		},
	}
	spikes := NewObstacle("spikes1", 1, 1, spikeArch)
	ctx.World.Obstacles = []*Obstacle{spikes}
	
	mm.UpdateProximityEffects(ctx)
	if mc.State.HealthPoints != 40 {
		t.Errorf("Expected health 40 after spike damage, got %d", mc.State.HealthPoints)
	}

	// 2. Healing Aura with Alignment Limit
	// Clear previous timers and obstacles for clean test
	ctx.World.Obstacles = nil
	mc.State.HealthPoints = 50
	
	wellArch := &ObstacleArchetype{
		ID: "holy_well",
		Actions: []ObstacleActionConfig{
			{Type: ActionHeal, Amount: 5, Aura: 5.0, AlignmentLimit: "ally"},
		},
	}
	well := NewObstacle("well1", 2, 2, wellArch)
	ctx.World.Obstacles = []*Obstacle{well}
	
	mm.UpdateProximityEffects(ctx)
	if mc.State.HealthPoints != 55 {
		t.Errorf("Expected health 55 after healing, got %d", mc.State.HealthPoints)
	}

	// 3. Enemy should not be healed by "ally" well
	enemynpc := NewCharacter(2, 2, nil, 1, false, nil)
	enemynpc.Alignment = AlignmentEnemy
	enemynpc.State.MaxHealthPoints = 100
	enemynpc.State.HealthPoints = 50
	ctx.World.Characters = []*Character{enemynpc}
	
	mm.UpdateProximityEffects(ctx)
	if enemynpc.State.HealthPoints != 50 {
		t.Errorf("Enemy should not be healed by ally-only well, got %d", enemynpc.State.HealthPoints)
	}
}

func TestMechanics_FogOfWar(t *testing.T) {
	ctx := NewTestContext()
	mm := NewMechanicsManager(ctx.World.Game)
	mc := NewCharacter(10, 10, nil, 1, true, nil)
	ctx.World.PlayableCharacter = mc
	
	ctx.World.PlayableCharacter.X, ctx.World.PlayableCharacter.Y = 10, 10
	mm.UpdateFogOfWar(ctx)
	
	if !ctx.World.ExploredTiles[image.Point{X: 10, Y: 10}] {
		t.Error("Center tile should be explored")
	}
	if !ctx.World.ExploredTiles[image.Point{X: 15, Y: 15}] {
		t.Error("Tile within radius should be explored")
	}
	if ctx.World.ExploredTiles[image.Point{X: 25, Y: 25}] {
		t.Error("Tile far away should not be explored")
	}
}
