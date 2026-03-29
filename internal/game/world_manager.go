package game

type WorldManager struct {
	game *Game
}

func NewWorldManager(g *Game) *WorldManager {
	return &WorldManager{game: g}
}

func (wm *WorldManager) UpdateChunks() {
	// Procedural spawning disabled
}

const TicksPerShift = TicksPerDay / 3 // 8 in-game hours

func (wm *WorldManager) UpdateDayCycle() {
	g := wm.game
	step := 1
	switch g.settings.TimePace {
	case "real":
		step = 1
	case "double":
		step = 2
	case "fast":
		step = 10
	case "standard":
		step = 60 // 24 minutes real-time per day
	case "month":
		step = 43200 // 1 month per minute
	case "year":
		step = 518400 // 1 year per minute
	}

	g.World.DayTick = (g.World.DayTick + step) % TicksPerDay
	
	shift := LaborShift(g.World.DayTick / TicksPerShift)
	for _, c := range g.characters {
		c.Shift = shift
	}
}
