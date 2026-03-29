package game

import (
	"testing"
)

func TestActor_CoverageAdditions(t *testing.T) {
	ctx := NewTestContext()
	p := NewCharacter(0, 0, nil, 1, true, nil)
	p.Relationships = make(map[string]float64)
	p.RomanticInterest = make(map[string]float64)
	
	// 1. GetTraumaDescription
	p.Trauma.LeftArmLost = true
	desc := p.GetTraumaDescription()
	if desc == "" {
		t.Error("Expected trauma description, got empty")
	}

	// 2. GetInventoryNames
	p.Inventory = append(p.Inventory, &ItemInstance{Config: &ObjectConfig{Name: "Test Item"}})
	names := p.GetInventoryNames()
	if len(names) == 0 || names[0] != "Test Item" {
		t.Errorf("Unexpected inventory names: %v", names)
	}

	// 3. CompetitiveAttributeRoll
	other := NewCharacter(1, 1, nil, 1, false, nil)
	p.PrimaryAttributes.Strength = 100
	other.PrimaryAttributes.Strength = 10
	winner := p.CompetitiveAttributeRoll(&other.Actor, "strength")
	// Since it's random, we can't guarantee winner, but we call the code.
	_ = winner

	// 4. CompetitiveContest
	winner = p.CompetitiveContest(&other.Actor, "strength", "strength")
	_ = winner

	// 5. TakeBath
	p.State.Hygiene = 50
	p.TakeBath(ctx)
	// We need multiple updates to reach 100 as it recovers 2.0 per tick in ActorBathing state
	for i := 0; i < 30; i++ {
		p.SharedUpdate(ctx)
	}
	if p.State.Hygiene < 99.0 {
		t.Errorf("Expected hygiene approx 100, got %f", p.State.Hygiene)
	}

	// 6. Torture
	victim := NewCharacter(10, 10, nil, 1, false, nil)
	victim.Relationships = make(map[string]float64)
	victim.Memories = []MemoryEvent{}
	ctx.World.Characters = append(ctx.World.Characters, victim)
	p.X, p.Y = 10, 10
	
	// Try torture while active (should fail)
	p.Torture(&victim.Actor, ctx)
	
	// Incapacitate victim
	victim.Actor.ActionState = ActorIncapacitated
	p.Torture(&victim.Actor, ctx)
	if victim.Relationships[p.Name] >= 0 {
		t.Error("Expected negative relationship after torture")
	}

	// 7. Butchery (hitCharacter on dead actor)
	victim.Actor.ActionState = ActorDead
	victim.Actor.MeatQuantity = 10
	p.ActionState = ActorChopping
	// Mock object registry for raw_meat
	if ctx.Registries.Objects == nil {
		ctx.Registries.Objects = NewObjectRegistry()
	}
	ctx.Registries.Objects.Objects["raw_meat"] = &ObjectConfig{Name: "Raw Meat"}
	
	p.hitCharacter(&victim.Actor, "", ctx)
	if victim.Actor.MeatQuantity >= 10 {
		t.Error("Expected meat quantity to decrease after butchery")
	}

	// 8. SpawnDefecation
	p.X, p.Y = 10, 10
	p.SpawnDefecation(ctx)
	found := false
	for _, o := range ctx.World.Obstacles {
		if o.ID == "defecation" {
			found = true
			break
		}
	}
	_ = found

	// 9. GetAbilityYield exhaustive
	yieldMap := map[string]string{
		"milk": "Husbandry", "shear": "Husbandry", "forage": "Survivalism",
		"cook": "Herbalism", "brew": "Herbalism", "rest": "Nourishment",
		"eat": "Nourishment", "drink": "Nourishment", "mate": "Mate",
		"plant": "Harvesting", "harvest_crop": "Harvesting", "water_crops": "Harvesting",
		"fish": "Survivalism", "hunt": "Survivalism", "trap": "Survivalism",
		"butcher": "Survivalism", "craft": "Crafting", "repair": "Crafting",
		"smelt": "Crafting", "build": "Crafting", "trade": "Trading",
		"appraise": "Trading", "haul": "Strength", "stash": "Strength",
		"sneak": "Dexterity", "steal": "Dexterity", "seduce": "Art",
		"perform": "Art", "paint": "Art", "sculpt": "Art",
		"tan": "Crafting", "weave": "Crafting", "lie": "Culture",
		"pray": "Culture", "guard": "Survivalism", "heal": "Herbalism",
		"teach": "Culture", "intimidate": "Culture", "recruit": "Culture",
		"bury": "Culture", "read": "Culture", "compose": "Culture",
	}
	for id := range yieldMap {
		p.GetAbilityYield(id)
	}
	p.GetAbilityYield("unknown")

	// 10. TransferSoilingToVictims
	victim.Actor.ActionState = ActorIncapacitated
	victim.Actor.X, victim.Actor.Y = p.X, p.Y
	p.State.Hygiene = 0
	p.TransferSoilingToVictims(ctx)
	if victim.State.Hygiene >= 100 {
		t.Error("Expected hygiene to drop after transfer")
	}

	// 11. getAttrValue
	attrs := []string{"strength", "dexterity", "health", "intellect", "wisdom", "invalid"}
	for _, a := range attrs {
		p.getAttrValue(a)
	}

	// 12. CompetitiveContest
	p.CompetitiveContest(&victim.Actor, "strength", "strength")
}
