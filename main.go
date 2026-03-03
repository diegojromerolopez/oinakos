package main

import (
	"embed"
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"

	"oinakos/internal/engine"
	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets data
var assets embed.FS

func main() {
	var initialMap string
	var initialMapType string
	var loadGame string
	var debug bool
	flag.StringVar(&initialMap, "map", "", "Map YAML file to load (save/instance)")
	flag.StringVar(&initialMapType, "map-type", "", "Named map type to generate from")
	flag.StringVar(&loadGame, "load-game", "", "Saved game file to load (e.g. quicksaves/save_20240101_120000.yaml)")
	flag.BoolVar(&debug, "debug", false, "Show collision perimeters (red borders)")
	flag.Parse()

	// --load-game overrides --map
	if loadGame != "" {
		initialMap = loadGame
	}

	engine.InitAudio(assets)

	// Load archetype sounds dynamically: walk each archetype's AudioDir and register
	// every .wav as  "<archetypeID>/<stem>"  (e.g. "orc_male/hit", "orc_male/menace_1").
	archetypeReg := game.NewArchetypeRegistry()
	archetypeReg.LoadAll(assets)
	for _, arch := range archetypeReg.Archetypes {
		if arch.AudioDir == "" {
			continue
		}
		entries, err := fs.ReadDir(assets, arch.AudioDir)
		if err != nil {
			continue // no audio for this archetype
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if len(name) < 5 || name[len(name)-4:] != ".wav" {
				continue
			}
			stem := name[:len(name)-4]
			key := arch.ID + "/" + stem
			engine.GlobalAudio.LoadSound(key, arch.AudioDir+"/"+name)
		}
	}

	// Load main character sounds: assets/audio/characters/main/<stem>.wav → "main/<stem>"
	const mainAudioDir = "assets/audio/characters/main"
	if entries, err := fs.ReadDir(assets, mainAudioDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if len(name) < 5 || name[len(name)-4:] != ".wav" {
				continue
			}
			stem := name[:len(name)-4]
			engine.GlobalAudio.LoadSound("main/"+stem, mainAudioDir+"/"+name)
		}
	}

	// Providers
	eg := engine.NewEbitenGraphics()
	ei := engine.NewEbitenInput()

	g := game.NewGame(assets, initialMap, initialMapType, ei, &game.DefaultAudioManager{}, debug)
	gr := game.NewGameRenderer(g, assets, eg)
	gr.LoadAssets(assets)

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Oinakos - Isometric RPG")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Set Window Icon from player character sprite
	f, err := assets.Open("assets/images/characters/main/static.png")
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
