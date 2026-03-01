package main

import (
	"embed"
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
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

	// Load sounds
	engine.GlobalAudio.LoadSound("knight_attack", "assets/audio/knight/knight_attack.wav")
	engine.GlobalAudio.LoadSound("knight_hit", "assets/audio/knight/knight_hit.wav")
	engine.GlobalAudio.LoadSound("orc_hit", "assets/audio/orc/orc_hit.wav")
	engine.GlobalAudio.LoadSound("demon_hit", "assets/audio/demon/demon_hit.wav")
	engine.GlobalAudio.LoadSound("peasant_hit", "assets/audio/peasant/peasant_hit.wav")

	// Menace lines
	engine.GlobalAudio.LoadSound("orc_menace_1", "assets/audio/orc/orc_menace_1.wav")
	engine.GlobalAudio.LoadSound("orc_menace_2", "assets/audio/orc/orc_menace_2.wav")
	engine.GlobalAudio.LoadSound("orc_menace_3", "assets/audio/orc/orc_menace_3.wav")
	engine.GlobalAudio.LoadSound("orc_menace_4", "assets/audio/orc/orc_menace_4.wav")
	engine.GlobalAudio.LoadSound("orc_menace_5", "assets/audio/orc/orc_menace_5.wav")

	engine.GlobalAudio.LoadSound("demon_menace_1", "assets/audio/demon/demon_menace_1.wav")
	engine.GlobalAudio.LoadSound("demon_menace_2", "assets/audio/demon/demon_menace_2.wav")
	engine.GlobalAudio.LoadSound("demon_menace_3", "assets/audio/demon/demon_menace_3.wav")
	engine.GlobalAudio.LoadSound("demon_menace_4", "assets/audio/demon/demon_menace_4.wav")
	engine.GlobalAudio.LoadSound("demon_menace_5", "assets/audio/demon/demon_menace_5.wav")

	engine.GlobalAudio.LoadSound("peasant_menace_1", "assets/audio/peasant/peasant_menace_1.wav")
	engine.GlobalAudio.LoadSound("peasant_menace_2", "assets/audio/peasant/peasant_menace_2.wav")
	engine.GlobalAudio.LoadSound("peasant_menace_3", "assets/audio/peasant/peasant_menace_3.wav")
	engine.GlobalAudio.LoadSound("peasant_menace_4", "assets/audio/peasant/peasant_menace_4.wav")
	engine.GlobalAudio.LoadSound("peasant_menace_5", "assets/audio/peasant/peasant_menace_5.wav")

	// Death cries
	engine.GlobalAudio.LoadSound("knight_death", "assets/audio/knight/knight_death.wav")
	engine.GlobalAudio.LoadSound("orc_death", "assets/audio/orc/orc_death.wav")
	engine.GlobalAudio.LoadSound("demon_death", "assets/audio/demon/demon_death.wav")
	engine.GlobalAudio.LoadSound("peasant_death", "assets/audio/peasant/peasant_death.wav")

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
