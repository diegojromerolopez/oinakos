# Oinakos (Knight's Path)

An infinite, isometric action combat game built with Go and Ebiten.

## 🛠 Tech Stack
-   **Language**: Go 1.21+
-   **Graphics**: [Ebiten v2](https://ebiten.org/) (2D Game Library)
-   **Assets**: Custom pixel art (isometric) and localized SFX.

## 🚀 Key Features
-   **Infinite Procedural World**: Chunk-based generation (10x10 tiles) creates an endless world of forests, ruins, and villages as you explore.
-   **Isometric Engine**: custom `internal/engine` handling Cartesian-to-Isometric transforms, camera following, and polygon-based collision detection. No scaling allowed at runtime (all assets must be pre-scaled).
-   **Dynamic Ambushes & Elites**: NPCs (Orcs, Demons, Peasants) spawn beyond the viewport. A 5% chance triggers "Elite" variants using unique color palettes and boosted stats.
-   **Unique NPCs & Bosses**: Unique characters like *Marcus Ardea* or *Stultus* feature unique descriptions, boss bars, and names.
-   **Depth-Correct Rendering**: implemented Y-sorting (Z-ordering) ensures correct occlusion between players, NPCs, and buildings.
-   **Combat System**: Hit-detection with generous isometric radii and persistent kill tracking. Supporting both melee and ranged (shouting) attacks.
-   **Dynamic Palette Swapping**: GPU-based shaders allow NPCs to share archetypes while maintaining unique identity via primary/secondary color masking.

## 📁 Project Structure
-   `/assets`: Organized by `images/` (player, npcs, environment) and `audio/`.
-   `/internal/engine`: Platform-agnostic game engine logic (Iso, Camera, Collision, Renderer).
-   `/internal/game`: Game-specific state, entities (NPC, Player, Obstacle), and level generation.

## 🏗 Development Patterns
-   **Dependency Injection**: Code in `internal/game` must **not** import `github.com/hajimehoshi/ebiten/v2` directly. Instead, use the interfaces defined in `internal/engine` (e.g., `Graphics`, `Image`, `Input`). This ensures UI components can be mocked, enabling 100% headless unit testing.
-   **Repository Pattern**: Use registries (e.g., `ArchetypeRegistry`, `MapTypeRegistry`) to manage configuration data, keeping data schemas decoupled from game simulation logic.
-   **Data-Driven Design (YAML)**: Avoid hardcoding entity or map parameters. Stats, sprite paths, and behaviors are loaded from YAML configurations.
-   **Asset Generation Rules**: Archetype and character images *must* have a solid lime green background (`#00FF00`).
    -   **Style**: Realistic human proportions and dark, medieval RPG aesthetic (Diablo/Hades).
    -   **Corpse Rules**: Just the body without any ground base or platform.
    -   **Palette Masking**: Use **Magenta (#FF00FF)** for primary color areas (e.g., armbands, capes) and **Yellow (#FFFF00)** for secondary areas. Shaders replace these at runtime.
    -   **Sizing**: No runtime scaling allowed; all assets must be 160x160 for characters.
-   **Python Tools**: Any Python scripts located in the `tools/` directory must be executed within a virtual environment. You must use `uv` for dependency management with Python 3.14.
    -   `tools/asset_processor`: Essential for preparing AI-generated images (BG removal, hex-snapping).

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
