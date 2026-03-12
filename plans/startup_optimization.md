# Startup Performance Optimization Plan ⚡️🚀

This plan identifies the bottlenecks in the Oinakos startup sequence and proposes a transition from eager loading to a hybrid lazy-loading architecture.

---

## 🔍 Bottleneck Analysis

Based on logs and code review, the following operations are the primary contributors to slow startup times:

1. **Eager Sprite Loading**: `GameRenderer.LoadAssets` currently iterates through the entire `ArchetypeRegistry`, `NPCRegistry`, `ObstacleRegistry`, and `PlayableCharacterRegistry`. For each entry, it performs multiple `fs.Stat` calls and loads up to 10 PNGs per entity (static, back, attack, etc.), regardless of whether that entity is present on the map.
2. **Audio Registry Scanning**: `registerEntitySounds` in `main.go` scans directories for every entity in every registry. This is a high-latency I/O operation when many small WAV files are involved.
3. **Serial Asset Loading**: All assets are loaded sequentially on the main thread, blocking the application from showing even the Main Menu until everything is processed.
4. **Redundant Palette Checks**: Shaders and color masks are initialized multiple times during the registry load.

---

## 🛠 Proposed Solutions

### 1. Hybrid Lazy Loading (Priority: High)
Move away from "Load all assets at start" to a system where:
- **Core Assets**: Only the Main Menu, HUD, and the Player's base archetype are loaded at startup.
- **On-Demand Entity Loading**: NPCs and Archetypes check if their sprites are loaded in their `Draw` call or during `NewNPC`. If not, they trigger a load.
- **Background Loading**: Start a background goroutine after the Main Menu appears to continue loading "likely used" assets.

### 2. Asset Manifests (Priority: Medium)
Instead of calling `fs.Stat` and `fs.ReadDir` repeatedly:
- Generate an `assets.json` manifest during the build process (or a small script).
- The game reads this manifest once, knowing exactly which files exist without probing the filesystem.

### 3. Audio Metadata Caching (Priority: Medium)
Modify `AudioManager` to only "Register" sounds at startup (knowing they exist) but only "Load" the actual pulse/buffer when the sound is first triggered.

### 4. Parallel Asset Processing (Priority: High)
Refactor `Registry.LoadAssets` to use `Worker Pools`:
- Use `sync.WaitGroup` and a limited number of goroutines (e.g., `runtime.NumCPU()`) to load PNG chunks in parallel.
- Ebitengine images must be created on the main thread, but PNG decoding can be offloaded.

---

## 📅 Roadmap

### Phase 1: Quick Wins
- [ ] Implement lazy loading for `Archetype.StaticImage`.
- [ ] Move HUD and Menu assets to a separate `InterfaceAssets` struct that loads first.
- [ ] Remove redundant/excessive debug logging during registry loads.

### Phase 2: Structural Changes
- [ ] Implement a `SpriteCache` that handles loading and deduplication.
- [ ] Offload PNG decoding to goroutines.
- [ ] Implement the `on-demand` loading trigger in `NPC.Draw`.

### Phase 3: I/O Optimization
- [ ] Implement the asset manifest system to eliminate `fs.Stat` calls.
- [ ] Refactor audio registration to avoid directory walks on the main thread.

---

## 📉 Expected Impact
- **Startup Time**: Projected reduction from ~5-7 seconds to **<1.5 seconds**.
- **Memory Usage**: Significant reduction as only the assets for the current map levels are held in GPU memory.
