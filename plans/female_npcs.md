# Plan: Female NPC Variants

## Objective
Introduce female variations for the three existing NPC types (Orc, Demon, Peasant). These variants will use distinct visual sprites and possess modified combat stats: they will deal less damage but move significantly faster than their male counterparts.

## Analysis
Currently, `NPCType` strictly defines the visual and mechanical properties of an NPC upon creation (`NewNPC`). To cleanly support variations without doubling the amount of `NPCType` constants (e.g., `NPCOrcMale`, `NPCOrcFemale`), we can introduce a new custom type `GenderType int` and add it to the `NPC` structure. This follows Go's idiomatic `iota` pattern, allowing for both Male/Female now while being completely extensible for other genders (e.g., non-binary, monstrous, or diverse forms) without string comparison overhead.

## Implementation Steps

### 1. New Visual Assets
- Generate new single-frame (or sprite sheet) assets for the female counterparts to ensure distinct visual representation:
  - `assets/images/npcs/orc_female_static.png` (and corpse/attack states)
  - `assets/images/npcs/demon_female_static.png` (and corpse/attack states)
  - `assets/images/npcs/peasant_female_static.png` (and corpse/attack states)
- Update `internal/game/game.go` to load these new image pointers into memory.

### 2. Extend the NPC Structure
- Modify `internal/game/npc.go`:
  - Define a new type and constants:
    ```go
    type GenderType int
    const (
        GenderMale GenderType = iota
        GenderFemale
    )
    ```
  - Add the new attribute to the `NPC` struct: `Gender GenderType`.

### 3. Modify NPC Generation Logic
- Update the `NewNPC` factory function.
- Whenever an NPC is spawned (Orc, Demon, or Peasant), determine its gender. For example, assign `n.Gender = GenderMale` or `n.Gender = GenderFemale` with equal probability (`rand.Float64() < 0.5`). 
- **Apply Stat Modifiers**:
  - In the initial switch statement setting stats, if `n.Gender == GenderFemale`:
    - **Speed Buff**: Multiply base speed (e.g., `n.Speed *= 1.3` for a 30% speed increase).
    - **Damage Nerf**: Reduce base damage (e.g., `n.Damage = int(math.Max(1, float64(n.Damage)*0.75))` to guarantee at least 1 damage but generally deal 25% less).
    - *(Optional)* **Health Tweak**: slightly reduce base health.

### 4. Separate Names Lists
- Currently, `npcNames` is a single slice of masculine names ("Grog", "Bob").
- Create distinct lists (e.g., `npcFemaleNames`) and assign the name based on the `Gender` attribute in `NewNPC`.

### 5. Update Rendering Logic
- Update `NPC.Draw()` to select the correct visual asset based on the `GenderType`.
- Because `game.go` currently passes the specific sprites directly into `NPC.Draw`, we must change how rendering assets are assigned.
- **Better approach**: Rather than passing 6 different `*ebiten.Image` pointers into `NPC.Draw` every frame, the `Game` struct (or an `AssetManager`) should store slices/maps of sprites. `NPC.Draw` can look up its own correct sprite using `n.Type` and `n.Gender`.

## Verification
- Run the game and watch the dynamic spawner.
- Verify visually distinct female NPCs spawn alongside male ones.
- Hover near a female NPC to verify the floating UI name draws from the female name pool.
- Verify female NPCs close the distance significantly faster when tracking you.
- Let a female NPC hit you to verify the damage reduction applies correctly.
