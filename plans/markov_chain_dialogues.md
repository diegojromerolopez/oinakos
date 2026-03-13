# Plan: Markov Chain Dialogue Generation ⛓️💬

## Objective
To enhance the `dialogue_system.md` by allowing NPCs to generate procedural, flavor-rich responses using Markov chains. This allows for high variety in repetitive interactions (like "Crazy Hermit" talk or "Street Rumors") without manually writing dozens of static string variations.

## Analysis
Static dialogue trees can feel repetitive. While the probabilistic starts in the main plan help, Markov chains provide a way to generate "novel" sentences that maintain the tone and vocabulary of a specific character by recombining fragments of a provided training text (corpus).

### Key Requirements:
1.  **Lightweight Markov Engine**: A simple Go implementation that builds a frequency-based transition table from a string.
2.  **Corpus-Driven YAML**: Add a `markov_corpus` field to dialogue nodes.
3.  **Dynamic Generation**: If a `markov_corpus` is present, generate a sentence of a specific length instead of using a static `text` field.
4.  **Seed Persistence**: (Optional) Allow the "feeling" of the chain to stay consistent for a session.

---

## Implementation Details

### 1. The Markov Engine (`internal/game/markov.go`)
- Create a `MarkovChain` struct:
    ```go
    type MarkovChain struct {
        Transitions map[string][]string
        Order       int // Usually 1 or 2 (Bigrams/Trigrams)
    }
    ```
- **Train**: Function to split input text into tokens and build the map.
- **Generate**: Function to pick a starting word and walk the chain until a sentence-ending punctuation or length limit is reached.

### 2. Integration with Dialogue System
Modify `DialogueNode` or `StartScenario` in `internal/game/dialogue.go`:
- Add `MarkovCorpus string` field.
- Add `MarkovLength int` (number of words to generate).
- During retrieval:
    - If `Node.Text` is empty and `Node.MarkovCorpus` is not, run the generator.
    - Cache the trained chain in memory to avoid re-parsing the large text on every interaction.

### 3. Usage in YAML (`data/npcs/hermit.go`)
Instead of a single line, provide a block of text that defines the character's "vocabulary space."

```yaml
id: "crazy_hermit"
name: "The Hermit"
dialogues:
  start_scenarios:
    - weight: 1.0
      markov_corpus: |
        The stars are bleeding tonight. I saw the knight with red eyes dancing in the fog. 
        The fog hungry. It ate the birds and the trees and the secrets of the cliff side.
        Do not go to the cliffs unless you want to be eaten by the red-eyed knight.
        Bleeding stars tell no lies. The trees are watching you.
      markov_length: 12
      choices:
        - text: "I'll be careful..."
          next: "exit"
```

---

## Behavior & Intelligence
- **Seeding**: The chain can be seeded with the player's last chosen response to make it feel slightly more "reactive," though it will mostly remain nonsensical/flavorful.
- **Stateful Chains**: Some NPCs could have "Learning" corpuses that grow as the game progresses (e.g., if you complete a quest, a sentence about your victory is added to the general "Rumors" Markov chain).

---

## Verification
- **Variety**: Interact with a Markov-enabled NPC 10 times and verify that the text is rarely identical.
- **Tone Consistency**: Ensure the generated text uses only words found in the training corpus.
- **Performance**: Ensure that training a 2KB corpus happens in under 1ms and doesn't hitch the game loop.
- **Fallback**: Ensure that if the generator fails to find a valid path, it falls back to a safe "..." or a generic static string.
