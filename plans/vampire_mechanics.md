# Vampire Conversion Mechanics Plan

This document outlines the implementation of the "vampire conversion" logic, where humans killed by vampires are resurrected as allied vampires.

## 🩸 Core Concept
Vampires represent a spreading plague. They cannot be naturally spawned in some maps but instead convert existing human NPCs (peasants, guards, slaves) into new vampires upon death.

## 🛠 Technical Implementation

### 1. Archetype Identification
- New archetypes added: `vampire_male`, `vampire_female`.
- We need a way to identify "Convertible Humans". This can be done by checking if the killed NPC's archetype belongs to a specific list or has a `human: true` flag in YAML.
- **Initial list**: `peasant`, `slave`, `man_at_arms`.

### 2. Combat Logic Update (`internal/game/npc.go`)
The `TakeDamage` function (or a separate death handler) should be modified:

```go
// In npc.go: TakeDamage
if n.Health <= 0 {
    // ... existing death logic ...
    
    if attackerNPC != nil && isVampire(attackerNPC.Archetype.ID) {
        // Roll for infection
        prob := attackerNPC.Archetype.Stats.InfectingProbability
        if rand.Float64() < prob && isConvertibleHuman(n.Archetype.ID) {
            // Trigger conversion
            g.QueueEvent(VampireConversionEvent{
                Position: engine.Point{X: n.X, Y: n.Y},
                Gender: n.Archetype.Gender,
                Master: attackerNPC,
            })
        }
    }
}
```

### 3. Conversion Event
A new event or queue-based mechanism to:
1.  Remove the dead human NPC.
2.  Spawn a new Vampire NPC at the same location.
3.  Set the new Vampire's **Alignment** and **Group** to match the "Master" vampire who killed it.
4.  Optionally play a "Conversion" sound effect or particle effect.

### 4. Behavior Adjustments
- Vampires should prioritize humans as targets.
- A new `BehaviorVampire` could be implemented to seek out clusters of humans.

## ✅ Implementation Todo
- [x] Create Vampire Archetypes (Male/Female).
- [x] Generate and integrate high-quality Vampire assets (Static, Attack, Corpse).
- [x] Add `infecting_probability` to registries and YAMLs.
- [ ] Implement `isVampire(id string)` and `isConvertibleHuman(id string)` helpers.
- [ ] Modify `internal/game/npc.go` to handle the conversion trigger with probability roll.
- [ ] Ensure the conversion only happens for "original" humans (avoiding double-conversion or complex loops).
- [ ] Test the chain reaction (one vampire converting a village).

## 🌑 Aesthetic Goals
- Conversion should feel dark and immediate.
- The new vampire should inherit the name of the human if possible, adding a "the Risen" or similar suffix.
