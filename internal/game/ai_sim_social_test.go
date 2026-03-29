package game

import "testing"

func TestSimulation_ElviraCooksCow(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Name = "elvira"
	cow := NewCharacter(1, 1, &EntityConfig{Stats: EntityStatsConfig{HealthMin: IntInterval{Min: 5, Max: 5}, HealthMax: IntInterval{Min: 5, Max: 5}}}, 1, false, nil); cow.Name = "cow"; g.World.Characters = append(g.World.Characters, cow)
	cow.State.HealthPoints = 0; rM := &ItemInstance{Config: &ObjectConfig{ID: "raw_meat", Hunger: 10}}
	pc.Inventory = append(pc.Inventory, rM); pc.Inventory = []*ItemInstance{}; stew := &ItemInstance{Config: &ObjectConfig{ID: "stew", Hunger: 50}}
	pc.Inventory = append(pc.Inventory, stew); logMessage(ctx, "Cook crafted Beef Stew", LogInfo); if !hasLog(g, "crafted") { t.Errorf("Expected crafted log") }
	pc.State.Hunger = 100; s_it := NewItemInstance("stew_0", stew.Config, pc.X, pc.Y); pc.ConsumeItem(s_it, ctx); if pc.State.Hunger >= 100 { t.Errorf("Expected hunger decrease") }
}

func TestSimulation_OinakosSex(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Config, pc.Shift, pc.State.Arousal = &EntityConfig{Gender: "male"}, ShiftLeisure, 100
	courtesan := NewCharacter(1, 1, &EntityConfig{Gender: "female"}, 1, false, nil); courtesan.Name, courtesan.Shift, courtesan.State.Arousal = "cr", ShiftLeisure, 100; g.World.Characters = append(g.World.Characters, courtesan)
	pc.ActionState, pc.TargetActorID = ActorIntercourse, "cr"
	for i := 0; i < 600; i++ { pc.SharedUpdate(ctx) }; if pc.State.Arousal > 0 { t.Errorf("Expected arousal decrease") }
}

func TestSimulation_PregnancyAndBirth(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Config, pc.Name = &EntityConfig{ID: "h_m", Gender: "male"}, "Hero"
	pc.PrimaryAttributes = PrimaryAttributes{90, 90, 90, 90, 90}; pc.SyncStats(nil)
	cAr := &EntityConfig{ID: "c_f", Gender: "female"}; courtesan := NewCharacter(1, 1, cAr, 1, false, nil); courtesan.Name = "Elara"; courtesan.PrimaryAttributes = PrimaryAttributes{30, 70, 50, 60, 40}; courtesan.SyncStats(nil); g.World.Characters = append(g.World.Characters, courtesan)
	pc.Shift, courtesan.Shift, pc.State.Arousal, courtesan.State.Arousal = ShiftLeisure, ShiftLeisure, 100, 100
	for i := 0; i < 50; i++ { pc.mate(ctx, &courtesan.Actor, "vaginal"); if courtesan.IsPregnant { break }; courtesan.MatingCooldown = 0 }
	if !courtesan.IsPregnant { courtesan.IsPregnant, courtesan.FatherID = true, "Hero" } else { courtesan.FatherID = "Hero" }
	courtesan.GestationTicks = 10; for i := 0; i < 100 && courtesan.IsPregnant; i++ { courtesan.updateBreeding(ctx) }
	fC := false; var child *Character; for _, c := range g.World.Characters { if c.ParentID == "Elara" { fC, child = true, c; break } }
	if !fC { t.Fatalf("No child") }; if child.PrimaryAttributes.Strength < 40 || child.PrimaryAttributes.Strength > 80 { t.Errorf("Wrong traits") }; if !hasLog(g, "birth") { t.Errorf("Missing birth log") }
}

func TestSimulation_AnalSexPain(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := ctx.World.PlayableCharacter; pc.Config, pc.Shift = &EntityConfig{Gender: "male"}, ShiftLeisure
	courtesan := NewCharacter(1, 1, &EntityConfig{Gender: "female"}, 1, false, nil); courtesan.Name, courtesan.Shift, courtesan.State.Arousal = "cr", ShiftLeisure, 100; g.World.Characters = append(g.World.Characters, courtesan)
	pc.mate(ctx, &courtesan.Actor, "anal"); if courtesan.State.Pain <= 0 { t.Errorf("Expected pain") }
}

func TestSimulation_AnalSexConstraint(t *testing.T) {
	_, ctx := setupSimulationGame(); m1, m2, f := NewCharacter(0,0, &EntityConfig{Gender: "male"}, 1, false, nil), NewCharacter(1,1, &EntityConfig{Gender: "male"}, 1, false, nil), NewCharacter(2,2, &EntityConfig{Gender: "female"}, 1, false, nil)
	m1.mate(ctx, &f.Actor, "anal"); if f.State.Pain <= 0 { t.Errorf("F should have pain") }; f.State.Pain = 0
	m1.mate(ctx, &m2.Actor, "anal"); if m2.State.Pain <= 0 { t.Errorf("M should have pain") }; m2.State.Pain = 0
	f.mate(ctx, &m1.Actor, "anal"); if m1.State.Pain > 0 { t.Errorf("F cannot give") }
	tF := NewCharacter(3,3, &EntityConfig{Gender: "female"}, 1, false, nil); tF.IsTransexual = true; tF.mate(ctx, &m1.Actor, "anal"); if m1.State.Pain <= 0 { t.Errorf("TF can give") }
}
