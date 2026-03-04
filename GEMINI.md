# Oinakos (Knight's Path) - Agent Memo 🛡️

Oinakos is a performance-optimized, infinite isometric action RPG built in Go. This memo serves as a technical source of truth for AI agents working on the codebase.

## ⚙️ Technical Core
- **Engine**: Custom `internal/engine` using [Ebiten v2](https://ebiten.org/).
- **Coordinate Systems**:
  - **Cartesian**: Used for all physics, AI, and collision logic.
  - **Isometric**: Used strictly for rendering. Transform: `isoX = (x - y)`, `isoY = (x + y) * 0.5`.
- **Concurrency**: Simulation runs at a locked 60 Ticks Per Second (TPS).

## 🚀 Architectural Pillars
- **Strict Dependency Injection**: Components in `internal/game` must use `engine` interfaces (`Graphics`, `Input`, `Audio`). Never import `ebiten` directly outside of `internal/engine` or `main.go`. This enables 100% headless unit testing.
- **Data-Driven Registry**: Entities (NPCs, Obstacles) and Maps are defined via YAML. The `ArchetypeRegistry` is the authority on shared stats and assets. Unique NPCs have specific behaviors, names, and a group alignment system to trigger faction-wide alerts.
- **Y-Sorting (Z-Ordering)**: The renderer maintains depth-correct occlusion by sorting all drawable entities by their `Y+X` cartesian coordinate before drawing.

## 💾 Persistence System
- **Format**: YAML-based serialization (`SaveData` struct). File extension: `.oinakos.yaml`.
- **Native Implementation**: Saves to the `saves/` directory (ignored by Git).
- **WASM Implementation**: Saves to browser `localStorage` under the key `quicksave`. Feature includes auto-resumption on page load.
- **Platform Bridge**: Split via build tags in `persistence_js.go` and `persistence_native.go`.

## 🎨 Asset Generation Standards
- **Characters**: 160x160 pixels, solid `#00FF00` background. Proportions: Realistic human.
- **Palette Masking**:
  - **Magenta (#FF00FF)**: Primary color area (Shader-swappable).
  - **Yellow (#FFFF00)**: Secondary color area (Shader-swappable).
- **Collision Footprints**: Defined as polygons in the archetype YAML. Must be refined using `make boundaries-editor`.

## 📁 Project Layout
- `/bin`: Native development binaries (`oinakos`, `boundaries_editor`, `map_editor`).
- `/dist`: Distribution packages (WASM, Minimal `index.html`, Native Bundles).
- `/data`: Game content definitions (YAML archetypes, maps, unique NPCs).
- `/assets`: Raw and processed game sprites and sound effects.
- `/internal/engine`: Low-level Ebiten abstractions and isometric math.
- `/internal/game`: Game loop, entity logic, and HUD rendering.
- `/scripts`: Platform-specific bundling and audio generation scripts.

## 🛠 Makefile Commands (Standard Workflow)
- `make run`: Compile and run natively.
- `make dist`: Generate minimal WASM package (Inlined JS, 2 files total).
- `make bundle-all`: Build standalone installers for macOS, Windows, and Linux.
- `make map-editor`: Launch the graphical map editor tool.
- `make boundaries-editor OBSTACLE=tree_oak`: Launch footprint editor.
- `make clean`: Purge `bin/` and `dist/` folders.

## 📝 Pending Roadmap
- [ ] **Animation System**: Implement sprite-sheet support for walk/atk/death states.
- [ ] **A* Navigation**: Move beyond linear lerp-tracking for NPCs (currently relies on simple collision avoidance).
- [ ] **Dynamic Biomes**: Procedural background changes based on chunk distance.
- [ ] **UI Refresh**: Replace debug-print HUD with textured elements and icons.

---
**Current Lead Character**: `Oinakos`
**Development Rule**: Always execute Python tools via `uv` in a virtual environment.
