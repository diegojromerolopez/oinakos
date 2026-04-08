//go:build headless

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"oinakos/internal/engine"
	"oinakos/internal/game"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	initialMap      string
	initialMapType  string
	heroID          string
	debug           bool
	fastLevel       int
	durationStr     string
	lastLoggedEvent int
	simStepFlag     int
	logChan         = make(chan string, 10000)
)

func logWorker() {
	for msg := range logChan {
		log.Println(msg)
	}
}

func asyncLog(msg string) {
	logChan <- msg
}

const (
	TicksPerHour  = 720
	TicksPerDay   = 17280
	TicksPerMonth = 17280 * 30
	TicksPerYear  = 17280 * 360
)

const HeadlessVersion = "1.0.0-headless"

func parseDurationTicks(s string) int {
	if s == "" { return 0 }
	
	// If it's just a number, treat as ticks for backward compatibility or direct control
	if val, err := strconv.Atoi(s); err == nil { return val }

	total := 0
	
	// Regex to find matches like "10y", "2d", "4h", "1m"
	re := regexp.MustCompile(`(\d+)\s*([ydmh])`)
	matches := re.FindAllStringSubmatch(strings.ToLower(s), -1)
	
	for _, m := range matches {
		val, _ := strconv.Atoi(m[1])
		unit := m[2]
		switch unit {
		case "y": total += val * TicksPerYear
		case "m": total += val * TicksPerMonth
		case "d": total += val * TicksPerDay
		case "h": total += val * TicksPerHour
		}
	}
	
	return total
}

func main() {
	flag.StringVar(&initialMap, "map", "data/maps/venburgo.yaml", "Initial map to load")
	flag.StringVar(&initialMapType, "map-type", "", "Initial map type ID")
	flag.StringVar(&heroID, "hero", "oinakos", "Initial hero ID")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.IntVar(&fastLevel, "fastlevel", 0, "Simulation speed level (0-5, 0=off, 5=ultra)")
	flag.StringVar(&durationStr, "duration", "0", "Simulation duration (e.g. '10y 2d 4h' or ticks)")
	flag.IntVar(&simStepFlag, "simstep", 10, "Ticks between biological updates (1-100, default 10)")
	flag.Parse()

	go logWorker()
	log.Printf("Starting Oinakos Headless Simulation (Fast Level: %d)...", fastLevel)

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
	g.GetContext().Settings.SimStep = simStepFlag
	game.SetGlobalHeadlessLogger(asyncLog)
	
	// Ensure the correct map is loaded even if it was passed as a path to a template
	if g.GetContext().World.CurrentMapType.ID != mapID {
		if m, ok := g.Registries.Maps.Types[mapID]; ok {
			g.SetCurrentMapType(m)
		}
	}
	
	g.TriggerMapLoad()
	
	// Wait a tiny bit for the world to populate
	time.Sleep(100 * time.Millisecond)
	
	// Disable external AI
	g.SetAIManager(nil)

	// Bypass menu
	g.BypassMenu()
	g.SetSimulationMode(true)
	if fastLevel > 0 { game.SetFastMode(true) }

	updatesPerTick := 1 
	switch fastLevel {
	case 1: updatesPerTick = 100
	case 2: updatesPerTick = 1000
	case 3: updatesPerTick = 10000
	case 4: updatesPerTick = 100000
	case 5: updatesPerTick = 500000
	default: updatesPerTick = 1
	}

	ticks := 0
	ticksGoal := parseDurationTicks(durationStr)
	lastLogTime := time.Now()

	log.Printf("Simulation Goal: %s (%d ticks)", durationStr, ticksGoal)

	for {
		for i := 0; i < updatesPerTick; i++ {
			err := g.Update()
			if err != nil {
				log.Printf("Error during update: %v", err)
				os.Exit(1)
			}
			ticks++
			
			if ticks%100000 == 0 {
				asyncLog(fmt.Sprintf("[SIM] Progress: Ticks=%d", ticks))
			}

			if g.IsGameOver() {
				asyncLog("Simulation ended: Playable character died.")
				asyncLog(fmt.Sprintf("Death reason: %s", g.GetDeathReason()))
				p := g.GetPlayableCharacter()
				if p != nil {
					asyncLog(fmt.Sprintf("Final State: HP=%.1f/%.1f | Hunger=%.2f | Thirst=%.2f | Fatigue=%.2f | Age=%.2f", 
						float64(p.State.HealthPoints), float64(p.State.MaxHealthPoints),
						p.State.Hunger, p.State.Thirst, p.State.Fatigue, p.State.Age.Current))
				}
				time.Sleep(100 * time.Millisecond) // Allow logs to drain
				os.Exit(0)
			}
			
			if ticksGoal > 0 && ticks >= ticksGoal {
				asyncLog(fmt.Sprintf("SUCCESS: Simulation duration reached: %s (%d ticks)!", durationStr, ticks))
				p := g.GetPlayableCharacter()
				if p != nil {
					asyncLog(fmt.Sprintf("Final Residency State: HP=%.1f/%.1f | Hunger=%.2f | Thirst=%.2f | Fatigue=%.2f | Age=%.2f", 
						float64(p.State.HealthPoints), float64(p.State.MaxHealthPoints),
						p.State.Hunger, p.State.Thirst, p.State.Fatigue, p.State.Age.Current))
				}
				time.Sleep(100 * time.Millisecond) // Allow logs to drain
				os.Exit(0)
			}

			if ticks%500000 == 0 { // ~Monthly
				p := g.GetPlayableCharacter()
				if p != nil {
					charCount := len(g.World.Characters)
					d := g.World.Demographics
					asyncLog(fmt.Sprintf("[SIM] Month: %d | Ticks: %d | Population: %d | Births: %d | Deaths(Nat/Viol): %d/%d | Mating: %d (%d Preg)", 
						ticks/500000, ticks, charCount, d.BirthsHumans, d.DeathsNatural, d.DeathsViolent, d.MatingActs, d.MatingPregancies))
				}
				lastLoggedEvent = len(g.EventLog)
			}
		}

		if time.Since(lastLogTime) > 60*time.Second {
			log.Printf("[SIM] Still running... Ticks=%d", ticks)
			lastLogTime = time.Now()
		}

		if fastLevel == 0 {
			time.Sleep(16 * time.Millisecond) // ~60 TPS
		}
	}

	log.Printf("Simulation complete. Total ticks: %d", ticks)
}
