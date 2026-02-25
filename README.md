# Oinakos (Knight's Path) 🛡️🗡️

> **⚠️ Warning:** This project has been vibecoded.

An infinite, isometric action combat game built with **Go** and the **Ebiten v2** 2D game library.

## 📖 Game Overview

**Oinakos** puts you in the boots of a lone knight in a vast, procedurally generated world. Your goal is simple: survive as long as possible while defeating the waves of Non-Player Characters (NPCs) that seek you out.

### Key Gameplay Features:
-   **Infinite Procedural World**: Chunk-based generation creates an endless world of forests, ruins, and stone walls as you explore.
-   **Dynamic Ambush System**: NPCs spawn from the edges of your view and track you down using specific behavior profiles (wander, hunter, etc.).
-   **Combat & Progression**:
    -   Attack NPCs with a precise hitbox system.
    -   Collect **XP** and track your **Kills** to measure your progress.
    -   Face diverse NPC types: **Orcs**, **Demons**, **Peasants**, **Goblins**, **Lame Devils**, and **Magi**, each with unique stats, weapons, and behaviors.
-   **Configurable Scenarios**: Maps and encounters are fully data-driven via YAML configuration files, allowing for easy content expansion.
-   **Testable Engine Architecture**: Game logic is decoupled from the rendering engine using strict Dependency Injection, enabling 100% headless testing.
-   **Survival Timer**: Every second counts. Your total survival time is tracked in real-time.
-   **High Fidelity Isometric Rendering**: Custom cartesian-to-isometric math handles depth-sorting (Y-sorting), precise footprints, and polygon-based collisions.
-   **Procedural Animations**: Characters feature dynamic directional mirroring, walking bob, and attack lunges calculated in real-time without relying on complex sprite sheets.
-   **Embedded Assets**: All sprites and audio are baked directly into the binary, making the game fully portable across desktop and web (WASM) environments.

## 🚀 Getting Started

### 🤖 AI Generation Disclaimer

**Please note:** This project has been heavily developed leveraging AI. Not only is the source code extensively written and refactored by AI, but **all** of the graphical assets (sprites, environments, items, UI elements) and sound effects have been created using AI generation models.

## Prerequisites
-   **Go 1.26+**
-   A system with GPU support (required by Ebiten for native desktop builds)
-   **Docker** (optional, for running the web server)

### 1. Play Natively (Desktop)
The easiest way to play on your desktop, with all assets embedded:
```bash
make run
```
To just build a standalone executable to the `bin/` directory:
```bash
make build
./bin/oinakos
```

### 2. Play in Browser (WebAssembly)
Oinakos can be compiled to WASM and runs entirely in your browser. We provide a Make target that builds the WASM binary and serves it locally.
```bash
make serve-wasm
# Then visit http://localhost:8000 in your browser
```
If you only want to build the files into `bin/` without serving:
```bash
make build-wasm
```

### 3. Play via Docker (Self-Contained Web Server)
We provide a multi-stage `wasm.Dockerfile` that builds the WASM game and serves it automatically using a minimal Go static server.
```bash
docker build -t oinakos-wasm -f wasm.Dockerfile .
docker run -p 8081:8000 oinakos-wasm
# Then visit http://localhost:8081 in your browser
```

## 🎮 Controls:
-   **WASD / Arrow Keys**: Move the Knight
-   **SPACE**: Attack
-   **ESC**: Pause / Exit confirmation
-   **ENTER**: Restart (on Game Over screen)

## 📁 Technical Architecture
-   **`internal/engine`**: The platform-agnostic core handling isometric transforms, camera lerping, and polygon-based collision detection.
-   **`internal/game`**: Game-specific logic: chunk generation, NPC AI, player state, and rendering tasks.
-   **`assets/`**: High-quality pixel art and localized sound effects.

## 📝 Future Improvements
-   [ ] **Full Animation System**: Transition from static sprites to multi-frame animations for walking, attacking, and death.
-   [ ] **A* Pathfinding**: Upgrade NPC AI to intelligently navigate around walls and buildings.
-   [ ] **Biome Diversity**: Add new terrain types like Snow, Desert, and Swamp with unique obstacles.
-   [ ] **Skill Tree**: Implement a level-up system where XP can be spent on new abilities or stat boosts.
-   [ ] **Equipment System**: Add loot drops and equipable items (swords, shields, armor).
-   [ ] **Persistent Save Data**: Locally save your high scores and character progress in JSON format.
-   [ ] **UI/UX Overhaul**: Replace debug text HUD with custom-textured bars, icons, and menus.

## 📜 License
This project is licensed under the MIT License. All AI-generated assets and code are intended for demonstration and educational purposes.
