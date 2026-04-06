package game

import (
	"oinakos/internal/engine"
	"math"
)

type Obstacle struct {
	ID            string  // Unique instance ID
	X, Y          float64 // Grid positions
	Z             float64
	Archetype     *ObstacleArchetype
	HealthPoints  int
	MaxHealthPoints int
	WeightLeft    float64
	CooldownTicks int
	TickCounter   int
	Alive         bool
	EffectTimers  map[ActorInterface]int 

	CachedFootprint *engine.Polygon 

	GrowthTicks int
	GrowthStage int 

	Inventory   []*ItemInstance
	TotalWeight float64
	OwnerID     string
	Locked      bool
	LockHealth  int
	LockBroken  bool
}

func (o *Obstacle) IsColliding(targetX, targetY, radius float64) bool {
	if !o.Alive { return false }
	dist := math.Sqrt(math.Pow(o.X-targetX, 2) + math.Pow(o.Y-targetY, 2))
	// Optimize: quick bounding circle check before polygon
	if dist > 5.0 { return false }
	
	p := o.GetFootprint()
	// Traditional Point-in-polygon logic
	return p.Contains(targetX, targetY)
}

func NewObstacle(id string, x, y float64, config *ObstacleArchetype) *Obstacle {
	hp := 0
	weight := 0.0
	if config != nil {
		hp = config.HealthPoints
		if config.Timber > 0 { hp = config.Timber }
		weight = config.Weight
		if weight <= 0 && config.Timber > 0 { weight = float64(config.Timber) }
		if weight > 0 && hp == 0 { hp = int(weight) }
	}

	obs := &Obstacle{
		ID:              id,
		X:               x,
		Y:               y,
		Archetype:       config,
		HealthPoints:    hp,
		MaxHealthPoints: hp,
		WeightLeft:      weight,
		CooldownTicks:   0,
		Alive:           true,
		Inventory:       make([]*ItemInstance, 0),
		EffectTimers:    make(map[ActorInterface]int),
		OwnerID:         "",
		Locked:          false,
		LockHealth:      0,
	}

	if config != nil {
		if config.OwnerID != "" {
			obs.OwnerID = config.OwnerID
		}
		if config.LockResistance > 0 {
			obs.Locked = true
			obs.LockHealth = config.LockResistance
		}
	}
	return obs
}

func (o *Obstacle) TryStash(it *ItemInstance) bool {
	if it == nil { return false }
	capacity := 0.0
	if o.Archetype != nil { capacity = o.Archetype.MaxCapacity }
	// If capacity is 0, this obstacle is not a container and cannot hold items
	if capacity <= 0 { return false }
	if o.TotalWeight + it.Weight > capacity { return false }
	o.Inventory = append(o.Inventory, it)
	o.TotalWeight += it.Weight
	return true
}

func (o *Obstacle) TryRetrieve(index int) *ItemInstance {
	if index < 0 || index >= len(o.Inventory) { return nil }
	it := o.Inventory[index]
	o.Inventory = append(o.Inventory[:index], o.Inventory[index+1:]...)
	o.TotalWeight -= it.Weight
	return it
}

func (o *Obstacle) Update() {
	if !o.Alive { return }
	if o.CooldownTicks > 0 { o.CooldownTicks-- }
	o.TickCounter++
	for entity, ticks := range o.EffectTimers { if ticks > 0 { o.EffectTimers[entity] = ticks - 1 } }
	if o.Archetype != nil {
		if o.Archetype.IsCrop && o.GrowthStage < 2 {
			growthLimit := o.Archetype.GrowthDuration
			if growthLimit > 0 {
				o.GrowthTicks++
				if o.GrowthStage == 0 && o.GrowthTicks > growthLimit/2 { o.GrowthStage = 1 } else if o.GrowthStage == 1 && o.GrowthTicks >= growthLimit { o.GrowthStage = 2 }
			}
		}
		if o.Archetype.ID == "defecation" && o.TickCounter > TicksPerMonth * 2 {
			o.Alive = false
		}
		if o.Archetype.ID == "urination" && o.TickCounter > TicksPerDay * 2 {
			o.Alive = false
		}
	}
}

func (o *Obstacle) TakeDamage(amount int) {
	if !o.Alive || o.Archetype == nil || !o.Archetype.Destructible { return }
	o.HealthPoints -= amount
	if o.HealthPoints <= 0 { o.Alive = false }
}

func (o *Obstacle) GetFootprint() engine.Polygon {
	if o.CachedFootprint != nil { return *o.CachedFootprint }
	var poly engine.Polygon
	if o.Archetype != nil && len(o.Archetype.Footprint) > 0 {
		poly = engine.Polygon{Points: make([]engine.Point, len(o.Archetype.Footprint))}
		for i, p := range o.Archetype.Footprint { poly.Points[i] = engine.Point{X: p.X, Y: p.Y} }
	} else {
		poly = engine.Polygon{Points: []engine.Point{ {X: -0.2, Y: -0.2}, {X: 0.2, Y: -0.2}, {X: 0.2, Y: 0.2}, {X: -0.2, Y: 0.2} }}
	}
	transformed := poly.Transformed(o.X, o.Y)
	o.CachedFootprint = &transformed
	return transformed
}

func (o *Obstacle) GetIsoPos() (float64, float64) { return engine.CartesianToIso(o.X, o.Y) }

func (o *Obstacle) GetSortY() float64 {
	sortY := o.X + o.Y
	if o.Archetype != nil {
		if o.Archetype.Type == "static" || o.Archetype.Type == "well" { sortY += 2.0 } else {
			p := o.GetFootprint()
			minX, minY, maxX, maxY := p.Bounds()
			sortY = (minX + maxX + minY + maxY) / 2
		}
	}
	return sortY
}
