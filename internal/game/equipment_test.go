package game

import (
	"testing"
)

func TestWeaponRollDamage_MinEqualsMax(t *testing.T) {
	w := &Weapon{Name: "Test", Damage: Damage{Min: 5, Max: 5}}
	for i := 0; i < 20; i++ {
		dmg := w.RollDamage()
		if dmg != 5 {
			t.Errorf("RollDamage with min=max=5: got %d, want 5", dmg)
		}
	}
}


func TestWeaponRollDamage_Range(t *testing.T) {
	w := &Weapon{Name: "Test", Damage: Damage{Min: 3, Max: 7}}
	for i := 0; i < 100; i++ {
		dmg := w.RollDamage()
		if dmg < 3 || dmg > 7 {
			t.Errorf("RollDamage: got %d, want [3,7]", dmg)
		}
	}
}


func TestWeaponRollDamage_Bonus(t *testing.T) {
	w := &Weapon{Name: "Test", Damage: Damage{Min: 5, Max: 5}, Bonus: 3}
	dmg := w.RollDamage()
	if dmg != 8 {
		t.Errorf("RollDamage with bonus=3: got %d, want 8", dmg)
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
func TestDamage_Average(t *testing.T) {
	d := Damage{Min: 10, Max: 20}
	if avg := d.Average(); avg != 15.0 {
		t.Errorf("expected 15.0, got %f", avg)
	}
}

func TestWeapon_GetMaxDistance(t *testing.T) {
	tests := []struct {
		dist string
		want float64
	}{
		{"touch", 2.5},
		{"", 2.5},
		{"10.5", 10.5},
		{"invalid", 2.5},
	}
	for _, tt := range tests {
		w := &Weapon{MaxDistance: tt.dist}
		if got := w.GetMaxDistance(); got != tt.want {
			t.Errorf("GetMaxDistance(%q) = %v, want %v", tt.dist, got, tt.want)
		}
	}
}

func TestWeapon_IsRanged(t *testing.T) {
	wMelee := &Weapon{Type: "melee", MaxDistance: "touch"}
	if wMelee.IsRanged() {
		t.Error("expected melee weapon to not be ranged")
	}
	
	wRanged := &Weapon{Type: "ranged", MaxDistance: "touch"}
	if !wRanged.IsRanged() {
		t.Error("expected weapon with type 'ranged' to be ranged")
	}
	
	wLongMelee := &Weapon{Type: "melee", MaxDistance: "10"}
	if wLongMelee.IsRanged() {
		t.Error("expected long distance melee weapon to not be ranged (remains melee sweep)")
	}
}
