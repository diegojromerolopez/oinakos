# Plan: Healing Well Feature

## Objective
Introduce a new interactive environmental structure, the "Well". Both the Knight and NPCs can drink from it to restore health. It features a unique drinking animation, a shared 5-minute cooldown, and visual UI indicating the remaining cooldown time.

## Analysis
This feature introduces several new systemic concepts to the game:
1.  **Interactive Obstacles**: Until now, obstacles (trees, houses) have been purely static collision objects. We need to introduce state (`CooldownTimer`) to specific obstacle types.
2.  **Contextual Player Actions**: The `SPACE` bar is hardcoded to attack. We must add contextual logic to check for a nearby, active Well and prioritize "Drinking" over "Attacking".
3.  **NPC Resource Seeking**: We need to expand NPC AI so that damaged NPCs actively search for and path toward an active Well rather than just wandering or attacking.
4.  **New Animation States**: `StateDrinking` for the Knight and `NPCDrinking` for the NPCs.

## Implementation Steps

### 1. Asset Requirements
-   **Static Sprites**: `assets/images/environment/well.png` (and optionally `well_empty.png` to visually indicate it's exhausted).
-   **Animation Frames**: Add a "drinking" pose/row to the sprite sheets planned in `sprite_animations.md` for both the Knight and all NPCs.
-   **Audio**: A new drinking sound effect (`assets/audio/environment/drink.wav`).

### 2. Extend Obstacle System
-   **Modify `Obstacle` struct** in `internal/game/obstacle.go`:
    -   Add `Type int` (e.g., `TypeTree`, `TypeHouse`, `TypeWell`).
    -   Add `CooldownTicks int`.
-   **Level Generation**: Update chunk generation to occasionally spawn a `TypeWell` instead of a rock or tree.
-   **Obstacle Update/Draw Loop**:
    -   In the `Update` loop (to be added to `Obstacle`), decrement `CooldownTicks` if it's > 0.
    -   In the `Draw` loop, if the obstacle is a `TypeWell` and `CooldownTicks > 0`, render the remaining time below it (e.g., convert ticks to `MM:SS` format and print using `ebitenutil.DebugPrintAt`).

### 3. Knight Interaction Mechanics
-   **Update `Player.Update()`** in `internal/game/player.go`:
    -   When `SPACE` is pressed, before transitioning to `StateAttacking`, iterate through nearby obstacles.
    -   If a `TypeWell` is within interaction range (e.g., 1.5 tiles distance) and its `CooldownTicks == 0`:
        -   Set `p.State = StateDrinking`.
        -   Start an animation timer for the drinking sequence (e.g., 60 frames).
        -   Set the Well's `CooldownTicks = 5 * 60 * 60` (5 minutes * 60 seconds * 60 ticks/sec).
        -   Restore `p.Health = p.MaxHealth`.
        -   Play the drinking sound.

### 4. NPC AI Integration
-   **Update `NPC.Update()`** in `internal/game/npc.go`:
    -   Introduce a new temporary behavior state (e.g., `BehaviorSeekingWell`).
    -   If an NPC's health drops below a certain threshold (e.g., 50%) and they are not actively being attacked by the player, scan `obstacles` for a nearby `TypeWell` with `CooldownTicks == 0`.
    -   If found, set `TargetObstacle = well` and path toward it.
    -   Upon reaching interacting range, trigger `NPCDrinking` state, restore health, and trigger the Well's 5-minute cooldown.

## Verification
-   Locate a well in the procedural generator.
-   Take damage, walk up to the well, and press Space. Verify the health fully restores, the drinking animation plays, and the 5:00 countdown appears.
-   Press Space again immediately -> verify the Knight swings his sword instead (well is empty).
-   Spawn an NPC, deal damage to it, and watch from a distance. Verify the NPC paths to a filled well, heals itself, and triggers the cooldown.
