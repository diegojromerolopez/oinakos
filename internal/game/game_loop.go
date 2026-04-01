package game

import (
	"fmt"
	"os"
	"sync/atomic"
	"oinakos/internal/engine"
)

func (g *Game) Update() error {
	if atomic.LoadInt32(&g.LoadingProgress) < 1000 { return nil }
	if g.input.IsKeyJustPressed(g.settings.GetKey("debug")) { g.showBoundaries = !g.showBoundaries; g.debug = g.showBoundaries; SetDebugMode(g.debug) }
	if g.input.IsKeyJustPressed(g.settings.GetKey("inventory")) && !g.isMainMenu && !g.isCharacterSelect && !g.isGameOver && !g.isMapWon && !g.isGameWon {
		g.isInventoryOpen = !g.isInventoryOpen; if g.isInventoryOpen { g.isMenuOpen = false }
	}
	if g.isMainMenu || g.isCharacterSelect || g.isCampaignSelect || g.isQuitConfirmationOpen || g.isAboutScreen || g.isSettingsScreen || g.isGameOver || g.isMapWon || g.isGameWon || g.isMenuOpen {
		return g.menuHandler.Update()
	}
	g.handleInteractionInput(); g.Tick++
	if g.worldManager != nil { g.worldManager.UpdateDayCycle() }
	g.updateAI()
	if g.input.IsKeyJustPressed(engine.KeyE) { g.tryUnloading() }
	if g.input.IsKeyJustPressed(engine.KeyQ) { g.performQuicksave() }
	if g.World != nil {
		g.World.Obstacles, g.World.Characters, g.World.Projectiles, g.World.PlayableCharacter, g.World.CurrentMapType, g.World.PlayTime = g.obstacles, g.characters, g.projectiles, g.playableCharacter, &g.currentMapType, g.playTime
	}
	ctx := g.GetContext()
	if !g.simulationMode {
		g.mechanicsManager.UpdateFogOfWar(ctx)
		g.updateWeatherVFX()
	}
	g.worldManager.UpdateChunks(); g.worldManager.UpdateNPCSpawning(); g.updateWorldState()
	activeProjectiles := []*Projectile{}
	for _, p := range g.productProjectiles() { p.Update(ctx); if p.Alive { activeProjectiles = append(activeProjectiles, p) } }
	g.projectiles = activeProjectiles
	if !g.isPaused && !g.isGameOver { g.playTime += 1.0 / 60.0; if g.saveMessageTimer > 0 { g.saveMessageTimer-- } }
	g.playableCharacter.CurrentTile = g.currentMapType.GetTileAt(g.playableCharacter.X, g.playableCharacter.Y)
	// Skip file IO in simulation mode
	if !g.simulationMode && g.playableCharacter.Tick%30 == 0 { os.WriteFile("/tmp/oinakos_pos.txt", []byte(fmt.Sprintf("%.2f,%.2f", g.playableCharacter.X, g.playableCharacter.Y)), 0644) }
	g.mechanicsManager.UpdateProximityEffects(ctx)
	if g.mechanicsManager.CheckWinConditions(ctx) { g.isMapWon = true; return nil }
	if !g.playableCharacter.IsAlive() { g.isGameOver, g.deathReason = true, g.playableCharacter.GetDeathReason() }
	aliveObstacles := make([]*Obstacle, 0, len(g.obstacles))
	for _, o := range g.obstacles { o.Update(); if o.Alive { aliveObstacles = append(aliveObstacles, o) } }; g.obstacles = aliveObstacles
	for _, n := range g.characters { n.CurrentTile = g.currentMapType.GetTileAt(n.X, n.Y); n.Update(ctx); if n.MustSurvive && !n.IsAlive() { g.isGameOver = true } }
	if len(g.World.Characters) > len(g.characters) { g.characters = g.World.Characters }
	g.UpdatePickups(); g.updateItemExpiration(ctx)
	activeTexts := make([]*FloatingText, 0)
	if !g.simulationMode {
		for _, ft := range g.floatingTexts { if ft.Update() { activeTexts = append(activeTexts, ft) } }; g.floatingTexts = activeTexts
		pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y); g.camera.Follow(pIsoX, pIsoY, 0.1); g.updateOcclusion(); g.ensurePlayerNotStuck()
	}
	if g.isInventoryOpen { g.menuHandler.updateInventoryScreen() } else if g.isTradeOpen { g.menuHandler.updateTradeScreen() } else if g.ActiveDialogue != nil { g.menuHandler.updateDialogueScreen() }
	return nil
}

func (g *Game) productProjectiles() []*Projectile { return g.projectiles }
