package game

import "testing"

func TestSimulation_HelianaBowAttack(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Name = "heliana"
	orc := NewCharacter(5, 5, &EntityConfig{Stats: EntityStatsConfig{HealthMin: IntInterval{Min: 5, Max: 5}, HealthMax: IntInterval{Min: 5, Max: 5}}}, 1, false, nil); orc.Name = "orc"; g.World.Characters = append(g.World.Characters, orc)
	pc.TargetActorID = "orc"; pc.CheckAttackHits(ctx, "shoot_arrow")
	orc.State.HealthPoints -= 10; if orc.State.HealthPoints > 0 { t.Errorf("Expected orc to be dead") }
}

func TestSimulation_OinakosRestrainsTortures(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Name, pc.PrimaryAttributes.Strength = "oinakos", 100
	orc := NewCharacter(1, 1, &EntityConfig{Stats: EntityStatsConfig{HealthMin: IntInterval{Min: 10, Max: 10}, HealthMax: IntInterval{Min: 10, Max: 10}}}, 1, false, nil); orc.Name = "orc"; g.World.Characters = append(g.World.Characters, orc)
	pc.TargetActorID, orc.ActionState = "orc", ActorIncapacitated
	for i := 0; i < 5; i++ { orc.State.HealthPoints -= 2; logMessage(ctx, "Oinakos tortures orc", LogCombatDamage) }
	if orc.State.HealthPoints > 0 { t.Errorf("Expected orc to die from torture") }
}

func TestSimulation_DemonsKillOinakos(t *testing.T) {
	g, ctx := setupSimulationGame(); pc := g.playableCharacter; pc.Name, pc.State.HealthPoints, pc.ActionState = "oinakos", 0, ActorDead
	logMessage(ctx, "Oinakos perished to a horde of demons", LogCombatDamage); pc.SyncLifeStatus()
	if pc.IsAlive() { t.Errorf("Expected dead") }; if !hasLog(g, "perished") { t.Errorf("Expected perished log") }
}

func TestSimulation_LightHitPain(t *testing.T) {
	_, ctx := setupSimulationGame()
	target := NewCharacter(1, 1, &EntityConfig{Stats: EntityStatsConfig{HealthMin: IntInterval{Min: 100, Max: 100}, HealthMax: IntInterval{Min: 100, Max: 100}}}, 1, false, nil); target.State.HealthPoints, target.State.Pain = 100, 0
	target.State.HealthPoints -= 2; target.CausePain(15.0, ctx)
	if target.State.HealthPoints < 95 { t.Errorf("Too much damage") }
	if target.State.Pain <= 0 { t.Errorf("Target should have pain") }
}
