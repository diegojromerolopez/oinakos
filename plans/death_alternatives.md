# Plan: Alternatives to Death (Incapacitation States) ⛓️

## Objective
Develop alternative states for characters instead of binary life/death, allowing for capturing, torture, or other non-lethal outcomes.

## Analysis
Currently, characters go straight to `ActorDead` when health hits zero. We want to introduce a "Vulnerable" or "Downed" window where the player or other NPCs can decide the character's fate.

### Key Requirements:
1. **New Actor States**: `ActorTortured`, `ActorEnslaved`, `ActorPinned`.
2. **Surrender Mechanism**: NPCs may "Surrender" at low HP instead of fighting to the death.
3. **Interactive Fates**: Player can interact with downed NPCs to determine their fate.

---

## Implementation Details

### 1. State Definition (`internal/game/actor.go`)
- Add new constants to `ActorState`:
    - `ActorIncapacitated`: Downed, unable to move or attack.
    - `ActorTortured`: Static state, often attached to a specific prop.
    - `ActorEnslaved`: Follower state where the NPC loses their weapon and follows a "Master".
    - `ActorPinned`: Immobile, stuck to a rock or wall.

### 2. Surrender & Incapacitation (`internal/game/npc.go`)
- Modify `TakeDamage`:
    - If `Health < 15%`, there is a chance the NPC enters `ActorIncapacitated` instead of dying.
    - Downed NPCs change their `Alignment` to `Neutral` and play a "Kneeling" or "Surrender" sprite frame.

### 3. Interactive Interaction
- When a player is near an `ActorIncapacitated` NPC, show a choice (e.g., via a small menu or context keys):
    - **Enslave**: NPC follows the player (using `BehaviorEscort`), but their weapon name is set to "None" and they cannot attack.
    - **Torture**: If a torture rack or post is nearby, the NPC is moved to it and set to `ActorTortured`. They periodically emit "Scream" floating texts.
    - **Pin to Rock**: If near a rock obstacle, the NPC is moved to the rock's edge and set to `ActorPinned`.

### 4. Rendering & Assets
- Use palette swapping or unique sprites (e.g., `shackles.png` overlay) to indicate capture.
- Update `actor_render.go` to handle the specific poses for these new states.

---

## Verification
- Attack an Orc until it surrenders.
- Press 'E' (or a dedicated interact key) to capture it.
- Verify the Orc follows the player around the map in a "captured" pose without attacking.
- Attempt to "Pin" an NPC to a nearby rock and verify they become immobile.
