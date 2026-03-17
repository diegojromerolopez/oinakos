package game

import (
	"math/rand"
	"strconv"
)

// Weapon defines offensive gear
type Damage struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

func (d Damage) Average() float64 {
	return float64(d.Min+d.Max) / 2.0
}

type Weapon struct {
	Name        string  `yaml:"name"`
	Type        string  `yaml:"type"`         // "ranged" or "melee"
	MaxDistance string  `yaml:"max_distance"` // "touch" or pixel distance
	Damage      Damage  `yaml:"damage"`
	Weight      float64 `yaml:"weight,omitempty"`
	Bonus       int     `yaml:"-"`
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
	WeaponFists          = &Weapon{Name: "Fists", Type: "melee", MaxDistance: "touch", Damage: Damage{Min: 1, Max: 2}}
	WeaponTizon          = &Weapon{Name: "Tizon", Type: "melee", MaxDistance: "touch", Damage: Damage{Min: 15, Max: 25}}
	WeaponIronBroadsword = &Weapon{Name: "Iron Broadsword", Type: "melee", MaxDistance: "touch", Damage: Damage{Min: 5, Max: 10}}
	WeaponOrcishAxe      = &Weapon{Name: "Orcish Axe", Type: "melee", MaxDistance: "touch", Damage: Damage{Min: 4, Max: 8}}
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
	dmg := w.Damage.Min
	if w.Damage.Max > w.Damage.Min {
		dmg = rand.Intn(w.Damage.Max-w.Damage.Min+1) + w.Damage.Min
	}
	return dmg + w.Bonus
}

func (w *Weapon) GetMaxDistance() float64 {
	if w.MaxDistance == "touch" || w.MaxDistance == "" {
		return 1.4 // Default melee range
	}
	dist, err := strconv.ParseFloat(w.MaxDistance, 64)
	if err != nil {
		return 1.4
	}
	return dist
}

func (w *Weapon) IsRanged() bool {
	return w.Type == "ranged" || w.GetMaxDistance() >= 2.0
}

