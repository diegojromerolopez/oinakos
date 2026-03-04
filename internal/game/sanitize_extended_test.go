package game

import (
	"math"
	"testing"
)

func TestHexToRGBA(t *testing.T) {
	cases := []struct {
		input string
		r, g  float32
		b, a  float32
	}{
		{"#FF0000", 1.0, 0.0, 0.0, 1.0},
		{"#00FF00", 0.0, 1.0, 0.0, 1.0},
		{"#0000FF", 0.0, 0.0, 1.0, 1.0},
		{"FF00FF", 1.0, 0.0, 1.0, 1.0},        // no leading #
		{"#FFFFFF", 1.0, 1.0, 1.0, 1.0},       // white
		{"#000000", 0.0, 0.0, 0.0, 1.0},       // black
		{"#808080", 0.502, 0.502, 0.502, 1.0}, // grey (approx)
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := HexToRGBA(tc.input)
			if got[3] != 1.0 {
				t.Errorf("alpha: got %f, want 1.0", got[3])
			}
			// Use approximate comparison for grey (rounding)
			approx := func(a, b float32) bool { return math.Abs(float64(a-b)) < 0.01 }
			if !approx(got[0], tc.r) || !approx(got[1], tc.g) || !approx(got[2], tc.b) {
				t.Errorf("HexToRGBA(%q) = %v, want [%f %f %f 1.0]", tc.input, got, tc.r, tc.g, tc.b)
			}
		})
	}
}

func TestHexToRGBA_Invalid(t *testing.T) {
	cases := []string{
		"",
		"#",
		"#FFF",    // too short
		"#GGGGGG", // invalid hex chars
		"not-hex",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			got := HexToRGBA(input)
			// Should fall back to white
			if got[0] != 1.0 || got[1] != 1.0 || got[2] != 1.0 || got[3] != 1.0 {
				t.Errorf("HexToRGBA(%q) = %v, want white [1 1 1 1]", input, got)
			}
		})
	}
}

func TestSanitizeMapType_Clamping(t *testing.T) {
	m := &MapType{
		ID:              "test",
		Name:            "Test",
		Difficulty:      -5,
		TargetKillCount: -10,
		TargetTime:      -1.0,
		TargetRadius:    -3.0,
		Spawns: []SpawnConfig{
			{Probability: -0.5, Frequency: -1.0},
			{Probability: 2.0, Frequency: 0.5},
		},
	}
	sanitizeMapType(m, "test")

	if m.Difficulty != 0 {
		t.Errorf("Difficulty: got %d, want 0", m.Difficulty)
	}
	if m.TargetKillCount != 0 {
		t.Errorf("TargetKillCount: got %d, want 0", m.TargetKillCount)
	}
	if m.TargetTime != 0 {
		t.Errorf("TargetTime: got %f, want 0", m.TargetTime)
	}
	if m.TargetRadius != 0 {
		t.Errorf("TargetRadius: got %f, want 0", m.TargetRadius)
	}
	if m.Spawns[0].Probability != 0 {
		t.Errorf("Spawn[0].Probability: got %f, want 0", m.Spawns[0].Probability)
	}
	if m.Spawns[0].Frequency != 0 {
		t.Errorf("Spawn[0].Frequency: got %f, want 0", m.Spawns[0].Frequency)
	}
	if m.Spawns[1].Probability != 1.0 {
		t.Errorf("Spawn[1].Probability: got %f, want 1.0", m.Spawns[1].Probability)
	}
}

func TestSanitizeMapType_EmptyIDAndName(t *testing.T) {
	m := &MapType{}
	sanitizeMapType(m, "test_source")

	if m.ID != "unknown" {
		t.Errorf("ID: got %q, want %q", m.ID, "unknown")
	}
	if m.Name != "unknown" {
		t.Errorf("Name: got %q, want %q", m.Name, "unknown")
	}
}

func TestSanitizePlayerSaveData(t *testing.T) {
	cases := []struct {
		name  string
		input PlayerSaveData
		check func(t *testing.T, p *PlayerSaveData)
	}{
		{
			name:  "negative health clamped to 1",
			input: PlayerSaveData{Health: -5, MaxHealth: 100, Level: 1},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.Health != 1 {
					t.Errorf("Health: got %d, want 1", p.Health)
				}
			},
		},
		{
			name:  "zero max_health clamped to 100",
			input: PlayerSaveData{Health: 10, MaxHealth: 0, Level: 1},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.MaxHealth != 100 {
					t.Errorf("MaxHealth: got %d, want 100", p.MaxHealth)
				}
			},
		},
		{
			name:  "health exceeds max_health clamped to max",
			input: PlayerSaveData{Health: 500, MaxHealth: 100, Level: 1},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.Health != 100 {
					t.Errorf("Health: got %d, want 100", p.Health)
				}
			},
		},
		{
			name:  "zero level clamped to 1",
			input: PlayerSaveData{Health: 10, MaxHealth: 100, Level: 0},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.Level != 1 {
					t.Errorf("Level: got %d, want 1", p.Level)
				}
			},
		},
		{
			name:  "negative XP clamped to 0",
			input: PlayerSaveData{Health: 10, MaxHealth: 100, Level: 1, XP: -50},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.XP != 0 {
					t.Errorf("XP: got %d, want 0", p.XP)
				}
			},
		},
		{
			name:  "negative kills clamped to 0",
			input: PlayerSaveData{Health: 10, MaxHealth: 100, Level: 1, Kills: -1},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.Kills != 0 {
					t.Errorf("Kills: got %d, want 0", p.Kills)
				}
			},
		},
		{
			name:  "negative base_attack clamped to 0",
			input: PlayerSaveData{Health: 10, MaxHealth: 100, Level: 1, BaseAttack: -3},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.BaseAttack != 0 {
					t.Errorf("BaseAttack: got %d, want 0", p.BaseAttack)
				}
			},
		},
		{
			name:  "negative base_defense clamped to 0",
			input: PlayerSaveData{Health: 10, MaxHealth: 100, Level: 1, BaseDefense: -2},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.BaseDefense != 0 {
					t.Errorf("BaseDefense: got %d, want 0", p.BaseDefense)
				}
			},
		},
		{
			name:  "valid data unchanged",
			input: PlayerSaveData{Health: 50, MaxHealth: 100, Level: 5, XP: 200, Kills: 10},
			check: func(t *testing.T, p *PlayerSaveData) {
				if p.Health != 50 || p.MaxHealth != 100 || p.Level != 5 || p.XP != 200 || p.Kills != 10 {
					t.Errorf("Valid data was unexpectedly modified: %+v", *p)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.input
			sanitizePlayerSaveData(&p, "test")
			tc.check(t, &p)
		})
	}
}

func TestSanitizeNPCSaveData(t *testing.T) {
	// Negative health → clamped to 0
	n := NPCSaveData{ArchetypeID: "orc", Health: -5, MaxHealth: 10, Level: 1}
	sanitizeNPCSaveData(&n, 0, "test")
	if n.Health != 0 {
		t.Errorf("Health: got %d, want 0", n.Health)
	}

	// Zero max_health with health > 0 → max_health set to health
	n2 := NPCSaveData{ArchetypeID: "orc", Health: 8, MaxHealth: 0, Level: 1}
	sanitizeNPCSaveData(&n2, 1, "test")
	if n2.MaxHealth != 8 {
		t.Errorf("MaxHealth: got %d, want 8", n2.MaxHealth)
	}

	// Zero max_health with health == 0 → max_health clamped to 1
	n3 := NPCSaveData{ArchetypeID: "orc", Health: 0, MaxHealth: 0, Level: 1}
	sanitizeNPCSaveData(&n3, 2, "test")
	if n3.MaxHealth != 1 {
		t.Errorf("MaxHealth: got %d, want 1", n3.MaxHealth)
	}

	// Zero level → clamped to 1
	n4 := NPCSaveData{ArchetypeID: "orc", Health: 5, MaxHealth: 10, Level: 0}
	sanitizeNPCSaveData(&n4, 3, "test")
	if n4.Level != 1 {
		t.Errorf("Level: got %d, want 1", n4.Level)
	}
}

func TestSanitizeObstacleArchetype_Clamping(t *testing.T) {
	o := &ObstacleArchetype{
		Health:       -5,
		CooldownTime: -1.0,
	}
	sanitizeObstacleArchetype(o, "test")

	if o.ID != "unknown" {
		t.Errorf("ID: got %q, want 'unknown'", o.ID)
	}
	if o.Health != 0 {
		t.Errorf("Health: got %d, want 0", o.Health)
	}
	if o.CooldownTime != 0 {
		t.Errorf("CooldownTime: got %f, want 0", o.CooldownTime)
	}
}

func TestSanitizeObstacleArchetype_HealthMinusOne(t *testing.T) {
	// Health == -1 is valid (indestructible), below -1 is clamped
	o := &ObstacleArchetype{ID: "wall", Health: -1}
	sanitizeObstacleArchetype(o, "test")
	if o.Health != -1 {
		t.Errorf("Health -1 should be unchanged, got %d", o.Health)
	}

	o2 := &ObstacleArchetype{ID: "wall", Health: -2}
	sanitizeObstacleArchetype(o2, "test")
	if o2.Health != 0 {
		t.Errorf("Health -2 should be clamped to 0, got %d", o2.Health)
	}
}
