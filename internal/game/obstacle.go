package game

import (
	"oinakos/internal/engine"
)

type Obstacle struct {
	ID            string  // Unique instance ID (e.g. from PreSpawnObstacle)
	X, Y          float64 // Grid positions
	Z             float64
	Archetype     *ObstacleArchetype
	HealthPoints  int
	MaxHealthPoints int
	WeightLeft    float64
	CooldownTicks int
	TickCounter   int
	Alive         bool
	EffectTimers  map[ActorInterface]int // Track intervals for hazards/healing per entity

	CachedFootprint *engine.Polygon // Optimization: obstacles don't move, cache world footprint

	// Growth state for crops
	GrowthTicks int
	GrowthStage int // 0: Seeded, 1: Growing, 2: Mature

	// Storage
	Inventory   []*ItemInstance
	TotalWeight float64
}

func NewObstacle(id string, x, y float64, config *ObstacleArchetype) *Obstacle {
	hp := 0 // Default indestructible
	weight := 0.0
	if config != nil {
		hp = config.HealthPoints
		if config.Timber > 0 {
			hp = config.Timber
		}
		weight = config.Weight
		if weight <= 0 && config.Timber > 0 {
			weight = float64(config.Timber)
		}
		// If it's a resource/tree, set health to weight units if not explicitly set
		if weight > 0 && hp == 0 {
			hp = int(weight)
		}
	}

	return &Obstacle{
		ID:              id,
		X:               x,
		Y:               y,
		Archetype:       config,
		HealthPoints:    hp,
		MaxHealthPoints: hp,
		WeightLeft:      weight,
		CooldownTicks:   0,
		Alive:           true,
		EffectTimers:    make(map[ActorInterface]int),
		Inventory:       make([]*ItemInstance, 0),
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

	// Growth logic if it is a crop
	if o.Archetype != nil && o.Archetype.IsCrop && o.GrowthStage < 2 {
		growthLimit := o.Archetype.GrowthDuration
		if growthLimit > 0 {
			o.GrowthTicks++
			if o.GrowthStage == 0 && o.GrowthTicks > growthLimit/2 {
				o.GrowthStage = 1
			} else if o.GrowthStage == 1 && o.GrowthTicks >= growthLimit {
				o.GrowthStage = 2
			}
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

	o.HealthPoints -= amount
	if o.HealthPoints <= 0 {
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
func (o *Obstacle) GetIsoPos() (float64, float64) {
	return engine.CartesianToIso(o.X, o.Y)
}

func (o *Obstacle) GetSortY() float64 {
	sortY := o.X + o.Y
	if o.Archetype != nil {
		if o.Archetype.Type == "static" || o.Archetype.Type == "well" {
			sortY += 2.0
		} else {
			p := o.GetFootprint()
			minX, minY, maxX, maxY := p.Bounds()
			sortY = (minX + maxX + minY + maxY) / 2
		}
	}
	return sortY
}
