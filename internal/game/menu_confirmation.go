package game

import (
	"os"
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updateQuitConfirmation() error {
	g := mh.game
	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my

	pw, ph := 400, 200
	px, py := (g.width-pw)/2, (g.height-ph)/2

	hoverIndex := -1
	for i := 0; i < 2; i++ {
		itemY := py + 100 + i*40
		if mx >= px+100 && mx <= px+300 && my >= itemY-10 && my <= itemY+30 {
			hoverIndex = i
			break
		}
	}

	if hoverIndex != -1 && mouseMoved {
		g.quitConfirmationIndex = hoverIndex
	}

	if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
		g.quitConfirmationIndex = (g.quitConfirmationIndex - 1 + 2) % 2
	}
	if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
		g.quitConfirmationIndex = (g.quitConfirmationIndex + 1) % 2
	}

	handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

	if handleSelect {
		if g.quitConfirmationIndex == 0 { // Yes, quit
			if !g.isWasm() {
				os.Exit(0)
			}
		} else { // No, stay here
			g.isQuitConfirmationOpen = false
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isQuitConfirmationOpen = false
	}
	return nil
}
