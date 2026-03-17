# Plan: Dwarf Fortress Simulation Mimicry

This plan outlines the technical roadmap to transform Oinakos into a deep simulation, mimicking core features of Dwarf Fortress (DF) while remaining optimized for laptop performance.

## 1. World Generation & Procedural History

The world should feel lived-in. Instead of static maps, we will shift towards a historical generation phase.

### Implementation
- **Seed-Based World Gen**: Create a `WorldGen` module in `internal/game/world_gen.go`.
- **Biome Mapping**: Generate a low-res global map with biomes (forest, desert, tundra) that determines the "vibe" of loaded maps.
- **Years of History**: Simulate 50-100 years of events:
    - **Founding**: Villages and fortresses appear.
    - **Deaths**: Leaders die in battles or of old age, creating "Historical Figures".
    - **Artifacts**: Unique items can be created during history and found later by the player.
- **NPC Integration**: Use history to name NPCs and set their initial moods/alignments.

## 2. Deep Entity Simulation (Body & Needs)

Replacing the generic `Health` bar with a granular system.

### Body Parts
- **Granular Health**: Actors get a `Body` struct with parts: Head, Torso, Left Arm, Right Arm, Left Leg, Right Leg.
- **Damage Effects**: 
    - Broken leg = Speed penalty.
    - Blinded (head injury) = Vision range penalty.
    - Bleeding = Damage over time until "bandaged" (new interaction).

### Needs & Moods
- **Basic Needs**: Hunger, Thirst, Sleep, and Entertainment.
- **Mood Engine**: Calculated based on satisfied needs and recent "Memories" (e.g., "Slept in a nice bed", "Was hit by a goblin").
- **AI Influence**: Mood modifies behavior. A "depressed" NPC might wander aimlessly, while a "berserk" one attacks everything.

## 3. Z-Levels & Elevation

Dwarf Fortress is famous for its verticality.

### Approach
- **Pseudo-3D Isometric**: Use the existing isometric math but add a `Z` coordinate to entities and obstacles.
- **Sorting**: Update the Y-sorting logic to include `Z*TileHeight` in the priority calculation.
- **Cliffs/Ramps**: Define specific obstacle types that allow `Z` transitions.

## 4. Resources & Production

Entities need to interact with the world beyond just combat.

### Resource System
- **Collection**: NPCs and the Player can gather resources (Wood, Stone, Iron, Food) from specific obstacles or map zones.
- **Stockpiles**: Designated areas or containers where resources are stored.

### Action Engine
- **Task Execution**: NPCs can be assigned "Actions" (e.g., "Forge Sword", "Cook Meal").
- **Productivity**: Executing an action consumes resources and produces a **Product** (requires a new `product.png` asset) or a world consequence (e.g., a new building).
- **In-Game Logic**: Actions take time (ticks) and can be interrupted by combat or needs (e.g., hunger).

## 5. AI-Extensible Maps

Static map YAMLs are the baseline; the AI and simulation will build upon them.

### Dynamic Map Modification
- **AI Construction**: The AIManager can request "Construction" events where the map layout changes (walls built, floor tiles swapped) based on the world's historical or current state.
- **Procedural Expansion**: When the player reaches a map edge, the AI can "stitch" a new procedurally generated chunk based on the current Biome.

## 6. Realism Dynamics: Illness & Contagion

Survival is not just about avoiding blades.

### The Sick State
- **Contraction**: Illness can be contracted from "Miasma" zones, tainted food, or contact with other sick actors.
- **Dynamics**: Illness affects stats (Speed, Strength) and can lead to death if not treated.
- **Contagion**: Actors in close proximity have a chance to spread the illness.
- **Visualization**: Sick actors display a specific **Sick Icon** (requires a new `sick.png` status icon).

## 7. Visualization (Simulation Debug Overlay)

To make the simulation "visible" on a laptop screen:

### The "U" Overlay
- **Thoughts Bubble**: Display the current "Major Thought" of an NPC above their head.
- **Status Icons**: Small icons for hunger (🍗), thirst (💧), and pain (💢), and illness (🤢).
- **Body Map**: A small HUD element when hovering over an NPC showing their limb status.

## 8. Optimization Strategy

- **Lazy Simulation**: Only "deeply" simulate NPCs within a certain distance of the player. NPCs further away are simulated via "Abbreviated History" (lightweight background math).
- **Seed Persistence**: Store only the seed and player changes to keep save files small.

---

**Next Steps**: Once this plan is reviewed in the context of the repository, we can begin creating the `Body` and `WorldGen` modules.
