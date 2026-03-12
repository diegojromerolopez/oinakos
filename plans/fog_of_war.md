# Fog of War Implementation Plan 🌫️

Implement a controllable "Fog of War" system for Oinakos to enhance navigation and tension.

## ⚙️ Logic & Terminology

The following modes will be available in the game settings:
- **Off (`none`)**: Standard visibility. The entire map is visible.
- **Vision (`vision`)**: Represents a "Torchlight" effect. Only the area around the playable character is visible. Everything else remains pitch black.
- **Exploration (`exploration`)**: Traditional Fog of War. Areas revealed by the player's vision stay visible forever (unshrouded landscape), but only entities currently in the vision radius are updated visually (though standard view will probably show the whole revealed landscape).

---

## 🛠 Architectural Changes

### 1. Settings Persistence
- Update `internal/game/settings.go`:
    - Add `FogOfWar` field to the `Settings` struct.
    - Default to `off`.
    - Possible values: `off`, `vision`, `exploration`.
- Ensure it is marshaled/unmarshaled correctly in `settings.yml`.

### 2. Game State
- Update `internal/game/game.go`:
    - Add `ExploredArea map[image.Point]bool` to the `Game` struct.
    - Each `image.Point` represents a Cartesian grid cell (e.g., 1x1 units) that has been "seen".
    - `Update` Loop:
        - If `FogOfWar` != `off`, calculate points within `radius` (e.g., 18 units) of `playableCharacter`.
        - Mark these points as "seen" in `ExploredArea`.
    - Exploration state should be part of the `SaveData` struct in `internal/game/save.go` to persist across sessions (per map).

### 3. Mask-Based Rendering
Instead of per-tile checks, a GPU-based mask approach is recommended for performance and visual fidelity (smooth edges):
- **Generate Mask**:
    - Every frame, create a low-res alpha mask (e.g., 1/4 screen size).
    - Clear to black.
    - Draw a white fuzzy circle (gradient) at the player's screen position.
    - If `exploration` mode is active, also draw white dots/circles for all `ExploredArea` points.
- **Apply Mask**:
    - Draw the game to a "Game Surface".
    - Draw the mask to a "Fog Surface".
    - Composite the Fog Surface over the Game Surface using `CompositeModeDestinationIn` (or similar) to "cut through" the darkness.
    - Alternatively, just draw the "Black Fog" image with transparency cutouts over the rendered world.

---

## 🚀 Step-by-Step Implementation

1.  **Phase 1: Setting Management (Now)**
    - Add `FogOfWar` string to `Settings`.
    - Update `GetOinakosDir` and `LoadSettings` if needed.
    - Add `FogOfWar` toggle to the `Settings` screen UI.
    - Save/Load from YAML.

2.  **Phase 2: Exploration Persistence**
    - Add `ExploredTiles` to `SaveData`.
    - Implement the bitset/map for tracking visited cells.

3.  **Phase 3: Shader & Masking (Next)**
    - Implement a "Vignette" / "Torch" shader or mask in `internal/engine` or `internal/game/game_render.go`.
    - Handle the `exploration` mode blending (old explored areas vs current vision).

---

## 🎨 Visual Aesthetics
- **Vision Mode**: Harsh black edges or a gentle gradient depending on the "Torch" feel.
- **Exploration Mode**: Explored but out-of-vision areas could be slightly desaturated or dimmed (standard RPG look) or fully visible as per the user's "it stays viewable" request.

