# Oinakos 🛡️🗡️

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)
![Ebiten v2](https://img.shields.io/badge/Engine-Ebiten%20v2-orange)

An infinite, isometric action combat game built with **Go** and the **Ebiten v2** 2D game library.

> *Feel like a kid again — loading up an RPG from a CD-ROM in late 1996.*

## 📖 Game Overview

**Oinakos** puts you in the boots of a lone knight in a vast, procedurally generated world. Survive against relentless waves of enemies, explore an infinite landscape, and carve your own path through a medieval world steeped in Spanish Ballad lore.

### 🎭 Playable Characters
Choose your knight before entering battle. Each character has unique stats, weapons, and voice lines in their native tongue:

| Character | Nationality | Weapon | Flavour |
| :--- | :--- | :--- | :--- |
| **Oinakos** | Unknown | Tizón | The lead hero. A mysterious knight forged in iron and ancient duty. |
| **Boris Stronesco** | Serbian | Claws | A noble who arrived in Cartagena carrying a cursed coffin. Speaks Serbian. |
| **Roland** | French | Durandal | A disciplined paladin of the Carolingian court. Speaks French. |
| **Conde Olinos** | Spanish | Long Sword | A tragic noble whose song moved the sea. Speaks Spanish. |
| **Durandarte** | Spanish | Sword | A knight whose heart was literally cut from his chest. Speaks Spanish. |
| **Gaiferos** | Spanish | Axe | A brave knight on a quest to free his beloved Melisendra. Speaks Spanish. |
| **Montesinos** | Spanish | Sword | Keeper of the deep cave and ancient knightly secrets. Speaks Spanish. |
| **Conde Estruch** | Catalan | Sword | A noble entangled in the mysteries of the old Catalan lands. |

### ⚔️ Key Gameplay Features
- **Campaigns & Maps**: Structured multi-map campaigns (*The Chronicles*, *Orc Invasion*, *Demonic Incursion*, *Kalot Embolot*) or freeform sandbox maps.
- **Infinite Procedural World**: Chunk-based generation creates an endless world of forests, ruins, and villages.
- **Dynamic Ambush System**: NPCs spawn from the edges of your view with distinct AI profiles (wander, hunter, patrol, chaotic, fighter, escort).
- **Unique Boss NPCs**: Encounter named enemies like **Stultus** (a ranged shouter), **Marcus Ardea**, **Virculus** (a clockwork automaton), **Tragantia** (a reptile-woman), **Lieutenant Varrick**, and **Staro Rovinec** — each with custom descriptions, stats, and voice lines.
- **Save System**: Full quicksave/load support. Native saves to `.oinakos.yaml` files; WASM auto-saves to browser `localStorage`.
- **NPC Faction System**: NPCs grouped by faction (e.g., *Peasants*, *Crimson Arm*) share hive-mind alert responses within a 20-unit radius.
- **Dynamic Palette Swapping**: GPU shaders recolor NPC unit armbands/capes at runtime using Magenta/Yellow color masks defined in YAML.
- **Interactive Environments**: Obstacles have an `actions` system — healing wells, damaging campfires — each with optional interaction gates.

### ⚔️ Combat & RPG Mechanics

Oinakos features a logarithmic RPG progression system:

- **Experience & Leveling**: Earn XP per kill (defined per archetype). `Level = (XP / 100) + 1`. Leveling up fully restores health.
- **RPG Attributes**: Health, Attack, Defense, Speed — all defined per character and scaled logarithmically.
- **Logarithmic Scaling**:
  $$\text{Stat} = \text{Base} + (\log_2(\text{Level}) \times 10)$$
  This keeps early levels exciting and high levels challenging without stat inflation.
- **Hit & Miss System**: Attack rolls are ratio-based: `hitChance = attack / (attack + defense) * 100`, clamped to [5, 95].
- **Weapon Quality**: Starting weapon gets a random bonus roll (0–3) for variety in each run.

### 🛡️ NPC Archetypes
Eleven archetypes form the spine of the enemy roster, most with distinct male and female variants:

`Demon` · `Giant` · `Goblin` · `Lame Devil` · `Magi` · `Man-at-Arms` · `Mythical` · `Orc` · `Peasant` · `Slave` · `Trasgo`

---

## 🚀 Getting Started

### Prerequisites
- **Go 1.21+**
- A system with GPU support (required by Ebiten for native desktop builds)
- **Python 3.14** (for asset and audio tools, managed via `uv`)

### 1. Play Natively (Desktop)
```bash
make run
```
To build a standalone executable to the `bin/` directory:
```bash
make build
./bin/oinakos
```

### 2. Debug Mode
```bash
make run-debug
# Enables collision boundary overlays (TAB key also toggles in-game)
```

### 3. Play in Browser (WebAssembly)
```bash
make serve-wasm
# Then visit http://localhost:8000
```
The WASM distribution is optimised into just **two files** (`index.html` and `oinakos.wasm`) in `dist/`.

---

## 🎮 Controls

| Key | Action |
| :--- | :--- |
| **WASD / Arrow Keys** | Move character |
| **SPACE** | Attack |
| **Q** | Quicksave |
| **ESC** | Open game menu (Resume / Load / Quit) |
| **ENTER** | Confirm selection / Restart on game over |
| **TAB** | Toggle collision boundary debug view |
| **Mouse** | Navigate menus |

---

## 📦 Deployment & Distribution

| Target | Command | Output |
| :--- | :--- | :--- |
| macOS App Bundle | `make bundle-mac` | `dist/Oinakos.app` |
| Windows Package | `make bundle-windows` | `dist/Oinakos_Windows.zip` |
| Linux Package | `make bundle-linux` | `dist/Oinakos_Linux.tar.gz` |
| All Platforms | `make bundle-all` | All of the above |
| WASM Only | `make dist` | `dist/index.html` + `dist/oinakos.wasm` |

---

## 🛠️ Development Tools

### Map Editor
Visual tool for creating and editing map levels with obstacles and inhabitants.
```bash
make map-editor
```

### Boundaries Editor
Graphical tool for editing collision footprints in isometric space.
```bash
make boundaries-editor OBSTACLE=tree_oak
make boundaries-editor NPC=stultus
make boundaries-editor CHARACTER=oinakos
```
Changes are saved automatically to the corresponding `.yaml` file in `data/`.

### Asset Processor
Utility to prepare sprites (background removal, hex-snapping to the 160×160 standard).
```bash
uv run tools/asset_processor/main.py input.png output.png
```

### Audio Generation
Audio is generated using [Piper TTS](https://github.com/rhasspy/piper). Scripts live in `scripts/`. See [assets/audio/README.md](assets/audio/README.md) for the full voice model assignment registry.

---

## 🛠️ Modding & Custom Content

Oinakos supports live moddding without recompiling. Create an `oinakos/` folder next to the game executable and override any asset or data file.

Loading priority:
1. **Local**: `oinakos/data/` or `oinakos/assets/`
2. **Embedded**: Files baked into the binary.

### Local Mod Directory Tree

```text
oinakos/
├── data/
│   ├── archetypes/          # Shared unit templates (Stats, AI profiles)
│   │   └── <category>/      # e.g., human/, orc/
│   │       └── <id>.yaml
│   ├── npcs/                # Unique/Named NPCs
│   │   └── <id>.yaml
│   ├── obstacles/           # Map objects (Trees, buildings, rocks)
│   │   └── <id>.yaml
│   ├── maps/                # Custom map levels
│   │   └── <id>.yaml
│   └── characters/          # Playable characters
│       └── <id>.yaml
├── assets/
│   ├── images/
│   │   ├── archetypes/
│   │   │   └── <category>/<id>/
│   │   │       ├── static.png   # Idle frame
│   │   │       ├── attack.png   # Attack frame
│   │   │       ├── corpse.png   # Death frame
│   │   │       ├── hit.png      # Flinch frame
│   │   │       ├── hit1.png     # (Optional) Random flinch variant 1
│   │   │       └── hit2.png     # (Optional) Random flinch variant 2
│   │   ├── obstacles/
│   │   │   └── <id>.png
│   │   └── characters/
│   │       └── <id>/
│   │           ├── static.png
│   │           ├── attack.png
│   │           └── corpse.png
│   └── audio/
│       ├── archetypes/
│       │   └── <category>/<id>/
│       │       ├── hit.wav
│       │       ├── death.wav
│       │       └── attack_*.wav
│       ├── npcs/
│       │   └── <npc_id>/        # Overrides archetype audio for unique NPCs
│       │       ├── hit.wav
│       │       └── attack_*.wav
│       └── characters/
│           └── <character_id>/  # Player character voice lines
│               ├── hit.wav
│               ├── death.wav
│               └── attack_*.wav
└── saves/                   # Save files (*.oinakos.yaml)
```

### Tips for Modders
- **Nested Folders**: Data YAML in `data/archetypes/human/male/peasant.yaml` → sprites in `assets/images/archetypes/human/male/peasant/`.
- **Palette Masking**: Paint **Magenta (#FF00FF)** for primary color and **Yellow (#FFFF00)** for secondary color on sprites. The engine swaps them at runtime using values from the YAML.
- **Audio Fallback**: Character audio uses `MainCharacter` as prefix. NPC audio falls back to archetype if no `npcs/<id>/` folder exists.
- **WAV Format**: All audio overrides must be WAV files.
- **Voice Registry**: See [assets/audio/README.md](assets/audio/README.md) for all voice model assignments.

### Interactive Environments (Actions System)

Obstacles can have `actions` — periodic effects that trigger every **1 second (60 ticks)**:

```yaml
# Damaging aura (e.g. cursed fire)
actions:
  - type: harm
    amount: 2
    aura: 1.5          # World-unit radius. 0 = use collision footprint.

# Healing well (player presses SPACE near it)
actions:
  - type: heal
    amount: 10
    requires_interaction: true
    alignment_limit: "ally"   # "all" | "ally" | "enemy"
```

---

## 📁 Project Structure

```text
/
├── cmd/              # Additional binary entry points
├── internal/
│   ├── engine/       # Low-level: Ebiten abstractions, isometric math, audio, shaders
│   └── game/         # High-level: NPC AI, combat, HUD, save system, registries
├── data/
│   ├── archetypes/   # Shared unit templates
│   ├── characters/   # Playable character definitions
│   ├── npcs/         # Unique named NPCs
│   ├── obstacles/    # Map object definitions
│   ├── maps/         # Standalone map levels
│   └── campaigns/    # Multi-map campaign definitions
├── assets/
│   ├── images/       # Sprites (archetypes, characters, obstacles, tiles)
│   └── audio/        # Voice lines and sound effects (WAV)
├── tools/
│   ├── boundaries_editor/  # Footprint editor tool
│   ├── map_editor/         # Map authoring tool
│   └── asset_processor/    # Sprite preprocessing (Python/uv)
├── scripts/          # Audio generation, bundling, and tooling scripts
├── models/           # Piper TTS voice model files (not committed)
├── bin/              # Compiled development binaries
└── dist/             # Distribution packages
```

---

## 🤝 Contributing

1. **Read the Technical Guide**: Review [GEMINI.md](GEMINI.md) for architecture rules — especially the Strict Dependency Injection constraint and coordinate system conventions.
2. **Test-Driven**: Run `make test` before submitting. The `internal/game` package is fully headless-testable.
3. **Use the Tools**: Use `make map-editor` and `make boundaries-editor` for visual content adjustments rather than editing YAML footprints by hand.

---

## 📝 Roadmap

- [ ] **Animation System**: Move from static sprites to full sprite-sheet support (walk / attack / death frames).
- [ ] **A\* Pathfinding**: Replace linear NPC tracking with proper grid navigation.
- [ ] **Dynamic Biomes**: Procedural background and ambient weather changes based on chunk distance.
- [ ] **UI Overhaul**: Replace debug-print HUD with textured panels and character portrait icons.
- [ ] **Occlusion Effect**: Render entities behind obstacles as greyscale silhouettes (plan in `plans/`).

---

## 📜 License
This project is licensed under the MIT License. Code and assets have been developed with AI-assisted tooling.
