package game

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
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
	ActorResting   // Sleeping / Resting to regain energy
	ActorChopping  // Gathering timber
	ActorDigging   // Gathering ore/excavating
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
	Z      float64
	VerticalVelocity float64
	Energy float64
	Config *EntityConfig
	Facing Direction
	State  ActorState
	Trauma PhysicalTrauma

	// Equipment
	Inventory []*ItemInstance
	Slots     map[string]*ItemInstance // Maps slot name (head, body, ring1, ring2, etc.) to item
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
	BaseProtection     int
	Speed              float64
	Weapon             *Weapon
	Alignment          Alignment
	LastAIDecisionTick int
	Group              string
	LeaderID           string
	MustSurvive        bool
	IsTarget           bool
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
	for _, it := range a.Slots {
		if it == nil || it.Config == nil {
			continue
		}
		for stat, effect := range it.Config.Effects {
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

	// Sync Active Weapon
	if it, ok := a.Slots["weapon"]; ok && it != nil && it.Config != nil && it.Config.Combat != nil {
		a.Weapon = it.Config.Combat
	} else {
		// Fallback to default fists
		a.Weapon = WeaponFists
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
	if weaponItem, ok := a.Slots["weapon"]; ok && weaponItem != nil {
		if weaponItem.Config != nil && weaponItem.Config.Combat != nil {
			a.Weapon = weaponItem.Config.Combat
		}
	} else {
		a.Weapon = nil
	}
}

// EquipItem tries to equip the given object into its slot.
// Returns true if the item was equipped (improves stats or fills empty slot).
// Old equipped item is moved to inventory if necessary.
func (a *Actor) EquipItem(it *ItemInstance) bool {
	if it == nil || it.Config == nil || it.Config.Slot == "" {
		return false
	}

	current := a.Slots[it.Config.Slot]
	shouldEquip := false

	if current == nil {
		shouldEquip = true
	} else if it.Config.Type == "weapon" && current.Config.Type == "weapon" {
		curDmg := current.Config.Combat.Damage.Average()
		newDmg := it.Config.Combat.Damage.Average()
		if newDmg > curDmg {
			shouldEquip = true
		}
	} else {
		// Compare stat totals
		curStats := 0.0
		newStats := 0.0
		for _, e := range current.Config.Effects {
			curStats += e.Increase
		}
		for _, e := range it.Config.Effects {
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
		a.Slots[it.Config.Slot] = it

		// Remove from inventory if it was there
		for i, item := range a.Inventory {
			if item == it {
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
func (a *Actor) EvaluateUpgrade(it *ItemInstance) bool {
	if it == nil || it.Config == nil || it.Config.Slot == "" {
		return false
	}

	current := a.Slots[it.Config.Slot]
	if current == nil {
		return true // Upgrade if empty slot
	}

	if it.Config.Type == "weapon" && current.Config.Type == "weapon" {
		curDmg := current.Config.Combat.Damage.Average()
		newDmg := it.Config.Combat.Damage.Average()
		return newDmg > curDmg
	}

	// Compare stat totals
	curStats := 0.0
	newStats := 0.0
	for _, e := range current.Config.Effects {
		curStats += e.Increase
	}
	for _, e := range it.Config.Effects {
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
		if weaponItem, ok := a.Slots["weapon"]; ok && weaponItem != nil {
			if weaponItem.Config != nil && weaponItem.Config.Combat == a.Weapon {
				inSlot = true
			}
		}
		if !inSlot {
			// Find the config that owns this weapon
			// This is tricky because a.Weapon is a *Weapon
			// For now, we assume simple case.
			// Actually, if it's not in a slot, it's probably WeaponFists or similar which have no weight in ObjectConfig
		}
	}
	// 2. Inventory (backpack)
	for _, item := range a.Inventory {
		if item != nil && item.Config != nil {
			total += item.Config.Weight
		}
	}
	// 3. Equipped items
	for _, item := range a.Slots {
		if item != nil && item.Config != nil {
			total += item.Config.Weight
		}
	}
	return total
}

func (a *Actor) SharedUpdate(ctx *SystemContext) {
	if !a.IsAlive() {
		return
	}
	a.UpdateEffects() // Refresh bonuses from items

	if ctx != nil && ctx.World != nil && ctx.World.CurrentMapType != nil {
		groundZ := ctx.World.CurrentMapType.GetElevationAt(a.X, a.Y)
		a.VerticalVelocity -= 0.05 // Gravity
		a.Z += a.VerticalVelocity
		if a.Z < groundZ {
			a.Z = groundZ
			a.VerticalVelocity = 0
		}
	}

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

	// Energy Mechanics
	if a.State == ActorResting {
		recoveryRate := 0.05
		isComfy := false
		if ctx != nil && ctx.World != nil {
			for _, o := range ctx.World.Obstacles {
				id := strings.ToLower(o.ID)
				if o.Alive && (strings.Contains(id, "tavern") || strings.Contains(id, "house") || strings.Contains(id, "farm")) {
					dist := math.Sqrt(math.Pow(a.X-o.X, 2) + math.Pow(a.Y-o.Y, 2))
					if dist < 8.0 {
						recoveryRate = 0.25
						isComfy = true
						break
					}
				}
			}
		}
		a.Energy += recoveryRate
		if a.Energy >= 100 {
			a.Energy = 100
			a.State = ActorIdle // Wake up fully rested
		}
		if a.Tick%60 == 0 {
			healthFactor := 0.20
			if isComfy { healthFactor = 0.60 }
			if a.Health < a.GetTotalMaxHealth() {
				regen := int(float64(a.GetTotalMaxHealth()) * healthFactor / 33.0)
				if regen < 1 { regen = 1 }
				a.Heal(regen)
			}
			if ctx != nil && ctx.World != nil && (a.Tick%120 == 0) {
				msg := "Zzz"
				if isComfy { msg = "Zzz (comfy rest)" }
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
					Text: msg, X: a.X, Y: a.Y, Life: 60, Color: ColorHeal,
				})
			}
		}
	} else if a.State == ActorAttacking || a.State == ActorChopping || a.State == ActorDigging {
		a.Energy -= 0.08 // Heavy labor drains fast
		if a.Energy < 0 { a.Energy = 0 }
	} else if a.State == ActorWalking {
		a.Energy -= 0.008 // Moderate drain for travel
		if a.Energy < 0 { a.Energy = 0 }
	} else if a.State == ActorCrouching || a.State == ActorDrinking {
		a.Energy -= 0.002
		if a.Energy < 0 { a.Energy = 0 }
	} else {
		a.Energy -= 0.0005 // Very slow passive drain while standing
		if a.Energy < 0 { a.Energy = 0 }
	}

	if a.Energy <= 0 && a.Tick%120 == 0 && (a.State == ActorWalking || a.State == ActorAttacking || a.State == ActorChopping || a.State == ActorDigging) {
		a.Health -= 1
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: "-Starving/Tired-", X: a.X, Y: a.Y, Life: 45, Color: ColorHarm,
			})
		}
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
	return a.BaseProtection + a.ProtectionBonus
}

// LoadEquipment loads items from Config.Equipment map into Slots and Config.Inventory array into Inventory.
func (a *Actor) LoadEquipment(objRegistry *ObjectRegistry) {
	if a.Config == nil || objRegistry == nil {
		return
	}
	a.Inventory = nil
	if a.Slots == nil {
		a.Slots = make(map[string]*ItemInstance)
	}

	// Load explicitly mapped slots first
	for slotName, objID := range a.Config.Equipment {
		if obj, ok := objRegistry.Objects[objID]; ok {
			if obj.Slot != "" && obj.Slot != slotName && obj.Slot != "ring" {
				// Mismatch warning (e.g., trying to equip a helmet in the weapon slot)
				log.Printf("[WARNING] Entity %s equipped %s in slot %s, but item defines slot as %s", a.Config.Name, obj.Name, slotName, obj.Slot)
			}
			a.Slots[slotName] = NewItemInstance(obj.ID, obj, a.X, a.Y)
		}
	}

	// Load backpack inventory
	for _, objID := range a.Config.Inventory {
		if obj, ok := objRegistry.Objects[objID]; ok {
			a.Inventory = append(a.Inventory, NewItemInstance(obj.ID, obj, a.X, a.Y))
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

// GetSpeedModifier returns a movement multiplier based on the current tile type and environment.
func (a *Actor) GetSpeedModifier(ctx *SystemContext) float64 {
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
		
		if ctx != nil {
			switch ctx.Weather {
			case WeatherRain:
				multiplier *= 0.9
			case WeatherSnow:
				multiplier *= 0.75
			case WeatherStorm:
				multiplier *= 0.85
			}
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
// DegradeWeapon reduces the resistance of the equipped weapon.
func (a *Actor) DegradeWeapon(ctx *SystemContext) {
	it, ok := a.Slots["weapon"]
	if !ok || it == nil || it.Config == nil {
		return
	}
	if it.Config.Resistance <= 0 { // 0 or -1 means infinite/no resistance set? 
		// Actually user said: "resistance makes the weapon stand 50-200 points... After that number of hits, it will break."
		// So if Resistance is 0 in config, maybe it shouldn't degrade.
		return
	}

	it.Resistance--
	if it.Resistance <= 0 {
		// Break weapon
		delete(a.Slots, "weapon")
		// Remove from inventory too
		for i, invItem := range a.Inventory {
			if invItem == it {
				a.Inventory = append(a.Inventory[:i], a.Inventory[i+1:]...)
				break
			}
		}
		a.UpdateEffects()
		if ctx != nil && ctx.Log != nil {
			ctx.Log(fmt.Sprintf("Your %s BROKE!", it.Config.Name), LogWarning)
		}
	}
}

// DegradeArmor reduces the resistance of all equipped armor pieces.
func (a *Actor) DegradeArmor(ctx *SystemContext) {
	broken := false
	for slot, it := range a.Slots {
		if slot == "weapon" || it == nil || it.Config == nil || it.Config.Resistance <= 0 {
			continue
		}

		it.Resistance--
		if it.Resistance <= 0 {
			// Break armor
			delete(a.Slots, slot)
			for i, invItem := range a.Inventory {
				if invItem == it {
					a.Inventory = append(a.Inventory[:i], a.Inventory[i+1:]...)
					break
				}
			}
			broken = true
			if ctx != nil && ctx.Log != nil {
				ctx.Log(fmt.Sprintf("Your %s BROKE!", it.Config.Name), LogWarning)
			}
		}
	}
	if broken {
		a.UpdateEffects()
	}
}
// GetActiveTraumas returns a list of human-readable strings for all active physical injuries.
func (a *Actor) GetActiveTraumas() []string {
	var list []string
	if a.Trauma.LeftArmLost {
		list = append(list, "Left Arm Lost")
	}
	if a.Trauma.RightArmLost {
		list = append(list, "Right Arm Lost")
	}
	if a.Trauma.LeftLegLost {
		list = append(list, "Left Leg Lost")
	}
	if a.Trauma.RightLegLost {
		list = append(list, "Right Leg Lost")
	}
	if a.Trauma.EyesLost == 1 {
		list = append(list, "Lost one Eye")
	} else if a.Trauma.EyesLost >= 2 {
		list = append(list, "Blind (Lost all eyes)")
	}
	if a.Trauma.BurnedAlive {
		list = append(list, "Severe Burn Scars")
	}
	if a.Trauma.SpineBroken {
		list = append(list, "Broken Spine")
	}
	return list
}
func (a *Actor) hitCharacter(target *Actor, ctx *SystemContext) {
	attk, def := float64(a.GetTotalAttack()), float64(target.GetTotalDefense())
	if def <= 0 { def = 1 }
	hitChance := clampInt(int(attk/(attk+def)*100), 5, 95)

	if rand.Intn(100)+1 <= hitChance {
		rawDmg := a.rollDamage()
		finalDmg := int(math.Max(1, float64(rawDmg-target.GetTotalProtection())))
		target.TakeDamage(finalDmg, a, ctx)
		a.DegradeWeapon(ctx)
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: fmt.Sprintf("-%d", finalDmg), X: target.X, Y: target.Y, Life: 45, Color: ColorHarm,
			})
		}
	} else {
		if ctx != nil && ctx.World != nil {
			ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
				Text: "MISS", X: target.X, Y: target.Y, Life: 45, Color: ColorMiss,
			})
		}
	}
}

func (a *Actor) rollDamage() int {
	if a.Weapon != nil {
		return a.GetTotalAttack() + a.Weapon.RollDamage()
	}
	return a.GetTotalAttack() + WeaponFists.RollDamage()
}

func (a *Actor) TakeDamage(amount int, attacker ActorInterface, ctx *SystemContext) {
	if a.State == ActorDead {
		return
	}
	a.Health -= amount
	a.HitTimer = 30
	a.DegradeArmor(ctx)
	
	prefix := "unknown"
	if a.Config != nil {
		prefix = a.Config.SoundID
		if prefix == "" { prefix = a.Config.ID }
	}
	
	if ctx != nil && ctx.Audio != nil {
		ctx.Audio.PlayRandomSound(prefix + "/hit")
	}

	// Permanent Trauma Acquisition
	if float64(a.Health) < float64(a.GetTotalMaxHealth())*0.10 {
		a.acquireRandomTrauma(attacker)
	} else if amount > int(float64(a.GetTotalMaxHealth())*0.15) && rand.Float64() < 0.01 {
		a.acquireRandomTrauma(attacker)
	}

	deathThreshold := a.GetDeathThreshold()
	if a.Health < deathThreshold {
		a.Health = deathThreshold
	}

	a.SyncLifeStatus()
	
	if !a.IsAlive() {
		a.die(attacker, ctx)
	}
}

func (a *Actor) acquireRandomTrauma(attacker ActorInterface) {
	r := rand.Intn(7)
	switch r {
	case 0:
		if !a.Trauma.LeftArmLost {
			a.Trauma.LeftArmLost = true
			DebugLog("Actor [%s] %s lost their LEFT ARM!", a.Alignment, a.Name)
		}
	case 1:
		if !a.Trauma.RightArmLost {
			a.Trauma.RightArmLost = true
			DebugLog("Actor [%s] %s lost their RIGHT ARM!", a.Alignment, a.Name)
		}
	case 2:
		if !a.Trauma.LeftLegLost {
			a.Trauma.LeftLegLost = true
			DebugLog("Actor [%s] %s lost their LEFT LEG!", a.Alignment, a.Name)
		}
	case 3:
		if !a.Trauma.RightLegLost {
			a.Trauma.RightLegLost = true
			DebugLog("Actor [%s] %s lost their RIGHT LEG!", a.Alignment, a.Name)
		}
	case 4:
		if a.Trauma.EyesLost < 2 {
			a.Trauma.EyesLost++
			DebugLog("Actor [%s] %s lost an EYE! (Total lost: %d)", a.Alignment, a.Name, a.Trauma.EyesLost)
		}
	case 5:
		if !a.Trauma.BurnedAlive {
			a.Trauma.BurnedAlive = true
			DebugLog("Actor [%s] %s was BURNED ALIVE and survived!", a.Alignment, a.Name)
		}
	case 6:
		if !a.Trauma.SpineBroken {
			a.Trauma.SpineBroken = true
			DebugLog("Actor [%s] %s suffered a BROKEN SPINE!", a.Alignment, a.Name)
		}
	}
}

func (a *Actor) die(attacker ActorInterface, ctx *SystemContext) {
	a.State = ActorDead
	if ctx != nil && ctx.World != nil && ctx.World.Game != nil { 
		ctx.World.Game.DropAllItems(a) 
	}
	
	prefix := "unknown"
	if a.Config != nil { 
		prefix = a.Config.SoundID 
		if prefix == "" { prefix = a.Config.ID }
	}
	
	if ctx != nil && ctx.Audio != nil { 
		ctx.Audio.PlayRandomSound(prefix + "/death") 
	}
	
	if attacker != nil {
		if act := attacker.GetActor(); act != nil {
			act.Kills++
			if a.Config != nil {
				act.MapKills[a.Config.ID]++
				xp := a.Config.XP
				if xp <= 0 { xp = 1 }
				act.AddXP(xp)
			}
			
			// Process OnKill actions from the attacker's config
			if act.Config != nil && act.Config.Actions != nil {
				for _, action := range act.Config.Actions.OnKill {
					if rand.Float64() < action.Probability {
						a.applyKillAction(action, attacker, ctx)
					}
				}
			}
		}
	}
}

func (a *Actor) applyKillAction(action KillAction, attacker ActorInterface, ctx *SystemContext) {
	if action.Type == "transform_victim" {
		e := action.Effect.Victim
		if e == nil { return }
		
		targetID := e.Transform
		// Replace {gender} if present
		if a.Config != nil {
			targetID = strings.ReplaceAll(targetID, "{gender}", a.Config.Gender)
		}
		
		var newConfig *EntityConfig
		var ok bool
		if ctx != nil && ctx.Registries != nil {
			if ctx.Registries.Archetypes != nil {
				newConfig, ok = ctx.Registries.Archetypes.Archetypes[targetID]
			}
			if !ok && ctx.Registries.Characters != nil {
				newConfig, ok = ctx.Registries.Characters.Characters[targetID]
			}
		}
		
		if ok {
			a.Config = newConfig
			a.Health = a.GetTotalMaxHealth()
			a.State = ActorIdle
			if e.Alignment == "inherit" {
				a.Alignment = attacker.GetActor().Alignment
			}
		}
	}
	
	if action.Type == "heal_attacker" || (action.Effect.Attacker != nil && action.Effect.Attacker.Heal > 0) {
		attk := attacker.GetActor()
		if action.Effect.Attacker != nil {
			attk.Health += action.Effect.Attacker.Heal
			if attk.Health > attk.GetTotalMaxHealth() {
				attk.Health = attk.GetTotalMaxHealth()
			}
		}
	}
}
