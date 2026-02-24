package game

import (
	"oinakos/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
)

type Obstacle struct {
	X, Y   float64 // Grid positions
	Config *ObstacleConfig
	Health int
	Alive  bool
}

func NewObstacle(x, y float64, config *ObstacleConfig) *Obstacle {
	hp := 0 // Default indestructible
	if config != nil {
		hp = config.Health
	}

	return &Obstacle{
		X:      x,
		Y:      y,
		Config: config,
		Health: hp,
		Alive:  true,
	}
}

func (o *Obstacle) TakeDamage(amount int) {
	if !o.Alive {
		return
	}
	// Indestructible obstacles have 0 or negative health starting point
	if o.Config == nil || o.Config.Health <= 0 {
		return
	}

	o.Health -= amount
	if o.Health <= 0 {
		o.Alive = false
	}
}

func (o *Obstacle) GetFootprint() engine.Polygon {
	// Dynamically map from config or fallback
	hw, hh := 0.4, 0.4
	if o.Config != nil {
		hw = o.Config.FootprintWidth
		hh = o.Config.FootprintHeight
	}

	return engine.Polygon{Points: []engine.Point{
		{X: -hw, Y: -hh}, {X: hw, Y: -hh}, {X: hw, Y: hh}, {X: -hw, Y: hh},
	}}.Transformed(o.X, o.Y)
}

func (o *Obstacle) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if o.Config == nil || o.Config.Image == nil || !o.Alive {
		return
	}

	isoX, isoY := engine.CartesianToIso(o.X, o.Y)

	op := &ebiten.DrawImageOptions{}
	scale := o.Config.Scale

	sw, sh := o.Config.Image.Size()

	// Pivot point for isometric depth
	pivotX := float64(sw) * scale / 2
	pivotY := float64(sh) * scale * 0.85

	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(isoX+offsetX-pivotX, isoY+offsetY-pivotY)

	screen.DrawImage(o.Config.Image, op)
}
