package game

import "testing"

// TestPlayerAddXP_LevelUp verifies that gaining enough XP increases the player's level and heals them.
func TestPlayerAddXP_LevelUp(t *testing.T) {
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.XP = 90
	mc.Level = 1
	mc.TemporalState.MaxHealthPoints = 100
	mc.TemporalState.HealthPoints = 20 // Wounded

	// Gaining 10 XP → Total 100. Level = 100/100 + 1 = 2
	mc.AddXP(10)

	if mc.Level != 2 {
		t.Errorf("Expected Level 2, got %d", mc.Level)
	}
	if mc.TemporalState.HealthPoints != 100 {
		t.Errorf("Expected full health on level up, got %d", mc.TemporalState.HealthPoints)
	}
}

// TestPlayerAddXP_MultipleLevels verifies gaining a large amount of XP at once works correctly.
func TestPlayerAddXP_MultipleLevels(t *testing.T) {
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	mc.XP = 0
	mc.Level = 1

	// Gaining 250 XP → Total 250. Level = 250/100 + 1 = 3
	mc.AddXP(250)

	if mc.Level != 3 {
		t.Errorf("Expected Level 3, got %d", mc.Level)
	}
	if mc.XP != 250 {
		t.Errorf("Expected XP 250, got %d", mc.XP)
	}
}

// TestPlayerStats_Reset verifies that NewCharacter sets sensible defaults.
func TestPlayerStats_Defaults(t *testing.T) {
	mc := NewCharacter(0, 0, nil, 1, true, nil)
	if mc.TemporalState.MaxHealthPoints <= 0 {
		t.Errorf("Expected positive MaxHealth, got %d", mc.TemporalState.MaxHealthPoints)
	}
	if mc.TemporalState.HealthPoints != mc.TemporalState.MaxHealthPoints {
		t.Errorf("Expected full health at start, got %d/%d", mc.TemporalState.HealthPoints, mc.TemporalState.MaxHealthPoints)
	}
	if mc.Level != 1 {
		t.Errorf("Expected start level 1, got %d", mc.Level)
	}
}
