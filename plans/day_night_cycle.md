# Plan: Day/Night Cycle 🌓

## Objective
Implement a dynamic global time system that shifts the map's visual ambiance between day and night.

## Analysis
The game needs a way to track "world time" and apply a visual filter across the entire map. This should be a global transformation that affects sprites and tiles consistently.

### Key Requirements:
1. **Time Tracking**: A `TimeOfDay` value (0.0 to 1.0) in the `Game` struct. 0.0 is midday, 0.5 is midnight.
2. **Visual Transformation**: A global tint applied during the rendering pass.
3. **Map Integration**: Maps should be able to specify a starting time or a fixed time (e.g., a "Midnight Raid" map).
4. **Light Sources**: Localized areas (like around the player or torches) should be exempt from the night tint to create a "vision" effect.

---

## Implementation Details

### 1. Game State Updates (`internal/game/game.go`)
- Add `TimeOfDay float64` to the `Game` struct.
- In `Game.Update`, if the map is not "fixed time", increment `TimeOfDay` by a small amount each tick (e.g., `1.0 / (60 * 60 * 10)` for a 10-minute full cycle at 60 TPS).
- Add `StartTime float64` and `FixedTime bool` to the map YAML structure to allow level-specific lighting.

### 2. Color Transformation (`internal/game/game_render.go`)
- Calculate a `NightAlpha` based on `TimeOfDay`. A cosine wave works well: `math.Max(0, math.Cos(TimeOfDay * 2 * math.Pi + math.Pi))` will give 0.0 at noon and 1.0 at midnight.
- In `GameRenderer.Draw`, after drawing all entities but before the HUD:
    - Draw a full-screen semi-transparent rectangle with a deep blue/purple color (e.g., `rgba(15, 15, 45, NightAlpha * 180)`).
- **Advanced (Shader)**: Alternatively, use a shader that adjusts brightness and blue-tint based on the time.

### 3. Light Sources (Torches)
- Utilize the existing `torchImage` in `GameRenderer`.
- Use `CompositeModeDestinationOut` with the torch image centered on the player to "cut out" the night overlay, revealing the true colors underneath. This creates a realistic "fog of war" or "vision" circle.

### 4. Configuration Changes (`data/maps/`)
- Add `start_time` and `is_static_time` fields to map YAMLs.
- `start_time: 0.5` would spawn the player at midnight.

---

## Verification
- Load `starter_village`. Wait 5 minutes and observe the gradual transition from bright day to deep blue night.
- Ensure the HUD remains fully legible (drawn after the overlay).
- Verify that the player's "torch" light allows them to see sprites clearly at night.
