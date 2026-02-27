# Unique NPCs & Dynamic Palette Swapping

## Overview
Currently, all NPCs (like Orcs or Demons) look identical across the entire map, and share the exact same stats defined by their `Archetype`. To introduce minibosses, named characters, and visual variety without requiring hundreds of new sprite sheets, we will implement a dynamic Color Palette swap system and a "Unique NPC" definition layer.

## Implementation Steps

### 1. The Asset Preparation (Color Masks)
- Modify our base archetype images (e.g., `knight/static.png`).
- Paint primary accent elements (like the tunic, core armor, or main weapon glow) with a pure, bright purple color (`#FF00FF`). This is the **Primary Mask**.
- Paint secondary accents (like the cape or shield trim) with a pure, bright yellow color (`#FFFF00`). This is the **Secondary Mask**.
- When rendering in-game, these strident masks will be replaced by the engine.

### 2. The `data/npcs/` Definitions
- Create a new directory alongside `data/archetypes` called `data/npcs/`.
- While Archetypes describe a generic *Class* (e.g., "Knight"), the new `NPCConfig` describes a *Specific Entity*.
- Example `data/npcs/green_knight.yaml`:
  ```yaml
  id: green_knight
  name: Green Knight
  description: A legendary warrior bound to the forest.
  archetype: knight
  unique: true
  primary_color: "#00FF00"    # Swaps out Purple
  secondary_color: "#000000" # Swaps out Yellow
  scale_multiplier: 1.5      # Bosses should be physically larger
  health_min: 500
  speed: 1.5
  unique_drop: "forest_amulet" # Specific item dropped on death
  ```

### 3. Dynamic Palette Swapping logic (Kage Shader)
- We MUST use an Ebitengine `*ebiten.Shader` (Kage language) to perform the color replacement on the GPU. Doing pixel-by-pixel replacement on the CPU in Go every frame will ruin the framerate.
- Write a simple `paletteswap.go` shader that reads the original pixel color. If the color is `vec4(1.0, 0.0, 1.0, 1.0)` (Purple), output the uniform `PrimaryColor`. If it's `vec4(1.0, 1.0, 0.0, 1.0)` (Yellow), output `SecondaryColor`.
- Update `internal/game/npc_render.go`: When drawing an NPC with custom colors, use `DrawRectShader` with the sprite bound as `image0`, passing the hex colors converted to `[]float32` RGBA uniforms.

### 4. UI: Names and Hover Descriptions
- **Names**: Draw the `NPC.Name` string directly beneath the NPC's feet (using `textRenderer.DebugPrintAt` or a custom font engine) if the entity has a name defined. Use a different font color (e.g., Gold) if `unique: true`.
- **Hover Tooltips**: In `game.go`, check the `mouseX` and `mouseY` (translated from Screen to Isometric space). If the cursor bounding box overlaps an NPC's isometric footprint, draw a floating UI panel containing the `NPC.Description`.

### 5. Unique Constraints & Elite Spawning
- **The Dead List**: Create a global `g.KilledUniqueEnemies = make(map[string]bool)`.
- When the chunk spawner selects an NPC, check its config. If `unique: true` AND `g.KilledUniqueEnemies[npc.id]` is true, abort the spawn or fallback to a generic pool. 
- This ensures named bosses (like the Green Knight) never appear twice in a single run.
- **Elite Upgrades**: If `unique: false` but colors are provided, the spawner can occasionally (e.g., 5% chance) load the custom NPC profile to create generic "Elite" enemies with boosted stats and alternate palettes, keeping infinite combat fresh.

## Challenges
- **Shader Pipeline**: Integrating `*ebiten.Shader` requires adjusting the main draw loop to ensure Z-ordering (Y-sorting) remains intact when switching between standard `DrawImage` (for normal NPCs) and `DrawRectShader` (for Palette Swapped NPCs).
- **Mouse Picking in Isometric Space**: Translating raw screen X/Y to the correct isometric Z-indexed entity for the hover description requires precise reverse-matrix math and hit-box validation.
