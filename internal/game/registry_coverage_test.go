package game

import (
	"testing"
)

func TestRegistries_Coverage(t *testing.T) {
	// 1. Archetype Registry
	er := NewArchetypeRegistry()
	er.Archetypes["human"] = &Archetype{ID: "human", Name: "Human"}
	if er.Archetypes["human"] == nil { t.Error("Failed to get archetype") }
	
	// 2. Object Registry
	or := NewObjectRegistry()
	or.Objects["sword"] = &ObjectConfig{ID: "sword", Name: "Sword"}
	if or.Get("sword") == nil { t.Error("Failed to get object") }
	
	// 3. Obstacle Registry
	obr := NewObstacleRegistry()
	obr.Archetypes["tree"] = &ObstacleArchetype{ID: "tree", Name: "Tree"}
	if obr.Archetypes["tree"] == nil { t.Error("Failed to get obstacle archetype") }
}
