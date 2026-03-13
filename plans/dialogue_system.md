# Plan: Dialogue & Event Log System 🗣️📄

## Objective
To implement a dynamic, deterministic dialogue system that allows the player to interact with NPCs through a dedicated UI. This system serves as a **lightweight, hardcoded alternative to the `llm_npc_integration.md` plan**, focusing on predictable branching paths and YAML-driven outcomes.

## Analysis
The game needs a way to transition from "Combat/Exploration" mode to "Dialogue" mode. This involves a UI component that persists at the bottom of the screen and a data-driven system to handle branching conversations.

### Key Requirements:
1.  **Dual-Purpose System**: The window serves as both a **Dialogue Interface** and a persistent **Event Log**.
2.  **Bottom UI Box**: A maximizable/minimizable panel at the bottom of the window.
3.  **Persistent Log**: Shows everything happening to the player (attack, hit, damage, recovery, etc.) in real-time.
4.  **Interaction Triggers**:
    - **Click-to-Talk**: Manual interaction by clicking on Neutral or Allied NPCs.
    - **Proximity/Initiator**: Some NPCs can automatically start a conversation when the player gets close.
5.  **Non-Blocking Logic**: The game world does not pause during dialogue or when viewing the log.
6.  **Visual Distinction**: Use different colors for Playable Character text vs. others/events.
7.  **Pseudo-Intelligence**: Probability-based variations for initial greeting texts.
8.  **Closing/Minimizing**: The dialog box or specific conversation can be closed/minimized at any time.

---

## Implementation Details

### 1. Data Structures (`internal/game/dialogue.go`)
-   **`DialogueRoot`**: The entry point for an NPC's dialogue. Contains `PlayerGreetings` (array of random strings) and `StartScenarios` (weighted list of responses).
-   **`DialogueNode`**: A standard interaction node containing the NPC's text and a list of `Choice` objects.
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
-   Add `ActiveDialogue *DialogueState` and `EventLog []LogEntry` to the `Game` struct.
-   `LogEntry` stores the text, timestamp (tick), and category (for coloring).
-   `DialogueState` tracks the current node, the speaker NPC, and the UI state (Maximized/Minimized).
-   **Dialogue Triggers**:
    - **Manual**: Detected via left-clicks on Neutral/Ally NPCs.
    - **Automatic (Proximity)**: The game tick loop checks distance to NPCs with the `auto_initiate` flag. If within `proximity_range`, the dialogue starts automatically.
    - **Hierarchy**: Manual click overrides proximity; proximity triggers only fire once per NPC session to avoid infinite loops.
-   **Non-Blocking**: Ensure game update loop continues running while dialogue is active.

### 4. UI: The Dialogue & Log Box (`internal/game/game_render.go`)
-   Implement `drawDialogueBox(screen engine.Image)`:
    -   **Position**: Bottom of the window.
    -   **Minimized State**: A slim unobtrusive bar (shows the last log entry or "Click NPC to talk").
    -   **Maximized State**: A larger panel showing a scrollable list of recent events and active dialogue.
    -   **Text Rendering**:
        - **Player Actions/Speech**: Blue or Green.
        - **NPC Speech**: White or Light Grey.
        - **Combat/World Events**: Red (Damage), Yellow (Recovery).
    -   **Choices**: A selectable list with a cursor (`>`) when in dialogue mode.
    -   **Controls**: A "Close" button ($X$) or key toggle (e.g., `ESC` or `BACKSPACE`) to dismiss active dialogue without closing the whole log.

---

## Intelligence & Variety
-   **Randomized Player Greetings**:
    - When clicking an NPC, the game picks a random greeting from the `player_greetings` list in the YAML.
    - This greeting is immediately printed to the log in the player's color.
-   **Probabilistic NPC Starts**:
    - The NPC then selects a response from `start_scenarios` based on weighted weights (e.g., 70% chance for "Friendly", 30% for "Auspicious").
    - The chosen response determines the branching path of the conversation.
-   **Contextual Events**: The event log hook should catch:
    - Player attacking: "You swung your weapon."
    - Player hit: "You took [X] damage from [Entity]."
    - Recovery: "You felt replenished."
-   **No Combat Pausing**: Enemies can still attack you while you are reading the log or talking! The UI must remain responsive.
-   **Procedural Content**: For complex or "crazy" characters, see [markov_chain_dialogues.md](file:///Users/diegoj/repos/oinakos/plans/markov_chain_dialogues.md) for generating text from a corpus.

---

---

## Probabilistic Example (`data/npcs/virculus.yaml`)
```yaml
id: "virculus"
name: "Virculus"
archetype: "merchant_male"
dialogues:
  player_greetings:
    - "Greetings, traveler!"
    - "Excuse me, do you have a moment?"
    - "Hey there!"
  
  start_scenarios:
    - weight: 0.8
      text: "Ah, a customer! Welcome to my humble patch of dirt. What can I do for you?"
      next: "main_menu"
    - weight: 0.2
      text: "I'm busy... but since you look like you have gold, I'll listen."
      next: "grumpy_start"

  main_menu:
    text: "I have the finest trinkets from the eastern wastes."
    choices:
      - text: "Show me your wares."
        next: "trade"
      - text: "Just passing through."
        next: "exit"

  grumpy_start:
    text: "Well? Speak up. I don't have all day."
    choices:
      - text: "Sorry to bother you. I'll leave."
        next: "exit"
      - text: "I'm looking for information."
        next: "intel_path"
```

---

## Embedded YAML Initiator Example (`data/npcs/stultus.yaml`)
```yaml
id: "stultus"
name: "Stultus"
archetype: "villager_male"
dialogues:
  start_scenarios:
    - weight: 1.0
      text: "You there! Have you seen the red-eyed knight passing through?"
      auto_initiate: true
      proximity_range: 3.5
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
-   **Log Persistence**: Verify that taking damage or attacking immediately prints a line to the bottom box, even when dialogue is not active.
-   **Click-to-Talk**: Click on a friendly NPC and verify the dialogue starts. Click on an Enemy and verify nothing happens.
-   **Auto-Initiation**: Approach an NPC with `auto_initiate: true` and verify the dialogue box pops up automatically when you enter their range.
-   **Color Coding**: Ensure damage numbers are colored differently from NPC speech.
-   **Probability Test**: Start a conversation multiple times with an NPC that has randomized greetings to ensure they vary.
-   **Non-Blocking**: Verify you can still be attacked and move while the dialogue box is maximized.
-   **Toggling**: Verify the box can be minimized back to a status bar without losing the conversation state.
