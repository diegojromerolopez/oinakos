package main

import (
	"fmt"
	"math/rand"
	"oinakos/internal/game"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== OINAKOS GENERATIONAL EVOLUTION SIMULATOR ===")
	rand.Seed(time.Now().UnixNano())
	
	// 1. Setup Mock Containers
	world := &game.World{
		State: game.WorldState{ Ticks: 0 },
	}
	ctx := &game.SystemContext{
		World: world,
		Registries: &game.RegistryContainer{
			Archetypes: &game.ArchetypeRegistry{
				Archetypes: make(map[string]*game.EntityConfig),
			},
			Objects: &game.ObjectRegistry{},
		},
		Log: func(msg string, category game.LogCategory) {
			if strings.Contains(msg, "grown into") || strings.Contains(msg, "has given birth") {
				fmt.Printf("[ENGINE LOG] %s\n", msg)
			}
		},
		Settings: &game.Settings{ AdultMode: true },
	}
	
	// Mock Registry setup
	ctx.Registries.Archetypes.Archetypes["archetypes/baby/male"] = &game.EntityConfig{ID: "baby_male", Name: "Baby Male", Archetype: "baby/male"}
	ctx.Registries.Archetypes.Archetypes["archetypes/kid/male"] = &game.EntityConfig{ID: "kid_male", Name: "Kid Male", Archetype: "kid/male"}
	ctx.Registries.Archetypes.Archetypes["archetypes/teenager/male"] = &game.EntityConfig{ID: "teenager_male", Name: "Teenager Male", Archetype: "teenager/male"}
	ctx.Registries.Archetypes.Archetypes["archetypes/man_at_arms"] = &game.EntityConfig{ID: "man_at_arms", Name: "Warrior", Archetype: "man_at_arms"}
	ctx.Registries.Archetypes.Archetypes["archetypes/peasant"] = &game.EntityConfig{ID: "peasant", Name: "Peasant", Archetype: "peasant"}

	// 2. Setup High-Stat Parents (to influence child career)
	mother := game.NewCharacter(0, 0, &game.EntityConfig{Gender: "female"}, 1, false, ctx.Registries.Objects)
	mother.Name = "Sylvanas"
	mother.PrimaryAttributes = game.PrimaryAttributes{Strength: 90, Intellect: 20} // Warrior family
	
	father := game.NewCharacter(0, 0, &game.EntityConfig{Gender: "male"}, 1, false, ctx.Registries.Objects)
	father.Name = "Arthas"
	father.PrimaryAttributes = game.PrimaryAttributes{Strength: 95, Intellect: 15}
	
	world.Characters = append(world.Characters, mother, father)
	
	// 3. Trigger Generational Event
	fmt.Println("Simulating Pregnancy & Birth...")
	mother.IsPregnant = true
	mother.FatherID = father.Name
	mother.GestationTicks = 1 
	mother.SharedUpdate(ctx) // birth!
	
	if len(world.Characters) < 3 { return }
	child := world.Characters[2]
	fmt.Printf("[OFFSPRING] %s born! Stats: Str:%d, Int:%d\n", child.Name, child.PrimaryAttributes.Strength, child.PrimaryAttributes.Intellect)
	
	// 4. Time Warp: Age to Adulthood (18 Years)
	tpy := float64(game.TicksPerYear)
	milestones := []struct{ age float64; label string }{
		{1.1, "KID"},
		{12.1, "TEENAGER"},
		{18.1, "ADULT"},
	}
	
	for _, m := range milestones {
		fmt.Printf("Warping to %s boundary (Age %.1f)...\n", m.label, m.age)
		for (child.Actor.AgeTicks / tpy) < m.age {
			child.Actor.AgeTicks += tpy / 12.0 // Advance 1 month per tick
			child.Actor.SharedUpdate(ctx)
		}
	}

	fmt.Printf("[FINAL] Character: %s | LifeStage: %s | Archetype: %s (%s)\n", 
		child.Name, child.LifeStage, child.Config.ID, child.Config.Name)
	fmt.Println("=== SIMULATION COMPLETE ===")
}
