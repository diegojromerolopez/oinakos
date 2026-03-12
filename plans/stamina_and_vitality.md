# Plan: Stamina, Exhaustion & Vitality-Based Combat 🌬️

## Objective
To move away from "click-spam" combat by introducing a Stamina system that rewards patience and penalizes over-exertion. Furthermore, tie the actor's physical health to their combat effectiveness, creating a "death spiral" where being wounded significantly weakens an entity.

## Analysis
Combating "non-stop attacks" requires a resource that regenerates over time and is consumed by actions. If this resource is depleted and the actor continues to push, the body suffers (the "crumbling" effect).

### Key Mechanics:
1.  **Stamina Bar**: A new resource for all `Actor` entities.
2.  **Exhaustion Damage**: Attacking with zero stamina drains `Health` instead.
3.  **Vitality Scaling**: Attack speed and damage output scale based on current HP %.
4.  **NPC Fatigue**: AI must be updated to "wait" and recover stamina before re-engaging.

---

## Implementation Details

### 1. Actor State Updates (`internal/game/actor.go`)
- Add `Stamina` and `MaxStamina` to `Actor`.
- Add `StaminaRegenRate` and `AttackStaminaCost`.
- In `Actor.Update`, regenerate Stamina if not attacking/running.

### 2. The "Crumbling" Effect (`internal/game/playable_character.go` & `npc.go`)
- In `CheckAttackHits`:
    - Check if `Stamina >= AttackStaminaCost`.
    - If YES: Consume Stamina, perform normal attack.
    - If NO:
        - Perform attack at 50% damage.
        - **Crumble**: Inflict `ExhaustionDamage` (e.g., 5-10 HP) on the attacker.
        - `DeadTimer` or `HitTimer` could trigger a brief "gasping" animation.

### 3. Vitality-Based Offense
- Modify damage calculation in `TakeDamage`:
    - `FinalDamage = WeaponRoll * HealthModifier`.
    - `HealthModifier` is a function of `CurrentHP / MaxHP`.

---

## Proposed Mathematical Models (Functions)

To implement the scaling of damage based on health and the exhaustion logic, several approaches can be used:

### Proposal A: Linear Scaling (Predictable)
*   **Damage/Attack Modifier**: `f(util) = 0.5 + (0.5 * currentHP / maxHP)`
    *   *Effect*: At full health, you deal 100%. At near-death, you deal 50%.
*   **Stamina Recovery**: Flat `+2` stamina per tick while idle.

### Proposal B: Sigmoid / S-Curve (Threshold-based)
*   **Damage Modifier**: Use a sigmoid function so that effectiveness stays high until about 40% health, then drops sharply.
    *   `f(HP_ratio) = 1 / (1 + exp(-10 * (HP_ratio - 0.3)))`
    *   *Effect*: Actors feel strong until they are "wounded," then become nearly useless. This rewards keeping health high.

### Proposal C: The "Adrenaline" Inverse (Controversial/High-Risk)
*   **Damage Modifier**: `f(HP_ratio) = 1.0 + (1.0 - HP_ratio)`
    *   *Effect*: You actually deal **more** damage the closer to death you are (berserker style), but your **Stamina Cost** doubles.
    *   *Exhaustion*: At 0 stamina, you take **triple** damage from "crumbling."

### Proposal D: Square Root Decay (Gentle)
*   **Damage Modifier**: `f(HP_ratio) = sqrt(HP_ratio)`
    *   *Effect*: A subtle drop. 50% health still results in ~70% damage. Good for preventing a frustratingly fast death spiral.

---

## NPC AI Integration (`internal/game/npc.go`)
- Update `NPC.Update`:
    - If `Stamina < AttackStaminaCost`, change state to `NPCIdle`.
    - NPC will "circle" the player or back away during this state to visualize "catching their breath."
    - This creates rhythmic "windows of opportunity" for the player.

---

## HUD & Visuals (`internal/game/game_render.go`)
- Add a **Stamina Bar** (Yellow/Green) below the Health Bar.
- When an actor "crumbles" (attacks without stamina), show a **"EXHAUSTED"** floating text (Grey/Purple).
- Use `palette_swap` to make the character look "pale" or "sweaty" when stamina is low.

---

## Verification
- Spam the attack key. Verify that after 5-6 swings, the Stamina bar is empty.
- Continue attacking while empty. Verify the player takes small amounts of damage and deals less damage to enemies.
- Get the player to 10% health. Verify that attacks are noticeably slower or weaker (depending on the chosen function).
- Observe an Orc combat. Verify that the Orc stops attacking for 1-2 seconds after a flurry to recover stamina.
