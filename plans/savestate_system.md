# Savestate: JSON Serialization

## Overview
Currently, the game only supports basic configuration loading. We need the ability to save the actual ongoing state of an infinite world run (Player stats, position, kills) and load back into it.

## Implementation Steps
1. **State Structs**:
    - Define a master `GameStateSave` struct in `internal/game/save.go` containing:
        - `PlayerState` (X, Y, HP, XP, Equipment, Kills)
        - `WorldState` (Generated Chunks, destroyed obstacles)
2. **Serialization**:
    - Use Go's built-in `encoding/json`.
    - Implement a `SaveGame()` function that maps the current active `game.mainCharacter` and `game.chunks` into the `GameStateSave` struct, then writes to `data/saves/slot_1.json`.
3. **Deserialization**:
    - Implement a `LoadGame()` function attached to the main menu.
    - It will parse the JSON, instantiate a new `MainCharacter` with the parsed stats, and pre-populate the `chunks` map so the player loads back into the exact world they left.
4. **Autosave**:
    - Hook `SaveGame()` to trigger automatically when transitioning maps, or periodically (every 5 minutes) during endless exploration.

## Challenges
- Storing the state of every chunk could result in massive file sizes if the player explores far. We may need to compress the chunk data or only save modified chunks (e.g., where a rock was destroyed), regenerating untouched chunks from the seed.
