# Plan: Ranged Enemies (Goblin Archer)

## Objective
Introduce the first non-melee combatant into the game: The Goblin Archer. This enemy will keep its distance from the player and fire linearly traveling projectiles (arrows) that can be dodged or blocked by environmental obstacles.

## Analysis
Ranged combat introduces a new core mechanic: **Projectiles**. Previously, all damage was calculated via direct proximity (distance < 1.0). We must create a new entity type (`Projectile`) that moves independently across the screen each tick and checks for collisions with both the player and the environment.

## Implementation Steps

### 1. The Projectile System
- Create `internal/game/projectile.go`.
- Define the `Projectile` struct:
  - `X, Y float64`
  - `Dx, Dy float64` (Velocity vector)
  - `Speed float64`
  - `Damage int`
  - `Owner *NPC` (to ensure they don't hit themselves instantly)
- Update `Game` struct in `game.go` to hold a slice of `[]*Projectile`.
- **Projectile Update Loop**:
  - Every tick, `p.X += p.Dx * p.Speed`, etc.
  - Check collisions against all active `Obstacle` structures (trees, houses). If it hits an obstacle, it is destroyed.
  - Check collision against the `Player`. If distance < 0.5 tiles, deal damage and destroy the projectile.

### 2. The Goblin Archer Asset & Type
- Add `NPCGoblin` (or `NPCObserver`/`NPCArcher`) to `NPCType`.
- Create visual and audio assets:
  - `assets/images/npcs/goblin_static.png` (and corpse)
  - `assets/images/projectiles/arrow.png`
  - `assets/audio/weapons/bow_shoot.wav`

### 3. Ranged AI Behavior
- Modify `NPC.Update()`:
  - When the NPC is `NPCGoblin` and has `TargetPlayer`:
    - Calculate distance to the player.
    - If distance > 5.0 tiles: move toward the player normally (Seek).
    - If distance < 3.0 tiles: enter a new behavior `BehaviorFlee` and move *away* from the player (Kiting).
    - If distance is between 3.0 and 5.0 tiles: Stop moving.
    - While stopped (or kiting), increment `AttackTimer`. When `AttackTimer >= AttackCooldown` (e.g., 90 frames), Fire!
- **Firing Logic**:
  - Calculate normalized vector `(dx, dy)` pointing exactly from the Goblin to the Player's current position.
  - Spawn a `NewProjectile(n.X, n.Y, dx, dy, damage)` and append it to the `Game`'s projectile list.
  - Play the `bow_shoot.wav` sound effect.

### 4. Integration with other features
- **Minimal AI (`plans/minimal_ai.md`)**: Ranged units naturally prefer the "Survival/Fleeing" protocol when low on health.
- **Advanced Combat (`plans/combat_rules.md`)**: Projectiles should still roll for a hit! A projectile might visually hit the Knight, but if the internal `rand() < HitChance` roll fails, it "glances" off the Knight's armor (perhaps showing a grey "BLOCK" text).

## Verification
- Spawn a Goblin Archer.
- Approach it. Verify it stops 5 tiles away and shoots an arrow sprite at the Knight.
- Move perpendicular to the arrow's flight path and verify it misses the Knight and continues flying until it either hits a Tree or travels far off-screen (and is cleaned up).
- Chase the Goblin. Verify it retreats to maintain distance (kiting).
