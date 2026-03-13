# Dynamic Audio System: Adaptive Music & SFX Polish

## Overview
Music is a key component to an atmospheric, immersive experience like classical late 90s RPGs. Currently, Oinakos supports basic, localized SFX, but lacks a musical soundscape. The game needs an adaptive soundtrack that matches the escalating tension of infinite exploration.

## Implementation Steps
1. **Audio Player Overhaul**:
    - Update `internal/engine/audio_engine.go` (and `internal/game/audio.go`) to support background streaming in Ebitengine, utilizing `audio.NewPlayer(audioContext, decodedFile)`.
    - Create an `AudioManager` that accepts `PlayBGM(path)` and handles looping audio using Ebiten's `InfiniteLoop`.
2. **State-Driven Crossfades**:
    - Add state logic to the game update loop: `Exploring` vs `In Combat`.
    - `In Combat` is triggered when `len(g.npcs) > 0` and an NPC is actively tracking or attacking the player.
    - Implement a `FadeOut(duration)` and `FadeIn(duration)` function on the audio players to allow smooth crossfading between an ambient "exploration" track and a rhythmic "combat" track.
3. **Sound Effect Spatialization (Juice)**:
    - Instead of playing the same *wack* sound uniformly, adjust the volume based on the distance between the sound source (NPC) and the camera center (Player).
    - If a goblin attacks off-screen, the sound should be barely audible; if an orc strikes point-blank, it should roar at 100% volume.
4. **Impact Sounds Variations**:
    - To prevent ear fatigue, add 3-4 variations of sword clangs and flesh impacts (e.g., `hit1.wav`, `hit2.wav`, `hit3.wav`).
    - The `AudioManager.PlaySFX()` function should randomly select one of the variations and slightly randomize the pitch (±10%) on every hit.

## Challenges
- Pitch shifting and spatial panning are non-trivial in pure Go without external C libraries. Ebiten supports raw waveform manipulation, but calculating real-time 3D panning requires custom math wrappers.
- Memory management: BGM must be streamed from disk (`os.Open`), not loaded fully into memory like short SFX files, to avoid ballooning RAM usage.
