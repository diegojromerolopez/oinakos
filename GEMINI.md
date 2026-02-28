# Oinakos (Knight's Path)

An infinite, isometric action combat game built with Go and Ebiten.

## 🛠 Tech Stack
-   **Language**: Go 1.21+
-   **Graphics**: [Ebiten v2](https://ebiten.org/) (2D Game Library)
-   **Assets**: Custom pixel art (isometric) and localized SFX.

## 🚀 Key Features
-   **Infinite Procedural World**: Chunk-based generation (10x10 tiles) creates an endless world of forests, ruins, and villages as you explore.
-   **Isometric Engine**: custom `internal/engine` handling Cartesian-to-Isometric transforms, camera following, and polygon-based collision detection. No scaling allowed at runtime (all assets must be pre-scaled).
-   **Dynamic Ambushes**: NPCs (Orcs, Demons, Peasants) spawn beyond the viewport and track the player, creating a persistent combat loop.
-   **Depth-Correct Rendering**: implemented Y-sorting (Z-ordering) ensures correct occlusion between players, NPCs, and buildings.
-   **Combat System**: Hit-detection with generous isometric radii and persistent kill tracking displayed in the HUD and Game Over screen.

## 📁 Project Structure
-   `/assets`: Organized by `images/` (player, npcs, environment) and `audio/`.
-   `/internal/engine`: Platform-agnostic game engine logic (Iso, Camera, Collision, Renderer).
-   `/internal/game`: Game-specific state, entities (NPC, Player, Obstacle), and level generation.

## 🏗 Development Patterns
-   **Dependency Injection**: Code in `internal/game` must **not** import `github.com/hajimehoshi/ebiten/v2` directly. Instead, use the interfaces defined in `internal/engine` (e.g., `Graphics`, `Image`, `Input`). This ensures UI components can be mocked, enabling 100% headless unit testing.
-   **Repository Pattern**: Use registries (e.g., `ArchetypeRegistry`, `MapTypeRegistry`) to manage configuration data, keeping data schemas decoupled from game simulation logic.
-   **Data-Driven Design (YAML)**: Avoid hardcoding entity or map parameters. Stats, sprite paths, and behaviors are loaded from YAML configurations.
-   **Asset Generation Rules**: Archetype and character images *must* have a solid lime green background (`#00FF00`). They must also feature realistic human proportions and follow a dark, medieval RPG aesthetic similar to Diablo, Baldur's Gate, or Hades. Corpse images must specifically be just the body without any ground base or platform. **No runtime scaling is allowed; all assets must be provided at the exact pixel size required for the game. The engine automatically replaces solid lime green (#00FF00) with transparency at load-time.**
-   **Python Tools**: Any Python scripts located in the `tools/` directory must be executed within a virtual environment. You must use `uv` for dependency management with Python 3.14. Define all dependencies inside the `pyproject.toml` file. Never run Python tools globally.

## 📝 Pending Improvements
-   [ ] **Animation System**: Implement sprite-sheet animation for walking and attacking (currently using static frames).
-   [ ] **Advanced AI**: Replace simple lerp-tracking with A* pathfinding to handle obstacle navigation.
-   [ ] **Biome Diversity**: Add new terrain types (Snow, Desert, Swamp) with unique obstacles and NPCs.
-   [ ] **UI Overhaul**: Replace debug-print HUD elements with custom textured health bars and menus.
-   [ ] **Savestate**: Implement JSON-based save/load for player position and kill count.

## 🎮 Running the Game
```bash
go run main.go
```
