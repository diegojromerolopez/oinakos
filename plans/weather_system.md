# Plan: Weather System 🌦️

## Objective
Introduce dynamic weather effects (rain, snow, storm) to enhance the atmospheric immersion of Oinakos maps.

## Analysis
Weather should be both a visual effect and a potential mechanical modifier (e.g., rain slowing down movement or causing slipping).

### Key Requirements:
1. **Weather State**: `WeatherType` enum in the `Game` struct.
2. **Particle System**: A lightweight particle engine to handle falling rain/snow.
3. **Audio**: Ambient sound loops for different weather conditions.
4. **Mechanical Impact**: Slight speed modifiers or visibility changes.

---

## Implementation Details

### 1. Data Structures (`internal/game/game.go`)
- Define `WeatherType` enum: `Clear`, `Rain`, `Snow`, `Storm`.
- Add `CurrentWeather WeatherType` and `WeatherIntensity float64` (0.0 to 1.0) to `Game`.

### 2. Particle Engine (`internal/game/particles.go`)
- Create a simple `Particle` struct: `X, Y, VX, VY, Life, Type`.
- Update loop: Particles move and die when they leave the screen or reach their lifespan.
- **Rain**: Thin slanted blue/white lines. High vertical velocity.
- **Snow**: Slow-falling white circles/dots with horizontal swaying (sine wave).
- **Optimization**: Use a pre-allocated slice of particles to avoid GC pressure. Render particles in screen-space.

### 3. Rendering (`internal/game/game_render.go`)
- Add `drawWeather(screen engine.Image)` to `GameRenderer`.
- For `Storm`, implement a "Lightning Flash" effect: occasionally fill the screen with bright white (low alpha) for 1-2 frames.
- Apply a slight "wet" tint or grey overlay to the world during rain/storm.

### 4. Audio & Map Config
- Update `internal/game/audio.go` to handle looping ambient tracks (e.g., `assets/audio/ambient/rain.wav`).
- Add `weather: "rain"` to map YAML files in `data/maps/`.

---

## Verification
- Load a map set to `weather: rain`. Verify rain particles fall correctly and the "rain" loop plays.
- Switch to `snow` via debug menu and verify the slow, swaying falling behavior.
- Ensure performance remains stable (60 FPS) with ~1000 active particles.
