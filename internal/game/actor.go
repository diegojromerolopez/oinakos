package game

func (a *Actor) IsAlive() bool { return a.ActionState != ActorDead }
func (a *Actor) IsTrulyDead() bool { return a.ActionState == ActorDead }
func (a *Actor) IsIncapacitated() bool { return a.ActionState == ActorIncapacitated }
func (a *Actor) GetDeathThreshold() int { return -int(float64(a.GetTotalMaxHealth()) * 0.10) }

func (a *Actor) SyncLifeStatus() {
	if a.ActionState == ActorDead { return }
	threshold := a.GetDeathThreshold()
	if a.State.HealthPoints <= threshold {
		a.State.HealthPoints, a.ActionState = threshold, ActorDead
		return
	}
	
	// CRITICAL: Survival actions (Drinking/Eating) take priority over the 'Incapacitated' state
	// even if HealthPoints == 0 or UnconsciousTimer > 0. This allows characters to 'crawl' to sources.
	if a.ActionState == ActorDrinking || a.ActionState == ActorEating {
		return 
	}

	if a.State.HealthPoints <= 0 || a.UnconsciousTimer > 0 {
		if a.ActionState != ActorIncapacitated { a.ActionState = ActorIncapacitated }
	} else if a.ActionState == ActorIncapacitated { 
		a.ActionState = ActorIdle 
	}
}

func (a *Actor) InitBodyStatus() {
	if a.BodyStatus == nil { a.BodyStatus = make(map[string]int) }
	maxH := a.GetTotalMaxHealth()
	a.BodyStatus["head"], a.BodyStatus["torso"] = maxH/4, maxH/2
	a.BodyStatus["l_arm"], a.BodyStatus["r_arm"] = maxH/4, maxH/4
	a.BodyStatus["l_leg"], a.BodyStatus["r_leg"] = maxH/4, maxH/4
}

func (a *Actor) GetActor() *Actor { return a }

func (a *Actor) SyncState() {
	// Clamp values using floor for consistency as requested
	a.State.Hunger = clampFloat(a.State.Hunger, 0, 100)
	a.State.Thirst = clampFloat(a.State.Thirst, 0, 100)
	a.State.Fatigue = clampFloat(a.State.Fatigue, 0, 100)
	a.State.Sanity = clampFloat(a.State.Sanity, 0, 100)
	a.State.Arousal = clampFloat(a.State.Arousal, 0, 100)
	a.State.Pain = clampFloat(a.State.Pain, 0, 100)
	a.State.Hygiene = clampFloat(a.State.Hygiene, 0, 100)
	a.State.BladderLevel = clampFloat(a.State.BladderLevel, 0, 100)
	a.State.BowelLevel = clampFloat(a.State.BowelLevel, 0, 100)
	a.State.AlcoholLevel = clampFloat(a.State.AlcoholLevel, 0, 100)
	a.IsConscious = a.State.IsConscious

	if a.State.HealthPoints <= a.GetDeathThreshold() && a.ActionState != ActorDead {
		a.ActionState = ActorDead
	}
}

func (a *Actor) Kill(reason string) {
	a.ActionState = ActorDead
	a.State.HealthPoints = a.GetDeathThreshold()
}

func (a *Actor) GetTotalMaxHealth() int { return a.State.MaxHealthPoints + a.MaxHealthBonus }

type ActorInterface interface {
	GetActor() *Actor
	Heal(amount int)
}
