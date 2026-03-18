package game

import (
	"log"
	"math"
	"oinakos/internal/engine"
)

// ActorState is the unified state enum for all living entities.
type ActorState int

const (
	ActorIdle ActorState = iota
	ActorWalking
	ActorAttacking
	ActorDead
	ActorDrinking  // Player-specific (well interaction)
	ActorCrouching // Picking up items
	ActorIncapacitated // Down but not yet Truly Dead
)

// Backward-compatible aliases for PlayableCharacterState
type PlayableCharacterState = ActorState

// Backward-compatible aliases for NPCState
type NPCState = ActorState

// Direction represents isometric facing direction.
type Direction int

const (
	DirSE Direction = iota
	DirSW
	DirNE
	DirNW
)

// Alignment represents faction membership.
type Alignment int

const (
	AlignmentEnemy Alignment = iota
	AlignmentNeutral
	AlignmentAlly
)

func (a Alignment) String() string {
	switch a {
	case AlignmentEnemy:
		return "ENEMY"
	case AlignmentNeutral:
		return "NEUTRAL"
	case AlignmentAlly:
		return "ALLY"
	default:
		return "UNKNOWN"
	}
}

// BehaviorType controls NPC AI decision-making.
type BehaviorType int

const (
	BehaviorWander BehaviorType = iota
	BehaviorPatrol
	BehaviorKnightHunter
	BehaviorNpcFighter
	BehaviorChaotic
	BehaviorEscort
	BehaviorFlee
)

// PhysicalTrauma tracks irreversible physical injuries.
type PhysicalTrauma struct {
	LeftArmLost   bool
	RightArmLost  bool
	LeftLegLost   bool
	RightLegLost  bool
	LeftHandLost  bool
	RightHandLost bool
	EyesLost      int  // 0, 1, or 2
	BurnedAlive   bool // Survivors of extreme fire
	SpineBroken   bool
}

// Actor holds all runtime state shared between the player character and any NPC.
type Actor struct {
	X, Y   float64
	Config *EntityConfig
	Facing Direction
	State  ActorState
	Trauma PhysicalTrauma

	// Equipment
	Inventory []*ObjectConfig
	Slots     map[string]*ObjectConfig // Maps slot name (head, body, ring1, ring2, etc.) to item
	MaxWeight float64

	// Bonus from items
	AttackBonus     int
	DefenseBonus    int
	ProtectionBonus int
	SpeedBonus      float64
	MaxHealthBonus  int
	RegenPerSecond  int

	Tick               int
	Health             int
	MaxHealth          int
	BaseAttack         int
	BaseDefense        int
	Speed              float64
	Weapon             *Weapon
	Alignment          Alignment
	LastAIDecisionTick int
	Group              string
	LeaderID           string
	MustSurvive        bool
	Level              int
	XP                 int
	Name               string
	CurrentTile        string // Set by Game loop before Update
	IsOccluded         bool   // Visual occlusion by an obstacle

	// Progress & Scoring (formerly player-only)
	Kills    int
	MapKills map[string]int

	// Timers
	HitTimer    int // How long to show hit sprite (BloodTimer on NPC)
	DeadTimer   int // Ticks since death
	CrouchTimer int // Ticks for crouch animation
}

// IsAlive returns true if the character is not in the Truly Dead state.
// They can be active OR incapacitated.
func (a *Actor) IsAlive() bool {
	return a.State != ActorDead
}

// IsTrulyDead returns true if the character has reached the final death threshold.
func (a *Actor) IsTrulyDead() bool {
	return a.State == ActorDead
}

// IsIncapacitated returns true if the character is downed but not yet truly dead.
func (a *Actor) IsIncapacitated() bool {
	return a.State == ActorIncapacitated
}

// GetDeathThreshold returns the negative health value at which the character truly dies.
func (a *Actor) GetDeathThreshold() int {
	return -int(float64(a.GetTotalMaxHealth()) * 0.10)
}

// SyncLifeStatus ensures the actor's state is in sync with their current health.
func (a *Actor) SyncLifeStatus() {
	if a.IsTrulyDead() {
		return
	}

	threshold := a.GetDeathThreshold()
	if a.Health <= threshold {
		a.Health = threshold
		a.State = ActorDead
		return
	}

	if a.Health <= 0 {
		if a.State != ActorIncapacitated {
			a.State = ActorIncapacitated
			DebugLog("Actor [%s] %s is INCAPACITATED!", a.Alignment, a.Name)
		}
	} else {
		if a.State == ActorIncapacitated {
			a.State = ActorIdle
			DebugLog("Actor [%s] %s has RECOVERED!", a.Alignment, a.Name)
		}
	}
}

// Heal increases health up to MaxHealth.
func (a *Actor) Heal(amount int) {
	if a.IsTrulyDead() {
		return
	}
	oldHealth := a.Health
	a.Health += amount
	maxH := a.GetTotalMaxHealth()
	if a.Health > maxH {
		a.Health = maxH
	}
	
	a.SyncLifeStatus()

	if a.Health > oldHealth {
		DebugLog("Actor Healed [%s] %s! +%d | Health: %d -> %d", a.Alignment, a.Name, amount, oldHealth, a.Health)
	}
}

// ApplyPermanentEffects permanently modifies the actor's base stats based on the object's effects.
func (a *Actor) ApplyPermanentEffects(obj *ObjectConfig) {
	if obj == nil || obj.Effects == nil {
		return
	}
	for stat, effect := range obj.Effects {
		switch stat {
		case "attack":
			a.BaseAttack += int(effect.Increase)
		case "defense":
			a.BaseDefense += int(effect.Increase)
		case "speed":
			a.Speed += effect.Increase
		case "max_health":
			a.MaxHealth += int(effect.Increase)
			a.Health += int(effect.Increase)
		case "xp":
			a.AddXP(int(effect.Increase))
		}
	}
}

func (a *Actor) UpdateEffects() {
	a.AttackBonus = 0
	a.DefenseBonus = 0
	a.ProtectionBonus = 0
	a.SpeedBonus = 0
	a.MaxHealthBonus = 0
	a.RegenPerSecond = 0

	// Apply effects from equipped slots
	for _, obj := range a.Slots {
		if obj == nil {
			continue
		}
		for stat, effect := range obj.Effects {
			switch stat {
			case "attack":
				a.AttackBonus += int(effect.Increase)
			case "defense":
				a.DefenseBonus += int(effect.Increase)
			case "protection":
				a.ProtectionBonus += int(effect.Increase)
			case "speed":
				a.SpeedBonus += effect.Increase
			case "max_health":
				a.MaxHealthBonus += int(effect.Increase)
			case "regen":
				a.RegenPerSecond += int(effect.Increase)
			}
		}
	}

	// Apply effects from Trauma
	if a.Trauma.LeftArmLost {
		a.AttackBonus -= 5
	}
	if a.Trauma.RightArmLost {
		a.AttackBonus -= 5
	}
	if a.Trauma.EyesLost > 0 {
		a.AttackBonus -= 5 * a.Trauma.EyesLost
	}
	if a.Trauma.BurnedAlive {
		a.MaxHealthBonus -= 30
	}
	if a.Trauma.SpineBroken {
		a.MaxHealthBonus -= 20
	}

	// Sync active weapon from "weapon" slot
	if weaponObj, ok := a.Slots["weapon"]; ok && weaponObj != nil {
		if weaponObj.Combat != nil {
			a.Weapon = weaponObj.Combat
		}
	} else {
		a.Weapon = nil
	}
}

// EquipItem tries to equip the given object into its slot.
// Returns true if the item was equipped (improves stats or fills empty slot).
// Old equipped item is moved to inventory if necessary.
func (a *Actor) EquipItem(obj *ObjectConfig) bool {
	if obj.Slot == "" {
		return false
	}

	current := a.Slots[obj.Slot]
	shouldEquip := false

	if current == nil {
		shouldEquip = true
	} else if obj.Type == "weapon" && current.Type == "weapon" {
		curDmg := current.Combat.Damage.Average()
		newDmg := obj.Combat.Damage.Average()
		if newDmg > curDmg {
			shouldEquip = true
		}
	} else {
		// Compare stat totals
		curStats := 0.0
		newStats := 0.0
		for _, e := range current.Effects {
			curStats += e.Increase
		}
		for _, e := range obj.Effects {
			newStats += e.Increase
		}
		if newStats > curStats {
			shouldEquip = true
		}
	}

	if shouldEquip {
		if current != nil {
			a.Inventory = append(a.Inventory, current)
		}
		a.Slots[obj.Slot] = obj

		// Remove from inventory if it was there
		for i, item := range a.Inventory {
			if item == obj {
				a.Inventory = append(a.Inventory[:i], a.Inventory[i+1:]...)
				break
			}
		}

		a.UpdateEffects()
		return true
	}

	return false
}

// EvaluateUpgrade checks if the item is better than what the actor has equipped.
func (a *Actor) EvaluateUpgrade(obj *ObjectConfig) bool {
	if obj.Slot == "" {
		return false
	}

	current := a.Slots[obj.Slot]
	if current == nil {
		return true // Upgrade if empty slot
	}

	if obj.Type == "weapon" && current.Type == "weapon" {
		curDmg := current.Combat.Damage.Average()
		newDmg := obj.Combat.Damage.Average()
		return newDmg > curDmg
	}

	// Compare stat totals
	curStats := 0.0
	newStats := 0.0
	for _, e := range current.Effects {
		curStats += e.Increase
	}
	for _, e := range obj.Effects {
		newStats += e.Increase
	}
	return newStats > curStats
}

// GetTotalWeight returns the total weight of everything carried and equipped.
func (a *Actor) GetTotalWeight() float64 {
	total := 0.0
	// 1. Starter/Active weapon weight (if it's not in a slot to avoid double counting)
	if a.Weapon != nil {
		// Optimization: if it's already in the slots, it will be counted in the slots loop
		inSlot := false
		if weaponObj, ok := a.Slots["weapon"]; ok && weaponObj != nil {
			if weaponObj.Combat == a.Weapon {
				inSlot = true
			}
		}
		if !inSlot {
			total += a.Weapon.Weight
		}
	}
	// 2. Inventory (backpack)
	for _, item := range a.Inventory {
		if item != nil {
			total += item.Weight
		}
	}
	// 3. Equipped items
	for _, item := range a.Slots {
		if item != nil {
			total += item.Weight
		}
	}
	return total
}

func (a *Actor) SharedUpdate(ctx *SystemContext) {
	if !a.IsAlive() {
		return
	}
	a.UpdateEffects() // Refresh bonuses from items

	// Regeneration (1 unit per second = 1 unit every 60 ticks)
	if a.RegenPerSecond > 0 && a.Health < a.GetTotalMaxHealth() {
		if a.Tick%60 == 0 {
			a.Heal(a.RegenPerSecond)
		}
	}

	// Trauma: Continuous Pain from BurnedAlive (-1 HP every 600 ticks = 10s)
	if a.Trauma.BurnedAlive && a.Tick%600 == 0 {
		a.Health -= 1
	}

	// Incapacitated Bleed-out (-1 HP per "hour" = 3600 ticks / 1 minute)
	if a.IsIncapacitated() && a.Tick%3600 == 0 {
		a.Health -= 1
	}

	a.SyncLifeStatus()

	if a.CrouchTimer > 0 {
		a.CrouchTimer--
		if a.CrouchTimer == 0 && a.State == ActorCrouching {
			a.State = ActorIdle
		}
	}
}

type ActorInterface interface {
	GetActor() *Actor
	Heal(amount int)
}

func (a *Actor) GetActor() *Actor {
	return a
}

func (a *Actor) GetSortY() float64 {
	sortY := a.X + a.Y
	if a.State == ActorDead {
		sortY -= 100.0
	}
	return sortY
}

// GetTotalProtection returns the sum of all equipped armor.
func (a *Actor) GetTotalProtection() int {
	return a.ProtectionBonus
}

// LoadEquipment loads items from Config.Equipment map into Slots and Config.Inventory array into Inventory.
func (a *Actor) LoadEquipment(objRegistry *ObjectRegistry) {
	if a.Config == nil || objRegistry == nil {
		return
	}
	a.Inventory = nil
	if a.Slots == nil {
		a.Slots = make(map[string]*ObjectConfig)
	}

	// Load explicitly mapped slots first
	for slotName, objID := range a.Config.Equipment {
		if obj, ok := objRegistry.Objects[objID]; ok {
			if obj.Slot != "" && obj.Slot != slotName && obj.Slot != "ring" {
				// Mismatch warning (e.g., trying to equip a helmet in the weapon slot)
				log.Printf("[WARNING] Entity %s equipped %s in slot %s, but item defines slot as %s", a.Config.Name, obj.Name, slotName, obj.Slot)
			}
			a.Slots[slotName] = obj
		}
	}

	// Load backpack inventory
	for _, objID := range a.Config.Inventory {
		if obj, ok := objRegistry.Objects[objID]; ok {
			a.Inventory = append(a.Inventory, obj)
		}
	}

	a.UpdateEffects()
	a.MaxWeight = a.Config.MaxWeight
}

// calculateStat applies logarithmic level scaling.
func (a *Actor) calculateStat(base, level int) int {
	if level <= 1 {
		return base
	}
	bonus := int(math.Log2(float64(level)) * 10)
	return base + bonus
}

// GetSpeedModifier returns a movement multiplier based on the current tile type.
func (a *Actor) GetSpeedModifier() float64 {
	switch a.CurrentTile {
	case "water.png", "dark_water.png":
		return 0.5
	case "mud.png":
		return 0.8
	default:
		multiplier := 1.0
		if a.Trauma.LeftLegLost {
			multiplier -= 0.5
		}
		if a.Trauma.RightLegLost {
			multiplier -= 0.5
		}
		if a.Trauma.SpineBroken {
			multiplier *= 0.2 // Broken spine is a massive hit
		}
		if multiplier < 0.1 {
			multiplier = 0.1 // Minimum crawl
		}
		return multiplier
	}
}

// GetTotalAttack returns the level-scaled attack value plus item bonuses.
func (a *Actor) GetTotalAttack() int {
	return a.calculateStat(a.BaseAttack, a.Level) + a.AttackBonus
}

// GetTotalDefense returns the level-scaled defense value plus item bonuses.
func (a *Actor) GetTotalDefense() int {
	return a.calculateStat(a.BaseDefense, a.Level) + a.DefenseBonus
}

// GetTotalMaxHealth returns the maximum health plus item bonuses.
func (a *Actor) GetTotalMaxHealth() int {
	return a.MaxHealth + a.MaxHealthBonus
}

// GetCollisionCircle returns the collision circle for this actor.
func (a *Actor) GetCollisionCircle() engine.Circle {
	radius := 0.9375 // Default radius (30 world units is too large, but 30px is 0.9375 world units)
	if a.Config != nil && a.Config.CollisionRadius > 0 {
		radius = a.Config.CollisionRadius
	}
	return engine.Circle{X: a.X, Y: a.Y, Radius: radius}
}

// checkCollisionAt tests whether moving this actor to (newX, newY) would collide with any obstacle.
func (a *Actor) checkCollisionAt(newX, newY float64, obstacles []*Obstacle) bool {
	circle := a.GetCollisionCircle()
	circle.X = newX
	circle.Y = newY

	for _, o := range obstacles {
		if !o.Alive {
			continue
		}
		if engine.CheckCirclePolygonCollision(circle, o.GetFootprint()) {
			return true
		}
	}
	return false
}

// AddXP adds experience and handles level-up logic.
func (a *Actor) AddXP(amount int) {
	a.XP += amount
	newLevel := a.XP/100 + 1
	if newLevel > a.Level {
		a.Level = newLevel
		a.Health = a.MaxHealth
	}
}
