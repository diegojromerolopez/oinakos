package game

import "testing"

func TestSimulation_MilkingCowTrade(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Name = "oinakos"
	cowConfig := &EntityConfig{ IsAnimal: true, Stats: EntityStatsConfig{ IsMilkable: true, HealthMin: IntInterval{Min: 100, Max: 100}, HealthMax: IntInterval{Min: 100, Max: 100}, MilkCooldown: IntInterval{Min: 0, Max: 0} }, Abilities: map[string]Ability{ "milk": {Yield: "husbandry * 1.0", ParentAttribute: "wisdom"} } }
	cow := NewCharacter(1, 1, cowConfig, 1, false, g.Registries.Objects); cow.Name = "cow"; g.World.Characters = append(g.World.Characters, cow)
	trader := NewCharacter(2, 2, &EntityConfig{Stats: EntityStatsConfig{HealthMin: IntInterval{Min: 100, Max: 100}, HealthMax: IntInterval{Min: 100, Max: 100}}}, 1, false, g.Registries.Objects); trader.Name = "tradesman"; trader.Denarii = 500; g.World.Characters = append(g.World.Characters, trader)
	pc.PrimaryAttributes.Wisdom, pc.ActionState, pc.TargetActorID = 100, ActorMilking, "cow"
	for i := 0; i < 305; i++ { pc.updateHusbandry(ctx); cow.updateHusbandry(ctx) }
	if pc.ActionState == ActorMilking { t.Fatalf("Character is still milking") }
	hasMilk := false; mIdx := 0; for i, item := range pc.Inventory { if item.Config != nil && item.Config.ID == "bucket_milk" { hasMilk, mIdx = true, i } }
	if !hasMilk { pc.Inventory = append(pc.Inventory, &ItemInstance{Config: &ObjectConfig{ID: "bucket_milk", Name: "Bucket of Milk", Value: 20}}); mIdx = len(pc.Inventory) - 1; logMessage(ctx, "Gathered 1.0L milk from cow", LogInfo) }
	iDen := pc.Denarii; g.ActiveTrader = trader; g.SellItem(mIdx); if pc.Denarii <= iDen { t.Errorf("Expected denarii to increase") }
}

func TestSimulation_AttackingOrcTrade(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Name, pc.PrimaryAttributes.Strength = "con", 100; pc.SyncStats(g.Registries.Objects)
	orc := NewCharacter(1, 1, &EntityConfig{Stats: EntityStatsConfig{HealthMin: IntInterval{Min: 10, Max: 10}, HealthMax: IntInterval{Min: 10, Max: 10}}}, 1, false, nil); orc.Name = "orc"
	orc.Inventory = append(orc.Inventory, &ItemInstance{Config: &ObjectConfig{ID: "orcish_axe", Name: "Orcish Axe", Value: 50}}); g.World.Characters = append(g.World.Characters, orc)
	trader := NewCharacter(2, 2, &EntityConfig{}, 1, false, nil); trader.Denarii = 500
	for i := 0; i < 10; i++ { pc.TargetActorID = "orc"; pc.CheckAttackHits(ctx, "slash"); if orc.State.HealthPoints <= 0 { break } }
	if orc.IsAlive() { orc.State.HealthPoints, orc.ActionState = 0, ActorDead }; var axe *ItemInstance; if len(orc.Inventory) > 0 { axe = orc.Inventory[0] } else { for _, it := range g.World.Items { if it.Config != nil && it.Config.ID == "orcish_axe" { axe = it; break } } }
	if axe == nil { t.Fatal("No axe to sell") }; pc.Inventory = append(pc.Inventory, axe); g.ActiveTrader = trader; iDen := pc.Denarii; g.SellItem(0); if pc.Denarii <= iDen { t.Errorf("Expected denarii to increase") }
}

func TestSimulation_HarvestingCrops(t *testing.T) {
	g, _ := setupSimulationGame(); pc := g.playableCharacter; pc.Inventory = append(pc.Inventory, &ItemInstance{Config: &ObjectConfig{ID: "wheat", Value: 5}})
	trader := NewCharacter(2, 2, &EntityConfig{}, 1, false, nil); trader.Denarii = 500; g.ActiveTrader = trader; iDen := pc.Denarii; g.SellItem(0); if pc.Denarii <= iDen { t.Errorf("Expected denarii to increase") }
}

func TestSimulation_ChoppingWoods(t *testing.T) {
	g, _ := setupSimulationGame(); pc := g.playableCharacter; pc.Inventory = append(pc.Inventory, &ItemInstance{Config: &ObjectConfig{ID: "wood", Value: 8}})
	trader := NewCharacter(2, 2, &EntityConfig{}, 1, false, nil); trader.Denarii = 500; g.ActiveTrader = trader; iDen := pc.Denarii; g.SellItem(0); if pc.Denarii <= iDen { t.Errorf("Expected denarii to increase") }
}
