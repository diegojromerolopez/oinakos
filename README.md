# Oinakos (Knight's Path) 🛡️🗡️

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)
![Ebiten v2](https://img.shields.io/badge/Engine-Ebiten%20v2-orange)

An infinite, isometric action combat game built with **Go** and the **Ebiten v2** 2D game library.

## 📖 Game Overview

**Oinakos** puts you in the boots of a lone knight in a vast, procedurally generated world. Your goal is simple: survive as long as possible while defeating the waves of Non-Player Characters (NPCs) that seek you out.

### Key Gameplay Features:
-   **Infinite Procedural World**: Chunk-based generation creates an endless world of forests, ruins, and villages as you explore.
-   **Dynamic Ambush System**: NPCs spawn from the edges of your view and track you down using specific behavior profiles (wander, hunter, etc.).
-   **Menu & Save System**: 
    -   Press **ESC** to open the Game Menu (Resume, Quicksave, Load, Quit).
    -   **Native**: Saves to `.oinakos.yaml` files in the `oinakos/saves/` directory. Use the "Load" dialog to select and resume any save.
    -   **WASM**: Automatic persistence to **localStorage**. The game automatically resumes your last session when you refresh the page.
-   **Unique NPC System**: Encounter unique "Boss" characters like **Stultus** or **Marcus Ardea** with custom descriptions, names, and gold-trimmed info boxes.
-   **NPC Group Alignment**: NPCs can belong to specific groups (e.g., "Peasants", "Crimson Arm"). Attacking any member of a group will alert all nearby members of the same faction, turning them hostile simultaneously.
-   **Dynamic Palette Swapping**: Engine uses GPU shaders to dynamically change NPC colors (e.g., unit armbands or capes) using primary/secondary color masks.
-   **Testable Engine Architecture**: Game logic is decoupled from the rendering engine using strict Dependency Injection, enabling 100% (headless) unit testing coverage.

### 🎭 NPC Factions & Groups
Oinakos implements a localized "Social AI" system. NPCs with a defined `group` field (e.g., `group: Peasants`) share a hive-mind response within a **20-unit radius**. 
-   **Radial Alerts**: When an NPC is damaged by the player, it broadcasts an alert.
-   **Shared Alignment**: All nearby members of the same group instantly switch their alignment to **Enemy** and initiate combat.

## 🚀 Getting Started

### Prerequisites
-   **Go 1.21+**
-   A system with GPU support (required by Ebiten for native desktop builds)
-   **Python 3.14** (for asset tools, managed via `uv`)

### 1. Play Natively (Desktop)
The easiest way to play on your desktop:
```bash
make run
```
To build a standalone executable to the `bin/` directory:
```bash
make build
./bin/oinakos
```

### 2. Play in Browser (WebAssembly)
Oinakos can be compiled to WASM and runs entirely in your browser with automatic state persistence.
```bash
make serve-wasm
# Then visit http://localhost:8000 in your browser
```
The WASM distribution is optimized into just **two files** (`index.html` and `oinakos.wasm`) located in the `dist/` folder.

## 📦 Deployment & Distribution

Oinakos features a robust multi-platform bundling system that generates production-ready, standalone packages. All generated bundles are placed in the `dist/` directory.

### Build Targets:
- **macOS Bundle**: `make bundle-mac`
  - Generates `dist/Oinakos.app`. A native, double-clickable application with Retina icons.
- **Windows Package**: `make bundle-windows`
  - Generates `dist/Oinakos_Windows.zip`. Standalone `.exe` built in GUI mode.
- **Linux Package**: `make bundle-linux`
  - Generates `dist/Oinakos_Linux.tar.gz`. Cross-compiled via Docker. 
- **All Platforms**: `make bundle-all`

## 🎮 Controls:
-   **WASD / Arrow Keys**: Move Oinakos
-   **SPACE**: Attack
-   **Q**: Quick Quicksave
-   **ESC**: Open Game Menu
-   **TAB**: Toggle Entity Boundaries (Debug View)
-   **ENTER**: Select (Menu) / Restart (on Game Over)

## 🛠️ Development Tools

### Boundaries Editor
Graphical tool for editing collision footprints in isometric space.
```bash
make boundaries-editor
```
Changes are saved automatically to the corresponding `.yaml` file in `data/`.

### Asset Processor
Utility in `tools/asset_processor` to prepare sprites (BG removal, hex-snapping).
```bash
uv run tools/asset_processor/main.py input.png output.png
```

## 🛠️ Modding & Custom Content

Oinakos supports easy modding and content overrides without needing to recompile the game. To override any game asset or data file, create an **`oinakos/`** folder in the same directory as the game executable.

The game priorities loading from the local folder in this order:
1.  **Local Folder**: `oinakos/data/` or `oinakos/assets/`
2.  **Internal Assets**: Files embedded within the binary.

### Local Mod Directory Tree

For your overrides, follow this exact structure inside the `oinakos/` folder:

```text
oinakos/
├── data/
│   ├── archetypes/          # Shared unit templates (Stats, AI profiles)
│   │   └── <category>/      # e.g., human/, orc/
│   │       └── <id>.yaml    
│   ├── npcs/                # Unique/Named NPCs (Names, unique colors)
│   │   └── <id>.yaml
│   ├── obstacles/           # Map objects (Trees, buildings, rocks)
│   │   └── <id>.yaml
│   ├── maps/                # Custom map levels (Created via map editor)
│   │   └── <id>.yaml
│   └── characters/
│       └── main/
│           └── character.yaml
├── assets/
│   ├── images/
│   │   ├── archetypes/
│   │   │   └── <category>/<id>/
│   │   │       ├── static.png   # Default idle frame
│   │   │       ├── attack.png   # Frame displayed during attack
│   │   │       ├── corpse.png   # Displayed when dead
│   │   │       ├── hit.png      # Standard flinch image
│   │   │       ├── hit1.png     # (Optional) Random flinch variant 1
│   │   │       └── hit2.png     # (Optional) Random flinch variant 2
│   │   ├── obstacles/
│   │   │   └── <id>.png        # Spritesheet or single frame for obstacles
│   │   └── characters/
│   │       └── main/           # Assets for the Knight (Oinakos)
│   │           ├── static.png
│   │           ├── attack.png
│   │           └── corpse.png
│   └── audio/
│       └── archetypes/
│           └── <category>/<id>/
│               ├── hit.wav      # Sound when damaged
│               └── death.wav    # Sound when killed
└── saves/                   # Save game files (*.oinakos.yaml)
```

### Tips for Modders:
- **Nested Folders**: If your data YAML is in `data/archetypes/human/male/peasant.yaml`, its images MUST be in `assets/images/archetypes/human/male/peasant/`.
- **Palette Masking**: The engine looks for Magenta (#FF00FF) and Yellow (#FFFF00) in archetype sprites to apply the primary and secondary colors defined in the NPC/Archetype YAML.
- **WAV Format**: All local audio must be provided in WAV format.

### Example: Adding a Halfling Archetype

To add a new "Halfling" unit in a "humanoid" category:

1.  **Define the Data**: Create `oinakos/data/archetypes/humanoid/halfling.yaml`:
    ```yaml
    id: halfling
    name: "Stout Halfling"
    behavior: "wander"
    stats:
      health_min: 20
      health_max: 30
      speed: 0.12
      base_attack: 4
      base_defense: 2
      attack_cooldown: 50
      attack_range: 0.8
    weapon: "Dagger"
    description: "Small but deceptively quick."
    ```

2.  **Add the Visuals**: Place your sprites in the matching asset subfolder:
    -   `oinakos/assets/images/archetypes/humanoid/halfling/static.png`
    -   `oinakos/assets/images/archetypes/humanoid/halfling/attack.png`
    -   `oinakos/assets/images/archetypes/humanoid/halfling/corpse.png`

3.  **Run the Game**: The halfling will now be available for use in the map editor or for spawning via new NPC definitions!

### 🌍 Interactive Environments (Actions System)

Obstacles can perform multiple effects through the `actions` system. This allows objects like a "Magic Campfire" to both heal allies and harm enemies simultaneously.

All periodic effects pulse exactly every **1 second (60 ticks)**.

- **Example: Damaging Aura (Campfire)**
  ```yaml
  actions:
    - type: harm
      amount: 2     # Damage amount
      aura: 1.5    # Radius in world units. If 0, uses collision.
  ```

- **Example: Manual Interaction (Healing Well)**
  ```yaml
  actions:
    - type: heal
      amount: 10    # Restoration amount
      requires_interaction: true # Require SPACE key
      alignment_limit: "ally"    # "all", "ally", "enemy"
  ```


## 📁 Project Structure

-   **`/bin`**: Compiled development binaries and tools.
-   **`/dist`**: Distribution packages (WASM, macOS `.app`, etc.).
-   **`/internal/engine`**: Low-level math, rendering interfaces, and platform abstractions.
-   **`/internal/game`**: High-level game logic: NPC AI, combat, and world generation.
-   **`/data`**: Game content definitions (YAML).
-   **`/assets`**: Raw and processed game sprites and sound effects.

## 🤝 Contributing

We welcome contributions! Whether you're fixing bugs, adding new entities, or expanding the engine, here's how to get started:

1.  **Read the Technical Guide**: Before writing code, please review our [Agent Memo / Technical Core (`GEMINI.md`)](GEMINI.md). It outlines crucial architectural pillars, such as our **Strict Dependency Injection** rule (keeping game logic completely decoupled from `ebiten`), our data-driven registry, and coordinate systems.
2.  **Test-Driven**: Because of our DI architecture, the `internal/game` package can be tested entirely headlessly. Run `make test` to ensure all tests pass before submitting changes.
3.  **Use the Tools**: Use `make map-editor` and `make boundaries-editor` for visual content adjustments to avoid manual YAML errors.

## 📝 Roadmap

- [ ] **Animation System**: Move from static sprites to full sprite-sheet support.
- [ ] **A* Pathfinding**: Implement advanced navigation for complex environments.
- [ ] **Dynamic Biomes**: Procedural background and ambient weather changes.
- [ ] **UI Overhaul**: Replace debug-print elements with textured HUD components.

## 📜 License
This project is licensed under the MIT License. All code and images have been generated leveraging AI tools, specifically designed for high-performance and modularity.
