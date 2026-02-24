# Plan: Healer NPC

## Objective
Introduce a new non-hostile NPC type ("Healer") that periodically spawns, runs toward the player, heals them upon contact, and then leaves or despawns.

## Analysis
This requires expanding the existing `NPC` system. Currently, NPCs are largely hostile (Orc, Demon) or passive/randomly hostile (Peasant). The Healer needs a unique behavior state (`BehaviorHealPlayer`), a distinct visual asset, and a timer-based spawn mechanic independent of the standard chunk-based enemy spawner.

## Implementation Steps

### 1. Asset Requirements
- Need a visual asset for the Healer (e.g., `assets/images/npcs/healer_static.png`). It can be male or female, perhaps wearing robes to distinguish them from peasants.

### 2. Extend NPCType and Behaviors
- In `internal/game/npc.go`, add `NPCHealer` to the `NPCType` constant list.
- Add a new behavior: `BehaviorHealPlayer`.
- Give the Healer high speed (to catch up to a running player) but 0 damage.

### 3. Spawning Logic
- In `internal/game/game.go` (specifically in `Update` or `updateNPCSpawning`), introduce a dedicated timer for the Healer.
- Logic:
  - Track `healerSpawnTimer`. Increment every tick.
  - Every 60 seconds (approx `60 * 60 = 3600` ticks), roll a probability check (e.g., 30% chance).
  - If successful, spawn an `NPCHealer` just outside the camera viewport.
  - Force their behavior to `BehaviorHealPlayer` with the Knight as the target.

### 4. Behavior Update
- Modify `NPC.Update()`:
  - If `Behavior == BehaviorHealPlayer`, calculate distance to `TargetPlayer`.
  - Move toward the player exactly like a hostile NPC.
  - **Intersection Event**: If distance < interactRange (e.g., 1.5 tiles):
    - Trigger a heal on the player: `player.Health += 30` (capped at `MaxHealth`).
    - Play a positive sound effect (e.g., "assets/audio/environment/heal.wav" - needs to be created/loaded).
    - Despawn the Healer (set state to `NPCDead` immediately without a corpse, or add a fade-out effect).

### 5. UI Improvements
- Optional but recommended: Draw a green "+" icon or text above the player when healed to provide visual feedback.

## Verification
- Play the game for several minutes.
- Survive and take some damage.
- Wait for a Healer to spawn, let them approach, and verify that the Health bar increases and the Healer subsequently disappears.
