# Plan: Multi-Tile Floor System & Zones 🌿

This plan outlines the implementation of a diverse, grid-based floor system and a "Floor Zone" architecture to replace the current single-texture "Infinite Grass". This will allow for paths, dirt patches, biomes, and hand-crafted regions while maintaining performance and infinite world support.

## 1. Data Model Enhancements
### `internal/game/config.go`
- **`FloorTileConfig`**: A new struct to define a floor tile archetype.
  - `ID`: Unique string ID (e.g., "grass_lush", "dirt_path").
  - `Path`: File path to the sprite.
- **`FloorZone`**: A geometric area with a specific tile override.
  - `Name`: String for debug purposes (e.g., "The Sacred Grove").
  - `Perimeter`: `engine.Polygon` defining the bounds.
  - `TileID`: The ID of the `FloorTileConfig` to use.
  - `Priority`: Integer to handle overlapping zones (higher numbers draw on top).
- **Update `MapType`**: Add `FloorZones []FloorZone` and a `DefaultTileID string`.

## 2. Asset Management
### `internal/game/game_render.go`
- **`floorSprites`**: A map `map[string]engine.Image` to cache loaded floor textures.
- **`LoadAssets`**: Update to load the default tile and all tiles referenced in the map's zones.

## 3. Rendering Architecture
### `internal/engine/renderer.go`
- **Generalize `DrawInfiniteGrass`**: Rename to `DrawTileMap`.
- **Selection Callback**: Instead of taking a single `grassSprite`, it should take a function: `func(x, y int) engine.Image`.

### `internal/game/game_render.go`
- Implement the zone-aware tile selection logic:
  ```go
  func (gr *GameRenderer) getTileAt(x, y float64) engine.Image {
      // 1. Sort zones by Priority (can be pre-cached)
      // 2. Check each zone:
      for _, zone := range gr.game.currentMapType.FloorZones {
          // Optimization: Check Bounding Box (AABB) first
          if zone.Bounds().Contains(x, y) {
              if zone.Perimeter.Contains(x, y) {
                  return gr.floorSprites[zone.TileID]
              }
          }
      }
      // 3. Fallback to default
      return gr.floorSprites[gr.game.currentMapType.DefaultTileID]
  }
  ```

## 4. Performance Optimizations
- **AABB Filtering**: Every `FloorZone` will calculate its axis-aligned bounding box (min/max X/Y) at load time. This allows the renderer to skip complex polygon math for 99% of tiles.
- **Ray-Casting (Point-in-Polygon)**: Add a `Contains(x, y)` method to `engine.Polygon` in `internal/engine/collision.go`.

## 5. Debugging & Logging
- **Debug Mode (Tab)**: When active, the HUD or `DebugLog` can display the name of the zone the player is currently standing in.
- **Visual Outlines**: Optional drawer to show the polygon perimeters of floor zones in debug mode (similar to character footprints).

## 6. Implementation Steps
1. **Engine Update**: Add `Contains(p Point)` to `engine.Polygon`.
2. **Registry**: Add a `FloorRegistry` (or update existing) to load floor tile and zone YAMLs.
3. **Renderer Refactor**: Modify `engine.Renderer` to support the selection callback.
4. **Integration**: Update `MapType` to support the new YAML structure.

## 🧪 Testing Strategy
- **Perimeter Accuracy**: Unit tests to verify that `Contains` correctly handles concave polygons and edges.
- **Priority Overlap**: Verify that if a "Path" zone crosses a "Meadow" zone, the Path tile correctly takes precedence.
- **Performance Leak**: Measure frame time on maps with 50+ zones to ensure AABB filtering is working.
