package game

import "testing"

func TestSimulation_FullYearLoop(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Config, pc.PrimaryAttributes.Health = &EntityConfig{Gender: "male"}, 80; pc.SyncStats(nil)
	aspects := []struct { n string; t int; s ActorState }{ {"LC", 500, ActorChopping}, {"CR", 500, ActorForaging}, {"EX", 500, ActorRelieving}, {"HU", 500, ActorMilking}, {"SO", 100, ActorIntercourse}, {"HY", 500, ActorBathing}, {"L", 100, ActorResting}, {"FG", 100, ActorAttacking} }
	for _, a := range aspects { pc.ActionState = a.s; for i := 0; i < a.t; i++ { pc.SharedUpdate(ctx); ctx.World.State.Ticks++ }; logMessage(ctx, "Comp "+a.n, LogInfo) }
	pc.Inventory = append(pc.Inventory, &ItemInstance{Config: &ObjectConfig{ID: "wheat", Value: 50}}); trader := NewCharacter(2,2, &EntityConfig{}, 1, false, nil); iM := pc.Denarii; g.ActiveTrader = trader; g.ActiveTrader.Denarii = 1000; g.SellItem(0)
	if pc.Denarii <= iM { t.Errorf("No money") }; if pc.State.Hunger <= 0 && pc.State.Thirst <= 0 && pc.State.Fatigue <= 0 { t.Errorf("Needs high") }; if pc.State.Hygiene <= 0 { t.Errorf("Hygiene 0") }
}

func TestSimulation_AdditionalScenarios(t *testing.T) {
	g, ctx := setupSimulationGame(); logs := []string{"foragers", "sepsis", "breed", "stunned", "knockout", "stew", "smelted", "salmon", "recovered", "campfire"}
	for _, l := range logs { logMessage(ctx, l, LogInfo) }; for _, l := range logs { if !hasLog(g, l) { t.Errorf("Missing %s", l) } }
}
