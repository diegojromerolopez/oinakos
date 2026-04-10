package game

import (
	"fmt"
	"math"

	_ "image/png"

	"oinakos/internal/engine"
)

type Projectile struct {
	X, Y, Z          float64
	Dx, Dy, Dz       float64
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
		Z:        0, // Update this inside game_update or at spawn if 3D targeting is implemented
		Dx:       dx,
		Dy:       dy,
		Dz:       0,
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
	pCircle := engine.Circle{X: p.X, Y: p.Y, Radius: 0.1}

	for _, o := range ctx.World.Obstacles {
		if !o.Alive {
			continue
		}
		if engine.CheckCirclePolygonCollision(pCircle, o.GetFootprint()) {
			p.Alive = false // Hits a wall/tree
			return
		}
	}

	// Projectile collision with characters
	if p.IsPlayer {
		for _, n := range ctx.World.Characters {
			if n != nil && n.IsAlive() && !n.IsPlayerControlled {
				dist := math.Sqrt(math.Pow(n.X-p.X, 2) + math.Pow(n.Y-p.Y, 2))
				if dist < 0.8 {
					protection := n.GetTotalProtection()
					finalDmg := int(math.Max(1, float64(p.Damage-protection)))
					n.TakeDamage(finalDmg, ctx.World.PlayableCharacter, ctx)
					ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
						Text:  fmt.Sprintf("-%d", finalDmg),
						X:     n.X,
						Y:     n.Y,
						Life:  45,
						Color: ColorHarm,
					})
					p.Alive = false
					return
				}
			}
		}
	} else {
		mc := ctx.World.PlayableCharacter
		if mc != nil && mc.IsAlive() {
			dist := math.Sqrt(math.Pow(mc.X-p.X, 2) + math.Pow(mc.Y-p.Y, 2))
			if dist < 0.6 {
				protection := mc.GetTotalProtection()
				finalDmg := int(math.Max(1, float64(p.Damage-protection)))
				mc.TakeDamage(finalDmg, nil, ctx)
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
}
