//go:build headless

package main

import (
	"log"
	"time"

	"oinakos/internal/engine"
	"oinakos/internal/game"
)

func main() {
	finalAssets, debug, initialMap, initialMapType, heroID := setupCommon()
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
	// Since we use MockGraphics, this will just "load" empty images.
	gr := game.NewGameRenderer(g, finalAssets, graphics)
	go func() {
		gr.LoadAssets(finalAssets)
		log.Printf("[SIM] Asset loading complete. Starting simulation loop.")
	}()

	// Wait until ready (mocked asset loading should be fast)
	// We'll give it a moment or check progress if needed.
	
	log.Printf("Simulation starting at 60 TPS...")
	ticker := time.NewTicker(16 * time.Millisecond) // Roughly 60 TPS
	defer ticker.Stop()

	// Initial progress report
	lastReport := time.Now()
	startTick := g.Tick

	for {
		select {
		case <-ticker.C:
			// Check if loading is complete
			if g.Tick == 0 && g.LoadingProgress < 1000 {
				if time.Since(lastReport) > 2*time.Second {
					log.Printf("Waiting for assets to load (Progress: %d/1000)...", g.LoadingProgress)
					lastReport = time.Now()
				}
				// We don't call Update yet, but we allow Tick to increment? No, Update increments Tick.
				// But Update returns early if LoadingProgress < 1000.
			}
			
			err := g.Update()
			if err != nil {
				log.Fatalf("Simulation error: %v", err)
			}

			if g.Tick % 600 == 0 && g.Tick > 0 && time.Since(lastReport) > 1*time.Second {
				now := time.Now()
				elapsed := now.Sub(lastReport)
				tps := float64(g.Tick-startTick) / elapsed.Seconds()
				
				log.Printf("[SIM] Tick: %d | Time: %.2fs | Player: %.1f, %.1f | TPS: %.1f", 
					g.Tick, g.World.PlayTime, g.World.PlayableCharacter.X, g.World.PlayableCharacter.Y, tps)
				
				lastReport = now
				startTick = g.Tick
			}

			// Condition to exit if needed, though for now it's infinite per request
			if !g.World.PlayableCharacter.IsAlive() {
				log.Printf("Simulation ended: Playable character died.")
				return
			}
		}
	}
}
