package game

import (
	"math"
	"math/rand"
)

func (c *Character) MoveTo(ctx *SystemContext, tx, ty float64) { c.ExecutePathTo(ctx, tx, ty) }

func (c *Character) ExecutePathTo(ctx *SystemContext, tx, ty float64) {
	dist := math.Sqrt(math.Pow(c.X-tx, 2) + math.Pow(c.Y-ty, 2))
	if dist < 0.5 {
		c.Path, c.ActionState = nil, ActorIdle
		return
	}
	if len(c.Path) == 0 || c.PathTimer <= 0 {
		if ctx.World != nil && ctx.World.Game != nil {
			c.Path, c.PathTimer = ctx.World.Game.FindAStarPath(c.X, c.Y, tx, ty), 120+rand.Intn(60)
		} else {
			c.PathTimer = 10 // Retry soon if world not ready
		}
	} else {
		c.PathTimer--
	}
	if len(c.Path) > 0 {
		nP := c.Path[0]
		if math.Sqrt(math.Pow(c.X-nP.X, 2)+math.Pow(c.Y-nP.Y, 2)) < 0.4 {
			c.Path = c.Path[1:]
			if len(c.Path) > 0 {
				nP = c.Path[0]
			}
		}
		c.executeMovement(ctx, nP.X-c.X, nP.Y-c.Y, ctx.World.Obstacles, false)
	} else {
		c.executeMovement(ctx, tx-c.X, ty-c.Y, ctx.World.Obstacles, false)
	}
}

func (c *Character) executeMovement(ctx *SystemContext, dx, dy float64, obstacles []*Obstacle, flee bool) {
	if c.ActionState == ActorIncapacitated && c.ActionState != ActorBerserk { return }
	mag := math.Sqrt(dx*dx + dy*dy)
	if mag < 0.01 { return }
	mvX, mvY := dx/mag, dy/mag
	if flee { mvX, mvY = -mvX, -mvY }
	spd := c.Speed * c.GetSpeedModifier(ctx)
	moveDist := spd
	
	if !flee && moveDist > mag { moveDist = mag }

	// 1. Try direct movement
	if !c.checkCollisionAt(c.X+mvX*moveDist, c.Y+mvY*moveDist, obstacles) {
		c.X, c.Y, c.ActionState = c.X+mvX*moveDist, c.Y+mvY*moveDist, ActorWalking
		c.updateFacing(mvX, mvY)
		return
	}

	// 2. Slanted movement (Sliding) as fallback
	angles := []float64{math.Pi / 4, -math.Pi / 4}
	for _, angle := range angles {
		cos, sin := math.Cos(angle), math.Sin(angle)
		rx, ry := mvX*cos - mvY*sin, mvX*sin + mvY*cos
		if !c.checkCollisionAt(c.X+rx*moveDist, c.Y+ry*moveDist, obstacles) {
			c.X, c.Y, c.ActionState = c.X+rx*moveDist, c.Y+ry*moveDist, ActorWalking
			c.updateFacing(rx, ry)
			return
		}
	}

	c.ActionState = ActorIdle
	if len(c.Path) > 0 { c.PathTimer = 0 }
}
