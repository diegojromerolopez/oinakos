package game

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewMapTypeRegistry(t *testing.T) {
	r := NewMapTypeRegistry()
	if r.Types == nil {
		t.Error("Types map should not be nil")
	}
}

func TestNewArchetypeRegistry(t *testing.T) {
	r := NewArchetypeRegistry()
	if r.Archetypes == nil {
		t.Error("Archetypes map should not be nil")
	}
}

func TestNewObstacleRegistry(t *testing.T) {
	r := NewObstacleRegistry()
	if r.Archetypes == nil {
		t.Error("Archetypes map should not be nil")
	}
}

func TestObjectiveTypeUnmarshalYAML(t *testing.T) {
	var ot ObjectiveType

	data := "kill_vip"
	if err := yaml.Unmarshal([]byte(data), &ot); err != nil {
		t.Errorf("Failed to unmarshal kill_vip: %v", err)
	}
	if ot != ObjKillVIP {
		t.Errorf("Expected ObjKillVIP, got %v", ot)
	}

	data = "99"
	if err := yaml.Unmarshal([]byte(data), &ot); err != nil {
		t.Errorf("Failed to unmarshal 99: %v", err)
	}
	if ot != ObjectiveType(99) {
		t.Errorf("Expected 99, got %v", ot)
	}

	data = "unknown"
	if err := yaml.Unmarshal([]byte(data), &ot); err == nil {
		t.Error("Expected error for unknown objective type")
	} else if !strings.Contains(err.Error(), "unknown objective type") {
		t.Errorf("Unexpected error message: %v", err)
	}
}
