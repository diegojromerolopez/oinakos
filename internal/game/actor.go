package game

import (
	"fmt"
	"math"
	"math/rand"
	"oinakos/internal/engine"
	"strings"
)

// IsAlive returns true if the character is not in the Truly Dead state.
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
	if a.State == ActorDead { return }

	threshold := a.GetDeathThreshold()
	if a.TemporalState.HealthPoints <= threshold {
		fmt.Printf("Actor %s (Age=%d) DIED: HP=%d <= Threshold=%d\n", a.Name, int(a.AgeTicks/float64(TicksPerDay)), a.TemporalState.HealthPoints, threshold)
		a.TemporalState.HealthPoints = threshold
		a.State = ActorDead
		return
	}

	if a.TemporalState.HealthPoints <= 0 || a.UnconsciousTimer > 0 {
		if a.State != ActorIncapacitated {
			a.State = ActorIncapacitated
			if a.UnconsciousTimer > 0 {
				DebugLog("Actor [%s] %s is UNCONSCIOUS!", a.Alignment, a.Name)
			} else {
				DebugLog("Actor [%s] %s is CRITICALLY WOUNDED!", a.Alignment, a.Name)
			}
		}
	} else {
		if a.State == ActorIncapacitated {
			a.State = ActorIdle
			DebugLog("Actor [%s] %s has RECOVERED consciousness!", a.Alignment, a.Name)
		}
	}
}

func (a *Actor) InitBodyStatus() {
	if a.BodyStatus == nil { a.BodyStatus = make(map[string]int) }
	maxH := a.GetTotalMaxHealth()
	a.BodyStatus["head"] = maxH / 4
	a.BodyStatus["torso"] = maxH / 2
	a.BodyStatus["l_arm"] = maxH / 4
	a.BodyStatus["r_arm"] = maxH / 4
	a.BodyStatus["l_leg"] = maxH / 4
	a.BodyStatus["r_leg"] = maxH / 4
}

func (a *Actor) GetActor() *Actor { return a }

func (a *Actor) GetSortY() float64 {
	sortY := a.X + a.Y
	if a.State == ActorDead { sortY -= 100.0 }
	return sortY
}

func (a *Actor) GetTotalProtection() int { return a.BaseProtection + a.ProtectionBonus }

func (a *Actor) GetTotalAttack() int { return a.calculateStat(a.BaseAttack, a.Level) + a.AttackBonus }
func (a *Actor) GetTotalDefense() int { return a.calculateStat(a.BaseDefense, a.Level) + a.DefenseBonus }
func (a *Actor) SyncStats(objReg *ObjectRegistry) {
	if a.Config == nil {
		return
	}

	// 1. Initialize Runtime Attributes from Config
	// NOTE: Starting attributes are set once in NewCharacter/Load.
	// SyncStats now only ensures they stay in range and handles bonus scaling.

	// Ensure 0-100 range
	a.PrimaryAttributes.Strength = clampInt(a.PrimaryAttributes.Strength, 0, 100)
	a.PrimaryAttributes.Dexterity = clampInt(a.PrimaryAttributes.Dexterity, 0, 100)
	a.PrimaryAttributes.Health = clampInt(a.PrimaryAttributes.Health, 0, 100)
	a.PrimaryAttributes.Intellect = clampInt(a.PrimaryAttributes.Intellect, 0, 100)
	a.PrimaryAttributes.Wisdom = clampInt(a.PrimaryAttributes.Wisdom, 0, 100)

	// 2. Apply Aging Modifiers
	age := float64(a.AgeTicks) / float64(TicksPerYear)
	pMult, mMult := 1.0, 1.0
	
	if age < 25 {
		penaltyPrc := (25.0 - age) / 25.0
		pMult = 1.0 - (0.25 * penaltyPrc)
		mMult = 1.0 - (0.30 * penaltyPrc)
	} else if age > 40 {
		mMult = 1.0 + (0.05 * math.Floor((age-40.0)/10.0))
		pPenalty := 0.25 * (age - 40.0) / (85.0 - 40.0)
		if pPenalty > 0.25 { pPenalty = 0.25 }
		pMult = 1.0 - pPenalty
	}

	str := float64(a.PrimaryAttributes.Strength) * pMult
	dex := float64(a.PrimaryAttributes.Dexterity) * pMult
	hlt := float64(a.PrimaryAttributes.Health) * pMult
	itl := float64(a.PrimaryAttributes.Intellect) * mMult
	wis := float64(a.PrimaryAttributes.Wisdom) * mMult

	if a.TemporalState.Arousal > 10 {
		penalty := a.TemporalState.Arousal * 0.5
		itl -= penalty
		wis -= penalty
		if itl < 1 { itl = 1 }
		if wis < 1 { wis = 1 }
	}
	
	if a.TemporalState.IsDrunk {
		dex *= 0.7
		itl *= 0.7
		wis *= 0.7
		if dex < 1 { dex = 1 }
		if itl < 1 { itl = 1 }
		if wis < 1 { wis = 1 }
	}

	// 3. Derived Stats from runtime Attributes (Scaling even the overrides by age)
	if a.RawStats.BaseAttack > 0 {
		a.BaseAttack = int(float64(a.RawStats.BaseAttack) * pMult)
	} else {
		a.BaseAttack = int(str * 2)
	}

	if a.RawStats.BaseDefense > 0 {
		a.BaseDefense = int(float64(a.RawStats.BaseDefense) * pMult)
	} else {
		a.BaseDefense = int(dex*1.5 + hlt*1.0)
	}
	
	a.RangedAttack = int(dex * 2)
	a.Speed = dex * 0.02
	if a.RawStats.Speed > 0 { a.Speed = a.RawStats.Speed * pMult }
	if a.Speed <= 0 { a.Speed = 0.01 }
	a.CriticalChance = str * 0.005

	// Productive / Social stats
	a.Nourishment = int(hlt * 2)
	a.Survivalism = int(str*0.5 + hlt*0.5)
	a.Mate = hlt * 0.01
	a.Crafting = int(itl*1.2 + str*0.3)
	a.Herbalism = int(wis*1.0 + itl*0.5)
	a.Trading = int(itl*1.2 + wis*0.3)
	a.Harvesting = int(wis*1.2 + dex*0.3)
	a.Husbandry = int(wis*1.0 + dex*0.5)
	a.Art = int(dex*0.5 + itl*0.5)
	a.Culture = int(itl*0.5 + wis*0.5)

	// 4. Update MaxHealthPoints based on (possibly aged) health attribute
	if a.RawStats.HealthMin > 0 {
		a.TemporalState.MaxHealthPoints = int(float64(a.RawStats.HealthMin) * pMult)
	} else {
		a.TemporalState.MaxHealthPoints = int(hlt * 10)
	}
	if a.TemporalState.MaxHealthPoints < 10 { a.TemporalState.MaxHealthPoints = 10 }
	if a.TemporalState.HealthPoints > a.TemporalState.MaxHealthPoints {
		a.TemporalState.HealthPoints = a.TemporalState.MaxHealthPoints
	}

	// Attack Cooldown scaling: 1.5x at Dexterity 0, 1.0x at Dexterity 50, 0.5x at Dexterity 100
	cooldownMult := 1.5 - (dex * 0.01)
	baseCD := a.RawStats.AttackCooldown
	if baseCD == 0 { baseCD = 60 }
	a.BaseAttackCooldown = int(float64(baseCD) * cooldownMult)
	if a.BaseAttackCooldown < 10 { a.BaseAttackCooldown = 10 }

	// Only use attribute-based weight if archetype doesn't define it
	if a.RawStats.MaxWeight > 0 {
		a.MaxWeight = a.RawStats.MaxWeight
	} else {
		a.MaxWeight = (str*1.5 + hlt*0.5) / 0.329
	}
	a.BaseWeapon = a.Config.Weapon.Resolve(objReg)
	if a.BaseWeapon == nil {
		a.BaseWeapon = WeaponFists
	}
	a.Weapon = a.BaseWeapon

	a.BaseProtection = a.calculateStat(a.RawStats.BaseProtection, a.Level)
}

func (a *Actor) calculateStat(base int, level int) int {
	if level <= 1 { return base }
	return int(float64(base) * math.Pow(1.15, float64(level-1)))
}

func clampInt(v, min, max int) int {
	if v < min { return min }
	if v > max { return max }
	return v
}

func (a *Actor) GetTotalMaxHealth() int { return a.TemporalState.MaxHealthPoints + a.MaxHealthBonus }

func (a *Actor) GetSpeedModifier(ctx *SystemContext) float64 {
	switch a.CurrentTile {
	case "water.png", "dark_water.png": return 0.5
	case "mud.png": return 0.8
	default:
		multiplier := 1.0
		if a.Trauma.LeftLegLost { multiplier -= 0.5 }
		if a.Trauma.RightLegLost { multiplier -= 0.5 }
		if a.Trauma.SpineBroken { multiplier *= 0.2 }
		if multiplier < 0.1 { multiplier = 0.1 }
		
		// Pain dizzy / incapacitated (dizziness occurs at > 50 pain)
		if a.TemporalState.Pain > 50 { multiplier *= (1.0 - (a.TemporalState.Pain-50)/100.0) }
		if a.TemporalState.Pain > 80 { multiplier = 0 } // Incapacitated
		
		if ctx != nil {
			switch ctx.Weather {
			case WeatherRain: multiplier *= 0.9
			case WeatherSnow: multiplier *= 0.75
			case WeatherStorm: multiplier *= 0.85
			}
		}
		return multiplier
	}
}

func (a *Actor) GetCollisionCircle() engine.Circle {
	radius := 0.4
	if a.Config != nil && a.Config.CollisionRadius > 0 { radius = a.Config.CollisionRadius }
	return engine.Circle{X: a.X, Y: a.Y, Radius: radius}
}

func (a *Actor) AddMemory(tick int, mType, source string, value float64) {
	if a.Memories == nil { a.Memories = []MemoryEvent{} }
	a.Memories = append(a.Memories, MemoryEvent{Tick: tick, Type: mType, Source: source, Value: value})
	if len(a.Memories) > 20 { a.Memories = a.Memories[1:] }
	
	// Memories influence sentiment
	a.ModifySentiment(source, value)
}

func (a *Actor) checkCollisionAt(nx, ny float64, obstacles []*Obstacle) bool {
	col := a.GetCollisionCircle()
	col.X, col.Y = nx, ny
	for _, o := range obstacles {
		if o.Alive && o.Archetype != nil && !o.Archetype.Passable && engine.CheckCirclePolygonCollision(col, o.GetFootprint()) {
			return true
		}
	}
	return false
}

func (a *Actor) AddXP(amount int) {
	a.XP += amount
	newLevel := a.XP/100 + 1
	if newLevel > a.Level {
		a.Level = newLevel
		a.TemporalState.HealthPoints = a.TemporalState.MaxHealthPoints
		if a.BodyStatus != nil { a.InitBodyStatus() }
	}
}

// GetTraumaDescription returns a summary of physical injuries.
func (a *Actor) GetTraumaDescription() string {
	res := []string{}
	if a.Trauma.BurnedAlive { res = append(res, "Severely Burned") }
	if a.Trauma.LeftArmLost { res = append(res, "Left Arm Amputated") }
	if a.Trauma.RightArmLost { res = append(res, "Right Arm Amputated") }
	if a.Trauma.LeftLegLost { res = append(res, "Left Leg Amputated") }
	if a.Trauma.RightLegLost { res = append(res, "Right Leg Amputated") }
	if a.Trauma.EyesLost >= 2 { 
		res = append(res, "Permanently Blind") 
	} else if a.Trauma.EyesLost == 1 { 
		res = append(res, "One Eye Lost") 
	}
	if a.Trauma.SpineBroken { res = append(res, "Broken Spine (Paralyzed)") }
	
	if len(res) == 0 { return "No permanent traumas." }
	return strings.Join(res, ", ")
}

func (a *Actor) GetInventoryNames() []string {
	var names []string
	for _, it := range a.Inventory { if it != nil && it.Config != nil { names = append(names, it.Config.Name) } }
	for _, it := range a.Slots { if it != nil && it.Config != nil { names = append(names, it.Config.Name+" (equipped)") } }
	return names
}

func (a *Actor) GetActiveTraumas() []string {
	var traumas []string
	if a.Trauma.LeftArmLost { traumas = append(traumas, "Left Arm Lost") }
	if a.Trauma.RightArmLost { traumas = append(traumas, "Right Arm Lost") }
	if a.Trauma.LeftLegLost { traumas = append(traumas, "Left Leg Lost") }
	if a.Trauma.RightLegLost { traumas = append(traumas, "Right Leg Lost") }
	if a.Trauma.EyesLost > 0 { traumas = append(traumas, fmt.Sprintf("%d Eyes Lost", a.Trauma.EyesLost)) }
	if a.Trauma.BurnedAlive { traumas = append(traumas, "Burned") }
	if a.Trauma.SpineBroken { traumas = append(traumas, "Spine Broken") }
	return traumas
}

type ActorInterface interface {
	GetActor() *Actor
	Heal(amount int)
}

// CheckAttributeSuccess performs a uniform 0-100 roll against a primary attribute.
// Returns true if roll <= attribute value.
// CheckAbilitySuccess resolves whether an ability attempt succeeds.
// modifier shifts the effective attribute threshold: positive = bonus (easier),
// negative = penalty (harder). Clamped to [0, 100].
func (a *Actor) CheckAbilitySuccess(abilityID string, modifier int) bool {
	// 1. Check if the character has a specific skill value for this ability
	if a.SkillValues != nil {
		if val, exists := a.SkillValues[abilityID]; exists {
			return a.checkThreshold(val, modifier)
		}
	}

	// 2. Check if the ability has a defined parent attribute in its config
	if a.Config != nil && a.Config.Abilities != nil {
		if ability, exists := a.Config.Abilities[abilityID]; exists && ability.ParentAttribute != "" {
			return a.CheckAttributeSuccess(ability.ParentAttribute, modifier)
		}
	}

	// 3. Fallback attribute mapping
	switch abilityID {
	case "punch", "kick", "heavy_strike", "chop", "dig", "build", "butcher", "throw", "knockout", "grapple":
		return a.CheckAttributeSuccess("strength", modifier)
	case "slap", "slash", "shoot_arrow", "milk", "shear", "sneak", "steal", "seduce", "weave":
		return a.CheckAttributeSuccess("dexterity", modifier)
	case "rest", "eat", "drink", "mate":
		return a.CheckAttributeSuccess("health", modifier)
	case "cook", "craft", "repair", "brew", "trade", "smelt", "read", "appraise", "intimidate", "tan", "lie":
		return a.CheckAttributeSuccess("intellect", modifier)
	case "forage", "plant", "harvest_crop", "tame", "fish", "pray", "guard",
		"hunt", "trap", "tend_animal", "breed", "water_crops", "bury", "recruit", "teach":
		return a.CheckAttributeSuccess("wisdom", modifier)
	}

	return false
}

func (a *Actor) CheckAttributeSuccess(attr string, modifier int) bool {
	val := a.getAttrValue(attr)
	return a.checkThreshold(val, modifier)
}

func (a *Actor) checkThreshold(val, modifier int) bool {
	effective := clampInt(val+modifier, 0, 100)
	return rand.Intn(101) <= effective
}

// CompetitiveAttributeRoll performs a contest between two actors using the same attribute.
func (a *Actor) CompetitiveAttributeRoll(other *Actor, attr string) bool {
	return a.CompetitiveContest(other, attr, attr)
}

// CompetitiveContest performs a contest between two actors using different attributes.
// Each actor rolls 0-100. If only one succeeds (roll <= attrVal), they win.
// If both succeed, the one with the LOWER roll wins.
func (a *Actor) CompetitiveContest(other *Actor, myAttr, theirAttr string) bool {
	valA := a.getAttrValue(myAttr)
	valB := other.getAttrValue(theirAttr)

	rollA := rand.Intn(101)
	rollB := rand.Intn(101)

	successA := rollA <= valA
	successB := rollB <= valB

	if successA && !successB { return true }
	if !successA && successB { return false }
	if successA && successB {
		return rollA < rollB // Lower roll wins
	}
	return false
}

func (a *Actor) getAttrValue(attr string) int {
	attr = strings.ToLower(attr)
	val := 0
	switch attr {
	case "strength":  val = a.PrimaryAttributes.Strength
	case "dexterity": val = a.PrimaryAttributes.Dexterity
	case "health":    val = a.PrimaryAttributes.Health
	case "intellect": val = a.PrimaryAttributes.Intellect
	case "wisdom":    val = a.PrimaryAttributes.Wisdom
	}

	if (attr == "intellect" || attr == "wisdom") && a.TemporalState.Arousal > 10 {
		penalty := int(a.TemporalState.Arousal * 0.5)
		val -= penalty
		if val < 1 { val = 1 }
	}
	return val
}

// GetAbilityYield returns the numerical output of a productive action (e.g. litres, units, potency).
// Formulas are defined in GEMINI.md based on derived productive stats.
func (a *Actor) GetAbilityYield(abilityID string) float64 {
	switch abilityID {
	case "milk":
		return float64(a.Husbandry) * 1.0
	case "shear":
		return float64(a.Husbandry) * 0.5
	case "forage":
		return float64(a.Survivalism) * 0.3
	case "cook", "brew":
		return float64(a.Herbalism) * 1.0
	case "rest":
		return float64(a.Nourishment) * 0.25
	case "eat":
		return float64(a.Nourishment) * 1.0
	case "drink":
		return float64(a.Nourishment) * 0.8
	case "mate":
		return a.Mate * 1.0
	case "plant":
		return float64(a.Harvesting) * 0.8
	case "harvest_crop":
		return float64(a.Harvesting) * 1.5
	case "water_crops":
		return float64(a.Harvesting) * 0.5
	case "fish", "butcher", "guard", "stash":
		return float64(a.Survivalism) * 0.5
	case "hunt":
		return float64(a.Survivalism) * 1.0
	case "trap":
		return float64(a.Survivalism) * 0.8
	case "craft", "build":
		return float64(a.Crafting) * 1.0
	case "repair":
		return float64(a.Crafting) * 0.5
	case "smelt", "tan":
		return float64(a.Crafting) * 0.8
	case "trade":
		return float64(a.Trading) * 1.0
	case "appraise":
		return float64(a.Trading) * 0.5
	case "haul":
		return float64(a.PrimaryAttributes.Strength) * 0.01
	case "sneak":
		return float64(a.PrimaryAttributes.Dexterity) * 1.0
	case "steal":
		return float64(a.PrimaryAttributes.Dexterity) * 0.5
	case "pray", "bury":
		return float64(a.Culture) * 0.3
	case "heal":
		return float64(a.Herbalism) * 1.0
	case "teach":
		return float64(a.Culture) * 0.5
	case "intimidate", "recruit", "lie", "seduce", "perform":
		return float64(a.Culture) * 1.0
	case "compose", "read":
		return float64(a.Culture) * 0.5
	case "paint", "sculpt":
		return float64(a.Art) * 0.8
	case "weave":
		return float64(a.Crafting) * 0.5
	}
	return 0.0
}

func (a *Actor) AlleviateOnSelf(ctx *SystemContext) {
	a.TemporalState.Miccionate = 0
	a.TemporalState.Defecate = 0
	a.TemporalState.Hygiene -= 40
	a.TemporalState.Pain = 0 // Immediate relief from urgent distress
	if a.TemporalState.Hygiene < 0 { a.TemporalState.Hygiene = 0 }
	
	// Spawning waste on ground
	a.SpawnDefecation(ctx)
	
	// Torture logic: if near victims, they get soiled
	a.TransferSoilingToVictims(ctx, 40.0)
}

func (a *Actor) TakeBath() {
	a.TemporalState.Hygiene = 100
}

func (a *Actor) AlleviateProperly(ctx *SystemContext) {
	a.TemporalState.Miccionate = 0
	a.TemporalState.Defecate = 0
	a.TemporalState.Hygiene -= 5.0
	a.TemporalState.Pain = 0 // Immediate relief from urgent distress
	if a.TemporalState.Hygiene < 0 { a.TemporalState.Hygiene = 0 }

	// Torture logic (can still soil victims if doing it intentionally near them)
	a.TransferSoilingToVictims(ctx, 30.0)
}

func (a *Actor) TransferSoilingToVictims(ctx *SystemContext, amount float64) {
	if ctx == nil || ctx.World == nil { return }
	for _, other := range ctx.World.Characters {
		if other.IsIncapacitated() && other.IsAlive() {
			dist := math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2))
			if dist < 1.0 {
				other.TemporalState.Hygiene -= amount
				other.TemporalState.Sanity -= 15.0 // Traumatic
				if other.TemporalState.Hygiene < 0 { other.TemporalState.Hygiene = 0 }
				if other.TemporalState.Sanity < 0 { other.TemporalState.Sanity = 0 }
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
					Text: "Soiled!", X: other.X, Y: other.Y, Life: 60, Color: ColorHarm,
				})
			}
		}
	}
}

func (a *Actor) SpawnDefecation(ctx *SystemContext) {
	if ctx == nil || ctx.World == nil { return }
	config := ctx.Registries.Obstacles.Archetypes["defecation"]
	if config == nil { return }
	
	id := fmt.Sprintf("waste_%d", ctx.World.DayTick + int(a.X * 100))
	obs := NewObstacle(id, a.X, a.Y, config)
	ctx.World.Obstacles = append(ctx.World.Obstacles, obs)
}
