package game

import "math/rand"

// Weapon defines offensive gear
type Weapon struct {
	Name      string
	MinDamage int
	MaxDamage int
	Bonus     int
}

// Armor defines defensive gear
type Armor struct {
	Name       string
	Protection int
	Slot       ArmorSlot
}

type ArmorSlot int

const (
	SlotHead ArmorSlot = iota
	SlotBody
	SlotShield
)

// Default weapons
var (
	WeaponFists          = &Weapon{Name: "Fists", MinDamage: 1, MaxDamage: 2}
	WeaponRustySword     = &Weapon{Name: "Rusty Sword", MinDamage: 2, MaxDamage: 5}
	WeaponIronBroadsword = &Weapon{Name: "Iron Broadsword", MinDamage: 5, MaxDamage: 10}
	WeaponOrcishAxe      = &Weapon{Name: "Orcish Axe", MinDamage: 4, MaxDamage: 8}
)

// Default armor
var (
	// Body
	ArmorLeather   = &Armor{Name: "Leather Armor", Protection: 1, Slot: SlotBody}
	ArmorChainmail = &Armor{Name: "Chainmail", Protection: 2, Slot: SlotBody}
	ArmorPlate     = &Armor{Name: "Plate Armor", Protection: 5, Slot: SlotBody}

	// Shield
	ArmorWoodShield  = &Armor{Name: "Wood Shield", Protection: 1, Slot: SlotShield}
	ArmorIronShield  = &Armor{Name: "Iron Shield", Protection: 2, Slot: SlotShield}
	ArmorTowerShield = &Armor{Name: "Tower Shield", Protection: 4, Slot: SlotShield}

	// Head
	ArmorCap      = &Armor{Name: "Cap", Protection: 1, Slot: SlotHead}
	ArmorFullHelm = &Armor{Name: "Full Helm", Protection: 2, Slot: SlotHead}
)

func (w *Weapon) RollDamage() int {
	dmg := w.MinDamage
	if w.MaxDamage > w.MinDamage {
		dmg = rand.Intn(w.MaxDamage-w.MinDamage+1) + w.MinDamage
	}
	return dmg + w.Bonus
}
