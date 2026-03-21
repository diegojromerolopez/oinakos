# Plan: Environment Elevation, Resource Management, and Survival Systems

This plan defines the implementation of elevation, geological accidents, resource extraction (mining/woodcutting), and character stamina in Oinakos.

## 1. Core Mechanics: Simulation of Elevation

To simulate height in a 2D isometric space, we introduce a third coordinate `Z` (Cartesian height) which translates to a vertical offset in screen space.

### 1.1 Mathematical Model
The new transformation including elevation `z` is:
- `isoX = (x - y)`
- `isoY = (x + y) * 0.5 - (z * VerticalScale)`

> [!NOTE]
> `VerticalScale` is set to `16px` per unit of Cartesian height. 
> 1 unit of `z` represents 1 Roman foot (`pes`).

### 1.2 Data Structures
- **Global Heightmap**: A 2D grid added to `MapType` for baseline elevation.
- **HeightZones**: YAML structure for regions with specific heights or gradients.
- **Entity State**: Add `Z float64` and `VerticalVelocity float64` to `Actor`.

## 2. Map Editor Integration

The `map-editor` must be enhanced to support terrain sculpting and visualization.

### 2.1 Elevation Tools
- **Height Brush**: A tool to increment/decrement `Z` values on the heightmap.
- **Flatten Tool**: Sets a selection of tiles to a uniform `Z` height.
- **Slope Tool**: Creates a smooth gradient between two points.

### 2.2 Visualization
- **Contour Lines**: Toggleable overlay showing elevation levels.
- **Height Shading**: Optional "Heatmap" overlay where higher elevations are brighter.
- **3D Preview**: A side-view or rotatable wireframe to verify verticality.

## 3. Excavation and Mining

Characters can modify the terrain through excavation using a **Pike**.

### 3.1 Excavation Mechanics
- **Difficulty**: Based on tile type (e.g., `dirt` is faster than `rock`).
- **Depth Limit**: The maximum excavation depth is **68 *passus*** (~100m).
- **Cave-In Risk**: At the maximum depth, the terrain becomes unstable. Reaching this limit results in:
  - Instant death of the character.
  - The hole is permanently covered by terrain.
  - The corpse is hidden forever.
- **Excavation Yield**: Digging creates a new floor tile at the bottom.

### 3.2 Mineral Veins
Excavating rock can reveal mineral veins (`iron`, `gold`, `silver`, `copper`).
- **Visuals**: Specialized floor tiles generated via AI for each mineral.
- **Harvesting**: Mining these tiles adds **Mineral Ore** to the inventory.
- **Weight**: Ore is extremely heavy. Inventory logic must enforce `CurrentWeight <= MaxWeight`.

### 3.3 Mining Yields (per 1,000 kg of Rock)
| Mineral | Concentration | Approx. Yield |
| :--- | :--- | :--- |
| **Copper** | 2–10% | 30 kg ore → 10–20 kg metal |
| **Silver** | 1–5% | ~200–500g silver (as byproduct) |
| **Gold** | Trace | 1–5g per ton |

## 4. Forestry and Woodcutting

Characters can harvest timber using an **Axe**.

### 4.1 Daily Yields (Fit worker with traditional axe)
- **Small trees (10–20 cm):** 15–30 trees/day
- **Medium trees (20–40 cm):** 5–15 trees/day
- **Large trees (40–60+ cm):** 2–5 trees/day
- **Usable Timber:** 500–1,500 kg (1–3 m³) per day.

## 5. Stamina and Energy System

Activities are physically taxing and require a balance of work and rest.

### 5.1 Energy (0–100)
- **Baseline**: 100 is fresh, 0 is exhausted.
- **Drain**: All activities (walking, attacking, mining, chopping, staying awake) reduce energy.
- **Exhaustion**: If energy hits 0 and the character continues working/moving, they lose **HP** over time.
- **Resting**:
  - **7–8 hours of rest/sleep** are needed to fully replenish energy and restore **20% of max Health**.
  - **Work Cycle**: Mining/Chopping involves 10–30 min bursts (burst of effort) followed by 5–10 min rest.
  - **Max Work**: Mining is limited to 4 hours/day; Woodcutting limited to 4–8 hours/day.

### 5.2 Animations and Graphics
- **Mining/Woodcutting**: The engine should attempt to load and use `digging.png` and `chopping.png` sprites for archetypes/characters if they exist. If these specific sprites do not exist, it must fallback to using `attack1.png`, and if that is also missing, fallback to `attack.png`. This logic will be paired with rendering the equipped Pike or Axe object and specific impact sound effects.
- **Resting**: Reuse the existing `crouch.png` sprite accompanied by newly generated "Zzz" sleeping particle effects to visually indicate rest without needing new character states.

## 6. Implementation Strategy

### 6.1 Simulation Rate
Stamina and resource yields will be tied to the 60 TPS simulation rate, but calculated in "Game Minutes" to match the intended work cycles.

### 6.2 AI Asset Generation
- **Minerals**: Use `generate_image` for `iron_vein.png`, `gold_vein.png`, etc.
- **Resting/Working States**: Use `generate_image` to create simple static props or tool icons (e.g., `pickaxe_icon.png`, `axe_icon.png`, `zzz_particle.png`) rather than trying to consistently regenerate complex character sprites.

## 7. Considerations
- **Weight limits**: Character strength should dictate the `MaxWeight` threshold.
- **Terrain Persistence**: excavated holes must be saved in the `.oinakos.yaml` save data.
- **Death Paradox**: Ensure the "instantly hidden corpse" logic cleans up the entity properly from the world manager.
