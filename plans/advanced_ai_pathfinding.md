# Advanced AI: A* Pathfinding

## Overview
Currently, NPCs track the player using simple linear interpolation (lerping). This causes them to get permanently stuck on obstacles like trees and rocks. We need to implement a dedicated pathfinding system to allow enemies to navigate the isometric terrain intelligently.

## Implementation Steps
1. **Grid Representation**:
    - The world is procedurally generated in chunks, so we don't have a static global grid.
    - We need to create a localized navigation grid around the player (e.g., a 20x20 tile area) that updates dynamically as the player moves.
2. **Obstacle Registration**:
    - Query the `obstacleRegistry` and currently spawned chunks to mark tiles as impassable on the local nav-grid.
3. **A* Algorithm**:
    - Implement a standard A* pathfinding algorithm in `internal/engine` or `internal/game`.
    - NPCs will query this system. Given their current (X,Y) and the target player's (X,Y), return a list of waypoints.
4. **Behavior Integration**:
    - Modify `BehaviorNpcFighter` and `BehaviorChaotic` in `npc.go`.
    - Instead of directly modifying `X` and `Y` toward the player, NPCs will follow the next waypoint in their generated path.
    - Paths should be recalculated periodically (e.g., every 60 ticks) to account for the player moving, rather than every frame to save CPU.

## Challenges
- **Performance**: Running A* for 50+ enemies every frame will tank performance. Paths must be cached and recalculated at staggered intervals.
- **Dynamic Obstacles**: If a rock is destroyed, the nav-grid needs to instantly update.
