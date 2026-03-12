# Plan: Permanent Physical Disfigurement & Trauma 🩸

## Objective
To implement a system of irreversible physical injuries that profoundly alter the gameplay experience, survivability, and visual identity of characters in *Oinakos*.

## Analysis
Permanent damage must go beyond temporary HP loss. It represents a fundamental shift in the character's capabilities. Every injury should carry a unique mechanical penalty and a distinct visual change to the character's sprite or the rendering of the world.

### Key Injury Types:
1.  **Amputation (Arm/Hand/Leg)**: Loss of limbs causing mobility or combat equipment restrictions.
2.  **Sensory Loss (Eyes)**: Permanent reduction in vision range or perspective.
3.  **Chronic Trauma (Burned Alive)**: Constant physical pain and severe health reduction.
4.  **Structural Collapse (Broken Spine)**: Near-total loss of mobility and heavy weapon capability.

---

## Implementation Details

### 1. The Trauma Registry (`internal/game/actor.go`)
Define a struct to track the physical state of every `Actor`:

```go
type PhysicalTrauma struct {
    LeftArmLost   bool
    RightArmLost  bool
    LeftLegLost    bool
    RightLegLost   bool
    LeftHandLost   bool
    RightHandLost  bool
    EyesLost      int  // 0, 1, or 2
    BurnedAlive   bool // Survivors of extreme fire
    SpineBroken   bool
}
```

Add this `Trauma` field to the `Actor` struct.

### 2. Mechanical Penalties & "The Death Spiral"
Injuries will apply cumulative modifiers in `Actor.Update()` and `Actor.TakeDamage()`:

| Injury | Mechanical Penalty | Health Impact |
| :--- | :--- | :--- |
| **Leg Loss** | `-50% Speed` (per leg). Crawl at 2-legs lost. | `-10 MaxHealth` (blood loss/trauma). |
| **Arm Loss** | Cannot equip items in that slot. `-30% Attack Speed`. | `-5 MaxHealth`. |
| **Hand Loss** | Can hold items but cannot use `Bow` or `Two-Handed`. | `-3 MaxHealth`. |
| **One Eye Lost** | `-30% Sight Range`. Screen "Blind Spot" overlay. | `-2 MaxHealth`. |
| **Blind (2 Eyes)** | Screen nearly black. Rely on audio/prox-icons. | `-5 MaxHealth`. |
| **Burned Alive** | `Continuous Pain`: `-1 HP every 600 ticks` permanently. | `-30 MaxHealth`. |
| **Broken Spine** | `-80% Speed`. No running. Cannot use `High-Weight` weapons. | `-20 MaxHealth`. |

### 3. Visual Transformations (`internal/game/game_render.go`)
Characters must look as broken as they are:

- **Folder Structure**: Character assets are organized into `default/` and `modified/` subfolders:
  ```
  assets/images/characters/<id>/
    ├── default/           # Original full-body sprites
    │   ├── static.png
    │   ├── attack1.png
    │   └── ...
    └── modified/          # Variations based on trauma
        ├── legless/       # Strategic amputation set
        ├── left-arm-lost/
        └── bleeding/      # Visual status overlays
  ```
- **Image Generation Strategy**: To ensure consistency across animation frames:
  1. **Master Reference**: Generate a `static.png` for the specific trauma using the original as a structural base (Image-to-Image).
  2. **Batch Refinement**: Once the "look" is approved, generate all other frames (`attack`, `hit`, `corpse`) using the original frames as structural guides to maintain pose and perspective.
- **Loading Logic**: The engine must check for the `default/` subfolder. If missing, it falls back to loading directly from the character ID folder for backwards compatibility.
- **Dynamic Swapping**: When an injury occurs, the `AssetDir` (or a specific `ActiveTraumaDir`) is updated to point to the relevant `modified/` subfolder.
- **Burning FX**: A "Charred" shader or overlay applied to the actor's sprite, turning skin tones to black/ash while keeping the primary/secondary colors muted.
- **Spine Pose**: If `SpineBroken` is true, force the character into the `Crouch` or `Limping` animation frame regardless of speed.
- **Vision Shaders**: If eyes are lost, apply a vignette or "half-screen" black-out shader to the `screen` image during the final render pass.

### 4. Acquisition Mechanics
Permanent damage should be rare but catastrophic:
- **Critical Fails**: A 1% chance when taking heavy damage from specific weapons (e.g., axes for limbs, fire for burns).
- **Execution Failures**: If an NPC attempts a "Fatality" on a player but is interrupted, the player survives but is maimed.
- **Traps**: Specific environmental hazards (e.g., "The Iron Maiden" or "The Fire Pit") that explicitly mention permanent risks.

---

## Verification
- **Test Case: The Veteran**: Spawn a character with `LeftArmLost` and `LeftLegLost`. Verify they move slowly and cannot equip a shield.
- **Test Case: The Burned King**: Set `BurnedAlive = true`. Observe the health slowly tick down every 10 seconds and the charred visual overlay on the sprite.
- **Test Case: Total Blindness**: Set `EyesLost = 2`. Verify the game screen becomes a void of black with only the HUD and minimal light circles visible.
- **Test Case: The Survivor**: Break a character's spine. Verify they can no longer reach "Walking" speed and the `MaxHealth` bar is significantly shortened.
