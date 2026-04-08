# Oinakos 🛡️🗡️

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)
![Ebiten v2](https://img.shields.io/badge/Engine-Ebiten%20v2-orange)
![Version](https://img.shields.io/badge/Version-0.1--alpha-green)

**Oinakos** is a performance-optimized, infinite isometric action RPG and biological ecosystem simulation built with **Go** and **Ebiten v2**. It blends visceral combat with a deep, uncompromising simulation of life, death, and genetics.

> *Feel like a kid again — loading up an RPG from a CD-ROM in late 1996.*

---

## 📖 Game Overview

In Oinakos, you aren't just playing a game; you are stepping into a living world. Survive relentless combat, manage your biological needs, and explore an infinite landscape steeped in Spanish Ballad lore and dark medieval grit.

### 🎭 Playable Characters
Choose your champion. Each has unique primary attributes (Strength, Dexterity, Health, Intellect, Wisdom) and voice lines in their native tongue:

| Hero | Nationality | Weapon | Specialty |
| :--- | :--- | :--- | :--- |
| **Oinakos** | Unknown | Tizón | The mysterious lead knight forged in iron. |
| **Boris Stronesco** | Serbian | Claws | A cursed noble from Cartagena. Speaks Serbian. |
| **Roland** | French | Durandal | A disciplined paladin of the Carolingian court. |
| **Conde Olinos** | Spanish | Long Sword | A tragic noble whose song moved the sea. |
| **Gaiferos** | Spanish | Axe | A brave knight seeking his beloved Melisendra. |
| **Conde Estruch** | Catalan | Sword | A noble entangled in the mysteries of old Catalan. |
| **Princess Elvira** | Spanish | Cursed Blast | A noble sorceress seeking family redemption. |

*...and many more, including the **Peasant King**, **Queen Urraca**, and the clockwork **Virculus**.*

---

## 🗺️ Campaigns & Maps

Oinakos offers both structured narrative experiences and freeform exploration.

- **The Chronicles**: A multi-map journey through the heart of the kingdom.
- **Invasion Maps**: Defend against the *Orc Invasion* or the *Demonic Incursion*.
- **Sandbox Mode**: Explore infinite, procedurally generated chunks with no fixed objective.
- **Venburgo**: A high-fidelity, hand-crafted map optimized for long-term biological simulation testing.

---

## 🧬 Biological & Ecosystem Simulation

Oinakos features a state-of-the-art biological engine where every entity is subject to the laws of nature.

- **Core Needs**: Every tick simulates Hunger, Thirst, and Fatigue. Neglect leads to Dehydration, Starvation, and eventual death.
- **Lifecycle & Aging**: Entities progress through life stages (Baby, Kid, Teenager, Adult, Elder). Natural aging begins to impact physical stats after 40 and can lead to death after 85 years.
- **Genetics & Procreation**: Offspring inherit attributes from parents via a 50/50 genetic blend with a 5% mutation chance.
- **Vampirism**: Immortal entities (Age Rate 0) reject normal food, satisfying their needs via **Bloodlust** (feeding on victims).
- **Professional Archetypes**: At age 18, NPCs select a career (Man-at-Arms, Peasant, Priest, etc.) based on their dominant attributes.

### 💀 Trauma & Environmental Hazards
- **Physical Trauma**: Combat can lead to irreversible damage, including lost limbs, blindness, or a broken spine.
- **Grief Cascade**: The death of a partner or friend triggers a **Mourning** state, causing massive Sanity drain and potentially leading to a **Psychotic Break**.
- **Miasma & Rot**: Fallen bodies rot after 24 hours (17,280 ticks), emitting a **Miasma Cloud** that causes Sickness and Sanity loss.

---

## 🧠 AI & NPC Intelligence

Every NPC in Oinakos is powered by a multi-layered intelligence system.

- **Dynamic Barks**: NPCs generate situation-aware dialogue based on their YAML-defined personality and current world context.
- **AI Simulation Mode**: Toggle `ai_simulation_mode` to watch the AI take full control of the player character, making tactical decisions from engagement to retreat.
- **Agent Bridge**: The engine serializes the world into a `WorldContext` JSON, allowing external AI models to "see" and "think" about the environment.
- **Social Drives**: NPCs seek leisure, form relationships, and respond to social impulses based on their Sanity and Arousal levels.

---

## ⚔️ Combat & Survival

Combat in Oinakos uses a balanced, logarithmic scaling system to prevent stat inflation while keeping high-level gameplay challenging.

- **Logarithmic RPG Progress**: `Stat = Base + (log2(Level) * 10)`. Early levels feel impactful; late levels reward mastery.
- **Precision Combat**: Hit chances are calculated as `attack / (attack + defense) * 100`, clamped between 5% and 95%.
- **Hunting & Butchering**: Target local wildlife (Oxen, Wolves, Boars) to survive. Success and meat yields scale with your **Survivalism** proficiency.
- **Industry & Logistics**: Manage your weight. Use **Warehouses** and **Smitheries** to unload raw materials like wood, lumber, and ore.

---

## 📐 Units of Measurement

Oinakos uses an Ancient Roman measurement system for consistent logic and flavor:

| Roman Unit | Ratio to `pes` (Foot) | Metric (Approx) | Context |
| :--- | :--- | :--- | :--- |
| **pes** | 1 pes | ~296 mm | Fundamental logic unit |
| **passus** | 5 pedes | ~1.48 m | Double step (pace) |
| **mille passus**| 5,000 pedes | ~1.48 km | Roman Mile |
| **libra** | 1 libra | ~329 g | Weight measurement |

---

## 🚀 Getting Started

### Prerequisites
- **Go 1.21+**
- **Python 3.14** (managed via `uv` for asset/audio tools)
- **GPU support** (for Ebiten native builds)
- **Ollama** (optional, for local high-performance AI)

### Installation & Run
```bash
# Run native desktop build
make run

# Run in Browser (WebAssembly)
make serve-wasm # Visit http://localhost:8000

# Run in Headless Mode (CLI Simulation)
go run -tags headless . -fast -debug
```

### AI Configuration
Edit your `settings.yml` (located in your platform's app data directory) to configure providers:
- **Local**: Ollama (Recommended)
- **Cloud**: OpenAI, Claude, Gemini, Mistral, Hugging Face
- **WASM**: Uses **WebGPU** to run a local LLM (`Llama-3-8B`) directly in your browser.

---

## 🛠️ Development & Modding

- **Headless Mode**: Use `-tags test` or `-tags headless` to run the simulation without a window—perfect for long-term balance testing.
- **Map Editor**: `make map-editor` - Create and edit isometric levels.
- **Boundaries Editor**: `make boundaries-editor OBSTACLE=tree_oak` - Graphically edit collision footprints.
- **Live Modding**: Place files in a local `oinakos/` folder to override any embedded asset or data YAML.

### Mod Directory Structure
```text
oinakos/
├── data/
│   ├── archetypes/    # Shared unit templates
│   ├── npcs/          # Unique/Named NPC definitions
│   ├── obstacles/     # Map objects (Trees, buildings)
│   └── maps/          # Custom level definitions
└── assets/
    ├── images/        # Sprites (static.png, attack.png, etc.)
    └── audio/         # WAV voice lines and SFX
```

---

## 📜 Roadmap
- [x] **Animation System**: Initial support for animated objects (e.g., campfires). Full sheet support in progress.
- [x] **A\* Pathfinding**: Advanced grid-based navigation logic implemented for complex NPC movement.
- [x] **UI Overhaul**: Textured panels, rich inventory management, and immersive character portraits.
- [x] **Dynamic Biomes**: Procedural weather cycles and terrain-specific logic based on world coordinates.
- [x] **Miasma & Trauma**: Implemented infectious diseases and irreversible physical/mental trauma mechanics.
- [ ] **Legacy & Lineage**: Marriage, family surnames, and land inheritance across multiple generations.
- [ ] **Dynamic Global Economy**: Resource prices reacting to local harvests, bandit raids, and trade caravans.
- [ ] **Alchemy & Crafting**: Deep crafting tree for refining raw ores into alloys and brewing medical potions.
- [ ] **Naval Navigation**: Implementation of boat travel, fishing at sea, and mysterious coastal biomes.
- [ ] **Campaign Editor**: A visual tool for defining multi-map progression, quest triggers, and branching dialogue.
- [ ] **Official Scripting API**: Formal Lua or JavaScript-based hooks for external modding control.

---

## ⚖️ License
Licensed under the **MIT License**.
*Fonts used: MedievalSharp, Modern Antiqua, Uncial Antiqua, Glass Antiqua, Kings, Eagle Lake (all SIL Open Font License 1.1).*
