# Oinakos — Agent Memo 🛡️

Oinakos is a performance-optimized, infinite isometric action RPG and biological ecosystem simulation built in Go. This memo is the **technical source of truth** for AI agents working on the codebase.

## 🛡️ Enforcement Rules

- **Strict File Limit**: **IMPERATIVE**. No source file may exceed **500 lines**. If an edit would push a file over this limit, you **MUST** refactor and split it before proceeding.
- **Dependency Isolation**: `internal/game` MUST NEVER import `ebiten`. All graphics/input must stay behind the `engine` interfaces.
- **Simulation Integrity**: All logic must be testable in **Headless Mode** (`-tags test` or `-tags headless`).

---

## ⚙️ Technical Core

- **Engine**: Custom `internal/engine` wrapping **Ebiten v2**.
- **Coordinate Systems**:
  - **Cartesian**: Physics, AI, collision, and logic. Standard units are **pedes** (feet).
  - **Isometric**: Rendering only. `isoX = (x - y)`, `isoY = (x + y) * 0.5`.
- **Simulation Timing**: Locked at **60 TPS**. 1 day = 17,280 ticks. 1 year = 360 days (standard).
- **Headless Mode**: Triggered via `go run -tags headless .` or `make run-headless`. Uses `MockGraphics` and `MockInput` to run the simulation without a window.
- **Adult Mode**: Toggleable in `settings.yml`. Enables arousal, physical trauma, and mature interaction mechanics.

---

## 📏 Units of Measurement (Ancient Roman)

| Roman Unit | Ratio to `pes` (Foot) | Metric (Approx) | Notes |
| :--- | :--- | :--- | :--- |
| **pes** | 1 pes | ~296 mm | Fundamental unit |
| **passus** | 5 pedes | ~1.48 m | Double step (pace) |
| **mille passus** | 5,000 pedes | ~1.48 km | Roman mile |
| **libra** | 1 libra | ~329 g | Weight unit |

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

**Age Modifiers:**
- **Young (<25)**: Physical stats penalty up to 25%, Mental up to 30%.
- **Elder (>40)**: Physical stats decay (up to 25% at age 85); Mental stats **improve** (5% per decade).

---

## 🧬 Biological & Temporal Simulation

### Core Needs (`State` Struct)
Living entities simulate needs every tick. Decay is modified by the **Health** attribute: `decayMult = 1.25 - (health * 0.01)`.

- **Hunger/Thirst**: Base rates `0.02` and `0.03`. Dehydration/Starvation leads to HP loss.
- **Fatigue**: Base rate `0.01`. Recovered by `ActorResting` near comforts.
- **Vampirism**: Immortal characters (Age Rate 0) satisfy Hunger/Thirst via **"Bloodlust"** (ActorFeeding on victims). Standard food is rejected.

### Lifecycle & Aging
- **Life Stages**: `Baby` (<1y), `Kid` (1-12y), `Teenager` (12-18y), `Adult` (18-65y), `Elder` (>65y).
- **Career Transition**: At Age 18, NPCs select a Professional Archetype (e.g., Man-at-Arms, Peasant, Priest) based on their highest Primary Attribute.
- **Death**: Natural death chance starts at **85 years**. `Age.Max` in YAML can set hard limits (0 = immortal).
- **Genetics**: Offspring inherit attributes from parents (50/50 blend) with stochastic variation (Mutation chance: 5%).

### Procreation & Romance
- **The Social Drive**: Intercourse occurs during `ShiftLeisure`. Triggered by **Relationship (>40)** or **Uninhibited Impulse** (Arousal > 50 or IsDrunk).
- **Biological Constraint**: Pregnancy only for **non-transexual biological females** mated with **non-transexual biological males**. 
- **Exclusivity**: Human procreation requires **Vaginal** intercourse.

### Trauma & Environmental Hazards
- **Physical Trauma**: Irreversible damage (Limbs, Blindness, Broken Spine).
- **Grief Cascade**: Death of a Partner/Friend triggers **"Mourning"** (Double GriefTicks). Causes massive Sanity drain, leading to Psychotic Breaks.
- **Miasma**: Dead bodies "Rot" after 1 day. Rotten corpses emit a **Miasma Cloud** (4 pedes radius). Causes Sanity loss and Sickness (Plague). AI will pathfind around Miasma.

---

## 🧠 AI Simulation Layer

### World Serialization
The engine serializes the environment into `WorldContext` JSON for AI reasoning:
- **Sanity & Breakpoints**: Critical Sanity (0) triggers `ActorBerserk` (Psychotic Break).

### Agent Bridge
External AI interacts via `agent-bridge`:
1. **Output**: Game writes `output.json` with situation and options.
2. **Input**: AI writes `input.json` with `ChosenOption` and `Reasoning`.

---

## 📜 Coding Standards (SOLID)

- **Dependency Inversion**: High-level game logic depends on `engine` interfaces, not Ebiten.
- **Composition**: Use struct embedding. `Actor` embeds `State`, `Attributes`, and `Trauma`.
- **Error Handling**: Check every error immediately. No panics in production code.
- **Interface Segregation**: Keep interfaces small (e.g., `engine.Graphics`, `engine.Input`).

---

## 🛠 Makefile Reference

| Command | Description |
| :--- | :--- |
| `make run` | Build & run native |
| `make run-headless` | Run pure simulation (JSON-AI Bridge) |
| `make test` | Run all headless unit tests |
| `make dist` | Build WASM package with WebLLM support |
| `make boundaries-editor` | Open collision footprint tool |
| `make map-editor` | Open map authoring tool |
| `make bundle-all` | Pack binaries for Mac, Win, Linux |
