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
				Stats: EntityStatsConfig{
					HealthMin: IntInterval{Min: -5, Max: -5},
					HealthMax: IntInterval{Min: -10, Max: -10},
					Speed:     FloatInterval{Min: -1.0, Max: -1.0},
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
				Stats: EntityStatsConfig{
					HealthMin: IntInterval{Min: 100, Max: 100},
					HealthMax: IntInterval{Min: 100, Max: 100},
					Speed:     FloatInterval{Min: 5.0, Max: 5.0},
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
			if config.Stats.HealthMin.Min != tt.wantHP {
				t.Errorf("HealthMin: got %d, want %d", config.Stats.HealthMin.Min, tt.wantHP)
			}
			if config.Stats.HealthMax.Max != tt.wantMax {
				t.Errorf("HealthMax: got %d, want %d", config.Stats.HealthMax.Max, tt.wantMax)
			}
			if config.Stats.Speed.Min != tt.wantSpd {
				t.Errorf("Speed: got %f, want %f", config.Stats.Speed.Min, tt.wantSpd)
			}
		})
	}
}

func TestSanitizeObstacleArchetype(t *testing.T) {
	tests := []struct {
		name   string
		input  ObstacleArchetype
		wantID string
		wantHP int
	}{
		{
			name: "invalid values",
			input: ObstacleArchetype{
				ID:     "",
				HealthPoints: -10,
			},
			wantID: "unknown",
			wantHP: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.input
			sanitizeObstacleArchetype(&config, "test")
			if config.ID != tt.wantID {
				t.Errorf("ID: got %s, want %s", config.ID, tt.wantID)
			}
			if config.HealthPoints != tt.wantHP {
				t.Errorf("HealthPoints: got %d, want %d", config.HealthPoints, tt.wantHP)
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
		TemporalState: TemporalState{
			HealthPoints:    -10,
			MaxHealthPoints: 0,
		},
		Level:     -1,
	}
	sanitizePlayerSaveData(&p, "test")
	if p.TemporalState.HealthPoints != 1 {
		t.Errorf("Player HealthPoints: got %d, want 1", p.TemporalState.HealthPoints)
	}
	if p.TemporalState.MaxHealthPoints != 100 {
		t.Errorf("Player MaxHealthPoints: got %d, want 100", p.TemporalState.MaxHealthPoints)
	}
	if p.Level != 1 {
		t.Errorf("Player Level: got %d, want 1", p.Level)
	}

	n := NPCSaveData{
		Name:   "Orc",
		TemporalState: TemporalState{HealthPoints: -5},
		Level:  0,
	}
	sanitizeNPCSaveData(&n, 0, "test")
	if n.TemporalState.HealthPoints != 0 {
		t.Errorf("NPC HealthPoints: got %d, want 0", n.TemporalState.HealthPoints)
	}
	if n.Level != 1 {
		t.Errorf("NPC Level: got %d, want 1", n.Level)
	}
}
