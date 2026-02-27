# Biome Diversity System

## Overview
The game currently only features a generic green grass aesthetic. Adding new biomes (Snow, Desert, Swamp) will wildly increase replayability and visual interest.

## Implementation Steps
1. **Biome Data Structure**:
    - Create a new `Biome` struct in `internal/game`.
    - Properties should include: `GroundSprite` (e.g., snow_tile.png), `MovementSpeedModifier` (swamps slow you down), and `SpawnTables` (which enemies spawn here).
2. **Procedural Generation Update**:
    - Modify the chunk generator. Instead of a uniform world, use a noise function (like Perlin noise) over macro-coordinates to determine the biome of a specific chunk.
3. **Asset Creation**:
    - **Desert**: Sand tiles, cactus obstacles, mummy/scorpion enemies.
    - **Snow**: Ice tiles, pine trees, frost-troll enemies.
    - **Swamp**: Mud tiles, dead trees, slime enemies.
4. **Transitions**:
    - To prevent harsh lines between chunks, we need transition tiles (e.g., grass-to-snow blending edges) based on neighboring chunk biomes.

## Challenges
- Maintaining the Y-sorting and isometric integrity while blending distinct ground tiles.
- Managing memory as the player explores and loads multiple different biome asset packs simultaneously.
