package game

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"

	"oinakos/internal/engine"
)

func (g *Game) Update() error {
	if atomic.LoadInt32(&g.LoadingProgress) < 1000 {
		return nil
	}
	if g.input.IsKeyJustPressed(engine.KeyTab) {
		g.showBoundaries = !g.showBoundaries
		g.debug = g.showBoundaries
		SetDebugMode(g.debug)
	}

	// Toggle inventory overlay
	if g.input.IsKeyJustPressed(engine.KeyI) && !g.isMainMenu && !g.isCharacterSelect && !g.isCampaignSelect && !g.isGameOver && !g.isMapWon && !g.isGameWon {
		g.isInventoryOpen = !g.isInventoryOpen
		if g.isInventoryOpen {
			g.isMenuOpen = false // Ensure pause menu is closed
		}
	}

	// Menu logic override
	if g.isMainMenu || g.isCharacterSelect || g.isCampaignSelect || g.isQuitConfirmationOpen || g.isAboutScreen || g.isSettingsScreen || g.isGameOver || g.isMapWon || g.isGameWon || g.isMenuOpen {
		return g.menuHandler.Update()
	}

	// Update AI
	g.Tick++
	if g.aiManager == nil && IsDebugEnabled() && g.Tick%60 == 0 {
		DebugLog("aiManager is NIL (Provider: %s)", g.settings.AIProvider)
	}

	if g.aiManager != nil {
		applied := g.aiManager.Poll()
		for _, a := range applied {
			if a.Decision.Err == nil {
				log.Printf("[AI-VERBOSE] Decision for %s: %s (%s)", a.NPCID, a.Decision.ChosenOption, a.Decision.Reasoning)
				if g.debug {
					g.LogEvent(fmt.Sprintf("AI [%s]: %s (%s)", a.NPCID, a.Decision.ChosenOption, a.Decision.Reasoning), LogInfo)
				}
				if a.NPCID == "PLAYER" {
					g.applyPlayerAIDecision(a.Decision)
					g.playableCharacter.LastAIDecisionTick = g.Tick
				} else {
					for _, n := range g.characters {
						if n.Name == a.NPCID || (n.Config != nil && n.Config.ID == a.NPCID) {
							choice := strings.ToLower(a.Decision.ChosenOption)
							if strings.Contains(choice, "talk") || strings.Contains(choice, "say") || strings.Contains(choice, "mutter") {
								msg := a.Decision.Reasoning
								if msg != "" {
									g.LogEvent(fmt.Sprintf("%s: %s", n.Name, msg), LogNPC)
								}
							}
							n.ApplyAIDecision(a.Decision)
							n.LastAIDecisionTick = g.Tick
							break
						}
					}
				}
			} else {
				DebugLog("AI Decision Error for %s: %v", a.NPCID, a.Decision.Err)
				if a.NPCID == "PLAYER" {
					g.playableCharacter.AIDecisionPending = false
					if g.playableCharacter.TargetActor == nil && g.playableCharacter.WanderDirX == 0 && g.playableCharacter.WanderDirY == 0 {
						g.playableCharacter.WanderDirX = rand.Float64()*2 - 1
						g.playableCharacter.WanderDirY = rand.Float64()*2 - 1
					}
				} else {
					for _, n := range g.characters {
						if n.Config.ID == a.NPCID || n.Name == a.NPCID {
							n.AIDecisionPending = false
							break
						}
					}
				}
			}
		}

		interval := 300 
		if IsDebugEnabled() { interval = 60 }

		if g.Tick%120 == 0 {
			DebugLog("[AI-DIAG] Tick: %d, SimMode: %v, Pending: %v, Manager: %v", g.Tick, g.settings.AISimulationMode, g.playableCharacter.AIDecisionPending, g.aiManager != nil)
		}

		if g.settings.AISimulationMode && !g.playableCharacter.AIDecisionPending && (g.Tick-g.playableCharacter.LastAIDecisionTick) >= interval {
			worldCtx := BuildWorldContext(g, nil)
			options := []string{"wander", "attack_nearest", "defend", "flee", "move_to_objective"}
			DebugLog("Simulation Mode: Requesting AI decision for PLAYER with options: %v", options)
			g.aiManager.RequestDecision(context.Background(), "PLAYER", worldCtx, options)
			g.playableCharacter.AIDecisionPending = true
			g.playableCharacter.LastAIDecisionTick = g.Tick
		}

		prob := g.settings.GetTalkingProbability()
		if prob > 0 && g.Tick%600 == 0 {
			for _, n := range g.characters {
				if !n.IsAlive() || n.AIDecisionPending { continue }
				dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
				if dist < 12.0 && rand.Float64() < prob {
					if g.aiManager != nil && g.settings.AIProvider != "none" {
						worldCtx := BuildWorldContext(g, n)
						options := []string{"wander", "talk_to_player", "mutter_to_self"}
						g.aiManager.RequestDecision(context.Background(), n.Name, worldCtx, options)
						n.AIDecisionPending = true
						n.LastAIDecisionTick = g.Tick
					} else {
						if n.Config != nil && n.Config.Dialogues != nil {
							bark := n.Config.Dialogues.PickIdleBark()
							if bark != "" {
								g.LogEvent(fmt.Sprintf("%s: %s", n.Name, bark), LogNPC)
							}
						}
					}
					break
				}
			}
		}
	}

	if g.ActiveDialogue != nil {
		if g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyUp) {
			g.ActiveDialogue.SelectedChoice--
			if g.ActiveDialogue.SelectedChoice < 0 {
				g.ActiveDialogue.SelectedChoice = len(g.ActiveDialogue.Choices) - 1
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyDown) {
			g.ActiveDialogue.SelectedChoice++
			if g.ActiveDialogue.SelectedChoice >= len(g.ActiveDialogue.Choices) {
				g.ActiveDialogue.SelectedChoice = 0
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyEnter) { g.AdvanceDialogue() }
		if g.input.IsKeyJustPressed(engine.KeyEscape) || g.input.IsKeyJustPressed(engine.KeyBackspace) {
			g.CloseDialogue()
			return nil
		}
	} else {
		g.handleDialogueInput()
		g.handleDialogueProximity()
		g.handleLogScrolling()
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			g.isMenuOpen = true
			return nil
		}
	}

	if g.input.IsKeyJustPressed(engine.KeyE) { g.tryUnloading() }
	if g.input.IsKeyJustPressed(engine.KeyQ) { g.performQuicksave() }

	ctx := &SystemContext{
		World:      g.World,
		Input:      g.input,
		Audio:      g.audio,
		Registries: g.Registries,
		Log:        g.LogEvent,
		AIManager:  g.aiManager,
		Weather:    g.CurrentWeather,
		Intensity:  g.WeatherIntensity,
	}

	g.mechanicsManager.UpdateFogOfWar(ctx)
	g.worldManager.UpdateChunks()
	g.worldManager.UpdateNPCSpawning()
	g.updateWeather()

	activeProjectiles := []*Projectile{}
	for _, p := range g.World.Projectiles {
		p.Update(ctx)
		if p.Alive { activeProjectiles = append(activeProjectiles, p) }
	}
	g.World.Projectiles = activeProjectiles
	g.projectiles = g.World.Projectiles

	if !g.isPaused && !g.isGameOver {
		g.playTime += 1.0 / 60.0
		g.World.PlayTime = g.playTime
		if g.saveMessageTimer > 0 { g.saveMessageTimer-- }
	}

	g.World.CurrentMapType = &g.currentMapType
	g.World.PlayableCharacter.CurrentTile = g.currentMapType.GetTileAt(g.World.PlayableCharacter.X, g.World.PlayableCharacter.Y)
	g.World.PlayableCharacter.Update(ctx)

	if g.playableCharacter.Tick%30 == 0 { g.logRealtimePosition() }
	os.WriteFile("/tmp/oinakos_pos.txt", []byte(fmt.Sprintf("%.2f,%.2f", g.playableCharacter.X, g.playableCharacter.Y)), 0644)

	g.mechanicsManager.UpdateProximityEffects(ctx)

	if g.mechanicsManager.CheckWinConditions(ctx) {
		if !g.isMapWon {
			DebugLog("Objective Completed! Level %d cleared. Objective: %v", g.mapLevel, g.currentMapType.Type)
		}
		g.isMapWon = true
		return nil
	}

	if !g.playableCharacter.IsAlive() { g.isGameOver = true }

	aliveObstacles := make([]*Obstacle, 0, len(g.obstacles))
	for _, o := range g.obstacles {
		o.Update()
		if o.Alive { aliveObstacles = append(aliveObstacles, o) }
	}
	g.obstacles = aliveObstacles

	ctx.World.Characters = g.characters
	for _, n := range ctx.World.Characters {
		n.CurrentTile = g.currentMapType.GetTileAt(n.X, n.Y)
		n.Update(ctx)
		if n.MustSurvive && !n.IsAlive() {
			if !g.isGameOver { DebugLog("CRITICAL FAILURE: [%s] was killed! Quest Failed.", n.Name) }
			g.isGameOver = true
		}
	}

	g.UpdatePickups()

	activeTexts := make([]*FloatingText, 0)
	for _, ft := range g.floatingTexts {
		if ft.Update() { activeTexts = append(activeTexts, ft) }
	}
	g.floatingTexts = activeTexts

	pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
	g.camera.Follow(pIsoX, pIsoY, 0.1)

	g.updateOcclusion()
	g.ensurePlayerNotStuck()

	if g.isInventoryOpen {
		g.menuHandler.updateInventoryScreen()
	}

	return nil
}
