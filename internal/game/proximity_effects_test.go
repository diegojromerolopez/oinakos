package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestProximityHazards(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.TemporalState.HealthPoints = 100
	mc.TemporalState.MaxHealthPoints = 100
	ctx.World.PlayableCharacter = mc
	mm := NewMechanicsManager(&Game{}) // We still need a Game for mm, but we pass ctx to its methods

	// 1. Test Aura Hazard
	campfireArchetype := &ObstacleArchetype{
		ID:   "campfire",
		Name: "Campfire",
		Actions: []ObstacleActionConfig{
			{
				Type:   ActionHarm,
				Amount: 10,
				Aura:   2.0,
			},
		},
	}
	campfire := NewObstacle("fire1", 1.0, 1.0, campfireArchetype)
	ctx.World.Obstacles = []*Obstacle{campfire}

	// Player is at (0,0), Fire is at (1,1). Distance is sqrt(2) approx 1.41. Radius is 2.0.
	// Should take damage.
	mm.UpdateProximityEffects(ctx)

	if mc.TemporalState.HealthPoints != 90 {
		t.Errorf("Expected health 90, got %d", mc.TemporalState.HealthPoints)
	}

	// 2. Test Interval Timer (should not take damage again immediately)
	mm.UpdateProximityEffects(ctx)
	if mc.TemporalState.HealthPoints != 90 {
		t.Errorf("Expected health to remain 90 due to interval, got %d", mc.TemporalState.HealthPoints)
	}

	// Tick the obstacle and timer
	campfire.Update()
	if campfire.EffectTimers[mc] != 59 {
		t.Errorf("Expected timer 59, got %d", campfire.EffectTimers[mc])
	}

	// Manually force timer to 0 to test re-application
	campfire.EffectTimers[mc] = 0
	mm.UpdateProximityEffects(ctx)
	if mc.TemporalState.HealthPoints != 80 {
		t.Errorf("Expected health 80 after timer reset, got %d", mc.TemporalState.HealthPoints)
	}

	// 3. Test Contact Hazard (no aura)
	spikeArchetype := &ObstacleArchetype{
		ID:   "spikes",
		Name: "Spikes",
		Actions: []ObstacleActionConfig{
			{
				Type:   ActionHarm,
				Amount: 5,
				Aura:   0, // Contact based
			},
		},
		Footprint: []FootprintPoint{
			{X: -0.5, Y: -0.5}, {X: 0.5, Y: -0.5}, {X: 0.5, Y: 0.5}, {X: -0.5, Y: 0.5},
		},
	}
	// Player footprint is small at center.
	spikes := NewObstacle("spikes1", 0, 0, spikeArchetype)
	ctx.World.Obstacles = []*Obstacle{spikes}

	mc.TemporalState.HealthPoints = 100
	mm.UpdateProximityEffects(ctx)
	if mc.TemporalState.HealthPoints != 95 {
		t.Errorf("Expected health 95 from contact hazard, got %d", mc.TemporalState.HealthPoints)
	}

	// Move player away from spikes
	mc.X = 10.0
	mc.Y = 10.0
	spikes.EffectTimers[mc] = 0
	mm.UpdateProximityEffects(ctx)
	if mc.TemporalState.HealthPoints != 95 {
		t.Errorf("Expected health 95 (no damage when away), got %d", mc.TemporalState.HealthPoints)
	}
}

func TestProximityHealing(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.TemporalState.HealthPoints = 50
	mc.TemporalState.MaxHealthPoints = 100
	ctx.World.PlayableCharacter = mc
	mm := NewMechanicsManager(&Game{})

	// 1. Test Aura Healing
	shrineArchetype := &ObstacleArchetype{
		ID:   "shrine",
		Name: "Healing Shrine",
		Actions: []ObstacleActionConfig{
			{
				Type:   ActionHeal,
				Amount: 10,
				Aura:   3.0,
			},
		},
	}
	shrine := NewObstacle("shrine1", 1.0, 1.0, shrineArchetype)
	ctx.World.Obstacles = []*Obstacle{shrine}

	mm.UpdateProximityEffects(ctx)
	if mc.TemporalState.HealthPoints != 60 {
		t.Errorf("Expected health 60, got %d", mc.TemporalState.HealthPoints)
	}

	// 2. Test Alignment Limit (Enemy-only healing shouldn't heal player)
	unholyAltarArch := &ObstacleArchetype{
		ID:   "unholy",
		Name: "Unholy Altar",
		Actions: []ObstacleActionConfig{
			{
				Type:           ActionHeal,
				Amount:         20,
				Aura:           5.0,
				AlignmentLimit: "enemy",
			},
		},
	}
	altar := NewObstacle("altar1", 0, 0, unholyAltarArch)
	ctx.World.Obstacles = []*Obstacle{altar}
	mc.TemporalState.HealthPoints = 50

	mm.UpdateProximityEffects(ctx)
	if mc.TemporalState.HealthPoints != 50 {
		t.Errorf("Expected health 50 (player is not an enemy), got %d", mc.TemporalState.HealthPoints)
	}

	// 3. Test Alignment Limit (Ally healing should heal player)
	holyStatueArch := &ObstacleArchetype{
		ID:   "holy",
		Name: "Holy Statue",
		Actions: []ObstacleActionConfig{
			{
				Type:           ActionHeal,
				Amount:         20,
				Aura:           5.0,
				AlignmentLimit: "ally",
			},
		},
	}
	statue := NewObstacle("statue1", 0, 0, holyStatueArch)
	ctx.World.Obstacles = []*Obstacle{statue}
	mm.UpdateProximityEffects(ctx)
	if mc.TemporalState.HealthPoints != 70 {
		t.Errorf("Expected health 70, got %d", mc.TemporalState.HealthPoints)
	}
}

func TestInteractiveHealing(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.TemporalState.HealthPoints = 10
	mc.TemporalState.MaxHealthPoints = 100
	ctx.World.PlayableCharacter = mc

	wellArchetype := &ObstacleArchetype{
		ID:   "well",
		Name: "Well",
		Actions: []ObstacleActionConfig{
			{
				Type:                ActionHeal,
				Amount:              999, // Full heal
				RequiresInteraction: true,
			},
		},
		CooldownTime: 1.0 / 60.0, // 1 second cooldown
	}
	well := NewObstacle("well1", 1.0, 0, wellArchetype)
	ctx.World.Obstacles = []*Obstacle{well}

	mockInput := ctx.Input.(*MockInputManager)
	// No key pressed -> no heal
	mc.Update(ctx)
	if mc.TemporalState.HealthPoints != 10 {
		t.Errorf("Expected health 10, got %d", mc.TemporalState.HealthPoints)
	}

	// Press Space -> Heal
	mockInput.PressedKeys[engine.KeySpace] = true
	mc.Update(ctx)
	if mc.TemporalState.HealthPoints != 100 {
		t.Errorf("Expected health 100 after using well, got %d", mc.TemporalState.HealthPoints)
	}
	if well.CooldownTicks != 60 {
		t.Errorf("Expected cooldown 60 ticks, got %d", well.CooldownTicks)
	}
}

func TestNPCProximityEffects(t *testing.T) {
	ctx := NewTestContext()
	arch := &EntityConfig{
		ID:   "peasant",
		Name: "Peasant",
	}
	n := &Character{
		Actor: Actor{
			X: 0,
			Y: 0,
			TemporalState: TemporalState{
				HealthPoints:    50,
				MaxHealthPoints: 100,
			},
			State:  ActorIdle,
			Config: arch,
		},
	}
	ctx.World.Characters = []*Character{n}
	ctx.World.PlayableCharacter = NewCharacter(100, 100, nil, 1, true, nil) // Keep MC away
	mm := NewMechanicsManager(&Game{})

	// 1. Hazard Effect on NPC
	campfireArch := &ObstacleArchetype{
		Actions: []ObstacleActionConfig{
			{
				Type:   ActionHarm,
				Amount: 10,
				Aura:   2.0,
			},
		},
	}
	fire := NewObstacle("f1", 0.5, 0.5, campfireArch)
	ctx.World.Obstacles = []*Obstacle{fire}

	mm.UpdateProximityEffects(ctx)
	if n.TemporalState.HealthPoints != 40 {
		t.Errorf("NPC should have 40 HP, got %d", n.TemporalState.HealthPoints)
	}

	// 2. Healing Effect on NPC
	wellArch := &ObstacleArchetype{
		Actions: []ObstacleActionConfig{
			{
				Type:   ActionHeal,
				Amount: 5,
				Aura:   2.0,
			},
		},
	}
	well := NewObstacle("w1", 0, 0, wellArch)
	ctx.World.Obstacles = []*Obstacle{well}
	mm.UpdateProximityEffects(ctx)
	if n.TemporalState.HealthPoints != 45 {
		t.Errorf("NPC should have 45 HP, got %d", n.TemporalState.HealthPoints)
	}
}
