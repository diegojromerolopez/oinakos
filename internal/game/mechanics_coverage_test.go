package game

import (
	"testing"
)

func TestMechanicsManager_Coverage(t *testing.T) {
	g := setupTestGame()
	ctx := &SystemContext{
		World:      g.World,
		Registries: g.Registries,
		Settings:   g.settings,
	}
	mm := NewMechanicsManager(g)
	
	// 1. UpdateFogOfWar
	mm.UpdateFogOfWar(ctx)
	if len(ctx.World.ExploredTiles) == 0 { t.Error("Fog of war not updated") }
	
	// 2. UpdateProximityEffects
	harmObs := &Obstacle{
		X: 1, Y: 1, 
		Alive: true,
		Archetype: &ObstacleArchetype{
			Actions: []ObstacleActionConfig{
				{Type: ActionHarm, Aura: 5.0, Amount: 10},
			},
		},
		EffectTimers: make(map[ActorInterface]int),
	}
	ctx.World.Obstacles = append(ctx.World.Obstacles, harmObs)
	ctx.World.PlayableCharacter.X, ctx.World.PlayableCharacter.Y = 1.1, 1.1
	mm.UpdateProximityEffects(ctx)
	
	// 3. CheckWinConditions
	ctx.World.CurrentMapType.Type = ObjKillCount
	ctx.World.CurrentMapType.TargetKillCount = 5
	if ctx.World.PlayableCharacter.MapKills == nil { ctx.World.PlayableCharacter.MapKills = make(map[string]int) }
	ctx.World.PlayableCharacter.MapKills["orc"] = 10
	if !mm.CheckWinConditions(ctx) { t.Error("Win condition not met for kill count") }
}
