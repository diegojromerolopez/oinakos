# UI Overhaul & HUD Polish

## Overview
The current UI relies on basic debug-print text utilizing standard fonts. We need to elevate the aesthetic to match the dark, medieval RPG style (Hades/Diablo inspired) of the actual gameplay.

## Implementation Steps
1. **Asset Integration**:
    - Generate textures for: Health Bar (Empty, Fill, Frame), Ability Cooldown frames, and an ornate HUD background.
2. **Custom Text Rendering**:
    - Replace `textRenderer.DebugPrintAt` with a custom bitmap font loader, or load a `.ttf` file (like a Gothic font) using `golang.org/x/image/font`.
3. **Entity Status Bars**:
    - Create a floating health bar system for NPCs and the Player. 
    - In `npc_render.go` and `playable_character_render.go`, draw the red health fill proportional to `Health / MaxHealth` overlaid on a dark frame sprite.
4. **Main HUD Canvas**:
    - Create a dedicated `DrawHUD()` phase in the render loop.
    - Display the Player's profile portrait, structured HP/XP bars, and the current Kill Count in an aesthetic frame anchored to the screen edges.

## Challenges
- Ensuring UI elements do not scale or warp when the camera zooms or the window is resized.
- Creating pixel-perfect text that doesn't blur at 1080p.
