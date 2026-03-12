# Plan: Damageable & Destructible Obstacles 🪵

## Objective
Make environmental obstacles (trees, crates, small buildings) interactable by allowing them to be damaged and destroyed, with visual feedback.

## Analysis
The `Obstacle` struct already contains `Health` and `TakeDamage` methods. We need to unify the attack logic and provide visual feedback for these interactions.

### Key Requirements:
1. **Unified Collision Combat**: Use `engine.CheckCollision` with `GetFootprint()` for all attack checks.
2. **Visual Feedback**: Show damage numbers (Floating Text) when an obstacle is hit.
3. **Destruction State**: Remove the obstacle or swap to a "broken" sprite when health reaching zero.

---

## Implementation Details

### 1. Unified Attack Collision (`internal/game/playable_character.go`)
- Refactor `CheckAttackHits` to move away from simple circle-distance checks.
- Build an **Attack Polygon** based on the actor's position and `Facing` direction (a small rectangle or wedge in front of the character).
- Check collision between this Attack Polygon and all nearby `Obstacle` footprints.

### 2. UI Feedback (`internal/game/floating_text.go`)
- When an obstacle's `TakeDamage` is called from an attack, spawn a `FloatingText` at the obstacle's coordinates.
- Use `ColorHarm` (red) for damage and potentially different colors for "Wood Splinters" or "Stone Cracks" effects.

### 3. Destruction Behavior (`internal/game/obstacle.go`)
- In `Obstacle.TakeDamage`, when `Health <= 0`:
    - Set `Alive = false`.
    - Drop potential loot (if defined in the archetype).
    - Play a destruction sound (e.g., `tree_fall.wav`, `crate_break.wav`).
- Optional: Add a `corpse_sprite` field to `ObstacleArchetype` to show a stump or debris instead of the obstacle just vanishing.

### 4. Data Updates (`data/obstacles/`)
- Update YAMLs for buildings and trees:
    - `destructible: true`
    - `health: 200` (adjustable)

---

## Verification
- Attack a tree. Verify that a red damage number appears above it.
- Verify that the attack only lands if the player is facing the tree and within the attack polygon.
- Break the tree and ensure it disappears or changes sprite, and the collision footprint is removed.
