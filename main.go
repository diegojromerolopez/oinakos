package main

import (
	"embed"
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"strings"

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
	var configDir string
	flag.StringVar(&initialMap, "map", "", "Map YAML file to load (save/instance)")
	flag.StringVar(&initialMapType, "map-type", "", "Named map type to generate from")
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

	engine.InitAudio(assets)

	// Register sounds from both archetypes and NPCs
	registerEntitySounds := func(configs map[string]*game.EntityConfig) {
		for _, conf := range configs {
			if conf.AudioDir == "" {
				continue
			}
			entries, err := fs.ReadDir(assets, conf.AudioDir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".wav") {
					continue
				}
				stem := e.Name()[:len(e.Name())-4]
				key := conf.ID + "/" + stem
				engine.GlobalAudio.LoadSound(key, conf.AudioDir+"/"+e.Name())
			}
		}
	}

	archetypeReg := game.NewArchetypeRegistry()
	archetypeReg.LoadAll(assets)
	registerEntitySounds(archetypeReg.Archetypes)

	npcReg := game.NewNPCRegistry()
	npcReg.LoadAll(assets)
	registerEntitySounds(npcReg.NPCs)

	// Load audio for ALL playable characters: assets/audio/characters/<id>/<stem>.wav → "<id>/<stem>"
	charReg := game.NewPlayableCharacterRegistry()
	charReg.LoadAll(assets)
	registerEntitySounds(charReg.Characters)

	// Providers
	eg := engine.NewEbitenGraphics()
	ei := engine.NewEbitenInput()

	g := game.NewGame(assets, initialMap, initialMapType, ei, &game.DefaultAudioManager{}, debug)
	gr := game.NewGameRenderer(g, assets, eg)
	gr.LoadAssets(assets)

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
