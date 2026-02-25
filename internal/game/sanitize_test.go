
package game

import (
	"testing"
)

func TestSanitizeEntityConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   EntityConfig
		wantID  string
		wantHP  int
		wantMax int
		wantSpd float64
	}{
		{
			name:    "empty id and name",
			input:   EntityConfig{},
			wantID:  "unknown",
			wantHP:  1,
			wantMax: 1,
			wantSpd: 0.01,
		},
		{
			name: "invalid health and speed",
			input: EntityConfig{
				ID: "orc",
				Stats: struct {
					HealthMin       int     `yaml:"health_min"`
					HealthMax       int     `yaml:"health_max"`
					Speed           float64 `yaml:"speed"`
					BaseAttack      int     `yaml:"base_attack"`
					BaseDefense     int     `yaml:"base_defense"`
					AttackCooldown  int     `yaml:"attack_cooldown"`
					AttackRange     float64 `yaml:"attack_range"`
					ProjectileSpeed float64 `yaml:"projectile_speed"`
				}{
					HealthMin: -5,
					HealthMax: -10,
					Speed:     -1.0,
				},
			},
			wantID:  "orc",
			wantHP:  1,
			wantMax: 1,
			wantSpd: 0.01,
		},
		{
			name: "speed too high",
			input: EntityConfig{
				ID: "hero",
				Stats: struct {
					HealthMin       int     `yaml:"health_min"`
					HealthMax       int     `yaml:"health_max"`
					Speed           float64 `yaml:"speed"`
					BaseAttack      int     `yaml:"base_attack"`
					BaseDefense     int     `yaml:"base_defense"`
					AttackCooldown  int     `yaml:"attack_cooldown"`
					AttackRange     float64 `yaml:"attack_range"`
					ProjectileSpeed float64 `yaml:"projectile_speed"`
				}{
					HealthMin: 100,
					HealthMax: 100,
					Speed:     5.0,
				},
			},
			wantID:  "hero",
			wantHP:  100,
			wantMax: 100,
			wantSpd: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.input
			sanitizeEntityConfig(&config, "test")
			if config.ID != tt.wantID {
				t.Errorf("ID: got %s, want %s", config.ID, tt.wantID)
			}
			if config.Stats.HealthMin != tt.wantHP {
				t.Errorf("HealthMin: got %d, want %d", config.Stats.HealthMin, tt.wantHP)
			}
			if config.Stats.HealthMax != tt.wantMax {
				t.Errorf("HealthMax: got %d, want %d", config.Stats.HealthMax, tt.wantMax)
			}
			if config.Stats.Speed != tt.wantSpd {
				t.Errorf("Speed: got %f, want %f", config.Stats.Speed, tt.wantSpd)
			}
		})
	}
}

func TestSanitizeObstacleArchetype(t *testing.T) {
	tests := []struct {
		name    string
		input   ObstacleArchetype
		wantID  string
		wantHP  int
		wantScl float64
	}{
		{
			name: "invalid values",
			input: ObstacleArchetype{
				ID:     "",
				Health: -10,
				Scale:  -1.0,
			},
			wantID:  "unknown",
			wantHP:  0,
			wantScl: 1.0,
		},
		{
			name: "excessive scale",
			input: ObstacleArchetype{
				ID:    "rock",
				Scale: 15.0,
			},
			wantID:  "rock",
			wantHP:  0,
			wantScl: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.input
			sanitizeObstacleArchetype(&config, "test")
			if config.ID != tt.wantID {
				t.Errorf("ID: got %s, want %s", config.ID, tt.wantID)
			}
			if config.Health != tt.wantHP {
				t.Errorf("Health: got %d, want %d", config.Health, tt.wantHP)
			}
			if config.Scale != tt.wantScl {
				t.Errorf("Scale: got %f, want %f", config.Scale, tt.wantScl)
			}
		})
	}
}

func TestSanitizeMapType(t *testing.T) {
	m := MapType{
		ID:              "",
		Difficulty:      -1,
		TargetKillCount: -1,
	}
	sanitizeMapType(&m, "test")
	if m.ID != "unknown" {
		t.Errorf("ID: got %s, want unknown", m.ID)
	}
	if m.Difficulty != 0 {
		t.Errorf("Difficulty: got %d, want 0", m.Difficulty)
	}
	if m.TargetKillCount != 0 {
		t.Errorf("TargetKillCount: got %d, want 0", m.TargetKillCount)
	}
}

func TestSanitizeSaveData(t *testing.T) {
	p := PlayerSaveData{
		Health:    -10,
		MaxHealth: 0,
		Level:     -1,
	}
	sanitizePlayerSaveData(&p, "test")
	if p.Health != 1 {
		t.Errorf("Player Health: got %d, want 1", p.Health)
	}
	if p.MaxHealth != 100 {
		t.Errorf("Player MaxHealth: got %d, want 100", p.MaxHealth)
	}
	if p.Level != 1 {
		t.Errorf("Player Level: got %d, want 1", p.Level)
	}

	n := NPCSaveData{
		Name:   "Orc",
		Health: -5,
		Level:  0,
	}
	sanitizeNPCSaveData(&n, 0, "test")
	if n.Health != 0 {
		t.Errorf("NPC Health: got %d, want 0", n.Health)
	}
	if n.Level != 1 {
		t.Errorf("NPC Level: got %d, want 1", n.Level)
	}
}
