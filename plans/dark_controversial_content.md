# Plan: Dark & Controversial Content🌑

## Objective
To introduce extreme survival horror elements, body horror, and psychological despair into Oinakos, shifting the tone toward a bleak, nihilistic experience.

## Analysis
To achieve this atmosphere, we must move beyond standard RPG stats and introduce persistent, brutal consequences for the player's actions and failures. The controversy arises from the game's willingness to inflict permanent disfigurement and psychological trauma on the protagonist.

### Key "Controversial" Features:
1.  **Body Part Dismemberment**: Attacks can target or randomly hit limbs, leading to permanent loss of functionality.
2.  **Hunger & Sanity**: Biological and mental decay that forces players into "evil" choices (cannibalism, sacrifice).
3.  **Grotesque Rituals**: Utilizing the "Captured" NPCs for dark empowerments or body-merging rituals.
4.  **Psychological Hallucinations**: Low sanity distorting the game world and spawning "Shadow" enemies.
5.  **Bleak Outcomes**: Missions with zero "heroic" path, only choices between different horrific sacrifices.

---

## Implementation Details

### 1. Limb System & Permanent Disfigurement (`internal/game/actor.go`)
- Add a bitmask or struct for `ActorLimbs`: `LeftArm`, `RightArm`, `LeftLeg`, `RightLeg`, `Eyes`.
- **Mechanical Impact**:
    - **Lose Leg**: Move speed reduced by 50% per leg. Character crawls if both are lost.
    - **Lose Arm**: Cannot use shields or two-handed weapons. Reduced attack speed.
    - **Lose Eye**: Screen goes half-dark or adds a "blind spot" shader.
- **Visuals**: Use `palette_swap` or sprite overlays (`missing_leg.png`) to show the loss on the character sprite.

### 2. Hunger, Sanity & Cannibalism (`internal/game/game.go`)
- Add `Hunger` and `Sanity` stats (0-100).
- **Hunger**: Decreases over time. At 0, the player takes constant damage.
- **Cannibalism**: Add an interaction with `ActorDead` entities (enemies or allies) to "Consume".
    - *Utility*: Restores hunger and health.
    - *Consequence*: Large sanity penalty; NPCs may turn hostile if they witness it.
- **Sanity**: Decreases when witnessing horrific events or being in the dark.
    - *Hallucinations*: At low sanity, spawn "Fake" enemies that only the player can see.
    - *Screen FX*: Use a shader for chromatic aberration or screen shaking that intensifies as sanity drops.

### 3. Dark Rituals & Sacrifices (`internal/game/interaction.go`)
- Expand the **Death Alternatives** plan:
    - **Human Sacrifice**: Sacrifice a captured NPC at a stone altar to regain Sanity or gain a "Dark Favor" (stat boost).
    - **Body Merging**: A high-level ritual that merges two captured NPCs into a "Marriage of Flesh" — a powerful but grotesque allied unit.

### 4. Body Horror Archetypes (`data/archetypes/`)
- Create "Controversial" enemy designs:
    - **The Amalgam**: A creature made of stitched-together peasant limbs.
    - **Stilt-Walker**: An emaciated NPC with elongated, bone-like legs.
    - **The Harvester**: An enemy that actively tries to "Pin" the player to harvest an organ.

### 5. Visceral Death & Game Over (`internal/game/game_render.go`)
- Implement "Bad Endings" for specific deaths:
    - If killed by a "Harvester," the game-over screen shows a unique text description of the player's fate (e.g., "Your parts were used to keep the machines of war turning").

---

## Content Pillars for Dark Atmosphere

| Feature | Psychological Impact | Controversial Element |
| :--- | :--- | :--- |
| **Dismemberment** | Feelings of vulnerability and permanence. | Graphic loss of player agency through disability. |
| **Cannibalism** | Survival at the cost of humanity. | Taboo interaction for gameplay advantage. |
| **Dark Rituals** | Exploitation of the weak for power. | Sacrificing allies/innocents as a core mechanic. |
| **Bleak Reality** | Hopelessness. | No traditional "Victory" ending for some levels. |

---

## Safety & Presentation
- Focus on the **horror of consequence** and **disturbing biology**.
- Avoid explicit sexual content (as per safety guidelines); focus instead on the **grotesque and the macabre**.
- Use the **Day/Night system** to amplify the dread (Sanity drops faster in deep night).

## Verification
- Enter a combat encounter and sustain a "Leg Break". Verify the player's speed is halved and the sprite is updated.
- Drop Sanity to 10% and verify that the screen begins to distort and "Fake" orcs begin to spawn.
- Use a captured NPC at a ritual site and verify the "Dark Favor" buff is received.
