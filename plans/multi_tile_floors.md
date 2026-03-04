# Plan: Multi-Tile Floor System & Zones 🌿

This plan outlines the implementation of a diverse, grid-based floor system and a "Floor Zone" architecture to replace the current single-texture "Infinite Grass". This will allow for paths, dirt patches, biomes, and hand-crafted regions while maintaining performance and infinite world support.

## 1. Data Model Enhancements
### `internal/game/config.go`
- **`FloorTileConfig`**: Define a floor tile archetype.
  - `ID`: Unique string ID (e.g., "grass_lush", "dirt_path").
  - `Path`: File path to the sprite inside `assets/`.
- **`FloorZone`**: A geometric area with a specific tile override. Coordinates MUST be strictly Cartesian (physics space).
  - `Name`: String for debug purposes.
  - `Perimeter`: `[]engine.Point` defining the bounds in Cartesian space.
  - `TileID`: The ID of the floor tile to use.
  - `Priority`: Integer to handle overlapping zones (higher numbers draw on top).
- **Update Map Data**: Update the map schemas and persistence structs (e.g., `SaveData` -> `MapData`) to handle a list of `FloorZones` and a `DefaultTileID`.

### Mock YAML Example
```yaml
id: "crossroads"
default_tile: "grass_lush"
zones:
  - name: "dirt_path"
    tile_id: "dirt_patch"
    priority: 1
    perimeter:
      - x: -5, y: -2
      - x: 5, y: -2
      - x: 5, y: 2
      - x: -5, y: 2
```

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
- **Ray-Casting (Point-in-Polygon)**: Add a `Contains(x, y float64)` method to `engine.Polygon` in `internal/engine/collision.go`.

## 5. Map Editor Integration
Since Oinakos relies on heavily visual tooling, the Map Editor (`make map-editor`) must be natively updated to support crafting these multi-tile floors manually:
- **Zone Polygon Tool**: Enable the mouse to click points and draw `FloorZone` polygons on the map view.
- **Tile Picker Sidebar**: Provide a list of loaded `FloorTileConfig` textures to assign to the active polygon.
- **Serialization**: Ensure the Map Editor saves the drawn `FloorZones` list directly into the `.oinakos.yaml` or map YAML via `gopkg.in/yaml.v3`.

## 6. Debugging & Logging
- **Debug Mode (Tab)**: When active, the HUD can display the name of the `FloorZone` the player's cartesian `(X,Y)` point falls into.
- **Visual Outlines**: Optional `DrawPolygon` usage to show the perimeters of floor zones in debug mode (similar to character footprints).

## 7. Implementation Steps (Strict DI Compliance)
1. **Engine Math**: Add `Contains(x, y float64) bool` and an AABB `Bounds()` method to `engine.Polygon` in `internal/engine/`. *CRITICAL*: Do not import `ebiten`.
2. **Data Structures**: Add `FloorZone` list to map YAML schemas.
3. **Renderer Refactor**: Generalize `engine.Renderer.DrawInfiniteGrass` to `DrawTileMap` accepting a tile resolver callback. Ensure rendering math converts the Cartesian checkerboard to Isometric properly.
4. **Zone Logic**: Implement AABB filtering and Point-in-Polygon in `internal/game/game_render.go` to select the right tile image per cell.
5. **Map Editor Update**: Add polygon drawing functionality in `tools/map_editor/main.go` using the `engine.Input` interface.

## 🧪 Testing Strategy
- **Perimeter Accuracy**: Unit tests to verify that `Contains(x, y)` accurately handles concave polygons and edges without a window/monitor (Headless).
- **Priority Overlap**: Verify that if a "Path" zone crosses a "Meadow" zone, the Priority index ensures the Path tile correctly takes precedence.
- **Performance Leak**: Measure frame time on maps with 50+ zones to ensure AABB filtering bounds-checks are halting O(n) ray-casting correctly.
