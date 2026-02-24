# Plan: Dark Knight Enemy

## Objective
Introduce a new formidable enemy type, the "Dark Knight". This NPC is a twisted reflection of the player's character that has fallen to the abyss. It requires separate visual assets and high combat stats, acting as an elite or rare enemy encounter.

## Analysis
Adding a new enemy type fits cleanly into the existing `NPCType` structure in `internal/game/npc.go`. The Dark Knight should be significantly harder to defeat than an Orc or Demon, perhaps having similar health and damage profiles to the player character themselves.

## Implementation Steps

### 1. Visual & Audio Assets
- **Sprite Generation**: Create new static, walking, and attacking sprites for the Dark Knight. These should mirror the player's `knight` sprites but feature a dark, corrupted color palette (e.g., black armor with glowing red or purple accents).
  - `assets/images/npcs/dark_knight_static.png` (and variants)
- **Audio Effects**:
  - `assets/audio/dark_knight/dark_knight_hit.wav`
  - `assets/audio/dark_knight/dark_knight_death.wav`
  - `assets/audio/dark_knight/dark_knight_menace_[1-5].wav` (Lower, distorted versions of grunt sounds)
- Update `internal/game/game.go` to load these assets into memory.

### 2. Extend NPCType Constants
- Modify `internal/game/npc.go`:
  - Add `NPCDarkKnight` to the `NPCType` constant generation block.
  ```go
  const (
      NPCOrc NPCType = iota
      NPCDemon
      NPCPeasant
      NPCDarkKnight
  )
  ```

### 3. Stat Initialization
- Update `NewNPC` factory function with a new case for `NPCDarkKnight`:
  - **Health**: Very high (e.g., `150 + rand.Intn(51)` -> 150-200 HP).
  - **Damage**: High (e.g., `6` or `8` damage per hit).
  - **Speed**: Match the player's speed or slightly slower (e.g., `0.06`).
  - **Attack Cooldown**: Match the player's swing speed (e.g., `45` frames).

### 4. Footprint and Hitbox
- Update `getTypeFootprint()` to return the exact same `engine.Polygon` footprint as the Player uses (since they are both knights wearing similar armor), ensuring accurate collision detection.

### 5. Spawn Logic Integration
- Update chunk generation in `internal/game/game.go` (`updateNPCSpawning`).
- Ensure Dark Knights spawn *rarely*. Instead of equal probability with Orcs and Demons, they should have a low chance (e.g., 5-10% of hostile spawns) to make their appearance meaningful and scary.
- *Optional*: Gate their spawning behind the player's kill count or survival timer (e.g., only spawn after 2 minutes of survival).

### 6. Minimal AI Integration
- Ensure Dark Knights interact smoothly with the new "Minimal AI" system (`plans/minimal_ai.md`). Because their health is so high, they will almost always pass the "Health Advantage" test and immediately charge the player, making them relentlessly aggressive until heavily wounded.

## Verification
- Run the game and increase the spawn rate temporarily for testing.
- Locate a Dark Knight. Verify its corrupted visual appearance.
- Engage in combat. Verify it deals high damage and takes significantly longer to kill than a Demon.
- Defeat it and ensure the specific `dark_knight_death` audio plays.
