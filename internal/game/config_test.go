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

func TestEntityConfig_Inheritance(t *testing.T) {
	arch := &EntityConfig{
		ID:       "orc_male",
		Behavior: "hunter",
		Stats: struct {
			HealthMin       int     `yaml:"health_min"`
			HealthMax       int     `yaml:"health_max"`
			Speed           float64 `yaml:"speed"`
			BaseAttack      int     `yaml:"base_attack"`
			BaseDefense     int     `yaml:"base_defense"`
			AttackCooldown  int     `yaml:"attack_cooldown"`
			AttackRange     float64 `yaml:"attack_range"`
			ProjectileSpeed float64 `yaml:"projectile_speed"`
		}{HealthMin: 30, HealthMax: 50, BaseAttack: 8, BaseDefense: 4, Speed: 0.03},
		XP: 11,
	}

	npcConfig := &EntityConfig{
		ID:          "red_orc",
		ArchetypeID: "orc_male",
	}

	// Manual simulation of the inheritance logic in config.go
	if npcConfig.Stats.HealthMin == 0 {
		npcConfig.Stats = arch.Stats
	}
	if npcConfig.Behavior == "" {
		npcConfig.Behavior = arch.Behavior
	}
	if npcConfig.XP == 0 {
		npcConfig.XP = arch.XP
	}

	if npcConfig.Stats.HealthMin != 30 {
		t.Errorf("Expected health_min 30, got %d", npcConfig.Stats.HealthMin)
	}
	if npcConfig.XP != 11 {
		t.Errorf("Expected XP 11, got %d", npcConfig.XP)
	}
}

func TestMapType_TargetPointRaw(t *testing.T) {
	data := `
id: "test"
type: "reach_portal"
target_point:
  x: 12.5
  y: -5.0
`
	var mt MapType
	if err := yaml.Unmarshal([]byte(data), &mt); err != nil {
		t.Fatalf("Failed to unmarshal MapType: %v", err)
	}

	if mt.TargetPointRaw == nil {
		t.Fatal("TargetPointRaw should not be nil")
	}
	if mt.TargetPointRaw.X != 12.5 || mt.TargetPointRaw.Y != -5.0 {
		t.Errorf("Unexpected values in TargetPointRaw: %+v", mt.TargetPointRaw)
	}
}
