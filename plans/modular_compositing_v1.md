# Plan: Modular Compositing v1 — Head + Body (Paper Doll Lite)

> **Status:** Ready for implementation.
> **Scope:** Two-zone compositing (Head + Body). Zero changes to existing Legacy characters.
> **Future:** v2 will split Body into Torso/Arms/Hips/Legs. See `modular_compositing.md` for the full vision.

---

## 1. What v1 Delivers

With only **two zones** (Head and Body), v1 provides:

| Feature | How |
| :--- | :--- |
| NPC variety (different faces) | Swap Head variants: `v1` (young), `v2` (old), `v3` (scarred) |
| Visible helmets | Head armor overlay drawn on top of Head zone |
| Visible body armor | Body armor overlay drawn on top of Body zone |
| Prisoner stripping | Unequip all overlay slots → naked base shows |
| Flaying | Swap Body to `_flayed` variant |
| Armor breakage | Remove body armor overlay → naked body shows |

What v1 does NOT do (deferred to v2):
- Individual arm/leg amputation visuals (trauma shader still handles this)
- Per-limb armor pieces (e.g. separate gauntlets vs breastplate)

---

## 2. The Two Zones

### Zone 1: Body
- Covers **everything from the shoulders to the feet** (torso, arms, hips, legs).
- Drawn FIRST (behind the Head).
- The naked base image shows a fully naked body from shoulders down.

### Zone 2: Head
- Covers the **head and neck**.
- Drawn SECOND (in front of the Body).
- The neck extends down far enough to overlap the Body's shoulder area. Because the Head is drawn after the Body, the overlap is invisible — the head naturally sits on top.

This is the **only seam** in the system, and it's the easiest one: neck-on-shoulders.

---

## 3. Layer Compositing Order (Back to Front)

All layers are drawn at the same `(isoX, isoY)` screen coordinate into a `CompositedSprite` cache buffer.

| Priority | Layer | Source | Condition |
| :---: | :--- | :--- | :--- |
| 1 | Naked Base — Body | `Actor.ActiveParts.Body` | Always |
| 2 | Naked Base — Head | `Actor.ActiveParts.Head` | Always |
| 3 | Undergarment Overlay | `Overlays["body"]` from item in `"undergarment"` slot | Slot occupied |
| 4 | Clothing Overlay | `Overlays["body"]` from item in `"clothing"` slot | Slot occupied |
| 5 | Body Armor Overlay | `Overlays["body"]` from item in `"body_armor"` slot | Slot occupied |
| 6 | Head Armor Overlay | `Overlays["head"]` from item in `"head_armor"` slot | Slot occupied |
| 7 | Weapon Overlay | `Overlays["weapon"]` from item in `"weapon"` slot | Slot occupied |
| 8 | Status Overlays | `Actor.StatusOverlays[]` images (chains, blood, etc.) | List not empty |

> **Overlay resolution:** Each equipped item's `Overlays` map is checked for the zone key. If the item has no `Overlays` map, it is stats-only (Legacy behaviour).

---

## 4. Asset Directory Layout

```
assets/images/
├── archetypes/
│   └── human/
│       └── male/
│           ├── static.png         ← Legacy fallback (used if no parts/ folder)
│           └── parts/             ← Presence enables Modular mode
│               ├── heads/
│               │   ├── v1/
│               │   │   ├── static.png   ← REQUIRED (idle, SW/SE)
│               │   │   ├── back.png     ← Optional (idle, NW/NE)
│               │   │   ├── attack.png   ← Optional
│               │   │   ├── crouch.png   ← Optional
│               │   │   └── corpse.png   ← Optional
│               │   ├── v2/
│               │   │   └── static.png
│               │   └── v3/
│               │       └── static.png
│               └── bodies/
│                   ├── v1_skin/
│                   │   ├── static.png
│                   │   ├── back.png
│                   │   ├── attack.png
│                   │   └── corpse.png
│                   └── v1_flayed/
│                       └── static.png
│
├── characters/
│   └── boris_stronesco/
│       ├── static.png             ← Legacy fallback
│       └── parts/                 ← Character-local (overrides archetype parts)
│           ├── heads/
│           │   └── v1/
│           │       └── static.png
│           └── bodies/
│               └── v1_skin/
│                   └── static.png
│
└── objects/
    ├── iron_plate_armor/
    │   └── overlays/
    │       └── body/              ← 160×160 breastplate+legs overlay
    │           ├── static.png
    │           └── corpse.png
    └── iron_helmet/
        └── overlays/
            └── head/              ← 160×160 helmet overlay
                └── static.png
```

**Rules:**
- `parts/` folder presence = Modular mode. Absence = Legacy mode.
- Both `heads/` and `bodies/` must have at least one variant with `static.png`.
- Object `overlays/` subfolders are named `head/` or `body/`.

---

## 5. Data Structures (Go)

### 5.1 New Types (add to `types_config.go`)

```go
// BodyPartPoses holds pose-specific images for one variant.
// Only Static is required; all others fall back to Static if nil.
type BodyPartPoses struct {
    Static engine.Image
    Back   engine.Image
    Attack engine.Image
    Crouch engine.Image
    Corpse engine.Image
}

// ActiveImage returns the correct image for a given state and facing direction.
// Action states (Attack/Crouch/Corpse) ignore facingBack.
func (p *BodyPartPoses) ActiveImage(state ActorState, facingBack bool) engine.Image {
    if p == nil { return nil }
    switch state {
    case ActorDead, ActorIncapacitated:
        if p.Corpse != nil { return p.Corpse }
    case ActorAttacking:
        if p.Attack != nil { return p.Attack }
    case ActorCrouching:
        if p.Crouch != nil { return p.Crouch }
    default:
        if facingBack && p.Back != nil { return p.Back }
    }
    return p.Static
}

// BodyPartVariants holds all loaded variants for one zone.
type BodyPartVariants struct {
    Variants map[string]BodyPartPoses // Key: subfolder name (e.g. "v1", "v1_flayed")
    Keys     []string                 // Sorted keys for stable random indexing
}

// TwoZoneParts holds all variant options for the two zones.
type TwoZoneParts struct {
    Heads  BodyPartVariants
    Bodies BodyPartVariants
}

// ActiveTwoZoneParts holds the selected variant for each zone for one actor.
type ActiveTwoZoneParts struct {
    Head    BodyPartPoses
    Body    BodyPartPoses
    HeadKey string // e.g. "v2" — stored for anatomy swap lookups and save/load
    BodyKey string // e.g. "v1_skin"
}

// StatusOverlay is a temporary visual effect drawn at the highest priority.
type StatusOverlay struct {
    ID    string       // e.g. "neck_chain", "blood_splatter"
    Image engine.Image // 160×160 transparent image
}

// RenderingMode determines which path DrawActor uses.
type RenderingMode int

const (
    RenderingModeLegacy  RenderingMode = iota
    RenderingModeModular
)
```

### 5.2 EntityConfig Changes

```go
type EntityConfig struct {
    // ── All existing fields unchanged ──

    RenderMode     RenderingMode `yaml:"-"`
    BaseVariations TwoZoneParts  `yaml:"-"` // Loaded from parts/
}
```

### 5.3 Actor Changes

```go
type Actor struct {
    // ── All existing fields unchanged ──

    ActiveParts      ActiveTwoZoneParts `yaml:"-"`
    StatusOverlays   []StatusOverlay    `yaml:"-"`
    CompositedSprite engine.Image       `yaml:"-"` // 160×160 cache buffer
    SpriteDirty      bool               `yaml:"-"`
}
```

### 5.4 ObjectConfig Changes

```go
type ObjectConfig struct {
    // ── All existing fields unchanged ──

    // Key: zone name ("head" or "body"). Value: pose frames for that overlay.
    Overlays map[string]BodyPartPoses `yaml:"-"`
}
```

### 5.5 New Slot Keys

| Slot Key | Priority | Purpose |
| :--- | :---: | :--- |
| `"undergarment"` | 3 | Loincloth, underwear — **new** |
| `"clothing"` | 4 | Tunic, pants — **new** |
| `"body_armor"` | 5 | Breastplate, chainmail — **new** |
| `"head_armor"` | 6 | Helmet, crown — **new** |
| `"weapon"` | 7 | Sword, axe — **existing**, now also reads `Overlays["weapon"]` |
| `"shield"` | — | Stats-only for now — **existing** |
| `"ring1"`, `"ring2"` | — | Stats-only — **existing** |

> **Backward compatibility:** Old items using `slot: "body"` or `slot: "head"` will continue to work for stats. They will only get visual overlays if their `ObjectConfig.Overlays` map is populated (i.e. they have an `overlays/` asset folder).

---

## 6. YAML Configuration

### Character YAML
```yaml
id: boris_stronesco
name: "Boris Stronesco"
# parts: section is OPTIONAL. Omit → Legacy mode.
parts:
  head: "v2"         # Force variant. Omit or "random" → seeded-random from NPC ID hash.
  body: "v1_skin"    # Force variant.
```

### Equipment YAML
```yaml
id: iron_plate_armor
name: "Iron Plate Armor"
slot: "body_armor"
effects:
  protection:
    increase: 15
overlays:
  body: true     # Loader scans assets/images/objects/iron_plate_armor/overlays/body/
```

```yaml
id: iron_helmet
name: "Iron Helmet"
slot: "head_armor"
overlays:
  head: true     # Loader scans assets/images/objects/iron_helmet/overlays/head/
```

---

## 7. Rendering Logic

### 7.1 Pose Frame Selection

| Actor State | `facingBack` = true (NW/NE) | `facingBack` = false (SW/SE) |
| :--- | :--- | :--- |
| Idle, Walking, Drinking | `Back` → fallback `Static` | `Static` |
| Attacking | `Attack` → fallback `Static` | `Attack` → fallback `Static` |
| Crouching | `Crouch` → fallback `Static` | `Crouch` → fallback `Static` |
| Dead, Incapacitated | `Corpse` → fallback `Static` | `Corpse` → fallback `Static` |

Horizontal flip: `DirSE` or `DirNE` → `flip = -1.0`. Same as Legacy.

### 7.2 Palette Swap Shader

The existing palette shader (`paletteSwapShaderSource`) is applied **once** to the final `CompositedSprite` when drawing to screen. It is NOT applied per-layer. No shader changes needed.

```
if actor.Config.RenderMode == RenderingModeModular {
    if actor.SpriteDirty || actor.CompositedSprite == nil {
        rebuildCompositeSprite(actor)  // draws 8 layers into buffer
        actor.SpriteDirty = false
    }
    // Apply shader on final composited buffer (same as Legacy applies to StaticImage)
    drawWithShader(screen, actor.CompositedSprite, paletteShader, uniforms, op)
} else {
    // Legacy path — completely unchanged
    drawWithShader(screen, DrawActorGetSprite(actor), paletteShader, uniforms, op)
}
```

### 7.3 SpriteDirty Triggers

Set `SpriteDirty = true` when:
- Any equipment slot changes.
- `Actor.State` changes to a different value.
- Any `PhysicalTrauma` flag changes.
- `ActiveParts` is modified (anatomy swap).
- A `StatusOverlay` is added or removed.

### 7.4 Archetype Inheritance

```
IF character has parts/ folder → use character's parts. Modular mode.
ELSE IF archetype has parts/ folder → inherit archetype's BaseVariations. Modular mode.
ELSE → Legacy mode. No change.
```

---

## 8. Dynamic Visual Events

### 8.1 Prisoner / Capture
1. Unequip items from: `undergarment`, `clothing`, `body_armor`, `head_armor`.
2. Add `StatusOverlay{ID: "chains", Image: chains_image}` to `Actor.StatusOverlays`.
3. `SpriteDirty = true`.

### 8.2 Flaying
1. Derive flayed key: `BodyKey "v1_skin"` → look for `"v1_flayed"` in `BaseVariations.Bodies`.
2. If found: swap `ActiveParts.Body` and `ActiveParts.BodyKey`.
3. `SpriteDirty = true`.

### 8.3 Armor Breakage
1. Unequip the broken item from its slot.
2. `SpriteDirty = true`.

---

## 9. Implementation Roadmap

| Phase | Task | Files | Test |
| :--- | :--- | :--- | :--- |
| **0** | Create `preview_composite.py` tool + `make preview-composite` | `tools/asset_processor/`, `Makefile` | Manual preview of test images |
| **I** | Define new types: `BodyPartPoses`, `BodyPartVariants`, `TwoZoneParts`, `ActiveTwoZoneParts`, `StatusOverlay`, `RenderingMode` | `types_config.go` | Unit test: `ActiveImage` fallback for all states |
| **II** | Extend `EntityConfig`, `Actor`, `ObjectConfig` with new fields | `registry_entity.go`, `actor.go`, `types_config.go` | All existing tests pass |
| **III** | Detect `parts/` in `LoadAssets`, scan `heads/` + `bodies/`, populate `BaseVariations`. Archetype inheritance. | `registry_entity.go` | Mock FS test |
| **IV** | Part selection at spawn: read YAML `parts:` or seeded-random (hash NPC ID) | `load_logic.go`, `world_manager.go` | Verify deterministic + random selection |
| **V** | Hybrid `DrawActor`: branch on `RenderMode`, composite 8 layers into buffer, apply shader once | `actor_render.go` | Mock image integration test |
| **VI** | `SpriteDirty` invalidation on all triggers | `actor.go`, `actor_render.go` | Unit test each trigger |

---

## 10. Asset Creation Workflow

### 10.1 Parts Must Be Generated Separately

> **Do NOT crop a full-body image.** Generate each zone as its own AI call.

### 10.2 The Seam (Head on Body)

There is exactly **one seam**: where the Head's neck meets the Body's shoulders.

- The **Body** image draws shoulders and neck-base area.
- The **Head** image draws head + neck extending down to overlap the shoulder area.
- Because Head is drawn AFTER Body (Priority 2 vs 1), the head naturally covers any rough edge.

This is the simplest possible seam. If the neck extends to ~y=70 on the head image and the shoulders start at ~y=45 on the body image, there is a 25px overlap zone that hides any mismatch.

### 10.3 Prompt Templates

**Body zone:**
```
Generate a full naked body (shoulders to feet) of [description].
160×160 pixels, chroma-key green #00FF00 background.
Isometric dark-fantasy, left-facing (SW).
Feet at bottom-center. Head is NOT included — cut off at the neck/shoulder line (~y=45).
Above y=45 must be fully green/transparent.
```

**Head zone:**
```
Generate ONLY the head and neck of [description].
160×160 pixels, chroma-key green #00FF00 background.
Isometric dark-fantasy, left-facing (SW).
Head in the upper portion (y=0 to ~y=70).
Neck extends to approximately y=70r to overlap with shoulder area.
Below y=70 must be fully green/transparent.
No shoulders or chest visible.
```

### 10.4 Validation
```bash
uv run tools/asset_processor/preview_composite.py \
    --character boris_stronesco --variant v1_skin --all
```

---

## 11. Memory Budget

| Item | Count | VRAM |
| :--- | :--- | :--- |
| 5 archetypes × 2 zones × 5 variants × 2 poses | 100 images | ~10 MB |
| 10 objects × 1 zone × 2 poses | 20 images | ~2 MB |
| 50 on-screen actors × 1 cache buffer | 50 images | ~5 MB |
| **Total** | | **~17 MB** |

This is negligible. No lazy loading or optimization needed for v1.

---

## 12. v1 → v2 Migration Path

When v1 is stable and the asset pipeline is proven, v2 splits the `Body` zone into 4 sub-zones (`Torso`, `Arms`, `Hips`, `Legs`):

- `TwoZoneParts` becomes `FiveZoneParts` (or a generic `map[string]BodyPartVariants`).
- Layer priorities 1–2 expand to 1–5.
- Existing `Body` variants become `Torso` variants (the broadest zone).
- Arms, Hips, Legs zones are added incrementally.
- All v1 overlay assets remain valid.

The architecture supports this because the compositing loop is already zone-agnostic — adding zones only means adding entries to the loop, not rewriting it.
