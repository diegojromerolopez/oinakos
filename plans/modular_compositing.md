# Plan: Modular Compositing & Character Generator System (Paper Doll)

> **Status:** Design complete. Ready for Phase 0 implementation.
> **Scope:** New system for Modular characters only. Zero changes to existing Legacy characters.
> **Canonical Rule:** When this document conflicts with any other source, this document wins.

---

## 1. Executive Summary

The current system renders each character from a single flat image (`static.png`, `back.png`, etc.). This makes dynamic visual changes — such as equipping armor that visually appears on the sprite, breaking that armor, stripping a prisoner, or flaying a torture victim — impossible without producing a combinatorial explosion of pre-baked permutation images.

The **Paper Doll** system introduces a composited layer stack. Every character's appearance is assembled at runtime from independently controlled transparent image layers, all drawn at the same screen coordinate:

```
Layer 1 → Naked Base Body Parts (Legs → Hips → Torso → Arms → Head)
Layer 2 → Undergarments (optional slot)
Layer 3 → Clothing (optional slot)
Layer 4 → Armor (per-bodyzone slots)
Layer 5 → Weapon (in-hand overlay)
Layer 6 → Status Effects (blood, chains, poison — always on top)
```

All layers are `160×160 px`, transparent by default. When a slot is empty, that layer is simply not drawn.

---

## 2. Core Principles

### 2.1 The Standardized Frame (160×160)

Every image in this system — base body parts, armor overlays, clothing, status effects — **must be exactly 160×160 pixels**.

- **Background:** Transparent PNG (preferred). Chroma-key green `#00FF00` is also accepted and stripped at load time by the existing `LoadSprite` logic.
- **Anchor point:** The character's feet are positioned at the **horizontal center, vertical bottom** of the 160×160 canvas. This matches the existing `static.png` convention exactly.
- **Rendering:** The engine passes the same `(isoX, isoY)` screen coordinate to every layer. No per-layer offset is computed or stored.
- **Consequence of this rule:** Because all images share the same anchor, any `Head_v2` will correctly sit on top of any `Torso_v5` without any manual alignment work. This is the foundational constraint that makes combinatorial variety possible. **Violating it breaks the entire system.**

### 2.2 Absolutely Naked Base Layer

The Base Body Part images (Head, Torso, Arms, Hips, Legs segments) show **raw anatomy only**: skin, muscle, fat, bone. There are **no built-in clothes, undergarments, or accessories** of any kind on these images.

- Clothes, undergarments, and armor are applied as separate Overlay layers on top.
- Removing all overlay layers reveals the naked base. This is the mechanism behind prisoner stripping, armor breakage, and anatomy swapping. No new images are needed for any of these scenarios.

### 2.3 Variation Sets (Combinatorial Variety)

Each archetype or character type can supply **multiple interchangeable variants** of any body part. Each variant is a separate subfolder containing pose frames (see §2.5).

Example variants for a Human Male archetype:

| Body Zone | Subfolder Name | Visual Description |
| :--- | :--- | :--- |
| Head | `v1` | Young, clean-shaven |
| Head | `v2` | Middle-aged, stubble |
| Head | `v3` | Scarred, missing ear |
| Torso | `v1_skin` | Muscular, normal skin |
| Torso | `v2_skin` | Slim, normal skin |
| Torso | `v1_flayed` | Muscular, flayed (see §2.4) |
| Arms | `v1` | Normal arms |
| Arms | `v2` | Tattooed arms |
| Legs | `v1` | Normal legs |

**Combinatorial math example:** 4 heads × 3 torsos × 2 arm sets × 2 leg sets = **48 visually unique characters** from 11 subfolder sets. At 10 variants each body zone: 10⁴ = **10,000 combinations**.

At NPC spawn time, the engine picks one variant per body zone. Selection is either:
- **Deterministic:** From the character YAML `parts:` section (e.g. `head: "v3"`).
- **Seeded random:** From a stable seed derived by hashing the NPC's unique `ID` string. This ensures the same NPC always gets the same random appearance across save/load cycles without storing the part selection explicitly.

Once selected, the combination is fixed for the lifetime of that actor (unless changed by a game event like flaying — see §6).

### 2.4 Alternative Anatomy Sets

The Base Layer is not limited to normal human skin. A torso subfolder can contain any anatomy type. The anatomy type is encoded **in the subfolder name** using a suffix convention:

| Anatomy Suffix | Meaning | Use Case |
| :--- | :--- | :--- |
| `_skin` | Normal human epidermis | Default living state |
| `_flayed` | Exposed muscle, no epidermis | Post-flaying torture state |
| `_skeleton` | Bare bone only | Skeletons, undead, death transition |
| `_undead` | Rotted/decomposed flesh | Zombies, revenants |

Example: `torsos/v1_skin/`, `torsos/v1_flayed/`, `torsos/v1_skeleton/` are three anatomy variants of the **same body proportions**. Switching anatomy means swapping the active `BodyPartPoses` pointer for that zone. Armor overlays remain unchanged because of the 160×160 anchor convention — a breastplate drawn over `v1_skin` will sit identically over `v1_flayed`.

If a character YAML specifies `torso: "v1_flayed"`, the engine loads from the `torsos/v1_flayed/` subfolder. If the game later triggers a flaying event, the engine swaps the pointer from `v1_skin` to `v1_flayed` at runtime and sets `SpriteDirty = true`.

### 2.5 Pose-per-Part Library

Pixel art is **never rotated in code**. All directional and action variants are pre-rendered as separate image files. Each part variant subfolder contains the following files:

| Filename | Required? | Used When |
| :--- | :---: | :--- |
| `static.png` | **Yes** | Idle, Walking — facing SW or SE |
| `back.png` | No | Idle, Walking — facing NW or NE. Falls back to `static.png` if absent. |
| `attack.png` | No | `ActorAttacking` state (all directions). Falls back to `static.png` if absent. |
| `crouch.png` | No | `ActorCrouching` state (all directions). Falls back to `static.png` if absent. |
| `corpse.png` | No | `ActorDead` or `ActorIncapacitated` (all directions). Falls back to `static.png` if absent. |

> **Facing directions — horizontal flip rule:**
> The engine only stores two orientations per pose: "facing left" (SW/NW) and "facing right" (SE/NE).
> "Facing right" is produced by horizontally flipping the "facing left" source image at render time (`flip = -1.0`).
> This is identical to the current legacy renderer behaviour. **Only left-facing source images need to be created.**

> **Attack/Crouch/Corpse and facing back:** These action states use the *same* `attack.png`/`crouch.png`/`corpse.png` regardless of whether the character is facing front or back. There is no `back_attack.png`. If a directionally-specific attack pose is needed in the future, it can be added as `back_attack.png` with a corresponding code path in `ActiveImage()`.

### 2.6 Archetype-to-Character Inheritance for Parts

Modular mode follows the **same inheritance chain** as the current Legacy system:

1. If a **character** has its own `parts/` folder → use character-local parts. `RenderingModeModular`.
2. Else if the character's **archetype** has a `parts/` folder → inherit the archetype's `BaseVariations`. `RenderingModeModular`.
3. Else → `RenderingModeLegacy`. Use `StaticImage` etc. as today.

This means you only need to create `parts/` folders for archetypes (e.g. `archetypes/human/male/parts/`) to get Modular rendering for all NPCs that use that archetype. Individual characters can override with their own `parts/` folder.

---

## 3. Asset Directory Layout

The presence of the `parts/` subfolder is the **single detection signal** that switches a character from Legacy to Modular rendering mode (see §7).

```
assets/images/
├── archetypes/
│   └── human/
│       └── male/
│           ├── static.png        ← Legacy fallback
│           └── parts/            ← Archetype-level Modular parts (inherited by NPCs)
│               ├── heads/
│               │   ├── v1/
│               │   │   └── static.png
│               │   └── v2/
│               │       └── static.png
│               ├── torsos/
│               │   ├── v1_skin/
│               │   │   └── static.png
│               │   └── v1_flayed/
│               │       └── static.png
│               ├── arms/
│               │   └── v1/
│               │       └── static.png
│               ├── hips/
│               │   └── v1/
│               │       └── static.png
│               └── legs/
│                   └── v1/
│                       └── static.png
│
├── characters/
│   └── boris_stronesco/
│       │
│       │   ← Legacy fallback files (used ONLY if no parts/ folder exists here or in archetype)
│       ├── static.png
│       ├── corpse.png
│       │
│       └── parts/      ← Character-local parts (OVERRIDES archetype parts entirely)
│           ├── heads/
│           │   └── v1/
│           │       ├── static.png   ← REQUIRED per variant
│           │       ├── back.png
│           │       ├── attack.png
│           │       ├── crouch.png
│           │       └── corpse.png
│           ├── torsos/
│           │   └── v1_skin/
│           │       ├── static.png
│           │       └── corpse.png
│           ├── arms/
│           │   └── v1/
│           │       └── static.png
│           ├── hips/
│           │   └── v1/
│           │       └── static.png
│           └── legs/
│               └── v1/
│                   └── static.png
│
└── objects/
    └── iron_plate_armor/
        │
        │   ← Overlay folder for equipment visual representation
        └── overlays/
            │   (One subfolder per body zone this item covers)
            ├── torso/
            │   ├── static.png    ← Drawn over torso when equipped
            │   └── corpse.png    ← Drawn over torso when ActorDead
            └── arms/
                └── static.png
```

**Rules:**
- Every body zone subfolder (`heads/`, `torsos/`, `arms/`, `hips/`, `legs/`) must contain at least one variant with `static.png`. If a zone folder is entirely absent, that zone renders as transparent (nothing drawn).
- Object overlay subfolders under `overlays/` are named by body zone: `head/`, `torso/`, `arms/`, `hips/`, `legs/`.
- Object overlay image selection follows the same pose fallback rules as Base Parts (§2.5).

---

## 4. Data Structures (Go)

All new types are added to `internal/game/` package. Existing types are extended with new fields tagged `yaml:"-"` so no YAML schema changes break existing files.

### 4.1 New Types (add to `types_config.go`)

```go
// BodyPartPoses holds the pose-specific images for one body part variant.
// Only Static is required; all others fall back to Static if nil.
type BodyPartPoses struct {
    Static engine.Image // Required
    Back   engine.Image // Optional: facing NW/NE idle. Nil → use Static.
    Attack engine.Image // Optional: ActorAttacking (all directions). Nil → use Static.
    Crouch engine.Image // Optional: ActorCrouching (all directions). Nil → use Static.
    Corpse engine.Image // Optional: ActorDead/Incapacitated (all directions). Nil → use Static.
}

// ActiveImage returns the correct image for a given ActorState and facing direction.
// Action states (Attack/Crouch/Corpse) ignore facingBack — they always use the
// front-facing action frame. facingBack only applies to idle/walk states.
// Returns nil only if Static is nil (broken asset).
func (p *BodyPartPoses) ActiveImage(state ActorState, facingBack bool) engine.Image {
    switch state {
    case ActorDead, ActorIncapacitated:
        if p.Corpse != nil { return p.Corpse }
    case ActorAttacking:
        if p.Attack != nil { return p.Attack }
    case ActorCrouching:
        if p.Crouch != nil { return p.Crouch }
    default: // ActorIdle, ActorWalking, ActorDrinking
        if facingBack && p.Back != nil { return p.Back }
    }
    return p.Static
}

// BodyPartVariants holds all loaded variants for one body zone.
type BodyPartVariants struct {
    // Key: subfolder name (e.g. "v1", "v2_tattoos", "v1_flayed")
    // Order: sorted alphabetically so that seeded random picks are stable.
    Variants map[string]BodyPartPoses
    Keys     []string // Sorted keys for indexed/random access
}

// BodyPartSets holds all variant options for all body zones of one character config.
// Loaded once at startup from the parts/ directory.
type BodyPartSets struct {
    Heads  BodyPartVariants
    Torsos BodyPartVariants
    Arms   BodyPartVariants
    Hips   BodyPartVariants
    Legs   BodyPartVariants
}

// ActiveBodyParts holds the single selected BodyPartPoses per body zone for one Actor instance.
// Populated during NPC/character spawn from BodyPartSets based on YAML config or random seed.
type ActiveBodyParts struct {
    Head     BodyPartPoses
    Torso    BodyPartPoses
    Arms     BodyPartPoses
    Hip      BodyPartPoses
    Legs     BodyPartPoses

    // Store the variant keys to support anatomy swaps and save/load.
    HeadKey  string
    TorsoKey string
    ArmsKey  string
    HipKey   string
    LegsKey  string
}

// StatusOverlay represents a temporary visual effect drawn on top of everything.
type StatusOverlay struct {
    ID    string       // Unique key (e.g. "neck_chain", "blood_splatter", "poison_drool")
    Image engine.Image // 160×160 transparent image
}

// RenderingMode determines which rendering path DrawActor uses.
type RenderingMode int

const (
    RenderingModeLegacy  RenderingMode = iota // Draws StaticImage/BackImage (existing behaviour)
    RenderingModeModular                       // Composites ActiveBodyParts + Overlays
)
```

### 4.2 EntityConfig Changes (add to `registry_entity.go`)

```go
type EntityConfig struct {
    // ── All existing fields remain here, unchanged ──

    // Modular rendering fields. Never serialized to YAML (yaml:"-").
    // Only populated when RenderMode == RenderingModeModular.
    RenderMode     RenderingMode `yaml:"-"` // Set by LoadAssets based on parts/ detection
    BaseVariations BodyPartSets  `yaml:"-"` // All loaded variant options for this config
}
```

### 4.3 Actor Changes (add to `actor.go`)

```go
type Actor struct {
    // ── All existing fields remain here, unchanged ──

    // Modular rendering state. Only used when Config.RenderMode == RenderingModeModular.
    ActiveParts      ActiveBodyParts  `yaml:"-"` // The selected variants for this actor instance
    StatusOverlays   []StatusOverlay  `yaml:"-"` // Temporary visual effects (chains, blood, poison)
    CompositedSprite engine.Image     `yaml:"-"` // Off-screen cache buffer (160×160)
    SpriteDirty      bool             `yaml:"-"` // true = rebuild cache before next draw
}
```

### 4.4 ObjectConfig Changes (add to `types_config.go`)

```go
type ObjectConfig struct {
    // ── All existing fields remain here, unchanged ──

    // Overlay images per body zone. Populated by LoadAssets if overlays/ folder exists.
    // Key: body zone name ("head", "torso", "arms", "hips", "legs").
    // Value: pose frames for that zone's overlay (same fallback rules as BodyPartPoses).
    Overlays map[string]BodyPartPoses `yaml:"-"`
}
```

### 4.5 Slot Name Mapping

The current `Actor.Slots` map uses keys like `"head"`, `"body"`, `"weapon"`, `"shield"`, `"ring1"`, `"ring2"`. The Modular system introduces **new slot keys** for per-zone armor. These are **additions** — existing slot keys remain valid and functional for Legacy mode.

| Slot Key | Layer Priority | Purpose | Notes |
| :--- | :---: | :--- | :--- |
| `"undergarment"` | 6 | Loincloth, underwear | **New slot.** |
| `"clothing"` | 7 | Tunic, pants, dress | **New slot.** Replaces the visual role of `"body"` for Modular characters. |
| `"legs_armor"` | 8 | Greaves, leg armor | **New slot.** |
| `"hip_armor"` | 9 | Faulds, belt armor | **New slot.** |
| `"torso_armor"` | 10 | Breastplate, chainmail | **New slot.** |
| `"arms_armor"` | 11 | Pauldrons, gauntlets | **New slot.** |
| `"head_armor"` | 12 | Helmet, crown | **New slot.** Replaces visual role of `"head"` for Modular characters. |
| `"weapon"` | 13 | Sword, axe, bow | **Existing slot.** Uses `ObjectConfig.Overlays["weapon"]` for Modular visual. |
| `"shield"` | — | Shield (drawn as part of arms) | **Existing slot.** Future: could get its own overlay. |
| `"ring1"`, `"ring2"` | — | Rings (no visual) | **Existing slots.** No overlay, stats only. |

**Migration rule:** When an equipped `ObjectConfig` has an `Overlays` map, the compositor resolves the visual layers from the object's `Overlays` field, regardless of which slot key the item is in. This means an item in slot `"body"` with `Overlays: {"torso": ..., "arms": ...}` will contribute its torso overlay at Priority 10 and arms overlay at Priority 11.

### 4.6 YAML Configuration

These are **additions** to existing YAML schemas. All new keys are optional. Omitting them does not break anything.

#### Character YAML (`data/characters/<id>.yaml` or `data/archetypes/<id>.yaml`)
```yaml
id: boris_stronesco
name: "Boris Stronesco"
# parts: section is OPTIONAL.
# If omitted entirely → Legacy rendering mode (existing behaviour, no change).
# If present → Modular rendering mode is activated.
parts:
  # Each key is a body zone. Value is a variant subfolder name.
  # Omit a zone → seeded-random selection (stable across save/load, seeded by NPC ID hash).
  # Set to "random" → same as omitting (explicit random).
  # Set to a specific name → that variant is always used.
  head: "v2"         # Always use heads/v2/
  torso: "v1_skin"   # Always use torsos/v1_skin/
  arms: "random"     # Pick randomly at spawn (same pick every time for this NPC ID)
  hips: "v1"         # Always use hips/v1/
  legs: "random"     # Pick randomly at spawn
```

#### Equipment YAML (`data/objects/<id>.yaml`)
```yaml
id: iron_plate_armor
name: "Iron Plate Armor"
slot: "torso_armor"    # Uses new Modular slot name
effects:
  protection:
    increase: 15
# overlays: section is OPTIONAL.
# If omitted → this item has no visual representation (stats-only, Legacy behaviour).
# If present → the loader looks for overlays/<zone>/ under this object's asset folder.
overlays:
  torso: true   # Load from assets/images/objects/iron_plate_armor/overlays/torso/
  arms: true    # Load from assets/images/objects/iron_plate_armor/overlays/arms/
  # head: true  # Uncomment to add a helmet overlay for this item too
```

> **Multi-zone items:** A single item (e.g. `iron_plate_armor` in slot `torso_armor`) can provide overlays for **multiple zones** (`torso` + `arms`). The compositor iterates the item's full `Overlays` map and draws each zone's overlay at its correct priority. The slot key determines equip/unequip logic; the `Overlays` keys determine which visual layers appear.

---

## 5. Rendering Pipeline

### 5.1 Layer Compositing Order (Back to Front)

All layers are drawn at the **same (isoX, isoY) screen coordinate** into the `CompositedSprite` buffer. Lower priority numbers are drawn first (appear behind higher numbers).

The compositor iterates ALL equipment slots and maps each overlay zone to its correct priority:

| Priority | Layer Name | Source | Condition to Draw |
| :---: | :--- | :--- | :--- |
| 1 | Naked Base — Legs | `Actor.ActiveParts.Legs` | Always (if not nil) |
| 2 | Naked Base — Hips | `Actor.ActiveParts.Hip` | Always (if not nil) |
| 3 | Naked Base — Torso | `Actor.ActiveParts.Torso` | Always (if not nil) |
| 4 | Naked Base — Arms | `Actor.ActiveParts.Arms` | Always; skip if both `LeftArmLost` and `RightArmLost` |
| 5 | Naked Base — Head | `Actor.ActiveParts.Head` | Always (if not nil) |
| 6 | Undergarment Overlay | Any item in any slot that has `Overlays["undergarment"]` | Item equipped |
| 7 | Clothing Overlay | Any item in any slot that has `Overlays["clothing"]` | Item equipped |
| 8 | Legs Armor Overlay | Any item in any slot that has `Overlays["legs"]` | Item equipped |
| 9 | Hips Armor Overlay | Any item in any slot that has `Overlays["hips"]` | Item equipped |
| 10 | Torso Armor Overlay | Any item in any slot that has `Overlays["torso"]` | Item equipped |
| 11 | Arms Armor Overlay | Any item in any slot that has `Overlays["arms"]` | Item equipped; skip if both arms lost |
| 12 | Head Armor Overlay | Any item in any slot that has `Overlays["head"]` | Item equipped |
| 13 | Weapon Overlay | Any item in any slot that has `Overlays["weapon"]` | Item equipped |
| 14 | Status Overlays | `Actor.StatusOverlays[]` images | List not empty |

> **Layer order is the same regardless of facing direction.** Facing direction only affects which pose frame image is selected within each layer (see §5.2).

> **Overlay resolution rule:** The compositor does NOT look up overlays by slot key. Instead, it iterates all equipped items across ALL slots, and for each item that has an `Overlays` map, draws each zone overlay at the priority defined above. This means a single item in slot `"torso_armor"` with `Overlays{"torso": ..., "arms": ...}` draws at both Priority 10 and Priority 11.

### 5.2 Pose Frame Selection Per Layer

For each layer drawn, the correct image is selected by calling `BodyPartPoses.ActiveImage(state, facingBack)`:

| Actor State | `facingBack = true` (NW or NE) | `facingBack = false` (SW or SE) |
| :--- | :--- | :--- |
| `ActorIdle`, `ActorWalking`, `ActorDrinking` | `Back` → fallback `Static` | `Static` |
| `ActorAttacking` | `Attack` → fallback `Static` | `Attack` → fallback `Static` |
| `ActorCrouching` | `Crouch` → fallback `Static` | `Crouch` → fallback `Static` |
| `ActorDead`, `ActorIncapacitated` | `Corpse` → fallback `Static` | `Corpse` → fallback `Static` |

> **Note:** Action states (`Attack`, `Crouch`, `Corpse`) use the **same image** regardless of facing direction. The `facingBack` parameter only applies to idle/walk states to select `Back` vs `Static`. This matches the current legacy sprite convention where there is no `back_attack.png`.

**Horizontal flip:** If `Actor.Facing` is `DirSE` or `DirNE`, the image is horizontally flipped (`scale(-1, 1)`) as the current renderer does. This applies identically to every layer including overlays.

### 5.3 Trauma — Layer Suppression Rules

When a `PhysicalTrauma` flag is active, certain layers are suppressed (not drawn):

| Trauma Flag | Layers Suppressed |
| :--- | :--- |
| `LeftArmLost && RightArmLost` | Entire Arms base (Priority 4) + Arms armor overlay (Priority 11) |
| `LeftLegLost && RightLegLost` | Entire Legs base (Priority 1) + Legs armor overlay (Priority 8) |

> **Partial limb loss (only one side lost):** The initial implementation suppresses the entire zone only when BOTH sides are lost. Single-side loss is visually handled by the existing palette-swap shader's geometric clipping logic (which clips left/right halves of the zone). In a future iteration, artists can provide `_missing_left.png` / `_missing_right.png` variants per zone as a higher-quality alternative.

### 5.4 Palette Swap Shader Interaction

The existing `paletteSwapShaderSource` handles:
- `PrimaryColor` / `SecondaryColor` palette swapping (magenta/yellow masks)
- Trauma geometric clipping (arm/leg removal)
- `BurnedAlive` darkening
- `StatusTint` (poison green, etc.)

**For Modular characters:**
1. The compositor first draws all layers (Priority 1–14) into the `CompositedSprite` buffer **without the shader**.
2. The shader is then applied **once** when `CompositedSprite` is drawn to the screen — exactly as it is applied to `StaticImage` today.
3. This means palette masking, trauma clipping, and tint all work identically on the final composited image. **No shader changes are needed.**

> **Exception — `BurnedAlive`:** For Modular characters, `BurnedAlive` can optionally swap anatomy to a `_burned` variant (similar to `_flayed`) for higher visual quality. The shader darkening remains as a fallback if no `_burned` variant exists.

### 5.5 Composite Caching (Performance)

Each Modular actor maintains an off-screen **`CompositedSprite`** (a single 160×160 `engine.Image`). All layer compositing is done into this buffer; only the final buffer is blitted to the screen each frame.

**Rebuild condition — `SpriteDirty = true` is set when:**
- Any equipment slot changes (item equipped, unequipped, or item breaks).
- `Actor.State` transitions to any different value.
- Any `PhysicalTrauma` flag changes.
- `ActiveParts` is modified (anatomy swap, e.g. `skin → flayed`).
- A `StatusOverlay` is added or removed from `Actor.StatusOverlays`.

**Draw logic (in `DrawActor`):**
```
if actor.Config.RenderMode == RenderingModeModular {
    if actor.SpriteDirty || actor.CompositedSprite == nil {
        rebuildCompositeSprite(actor)  // draws all layers into CompositedSprite
        actor.SpriteDirty = false
    }
    // Apply palette swap shader on the composited buffer (same as legacy single-sprite)
    screen.DrawImageWithShader(actor.CompositedSprite, paletteShader, uniforms, op)
} else {
    // Legacy path — unchanged current code
    screen.DrawImage(DrawActorGetSprite(actor), op)
}
```

### 5.6 Memory Budget Estimate

Each 160×160 RGBA image uses `160 × 160 × 4 = 102,400 bytes ≈ 100 KB`.

| Scenario | Image Count | VRAM |
| :--- | :--- | :--- |
| 5 archetypes × 5 zones × 5 variants × 2 poses | 250 images | ~25 MB |
| 10 archetypes × 5 zones × 10 variants × 5 poses | 2,500 images | **~250 MB** |
| + 20 objects × 3 zones × 2 poses | +120 images | ~12 MB |
| + per-actor CompositedSprite (50 on-screen actors) | +50 images | ~5 MB |

**Mitigation strategies (implement if budget exceeds 100 MB):**
1. **Shared archetype pools:** NPCs that share an archetype share the same `BodyPartSets`. Only one set of variant images is loaded per archetype, not per NPC. This is already the design (variations live on `EntityConfig`).
2. **Lazy pose loading:** Only load `static.png` at startup. Load `attack.png`, `corpse.png`, etc. on first use and cache. Reduces initial load by ~60%.
3. **Variant cap per archetype:** Limit to 5 variants per zone per archetype. 5 archetypes × 5 zones × 5 variants × 2 poses = 250 images = ~25 MB. This is very comfortable.

---

## 6. Dynamic Visual Events (Reference Implementation)

### 6.1 Prisoner / Capture

**Trigger:** A game event (script, story node, or mechanic) calls `CaptureActor(actor)`.

**Steps:**
1. Unequip all items from slots: `undergarment`, `clothing`, `legs_armor`, `hip_armor`, `torso_armor`, `arms_armor`, `head_armor`. Items go to the world floor or are confiscated (game-design decision, not renderer concern).
2. Add `StatusOverlay` entries to `Actor.StatusOverlays`:
   - `StatusOverlay{ID: "neck_chain", Image: <loaded from assets/images/status/neck_chain.png>}`
   - `StatusOverlay{ID: "arm_shackles", Image: <loaded from assets/images/status/arm_shackles.png>}`
3. Set `Actor.SpriteDirty = true`.

**Result:** On the next frame, the compositor draws only the naked base anatomy + chains at Priority 14. No new sprites were needed.

### 6.2 Flaying

**Trigger:** A game event calls `FlayActor(actor)`.

**Steps:**
1. For each body zone in `actor.ActiveParts`, derive the `_flayed` key from the current key (e.g., `HeadKey = "v1"` → look for `"v1_flayed"`; `TorsoKey = "v1_skin"` → look for `"v1_flayed"`). If the `_flayed` variant exists in `actor.Config.BaseVariations`, swap the active poses and key. If absent, the zone's current base image remains unchanged.
2. Set `Actor.SpriteDirty = true`.

**Result:** The anatomy changes on the next compositor rebuild. Armor overlays remain identical because the canvas anchor is unchanged.

### 6.3 Armor Breakage

**Trigger:** An armor item's durability reaches zero (future mechanic).

**Steps:**
1. Unequip the item from its slot (same as prisoner, but only one slot).
2. Optionally replace it with a `<item_id>_broken` variant that has a damaged overlay sprite.
3. Set `Actor.SpriteDirty = true`.

**Result:** The broken or missing armor is visible within one frame.

---

## 7. Hybrid Rendering (Backward Compatibility)

**Detection** happens once during `LoadAssets` for each character/archetype config:

```
IF character asset dir has parts/ directory:
    config.RenderMode = RenderingModeModular
    → Load all subfolders under parts/ into config.BaseVariations
ELSE IF character's archetype has parts/ directory:
    config.RenderMode = RenderingModeModular
    → Inherit archetype's BaseVariations
ELSE:
    config.RenderMode = RenderingModeLegacy
    → Load StaticImage, BackImage, CorpseImage etc. as today (no change)
```

**Legacy Mode:** `DrawActor` draws `StaticImage` (or `BackImage`) exactly as it does today. Equipment is stats-only and has no visual representation. All existing characters (Oinakos, Stult, Marca, Queen Urraca, all archetypes) continue to work with zero changes.

**Modular Mode:** `DrawActor` uses the compositing pipeline in §5.

**Migration path:** To upgrade an existing character to Modular, create the `parts/` folder structure under their asset directory (or their archetype's). The engine will switch modes automatically on the next startup.

---

## 8. Implementation Roadmap

Phases must be completed in order. Each phase is independently testable.

| Phase | Task | Files Modified | Test Strategy |
| :--- | :--- | :--- | :--- |
| **0** | Implement `preview_composite.py` validation tool (see §9.5) and add `make preview-composite` to Makefile | `tools/asset_processor/preview_composite.py`, `Makefile` | Manual: run with test images, verify output PNG |
| **I** | Define all new types: `BodyPartPoses`, `BodyPartVariants`, `BodyPartSets`, `ActiveBodyParts`, `StatusOverlay`, `RenderingMode` | `types_config.go` | Unit test: type construction, `ActiveImage` fallback logic for all states+directions |
| **II** | Extend `EntityConfig` with `RenderMode`, `BaseVariations`; extend `Actor` with `ActiveParts`, `StatusOverlays`, `CompositedSprite`, `SpriteDirty`; extend `ObjectConfig` with `Overlays` | `registry_entity.go`, `types_config.go`, `actor.go` | Unit test: existing tests still pass (no regression) |
| **III** | Refactor `ArchetypeRegistry.LoadAssets` and `CharacterRegistry.LoadAssets` to detect `parts/`, scan subfolders, populate `BaseVariations`. Implement archetype inheritance. | `registry_entity.go` | Unit test: mock FS with `parts/` folder; verify correct variant loading and inheritance |
| **IV** | Add part-selection logic: at NPC/character spawn, read `parts:` YAML map and populate `Actor.ActiveParts` (fixed or seeded-random per zone using hashed NPC ID) | `load_logic.go`, `world_manager.go` | Unit test: YAML with fixed head; verify `ActiveParts.Head` is correct variant. Test seeded random is stable. |
| **V** | Implement Hybrid Renderer in `DrawActor`: branch on `RenderMode`, add compositing loop, apply palette shader on final buffer | `actor_render.go` | Integration test with mock images; verify layer order and shader application |
| **VI** | Implement `CompositedSprite` cache: `SpriteDirty` flag, `rebuildCompositeSprite`, invalidation triggers | `actor.go`, `actor_render.go` | Unit test: verify `SpriteDirty` is set correctly on state change, equip, trauma, overlay add/remove |

---

## 9. Asset Creation Workflow

### 9.1 Why Parts Must Be Generated Separately (Not Cropped)

> **Do NOT attempt to crop a full-body image into parts using a script.** Anatomical boundaries do not follow horizontal pixel lines. The neck is simultaneously part of the head and the torso. Any hard rectangular crop creates an unnatural cut that is visually broken.

Each body part **must be generated as a standalone, self-contained transparent image** from the start. The seam between adjacent zones is handled by the **Overlap Design Pattern** below.

### 9.2 The Overlap Design Pattern

Adjacent zones are designed so that **higher-priority layers visually cover the edge of lower-priority layers** at the seam.

How it works: Each zone draws its own content fully, with its edges fading into transparency. The zone that is drawn LATER (higher priority) naturally paints over any rough edge from the zone that was drawn EARLIER (lower priority).

| Zone | Priority | What it draws | How the seam is handled |
| :--- | :---: | :--- | :--- |
| Legs | 1 | Full legs from upper thigh to feet | Top of thighs fade to transparent. |
| Hips | 2 | Belt/waist area | Drawn AFTER legs. Hip bottom covers the leg-top transparent edge. |
| Torso | 3 | Shoulders, chest, abdomen | Drawn AFTER hips. Lower torso covers hip-top edge. Upper torso draws shoulder area. |
| Arms | 4 | Both arms from shoulder cap to wrist | Drawn AFTER torso. Shoulder caps paint over the outer torso edge. |
| Head | 5 | Head + neck | Drawn AFTER torso AND arms. The neck extends downward enough that the Torso's shoulder area (already painted) covers any rough neck-bottom edge. |

**Key insight:** Because the Head is drawn LAST (Priority 5) among base parts, the head's neck paints OVER the torso's shoulder area — not the other way around. The seam strategy is: the HEAD owns the neck, and the neck extends far enough down that it overlaps with the shoulder zone. Because the head is painted after the torso, it sits on top.

### 9.3 Generation Prompt Templates Per Zone

Each zone is generated in a **separate AI call**. All prompts share these base constraints:
```
160×160 pixels. Chroma-key green #00FF00 background.
Isometric dark-fantasy style, left-facing (SW direction).
Character feet would be at the bottom-center of the full-body 160×160 frame.
```

#### Head
```
Generate ONLY the head and neck of [description].
Head occupies the top ~40% of the frame (y=0 to ~y=64).
Neck terminates at approximately y=64 with a clean lower edge.
No shoulders or chest visible.
```

#### Torso
```
Generate ONLY the torso (shoulders, chest, abdomen) of [description].
Shoulders begin at ~y=45, waist ends at ~y=110.
Top of image (y=0 to y=45) must be fully green/transparent.
Shoulders must sit naturally as if the head is present above.
Do NOT draw the arms — they are a separate layer.
```

#### Arms
```
Generate ONLY both arms of [description], from shoulder cap to wrist.
Arms occupy x=0–55 (left) and x=105–160 (right). y=50 to y=130.
Center of image (x=55–105) must be fully green/transparent.
Shoulder cap must overlap the outer torso to avoid a visible seam.
```

#### Hips
```
Generate ONLY the hip/belt/waist area of [description].
Occupies approximately y=95 to y=130. Outside that band: green/transparent.
Sits OVER the torso lower edge and UNDER the legs upper edge.
```

#### Legs
```
Generate ONLY both legs of [description], from upper thigh to feet.
Legs begin at ~y=110. Feet at y=160 (bottom-center of frame).
Above y=110 must be fully green/transparent.
```

### 9.4 Consistency Across Variants and Poses

When generating a second variant (`v2`) or a second pose (`attack`), provide the **accepted `v1/static.png` part images as reference**:

> *"Using the attached `head_v1_static.png` as anatomical reference for position, scale, and facing direction, generate a NEW head variant with [different feature]. Keep the same positioning and frame conventions exactly."*

### 9.5 Validation After Generation

After generating any part, visually validate it by compositing it with the adjacent zones using the preview tool:

```bash
# Preview head + torso alignment
uv run tools/asset_processor/preview_composite.py \
    --character boris_stronesco \
    --variant v1_skin \
    --zones head torso

# Preview full character stack
uv run tools/asset_processor/preview_composite.py \
    --character boris_stronesco \
    --variant v1_skin \
    --all
```

**What the tool does:** Loads the specified part images from the character's `parts/` directory, composites them in the correct priority order (§5.1) onto a single 160×160 canvas, and saves the result as a PNG file at `<output_dir>/preview_<character>_<variant>.png`. If `--all` is specified, all five base zones are composited.

**Input:** Part images from the `parts/` subdirectories under the character's asset folder.
**Output:** A single 160×160 PNG showing the composited result.

If a seam is visible or a zone is misaligned, **regenerate only that specific zone** before committing. This tool must be implemented in **Phase 0** before any character migration begins.

---

## 10. Feasibility

**Status: Feasible with disciplined asset creation.**

The compositing engine itself is straightforward. The primary risk — seam visibility between body zones — is fully mitigated by:
1. **Overlap Design Pattern** (§9.2): higher-priority layers cover adjacent zone edges by design.
2. **Per-zone prompt templates** (§9.3): each zone is generated with explicit transparency bounds.
3. **Validation preview tool** (§9.5): seams are caught before assets are committed.

The additional asset discipline is offset by: combinatorial character variety (§2.3), dynamic visual events with zero new sprites (§6), and the ability to upgrade characters one-at-a-time without breaking anything (§7).
