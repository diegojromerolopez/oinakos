package main

import (
	"log"
	"oinakos/internal/engine"
	"github.com/hajimehoshi/ebiten/v2"
)

func NewMapEditor(g engine.Graphics, in engine.Input) *MapEditor {
	me := &MapEditor{
		Graphics:    g,
		Input:       in,
		Renderer:    engine.NewRenderer(),
		Mode:        "DIALOG",
		InWidth:     "640",
		InHeight:    "640",
		FloorImages: make(map[string]engine.Image),
	}
	me.loadLibrary()
	me.loadFloors()
	return me
}

func (m *MapEditor) Update() error {
	if m.Mode == "DIALOG" { return m.updateDialog() }
	return m.updateEditor()
}

func (m *MapEditor) Draw(screen *ebiten.Image) {
	if m.Mode == "DIALOG" { m.drawDialog(screen); return }
	m.drawEditor(screen)
}

func (m *MapEditor) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	graphics := engine.NewEbitenGraphics()
	input := engine.NewEbitenInput()
	editor := NewMapEditor(graphics, input)

	ebiten.SetWindowTitle("Oinakos Map Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(editor); err != nil {
		log.Fatal(err)
	}
}
