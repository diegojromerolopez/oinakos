# Plan: Debate-Incite Content & Moral Dilemmas ⚖️

## Objective
To introduce narrative depth and moral complexity that forces the player to make difficult choices, causing conflict between NPCs and sparking "debate" (internal or within the game's community) about the "right" path.

## Analysis
The current world is binary: Allied or Enemy. To incite debate, we must introduce **Moral Gray Areas**. This is achieved by making factions have valid but conflicting goals, and by making the player's survival dependent on potentially unethical choices.

### Key Pillars of Debate:
1. **The Cost of Security**: Is "Enslaving" or "Torturing" captured NPCs (from the Capture Plan) justified if it gives the village better defenses against an imminent Orc invasion?
2. **Faction Rivalries**: Two allied factions with different philosophies (e.g., The Church of Light vs The Order of Steel).
3. **Sacrifice of the Few**: Decisions where an "Objective" requires sacrificing an NPC the player has grown to like.
4. **Historical Revisionism**: NPCs who claim the "Heroes" were actually the original aggressors.

---

## Implementation Details

### 1. Faction Ideologies (`data/archetypes/`)
- Define distinct sub-factions within the "Ally" alignment:
    - **Traditionalists**: Want to preserve the old village ways, even if it means sticking to weak wooden walls.
    - **Modernists/Pragmatists**: Want to use captured orcs (Slaves) and industrial tools to build iron fortifications, even if it "pollutes" or "corrupts" the land.
- Create specific NPCs from each faction that constantly argue in the "Safe Zone."

### 2. The Rationing System (Map Objective)
- Create a "Siege" map type where resources (Wells/Food Caches) are finite.
- The player must choose:
    - **Give extra water to the Guards**: They fight better, but the Civilians start losing HP.
    - **Give water to the Civilians**: The Guards might desert or fail to hold the line.

### 3. The "Captive's Dilemma" (Interaction)
- Combine with the **Death Alternatives** plan. 
- When an NPC is captured:
    - An NPC (The Inquisitor) urges you to **Torture** them for information (reveals enemy spawn points on the map).
    - An NPC (The Priest) urges you to **Release** them (gives a "Mercy" buff/reputation but makes the next wave harder).
- Display the consequences through **Floating Text** or subsequent map events.

### 4. Environmental Narrative
- Add "Scrolls" or "Tattered Notes" as Obstacles that provide contradicting lore.
- Note A: "The King saved us from the Orcs."
- Note B: "The King burned the Orc villages first to steal their gold."

---

## Content Examples for Debate

| Choice | Pro | Con |
| :--- | :--- | :--- |
| **Enslave Captured Orcs** | Faster construction of `barricades` and `walls`. | Moral corruption; possibility of a slave revolt. |
| **Defend the Sacred Grove** | Gain "Nature's Favor" (Healing regen). | Lose the village's only source of lumber for arrows. |
| **Execute the Traitor** | Maintains order and discipline in the camp. | They were only stealing to feed their starving family. |

---

## Verification
- Load a "Village" map and witness a scripted argument between two NPCs about a specific policy (e.g., the use of slaves).
- Complete a quest by choosing the "Cruel but Efficient" path and see how the village visual state changes (e.g., more iron, but more misery/squalor).
- Verify that choosing one faction's side lowers the player's "Loyalty" stat with the other.
