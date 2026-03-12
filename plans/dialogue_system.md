# Plan: Dialogue System & Social Influence 🗣️

## Objective
To implement a dynamic, deterministic dialogue system that allows the player to interact with NPCs through a dedicated UI. This system serves as a **lightweight, hardcoded alternative to the `llm_npc_integration.md` plan**, focusing on predictable branching paths and YAML-driven outcomes.

## Analysis
The game needs a way to transition from "Combat/Exploration" mode to "Dialogue" mode. This involves a UI component that persists at the bottom of the screen and a data-driven system to handle branching conversations.

### Key Requirements:
1.  **Dialogue Data**: YAML-based dialogue trees with choices and effects.
2.  **Bottom UI Box**: A maximizable/minimizable panel for text and response selection.
3.  **Minimal AI Integration**: NPCs have "Social States" or "Moods" influenced by dialogue.
4.  **Behavior & Alignment Shifting**: Choices can trigger immediate logic changes (e.g., turning an Enemy into an Ally through persuasion).

---

## Implementation Details

### 1. Data Structures (`internal/game/dialogue.go`)
-   **`DialogueNode`**: Contains the NPC's text and a list of `Choice` objects.
-   **`Choice`**: Contains the player's response text, the ID of the next node, and an optional list of `Effect` triggers.
-   **`Effect`**: A struct defining changes:
    -   `ChangeAlignment`: Set target NPC to `Enemy/Ally/Neutral`.
    -   `ChangeBehavior`: Set target NPC to `Wander/Patrol/Flee/Follow`.
    -   `GrantItem`: (Future) Give weapon/gold.
    -   `SetFlag`: Global game flag for quest tracking.

### 2. Dialogue Integration (`internal/game/config.go`)
- Update `EntityConfig` (used by both Archetypes and NPCs) to include an optional `Dialogues` field: `Dialogues map[string]*DialogueNode`.
- This allows conversation trees to be hardcoded directly into the NPC/Character YAML files.
- The system will attempt to load dialogue from the `EntityConfig` first, falling back to a shared `data/dialogues/*.yaml` registry for generic conversations (like "Standard Villager Talk").

### 3. Game State & Logic (`internal/game/game.go`)
-   Add `ActiveDialogue *DialogueState` to the `Game` struct.
-   `DialogueState` tracks the current node, the speaker NPC, and whether the box is minimized.
-   **Input**: When in dialogue mode, `W/S` or `Up/Down` navigates choices, `Enter` confirms. `TAB` could toggle minimize/maximize.

### 4. UI: The Dialogue Box (`internal/game/game_render.go`)
-   Implement `drawDialogueBox(screen engine.Image)`:
    -   **Position**: Bottom 25% of the screen.
    -   **Minimized State**: A slim bar showing the speaker's name and a "Talk" prompt.
    -   **Maximized State**: A large parchment or dark stone textured box.
    -   **Text Rendering**: Gradual typewriter effect for NPC text (1 character per tick).
    -   **Choices**: A selectable list with a cursor (`>`).

---

## Minimal Dialogue AI
-   **Proximity Trigger**: When the player is near a neutral NPC and presses 'E', the dialogue box maximizes and the conversation begins.
-   **Reactionary Behavior**: NPCs can have a "Patience" or "Fear" meter. If you choose aggressive options, their behavior might flip to `BehaviorFlee` or `BehaviorKnightHunter` instantly.

---

## Embedded YAML Example (`data/npcs/stultus.yaml`)
```yaml
id: "stultus"
name: "Stultus"
archetype: "villager_male"
dialogues:
  start:
    text: "You there! Have you seen the red-eyed knight passing through?"
    choices:
      - text: "I have. He went toward the cliffs."
        next: "cliff_warn"
      - text: "None of your business, old man."
        next: "insult_path"
        effects: [{ type: "change_behavior", value: "flee" }]
  cliff_warn:
    text: "The cliffs? By the gods... that path is cursed."
```

---

## Verification
-   Approach a neutral Villager. Verify that pressing the interaction key opens the bottom dialogue box.
-   Choose a "Hostile" response and verify that the NPC instantly turns red (Enemy alignment) and starts attacking.
-   Choose a "Persuasive" response on a hostile-leaning NPC and verify they become Neutral and stop their pursuit.
-   Test the **Minimization**: Toggle the box with a key and verify it shrinks to a small status bar while the game continues to run in the background.
