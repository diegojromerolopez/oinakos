//go:build !headless

package main

import (
	"image"
	"log"

	"oinakos/internal/engine"
	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	finalAssets, debug, initialMap, initialMapType, heroID, _, _ := setupCommon()
	log.Printf("Starting Ebiten GUI mode...")

	loadRegistries(finalAssets)

	// Providers
	eg := engine.NewEbitenGraphics()
	ei := engine.NewEbitenInput()

	g := game.NewGame(finalAssets, eg, initialMap, initialMapType, heroID, ei, &game.DefaultAudioManager{}, debug, Version)

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
	f, err := finalAssets.Open("assets/images/characters/" + g.World.PlayableCharacter.Config.ID + "/static.png")
	if err != nil {
		log.Printf("Warning: failed to open icon file for hero %s: %v", g.World.PlayableCharacter.Config.ID, err)
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
