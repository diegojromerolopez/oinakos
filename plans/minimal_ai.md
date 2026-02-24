# Plan: Minimal AI System

## Objective
Overhaul the current simplistic NPC tracking logic with a "minimal" but intelligent AI system. NPCs will no longer blindly attack the Knight. They will evaluate combat conditions, use group tactics, retreat when wounded, and seek out Healing Wells.

## Analysis
The current NPC `Update` loop simply checks for `TargetPlayer` and walks directly toward them to attack until death. 
We need to introduce a **Threat Assessment** phase to the behavior loop before choosing to attack, and a **Survival Protocol** phase if health becomes critically low.

## Implementation Steps

### 1. Define Threat Assessment Criteria
- Add `HostilityState` to the `NPC` setup to track if they are currently passive, fleeing, or actively hunting.
- Before transitioning from passive/wandering to tracking the Knight, an NPC must evaluate:
    - **A. Health Advantage**: Is `n.Health > player.Health`?
    - **B. Number Advantage**: Are there other living NPCs within a certain radius (e.g., 5 tiles) of the checking NPC? If `nearbyNPCs > 1`, they feel emboldened to attack.
    - **C. Provocation**: The NPC instantly becomes hostile and bypasses checks A and B if `n.Health < n.MaxHealth` due to an attack from the Knight.

### 2. Implement Survival Protocol (Fleeing)
- Update `NPC.Update()`:
    - If `n.Health` drops below a critical threshold (e.g., <= 30% of `n.MaxHealth`):
        - The NPC enters the `BehaviorFlee` state.
        - Instead of calculating a movement vector *toward* the Knight, calculate a vector *away* from the Knight (`targetX = n.X - knightDx`, `targetY = n.Y - knightDy`).
        - Increase their movement speed slightly (panic sprint) to give them a chance to escape.

### 3. Seek Healing Wells
- This integrates directly with the `plans/healing_well.md` logic.
- If an NPC is in the `BehaviorFlee` state or simply has low health:
    - Scan the active obstacles array for a `TypeWell` where `CooldownTicks == 0`.
    - If a usable well is found within a broad search radius (e.g., 20 tiles):
        - Change state from `BehaviorFlee` to `BehaviorSeekingWell`.
        - Path directly to the well instead of just running randomly away from the Knight.
        - Trigger the `NPCDrinking` sequence upon arrival to restore health, then revert to standard threat assessment logic.

### 4. Code Integration Points
- `internal/game/npc.go`: Expand `NPC.Update()` with the new `calculateThreat` and `calculateEscape` helper methods to keep the `Update` loop readable.
- Implement distance checks using standard Pythagorean calculations against the array of active chunks/entities.

## Verification
- Spawn several NPCs with lower health than the Knight. Verify they wander passivly.
- Attack one. Verify it turns hostile (provoked).
- Lure the Knight toward a large group of 3+ NPCs. Verify they all turn hostile simultaneously due to numbers advantage.
- Damage an NPC to <30% health. Verify it stops attacking, turns around, and moves rapidly in the opposite direction.
- Ensure an injured fleeing NPC successfully paths to an active Well, heals, and optionally resumes hostility if conditions allow.
