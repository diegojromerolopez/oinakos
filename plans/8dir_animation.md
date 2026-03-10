# Plan: 8-Directional Sprites & Run Animation

**Status**: Draft  
**Difficulty**: ⭐⭐⭐⭐ Hard (significant asset, data, and rendering work)

---

## Overview

Currently every character and NPC has a single static idle sprite (`static.png`) that is
mirrored or reused for all movement directions. This plan describes what it would take to
replace that with **8 directional walk/run frames** and a **sprite-sheet animation loop**
while the entity is moving.

---

## Industry Precedent

Understanding how 1990s isometric RPGs solved this problem informs the recommended
asset pipeline.

### Baldur's Gate (1998) & Diablo (1996)

Both games used **pre-rendered 3D sprites**: characters were rigged and animated in 3D
software (3D Studio Max), then **rendered to flat 2D images** from fixed isometric camera
angles. This meant:

- Author a walk cycle **once** in 3D — re-render at any angle for free.
- **5 unique directions rendered** (SE, S, SW, W, NW); the other 3 are runtime mirrors.
- Sprites stored in packed formats (BAM for BG, DC6 for Diablo) with palette compression.
- A single character like Minsc had ~600 sprites (5 dirs × 12 states × ~10 frames).

### Why switching to a full 3D engine is not the answer

Switching Oinakos to a real-time 3D engine would be a **complete rewrite**:
- Ebiten is a 2D library — no scene graph, no depth buffer, no 3D mesh rendering.
- All of `internal/engine`, the isometric math, Y-sorting, and current sprites would be
  discarded.
- A replacement would target Godot, Unity, or Raylib — a **3–6 month** effort minimum.

### Recommended asset strategy: Blender pre-render pipeline

The pragmatic middle ground — exactly what Baldur's Gate did — is:

1. Model each character in **Blender** (free, open source).
2. Set up an isometric camera (≈30° elevation, 45° azimuth).
3. Author one walk cycle (~8 keyframes) on the rig.
4. Batch render a **sprite strip for each of the 5 directions** automatically.
5. Post-process: resize to 160×160, set background to `#00FF00`.

This produces **pixel-perfect, spatially consistent frames** that individual AI image
generation cannot reliably match. Any future animation state (attack, death) is just a new
action on the existing rig — not a new set of image requests.

**Estimated setup time per character**: 3–5 days (modelling + rigging) + 1 day (render pipeline).
**Subsequent animation states**: 2–4 hours each (keyframe new action, re-render, done).

---

## Scope

| Feature | In scope |
| :--- | :--- |
| 8 cardinal/intercardinal walk directions (N, NE, E, SE, S, SW, W, NW) | ✅ |
| Multi-frame animation loop for each direction (simulating running) | ✅ |
| Idle fallback to existing `static.png` when standing still | ✅ |
| Attack and hit directional sprites | ❌ (future, not in this plan) |
| True sprite-sheet (single PNG with frames on a grid) vs. individual PNGs | discussed below |

---

## Layer 1 — Assets (Hardness: ⭐⭐⭐⭐⭐ Very Hard)

This is the **dominant cost** of the entire feature.

### What is needed

For every character and NPC archetype, we need a **walk animation** for all 8 directions.
A minimal implementation with **4 frames per direction** gives:

```
8 directions × 4 frames = 32 sprites per character
```

We currently have **8 playable characters + 11 archetypes (×2 genders) + 9 unique NPCs**
≈ **36 distinct entities**.

```
36 entities × 32 frames = 1,152 individual 160×160 sprite images
```

At roughly **30 KB per PNG** that is ~**35 MB** of new image data added to the binary
(before any WASM size concerns).

### Isometric direction notes

In isometric view, the camera is fixed at a NE–SW diagonal. The 8 logical directions map to
visually distinct angles:

| Logical Dir | Isometric appearance | Notes |
| :--- | :--- | :--- |
| SE | directly toward viewer (right-front) | easiest to draw |
| SW | toward viewer (left-front) | mirror of SE in theory |
| NE | away from viewer (right-back) | often simplified |
| NW | away from viewer (left-back) | mirror of NE |
| E / W | pure right / left | odd diagonal in iso |
| N / S | pure away / toward | least common in classic iso |

In practice, many classic isometric RPGs (Diablo included) **only ship 5 directions**
(SE, S, SW, W, NW) and **mirror** the other 3 (NE mirrors NW, N mirrors S, E mirrors W).
This halves the asset cost:

```
5 directions × 4 frames = 20 sprites per entity (+ mirroring at runtime)
```

### Sprite-sheet vs. individual files

| Approach | Pros | Cons |
| :--- | :--- | :--- |
| **Individual PNGs per frame** | Simple to author, easy to load | 1,152 files, filesystem clutter |
| **Single sprite-sheet per direction** | One file per dir, standard industry approach | Requires sheet-slicing logic in engine |
| **Single mega-sheet per entity** | Minimal files | Complex to address, harder to author |

**Recommendation**: one sprite-sheet per entity per direction, e.g.  
`assets/images/characters/oinakos/walk_se.png` — a horizontal strip of 4 frames.

### Asset generation difficulty

- AI-generated images must be **spatially consistent across frames** (same pose proportions,
  same lighting, same clothing) — currently each image is generated independently, which
  makes frame consistency very difficult without fine-tuned models or manual cleanup.
- Each frame must be **pixel-level aligned** so movement doesn't "jitter" between frames.
- This is the reason the roadmap item says "sprite-sheet support" not "sprite-sheet done".

**Estimated effort**: 2–4 weeks of asset work per character if generated and cleaned manually.
With a consistent AI pipeline (e.g. ControlNet pose conditioning): perhaps 3–5 days per
character but still requiring manual cleanup.

---

## Layer 2 — Data / YAML (Hardness: ⭐⭐ Easy)

The `EntityConfig` struct would need two new fields:

```go
WalkFrameCount int    `yaml:"walk_frame_count"` // default: 4
WalkFPS        int    `yaml:"walk_fps"`          // default: 8 (frames per second)
```

Example YAML:
```yaml
walk_frame_count: 4
walk_fps: 8
```

No breaking changes to existing YAML files — these fields would be purely optional with
sensible defaults. Entities without them fall back to the current `static.png` idle.

**Estimated effort**: 1–2 hours.

---

## Layer 3 — Engine / Asset Loading (Hardness: ⭐⭐⭐ Medium)

### New field on `EntityConfig`

```go
// One entry per direction (8 max), each is a slice of frame images
WalkImages [8][]engine.Image `yaml:"-"`
```

### Loading logic (in `ArchetypeRegistry.LoadAssets` / `PlayableCharacterRegistry.LoadAssets`)

For each of the 8 directions (or 5 if using mirroring), attempt to load:
```
assets/images/<kind>/<id>/walk_se.png   → strip of N frames of width 160×N, height 160
```

The loader would slice the strip horizontally into `walk_frame_count` frames of 160×160 each.

Ebiten supports sub-image slicing natively via `image.SubImage`, so no extra dependencies
are needed.

### Mirroring

Ebiten's `DrawImageOptions.GeoM.Scale(-1, 1)` flips horizontally for free at render time.
NE can mirror NW, E can mirror W, N can mirror S — at zero additional asset cost.

**Estimated effort**: 1–2 days.

---

## Layer 4 — Direction Enum (Hardness: ⭐ Trivial)

Currently the `Direction` enum has 4 values:

```go
const (
    DirSE Direction = iota
    DirSW
    DirNE
    DirNW
)
```

Expanding to 8:

```go
const (
    DirSE Direction = iota // 0
    DirS                   // 1
    DirSW                  // 2
    DirW                   // 3
    DirNW                  // 4
    DirN                   // 5
    DirNE                  // 6
    DirE                   // 7
)
```

All facing-assignment logic in `playable_character.go` and `npc.go` currently uses 4-way checks
(`dx > 0`, `dx < 0`, `dy > 0`, `dy < 0`). Expanding to 8 requires using the **angle** of the
movement vector:

```go
angle := math.Atan2(dy, dx) // radians
// bucket into 8 × 45° sectors
dir := Direction(int((angle + math.Pi + math.Pi/8) / (math.Pi / 4)) % 8)
```

**Estimated effort**: 2–3 hours.

---

## Layer 5 — Animation Tick (Hardness: ⭐⭐ Easy)

A simple frame counter on each entity:

```go
type PlayableCharacter struct {
    ...
    AnimFrame int // current walk frame index
    AnimTick  int // ticks since last frame advance
}
```

In `Update()`, while `StateWalking`:
```go
pc.AnimTick++
if pc.AnimTick >= (60 / pc.Config.WalkFPS) { // 60 TPS / FPS = ticks per frame
    pc.AnimTick = 0
    pc.AnimFrame = (pc.AnimFrame + 1) % pc.Config.WalkFrameCount
}
```

When `StateIdle`, reset `AnimFrame = 0` and show `static.png`.

NPCs follow the same pattern in their `Update()`.

**Estimated effort**: 2–4 hours.

---

## Layer 6 — Renderer (Hardness: ⭐⭐ Easy)

Currently `game_render.go` uses a single `config.StaticImage` for movement.

New selection logic:
```go
func spriteForEntity(config *EntityConfig, state State, facing Direction, frame int) engine.Image {
    if state == StateWalking && config.WalkImages[facing] != nil {
        frames := config.WalkImages[facing]
        return frames[frame % len(frames)]
    }
    return config.StaticImage // fallback
}
```

This is purely additive — existing behaviour is preserved for entities without walk sprites.

**Estimated effort**: 2–4 hours.

---

## Summary

| Layer | Task | Effort | Difficulty |
| :--- | :--- | :--- | :--- |
| Assets | Generate & clean 8-dir walk strips | **weeks** | ⭐⭐⭐⭐⭐ |
| Data/YAML | Add `walk_frame_count`, `walk_fps` | 1–2 h | ⭐⭐ |
| Asset loading | Strip slicing + mirror logic | 1–2 days | ⭐⭐⭐ |
| Direction enum | Expand 4→8 directions, angle math | 2–3 h | ⭐ |
| Animation tick | Frame counter in Update() | 2–4 h | ⭐⭐ |
| Renderer | Directional frame selection | 2–4 h | ⭐⭐ |

> **The code changes are straightforward (~3–4 days total). The bottleneck is entirely
> asset production.** Generating 1,152 spatially consistent 160×160 sprites with current
> AI tooling is the hard part, not the Go code.

---

## Recommended Rollout Strategy

1. **Code first, assets second**: Implement all 6 code layers using a simple hand-crafted
   test strip (e.g. a coloured rectangle sliding across frames) to validate the pipeline
   end-to-end before producing real art.

2. **Ship behind a feature flag**: `walk_frame_count: 0` is the default — no entity is
   affected until it explicitly opts in with real assets.

3. **Mirror first**: Implement horizontal mirroring via `GeoM.Scale(-1, 1)` before producing
   any assets — this immediately reduces the required directions from 8 to 5, halving the
   asset workload.

4. **Blender pipeline first, one character**: Set up the Blender render pipeline for
   **Oinakos** first (as the lead character). This produces the template camera, lighting,
   and post-process chain that all subsequent characters will reuse.
   - Isometric camera: ≈30° elevation, 45° azimuth
   - Render output: horizontal sprite strip, 4 frames wide, 160×160 per frame
   - Background: solid `#00FF00` (Ebiten chroma-key)
   - Automate the 5-direction batch render with a short Blender Python script

5. **Roll out entity by entity**: Once the pipeline exists, each new character is a matter
   of importing a new mesh, re-using the rig, and running the batch render script.
   Estimated **1–2 days per new character** once the pipeline is established.
