# Plan: Fix Obstacle Backgrounds

## Objective
Remove the "horrible" transparent lines or incorrect backgrounds present in some environmental obstacle sprites (trees, rocks, houses, etc.) so they render seamlessly on the isometric grid.

## Analysis
Currently, some environment sprites (e.g., `assets/images/environment/*.png`) have visual artifacts. This happens because pixel art assets often require exact masking, and any anti-aliased or semi-transparent edge pixels can show up as hard lines or glitches when drawn over a background in a game engine like Ebiten.
Additionally, the `engine.Transparentize` function used currently attempts color-keyed transparency (removing pure white/magenta/black backgrounds), which leaves jagged edges if the source image had anti-aliasing.

## Implementation Steps

### 1. Identify Problematic Sprites
- Run the game and visually identify exactly which files have the transparent lines.
  - Likely candidates: `tree1.png`, `house1.png`, `rock1.png`, etc.

### 2. Manual/Automated Sprite Cleanup
- Since the AI image generator struggles with precise pixel-perfect slicing and alpha mask generation without altering the core image, the best approach is to edit the raw PNG files to have true alpha transparency (Alpha channel) instead of relying on `engine.Transparentize`.
- **Action**: Use a script (or manual editing if the user prefers) to bulk-convert the background color (e.g., pure white `rgb(255,255,255)`) to transparent alpha `rgba(0,0,0,0)` using a strict threshold, ensuring no semi-transparent fringe pixels remain.

### 3. Engine Update
- Modify `loadSprite` in `internal/game/game.go`.
- If we convert the images to true PNG alpha, we can completely remove the `removeBg` parameter and the call to `engine.Transparentize(img)`, as relying on Ebiten's native PNG alpha rendering is significantly cleaner and faster.

## Verification
- Load into the game and walk around dense forests and villages.
- Verify that trees and houses overlap cleanly without geometric lines or "halos" around their borders.
