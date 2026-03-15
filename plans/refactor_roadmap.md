# Oinakos Refactor Plan: Easing LLM-Aided Development

To improve code reasoning, maintainability, and efficiency for both human developers and AI agents, we propose the following refactoring steps. The core goal is to reduce cognitive load by decoupling state, logic, and UI, and by simplifying function signatures.

---

## 🏗️ 1. Core Architecture: The "World" & "Context" Pattern

Currently, the `Game` struct is a "God Object" holding both state and logic. Functional calls often take 10+ parameters.

### Propose: `World` Struct
Create a `World` struct to hold all live game entities and spatial data.
- **Location**: `internal/game/world.go`
- **Contents**:
  - `NPCs []*NPC`
  - `Obstacles []*Obstacle`
  - `Projectiles []*Projectile`
  - `FloatingTexts []*FloatingText`
  - `Map MapType`
  - `ExploredTiles map[image.Point]bool`

### Propose: `SystemContext` Struct
Encapsulate all external dependencies needed for updates into a single context.
- **Location**: `internal/game/context.go`
- **Contents**:
  - `World *World`
  - `Input engine.Input`
  - `Audio AudioManager`
  - `Registries *RegistryContainer` (bundle all archetypes/NPC/map registries)
  - `Log func(string, LogCategory)`

---

## 📦 2. Decoupling the `Game` Struct

The `internal/game/game.go` file (currently ~850 lines) should be split to adhere to the 500-line limit.

### Action A: Extract Dialogue System
- Move `ActiveDialogue`, `InitiateDialogue`, `AdvanceDialogue`, and `ApplyDialogueEffect` to `internal/game/dialogue_manager.go`.
- The manager will handle the state machine for conversations.

### Action B: Extract Logging & Event Handling
- Move `EventLog`, `LogEvent`, and UI scrolling logic to `internal/game/logger.go`.

### Action C: Consolidate Menus
- Move all menu state fields (`isMainMenu`, `isSettingsScreen`, `isCharacterSelect`, etc.) into the `MenuHandler`.
- `Game` should only ask `MenuHandler` if a menu is active and let it handle the response.

---

## 🏃 3. Streamlining Entity Updates

### Refactor `Update` Signatures
Change the giant parameter lists to use `SystemContext`.

**Before:**
```go
func (n *NPC) Update(pc *PlayableCharacter, obs *[]*Obstacle, reg *ObstacleRegistry, npcs []*NPC, proj *[]*Projectile, fts *[]*FloatingText, mapW, mapH float64, audio AudioManager, logFunc func(string, LogCategory), archs *ArchetypeRegistry)
```

**After:**
```go
func (n *NPC) Update(ctx *SystemContext)
```

This makes it trivial to add new dependencies (like a weather system or time-of-day) without breaking every function signature in the game.

---

## 🛠️ 4. Improved Logic Separation (Managers to Systems)

Currently, managers like `WorldManager` and `MechanicsManager` hold a pointer to `Game`. This creates tight coupling.

### Shift to "Stateless" Systems
Managers should ideally be "Systems" that provide functional updates:
1. `Systems.UpdatePhysics(world *World, dt float64)`
2. `Systems.UpdateAI(npc *NPC, ctx *SystemContext)`
3. `Systems.CheckObjectives(world *World) bool`

---

## 📝 5. Implementation Roadmap

### Phase 1: Context & World (Foundation)
1. Create `World` and `SystemContext` structs.
2. Update `NPC.Update` and `PlayableCharacter.Update` to use them.
3. Move entity slices from `Game` to `World`.

### Phase 2: Game.go Decomposition
1. Extract `DialogueManager`.
2. Extract `EventLogger`.
3. Move Menu state to `MenuHandler`.

### Phase 3: Cleanup
1. Ensure all files are under 500 lines.
2. Remove redundant helper functions (consolidation).
3. Update tests to use the new `World` / `Context` structure.

---

## 💡 Why this helps LLMs
- **Shorter Context**: LLMs can read a single "System" or "Manager" file without needing the entire 800-line `game.go`.
- **Predictable API**: `Update(ctx)` is a consistent pattern.
- **Isolated State**: It's clear what data a function can mutate (anything in `World`).
- **Scalability**: New features can be added as new files/managers without cluttering core files.
