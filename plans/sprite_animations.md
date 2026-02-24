# Plan: Sprite Sheet Animations

## Objective
Improve the animations of the Knight and NPCs by transitioning from static single-frame images to proper sprite sheets, allowing for smooth walking and attacking animations.

## Analysis
Currently, the game loads individual images for states (e.g., `knight_static.png`, `knight_attack.png`). By consolidating these into a single sprite sheet per character, we can slice the image into uniform "frames" and cycle through them based on the game's tick timer.

## Implementation Steps

### 1. Asset Preparation
- Obtain or generate uniform sprite sheets for each character (Knight, Orc, Demon, Peasant).
- **Structure**: Ideally, the sprite sheet should be organized in a grid.
  - Row 0: Idle/Static (1 frame)
  - Row 1: Walk Cycle (e.g., 4 or 6 frames)
  - Row 2: Attack Cycle (e.g., 3 frames)
  - Row 3: Death/Corpse (1 frame)
- **Constraint Calculation**: Define the `frameWidth` and `frameHeight` constants for each character type.

### 2. Engine Module: Sprite Animator
- Create a new struct in `internal/engine/animator.go` (or similar) to handle generic sprite slicing.
```go
type Animator struct {
    Sheet       *ebiten.Image
    FrameW      int
    FrameH      int
    FrameCount  int
    CurrentTick int
    Speed       int
}
func (a *Animator) GetFrame(state string, tick int) *ebiten.Image
```

### 3. Integrate into Game Entities
- Update `NPC` and `Player` structs in `internal/game/` to hold an `Animator` instead of multiple `*ebiten.Image` pointers.
- Modify the `Draw` methods:
  - Calculate the current frame index: `currentFrame := (tick / animationSpeed) % maxFramesForState`
  - Slice the main sprite sheet using `ebiten.Image.SubImage`:
    ```go
    rect := image.Rect(currentFrame*frameWidth, row*frameHeight, (currentFrame+1)*frameWidth, (row+1)*frameHeight)
    frameObj := sheet.SubImage(rect).(*ebiten.Image)
    ```
  - Draw `frameObj` instead of the static image.

### 4. Retain Directional Mirroring
- The existing logic that flips the sprite horizontally based on `Facing == DirSE || DirNE` will remain perfectly intact and will automatically apply to the animated frames.

## Verification
- Walk the Knight around and observe the leg cycling.
- Press Space to attack and observe the multi-frame swing.
- Ensure NPCs loop their walk cycles seamlessly while tracking the player.
