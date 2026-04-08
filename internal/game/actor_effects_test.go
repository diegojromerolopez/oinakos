package game

import (
	"testing"
)

func TestActor_UpdateEffectsAdvanced(t *testing.T) {
	a := &Actor{
		Slots: make(map[string]*ItemInstance),
	}
	
	// 1. No items
	a.UpdateEffects()
	if a.AttackBonus != 0 { t.Error("Base attack bonus should be 0") }
	
	// 2. Add an item with effects
	it := &ItemInstance{
		Config: &ObjectConfig{
			Effects: map[string]StatEffect{
				"attack":     {Increase: 10},
				"defense":    {Increase: 5},
				"protection": {Increase: 2},
				"speed":      {Increase: 0.1},
				"max_health": {Increase: 20},
				"regen":      {Increase: 1},
			},
		},
	}
	a.Slots["ring"] = it
	a.UpdateEffects()
	
	if a.AttackBonus != 10 { t.Errorf("Expected 10, got %d", a.AttackBonus) }
	if a.DefenseBonus != 5 { t.Errorf("Expected 5, got %d", a.DefenseBonus) }
	if a.ProtectionBonus != 2 { t.Errorf("Expected 2, got %d", a.ProtectionBonus) }
	if a.SpeedBonus != 0.1 { t.Errorf("Expected 0.1, got %v", a.SpeedBonus) }
	if a.MaxHealthBonus != 20 { t.Errorf("Expected 20, got %d", a.MaxHealthBonus) }
	if a.RegenPerSecond != 1 { t.Errorf("Expected 1, got %d", a.RegenPerSecond) }
	
	// 3. Trauma effects
	a.Trauma.LeftArmLost = true
	a.UpdateEffects()
	if a.AttackBonus != 5 { t.Errorf("Expected 5 (10-5), got %d", a.AttackBonus) }

	a.Trauma.RightArmLost = true
	a.UpdateEffects()
	if a.AttackBonus != 0 { t.Errorf("Expected 0 (10-5-5), got %d", a.AttackBonus) }
	
	a.Trauma.EyesLost = 1
	a.UpdateEffects()
	// Should hit the EyesLost logic
}
