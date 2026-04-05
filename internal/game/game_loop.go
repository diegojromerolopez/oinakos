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
	isSuperFast := g.simulationMode && IsFastMode()
	
	cycleFreq := 1; if isSuperFast { cycleFreq = 60 } // Throttle world systems
	if g.worldManager != nil && g.Tick % cycleFreq == 0 { g.worldManager.UpdateDayCycle() }
	
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
	if g.Tick % cycleFreq == 0 {
		g.worldManager.UpdateChunks(); g.worldManager.UpdateNPCSpawning(); g.updateWorldState()
	}
	activeProjectiles := []*Projectile{}
	for _, p := range g.productProjectiles() { p.Update(ctx); if p.Alive { activeProjectiles = append(activeProjectiles, p) } }
	g.projectiles = activeProjectiles
	if !g.isPaused && !g.isGameOver { g.playTime += 1.0 / 60.0; if g.saveMessageTimer > 0 { g.saveMessageTimer-- } }
	
	skipFactor := 1; if isSuperFast { skipFactor = 10 }
	
	if g.simulationMode {
		if g.Tick % skipFactor == 0 {
			g.playableCharacter.CurrentTile = g.currentMapType.GetTileAt(g.playableCharacter.X, g.playableCharacter.Y)
			oldX, oldY := g.playableCharacter.X, g.playableCharacter.Y
			g.playableCharacter.Update(ctx)
			g.playableCharacter.X = oldX + (g.playableCharacter.X - oldX) * float64(skipFactor)
			g.playableCharacter.Y = oldY + (g.playableCharacter.Y - oldY) * float64(skipFactor)
		}
	} else {
		g.playableCharacter.CurrentTile = g.currentMapType.GetTileAt(g.playableCharacter.X, g.playableCharacter.Y)
	}

	// Skip file IO in simulation mode
	if !g.simulationMode && g.playableCharacter.Tick%30 == 0 { os.WriteFile("/tmp/oinakos_pos.txt", []byte(fmt.Sprintf("%.2f,%.2f", g.playableCharacter.X, g.playableCharacter.Y)), 0644) }
	
	if g.Tick % cycleFreq == 0 { g.mechanicsManager.UpdateProximityEffects(ctx) }
	if g.mechanicsManager.CheckWinConditions(ctx) { g.isMapWon = true; return nil }
	if !g.playableCharacter.IsAlive() { g.isGameOver, g.deathReason = true, g.playableCharacter.GetDeathReason() }
	
	// Memory Optimization: Periodic entity scrubbing
	if g.Tick%1000 == 0 {
		aliveObstacles := make([]*Obstacle, 0, len(g.obstacles))
		for _, o := range g.obstacles { if o.Alive { aliveObstacles = append(aliveObstacles, o) } }; g.obstacles = aliveObstacles
		
		aliveChars := make([]*Character, 0, len(g.characters))
		for _, c := range g.characters {
			// keep the player (Game Over handles it), alive entities, and recent corpses. Cull after 3 months (TicksPerMonth * 3).
			if c == g.playableCharacter || c.IsAlive() || c.Actor.RotTicks <= TicksPerMonth * 3 {
				aliveChars = append(aliveChars, c)
			}
		}
		g.characters = aliveChars

		if g.World != nil {
			g.World.FloatingTexts = g.floatingTexts
			g.World.Projectiles = g.projectiles
			g.World.Obstacles = g.obstacles
			g.World.Characters = g.characters
		}
	}

	// PARALLEL CHARACTER DISPATCH (Signal persistent workers only when needed)
	if g.Tick%skipFactor == 0 {
		for i := 0; i < 8; i++ { g.charStartChan <- ctx }
		for i := 0; i < 8; i++ { <-g.charDoneChan }
	}

	if len(g.World.Characters) > len(g.characters) { g.characters = g.World.Characters }
	g.UpdatePickups(); 
	if g.Tick % cycleFreq == 0 { g.updateItemExpiration(ctx) } 
	
	if !g.simulationMode {
		activeTexts := make([]*FloatingText, 0)
		for _, ft := range g.floatingTexts { if ft.Update() { activeTexts = append(activeTexts, ft) } }; g.floatingTexts = activeTexts
		pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y); g.camera.Follow(pIsoX, pIsoY, 0.1); g.updateOcclusion(); g.ensurePlayerNotStuck()
	}
	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		mx, my := g.input.MousePosition()
		if mx >= 10 && mx <= 35 && my >= 70 && my <= 95 { g.ToggleHUD() }
	}
	if g.isInventoryOpen { g.menuHandler.updateInventoryScreen() } else if g.isTradeOpen { g.menuHandler.updateTradeScreen() } else if g.ActiveDialogue != nil { g.menuHandler.updateDialogueScreen() }
	return nil
}

func (g *Game) processCharacter(ctx *SystemContext, n *Character) {
	isSuperFast := g.simulationMode && IsFastMode()
	skipFactor := 1; if isSuperFast { skipFactor = 10 }

	if g.simulationMode {
		if g.Tick%skipFactor != 0 { return }
		n.CurrentTile = g.currentMapType.GetTileAt(n.X, n.Y)
		oldX, oldY := n.X, n.Y
		n.Update(ctx)
		n.X = oldX + (n.X-oldX)*float64(skipFactor)
		n.Y = oldY + (n.Y-oldY)*float64(skipFactor)
	} else {
		n.CurrentTile = g.currentMapType.GetTileAt(n.X, n.Y)
		n.Update(ctx)
	}
	if n.MustSurvive && !n.IsAlive() { g.isGameOver = true }
}

func (g *Game) productProjectiles() []*Projectile { return g.projectiles }
