# Unique NPCs & Dynamic Palette Swapping

## Overview
Currently, all NPCs (like Orcs or Demons) look identical across the entire map, and share the exact same stats defined by their `Archetype`. To introduce minibosses, named characters, and visual variety without requiring hundreds of new sprite sheets, we will implement a dynamic Color Palette swap system and a "Unique NPC" definition layer.

## Implementation Steps

### 1. The Asset Preparation (Color Masks)
- Modify our base archetype images (e.g., `knight/static.png`).
- Paint specific elements (like the tunic or armor trim) with a pure, bright purple color (`#FF00FF`). This is the **Primary Mask**.
- Paint secondary elements (like the cape or shield) with a pure, bright yellow color (`#FFFF00`). This is the **Secondary Mask**.
- When rendering in-game, these strident colors will be swapped out dynamically.

### 2. The `data/npcs/` Definitions
- Create a new directory alongside `data/archetypes` called `data/npcs/`.
- While Archetypes describe a *Class* (e.g., "Knight"), the new `NPCConfig` describes a *Specific Entity*.
- Example `data/npcs/green_knight.yaml`:
  ```yaml
  id: green_knight
  name: Green Knight
  description: A legendary warrior bound to the forest.
  archetype: knight
  unique: true
  primary_color: "#00FF00"    # Swaps out Purple
  secondary_color: "#000000" # Swaps out Yellow
  health_min: 500
  speed: 1.5
  ```

### 3. Dynamic Palette Swapping logic
- Ebitengine allows custom shader programs (Kage) or Colorm functions for drawing loops.
- Update `internal/game/npc_render.go`. When `Draw()` is called, if the NPC has a custom `PrimaryColor`, loop through the pixels (or use a Kage shader for performance) to replace any `#FF00FF` pixels with the defined custom color before rendering to the screen.

### 4. UI: Names and Hover Descriptions
- **Names**: Draw the `NPC.Name` string directly beneath the NPC's feet (using `textRenderer.DebugPrintAt` or a custom font engine) if the entity has a name defined.
- **Hover**: In `game.go`, check the `mouseX` and `mouseY` (translated to isometric space). If the cursor bounding box overlaps an NPC's isometric footprint, draw a floating tooltip box containing the `NPC.Description`.

### 5. The "Unique" Constraint
- When map chunks spawn, they query the available NPCs. 
- Create a global `g.KilledUniqueEnemies` map `map[string]bool`.
- If a spawned enemy's config has `unique: true`, first check the dead map. If `g.KilledUniqueEnemies["green_knight"] == true`, abort the spawn or downgrade it to a standard archetype.
- This ensures named bosses never appear twice in a single run.

## Challenges
- **Palette Performance**: Doing pixel-by-pixel color replacement on the CPU in Go every frame will kill framerates. We MUST use Ebitengine's `*ebiten.Shader` (Kage) to perform the color swap on the GPU during the `DrawRectShader` call.
- **Mouse Picking in Isometric Space**: Translating raw screen X/Y to the correct isometric Z-indexed entity for the hover description requires precise reverse-matrix math.
