package main

import (
	"embed"
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"os"

	"oinakos/internal/engine"
	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

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

func main() {
	var initialMap string
	var initialMapType string
	var heroID string
	var loadGame string
	var debug bool
	var configDir string
	flag.StringVar(&initialMap, "map", "", "Map YAML file to load (save/instance)")
	flag.StringVar(&initialMapType, "map-type", "", "Named map type to generate from")
	flag.StringVar(&heroID, "hero", "", "Character ID to use as the playable character")
	flag.StringVar(&loadGame, "load-game", "", "Saved game file to load (e.g. quicksaves/save_20240101_120000.yaml)")
	flag.BoolVar(&debug, "debug", false, "Show collision perimeters (red borders)")
	flag.StringVar(&configDir, "config", "", "Config directory to use for settings and saves")
	flag.Parse()

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

	archetypeReg := game.NewArchetypeRegistry()
	archetypeReg.LoadAll(finalAssets)

	npcReg := game.NewNPCRegistry()
	npcReg.LoadAll(finalAssets)

	// Load audio for ALL playable characters: assets/audio/characters/<id>/<stem>.wav → "<id>/<stem>"
	charReg := game.NewPlayableCharacterRegistry()
	charReg.LoadAll(finalAssets)

	// Providers
	eg := engine.NewEbitenGraphics()
	// Load initial font (default to first available or first from settings)
	ei := engine.NewEbitenInput()

	g := game.NewGame(finalAssets, initialMap, initialMapType, heroID, ei, &game.DefaultAudioManager{}, debug)
	
	// Hook font update
	g.SetOnFontUpdate(func(fontName string) {
		if fontName == "default" {
			eg.LoadFont(nil, "")
			return
		}
		fontPath := "assets/fonts/" + fontName + ".ttf"
		if err := eg.LoadFont(finalAssets, fontPath); err != nil {
			log.Printf("Warning: failed to reload font %s: %v", fontPath, err)
		}
	})
	// Initial font apply from loaded settings
	s := game.LoadSettings()
	if s.Font != "" {
		g.UpdateFont()
	} else {
		// Fallback to medieval if available, or just use default
		eg.LoadFont(finalAssets, "assets/fonts/medieval.ttf")
	}

	gr := game.NewGameRenderer(g, finalAssets, eg)
	go gr.LoadAssets(finalAssets)

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Oinakos")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Set Window Icon from player character sprite
	f, err := assets.Open("assets/images/characters/oinakos/static.png")
	if err != nil {
		log.Printf("Warning: failed to open icon file: %v", err)
	} else {
		defer f.Close()
		iconImg, _, err := image.Decode(f)
		if err != nil {
			log.Printf("Warning: failed to decode icon image: %v", err)
		} else {
			// Apply project standard transparency (removing lime green)
			transparentIcon := engine.Transparentize(iconImg)
			ebiten.SetWindowIcon([]image.Image{transparentIcon})
			log.Println("Success: Window icon set from player character sprite.")
		}
	}

	// Create a single screen wrapper to avoid reflecting/allocating every frame
	screenWrapper := engine.NewEbitenImageWrapper(nil)

	if err := ebiten.RunGame(&gameWithRenderer{g, gr, screenWrapper}); err != nil {
		log.Fatal(err)
	}
}

type gameWithRenderer struct {
	*game.Game
	gr *game.GameRenderer

	screenWrapper *engine.EbitenImageWrapper
}

func (g *gameWithRenderer) Draw(screen *ebiten.Image) {
	if g.screenWrapper == nil {
		g.screenWrapper = engine.NewEbitenImageWrapper(screen)
	}
	// We must ensure the wrapper always points to the current active screen
	// because Ebiten replaces the *ebiten.Image behind the scenes!
	g.screenWrapper.UpdateRaw(screen)
	// Draw the game
	g.gr.Draw(g.screenWrapper)
}
