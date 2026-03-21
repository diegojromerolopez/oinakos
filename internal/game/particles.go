package game

import (
	"math"
	"math/rand"
)

type ParticleType int

const (
	ParticleRain ParticleType = iota
	ParticleSnow
)

type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    int
	MaxLife int
	Type    ParticleType
}

func (g *Game) updateWeather() {
	if g.audio != nil {
		switch g.CurrentWeather {
		case WeatherRain:
			g.audio.PlayAmbientLoop("ambient/rain")
			g.audio.StopAmbientLoop("ambient/storm")
		case WeatherStorm:
			g.audio.PlayAmbientLoop("ambient/storm")
			g.audio.StopAmbientLoop("ambient/rain")
		default:
			g.audio.StopAmbientLoop("ambient/rain")
			g.audio.StopAmbientLoop("ambient/storm")
		}
	}

	if g.CurrentWeather == WeatherClear {
		if len(g.particles) > 0 {
			g.particles = g.particles[:0]
		}
		return
	}

	// Spawn particles
	spawnRate := int(g.WeatherIntensity * 8)
	if g.CurrentWeather == WeatherRain || g.CurrentWeather == WeatherStorm {
		for i := 0; i < spawnRate; i++ {
			g.spawnParticle(ParticleRain)
		}
	} else if g.CurrentWeather == WeatherSnow {
		if g.Tick%4 == 0 {
			for i := 0; i < spawnRate/2+1; i++ {
				g.spawnParticle(ParticleSnow)
			}
		}
	}

	// Update particles
	activeParticles := g.particles[:0]
	for _, p := range g.particles {
		p.X += p.VX
		p.Y += p.VY
		p.Life--

		if p.Type == ParticleSnow {
			// Swaying effect for snow
			p.VX = math.Sin(float64(g.Tick)/20.0+p.X) * 0.5
		}

		// Keep particle if it's still alive and on screen (roughly)
		if p.Life > 0 && p.Y < float64(g.height)+20 && p.X > -220 && p.X < float64(g.width)+220 {
			activeParticles = append(activeParticles, p)
		}
	}
	g.particles = activeParticles
}

func (g *Game) spawnParticle(t ParticleType) {
	// Use a wider spawn area to account for slant and ensure full screen coverage
	spawnWidth := float64(g.width) + 400
	p := &Particle{
		X:    (rand.Float64() * spawnWidth) - 200,
		Y:    -20,
		Type: t,
		Life: 200 + rand.Intn(100),
	}

	if t == ParticleRain {
		p.VY = 12 + rand.Float64()*8
		p.VX = 1.5 // Slanted rain
	} else if t == ParticleSnow {
		p.VY = 1.5 + rand.Float64()*1.5
		p.VX = rand.Float64()*0.5 - 0.25
	}

	g.particles = append(g.particles, p)
}
