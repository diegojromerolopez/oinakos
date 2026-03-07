# 🎙️ Oinakos Voice Registry

This document catalogs the specific **Piper TTS voice models** used for every character, unique NPC, and archetype in the game. All models are stored in the `models/` directory.

## 🎭 Main Characters
These characters feature custom dialogue and specific language/accent assignments.

| Character | Language | Voice Model | Description |
| :--- | :--- | :--- | :--- |
| **Boris Stronesco** | Serbian (sr_RS) | `sr_RS-serbski_institut-medium` | A mysterious noble with a deep, official tone. |
| **Roland** | French (fr_FR) | `fr_FR-tom-medium` | A martial, disciplined paladin voice. |
| **Conde Olinos** | Spanish (es_ES) | `es_ES-davefx-medium` | A tragic noble with a clear, resonant tone. |
| **Durandarte** | Spanish (es_ES) | `es_ES-carlfm-x_low` | A knightly voice with a slightly rougher texture. |
| **Gaiferos** | Spanish (es_ES) | `es_ES-sharvard-medium` (Spkr 0) | An articulate, standard Spanish male voice. |
| **Montesinos** | Spanish (es_ES) | `es_ES-mls_9972-low` | A deep, grounded voice for the mountain knight. |
| **Oinakos** | English (en_US) | `en_US-john-medium` | A deep, resonant, and authoritative noble knight. |

## 👥 Unique NPCs
Unique characters encountered in the world with distinct personalities.

| NPC | Language | Voice Model | Description |
| :--- | :--- | :--- | :--- |
| **Stultus** | English (en_US) | `en_US-hfc_male-medium` | Intense and powerful, suitable for shouting attacks. |
| **Marcus Ardea** | English (en_US) | `en_US-lessac-low` | Grounded, simple, and slightly gravelly. |
| **Peasant King** | English (en_US) | `en_US-ljspeech-medium` | Charismatic and persuasive narrative voice. |
| **Virculus** | English (en_US) | `en_US-lessac-high` | Clean, neutral, and consistent automaton response. |
| **Tragantia** | Spanish (es_ES) | `es_ES-mls_10246-low` | Sibilant/hissing quality for a reptile-woman. |

## 🛡️ Archetypes
Automatic voice variants used for generic mob behavior.

| Archetype | Gender | Voice Model | Characteristics |
| :--- | :--- | :--- | :--- |
| **Demon** | Female | `en_US-kathleen-low` | Hollow and ominous. |
| | Male | `en_US-bryce-medium` | Deep and sinister. |
| **Giant** | Female | `en_GB-cori-high` | Grand and authoritative. |
| | Male | `en_US-ryan-high` | Booming and powerful. |
| **Orc** | Female | `en_GB-jenny_dioco-medium` | Rugged and rough. |
| | Male | `en_GB-northern_english_male-medium`| Gritty warrior. |
| **Magi** | Female | `en_US-ljspeech-high` | Wise and articulate. |
| | Male | `en_US-libritts-high` | Precise and academic. |
| **Goblin** | Female | `en_US-amy-low` | High-pitched and shrill. |
| | Male | `en_US-danny-low` | Raspy and devious. |
| **Peasant** | Female | `en_GB-southern_english_female-low` | Simple and rustic. |
| | Male | `en_GB-alan-medium` | Friendly townfolk. |
| **Slave** | Female | `en_GB-alba-medium` | Tired and soft. |
| | Male | `en_US-norman-medium` | Exhausted and deep. |
| **Man-at-Arms**| Male | `en_US-joe-medium` | Direct and soldierly. |
| **Lame Devil** | Female | `en_US-hfc_female-medium` | Sharp and intense. |
| | Male | `en_US-kusal-medium` | Sly and cunning. |
| **Trasgo** | Male | `en_US-arctic-medium` | Clean and digital. |
| **Mythical** | Female | `en_US-kristin-medium` | Narrative and clear. |
| | Male | `en_US-reza_ibrahim-medium` | Otherworldly. |


## 📜 Voice Origin & Licensing
All voice assets in Oinakos are generated using the **[Rhasspy Piper TTS](https://github.com/rhasspy/piper)** engine. The voice models themselves are sourced from the **[rhasspy/piper-voices](https://huggingface.co/rhasspy/piper-voices)** repository on Hugging Face.

### Datasets & Permissions
The models are trained on various open-source datasets, which are generally provided under permissive licenses suitable for open-source development:
- **LibriTTS / VCTK**: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- **M-AILABS**: BSD-3-Clause
- **LJSpeech**: Public Domain ([CC0](https://creativecommons.org/publicdomain/zero/1.0/))
- **HFC (Hugging Face Community)**: MIT or CC-BY-4.0

*For specific license details for a particular voice, refer to the `MODEL_CARD` file in the `models/` directory.*

## ⚙️ Technical: NPC Audio Fallback
Oinakos handles NPC audio using a hierarchical fallback system to ensure every entity has a voice even without unique assets:

1.  **Unique Override**: The engine first looks for WAV files in `assets/audio/npcs/<npc_id>/`.
2.  **Archetype Inheritance**: If overrides aren't found, the NPC inherits its audio path from its base Archetype (e.g., `assets/audio/archetypes/orc/male/`).
3.  **Main Character**: The player character always uses `EntityConfig.MainCharacter` as the audio prefix (e.g., `boris_stronesco/attack`), mapped directly to `assets/audio/characters/<id>/`.

---
*Note: All generation is handled via Piper TTS using raw output (no post-processing filters).*
