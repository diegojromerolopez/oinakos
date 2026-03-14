package game

import (
	"fmt"
	"math"

	_ "image/png"

	"oinakos/internal/engine"
)

type Projectile struct {
	X, Y             float64
	Dx, Dy           float64
	Speed            float64
	Damage           int
	Alive            bool
	IsPlayer         bool    // true if fired by player, false if by NPC (to prevent friendly-fire on themselves)
	MaxRange         float64 // Despawn after this distance
	DistanceTraveled float64
}

func NewProjectile(x, y, dx, dy, speed float64, damage int, isPlayer bool, maxRange float64) *Projectile {
	// Normalize dx/dy
	mag := math.Sqrt(dx*dx + dy*dy)
	if mag != 0 {
		dx /= mag
		dy /= mag
	}
	return &Projectile{
		X:        x,
		Y:        y,
		Dx:       dx,
		Dy:       dy,
		Speed:    speed,
		Damage:   damage,
		Alive:    true,
		IsPlayer: isPlayer,
		MaxRange: maxRange,
	}
}

func (p *Projectile) Update(ctx *SystemContext) {
	if !p.Alive {
		return
	}
	p.X += p.Dx * p.Speed
	p.Y += p.Dy * p.Speed
	p.DistanceTraveled += p.Speed

	if p.MaxRange > 0 && p.DistanceTraveled >= p.MaxRange {
		p.Alive = false
		return
	}

	// Check environment collision
	pFootprint := engine.Polygon{Points: []engine.Point{
		{X: -0.1, Y: -0.1}, {X: 0.1, Y: -0.1}, {X: 0.1, Y: 0.1}, {X: -0.1, Y: 0.1},
	}}.Transformed(p.X, p.Y)

	for _, o := range ctx.World.Obstacles {
		if !o.Alive {
			continue
		}
		if engine.CheckCollision(pFootprint, o.GetFootprint()) {
			p.Alive = false // Hits a wall/tree
			return
		}
	}

	// For now, projectiles are enemy arrows aiming at player
	mc := ctx.World.PlayableCharacter
	if !p.IsPlayer && mc.IsAlive() {
		dist := math.Sqrt(math.Pow(mc.X-p.X, 2) + math.Pow(mc.Y-p.Y, 2))
		if dist < 0.6 {
			// Hit!
			protection := mc.GetTotalProtection()
			finalDmg := int(math.Max(1, float64(p.Damage-protection)))
			DebugLog("Projectile HIT Player for %d damage", finalDmg)
			mc.TakeDamage(finalDmg, ctx)

			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text:  fmt.Sprintf("-%d", finalDmg),
				X:     mc.X,
				Y:     mc.Y,
				Life:  45,
				Color: ColorHarm,
			})
			p.Alive = false
		}
	}
}
