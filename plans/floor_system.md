# Floor System & Tile Mapping Plan 🗺️🏗️

This plan outlines the standardization of floor assets, the implementation of complex map zoning, and the technical specifications for isometric floor tiles in Oinakos.

---

## 📐 Tile Specifications

To ensure pixel-perfect alignment and consistent rendering, all floor tiles must follow these strict standards:

- **Format**: PNG with 8-bit alpha (or solid with chroma-key).
- **Dimensions**: **64x64 pixels**.
- **Usable Area**: The floor surface itself is a **64x32 diamond** centered horizontally within the 64x64 box.
- **Verticality**: The extra 32px of height allows for "depth" (e.g., water ripples below the surface level) or "height" (e.g., tall grass or raised floor panels) while keeping the base alignment consistent.
- **Chroma-Key**: All pixels outside the intended floor diamond must be set to **Vibrant Lime Green (`#00FF00`)**. The `engine.Graphics.LoadSprite` routine processes this as transparent at load time to avoid expensive per-pixel masking during the render loop.
- **Variation/Anti-Tiling**: To break up visual repetition, organic biomes (like grass or sand) should support numeric variants (e.g., `grass_1.png`, `grass_2.png`). The renderer will select a variant using a fast deterministic hash based on the logical `(x, y)` tile coordinate.

---

## 🎨 Asset Categories (Floor Types)

We will standardize on a set of core biomes/textures:

1. **Natural**:
   - `grass.png`: Standard lush green grass.
   - `dirt.png`: Dry, packed earth.
   - `mud.png`: Wet, dark, reflective soil.
   - `desert_sand.png`: Golden sand with wind ripples.
2. **Artificial**:
   - `big_stones.png`: Rough, irregular castle/temple pavers.
   - `paved_ground.png`: Civilized, flat stone tiles.
   - `stonestep.png`: Steps or elevated edge tiles.
3. **Liquids**:
   - `water.png`: Clear blue water.
   - `dark_water.png`: Murky, deep swamp or dungeon water.

---

## 🗺️ Map Zoning & Transitions

Currently, the game supports a single `FloorTile` (base map-wide texture) and multiple overlaid `FloorZones` (polygonal patches of alternate textures). We will expand this to support more organic maps.

### 1. Zone Definitions
Zones are defined in the Map YAML. *(Note: Perimeter points are defined in logical **Cartesian** coordinates, not isometric or screen pixels.)*

```yaml
floor_zones:
  - name: "Dungeon Entrance"
    tile: "big_stones"
    priority: 10
    # Coordinates are in the Cartesian physics grid
    perimeter: [{x: 0, y: 0}, {x: 20, y: 0}, {x: 20, y: 10}, {x: 0, y: 10}]
```

### 2. Implementation of "Blurry" Zones (Future)
To avoid sharp aliased edges between grass and stone:
- **Mask-Based Zones**: Allow a `zones.png` mask image where different colors represent different tile indices.
- **Blending Shaders**: Use a shader that samples adjacent tiles and blends them at the edges.

### 3. Obstacle Interaction & Feedback
Floor types should eventually affect gameplay mechanics and player feedback:
- **Movement Speed Modifiers**: 
  - **Water**: Slows movement by 50%.
  - **Mud**: Adds a minor slow.
  - **Paved/Grass**: Standard speed.
- **Dynamic Audio**: Footstep sounds should change based on the surface beneath the character (e.g., splashing in water, hollow thuds on wood, hard clicks on stone).

---

## ⚙️ Rendering Pipeline & Performance

1. **Dynamic Registry**: `FloorRegistry` will load all PNGs from `assets/images/floors`.
2. **Async Rasterization**: Tiles are rendered to a background buffer or drawn using a tilemap renderer that handles the `-32px` vertical offset for the 64x64 sprites.
3. **Culling**: Continue using the current `dim=25` Cartesian-to-Iso culling for performance.
4. **Optimized Zone Lookups**: Getting the tile under a coordinate (`getTileAt`) is currently performing O(N) polygon intersection queries per tile drawn. This must be optimized by:
   - Calculating and checking AABB (Axis-Aligned Bounding Box) for the zone before executing point-in-polygon math.
   - Generating a static chunked grid array of resolved tile IDs on map load, converting the dynamic `GetPolygon().Contains` runtime check into a strict `O(1)` slice index look-up.

---

## ✅ Immediate Action Items

- [ ] Re-generate/Refactor existing floor assets to **64x64** with lime (`#00FF00`) backgrounds.
- [ ] Add `grass_2.png` and `grass_3.png` to test deterministic variant-picking.
- [ ] Update `engine.Renderer.DrawTileMap` to align 64x64 sprites correctly (centered diamond) and apply the variant hashing.
- [ ] Implement Fast AABB bounding box properties on `FloorZone` structs to cull unneeded point-in-polygon math during rendering.
- [ ] Implement `MoveSpeed` modifiers in `NPC` and `PlayableCharacter` based on the `resolvedTile` at their current Cartesian coordinate.
- [ ] Create a "Sandbox Map" that showcases all floor types in distinct zones.
