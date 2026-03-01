package game

import (
	"oinakos/internal/engine"
)

type Obstacle struct {
	X, Y          float64 // Grid positions
	Archetype     *ObstacleArchetype
	Health        int
	CooldownTicks int
	Alive         bool
}

func NewObstacle(x, y float64, config *ObstacleArchetype) *Obstacle {
	hp := 0 // Default indestructible
	if config != nil {
		hp = config.Health
	}

	return &Obstacle{
		X:             x,
		Y:             y,
		Archetype:     config,
		Health:        hp,
		CooldownTicks: 0,
		Alive:         true,
	}
}

func (o *Obstacle) Update() {
	if !o.Alive {
		return
	}
	if o.CooldownTicks > 0 {
		o.CooldownTicks--
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
	}
}

func (o *Obstacle) GetFootprint() engine.Polygon {
	if o.Archetype != nil && len(o.Archetype.Footprint) > 0 {
		poly := engine.Polygon{Points: make([]engine.Point, len(o.Archetype.Footprint))}
		for i, p := range o.Archetype.Footprint {
			poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
		}
		return poly.Transformed(o.X, o.Y)
	}
	// Absolute fallback for nil archetype or empty footprint.
	return engine.Polygon{Points: []engine.Point{
		{X: -0.2, Y: -0.2}, {X: 0.2, Y: -0.2}, {X: 0.2, Y: 0.2}, {X: -0.2, Y: 0.2},
	}}.Transformed(o.X, o.Y)
}
