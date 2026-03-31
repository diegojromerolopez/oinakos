//go:build headless

package main

import (
	"log"
	"time"

	"oinakos/internal/engine"
	"oinakos/internal/game"
)

func main() {
	finalAssets, debug, initialMap, initialMapType, heroID, fastSim, simDuration := setupCommon()
	log.Printf("Starting Headless Oinakos Simulation Mode (No Ebiten)...")

	loadRegistries(finalAssets)

	// Use Mock providers for headless mode
	graphics := &engine.MockGraphics{}
	input := engine.NewMockInput()

	g := game.NewGame(finalAssets, graphics, initialMap, initialMapType, heroID, input, &game.DefaultAudioManager{}, debug, Version)
	g.SetSimulationMode(true)
	g.BypassMenu()

	// In headless mode, we still need to run the asset loading process 
	// because the game loop waits for LoadingProgress to reach 1000.
	gr := game.NewGameRenderer(g, finalAssets, graphics)
	go func() {
		gr.LoadAssets(finalAssets)
		log.Printf("[SIM] Asset loading complete. Starting simulation loop.")
	}()

	tickerDuration := 16 * time.Millisecond
	if fastSim {
		// Target ~8640 TPS (1 month game time = 1 minute real time)
		// Since we can't reliably tick at 115 microseconds, we'll try to run in a semi-busy loop
		tickerDuration = 100 * time.Microsecond 
		log.Printf("[SIM] Fast Simulation Mode Enabled (Target ~8640 TPS)")
	}
	
	log.Printf("Simulation starting...")
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	// Initial progress report
	lastReport := time.Now()
	startTick := g.Tick
	realStartTime := time.Now()

	for {
		select {
		case <-ticker.C:
			// Run multiple updates per tick if in fast mode to overcome system timer overhead
			updatesPerTick := 1
			if fastSim {
				updatesPerTick = 10 // Try for ~100k TPS theoretical max, limited by CPU
			}

			for i := 0; i < updatesPerTick; i++ {
				// Check if loading is complete
				if g.Tick == 0 && g.LoadingProgress < 1000 {
					if time.Since(lastReport) > 2*time.Second {
						log.Printf("Waiting for assets to load (Progress: %d/1000)...", g.LoadingProgress)
						lastReport = time.Now()
					}
					break
				}
				
				err := g.Update()
				if err != nil {
					log.Fatalf("Simulation error: %v", err)
				}

				if !g.World.PlayableCharacter.IsAlive() {
					log.Printf("Simulation ended: Playable character died.")
					log.Printf("Death reason: %s", g.World.PlayableCharacter.GetDeathReason())
					log.Printf("Final State: HP=%.1f/%.1f | Hunger=%.2f | Thirst=%.2f | Fatigue=%.2f", 
						g.World.PlayableCharacter.State.HealthPoints, 
						g.World.PlayableCharacter.State.MaxHealthPoints,
						g.World.PlayableCharacter.State.Hunger,
						g.World.PlayableCharacter.State.Thirst,
						g.World.PlayableCharacter.State.Fatigue)
					return
				}

				// Handle duration limit
				if simDuration > 0 && time.Since(realStartTime).Minutes() >= simDuration {
					log.Printf("Simulation ended: Duration of %.1f minutes reached.", simDuration)
					return
				}
			}

			if g.Tick > 0 && time.Since(lastReport) > 1*time.Second {
				now := time.Now()
				elapsed := now.Sub(lastReport)
				tps := float64(g.Tick-startTick) / elapsed.Seconds()
				
				log.Printf("[SIM] Tick: %d | Time: %.2fs | Player: %.1f, %.1f | TPS: %.1f", 
					g.Tick, g.World.PlayTime, g.World.PlayableCharacter.X, g.World.PlayableCharacter.Y, tps)
				
				lastReport = now
				startTick = g.Tick
			}
		}
	}
}
