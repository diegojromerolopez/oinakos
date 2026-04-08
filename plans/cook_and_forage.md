# Plan: Cooking & Foraging Dynamics 🍖🍎

This plan outlines the implementation of cooking mechanics for raw food and foraging activities in Oinakos.

## 1. Objectives
-   **Cooking**: Allow players to transform raw ingredients (like `raw_meat`) into more nutritious cooked food (like `cooked_meat`) using campfires.
-   **Foraging**: Enable players to gather fruits, berries, and vegetables from trees and bushes using the `F` key.
-   **Nutrition**: Eating food (raw or cooked) replenishes health and energy, with cooked food providing significantly better bonuses.

## 2. New Objects & Assets
We need to define new items in `data/objects/` and generate corresponding 160x160 sprites with chroma-key green (`#00FF00`) backgrounds.

| Item ID | Name | Type | Description |
| :--- | :--- | :--- | :--- |
| `cooked_meat` | Cooked Meat | consumable | Nutritious and savory. Restores high Health and moderate Energy. |
| `wild_fruit` | Apple | consumable | A sweet wild fruit. Restores small amount of Health and Energy. |
| `wild_berries` | Berries | consumable | Handful of tart berries. Restores tiny amount of Health and Energy. |
| `wild_veg` | Root Vegetable | consumable | Earthy and filling. Restores moderate Health. |

### Asset Generation Requirements:
-   `assets/images/objects/cooked_meat.png`
-   `assets/images/objects/wild_fruit.png`
-   `assets/images/objects/wild_berries.png`
-   `assets/images/objects/wild_veg.png`

## 3. Foraging Mechanics
-   **Action Key**: `F`
-   **Logic**:
    -   Check for proximity to obstacles of type `tree` or `bush`.
    -   If near, play a "Gathering" sound/animation.
    -   Spawn a random item (`wild_fruit`, `wild_berries`, or `wild_veg`) at the character's feet.
    -   Add a small energy cost to foraging to prevent infinite spam without consequence.
    -   (Optional) Add a cooldown or "depletion" state to the tree so it doesn't give infinite fruit immediately.

## 4. Cooking Mechanics
-   **Requirement**: Proximity to a `campfire` obstacle.
-   **Process**:
    -   When near a campfire, pressing a designated key (e.g., `F` or `SPACE`) checks the inventory for `raw_meat`.
    -   Remove 1 `raw_meat` and add 1 `cooked_meat` to the inventory.
    -   Display floating text: `*Cooking...*` followed by `+Cooked Meat`.
    -   Play a "sizzle" sound effect.

## 5. Nutrition System
Update the consumable logic in `internal/game` (actor.go/menu_handler.go):
-   **Energy**: Replenish `Actor.Energy`.
-   **Health**: Replenish `Actor.Health` (Heal).
-   **Stat Comparison**:
    -   `raw_meat`: +2 Health, +15 Energy.
    -   `cooked_meat`: +15 Health, +25 Energy.
    -   `wild_fruit`: +5 Health, +10 Energy.
    -   `wild_berries`: +2 Health, +5 Energy.

## 6. Implementation Steps
1.  **[Assets]** Generate images for all new food items.
2.  **[Data]** Create YAML definitions for `cooked_meat`, `wild_fruit`, `wild_berries`, and `wild_veg`.
3.  **[Code]** Implement `ActorForaging` state and `F` key handler in `character_update.go`.
4.  **[Code]** Add proximity detection for campfires and cooking logic.
5.  **[Code]** Enhance inventory "Use" logic to support `Energy` and `Health` replenishment for non-permanent consumables.
