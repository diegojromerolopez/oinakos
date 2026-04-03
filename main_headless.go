//go:build headless

package main

import (
	"flag"
	"io/fs"
	"log"
	"oinakos/internal/engine"
	"oinakos/internal/game"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	initialMap     string
	initialMapType string
	heroID         string
	debug          bool
	fastMode       bool
	durationSec    int
)

const HeadlessVersion = "1.0.0-headless"

func main() {
	flag.StringVar(&initialMap, "map", "data/maps/venburgo.yaml", "Initial map to load")
	flag.StringVar(&initialMapType, "map-type", "", "Initial map type ID")
	flag.StringVar(&heroID, "hero", "oinakos", "Initial hero ID")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&fastMode, "fast", false, "Enable hyper-fast simulation")
	flag.IntVar(&durationSec, "duration", 0, "Simulation duration in seconds (0 = infinite)")
	flag.Parse()

	log.Printf("Starting Oinakos Headless Simulation (Hyper-Fast: %v)...", fastMode)

	// Use the embedded assets if possible, or local DirFS
	var finalAssets fs.FS = assets 
	if _, err := os.Stat("oinakos"); err == nil {
		finalAssets = &combinedFS{local: os.DirFS("oinakos"), embed: assets}
	} else if _, err := os.Stat("assets"); err == nil {
		finalAssets = &combinedFS{local: os.DirFS("."), embed: assets}
	}

	graphics := &engine.MockGraphics{}
	input := engine.NewMockInput()
	
	// Set testing mode to avoid some Ebiten UI dependencies if any
	game.SetTestingMode(true)

	// Providers
	// Handle map path vs ID
	mapID := initialMap
	if strings.Contains(initialMap, "/") || strings.Contains(initialMap, ".yaml") {
		mapID = strings.TrimSuffix(filepath.Base(initialMap), ".yaml")
		mapID = strings.TrimSuffix(mapID, ".yml")
	}

	g := game.NewGame(finalAssets, graphics, mapID, initialMapType, heroID, input, &game.DefaultAudioManager{}, debug, HeadlessVersion)
	
	// Ensure the correct map is loaded even if it was passed as a path to a template
	if g.GetContext().World.CurrentMapType.ID != mapID {
		if m, ok := g.Registries.Maps.Types[mapID]; ok {
			g.SetCurrentMapType(m)
		}
	}
	
	g.TriggerMapLoad()
	
	// Wait a bit for the world to populate
	time.Sleep(3 * time.Second)
	
	// Disable external AI
	g.SetAIManager(nil)

	// Bypass menu
	g.BypassMenu()
	g.SetSimulationMode(true)

	updatesPerTick := 1
	if fastMode {
		updatesPerTick = 5000 
	}

	ticks := 0
	startTime := time.Now()
	lastLogTime := time.Now()

	for {
		if durationSec > 0 && time.Since(startTime).Seconds() >= float64(durationSec) {
			log.Printf("Simulation duration reached: %v seconds", durationSec)
			break
		}

		for i := 0; i < updatesPerTick; i++ {
			err := g.Update()
			if err != nil {
				log.Printf("Error during update: %v", err)
				os.Exit(1)
			}
			ticks++
			
			if g.IsGameOver() {
				log.Printf("Simulation ended: Playable character died.")
				log.Printf("Death reason: %s", g.GetDeathReason())
				p := g.GetPlayableCharacter()
				if p != nil {
					log.Printf("Final State: HP=%.1f/%.1f | Hunger=%.2f | Thirst=%.2f | Fatigue=%.2f | Age=%.2f", 
						float64(p.State.HealthPoints), float64(p.State.MaxHealthPoints),
						p.State.Hunger, p.State.Thirst, p.State.Fatigue, p.State.Age.Current)
				}
				os.Exit(0)
			}
			
			// 10 year check
			// 1 year = 360 days = 360 * 17280 ticks = 6,220,800 ticks
			// 10 years = 62,208,000 ticks
			if ticks >= 62208000 {
				log.Printf("SUCCESS: Character survived for 10 years (%d ticks)!", ticks)
				p := g.GetPlayableCharacter()
				if p != nil {
					log.Printf("Final State: HP=%.1f/%.1f | Hunger=%.2f | Thirst=%.2f | Fatigue=%.2f | Age=%.2f", 
						float64(p.State.HealthPoints), float64(p.State.MaxHealthPoints),
						p.State.Hunger, p.State.Thirst, p.State.Fatigue, p.State.Age.Current)
				}
				os.Exit(0)
			}
		}

		if time.Since(lastLogTime) > 10*time.Second {
			p := g.GetPlayableCharacter()
			if p != nil {
				charCount := len(g.World.Characters)
				obsCount := len(g.World.Obstacles)
				piglets := 0
				matures := 0
				for _, c := range g.World.Characters {
					if strings.Contains(strings.ToLower(c.Config.ID), "piglet") { piglets++ }
				}
				for _, o := range g.World.Obstacles {
					if o.Alive && o.Archetype != nil && o.Archetype.IsCrop && o.GrowthStage >= 2 { matures++ }
				}
				log.Printf("[SIM] Progress: %d ticks | Age: %.2f | Hunger: %.1f | Thirst: %.1f | Pop: %d (%d piglets) | Mature Crops: %d | Obs: %d", 
					ticks, p.State.Age.Current, p.State.Hunger, p.State.Thirst, charCount, piglets, matures, obsCount)
			}
			lastLogTime = time.Now()
		}

		if !fastMode {
			time.Sleep(16 * time.Millisecond) // ~60 TPS
		}
	}

	log.Printf("Simulation complete. Total ticks: %d", ticks)
}
