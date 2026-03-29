This plan outlines the technical roadmap to transform Oinakos into a deep simulation, mimicking core features of Dwarf Fortress (DF) while remaining optimized for laptop performance.

> [!NOTE]
> Foundations for **Cooking, Foraging, and Resting** are now implemented. This plan builds upon these to create an autonomous ecological and social cycle.

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
    - **Hunger**: Connected to the `ConsumeItem` system. Drains over time; if it reaches 0, it reduces Max Health.
    - **Sleep/Stamina**: Connected to the `ActorResting` state. NPCs now need to find a "comfy" spot (campfire, tavern) to recover.
- **Mood Engine**: Calculated based on satisfied needs and recent "Memories" (e.g., "Slept in a nice bed", "Ate delicious cooked meat").
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
- **Autonomous Foraging**: NPCs with low hunger will now actively use the `ActorForaging` state on nearby trees/bushes.
- **Autonomous Cooking**: NPCs with `raw_meat` in their inventory will seek out a `campfire` obstacle to trigger the cooking logic.
- **Productivity**: Executing an action consumes resources and produces a **Product** or a world consequence (e.g., a new building).
- **In-Game Logic**: Actions take time (ticks) and can be interrupted by combat or higher-priority needs.

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

## 9. Social & Memory (NEW)

Simulation depth requires NPCs to form opinions based on interactions.

### The Memory System
- **Event Log**: Actors track a small queue of `RecentEvents` (e.g., "Player fed me cooked meat", "Player attacked my friend").
- **Relationships**: A map of `ActorID -> Sentiment`. Sentiment influences trade prices, dialogue barks, and combat help.
- **Grudges**: High negative sentiment leads to `BehaviorKnightHunter` targeting the player on sight.

## 11. Economic System (Trade & Currency)

A living world needs a flow of goods and value.

### Currency: The Denarius
- **Standard Unit**: The `Denarius` (plural `Denarii`) is the primary silver coin.
- **Storage**: The actor stores Denarii in their `Denarii` field (persists in save files).
- **Looting**: NPCs can carry Denarii, dropped on death or successfully pickpocked.

### Trading & Barter
- **Trader Behavior**: Some NPCs are tagged with `BehaviorTrader`. Interacting with them (via `SPACE` or dynamic key) opens the **Market Overlay**.
- **Dynamic Pricing**: The value of an item is `BaseValue * SentimentMultiplier * ScarcityFactor`.
- **Barter**: Players can swap high-value artifacts for multiple low-value items (e.g., a Golden Ring for 20 pieces of Cooked Meat).

### Supply and Demand
- **Stock replenishment**: Traders gain new stock periodically if the world simulation is in a "Peaceful" state.
- **Wealth Accumulation**: NPCs gain Denarii by selling their gathered resources (e.g., wood from `ActorChopping`).

---

## 10. Visual Asset Requirements (Asset Registry)

To support these simulation features, the following **160x160px isometric sprites** (with `#00FF00` background) are required:

### Items & Resources
- [ ] `meat.png`: A steaming, cooked piece of meat on a simple wooden platter.
- [ ] `wild_fruit.png`: A cluster of wild apples or forest fruit.
- [ ] `wild_berries.png`: A small bush-like cluster of dark forest berries.
- [ ] `wild_veg.png`: A rough, earthy root vegetable (like a wild carrot or turnip).
- [ ] `artifact.png`: A glowing, ornate box or historical relic found in ruins.

### Status Indicators (UI)
- [ ] `sick.png`: A green-tinted nauseous face or miasma cloud.
- [ ] `hungry.png`: A small icon of a bone or empty bowl.
- [ ] `tired.png`: A small icon of closed eyes or a moon.

### Character States (Animation Frames)
- [ ] `rest.png`: Character lying down in a sleeping posture.
- [ ] `eat.png`: Character holding a piece of food in a consumption pose.
- [ ] `forage.png`: Character crouching/reaching into bushes.

---

**Next Steps**: 
1.  **NPC Autonomous Cycles**: Update the AI manager to make NPCs forage and cook when hungry.
2.  **Body Part Module**: Refactor `Actor` health into a `BodyStatus` map with limb-specific debuffs.
3.  **WorldGen**: Initial implementation of seed-based map variation.
