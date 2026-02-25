package game

import (
	"testing"
)

func TestWeaponRollDamage_MinEqualsMax(t *testing.T) {
	w := &Weapon{Name: "Test", MinDamage: 5, MaxDamage: 5}
	for i := 0; i < 20; i++ {
		dmg := w.RollDamage()
		if dmg != 5 {
			t.Errorf("RollDamage with min=max=5: got %d, want 5", dmg)
		}
	}
}

func TestWeaponRollDamage_Range(t *testing.T) {
	w := &Weapon{Name: "Test", MinDamage: 3, MaxDamage: 7}
	for i := 0; i < 100; i++ {
		dmg := w.RollDamage()
		if dmg < 3 || dmg > 7 {
			t.Errorf("RollDamage: got %d, want [3,7]", dmg)
		}
	}
}

func TestWeaponRollDamage_Bonus(t *testing.T) {
	w := &Weapon{Name: "Test", MinDamage: 5, MaxDamage: 5, Bonus: 3}
	dmg := w.RollDamage()
	if dmg != 8 {
		t.Errorf("RollDamage with bonus=3: got %d, want 8", dmg)
	}
}

func TestGetWeaponByName(t *testing.T) {
	cases := []struct {
		name    string
		wantMin int
		wantMax int
	}{
		{"Tizon", 15, 25},
		{"Orcish Axe", 4, 8},
		{"Iron Broadsword", 5, 10},
		{"Fists", 1, 2},
		{"Cleaver", 3, 7},
		{"Trident", 6, 12},
		{"Whip", 4, 9},
		{"Bow", 3, 6},
		{"Dagger", 2, 5},
		{"Unknown", 1, 2}, // falls back to Fists
	}
	for _, tc := range cases {
		w := GetWeaponByName(tc.name)
		if w == nil {
			t.Errorf("GetWeaponByName(%q): got nil", tc.name)
			continue
		}
		if w.MinDamage != tc.wantMin || w.MaxDamage != tc.wantMax {
			t.Errorf("GetWeaponByName(%q): got min=%d max=%d, want %d/%d",
				tc.name, w.MinDamage, w.MaxDamage, tc.wantMin, tc.wantMax)
		}
	}
}

func TestArmorSlotConstants(t *testing.T) {
	if SlotHead != 0 {
		t.Errorf("SlotHead should be 0, got %d", SlotHead)
	}
	if SlotBody != 1 {
		t.Errorf("SlotBody should be 1, got %d", SlotBody)
	}
	if SlotShield != 2 {
		t.Errorf("SlotShield should be 2, got %d", SlotShield)
	}
}

func TestDefaultWeapons(t *testing.T) {
	if WeaponFists == nil {
		t.Error("WeaponFists is nil")
	}
	if WeaponTizon == nil {
		t.Error("WeaponTizon is nil")
	}
	if WeaponIronBroadsword == nil {
		t.Error("WeaponIronBroadsword is nil")
	}
	if WeaponOrcishAxe == nil {
		t.Error("WeaponOrcishAxe is nil")
	}
}

func TestDefaultArmor(t *testing.T) {
	armors := []*Armor{
		ArmorLeather, ArmorChainmail, ArmorPlate,
		ArmorWoodShield, ArmorIronShield, ArmorTowerShield,
		ArmorCap, ArmorFullHelm,
	}
	for _, a := range armors {
		if a == nil {
			t.Error("Armor piece is nil")
		}
	}
}
