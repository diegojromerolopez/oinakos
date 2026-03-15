# Oinakos — Agent Memo 🛡️

Oinakos is a performance-optimized, infinite isometric action RPG built in Go. This memo is the **technical source of truth** for AI agents working on the codebase. Read this before touching any file.

---

## ⚙️ Technical Core

- **Engine**: Custom `internal/engine` wrapping [Ebiten v2](https://ebiten.org/).
- **Coordinate Systems** (two separate spaces — never mix them):
  - **Cartesian**: All physics, AI, collision, and game-logic coordinates live here.
  - **Isometric**: Used **only** for rendering. Transform: `isoX = (x - y)`, `isoY = (x + y) * 0.5`.
- **Simulation Rate**: Locked at **60 TPS** (`ebiten.SetTPS(60)`). All timers/cooldowns are in ticks.
- **TileSize**: `engine.TileWidth = 64px`, `engine.TileHeight = 32px`. Map dimensions in pixels are divided by these to get Cartesian units.
- **Versioning**: Initial version is **0.1**. Version is injected via `-ldflags "-X main.Version=$(VERSION)"`.

---

### Data-Driven Registries
All game content is defined in YAML under `data/` and loaded at startup:
- **`ArchetypeRegistry`** — shared stats, sprites, audio dir, and AI profile for a category of NPC (e.g. `orc/male`).
- **`NPCRegistry`** — unique named NPCs (e.g. Stultus, Virculus). They can override archetype stats and have their own audio folder.
- **`PlayableCharacterRegistry`** — player-selectable characters. Each sets `EntityConfig.PlayableCharacter = config.ID`.
- **`MapTypeRegistry`** — both standalone maps and individual campaign map levels.
- **`CampaignRegistry`** — ordered sequences of map IDs.

### Playable Characters
- Defined in `data/characters/`. Loaded by `PlayableCharacterRegistry`.
- The character the **player is currently controlling** is the **Playable Character**.
- `EntityConfig.PlayableCharacter` is set to `config.ID` for every entry in this registry (e.g. `"boris_stronesco"`).
- This field is the **canonical token** for all character-specific runtime logic:
  - Audio prefix: `PlayableCharacter + "/attack"` → plays from `assets/audio/characters/<id>/`
  - Future uses: HUD portrait, dialogue triggers, quest flags, unique mechanics.
- **Do not hardcode `"oinakos"` anywhere.** Always use `pc.Config.PlayableCharacter`.

### Y-Sorting (Z-Ordering)
- The renderer sorts all drawable entities by `Y + X` (Cartesian) before each draw call.
- This achieves correct depth occlusion without a Z-buffer.

### NPC Audio Fallback Chain
1. Check `assets/audio/npcs/<npc_id>/` for WAV files → use NPC-specific audio.
2. Else fall back to `assets/audio/archetypes/<archetype_id>/` (the archetype's voice).
3. Player character audio always uses `PlayableCharacter` as the key prefix.

---

## 📜 Coding Standards

### Go Best Practices
- **`gofmt`**: All code must be formatted with the standard Go formatter.
- **Explicit Error Handling**: Check every error immediately. Keep the "happy path" on the left.
- **Naming**: Use `CamelCase` for all names. Short names for local scope (e.g., `err`, `i`), descriptive for package/struct level.
- **Composition**: Use struct embedding over complex heirarchies.
- **`interface{}` vs `any`**: Do not use `interface{}`, use `any` instead.
- **Minimize `any`**: Do not use `any` unless absolutely necessary (e.g. library requirements, generic loading). Prefer specific interfaces.

### SOLID Principles & Dependency Injection
- **S.O.L.I.D.**: All new features and refactors must adhere to SOLID principles:
  - **Single Responsibility**: Each struct/package should have one clear purpose.
  - **Open/Closed**: Code should be open for extension but closed for modification (e.g., registries, interfaces).
  - **Liskov Substitution**: Interfaces (like `engine.Graphics`) must be fully replaceable by mocks in tests.
  - **Interface Segregation**: Prefer small, focused interfaces over large "God" interfaces.
  - **Dependency Inversion**: High-level game logic must depend on abstractions, not Ebiten directly.
- **Dependency Injection**: 
  - Never use global state for game logic. 
  - Pass dependencies through constructors or context structs (e.g., `SystemContext`). 
  - This is non-negotiable for testability.
  - **`internal/game` must never import `ebiten` directly**. Only `internal/engine` and `main.go` may.
  - All Ebiten types are behind interfaces (`engine.Graphics`, `engine.Input`, `engine.Image`). This enables **100% headless unit testing** of game logic.
- **Headless Testing**: 
  - All unit tests must be run with `-tags test` (e.g., `go test -tags test ./internal/...`). 
  - This build tag swaps Ebiten-dependent code in `internal/engine` for headless stubs in `internal/engine/test_stubs.go`.
  - Use `make test` as the canonical way to run tests.

### File Hygiene
- **Max File Length**: No file should exceed **500 lines**. If a file grows beyond this, refactor and split it by responsibility.
- **Package Names**: Name folders after the service they provide (e.g., `internal/game`, not `internal/utils`).

---

## 💾 Persistence System

- **Format**: YAML (`SaveData` struct). Extension: `.oinakos.yaml`.
- **Native**: Saves to `oinakos/saves/` beside the binary. Supports multiple named saves + load dialog.
- **WASM**: Persists to browser `localStorage` under key `quicksave`. Auto-resumes on page load.
- **Platform bridge**: `persistence_native.go` vs `persistence_js.go`, split via Go build tags.
- **Character identity** is stored as `player.archetype_id` in the save file and looked up in `PlayableCharacterRegistry` on load — `PlayableCharacter` is then set automatically from the registry.

---

## 🎨 Asset Generation Standards

### Sprites
- **Characters & NPCs**: **160×160 px**, solid **`#00FF00`** (chroma-key green) background.
- **Proportions**: Realistic human scale relative to the isometric tile.
- Required frames: `static.png`, `attack.png`, `corpse.png`. Optional: `back.png`, `hit.png`, `hit1.png`, `hit2.png`, `attack1.png`, `attack2.png`.

### Palette Masking (Shader-Swappable Colors)
- **Magenta (`#FF00FF`)**: Primary color zone — swapped at runtime with `primary_color` from YAML.
- **Yellow (`#FFFF00`)**: Secondary color zone — swapped with `secondary_color`.
- This is how faction armbands, cape colors, etc. are done without duplicate sprites.

### Collision Footprints
- Defined as `footprint: [{x, y}, ...]` polygon in the archetype/NPC/character YAML.
- **Always** use `make boundaries-editor` to visually define footprints. Do **not** hand-edit polygon coordinates blindly.
- `make boundaries-editor OBSTACLE=tree_oak`
- `make boundaries-editor NPC=stultus`
- `make boundaries-editor CHARACTER=oinakos`

### Audio
- Format: **WAV**, single-channel or stereo, any sample rate (engine resamples to 44100 Hz).
- Generated via [Piper TTS](https://github.com/rhasspy/piper). Scripts in `scripts/`. Models in `models/` (not committed).
- See [`assets/audio/README.md`](assets/audio/README.md) for the full voice-model registry.
- Standard sound files per entity: `hit.wav`, `death.wav`, `attack_1.wav` … `attack_5.wav`.

---

## 📁 Project Layout

```
/
├── cmd/                    # Additional binary entry points (if any)
├── internal/
│   ├── engine/             # Ebiten abstractions, isometric math, audio manager, shaders
│   └── game/               # Game loop, NPC AI, combat, HUD, save, registries, world gen
├── data/
│   ├── archetypes/         # <category>/<gender>.yaml → shared mob templates
│   ├── characters/         # <id>.yaml → playable character definitions
│   ├── npcs/               # <id>.yaml → unique/named NPCs
│   ├── obstacles/          # <id>.yaml → map object definitions
│   ├── maps/               # <id>.yaml → standalone sandbox maps
│   └── campaigns/          # <id>.yaml → ordered map sequences
│       └── <id>/           # Per-campaign map level YAMLs
├── assets/
│   ├── images/
│   │   ├── archetypes/     # Sprites per archetype category/gender
│   │   ├── characters/     # Sprites per playable character
│   │   ├── obstacles/      # Obstacle sprites
│   │   └── tiles/          # Floor tile textures
│   └── audio/
│       ├── archetypes/     # <category>/<id>/ → archetype voice lines
│       ├── npcs/           # <npc_id>/ → unique NPC voice overrides
│       └── characters/     # <character_id>/ → player character voice lines
├── tools/
│   ├── boundaries_editor/  # Footprint editor (Go + Ebiten)
│   ├── map_editor/         # Map authoring tool (Go + Ebiten)
│   └── asset_processor/    # Sprite preprocessing (Python, run via `uv`)
├── scripts/                # Audio gen, platform bundling scripts
├── models/                 # Piper TTS ONNX model files (gitignored)
├── bin/                    # Compiled development binaries
└── dist/                   # Production distribution packages
```

---

## 🛠 Makefile Commands

| Command | Description |
| :--- | :--- |
| `make run` | Build & run natively |
| `make run-debug` | Build & run with debug overlays |
| `make test` | Run all unit tests (headless) |
| `make build` | Compile native binary to `bin/` |
| `make dist` | Build minimal 2-file WASM package |
| `make serve-wasm` | Build WASM + serve on `localhost:8000` |
| `make map-editor` | Launch the graphical map editor |
| `make boundaries-editor OBSTACLE=id` | Launch footprint editor for an obstacle |
| `make boundaries-editor NPC=id` | Launch footprint editor for a unique NPC |
| `make boundaries-editor CHARACTER=id` | Launch footprint editor for a character |
| `make bundle-mac` | Build `dist/Oinakos.app` |
| `make bundle-windows` | Build `dist/Oinakos_Windows.zip` |
| `make bundle-linux` | Build `dist/Oinakos_Linux.tar.gz` |
| `make bundle-all` | Build all platform packages |
| `make clean` | Delete `bin/` and `dist/` |

---

## 📝 Pending Roadmap

- [ ] **Animation System**: Sprite-sheet support for walk / attack / death states.
- [ ] **A\* Navigation**: Replace linear NPC tracking with proper grid pathfinding.
- [ ] **Dynamic Biomes**: Procedural background changes based on chunk distance from origin.
- [ ] **UI Refresh**: Replace debug-print HUD with textured panels and portrait icons.
- [ ] **Occlusion Effect**: Greyscale silhouette for entities behind obstacles (plan in `plans/`).

---

**Default Lead Character**: `Oinakos` — any character in `data/characters/` can be selected; the active one is identified at runtime by `EntityConfig.PlayableCharacter`.

**Development Rule**: Always execute Python tools via `uv` in a virtual environment (`uv run …` or `.venv/bin/python`).
