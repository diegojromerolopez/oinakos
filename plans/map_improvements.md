# Oinakos — Map Improvement Plan 🗺️

## Objective
To overhaul and improve all maps within Oinakos (`starter_village.yaml`, `conde_olinos_hunt.yaml`, etc.) to create realistic, breathing environments. Maps will feature distinct paved roads, strategically clustered buildings, natural environmental obstacles, and dense, populated NPC distributions.

---

## 1. Floor Zones & Roads Architecture
The newly integrated `FloorZone` AABB clipping pipeline makes it cheap and performant to define complex road networks natively in YAML.

### **Paved Roads**
- **Core Town Roads**: Use `paved_ground.png` defined via `FloorZones` to lay down logical streets.
- **Dirt Paths**: Use `mud.png` or an equivalent to branch off main paved roads into farmsteads, forests, or hunting areas (applying the `x0.8` slowness modifier where appropriate).
- **Perimeter Design**: Instead of simple rectangles, road perimeters must use polygon loops connecting major POIs (Points of Interest) such as spawning points, town squares, and key buildings.

---

## 2. Building Placement & Clustering
Buildings must feel organic and clustered "along" the roads without their collision footprints overlapping. 

### **The "Close but Not Overlapping" Rule**
- **Distance Tolerances**: Given Cartesian units, standard houses (`house1`–`house3`) occupy roughly a `2x2` up to `4x4` Cartesian footprint. 
- **Row Formations**: Buildings should be spaced ~`1.5` units apart to form "blocks."
- **Road Proximity**: Place buildings functionally *next* to `paved_ground.png` `FloorZones` (within 1–3 units of a road node) so the town looks intentionally settled.
- **Town Squares**: Leave deliberate open Cartesian areas (e.g., `10x10`) at the center of road intersections. Place `well` or `temple` obstacles in the middle.

---

## 3. Natural Obstacles & Environment Scattering
Maps cannot just be buildings on an endless grass void.
- **Forests**: Group `tree1`...`tree7` in dense clusters (`tree_pine`, `tree_oak`) away from the paved roads to create natural borders. 
- **Rocks & Bushes**: Scatter `rock1`...`rock5` and `bush1`...`bush5` along dirt paths (`mud.png`) and near building walls to soften the transition between structures and nature. 
- **Water Bodies**: Use `FloorZones` with `water.png` (applies the `x0.5` slowness) enclosed by rocks or thickets.

---

## 4. Dense Population (NPCs)
A "lot" of NPCs will make the towns feel alive.
- **Civilians**: Populate the roads and town squares heavily with `peasant_male`, `villager`, and other neutral alignment archetypes. Set their behaviors to `BehaviorWander` so they dynamically traverse the roads. 
- **Guards / Soldiers**: Place named or generic `queens_guard`, `crimson_guard`, or `golden_guard` NPCs near critical buildings (e.g., temples or mayor's houses). Assign `BehaviorPatrol` to have them walk up and down specific roads, or `BehaviorWander` near the borders.
- **Density Guidelines**: Instead of 2–3 NPCs per map, a standard starter village or town center should house **20–30 NPCs**. (Engine handles up to ~100 with AABB easily).

---

## 5. Execution Steps
1. **Map Expansion**: Update the raw size of `starter_village.yaml` to ensure enough Cartesian floor space (`width_px: 3840`, `height_px: 3840` -> 60x60 logical units).
2. **Implement Zones**: Write Cartesian polygonal coordinates for `paved_surface.png` to create a "T" or "Cross" intersection road network in the center.
3. **Plop Obstacles**: 
    - Surround the roads with a tight, non-overlapping cluster of `house` and `farm` obstacles.
    - Surround the cluster perimeter with dense `tree` and `bush` obstacles. 
4. **Populate**: Add ~30 NPCs to the `inhabitants` list. Add 3–4 static patrol paths. 
5. **Validation**: Boot up the maps, run physically through them to verify collision is mathematically rigid and that the paths logically match the road zones.

---

## 6. Detailed Map-by-Map Analysis & Required Improvements

### **1. `starter_village.yaml`**
- **Current State**: Extremely barebones. Contains 2 NPCs and 1 well. Has a wildly absurd boundary size (`1000000x1000000`), which breaks local geometry limits and causes floating-point errors over time.
- **Required Fixes**:
  - **Boundaries**: Reset to a standard logical `5120x5120` (or ~80x80 units).
  - **Roads**: Design a true village square using `paved_ground.png` and several intersecting side streets (dirt paths).
  - **Buildings**: Construct a minimum of 8–10 clustered houses functioning as the town center, a farm on the outskirts, and a tavern or temple.
  - **NPCs**: Inject ~20-30 wandering neutral villagers, a few guarding patrols, and a designated mayor/VIP in the center.

### **2. `conde_olinos_hunt.yaml`**
- **Current State**: Currently features a basic linear `Town Road` zone, 5 buildings arranged artificially, and a scattering of enemies and allies.
- **Required Fixes**:
  - **Zoning & Roads**: The simple rectangular `Town Road` must be broken up into a dynamic village layout. Transition the escape path toward the portal into a winding `mud.png` trail that slows pursuers.
  - **Clustering**: Move the buildings into tight rows (1.5 units apart) lining the new `paved_ground.png` zones instead of sparse placement.
  - **Obstacles**: Scatter heavy timber (`tree_oak`, `tree_pine`) near the portal escape sequence to create a gauntlet feel.
  - **NPCs**: Multiply the ambient `peasant_male/female` count to simulate a populated town erupting into chaos, and increase the Queen's Guard volume.

### **3. `chronicles/safe_zone.yaml`**
- **Current State**: Already features foundational `FloorZones` (central road, town square, wheat fields) and decent symmetric obstacle groupings.
- **Required Fixes**:
  - **Organic Flow**: Break up the artificial symmetry of `house1` and `tree_pine` clusters. Add offset angles and natural overlapping (using rocks/bushes to hide harsh gaps). 
  - **Patrol Paths**: The 11 `crimson_guard` NPCs are static. Modify their behavior to `BehaviorPatrol`, providing them logical vertices to march back and forth along the `paved_ground.png` roads.

### **4. Campaign Maps (`orc_invasion/`, `demonic_incursion/`, `kalot_embolot/`)**
- **Current State**: Most narrative combat maps lack advanced floor aesthetics and feel like empty arenas with spawn nodes.
- **Required Fixes**:
  - **Thematic Zones**: Use specialized floor modifiers heavily. For `demonic_incursion/`, utilize `dark_water.png` pools to restrict movement. For `orc_invasion/`, use widespread `mud.png` to simulate trampled war camps.
  - **Military Clustering**: Rather than standard houses, group `tent`, `camp_fire`, and `barricade` objects in dense clusters resembling fortresses or outposts interconnected by dirt paths.
  - **Army Density**: Increase base enemy counts (`orc`, `golden_guard`, `knight`) per map to 40+ to provide genuine hack-and-slash density. Create specific defensive lines where archers have clear line of sight over `mud.png` zones where the player gets slowed.
