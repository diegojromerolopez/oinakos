package game

import (
	"context"
	"fmt"
	"image/color"
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
	if g.input.IsKeyJustPressed(g.settings.GetKey("debug")) {
		g.showBoundaries = !g.showBoundaries
		g.debug = g.showBoundaries
		SetDebugMode(g.debug)
	}

	// Toggle inventory overlay
	if g.input.IsKeyJustPressed(g.settings.GetKey("inventory")) && !g.isMainMenu && !g.isCharacterSelect && !g.isCampaignSelect && !g.isGameOver && !g.isMapWon && !g.isGameWon {
		g.isInventoryOpen = !g.isInventoryOpen
		if g.isInventoryOpen {
			g.isMenuOpen = false // Ensure pause menu is closed
		}
	}

	// Menu logic override (these PAUSE the game)
	if g.isMainMenu || g.isCharacterSelect || g.isCampaignSelect || g.isQuitConfirmationOpen || g.isAboutScreen || g.isSettingsScreen || g.isGameOver || g.isMapWon || g.isGameWon || g.isMenuOpen {
		return g.menuHandler.Update()
	}

	// Interaction and Dialogue Handling (these do NOT pause the game)
	if g.ActiveDialogue == nil {
		g.handleDialogueInput()
		g.handleDialogueProximity()
		g.handleLogScrolling()

		// --- Character Pinning Logic ---
		if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
			mx, my := g.input.MousePosition()
			offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)
			
			// 1. Check for command clicks if we have a pin
			if g.pinnedCharacter != nil {
				n := g.pinnedCharacter
				isoX, isoY := engine.CartesianToIso(n.X, n.Y)
				scrX, scrY := isoX+offsetX, isoY+offsetY
				
				boxW, boxH := 320.0, 480.0
				bx, by := float32(scrX+20), float32(scrY+20)
				if float64(bx)+boxW > float64(g.width) { bx = float32(float64(scrX) - boxW - 20) }
				if float64(by)+boxH > float64(g.height) { by = float32(float64(scrY) - boxH - 20) }
				
				yCmds := int(by) + int(boxH) - 120 + 20
				commands := []string{"TALK", "ATTACK", "TRADE", "SEDUCE", "INTIMIDATE", "STEAL", "RESTRAIN", "HEAL", "GIVE ITEM", "SEX", "TORTURE"}
				for i, cmd := range commands {
					cx, cy := int(bx)+10 + (i%3)*100, yCmds + (i/3)*25
					if mx >= cx && mx <= cx+85 && my >= cy-15 && my <= cy+10 {
						g.handlePinnedCommand(cmd, n)
						return nil 
					}
				}
			}

			// 2. Click on world to pin/unpin
			var found *Character
			for _, n := range g.characters {
				if !n.IsAlive() { continue }
				isoX, isoY := engine.CartesianToIso(n.X, n.Y)
				scrX, scrY := isoX+offsetX, isoY+offsetY
				dist := math.Sqrt(math.Pow(float64(mx)-scrX, 2) + math.Pow(float64(my)-(scrY-40), 2))
				if dist < 40 {
					found = n
					break
				}
			}

			if found != nil {
				g.pinnedCharacter = found
			} else {
				// Only clear if we didn't click on UI elements 
				// (Simple check: most UI is on top/bottom/sides)
				// Log box is at bottom, status panel is at top-left.
				// We'll just clear for now, unless it's a very specific UI area.
				isUI := my < 50 || (my > g.height-180 && mx < g.width-20) || (mx < 360 && my < 300)
				if !isUI {
					g.pinnedCharacter = nil
				}
			}
		}

		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			if g.pinnedCharacter != nil {
				g.pinnedCharacter = nil
			} else {
				g.isMenuOpen = true
			}
			return nil
		}
	}

	// Update AI
	g.Tick++
	if g.worldManager != nil { g.worldManager.UpdateDayCycle() }
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
							n.ApplyAIDecision(g.GetContext(), a.Decision)
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

	if g.input.IsKeyJustPressed(engine.KeyE) { g.tryUnloading() }
	if g.input.IsKeyJustPressed(engine.KeyQ) { g.performQuicksave() }

	if g.World != nil {
		g.World.Obstacles = g.obstacles
		g.World.Characters = g.characters
		g.World.Projectiles = g.projectiles
		g.World.PlayableCharacter = g.playableCharacter
		g.World.CurrentMapType = &g.currentMapType
		g.World.PlayTime = g.playTime
	}

	ctx := &SystemContext{
		World:      g.World,
		Input:      g.input,
		Audio:      g.audio,
		Registries: g.Registries,
		Log:        g.LogEvent,
		AIManager:  g.aiManager,
		Weather:    g.World.State.Weather,
		Intensity:  g.World.State.Intensity,
		Settings:   g.settings,
	}

	g.mechanicsManager.UpdateFogOfWar(ctx)
	g.worldManager.UpdateChunks()
	g.worldManager.UpdateNPCSpawning()
	g.updateWorldState()
	g.updateWeatherVFX()

	activeProjectiles := []*Projectile{}
	for _, p := range g.World.Projectiles {
		p.Update(ctx)
		if p.Alive { activeProjectiles = append(activeProjectiles, p) }
	}
	g.projectiles = activeProjectiles

	if !g.isPaused && !g.isGameOver {
		g.playTime += 1.0 / 60.0
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

	for _, n := range g.characters {
		n.CurrentTile = g.currentMapType.GetTileAt(n.X, n.Y)
		n.Update(ctx)
		if n.MustSurvive && !n.IsAlive() {
			if !g.isGameOver { DebugLog("CRITICAL FAILURE: [%s] was killed! Quest Failed.", n.Name) }
			g.isGameOver = true
		}
	}
	// Re-sync if birth happened
	if len(g.World.Characters) > len(g.characters) {
		g.characters = g.World.Characters
	}

	g.UpdatePickups()
	g.updateItemExpiration(ctx)

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
	if g.isTradeOpen {
		g.menuHandler.updateTradeScreen()
	}
	if g.ActiveDialogue != nil {
		g.menuHandler.updateDialogueScreen()
	}

	return nil
}

func (g *Game) updateItemExpiration(ctx *SystemContext) {
	if g.World == nil { return }
	// 1. Items on ground
	for _, it := range g.World.Items {
		if it != nil { it.Update(ctx) }
	}
	// 2. Playable Character
	if g.playableCharacter != nil {
		for _, it := range g.playableCharacter.Inventory {
			if it != nil { it.Update(ctx) }
		}
		for _, it := range g.playableCharacter.Slots {
			if it != nil { it.Update(ctx) }
		}
	}
	// 3. All NPCs
	for _, n := range g.characters {
		if n != nil {
			for _, it := range n.Inventory {
				if it != nil { it.Update(ctx) }
			}
			for _, it := range n.Slots {
				if it != nil { it.Update(ctx) }
			}
		}
	}
}

func (g *Game) updateWorldState() {
	if g.World == nil { return }
	s := &g.World.State
	s.Ticks++
	
	// 1 Hour = 720 Ticks (12 seconds IRT)
	if s.Ticks >= 720 {
		s.Ticks = 0
		s.Hour++
		if s.Hour >= 24 {
			s.Hour = 0
			s.Day++
			// 1 Month = 4 Days. 1 Year = 48 Days.
			monthVal := ((s.Day - 1) / 4) + 1
			s.Month = monthVal
			if s.Month > 12 {
				s.Month = 1
				s.Day = 1
			}
			// Season Mapping: 3 months each
			if s.Month >= 3 && s.Month <= 5 { s.Season = SeasonSpring
			} else if s.Month >= 6 && s.Month <= 8 { s.Season = SeasonSummer
			} else if s.Month >= 9 && s.Month <= 11 { s.Season = SeasonAutumn
			} else { s.Season = SeasonWinter }
		}
	}

	// Ambient Temperature Calculation
	baseTemp := 15.0 // Spring
	switch s.Season {
	case SeasonSummer: baseTemp = 28.0
	case SeasonAutumn: baseTemp = 12.0
	case SeasonWinter: baseTemp = -2.0
	}
	// Diurnal cycle: coldest at 04:00, hottest at 16:00
	hourOffset := math.Sin(float64(s.Hour-10)*math.Pi/12.0) * 8.0
	s.Temperature = baseTemp + hourOffset
	
	if s.Weather == WeatherRain { s.Temperature -= 4.0 }
	if s.Weather == WeatherSnow { s.Temperature -= 7.0 }

	// Weather Management (Refactored from updateWeather)
	if s.WeatherTimer > 0 {
		s.WeatherTimer--
	} else {
		// Determine next weather based on season
		roll := rand.Float64()
		s.WeatherTimer = 3600 + rand.Intn(7200) // 1 to 2 hours (game time)
		
		switch s.Season {
		case SeasonWinter:
			if roll < 0.4 { s.Weather = WeatherSnow
			} else if roll < 0.6 { s.Weather = WeatherClear
			} else { s.Weather = WeatherFog }
		case SeasonSummer:
			if roll < 0.1 { s.Weather = WeatherRain
			} else { s.Weather = WeatherClear }
		case SeasonAutumn:
			if roll < 0.4 { s.Weather = WeatherRain
			} else if roll < 0.7 { s.Weather = WeatherFog
			} else { s.Weather = WeatherClear }
		default: // Spring
			if roll < 0.2 { s.Weather = WeatherRain
			} else { s.Weather = WeatherClear }
		}
		s.Intensity = 0.3 + rand.Float64()*0.7
	}
}

func (g *Game) handlePinnedCommand(cmd string, n *Character) {
	pc := g.playableCharacter
	ctx := g.GetContext()
	// Proximity check for interaction (approx 3 Cartesian units)
	dist := math.Sqrt(math.Pow(pc.X-n.X, 2) + math.Pow(pc.Y-n.Y, 2))
	if dist > 3.0 && cmd != "ATTACK" {
		g.LogEvent(fmt.Sprintf("%s is too far away!", n.Name), LogInfo)
		return
	}

	switch cmd {
	case "TALK":
		g.InitiateDialogue(n)
	case "ATTACK":
		pc.TargetActor = &n.Actor
		g.LogEvent(fmt.Sprintf("Attacking %s!", n.Name), LogPlayer)
		n.Say("Wait! What are you doing?!", ctx)
	case "TRADE":
		g.ActiveTrader = n
		g.isTradeOpen = true
		g.LogEvent(fmt.Sprintf("Trading with %s...", n.Name), LogPlayer)
		n.LastReaction = "Let's see what you have."
	case "SEDUCE":
		if n.State == ActorIncapacitated || pc.CheckAbilitySuccess("seduce", 0) {
			n.ModifyRomanticInterest(pc.Name, 10.0)
			g.AddFloatingText("❤", n.X, n.Y-1, color.RGBA{255, 100, 200, 255})
			g.LogEvent(fmt.Sprintf("Successful seduction of %s!", n.Name), LogPlayer)
			n.Say("You... you are quite charming.", ctx)
		} else {
			n.ModifySentiment(pc.Name, -5.0)
			g.LogEvent(fmt.Sprintf("%s was not impressed.", n.Name), LogNPC)
			n.Say("Stop that. It's embarrassing.", ctx)
		}
	case "INTIMIDATE":
		if n.State == ActorIncapacitated || pc.CheckAbilitySuccess("intimidate", 0) {
			n.ModifySubmission(pc.Name, 15.0)
			g.LogEvent(fmt.Sprintf("%s is cowed by your presence.", n.Name), LogPlayer)
			n.Say("P-please... I'll do whatever you want.", ctx)
		} else {
			n.ModifySentiment(pc.Name, -10.0)
			g.LogEvent(fmt.Sprintf("%s stands their ground.", n.Name), LogNPC)
			n.Say("You don't scare me.", ctx)
		}
	case "STEAL":
		if n.State == ActorIncapacitated || pc.CheckAbilitySuccess("steal", 0) {
			g.LogEvent(fmt.Sprintf("You successfully pilfered from %s!", n.Name), LogPlayer)
			n.LastReaction = "(They haven't noticed yet...)"
		} else {
			n.Alignment = AlignmentEnemy
			g.LogEvent(fmt.Sprintf("Caught! %s is calling for the guard!", n.Name), LogNPC)
			n.Say("THIEF! GUARDS! HELP!", ctx)
		}
	case "RESTRAIN":
		if pc.CompetitiveContest(&n.Actor, "dexterity", "strength") {
			n.State = ActorIncapacitated
			g.LogEvent(fmt.Sprintf("You have RESTRAINED %s!", n.Name), LogPlayer)
			g.AddFloatingText("Bound!", n.X, n.Y-1, color.RGBA{200, 200, 200, 255})
			n.Say("Get these off me! Let me go!", ctx)
		} else {
			n.Alignment = AlignmentEnemy
			g.LogEvent(fmt.Sprintf("%s broke free and is attacking!", n.Name), LogNPC)
			n.Say("Nice try! Now you'll pay!", ctx)
		}
	case "HEAL":
		if pc.CheckAbilitySuccess("heal", 0) {
			n.Heal(20)
			g.LogEvent(fmt.Sprintf("You tended to %s's wounds.", n.Name), LogPlayer)
			n.Say("Thank you... the pain is fading.", ctx)
		}
	case "GIVE ITEM":
		g.LogEvent("Giving items is currently managed via the Trade screen.", LogInfo)
	case "SEX":
		if dist > 2.0 {
			g.LogEvent(fmt.Sprintf("%s is too far away for that!", n.Name), LogInfo)
			return
		}
		pc.Actor.mate(ctx, &n.Actor, "vaginal")
	case "TORTURE":
		if dist > 2.0 {
			g.LogEvent(fmt.Sprintf("%s is too far away!", n.Name), LogInfo)
			return
		}
		pc.Actor.Torture(&n.Actor, ctx)
	}
}
