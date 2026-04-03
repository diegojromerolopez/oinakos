package game

import (
	"oinakos/internal/engine"
)

func (a *Actor) GetSortY() float64 {
	sortY := a.X + a.Y
	if a.ActionState == ActorDead { sortY -= 100.0 }
	return sortY
}

func (a *Actor) GetCollisionCircle() engine.Circle {
	radius := 0.4
	if a.Config != nil && a.Config.CollisionRadius > 0 { radius = a.Config.CollisionRadius }
	return engine.Circle{X: a.X, Y: a.Y, Radius: radius}
}

func (a *Actor) checkCollisionAt(nx, ny float64, obstacles []*Obstacle) bool {
	col := a.GetCollisionCircle()
	col.X, col.Y = nx, ny
	for _, o := range obstacles {
		if o.Alive && o.Archetype != nil && !o.Archetype.Passable && engine.CheckCirclePolygonCollision(col, o.GetFootprint()) {
			return true
		}
	}
	return false
}

func (a *Actor) GetSpeedModifier(ctx *SystemContext) float64 {
	switch a.CurrentTile {
	case "water.png", "dark_water.png": return 0.5
	case "mud.png": return 0.8
	default:
		multiplier := 1.0
		if a.Trauma.LeftLegLost { multiplier -= 0.5 }
		if a.Trauma.RightLegLost { multiplier -= 0.5 }
		if a.Trauma.SpineBroken { multiplier *= 0.2 }
		if multiplier < 0.1 { multiplier = 0.1 }
		if a.State.Pain > 50 { multiplier *= (1.0 - (a.State.Pain-50)/100.0) }
		if a.State.Pain > 80 { multiplier = 0 }
		if a.State.Sanity < 30 { multiplier *= 0.5 } // Depression/Lethargy speed penalty
		if ctx != nil {
			switch ctx.Weather {
			case WeatherRain: multiplier *= 0.9
			case WeatherSnow: multiplier *= 0.75
			case WeatherStorm: multiplier *= 0.85
			}
		}
		return multiplier
	}
}

func (a *Actor) AddMemory(tick int, mType, source string, value float64) {
	if a.Memories == nil { a.Memories = []MemoryEvent{} }
	a.Memories = append(a.Memories, MemoryEvent{Tick: tick, Type: mType, Source: source, Value: value})
	if len(a.Memories) > 20 { a.Memories = a.Memories[1:] }
	a.ModifySentiment(source, value)
}

func (a *Actor) AddXP(amount int) {
	a.XP += amount
	newLevel := a.XP/100 + 1
	if newLevel > a.Level {
		a.Level = newLevel
		a.State.HealthPoints = a.State.MaxHealthPoints
		if a.BodyStatus != nil { a.InitBodyStatus() }
	}
}

func (a *Actor) GetInventoryNames() []string {
	var names []string
	for _, it := range a.Inventory { if it != nil && it.Config != nil { names = append(names, it.Config.Name) } }
	for _, it := range a.Slots { if it != nil && it.Config != nil { names = append(names, it.Config.Name+" (equipped)") } }
	return names
}

func clampInt(v, min, max int) int {
	if v < min { return min }; if v > max { return max }; return v
}

func clampFloat(v, min, max float64) float64 {
	if v < min { return min }; if v > max { return max }; return v
}
