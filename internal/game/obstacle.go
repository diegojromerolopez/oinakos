package game

import (
	"oinakos/internal/engine"
)

type Obstacle struct {
	ID            string  // Unique instance ID (e.g. from PreSpawnObstacle)
	X, Y          float64 // Grid positions
	Archetype     *ObstacleArchetype
	Health        int
	CooldownTicks int
	TickCounter   int
	Alive         bool
	EffectTimers  map[ActorInterface]int // Track intervals for hazards/healing per entity

	CachedFootprint *engine.Polygon // Optimization: obstacles don't move, cache world footprint
}

func NewObstacle(id string, x, y float64, config *ObstacleArchetype) *Obstacle {
	hp := 0 // Default indestructible
	if config != nil {
		hp = config.Health
	}

	return &Obstacle{
		ID:            id,
		X:             x,
		Y:             y,
		Archetype:     config,
		Health:        hp,
		CooldownTicks: 0,
		Alive:         true,
		EffectTimers:  make(map[ActorInterface]int),
	}
}

func (o *Obstacle) Update() {
	if !o.Alive {
		return
	}
	if o.CooldownTicks > 0 {
		o.CooldownTicks--
	}
	o.TickCounter++

	// Age the effect timers
	for entity, ticks := range o.EffectTimers {
		if ticks > 0 {
			o.EffectTimers[entity] = ticks - 1
		} else {
			// Cleanup old timers? Maybe not strictly necessary if map is small
		}
	}
}

func (o *Obstacle) TakeDamage(amount int) {
	if !o.Alive {
		return
	}
	// If the archetype marks the obstacle as indestructible, ignore all damage.
	if o.Archetype == nil || !o.Archetype.Destructible {
		return
	}

	o.Health -= amount
	if o.Health <= 0 {
		o.Alive = false
		DebugLog("Obstacle [%s] Destroyed at (%.2f, %.2f)", o.ID, o.X, o.Y)
	}
}

func (o *Obstacle) GetFootprint() engine.Polygon {
	if o.CachedFootprint != nil {
		return *o.CachedFootprint
	}

	var poly engine.Polygon
	if o.Archetype != nil && len(o.Archetype.Footprint) > 0 {
		poly = engine.Polygon{Points: make([]engine.Point, len(o.Archetype.Footprint))}
		for i, p := range o.Archetype.Footprint {
			poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
		}
	} else {
		// Absolute fallback for nil archetype or empty footprint.
		poly = engine.Polygon{Points: []engine.Point{
			{X: -0.2, Y: -0.2}, {X: 0.2, Y: -0.2}, {X: 0.2, Y: 0.2}, {X: -0.2, Y: 0.2},
		}}
	}

	transformed := poly.Transformed(o.X, o.Y)
	o.CachedFootprint = &transformed
	return transformed
}
