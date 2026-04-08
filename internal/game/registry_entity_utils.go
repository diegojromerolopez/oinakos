package game

import "oinakos/internal/engine"

func (e *EntityConfig) PickAttackImage(seed int) engine.Image {
	if e.Attack1Image != nil {
		if e.Attack2Image != nil { if seed%2 == 0 { return e.Attack1Image }; return e.Attack2Image }
		return e.Attack1Image
	}
	return e.AttackImage
}

func (e *EntityConfig) PickHitImage(seed int) engine.Image {
	if e.Hit1Image != nil {
		if e.Hit2Image != nil { if seed%2 == 0 { return e.Hit1Image }; return e.Hit2Image }
		return e.Hit1Image
	}
	return e.HitImage
}

func (c *EntityConfig) GetFootprint() engine.Polygon {
	if c.CachedBaseFootprint != nil { return *c.CachedBaseFootprint }
	if len(c.Footprint) == 0 {
		p := engine.Polygon{Points: []engine.Point{{X: -0.15, Y: -0.15}, {X: 0.15, Y: -0.15}, {X: 0.15, Y: 0.15}, {X: -0.15, Y: 0.15}}}
		c.CachedBaseFootprint = &p; return p
	}
	p := engine.Polygon{Points: make([]engine.Point, len(c.Footprint))}
	for i, f := range c.Footprint { p.Points[i] = engine.Point{X: f.X, Y: f.Y} }
	c.CachedBaseFootprint = &p; return p
}
