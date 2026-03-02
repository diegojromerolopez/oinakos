# Oinakos (Knight's Path) 🛡️🗡️

> **⚠️ Warning:** This project has been vibecoded.

An infinite, isometric action combat game built with **Go** and the **Ebiten v2** 2D game library.

## 📖 Game Overview

**Oinakos** puts you in the boots of a lone knight in a vast, procedurally generated world. Your goal is simple: survive as long as possible while defeating the waves of Non-Player Characters (NPCs) that seek you out.

### Key Gameplay Features:
-   **Infinite Procedural World**: Chunk-based generation creates an endless world of forests, ruins, and villages as you explore.
-   **Dynamic Ambush System**: NPCs spawn from the edges of your view and track you down using specific behavior profiles (wander, hunter, etc.).
-   **Menu & Save System**: 
    -   Press **ESC** to open the Game Menu (Resume, Quicksave, Load, Quit).
    -   **Native**: Saves to `.oinakos.yaml` files in the `saves/` directory. Use the "Load" dialog to select and resume any save.
    -   **WASM**: Automatic persistence to **localStorage**. The game automatically resumes your last session when you refresh the page.
-   **Unique NPC System**: Encounter unique "Boss" characters like **Stultus** or **Marcus Ardea** with custom descriptions, names, and gold-trimmed info boxes.
-   **Dynamic Palette Swapping**: Engine uses GPU shaders to dynamically change NPC colors (e.g., unit armbands or capes) using primary/secondary color masks.
-   **Testable Engine Architecture**: Game logic is decoupled from the rendering engine using strict Dependency Injection, enabling 100% headless testing.

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

## 📁 Technical Architecture
-   **`internal/engine`**: Platform-agnostic core (Iso transforms, Camera, Collision, Graphics interfaces).
-   **`internal/game`**: Game-specific logic: chunking, NPC AI, Persistence (Platform-split), and HUD.
-   **`assets/`**: High-quality pixel art and localized sound effects.

## 📜 License
This project is licensed under the MIT License. All AI-generated assets and code are intended for demonstration purposes.
