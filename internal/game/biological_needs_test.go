package game

import (
	"math"
	"testing"
)

func TestBiologicalNeeds_Mechanics(t *testing.T) {
	ctx := NewTestContext()
	
	t.Run("Decay Multiplier and Weather", func(t *testing.T) {
		a := &Actor{Name: "Need Tester", ActionState: ActorIdle}
		a.State.HealthPoints = 100
		a.State.MaxHealthPoints = 100
		a.PrimaryAttributes.Health = 100
		
		ctx.World.State.Weather = WeatherClear
		initialHunger := a.State.Hunger
		a.updateNeeds(ctx)
		hungerChange := a.State.Hunger - initialHunger
		
		if math.Abs(hungerChange - 0.0016875) > 0.0001 {
			t.Errorf("Expected hunger change 0.0016875, got %v", hungerChange)
		}
		
		ctx.World.State.Weather = WeatherRain
		ctx.World.State.Intensity = 1.0
		a.State.Hunger = 0
		a.updateNeeds(ctx)
		hungerChange = a.State.Hunger
		if math.Abs(hungerChange - 0.00253125) > 0.0001 {
			t.Errorf("Expected hunger change 0.00253125, got %v", hungerChange)
		}
	})

	t.Run("Retentive Pain Logic", func(t *testing.T) {
		a := &Actor{Name: "Need Tester", ActionState: ActorIdle}
		a.State.HealthPoints = 100
		a.State.MaxHealthPoints = 100
		a.State.BladderLevel = 90
		a.State.BowelLevel = 90
		a.Tick = 300
		initialPain := a.State.Pain
		a.updateNeeds(ctx)
		if a.State.Pain <= initialPain {
			t.Errorf("Pain should increase due to full bladder/bowel at 90%%")
		}
		
		a.State.BladderLevel = 100
		a.Tick = 600
		a.updateNeeds(ctx)
		if a.State.Pain != 0 {
			t.Errorf("Pain should reset to 0 after soiling pants at 100%%")
		}
	})

	t.Run("Alcohol Metabolism", func(t *testing.T) {
		a := &Actor{Name: "Drunkard", ActionState: ActorIdle}
		a.State.HealthPoints = 100
		a.State.MaxHealthPoints = 100
		a.State.AlcoholLevel = 10
		a.State.IsDrunk = true
		a.updateNeeds(ctx)
		if a.State.AlcoholLevel >= 10 {
			t.Errorf("AlcoholLevel should decay")
		}
		
		a.State.AlcoholLevel = 0.0001
		a.updateNeeds(ctx)
		if a.State.IsDrunk {
			t.Errorf("AlcoholLevel 0 should sober up the actor")
		}
	})

	t.Run("Action Impacts", func(t *testing.T) {
		a := &Actor{Name: "Consumer", State: State{Hunger: 50, Thirst: 50, Hygiene: 50, HealthPoints: 100, MaxHealthPoints: 100}}
		a.Inventory = []*ItemInstance{
			{Config: &ObjectConfig{Name: "Water", Refillable: true, MaxLiquid: 1.0}, LiquidContent: 1.0, LiquidMax: 1.0, Refillable: true},
		}
		
		a.ActionState = ActorDrinking
		a.updateNeeds(ctx)
		if a.State.Thirst >= 50 { t.Errorf("Drinking should satiate thirst, got %v", a.State.Thirst) }
		
		a.ActionState = ActorEating
		a.updateNeeds(ctx)
		if a.State.Hunger >= 50 { t.Errorf("Eating should satiate hunger") }
		
		a.ActionState = ActorBathing
		a.updateNeeds(ctx)
		if a.State.Hygiene <= 50 { t.Errorf("Bathing should increase hygiene") }

		a.State.Hunger = 50
		a.ActionState = ActorWalking
		a.updateNeeds(ctx)
		if a.State.Hunger <= 50 { t.Errorf("Walking costs more hunger") }

		a.ActionState = ActorAttacking
		a.updateNeeds(ctx)
		// Fatigue increases by base (0.001) + action (0.01) = 0.011
		if a.State.Fatigue < 0.01 { t.Errorf("Attacking costs fatigue, got %v", a.State.Fatigue) }
	})

	t.Run("Resting Mechanics", func(t *testing.T) {
		ctx.World.DayTick = TicksPerHour * 8 // Day time
		ctx.World.Obstacles = []*Obstacle{
			{ID: "campfire", Alive: true, X: 0, Y: 0},
		}
		a := &Actor{Name: "Rester", ActionState: ActorResting, X: 0, Y: 0, State: State{Fatigue: 50, HealthPoints: 10, MaxHealthPoints: 100, Sanity: 100}}
		a.Tick = 60
		a.updateNeeds(ctx)
		if a.State.Fatigue >= 50 { t.Errorf("Resting should decrease fatigue, got %v", a.State.Fatigue) }
		if a.State.HealthPoints <= 10 { t.Errorf("Resting should heal") }

		// Full rest wake up
		a.State.Fatigue = 0.001
		a.updateNeeds(ctx)
		if a.ActionState != ActorIdle { t.Errorf("Should wake up after fully resting, got state %v", a.ActionState) }
	})

	t.Run("Status Thresholds and Penalties", func(t *testing.T) {
		a := &Actor{Name: "Neglected", ActionState: ActorIdle, State: State{HealthPoints: 100, MaxHealthPoints: 100, Hunger: 100}}
		a.Tick = TicksPerHour
		a.updateNeeds(ctx)
		if a.State.HealthPoints >= 100 { t.Errorf("Starvation at 100%% should drain HP") }

		a.ActionState = ActorIncapacitated
		a.State.HealthPoints = 100
		a.updateNeeds(ctx)
		if a.State.HealthPoints >= 100 { t.Errorf("Incapacitated should bleed out HP over time") }
	})


}
