# LLM-Powered NPC Integration Plan 🧠🦾

This document outlines the architecture for integrating Large Language Models (LLMs) into Oinakos NPCs to enable dynamic dialogue and autonomous action-taking (movement and combat).

---

## 🎯 Objectives
1. **Dynamic Dialogue**: NPCs can "speak" with the player and each other based on personality, history, and current situation.
2. **Autonomous Actions**: NPCs can decide to follow, flee, wander, or attack based on LLM-driven reasoning rather than static state machines.
3. **Multi-Model Support**: Support for local LLMs (Ollama/LocalAI) and Cloud APIs (Gemini/OpenAI).

---

## 🏗️ Technical Architecture

### 1. The LLM Provider Interface
We will create a provider-agnostic interface in `internal/engine` to handle requests asynchronously.

```go
type LLMProvider interface {
    Query(prompt string) (string, error)
    QueryJSON(prompt string, schema interface{}) (string, error) // For structured actions
}
```

### 2. The "Sensor" System (World Context)
NPCs need to "perceive" the world to make decisions. We will implement a `Perception` struct that serializes local world state into text for the LLM.

**Context includes:**
- **Self**: Name, Archetype, Health, Faction, Position.
- **Environment**: Nearby obstacles (trees, wells), Map bounds.
- **Entities**: Nearby NPCs (friends/foes), Player distance and visible health.
- **History**: Last 3-5 events (e.g., "The player hit me", "I found a well").

### 3. State Additions to NPC Struct
We need to extend `internal/game/npc.go` to store LLM-related state.

```go
type NPC struct {
    Actor
    // ...
    LLMContext    []string // Short-term memory
    LastThought   int      // Tick count of last LLM update
    IsThinking    bool     // ASYNC flag to prevent multiple overlapping requests
    SpeechBubble  string
    SpeechTimer   int
}
```

---

## 💬 Dialogue System (Speaking)

### Implementation Steps:
1. **Archetype Personality**: Add `personality` and `goal` fields to archetype YAMLs in `data/archetypes/`.
2. **Speech Trigger**: Dialogue triggers when the player is close enough AND a timer expires, or if the NPC is hit.
3. **Speech Bubbles**: Implement a rendering system in `internal/game/game_render.go` to show a pixel-perfect gothic bubble above the NPC's head.
4. **TTS Hook**: Trigger `engine.GlobalAudio.PlaySpeech(text)` to use the existing Piper TTS system.

---

## ⚔️ Action System (Movement/Attack)

The LLM will not return raw text, but rather a structured JSON response (or a specific syntax) that maps to game actions.

**Desired Output Format:**
```json
{
  "speech": "Stop right there, traveler!",
  "action": "ATTACK",
  "target": "PLAYER",
  "coordinate": {"x": 10.5, "y": 12.0}
}
```

**Actions to Implement:**
- `IDLE`: Stay put.
- `WANDER`: Move to a random nearby coordinate.
- `FLEE`: Move away from the player.
- `FOLLOW`: Move toward the player (friendly).
- `ATTACK`: Engage the nearest valid target.

---

## 🔄 The Async Reasoning Loop

Since LLMs are slow (ms to seconds), they **must never** run on the main game thread.

1. **Main Loop**: NPC checks `IsThinking`. If false and `Tick - LastThought > ThoughtCooldown`, fire a Goroutine.
2. **Goroutine**:
   - Assemble World Context string.
   - Send to `LLMProvider`.
   - On success: Parse response.
   - Inject commands into a `CommandQueue`.
3. **Main Loop (Next Tick)**: NPC consumes the `CommandQueue` and executes the movement/attack logic.

---

## 📅 Roadmap phases

### Phase 1: Infrastructure
- [ ] Create `internal/engine/llm_provider.go`.
- [ ] Implement a mock provider and an Ollama/Gemini provider.
- [ ] Add `LLMBehavior` to `BehaviorType`.

### Phase 2: Perception & Speaking
- [ ] Implement `AssembleContext()` method for NPCs.
- [ ] Implement the async thinking loop.
- [ ] Render Speech Bubbles above internal actors.
- [ ] Link Piper TTS to LLM output.

### Phase 3: Autonomous Actions
- [ ] Define the JSON Command Schema.
- [ ] Implement the Command Parser.
- [ ] Connect LLM commands to `MoveTo` and `SetState(NPCAttacking)` logic.

---

## ⚠️ Performance Considerations
- **Throttling**: Limit logic updates to once every 2-5 seconds per NPC.
- **Batching**: If many NPCs are present, prioritize those closest to the player.
- **Fallback**: If the LLM times out or fails, the NPC falls back to its standard static Behavior (e.g. `BehaviorKnightHunter`).
