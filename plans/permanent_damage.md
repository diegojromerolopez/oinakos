# Plan: Permanent Physical Disfigurement & Trauma 🩸

## Objective
To implement a system of irreversible physical injuries that profoundly alter the gameplay experience, survivability, and visual identity of characters in *Oinakos*.

## Analysis
Permanent damage representation goes beyond temporary HP loss. It represents a fundamental shift in the character's capabilities. Every injury should carry a unique mechanical penalty and a distinct visual change.

### Key Injury Types:
1.  **Amputation (Arm/Leg)**: Loss of limbs causing mobility or equipment restrictions.
2.  **Sensory Loss (Eyes)**: Permanent reduction in perspective or vision range.
3.  **Chronic Trauma (Burned Alive)**: Constant physical pain and severe health reduction.
4.  **Structural Collapse (Broken Spine)**: Near-total loss of mobility.

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
    EyesLost      int  // 0, 1, or 2
    BurnedAlive   bool // Survivors of extreme fire
    SpineBroken   bool
}
```

### 2. Mechanical Penalties & "The Death Spiral"
| Injury | Mechanical Penalty | Health Impact |
| :--- | :--- | :--- |
| **Leg Loss** | `-50% Speed` (per leg). Crawl at 2-legs lost (10% speed). | `-10 MaxHealth`. |
| **Arm Loss** | Cannot equip items in that slot. `-30% Attack Speed`. | `-5 MaxHealth`. |
| **One Eye Lost** | `-30% Attack Bonus`. | `-2 MaxHealth`. |
| **Burned Alive** | `Continuous Pain`: `-1 HP every 600 ticks` permanently. | `-30 MaxHealth`. |
| **Broken Spine** | `-80% Speed`. No running. | `-20 MaxHealth`. |

### 3. Visual Transformations (Non-Destructive Shader)
Instead of generating thousands of new PNGs, we use the **Universal Actor Shader**:
- **Geometric Clipping**: Shader discards pixels based on sprite coordinates (160x160) for limbs and eyes.
- **Charred Effect**: `BurnedAlive` triggers a global darkening and shadow "ember" glow.
- **Incapacitated State**: At 0 HP, characters use their `corpse.png` but remain in the world (not yet truly dead).

### 4. Acquisition & Bleed-out
- **Deterministic Trauma**: Every hit taken while `< 10% Health` results in a new trauma.
- **Incapacitated State**: At 0 HP, characters are downed but still "alive."
- **Bleed-out**: While incapacitated, characters lose **1 HP every hour** (3600 ticks / 1 minute real-time).
- **True Death**: Irremediable death only occurs when health drops to **-10% of MaxHealth**.
- **Heavy Hits**: Any hit `> 15% MaxHealth` has a 1% base trauma risk.

---

## Verification
- **Test Case: Incapacitation**: Damage an actor to 0. Verify they enter `ActorIncapacitated` state and use the corpse sprite but are still "there."
- **Test Case: Bleed-out**: Leave an incapacitated actor alone. Verify their HP drops to -1, -2, etc., every minute.
- **Test Case: Final Death**: Continue damaging an incapacitated actor until they hit -10% MaxHealth. Verify they trigger the full `die()` logic (drop items, XP awarded).
- **Test Case: Maiming**: Hit an actor while they are incapacitated (HP <= 0). Verify that they continue to lose limbs until they die.
