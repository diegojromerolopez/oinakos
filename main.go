package main

import (
	"embed"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/* data/*
var assets embed.FS

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Oinakos - Isometric RPG")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := game.NewGame(assets)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
