# Character Movement and Animation Juice Plan 🎭✨

This plan outlines how to achieve a "Darkest Dungeon" style feel for NPCs and players—providing weight, impact, and a sense of life—without resorting to complex skeletal rigs or paperdoll systems.

---

## 🎨 Design Philosophy: "Juice" over Frames
Instead of drawing hundreds of unique frames, we will use **Procedural Squash, Stretch, and Secondary Motion**. Darkest Dungeon uses 2D deformation to make static-ish images feel alive.

---

## 🏗️ Technical Specifications

### 1. Dynamic Squash & Stretch
We will apply sine-wave based scaling to the actor's sprite based on their state.

- **Idling (Breathing)**: Subtle vertical scaling (Y-axis) oscillating at a slow frequency (e.g., 2-3 seconds).
- **Walking (The "Hop")**: Faster vertical scaling + slight lateral tilt. The character "squashes" when their foot hits the ground and "stretches" at the peak of their step.
- **Attacking (The "Lunge")**: Significant lateral stretch on the thrust, immediately followed by a "snap back" to idle.

### 2. Secondary Motion (Bobbing/Floating)
- **Floating**: Non-terrestrial entities (like ghosts or floating eyes) will have a periodic vertical offset added to their `Y` coordinate.
- **Tilting**: When changing direction, the actor will subtly "lean" into the direction of movement for several frames.

---

## 🖼️ Sprite Standards

Since we are avoiding paperdolls, we will use **Mini-Sprite Sheets**:

1. **Standard Set (3 frames)**:
   - `static.png`: Base pose.
   - `action_1.png`: Slight variation for "preparing".
   - `action_2.png`: Peak of an action.
2. **Standardization**:
   - All sprites should be centered on the **bottom-middle** (the base of the character's feet).
   - Standard size remains **160x160 pixels**.

---

## 🛠️ Implementation Workflow

### Step 1: Geometry Deform Module
Add helper functions to `internal/game/actor_render.go` to calculate dynamic `GeoM` transforms.

```go
func getJuiceScale(a *Actor) (sx, sy float64) {
    if a.State == ActorIdle {
        sy = 1.0 + math.Sin(float64(a.Tick)*0.05)*0.02 // Breathe
        sx = 1.0 / sy // Maintain volume
    }
    // ... logic for walk/hit
    return sx, sy
}
```

### Step 2: Impact Effects (Darkest Dungeon Style)
- **Screen Shake**: When the player takes a heavy hit.
- **Freeze Frame**: On a killing blow, pause the entity's animation for 2-3 frames to provide "hit stop" impact.
- **Damage Numbers**: Refine `FloatingText` with a slight "pop" scale animation (starting large and shrinking to target size).

### Step 3: Groundedness (Shadows)
Every actor needs a procedural shadow:
- A simple semi-transparent black ellipse drawn at `(a.X, a.Y)`.
- The shadow's scale changes with the actor's procedural vertical bobbing (smaller/lighter when jumping, larger/darker when landing).

---

## 📅 Roadmap

### Phase 1: Procedural Juice
- [ ] Implement vertical breathing oscillation for `ActorIdle`.
- [ ] Implement the "Step-Hop" squash/stretch for `ActorWalking`.
- [ ] Add the "Hit-Shake" (lateral jitter) for `a.HitTimer > 0`.

### Phase 2: Sprite Variation
- [ ] Update `ArchetypeRegistry` to support matching image sequences (e.g., `walk_0.png`, `walk_1.png`).
- [ ] Implement a simple frame cycler based on `a.Tick % frameCount`.

### Phase 3: Secondary FX
- [ ] Implement procedural shadows that react to bobbing.
- [ ] Add "Hit-Stop" (time dilation for specific entities).
- [ ] Implement "Red-Flash" shader effect when taking damage.
