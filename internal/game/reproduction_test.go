package game

import (
	"strings"
	"testing"
)

func TestReproductionCycle(t *testing.T) {
	ctx := NewTestContext()

	// 1. Setup Female and Male characters
	femaleConfig := &EntityConfig{ID: "peasant_female", Name: "Female NPC", Gender: "female"}
	maleConfig := &EntityConfig{ID: "peasant_male", Name: "Male NPC", Gender: "male"}

	mother := NewCharacter(10, 10, femaleConfig, 1, false, ctx.Registries.Objects)
	mother.SyncStats(ctx.Registries.Objects)
	mother.State.Age.Current = 25.0
	mother.AgeTicks = 25.0 * float64(TicksPerYear)

	father := NewCharacter(11, 11, maleConfig, 1, false, ctx.Registries.Objects)
	father.SyncStats(ctx.Registries.Objects)
	father.State.Age.Current = 25.0
	father.AgeTicks = 25.0 * float64(TicksPerYear)

	// 2. Test Pregnancy
	pregnant := false
	for i := 0; i < 100; i++ {
		mother.haveSex(ctx, &father.Actor, "vaginal")
		if mother.IsPregnant {
			pregnant = true
			break
		}
		mother.MatingCooldown = 0
	}

	if !pregnant {
		t.Errorf("Failed to make character pregnant after 100 attempts")
	}

	// 3. Test Gestation and Birth
	mother.GestationTicks = 10
	mother.updateBreeding(ctx)

	if mother.IsPregnant {
		t.Errorf("Expected mother to NOT be pregnant after gestation ends")
	}

	// 4. Verify Baby exists
	foundBaby := false
	for _, char := range ctx.World.Characters {
		if char.ParentID == mother.Name {
			foundBaby = true
			break
		}
	}
	if !foundBaby {
		t.Errorf("Offspring was not found after birth")
	}
}

func TestAnimalReproduction(t *testing.T) {
	ctx := NewTestContext()

	// Setup Sheep
	sheepConfig := &EntityConfig{ID: "sheep", Name: "Sheep", Gender: "female", IsAnimal: true}
	ramConfig := &EntityConfig{ID: "ram", Name: "Ram", Gender: "male", IsAnimal: true}
	lambConfig := &EntityConfig{ID: "lamb", Name: "Lamb", Gender: "female", IsAnimal: true}

	ctx.Registries.Archetypes.Archetypes["sheep"] = sheepConfig
	ctx.Registries.Archetypes.Archetypes["ram"] = ramConfig
	ctx.Registries.Archetypes.Archetypes["lamb"] = lambConfig

	mother := NewCharacter(10, 10, sheepConfig, 1, false, ctx.Registries.Objects)
	mother.SyncStats(ctx.Registries.Objects)
	mother.State.Age.Current = 3.0

	father := NewCharacter(11, 11, ramConfig, 1, false, ctx.Registries.Objects)
	father.SyncStats(ctx.Registries.Objects)
	father.State.Age.Current = 3.0

	pregnant := false
	for i := 0; i < 10; i++ {
		mother.haveSex(ctx, &father.Actor, "vaginal")
		if mother.IsPregnant {
			pregnant = true
			break
		}
		mother.MatingCooldown = 0
	}

	if !pregnant {
		t.Errorf("Failed to make sheep pregnant")
	}

	mother.GestationTicks = 1
	mother.updateBreeding(ctx)

	foundLamb := false
	for _, char := range ctx.World.Characters {
		if char.ParentID == mother.Name {
			foundLamb = true
			if !strings.Contains(strings.ToLower(char.Config.ID), "lamb") {
				t.Errorf("Expected sheep offspring to be a lamb, got %s", char.Config.ID)
			}
			break
		}
	}
	if !foundLamb {
		t.Errorf("Lamb was not born")
	}
}

func TestSameSexIntercourse(t *testing.T) {
	ctx := NewTestContext()
	female1 := NewCharacter(10, 10, &EntityConfig{ID: "f1", Gender: "female"}, 25, false, nil)
	female2 := NewCharacter(11, 11, &EntityConfig{ID: "f2", Gender: "female"}, 25, false, nil)

	for i := 0; i < 50; i++ {
		female1.haveSex(ctx, &female2.Actor, "vaginal")
		if female1.IsPregnant || female2.IsPregnant {
			t.Errorf("Pregnancy occurred in female-female encounter")
		}
		female1.MatingCooldown = 0
	}
}

func TestUnderagePregnancy(t *testing.T) {
	ctx := NewTestContext()
	
	// 1. Child (Age 10) - Still sterile
	childMother := NewCharacter(10, 10, &EntityConfig{ID: "child", Gender: "female"}, 1, false, nil)
	childMother.State.Age.Current = 10.0 
	adultFather := NewCharacter(11, 11, &EntityConfig{ID: "adult", Gender: "male"}, 25, false, nil)
	adultFather.State.Age.Current = 25.0

	for i := 0; i < 50; i++ {
		childMother.haveSex(ctx, &adultFather.Actor, "vaginal")
		if childMother.IsPregnant {
			t.Errorf("Pregnancy occurred in pre-pubescent character (age 10)")
		}
		childMother.MatingCooldown = 0
	}

	// 2. Teenager (Age 13) - Now fertile
	teenMother := NewCharacter(10, 10, &EntityConfig{ID: "teen", Gender: "female"}, 1, false, nil)
	teenMother.State.Age.Current = 13.0
	
	pregnant := false
	for i := 0; i < 100; i++ {
		teenMother.haveSex(ctx, &adultFather.Actor, "vaginal")
		if teenMother.IsPregnant {
			pregnant = true
			break
		}
		teenMother.MatingCooldown = 0
	}

	if !pregnant {
		t.Errorf("Teenager (age 13) failed to get pregnant after 100 attempts")
	}
}

func TestElderlyIntercourse(t *testing.T) {
	ctx := NewTestContext()
	elderMother := NewCharacter(10, 10, &EntityConfig{ID: "elder", Gender: "female"}, 60, false, nil)
	elderMother.State.Age.Current = 60.0 // Menopause is at 45
	adultFather := NewCharacter(11, 11, &EntityConfig{ID: "adult", Gender: "male"}, 25, false, nil)
	adultFather.State.Age.Current = 25.0

	for i := 0; i < 50; i++ {
		elderMother.haveSex(ctx, &adultFather.Actor, "vaginal")
		if elderMother.IsPregnant {
			t.Errorf("Pregnancy occurred in elderly character (age 60)")
		}
		elderMother.MatingCooldown = 0
	}
}

func TestTransexualReproduction(t *testing.T) {
	ctx := NewTestContext()
	tsMother := NewCharacter(10, 10, &EntityConfig{ID: "ts", Gender: "female"}, 25, false, nil)
	tsMother.IsTransexual = true
	tsMother.State.Age.Current = 25.0
	adultFather := NewCharacter(11, 11, &EntityConfig{ID: "adult", Gender: "male"}, 25, false, nil)
	adultFather.State.Age.Current = 25.0

	for i := 0; i < 50; i++ {
		tsMother.haveSex(ctx, &adultFather.Actor, "vaginal")
		if tsMother.IsPregnant {
			t.Errorf("Pregnancy occurred in transexual character")
		}
		tsMother.MatingCooldown = 0
	}
}

func TestOralSexConstraints(t *testing.T) {
	ctx := NewTestContext()
	mother := NewCharacter(10, 10, &EntityConfig{ID: "f", Gender: "female"}, 25, false, nil)
	father := NewCharacter(11, 11, &EntityConfig{ID: "m", Gender: "male"}, 25, false, nil)
	mother.State.Age.Current = 25.0

	for i := 0; i < 50; i++ {
		mother.haveSex(ctx, &father.Actor, "cunnilingus")
		if mother.IsPregnant {
			t.Errorf("Pregnancy occurred after oral sex")
		}
		mother.MatingCooldown = 0
	}
}

func TestTribadismConstraints(t *testing.T) {
	ctx := NewTestContext()
	f1 := NewCharacter(10, 10, &EntityConfig{ID: "f1", Gender: "female"}, 25, false, nil)
	f2 := NewCharacter(11, 11, &EntityConfig{ID: "f2", Gender: "female"}, 25, false, nil)

	for i := 0; i < 50; i++ {
		f1.haveSex(ctx, &f2.Actor, "tribadism")
		if f1.IsPregnant || f2.IsPregnant {
			t.Errorf("Pregnancy occurred after tribadism")
		}
		f1.MatingCooldown = 0
	}
}

func TestBestialityConstraints(t *testing.T) {
	ctx := NewTestContext()
	human := NewCharacter(10, 10, &EntityConfig{ID: "human", Gender: "female", IsAnimal: false}, 25, false, nil)
	animal := NewCharacter(11, 11, &EntityConfig{ID: "animal", Gender: "male", IsAnimal: true}, 5, false, nil)
	human.State.Age.Current = 25.0
	startSanity := human.State.Sanity

	for i := 0; i < 50; i++ {
		human.haveSex(ctx, &animal.Actor, "bestiality")
		if human.IsPregnant {
			t.Errorf("Pregnancy occurred after bestiality")
		}
		human.MatingCooldown = 0
	}

	if human.State.Sanity >= startSanity {
		t.Errorf("Expected sanity penalty for human engaging in bestiality, but sanity was %.2f", human.State.Sanity)
	}
}

func TestAnalSexConstraints(t *testing.T) {
	ctx := NewTestContext()
	mother := NewCharacter(10, 10, &EntityConfig{ID: "f", Gender: "female"}, 25, false, nil)
	father := NewCharacter(11, 11, &EntityConfig{ID: "m", Gender: "male"}, 25, false, nil)
	mother.State.Age.Current = 25.0

	for i := 0; i < 50; i++ {
		father.haveSex(ctx, &mother.Actor, "anal")
		if mother.IsPregnant {
			t.Errorf("Pregnancy occurred after anal sex")
		}
		father.MatingCooldown = 0
	}
}

func TestPregnancyLock(t *testing.T) {
	ctx := NewTestContext()
	mother := NewCharacter(10, 10, &EntityConfig{ID: "f", Gender: "female"}, 25, false, nil)
	father1 := NewCharacter(11, 11, &EntityConfig{ID: "m1", Name: "Father1", Gender: "male"}, 25, false, nil)
	father2 := NewCharacter(12, 12, &EntityConfig{ID: "m2", Name: "Father2", Gender: "male"}, 25, false, nil)
	mother.State.Age.Current = 25.0

	// 1. Get pregnant from Father1
	mother.haveSex(ctx, &father1.Actor, "vaginal")
	mother.IsPregnant = true // Force it for test reliability
	mother.FatherID = "Father1"
	mother.GestationTicks = 100

	// 2. Attempt to get pregnant from Father2
	mother.MatingCooldown = 0
	mother.haveSex(ctx, &father2.Actor, "vaginal")

	// 3. Check that FatherID and Gestation are NOT reset
	if mother.FatherID != "Father1" {
		t.Errorf("FatherID was overwritten during active pregnancy")
	}
	if mother.GestationTicks != 100 {
		t.Errorf("GestationTicks were reset during active pregnancy")
	}
}

func TestGeneticMutation(t *testing.T) {
	ctx := NewTestContext()
	mother := NewCharacter(10, 10, &EntityConfig{ID: "f", Gender: "female"}, 25, false, nil)
	father := NewCharacter(11, 11, &EntityConfig{ID: "m", Gender: "male"}, 25, false, nil)
	
	// Set specific attributes
	mother.PrimaryAttributes = PrimaryAttributes{Strength: 50, Dexterity: 50, Health: 50, Intellect: 50, Wisdom: 50}
	father.PrimaryAttributes = PrimaryAttributes{Strength: 50, Dexterity: 50, Health: 50, Intellect: 50, Wisdom: 50}
	mother.FatherID = father.Name
	mother.IsPregnant = true
	mother.GestationTicks = 0 // Immediate birth
	
	ctx.World.Characters = append(ctx.World.Characters, mother, father)

	mother.updateBreeding(ctx)
	
	child := ctx.World.Characters[len(ctx.World.Characters)-1]
	// If mutation works, Strength won't be exactly 50 (in some runs, we check it differs)
	// Since mutation is random, we check that it's within 47-53 (50 * 0.95 to 50 * 1.05)
	if child.PrimaryAttributes.Strength < 47 || child.PrimaryAttributes.Strength > 53 {
		t.Errorf("Mutation out of expected bounds: %d", child.PrimaryAttributes.Strength)
	}
}

func TestIncestPenalty(t *testing.T) {
	ctx := NewTestContext()
	parent := NewCharacter(10, 10, &EntityConfig{ID: "parent", Name: "Parent", Gender: "male"}, 40, false, nil)
	child := NewCharacter(11, 11, &EntityConfig{ID: "child", Name: "Child", Gender: "female"}, 20, false, nil)
	child.ParentID = "Parent"
	child.State.Age.Current = 20.0
	
	startSanity := child.State.Sanity
	child.haveSex(ctx, &parent.Actor, "vaginal")

	if child.State.Sanity >= startSanity {
		t.Errorf("Expected sanity penalty for incestuous act, got %.2f", child.State.Sanity)
	}
}

func TestMiscarriage(t *testing.T) {
	ctx := NewTestContext()
	mother := NewCharacter(10, 10, &EntityConfig{ID: "f", Gender: "female"}, 25, false, nil)
	mother.IsPregnant = true
	mother.GestationTicks = 1000
	mother.State.Pain = 90.0 // Critical pain

	// Run update search for miscarriage (low chance, so we loop)
	miscarried := false
	for i := 0; i < 1000; i++ {
		mother.updateBreeding(ctx)
		if !mother.IsPregnant {
			miscarried = true
			break
		}
	}

	if !miscarried {
		t.Errorf("Expected miscarriage from extreme pain")
	}
}

func TestMultipleBirths(t *testing.T) {
	ctx := NewTestContext()
	
	// Setup Piglet archetype
	ctx.Registries.Archetypes.Archetypes["piglet"] = &EntityConfig{ID: "piglet", IsAnimal: true}

	mother := NewCharacter(10, 10, &EntityConfig{ID: "pig", Gender: "female", IsAnimal: true}, 3, false, nil)
	father := NewCharacter(11, 11, &EntityConfig{ID: "boar", Gender: "male", IsAnimal: true}, 3, false, nil)
	mother.FatherID = father.Name
	mother.IsPregnant = true
	mother.GestationTicks = 0

	mother.updateBreeding(ctx)

	count := 0
	for _, char := range ctx.World.Characters {
		if char.ParentID == mother.Name {
			count++
		}
	}

	if count < 4 {
		t.Errorf("Expected a litter of piglets (at least 4), got %d", count)
	}
}
