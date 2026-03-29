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
- **Versioning**: Current version is **0.1-alpha**.
  - Centralized in `main.go` and injected via `-ldflags "-X main.Version=$(VERSION)"` in the `Makefile`.
  - Passed to `game.NewGame` to be stored in the `Game` struct.
  - Rendered dynamically in the main menu via `g.Version`.

---

## 📏 Units of Measurement

For historical accuracy and consistency in world-building, distances and lengths in Oinakos are measured in Ancient Roman units.

| Roman Unit | English Name | Ratio to `pes` (Foot) | Metric Equivalent | Imperial Equivalent | Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **digitus** | finger | 1/16 pes | ~18.5 mm | ~0.728 in | Smallest unit |
| **uncia** | inch (thumb) | 1/12 pes | ~24.6 mm | ~0.971 in | |
| **palmus** | palm (minor) | 1/4 pes | ~74 mm | ~2.912 in | |
| **palmus maior** | palm length (major) | 3/4 pes | ~222 mm | ~8.737 in | |
| **pes** | foot | 1 pes | ~296 mm | ~0.971 ft | Fundamental unit |
| **palmipes** | foot and a palm | 1 1/4 pedes | ~370 mm | ~1.214 ft | |
| **cubitum** | cubit | 1 1/2 pedes | ~444 mm | ~1.456 ft | |
| **gradus** / **pes sestertius** | step | 2 1/2 pedes | ~0.74 m | ~2.427 ft | |
| **passus** | pace | 5 pedes | ~1.48 m | ~4.854 ft | Double step |
| **decempeda** / **pertica** | perch | 10 pedes | ~2.96 m | ~9.708 ft | Measuring rod |
| **actus** | path, track | 120 pedes | ~35.5 m | ~116.496 ft | |
| **stadium** | stade | 625 pedes | ~185 m | ~607.14 ft | 1/8 mile |
| **mille passus** / **mille passuum** | (Roman) mile | 5,000 pedes | ~1.48 km | ~0.919 mi | 1,000 paces |
| **leuga** / **leuca** | (Gallic) league | 7,500 pedes | ~2.22 km | ~1.379 mi | 1.5 miles |
| **parasanga** | parasang | 18,750 pedes | ~5.55 km | ~3.45 mi | 30 stadia (Persian origin) |

**Conversion Summary:**
*   1 **pes** (foot) ≈ 296 mm.
*   1 **passus** (pace) = 5 **pedes**.
*   1 **mille passus** (Roman mile) = 1,000 **passus** = 5,000 **pedes**.

---

### Data-Driven Registries
All game content is defined in YAML under `data/` and loaded at startup:
- **`ArchetypeRegistry`** — shared stats, sprites, audio dir, and AI profile for a category of NPC (e.g. `orc/male`).
- **`NPCRegistry`** — unique named NPCs (e.g. Stultus, Virculus). They can override archetype stats and have their own audio folder.
- **`PlayableCharacterRegistry`** — player-selectable characters. Each sets `EntityConfig.PlayableCharacter = config.ID`.
- **`MapTypeRegistry`** — both standalone maps and individual campaign map levels.
- **`CampaignRegistry`** — ordered sequences of map IDs.
- **`ObjectRegistry`** — definitions for items, weapons, and equipment (loaded from `data/objects`).

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
2. Else fall back to `assets/audio/archetypes/<archetype>/` (the archetype's voice).
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

## ⚔️ Entity Attribute & Ability System

All entity strength, behavior, and action success is driven by **Primary Attributes**. Never hardcode flat stat values — always use the attribute formulas below.

### Primary Attributes (`attributes:` in YAML)

Each attribute is an integer from **0–100**, defined on the `attributes` block of every character and archetype YAML.

| Attribute | Key | Role |
| :--- | :--- | :--- |
| **Strength** | `strength` | Melee damage, carrying capacity, chopping/digging |
| **Dexterity** | `dexterity` | Speed, attack cooldown, ranged accuracy, milking, shearing |
| **Health** | `health` | Max HP, endurance, nourishment absorption, resting |
| **Intellect** | `intellect` | Cooking quality, crafting, trade negotiation, brewing |
| **Wisdom** | `wisdom` | Foraging yield, taming, farming, prayer, leadership |

### Combat Derived Stats (`stats:` block in YAML)

Computed at runtime in `Actor.SyncStats()`. YAML values act as **overrides**.

| Stat | Formula | Max | Notes |
| :--- | :--- | :--- | :--- |
| `melee_attack` | `strength * 2` | 200 | Used by all melee damage rolls |
| `ranged_attack` | `dexterity * 2` | 200 | Used by all ranged damage rolls |
| `defense` | `dexterity * 1.5 + health * 1.0` | 250 | Damage reduction |
| `health_points` | `health * 10` | 1000 | Overridden by `health_min`/`health_max` in YAML |
| `speed` | `dexterity * 0.02` | 2.0 | Units per tick |
| `critical_chance` | `strength * 0.005` | 0.5 (50%) | Doubles damage on proc |
| `attack_cooldown` | `stats.attack_cooldown * (1.5 - dexterity * 0.01)` | — | Min clamp: 10 ticks |
| `max_weight` | `(strength * 1.5 + health * 0.5) / 0.329` | — | Roman **librae**. Must be inside `stats:` |

### Productive / Social Derived Stats

These stats power the `yield` formula of non-combat abilities. Also computed in `SyncStats()`.

| Stat | Formula | Max | Used By |
| :--- | :--- | :--- | :--- |
| `nourishment` | `health * 2` | 200 | `eat`, `drink`, `rest` — metabolic efficiency |
| `survivalism` | `strength * 0.5 + health * 0.5` | 100 | `chop`, `dig`, `forage`, `fish`, `hunt`, `trap`, `butcher`, `guard` |
| `mate` | `health * 0.01` | 1.0 | `mate` — conception probability (0–1) |
| `crafting` | `intellect * 1.2 + strength * 0.3` | 150 | `smelt`, `craft`, `repair`, `build` |
| `herbalism` | `wisdom * 1.0 + intellect * 0.5` | 150 | `cook`, `brew`, `heal` |
| `trading` | `intellect * 1.2 + wisdom * 0.3` | 150 | `trade`, `appraise` |
| `harvesting` | `wisdom * 1.2 + dexterity * 0.3` | 150 | `plant`, `harvest_crop`, `water_crops` |
| `husbandry` | `wisdom * 1.0 + dexterity * 0.5` | 150 | `milk`, `shear`, `tame`, `tend_animal`, `breed` |
| `art` | `dexterity * 0.5 + intellect * 0.5` | 100 | `perform`, `paint`, `sculpt`, `seduce` |
| `culture` | `intellect * 0.5 + wisdom * 0.5` | 100 | `compose`, `teach`, `pray`, `bury`, `read`, `intimidate`, `recruit` |

> **`nourishment` is an entity trait, not a food property.** Food items carry a fixed `nourishment_value` in their YAML. At consumption, `hunger_restored = item.nourishment_value + entity.nourishment * 0.2`.

### Action Success Resolution

To check if an action (ability) succeeds, the system performs a **uniform roll 0–100**. Success if `roll ≤ effective_threshold`. For a competitive roll (e.g. two combatants), the lower roll that is still ≤ its effective threshold wins.

**Situational Modifiers:**  
`CheckAbilitySuccess(abilityID, modifier int)` accepts an integer modifier that shifts the effective threshold:
- **Bonus** (`modifier > 0`): raises the threshold — a higher roll can still succeed.  
  e.g. `attribute=60, bonus=+20` → success if `roll ≤ 80`.
- **Penalty** (`modifier < 0`): lowers the threshold — only a better roll succeeds.  
  e.g. `attribute=60, penalty=−20` → success if `roll ≤ 40`.
- Clamped to `[0, 100]` — no modifier can make an action impossible or guaranteed.

Example modifiers:
| Situation | Modifier |
| :--- | :--- |
| Chopping in heavy rain | −10 |
| Foraging at night | −20 |
| Cooking in a fully equipped kitchen | +15 |
| Hunting downwind | +10 |
| Taming a previously encountered animal | +20 |
| Intimidating someone much weaker | +30 |

Priority order in `CheckAbilitySuccess(abilityID, modifier)`:
1. **Skill value** — check `Actor.SkillValues[abilityID]` (specialized proficiency, e.g. `sword_mastery`).
2. **Ability config** — check `Config.Abilities[abilityID].ParentAttribute` and roll against that attribute.
3. **Hardcoded fallback** — built-in mapping for common IDs (below).

**Default attribute fallbacks:**
| Action IDs | Attribute |
| :--- | :--- |
| `punch`, `kick`, `heavy_strike`, `chop`, `dig`, `build`, `butcher`, `throw`, `knockout`, `grapple` | `strength` |
| `slap`, `slash`, `shoot_arrow`, `milk`, `shear`, `sneak`, `steal`, `seduce`, `weave` | `dexterity` |
| `rest`, `eat`, `drink`, `mate` | `health` |
| `cook`, `craft`, `repair`, `brew`, `trade`, `smelt`, `read`, `appraise`, `intimidate`, `tan`, `lie` | `intellect` |
| `forage`, `plant`, `harvest_crop`, `tame`, `fish`, `pray`, `guard`, `hunt`, `trap`, `tend_animal`, `breed`, `water_crops`, `bury`, `recruit`, `teach` | `wisdom` |

**Competitive Rolls (character vs. character):**

Used when two characters actively contest the same action — e.g. one tries to `steal` while the target resists, or one tries to `intimidate` while the other defies.

1. Both characters roll 0–100 against their relevant attribute (with any applicable modifier).
2. **Both fail** → action fails; no winner.
3. **Only one succeeds** (roll ≤ threshold) → that character wins.
4. **Both succeed** → the character with the **lower roll** wins (they performed more decisively).

Examples:
| Action | Attacker rolls | Defender rolls | Winner |
|---|---|---|---|
| `steal` (attacker dex=70) vs `guard` (defender wis=60) | – | – | Both roll; lower success wins |
| `intimidate` (str=80) vs `lie` (target int=50) | attacker succeeds at 45 | defender fails at 62 | Attacker wins |
| `seduce` (art=90) vs `resist` (health=80) | both succeed; attacker 20, defender 35 | – | Attacker wins (lower roll) |

Implemented via `CompetitiveAttributeRoll(other *Actor, attr string)` in Go.

### Abilities (`abilities:` block in YAML)

Every character and archetype YAML must include an `abilities` block. Each entry defines:
- `damage` *(combat)* — formula referencing `attack` (e.g. `"attack * 1.5"`).
- `yield` *(productive)* — formula referencing a productive stat (e.g. `"cook_quality * 1.0"`).
- `parent_attribute` — attribute used for the success roll if no skill value exists.
- `required_weapon` *(optional)* — weapon type required (e.g. `sword`, `bow`).
- `effects` *(optional)* — list of status effect objects.

#### Combat Abilities

| Ability | `damage` Formula | Parent Attr | Weapon | Effects |
| :--- | :--- | :--- | :--- | :--- |
| `punch` | `melee_attack * 1.0` | `strength` | — | — |
| `slap` | `melee_attack * 0.6` | `dexterity` | — | stun 20% / 1 tick |
| `kick` | `melee_attack * 1.3` | `strength` | — | knockback 2 units |
| `slash` | `melee_attack * 1.5` | `dexterity` | sword | armor break 20% / 3s |
| `heavy_strike` | `melee_attack * 1.8` | `strength` | sword | armor break 30% / 3s |
| `shoot_arrow` | `ranged_attack * 1.2` | `dexterity` | bow | — |
| `power_shot` | `ranged_attack * 1.7` | `dexterity` | bow | pierce 2 targets |
| `throw` | `ranged_attack * 0.8` | `strength` | — | — |
| `infect_bite` | `melee_attack * 0.8` | `strength` | — | poison 5 dmg/s / 5s (90%) |
| `knockout` | `melee_attack * 0.3` | `strength` | — | stun 85% / 10s — deliberate non-lethal takedown |
| `grapple` | `melee_attack * 0.1` | `strength` | — | immobilise target while held; no kill intent |
| `chop` | `melee_attack * 1.0` | `strength` | axe | Also has `yield: survivalism * 1.0` (wood out) |
| `dig` | `melee_attack * 1.0` | `strength` | pike | Also has `yield: survivalism * 1.0` (ore/stone out) |

#### Productive / Social Abilities

| Ability | `yield` Formula | Parent Attr | Notes |
| :--- | :--- | :--- | :--- |
| `forage` | `survivalism * 0.3` | `wisdom` | Items gathered from environment |
| `cook` | `herbalism * 1.0` | `intellect` | Nourishment value of produced meal |
| `rest` | `nourishment * 0.25` | `health` | HP regen per rest tick |
| `eat` | `nourishment * 1.0` | `health` | Hunger restored when eating |
| `drink` | `nourishment * 0.8` | `health` | Thirst restored when drinking |
| `milk` | `husbandry * 1.0` | `wisdom` | Litres of milk per session |
| `shear` | `husbandry * 0.5` | `wisdom` | Wool yield per shearing session |
| `mate` | `mate * 1.0` | `health` | Conception probability (0–1) |
| `tame` | `husbandry * 1.0` | `wisdom` | Taming success roll modifier |
| `tend_animal` | `husbandry * 1.0` | `wisdom` | Healing and caring for sick livestock |
| `breed` | `husbandry * 1.0` | `wisdom` | Actively managing animal reproduction |
| `plant` | `harvesting * 0.8` | `wisdom` | Crops planted per session |
| `harvest_crop` | `harvesting * 1.5` | `wisdom` | Yield bonus over base crop output |
| `water_crops` | `harvesting * 0.5` | `wisdom` | Maintaining crops between plant and harvest |
| `fish` | `survivalism * 0.5` | `wisdom` | Fish caught per session |
| `hunt` | `survivalism * 1.0` | `wisdom` | Actively tracking and killing wild animals |
| `trap` | `survivalism * 0.8` | `wisdom` | Setting snares for passive trapping |
| `butcher` | `survivalism * 0.5` | `strength` | Processing dead animals into meat and hide |
| `brew` | `herbalism * 1.0` | `intellect` | Potency of produced drink/potion |
| `craft` | `crafting * 1.0` | `intellect` | Quality tier of crafted item |
| `repair` | `crafting * 0.5` | `intellect` | Durability restored per repair |
| `smelt` | `crafting * 0.8` | `intellect` | Purity of smelted metal ingot |
| `build` | `crafting * 1.0` | `strength` | Quality and speed of constructed structure |
| `trade` | `trading * 1.0` | `intellect` | Price negotiation modifier (%) |
| `appraise` | `trading * 0.5` | `intellect` | Item value estimation without transaction |
| `haul` | `strength * 0.01` | `strength` | Speed modifier while hauling goods |
| `stash` | `strength * 0.005` | `strength` | Organizing items into stockpiles |
| `sneak` | `dexterity * 1.0` | `dexterity` | Moving without being detected |
| `steal` | `dexterity * 0.5` | `dexterity` | Taking items without the owner noticing |
| `pray` | `culture * 0.3` | `wisdom` | Sanity/morale restored |
| `guard` | `survivalism * 0.5` | `wisdom` | Alert radius modifier |
| `heal` | `herbalism * 1.0` | `intellect` | HP restored when tending wounds |
| `teach` | `culture * 0.5` | `wisdom` | Skill XP bonus granted to student |
| `intimidate` | `culture * 1.0` | `strength` | Social coercion — forcing compliance |
| `seduce` | `art * 1.0` | `dexterity` | Romantic persuasion |
| `recruit` | `culture * 1.0` | `wisdom` | Convincing NPCs to join a faction or group |
| `bury` | `culture * 0.3` | `wisdom` | Burying the dead — affects settlement morale |
| `read` | `culture * 0.5` | `intellect` | Literacy — reading signs, books, quest items |
| `perform` | `art * 1.0` | `dexterity` | Entertainment performance |
| `compose` | `culture * 0.5` | `intellect` | Writing music, poetry, or other works |
| `paint` | `art * 0.8` | `dexterity` | Creating visual art |
| `sculpt` | `art * 0.8` | `strength` | Carving or moulding sculpture |
| `tan` | `crafting * 0.8` | `intellect` | Processing raw animal hide into leather |
| `weave` | `crafting * 0.5` | `dexterity` | Making cloth or textiles from wool or fibre |
| `lie` | `culture * 1.0` | `intellect` | Covert deception — distinct from intimidate (overt) or trade (transactional) |

**Available effect fields:**
```yaml
effects:
  - stun_chance: 0.2          # Probability of stunning (0–1)
    duration: 1.0              # Duration in seconds
  - knockback_distance: 2.0   # Cartesian units pushed back
  - armor_break_percentage: 0.2
    duration: 3.0
  - poison_damage_per_second: 5
    duration: 5.0
    probability: 0.9           # Chance to apply the effect
  - pierce_targets: 2         # Projectile passes through N targets
```

---

## 💾 Persistence System

- **Format**: YAML (`SaveData` struct). Extension: `.oinakos.yaml`.
- **Native**: Saves to `oinakos/saves/` beside the binary. Supports multiple named saves + load dialog.
- **WASM**: Persists to browser `localStorage` under key `quicksave`. Auto-resumes on page load.
- **Platform bridge**: `persistence_native.go` vs `persistence_js.go`, split via Go build tags.
- **Character identity** is stored as `player.archetype` in the save file and looked up in `PlayableCharacterRegistry` on load — `PlayableCharacter` is then set automatically from the registry.

---

## 🎨 Asset Generation Standards

### Sprites
- **Characters & NPCs**: **160×160 px**, solid **`#00FF00`** (chroma-key green) background.
- **Proportions**: Realistic human scale relative to the isometric tile.
- Required frames: `static.png`, `attack.png`, `corpse.png`. 
- Optional: `crouch.png` (required for picking up items), `back.png`, `hit.png`, `hit1.png`, `hit2.png`, `attack1.png`, `attack2.png`.

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
- `make boundaries-editor OBJECT=iron_sword`

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
│   ├── objects/            # <id>.yaml → items, weapons, and equipment definitions
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
| `make boundaries-editor OBJECT=id` | Launch footprint editor for an object |
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
