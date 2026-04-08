package game

import (
	"fmt"
	"math"
	"sync/atomic"
	"testing"
)

func TestAging_AttributeScaling(t *testing.T) {
	g := setupTestGame()
	p := g.playableCharacter
	// Use explicit 100 for all attributes
	p.PrimaryAttributes = PrimaryAttributes{Strength: 100, Dexterity: 100, Health: 100, Intellect: 100, Wisdom: 100}
	
	// Ensure we don't have overrides in RawStats that would complicate the test
	p.RawStats = EntityStats{}
	
	ticksPerYear := float64(TicksPerYear)
	
	scenarios := []struct {
		ageYears float64
		expectedStr int // 100 * pMult
		expectedWis int // 100 * mMult
	}{
		{0, 75, 70},   // -25% Phys, -30% Ment (at birth)
		{12.5, 87, 85}, // Middle of growth
		{25, 100, 100}, // Peak
		{40, 100, 100}, // Start of decline
		{60, 88, 110},  // 60-40=20y past peak. Phys: -11% (approx). Ment: +10% (2 decades)
		{85, 75, 120},  // 85y: -25% Phys. Ment: +20% (4 decades past 40. 50,60,70,80)
	}
	
	for _, s := range scenarios {
		p.AgeTicks = s.ageYears * ticksPerYear
		p.SyncStats(nil)
		
		// BaseAttack is str * 2. str = 100 * pMult.
		// So BaseAttack = 200 * pMult.
		pMult := 1.0
		if s.ageYears < 25 {
			pMult = 1.0 - 0.25 * (25.0 - s.ageYears) / 25.0
		} else if s.ageYears > 40 {
			pPenalty := 0.25 * (s.ageYears - 40.0) / (85.0 - 40.0)
			if pPenalty > 0.25 { pPenalty = 0.25 }
			pMult = 1.0 - pPenalty
		}
		expectedAttack := int(100.0 * pMult * 2.0)
		
		if p.BaseAttack != expectedAttack {
			t.Errorf("Age %v: expected Attack %d, got %d", s.ageYears, expectedAttack, p.BaseAttack)
		}
		
		// Mentals: MentMult = 1.0 + 0.05 * floor((age-40)/10)
		mMult := 1.0
		if s.ageYears < 25 {
			mMult = 1.0 - 0.30 * (25.0 - s.ageYears) / 25.0
		} else if s.ageYears > 40 {
			mMult = 1.0 + 0.05 * math.Floor((s.ageYears - 40.0) / 10.0)
		}
		
		// Trading formula: int(itl*1.2 + wis*0.3) = int(MentVal * 1.5)
		expectedMentVal := 100.0 * mMult
		expectedTrade := int(expectedMentVal * 1.5)
		
		if p.Trading != expectedTrade {
			t.Errorf("Age %v: expected Trading %d, got %d (mMult=%.2f)", s.ageYears, expectedTrade, p.Trading, mMult)
		}
	}
}

func TestSimulation_Longevity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping longevity test in short mode")
	}

	g := setupTestGame()
	
	// Add survival obstacles to the world
	// 1. Water source
	g.World.Obstacles = append(g.World.Obstacles, &Obstacle{
		ID: "well_0", X: 105, Y: 105, Alive: true,
		Archetype: &ObstacleArchetype{ID: "well", Type: TypeBuilding, Destructible: false},
	})
	// 2. Food source
	for i := 0; i < 5; i++ {
		g.World.Obstacles = append(g.World.Obstacles, &Obstacle{
			ID: fmt.Sprintf("apple_tree_%d", i), X: 110 + float64(i*2), Y: 110 + float64(i*2), Alive: true,
			Archetype: &ObstacleArchetype{ID: "tree_apple", Type: TypeTree, Yield: "apple", Weight: 100.0, Destructible: true},
		})
	}

	// 3. Hygiene/Excretion
	g.World.Obstacles = append(g.World.Obstacles, &Obstacle{
		ID: "latrine_0", X: 105, Y: 100, Alive: true,
		Archetype: &ObstacleArchetype{ID: "latrine", Type: TypeBuilding, Destructible: false},
	})
	g.World.Obstacles = append(g.World.Obstacles, &Obstacle{
		ID: "bath_0", X: 100, Y: 105, Alive: true,
		Archetype: &ObstacleArchetype{ID: "bath", Type: TypeBuilding, Destructible: false},
	})

	// Add apple item definition
	g.Registries.Objects.Objects["apple"] = &ObjectConfig{
		ID: "apple", Name: "Apple", Type: "consumable",
		Hunger: 20, Thirst: 5, Fatigue: 0, Weight: 0.1,
	}
	
	// Ensure the playable character doesn't die of neglect
	g.playableCharacter.Name = "Oinakos"
	g.playableCharacter.X, g.playableCharacter.Y = 110, 110
	g.playableCharacter.State.HealthPoints = 1000
	g.playableCharacter.State.MaxHealthPoints = 1000
	g.playableCharacter.Behavior = BehaviorChaotic
	g.playableCharacter.Alignment = AlignmentAlly
	g.playableCharacter.Shift = ShiftLeisure


	// Noah's Ark 2.0: 10 characters (5 male, 5 female)
	for i := 0; i < 5; i++ {
		m := spawnTestActor(g, "peasant_male", 100+float64(i*10), 100)
		m.Name = fmt.Sprintf("Adam_%d", i)
		m.LifeStage = StageAdult
		m.AgeTicks = 20.0 * float64(TicksPerYear)
		m.Behavior = BehaviorChaotic
		m.SyncStats(nil)
		
		f := spawnTestActor(g, "peasant_female", 100+float64(i*10), 110)
		f.Name = fmt.Sprintf("Eve_%d", i)
		f.LifeStage = StageAdult
		f.AgeTicks = 20.0 * float64(TicksPerYear)
		f.Behavior = BehaviorChaotic
		f.SyncStats(nil)
	}
	
	// Force pregnancy for Eve_0 (index 1)
	g.characters[1].IsPregnant = true
	g.characters[1].GestationTicks = (1 * TicksPerDay) // Birth tomorrow
	
	atomic.StoreInt32(&g.LoadingProgress, 1000)
	
	testPeriod := 2 * TicksPerDay // 2 Days
	g.World.State.Temperature = 37.0 // Perfect temp
	fmt.Printf("Starting simulation for 2 days (%d ticks)...\n", testPeriod)
	
	for i := 0; i < testPeriod; i++ {
		g.Update()
		
		if i % TicksPerDay == 0 {
			day := i / TicksPerDay
			pop := 0
			pregnantCount := 0
			for _, c := range g.characters {
				if c.IsAlive() { 
					pop++ 
					if c.IsPregnant {
						pregnantCount++
					}
				}
			}
			// Print every day
			c := g.characters[0]
			fmt.Printf("Day %d: Pop=%d, Preg=%d, AgeY=%.1f, HP=%d, State=%s, Reason=%s, Target=%s, H=%.1f, P=%.1f, Hy=%.1f\n", 
				day, pop, pregnantCount, float64(c.AgeTicks)/float64(TicksPerYear), 
				c.State.HealthPoints, c.ActionState.String(), c.LastAIReasoning, c.TargetActorID,
				c.State.Hunger, c.State.Pain, c.State.Hygiene)
		}
	}

	aliveCount := 0
	for _, c := range g.characters { if c.IsAlive() { aliveCount++ } }
	fmt.Printf("Simulation ended. Final Population: %d\n", aliveCount)
	if aliveCount == 0 {
		t.Errorf("Population went extinct after 2 days")
	}
}

func spawnTestActor(g *Game, archetype string, x, y float64) *Character {
	config := g.archetypeRegistry.Archetypes[archetype]
	c := NewCharacter(x, y, config, 1, false, nil)
	c.Actor.Shift = ShiftLeisure // Ensure they can mate/simulate
	g.characters = append(g.characters, c)
	g.World.Characters = append(g.World.Characters, c)
	return c
}
