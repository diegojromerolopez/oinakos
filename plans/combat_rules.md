# Plan: Advanced Combat Rules & Stats

## Objective
Overhaul the current deterministic combat system (every attack hits for flat damage) into a hit-probability system. Introduce weapons as lootable items, Defense stats, and an experience-based progression system that logarithmically scales character abilities.

## Analysis
This represents a major shift toward traditional RPG mechanics. 
1. **Hit Calculation**: We need a `hit chance` formula: `Roll(1-100) < (Attacker Attack - Defender Defense)`.
2. **Weapons**: Characters need an active `Weapon` struct that defines their damage interval (e.g., 2-5).
3. **Progression**: Characters need an `Experience` tracker that slowly increases their base `Attack` and `Defense` logarithmically to prevent infinite power scaling.
4. **Armor & Shields**: In addition to a `Defense` stat (accuracy/evasion), characters need `Armor` points that reduce damage from successful hits. Armor comes from multiple slots: Body, Shield, Helmet, etc.
5. **HUD/UI**: The top-left menu must display current Weapon Damage, Defense, and total Shield/Armor points.

## Implementation Steps

### 1. Equipment (Weapons & Armor)
- Create `internal/game/equipment.go` (expanding on previous weapon plan).
- Define `Armor` struct: `Name string`, `Protection int`, `Slot string` (Head, Body, Shield).
- Define constants for Armor types:
  - **Body (Corporal)**: Leather (1), Chainmail (2), Plate (5).
  - **Shield**: Wood (1), Iron (2), Tower (4).
  - **Helmet**: Cap (1), Full Helm (2).
- Add `EquippedArmor` map/struct to `NPC` and `Player` to track these slots.
- *Spawning Equipment*:
  - Update `updateObstacles` to occasionally drop `TypeWeaponLoot` or `TypeArmorLoot` on the ground.
  - Picking up armor replaces the item in the corresponding slot.

### 2. Base Stats & Progression
- Update `NPC` and `Player` structs:
  - Add `BaseAttack int` and `BaseDefense int` (representing their skill/accuracy/evasion).
  - Add `Level int` and `Experience int`.
- **Logarithmic Scaling**:
  - Create a progression helper function: `func calculateStat(base, level int) int { return base + int(math.Log2(float64(level+1)) * scalingFactor) }`
  - When the player gains XP from a kill, check if `XP > threshold(Level)`. If so, increment `Level` and recalculate `Attack` and `Defense`.
  - *NPC Scaling*: As the player's survival timer increases, newly spawned NPCs should spawn at higher base levels.

### 3. Hit/Miss Combat Logic
- Modify `Player.Attack()` and `NPC.Update()` (combat phase).
- **The Roll**:
  - `hitChance := attacker.Attack - defender.Defense`
  - Clamp `hitChance` to a reasonable range (e.g., minimum 5%, maximum 95%).
  - `roll := rand.Intn(100) + 1`
  - If `roll <= hitChance`: **HIT!**
- **The Damage**:
  - If it's a hit, calculate raw damage from the weapon interval: `rawDamage := rand.Intn(CurrentWeapon.MaxDamage - CurrentWeapon.MinDamage + 1) + CurrentWeapon.MinDamage`
  - Total Protection is summed from all equipped armor: `protection := player.GetTotalProtection()`
  - Final Damage: `finalDamage := int(math.Max(1, float64(rawDamage - protection)))`
  - `defender.Health -= finalDamage`

### 4. UI & HUD (Top Left Menu)
- Implement a persistent UI overlay in the top-left corner using `ebitenutil.DebugPrintAt` or a custom text renderer.
- **Display Fields**:
  - **Weapon**: `[Name] (Min-Max)`
  - **Attack**: current accuracy stat.
  - **Defense**: current evasion stat.
  - **Shield/Armor**: sum of all equipped armor protection values.

### 5. UI Feedback
- To make misses obvious to the player, we should render temporary floating text over the target.
- If a miss occurs, draw `"MISS"` in grey text above the target's head for 30 frames.
- If a hit occurs, draw `"-X"` (where X is damage) in red text.

## Verification
- Engage an Orc. Observe that some swings naturally miss the Orc based on the RNG roll.
- Kill enemies to level up. Verify the `Player.Level` increases, which slightly boosts the `Attack` stat, making consecutive hits on basic Orcs more frequent.
- Pick up a new weapon from the ground. Verify the damage range increases on successful hits.
