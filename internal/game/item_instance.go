package game

import (
	"oinakos/internal/engine"
)

type ItemInstance struct {
	ID         string
	Config     *ObjectConfig
	Resistance int
	Weight     float64 `json:"weight"`
	HoursLeft  float64 `json:"hours_left"` // Runtime hours until expiration
	DroppedBy  string  `json:"dropped_by,omitempty"`
	X, Y, Z    float64
	Pickable   bool
}

func NewItemInstance(id string, config *ObjectConfig, x, y float64) *ItemInstance {
	res := 0
	weight := 0.0
	hours := 0.0
	if config != nil {
		res = config.Resistance
		weight = config.Weight
		hours = config.MaxHours
	}
	return &ItemInstance{
		ID:         id,
		Config:     config,
		Resistance: res,
		Weight:     weight,
		HoursLeft:  hours,
		X:          x,
		Y:          y,
		Z:          0,
		Pickable:   true,
	}
}

func (it *ItemInstance) Draw(screen engine.Image, offsetX, offsetY float64) {
	if it.Config == nil || it.Config.Sprite == nil {
		return
	}
	isoX, isoY := engine.CartesianToIsoZ(it.X, it.Y, it.Z)
	
	w, h := it.Config.Sprite.Size()
	op := engine.NewDrawImageOptions()
	
	// Center the sprite on the isometric point
	tx := isoX + offsetX - float64(w)/2
	ty := isoY + offsetY - float64(h)*0.9 // Grounded item anchoring
	
	op.Translate(tx, ty)
	screen.DrawImage(it.Config.Sprite, op)
}
func (it *ItemInstance) GetFootprint() engine.Polygon {
	var poly engine.Polygon
	if it.Config != nil && len(it.Config.Footprint) > 0 {
		poly = engine.Polygon{Points: make([]engine.Point, len(it.Config.Footprint))}
		for i, p := range it.Config.Footprint {
			poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
		}
	} else {
		// Fallback footprint
		poly = engine.Polygon{Points: []engine.Point{
			{X: -0.2, Y: -0.2}, {X: 0.2, Y: -0.2}, {X: 0.2, Y: 0.2}, {X: -0.2, Y: 0.2},
		}}
	}
	return poly.Transformed(it.X, it.Y)
}

func (it *ItemInstance) Update(ctx *SystemContext) {
	if it.Config == nil || it.Config.MaxHours <= 0 {
		return
	}

	// 1 Hour = 720 ticks (DayCycle threshold in game_loop.go)
	// it.HoursLeft -= 1.0 / 720.0
	it.HoursLeft -= (1.0 / 720.0)
	if it.HoursLeft < 0 {
		it.HoursLeft = 0
	}
}
