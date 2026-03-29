package game

import "testing"

func TestSimulation_GaiferosRests(t *testing.T) {
	_, ctx := setupSimulationGame(); pc := ctx.World.PlayableCharacter; pc.State.Fatigue, pc.ActionState = 100, ActorResting
	for i := 0; i < 300; i++ { pc.updateNeeds(ctx) }; if pc.State.Fatigue >= 100 { t.Errorf("Expected fatigue to decrease") }
}

func TestSimulation_OinakosDrinks(t *testing.T) {
	_, ctx := setupSimulationGame(); pc := ctx.World.PlayableCharacter; pc.State.Thirst = 100; it := NewItemInstance("w", &ObjectConfig{Thirst: 50}, pc.X, pc.Y); pc.ConsumeItem(it, ctx)
	if pc.State.Thirst >= 100 { t.Errorf("Expected thirst to decrease") }
}

func TestSimulation_OinakosRelief(t *testing.T) {
	_, ctx := setupSimulationGame(); pc := ctx.World.PlayableCharacter; pc.State.BowelLevel, pc.State.BladderLevel, pc.ActionState = 100, 100, ActorRelieving
	for i := 0; i < 350; i++ { pc.updateNeeds(ctx); if pc.State.BowelLevel <= 0 && pc.State.BladderLevel <= 0 { break } }
	if pc.State.BowelLevel > 0 || pc.State.BladderLevel > 0 { t.Errorf("Expected relief") }
}

func TestSimulation_ElviraCleans(t *testing.T) {
	_, ctx := setupSimulationGame(); pc := ctx.World.PlayableCharacter; pc.State.Hygiene, pc.ActionState = 0, ActorBathing
	for i := 0; i < 300; i++ { pc.updateNeeds(ctx) }; if pc.State.Hygiene <= 0 { t.Errorf("Expected hygiene to increase") }
}

func TestSimulation_BiologicalReliefLoop(t *testing.T) {
	_, ctx := setupSimulationGame(); pc := ctx.World.PlayableCharacter; pc.Name = "oinakos"
	for d := 1; d <= 2; d++ {
		w := &ObjectConfig{Thirst: 50}; w_it := NewItemInstance("w", w, pc.X, pc.Y); for i := 0; i < 20; i++ { pc.ConsumeItem(w_it, ctx) }
		if pc.State.BladderLevel < 80 { t.Errorf("Day %d: Expected high bladder", d) }
		for i := 0; i < 901; i++ { pc.SharedUpdate(ctx); pc.Tick++ }
		if pc.State.Pain < 3.0 { t.Errorf("Day %d: Expected pain", d) }
		f := &ObjectConfig{Hunger: 50}; f_it := NewItemInstance("f", f, pc.X, pc.Y); for i := 0; i < 20; i++ { pc.ConsumeItem(f_it, ctx) }
		if pc.State.BowelLevel < 80 { t.Errorf("Day %d: Expected high bowel", d) }
		pc.ActionState = ActorRelieving; for i := 0; i < 100; i++ { pc.SharedUpdate(ctx); pc.Tick++ }
		if pc.State.BladderLevel > 2.0 || pc.State.BowelLevel > 2.0 { t.Errorf("Day %d: Expected relief", d) }
		if pc.State.Pain > 0 { t.Errorf("Day %d: Expected pain 0", d) }
		ctx.World.State.Ticks += 86400
	}
}

func TestSimulation_SoilYourselfUnbearable(t *testing.T) {
	_, ctx := setupSimulationGame(); pc := ctx.World.PlayableCharacter; pc.State.Hygiene = 100
	w := &ObjectConfig{Thirst: 50}; w_it := NewItemInstance("w", w, pc.X, pc.Y); for i := 0; i < 25; i++ { pc.ConsumeItem(w_it, ctx) }
	pc.Tick, pc.State.BladderLevel = 600, 100; pc.SharedUpdate(ctx)
	if pc.State.BladderLevel > 0.1 { t.Errorf("Expected bladder reset") }
	if pc.State.Hygiene > 70 { t.Errorf("Expected hygiene drop") }
	if pc.State.Pain > 0 { t.Errorf("Expected pain reset") }
}
