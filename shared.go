package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"oinakos/internal/engine"
	"oinakos/internal/game"
)

// Version is the current build version, injected at compile time via -ldflags
var Version = "0.2-alpha"

//go:embed assets data
var assets embed.FS

type combinedFS struct {
	local fs.FS
	embed embed.FS
}

func (c *combinedFS) Open(name string) (fs.File, error) {
	if c.local != nil {
		f, err := c.local.Open(name)
		if err == nil {
			return f, nil
		}
	}
	return c.embed.Open(name)
}

func setupCommon() (fs.FS, bool, string, string, string, bool, float64) {
	var initialMap string
	var initialMapType string
	var heroID string
	var loadGame string
	var debug bool
	var configDir string
	var fastSim bool
	var simDuration float64
	flag.StringVar(&initialMap, "map", "", "Map YAML file to load (save/instance)")
	flag.StringVar(&initialMapType, "map-type", "", "Named map type to generate from")
	flag.StringVar(&heroID, "hero", "", "Character ID to use as the playable character")
	flag.StringVar(&loadGame, "load-game", "", "Saved game file to load (e.g. quicksaves/save_20240101_120000.yaml)")
	flag.BoolVar(&debug, "debug", false, "Show collision perimeters (red borders)")
	flag.StringVar(&configDir, "config", "", "Config directory to use for settings and saves")
	flag.BoolVar(&fastSim, "fast", false, "Run simulation at maximum speed (1min real = 1mo game)")
	flag.Float64Var(&simDuration, "duration", 0, "Duration of simulation in minutes (0 = infinite)")
	flag.Parse()

	if debug {
		game.SetDebugMode(true)
	}
	log.Printf("Oinakos Engine %s starting...", Version)

	if configDir != "" {
		game.SetOinakosDir(configDir)
	}

	// --load-game overrides --map
	if loadGame != "" {
		initialMap = loadGame
	}

	// Setup combined filesystem: check for local "oinakos/" folder OR local "assets/" (dev mode)
	var finalAssets fs.FS = assets
	if _, err := os.Stat("oinakos"); err == nil {
		finalAssets = &combinedFS{local: os.DirFS("oinakos"), embed: assets}
	} else if _, err := os.Stat("assets"); err == nil {
		finalAssets = &combinedFS{local: os.DirFS("."), embed: assets}
	}

	// Discover fonts dynamically
	fonts := game.DiscoverFonts(finalAssets)
	game.SetFontOptions(fonts)

	engine.InitAudio(finalAssets)

	return finalAssets, debug, initialMap, initialMapType, heroID, fastSim, simDuration
}

func loadRegistries(finalAssets fs.FS) (*game.ArchetypeRegistry, *game.CharacterRegistry) {
	archetypeReg := game.NewArchetypeRegistry()
	archetypeReg.LoadAll(finalAssets)

	charReg := game.NewCharacterRegistry()
	charReg.LoadAll(finalAssets)

	return archetypeReg, charReg
}
