package game

import (
	"strings"
	"testing"
)

func TestActor_StrangeMood(t *testing.T) {
	ctx := NewTestContext()
	a := NewCharacter(0, 0, nil, 1, false, nil)
	a.Name = "Artisan"
	a.Mood = MoodNeutral

	t.Run("transition to strange mood", func(t *testing.T) {
		// Manually trigger
		a.Mood = MoodStrange
		a.WorkTicks = 0
		
		a.updateMood(ctx)
		if a.Mood != MoodStrange {
			t.Errorf("expected to stay in Strange Mood, got %v", a.Mood)
		}
		if a.LastAIReasoning == "" {
			t.Error("expected AI reasoning to be set for strange mood")
		}
	})

	t.Run("melancholy timeout", func(t *testing.T) {
		a.Mood = MoodStrange
		a.WorkTicks = 36001 // Past timeout
		
		a.updateMood(ctx)
		if a.Mood != MoodMelancholy {
			t.Errorf("expected transition to Melancholy, got %v", a.Mood)
		}
	})

	t.Run("artifact creation", func(t *testing.T) {
		a.Mood = MoodStrange
		a.WorkTicks = 100
		
		// Reset inventory
		a.Inventory = nil
		
		// Mock materials: Wood and Bone
		woodCfg := &ObjectConfig{ID: "pine_wood", Name: "Pine Wood"}
		boneCfg := &ObjectConfig{ID: "large_bone", Name: "Large Bone"}
		
		a.Inventory = append(a.Inventory, &ItemInstance{ID: "w1", Config: woodCfg})
		a.Inventory = append(a.Inventory, &ItemInstance{ID: "b1", Config: boneCfg})
		
		// Setup registry for artifact base
		ctx.Registries.Objects.Objects["iron_sword"] = &ObjectConfig{ID: "iron_sword", Name: "Iron Sword", Resistance: 100}
		ctx.Registries.Objects.Objects["bone_amulet"] = &ObjectConfig{ID: "bone_amulet", Name: "Bone Amulet", Resistance: 100}

		a.updateMood(ctx)
		
		if a.Mood != MoodNeutral {
			t.Errorf("expected return to Neutral mood after success, got %v", a.Mood)
		}
		
		foundArtifact := false
		for _, it := range a.Inventory {
			if strings.HasPrefix(it.ID, "artifact_") {
				foundArtifact = true
				if it.Resistance <= 100 {
					t.Errorf("artifact resistance should be boosted, got %d", it.Resistance)
				}
				break
			}
		}
		
		if !foundArtifact {
			t.Error("expected artifact to be in inventory")
		}
	})
}
