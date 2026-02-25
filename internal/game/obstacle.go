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
	// Indestructible obstacles have 0 or negative health starting point
	if o.Archetype == nil || o.Archetype.Health <= 0 {
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

	// Fallback to width/height
	hw, hh := 0.3, 0.3 // Default
	if o.Archetype != nil {
		if o.Archetype.FootprintWidth > 0 {
			hw = o.Archetype.FootprintWidth / 2
		}
		if o.Archetype.FootprintHeight > 0 {
			hh = o.Archetype.FootprintHeight / 2
		}
	}

	return engine.Polygon{Points: []engine.Point{
		{X: -hw, Y: -hh}, {X: hw, Y: -hh}, {X: hw, Y: hh}, {X: -hw, Y: hh},
	}}.Transformed(o.X, o.Y)
}
