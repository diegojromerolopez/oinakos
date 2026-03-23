package game

import "testing"

func TestActor_Degradation(t *testing.T) {
	pc := NewCharacter(0, 0, nil, 1, true, nil)
	item := &ItemInstance{
		Config: &ObjectConfig{Resistance: 5},
		Resistance: 5,
	}
	pc.Slots = map[string]*ItemInstance{"weapon": item}
	pc.Inventory = []*ItemInstance{item}

	// 1. Degrade Weapon
	pc.DegradeWeapon(nil)
	if item.Resistance != 4 {
		t.Errorf("Weapon Resistance: got %d, want 4", item.Resistance)
	}
	
	// Break weapon
	for i := 0; i < 4; i++ {
		pc.DegradeWeapon(nil)
	}
	if pc.Slots["weapon"] != nil {
		t.Error("Weapon should be removed from slots after breaking")
	}
	if len(pc.Inventory) != 0 {
		t.Error("Weapon should be removed from inventory after breaking")
	}

	// 2. Degrade Armor
	armor := &ItemInstance{
		Config: &ObjectConfig{Resistance: 2},
		Resistance: 2,
	}
	pc.Slots = map[string]*ItemInstance{"body": armor}
	pc.Inventory = []*ItemInstance{armor}
	
	pc.DegradeArmor(nil)
	if armor.Resistance != 1 {
		t.Errorf("Armor Resistance: got %d, want 1", armor.Resistance)
	}
	
	pc.DegradeArmor(nil)
	if pc.Slots["body"] != nil {
		t.Error("Armor should be removed from slots after breaking")
	}
}
