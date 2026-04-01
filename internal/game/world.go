package game

import (
	"image"
)
const (
	TicksPerSecond = 60
	TicksPerHour   = 720 // 12 seconds IRL
	TicksPerDay    = TicksPerHour * 24
	TicksPerMonth  = 30 * TicksPerDay // Aligning with 360-day year standard
	TicksPerSeason = 3 * TicksPerMonth
	TicksPerYear   = 12 * TicksPerMonth // 360 Days per year
)

// World holds all live game entities and spatial data.
type World struct {
	PlayableCharacter *Character
	Characters        []*Character
	Obstacles         []*Obstacle
	Projectiles       []*Projectile
	FloatingTexts     []*FloatingText
	
	CurrentMapType    *MapType
	ExploredTiles     map[image.Point]bool
	PlayTime          float64
	DayTick           int      // 0 - 5,184,000 (ticks in 24h)
	Game              *Game
	Items             []*ItemInstance
	State             WorldState
}

type Season int

const (
	SeasonSpring Season = iota
	SeasonSummer
	SeasonAutumn
	SeasonWinter
)

func (s Season) String() string {
	switch s {
	case SeasonSpring:
		return "SPRING"
	case SeasonSummer:
		return "SUMMER"
	case SeasonAutumn:
		return "AUTUMN"
	case SeasonWinter:
		return "WINTER"
	default:
		return "UNKNOWN"
	}
}

type WeatherType int

const (
	WeatherClear WeatherType = iota
	WeatherRain
	WeatherSnow
	WeatherStorm
	WeatherFog
)
func (w WeatherType) String() string {
	switch w {
	case WeatherClear: return "CLEAR"
	case WeatherRain:  return "RAIN"
	case WeatherSnow:  return "SNOW"
	case WeatherStorm: return "STORM"
	case WeatherFog:   return "FOG"
	default:           return "UNKNOWN"
	}
}

type WorldState struct {
	Ticks        int
	Hour         int     // 0-23
	Day          int     // 1 to 48 (12 days per season)
	Month        int     // 1 to 12
	Season       Season  // Calculated from Month
	Temperature  float64 // Celsius
	Weather      WeatherType
	Intensity    float64 // 0.0 to 1.0
	WeatherTimer int
	GroupSentiment map[string]map[string]float64 // FactionA -> FactionB -> Sentiment
}

func NewWorld() *World {
	return &World{
		Characters:    make([]*Character, 0),
		Obstacles:     make([]*Obstacle, 0),
		Projectiles:   make([]*Projectile, 0),
		FloatingTexts: make([]*FloatingText, 0),
		ExploredTiles: make(map[image.Point]bool),
		State: WorldState{
			Season: SeasonSpring, 
			Hour: 12, 
			Day: 1, 
			Month: 4, 
			Temperature: 18.0,
			GroupSentiment: make(map[string]map[string]float64),
		},
	}
}
