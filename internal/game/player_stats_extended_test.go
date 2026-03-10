package game

import "testing"

// TestPlayerAddXP_LevelUp verifies that gaining enough XP increases the player's level and heals them.
func TestPlayerAddXP_LevelUp(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
	mc.XP = 90
	mc.Level = 1
	mc.MaxHealth = 100
	mc.Health = 20 // Wounded

	// Gaining 10 XP → Total 100. Level = 100/100 + 1 = 2
	mc.AddXP(10)

	if mc.Level != 2 {
		t.Errorf("Expected Level 2, got %d", mc.Level)
	}
	if mc.Health != 100 {
		t.Errorf("Expected full health on level up, got %d", mc.Health)
	}
}

// TestPlayerAddXP_MultipleLevels verifies gaining a large amount of XP at once works correctly.
func TestPlayerAddXP_MultipleLevels(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
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

// TestPlayerStats_Reset verifies that NewPlayableCharacter sets sensible defaults.
func TestPlayerStats_Defaults(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
	if mc.MaxHealth <= 0 {
		t.Errorf("Expected positive MaxHealth, got %d", mc.MaxHealth)
	}
	if mc.Health != mc.MaxHealth {
		t.Errorf("Expected full health at start, got %d/%d", mc.Health, mc.MaxHealth)
	}
	if mc.Level != 1 {
		t.Errorf("Expected start level 1, got %d", mc.Level)
	}
}
