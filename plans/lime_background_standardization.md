# Standardization of Image Backgrounds to Lime Green

The current `Transparentize` function removes white/near-white colors, causing unexpected transparencies inside sprites (e.g., eyes, teeth, clothes becoming transparent). To fix this, we need to enforce a standard "lime green" background for all sprites instead of white ones.

## Proposed Changes
### Asset Updates
- Identify all `.png` images in `assets/images` that currently use a white background.
- Replace these white backgrounds with a solid lime-green background.

### Engine Modifications
#### [MODIFY] [collision.go](file:///Users/diegoj/repos/oinakos/internal/engine/collision.go)
- Remove the block of code inside `Transparentize` that explicitly zeroes out literal white/near-white pixels (`r > whiteThreshold && g > whiteThreshold && b > whiteThreshold`). This ensures white colors inside the character sprites are preserved.
- The existing logic that identifies edge colors and lime green colors will naturally handle the new lime green backgrounds without false positives on internal sprite colors.
