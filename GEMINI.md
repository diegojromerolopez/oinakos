# Oinakos — Agent Memo 🛡️

Oinakos is a performance-optimized, infinite isometric action RPG and biological ecosystem simulation built in Go. This memo is the **technical source of truth** for AI agents working on the codebase.

## 🛡️ Enforcement Rules

- **Strict File Limit**: **IMPERATIVE**. No source file may exceed **500 lines**. If an edit would push a file over this limit, you **MUST** refactor and split it before proceeding.
- **Dependency Isolation**: `internal/game` MUST NEVER import `ebiten`. All graphics/input must stay behind the `engine` interfaces.
- **YAML & Asset Integrity**: 
  - Every time you add or delete YAML files in `data/`, you MUST update or verify `TestForEachYAML` to reflect the expected file count or structure.
  - Assets must follow strict naming: `<id>.png` for obstacles, or a folder `<id>/` containing `static.png`, `attack.png`, etc., for archetypes.
- **Simulation Integrity**:
  - All logic must be testable in **Headless Mode** (`-tags test` or `-tags headless`).
  - When in doubt, if the simulation fails or crashes, fix it according to the behavior of the real world.
- **Tests**:
  - If you add a new feature, add a test for it.
  - Tests must pass after each change.
  - Every fix you do needs to be covered by a test.
  - If a test fails, fix it before proceeding.

---

## ⚙️ Technical Core

- **Engine**: Custom `internal/engine` wrapping **Ebiten v2**.
- **Coordinate Systems**:
  - **Cartesian**: Physics, AI, collision, and logic. Standard units are **pedes** (feet).
  - **Isometric**: Rendering only. `isoX = (x - y)`, `isoY = (x + y) * 0.5`.
- **Simulation Timing**: Locked at **60 TPS**. 
  - 1 hour = 720 ticks.
  - 1 day = 17,280 ticks. 
  - 1 year = 360 days (6,220,800 ticks).
- **Headless Mode**: Triggered via `go run -tags headless .` or `make run-headless`. Uses `MockGraphics` and `MockInput`.
- **Long-term Simulation**: 
  `go run -tags headless . -fast -debug > simulation_1year.log 2>&1`
  *Use `-fast` for $10\times$ speed and `-debug` for detailed biological/breeding logs.*
- **Versioning**: Distributed via `LDFLAGS=-ldflags "-X main.Version=$(VERSION)"`.

---

## 📏 Units & Conversions (Ancient Roman)

| Roman Unit | Ratio to `pes` (Foot) | Metric (Approx) | Context |
| :--- | :--- | :--- | :--- |
| **pes** | 1 pes | ~296 mm | Logic fundamental |
| **passus** | 5 pedes | ~1.48 m | Pace (Double step) |
| **mille passus**| 5,000 pedes | ~1.48 km | Roman Mile |
| **libra** | 1 libra | ~329 g | Weight (R. Pound) |

---

## 🖼️ Asset Standards & Loading

- **Green Screen**: Assets use **Lime Green (`#00FF00`)** as a transparency key. The `engine.Transparentize()` function handles this at runtime.
- **Obstacle States**: Registries look for specific files in `assets/images/obstacles/<id>/`:
  - `open.png`, `closed.png` (Chests/Doors)
  - `growing.png`, `ready.png` (Crops)
- **Archetype Sprites**: Found in `assets/images/archetypes/<id>/` or `animals/<id>/`:
  - `static.png`, `back.png`, `attack.png`, `hit.png`, `corpse.png`, `crouch.png`, `pregnant.png`, `resting.png`.
- **Object Footprints**: Obstacles and large objects require a `footprint` (list of `x,y` points) in their YAML for proper collision.

---

## ⚔️ Entity Attribute & Ability System

### Primary Attributes (`0–100`)
- **Strength**: Melee damage, carrying capacity, chopping/digging.
- **Dexterity**: Speed, attack cooldown, ranged accuracy, fine motor skills.
- **Health**: Max HP, decay resistance, physical endurance.
- **Intellect**: Crafting, trade negotiation, memory processing.
- **Wisdom**: Foraging, taming, leadership, morale.

### Derived Stats (Calculated in `SyncStats()`)
| Stat | Formula | Notes |
| :--- | :--- | :--- |
| `melee_attack` | `strength * 2` | Melee damage basis |
| `ranged_attack` | `dexterity * 2` | Ranged damage basis |
| `defense` | `dexterity * 1.5 + health * 1.0` | Damage reduction |
| `health_points` | `health * 10` | Max HP |
| `speed` | `dexterity * 0.02` | Cartesian units/tick |
| `attack_cooldown`| `baseCD * (1.5 - dexterity * 0.01)` | Min clamp: 10 ticks |
| `max_weight` | `(strength * 1.5 + health * 0.5) / 0.329` | Roman **librae** |
| `survivalism` | `wisdom * 0.4 + intellect * 0.3 + health * 0.2 + dexterity * 0.1` | Survival proficiency |

**Combat Mechanics:**
- **Hit Chance**: `attack / (attack + defense) * 100` (Clamped 5%–95%).
- **RPG Scaling**: `Stat = Base + (log2(Level) * 10)` (Prevents inflation).

---

## 🧬 Biological & Temporal Simulation

### Core Needs (`State` Struct)
Living entities simulate needs every `SimStep` (default 10 ticks). Decay mult: `1.25 - (health * 0.01)`.
- **Hunger**: Base rate `0.02`. **Thirst**: Base rate `0.03`. **Fatigue**: Base rate `0.01`.
- **Vampirism**: satisfy Hunger/Thirst via **"Bloodlust"** (feeding). Normal food rejected.

### Lifecycle & Aging
- **Stages**: `Baby` (<1y), `Kid` (1-12y), `Teenager` (12-18y), `Adult` (18-65y), `Elder` (>65y).
- **Death**: Natural death chance starts at **85 years**. `Age.Max` in YAML can set hard limits.
- **Genetics**: 50/50 parental blend + 5% mutation chance.

### Adult Mode & Trauma
- **Procreation**: Triggered during `ShiftLeisure`. Requires **Relationship (>40)** or **Arousal (>50)**.
- **Physical Trauma**: Irreversible (Limbs, Blindness, Broken Spine).
- **Miasma**: Dead bodies "Rot" after 1 day, emitting a **4 pedes** radius cloud causing Sickness/Sanity drain.

---

## 🧠 AI & NPC Intelligence

### AI Providers
- **Native**: Local models (native to Go or via local services).
- **Cloud**: OpenAI, Claude, Gemini, Mistral, Hugging Face.
- **WASM**: WebGPU-accelerated models (e.g., `Llama-3-8B`) running in-browser.
- **Bridge**: Polls `~/.oinakos/headless/output.json` for situation and waits for `input.json` decision.

### External Interaction (`agent-bridge`)
1. **Output**: Game writes current `WorldContext` and valid options.
2. **Input**: AI writes `ChosenOption` and `Reasoning`.

---

## 🛠 Developer Tooling

- **Boundaries Editor**: `make boundaries-editor OBSTACLE=tree_oak`
  - Edit collision footprints visually.
  - Supports `--obstacle`, `--npc`, `--character`, `--object`.
- **Map Editor**: `make map-editor`
  - Paint tiles, place obstacles, and define spawn points.
- **Headless Profiling**: Use `-fast` for high-speed simulation testing.

---

## 📜 Coding Standards (SOLID)

- **Dependency Inversion**: High-level game logic depends on `engine` interfaces.
- **Composition**: `Actor` embeds `State`, `Attributes`, and `Trauma`.
- **Error Handling**: Check every error immediately. No panics in production.
- **Package Separation**: `internal/engine` for infra, `internal/game` for logic.

---

## 📂 Repository Structure (Modding)

```text
oinakos/ (or root)
├── data/              # Simulation Definitions (YAML)
│   ├── animals/       # Animal species and stats
│   ├── archetypes/    # Shared human/NPC templates
│   ├── obstacles/     # Buildings, Trees, Containers
│   └── maps/          # Level and Chunk definitions
├── assets/            # Multimedia Assets
│   ├── images/        # Sprites (Lime Green #00FF00 background)
│   ├── audio/         # Voice lines and SFX
│   └── fonts/         # TTF Medieval/Antique fonts
```
