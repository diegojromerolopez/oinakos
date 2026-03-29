package game

import (
	"math"
	"math/rand"
)

func (g *Game) updateItemExpiration(ctx *SystemContext) {
	if g.World == nil { return }
	for _, it := range g.World.Items { if it != nil { it.Update(ctx) } }
	if g.playableCharacter != nil {
		for _, it := range g.playableCharacter.Inventory { if it != nil { it.Update(ctx) } }
		for _, it := range g.playableCharacter.Slots { if it != nil { it.Update(ctx) } }
	}
	for _, n := range g.characters {
		if n != nil {
			for _, it := range n.Inventory { if it != nil { it.Update(ctx) } }
			for _, it := range n.Slots { if it != nil { it.Update(ctx) } }
		}
	}
}

func (g *Game) updateWorldState() {
	if g.World == nil { return }
	s := &g.World.State; s.Ticks++
	if s.Ticks >= 720 {
		s.Ticks, s.Hour = 0, s.Hour+1
		if s.Hour >= 24 {
			s.Hour, s.Day = 0, s.Day+1
			if s.Month = ((s.Day - 1) / 4) + 1; s.Month > 12 { s.Month, s.Day = 1, 1 }
			if s.Month >= 3 && s.Month <= 5 { s.Season = SeasonSpring } else if s.Month >= 6 && s.Month <= 8 { s.Season = SeasonSummer } else if s.Month >= 9 && s.Month <= 11 { s.Season = SeasonAutumn } else { s.Season = SeasonWinter }
		}
	}
	baseTemp := 15.0; if s.Season == SeasonSummer { baseTemp = 28.0 } else if s.Season == SeasonAutumn { baseTemp = 12.0 } else if s.Season == SeasonWinter { baseTemp = -2.0 }
	s.Temperature = baseTemp + math.Sin(float64(s.Hour-10)*math.Pi/12.0)*8.0
	if s.Weather == WeatherRain { s.Temperature -= 4.0 } else if s.Weather == WeatherSnow { s.Temperature -= 7.0 }
	if s.WeatherTimer > 0 { s.WeatherTimer-- } else {
		roll := rand.Float64(); s.WeatherTimer = 3600 + rand.Intn(7200)
		switch s.Season {
		case SeasonWinter: if roll < 0.4 { s.Weather = WeatherSnow } else if roll < 0.6 { s.Weather = WeatherClear } else { s.Weather = WeatherFog }
		case SeasonSummer: if roll < 0.1 { s.Weather = WeatherRain } else { s.Weather = WeatherClear }
		case SeasonAutumn: if roll < 0.4 { s.Weather = WeatherRain } else if roll < 0.7 { s.Weather = WeatherFog } else { s.Weather = WeatherClear }
		default: if roll < 0.2 { s.Weather = WeatherRain } else { s.Weather = WeatherClear }
		}
		s.Intensity = 0.3 + rand.Float64()*0.7
	}
}
