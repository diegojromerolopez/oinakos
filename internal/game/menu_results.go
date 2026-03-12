package game

import (
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updateGameWon() error {
	g := mh.game
	if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
		g.mapWonMenuIndex = 0
	}
	if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
		g.mapWonMenuIndex = 1
	}

	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my

	hoverIndex := -1
	for i := 0; i < 2; i++ {
		itemY := 200 + i*40
		if mx >= g.width/2-100 && mx <= g.width/2+100 && my >= itemY && my <= itemY+30 {
			hoverIndex = i
		}
	}
	if hoverIndex != -1 && mouseMoved {
		g.mapWonMenuIndex = hoverIndex
	}

	handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

	if handleSelect {
		if g.mapWonMenuIndex == 0 { // Replay
			*g = *NewGame(g.assets, g.initialMapID, g.initialMapTypeID, g.initialHeroID, g.input, g.audio, g.debug)
		} else { // Quit
			g.isQuitConfirmationOpen = true
			g.quitConfirmationIndex = 1
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isQuitConfirmationOpen = true
		g.quitConfirmationIndex = 1
	}
	return nil
}

func (mh *MenuHandler) updateGameOver() error {
	g := mh.game
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.CloseWindow()
	}
	if g.input.IsKeyJustPressed(engine.KeyEnter) || g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		*g = *NewGame(g.assets, g.initialMapID, g.initialMapTypeID, g.initialHeroID, g.input, g.audio, g.debug)
	}
	return nil
}

func (mh *MenuHandler) updateMapWon() error {
	g := mh.game
	if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
		g.mapWonMenuIndex = 0
	}
	if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
		g.mapWonMenuIndex = 1
	}

	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my

	hoverIndex := -1
	if mx >= g.width/2-100 && mx <= g.width/2+100 {
		if my >= g.height/2+50 && my <= g.height/2+80 {
			hoverIndex = 0 // Continue
		} else if my >= g.height/2+90 && my <= g.height/2+120 {
			hoverIndex = 1 // Quit
		}
	}
	if hoverIndex != -1 && mouseMoved {
		g.mapWonMenuIndex = hoverIndex
	}

	handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft)

	if handleSelect {
		if g.mapWonMenuIndex == WinMenuContinue {
			if g.isCampaign && g.currentCampaign != nil {
				g.campaignIndex++
				if g.campaignIndex >= len(g.currentCampaign.Maps) {
					g.isGameWon = true
					g.isMapWon = false
				} else {
					go g.worldManager.LoadMapLevel()
					g.isMapWon = false
				}
			} else {
				g.isGameWon = true
				g.isMapWon = false
			}
		} else if g.mapWonMenuIndex == WinMenuQuit {
			g.isQuitConfirmationOpen = true
			g.quitConfirmationIndex = 1
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isQuitConfirmationOpen = true
		g.quitConfirmationIndex = 1
	}
	return nil
}
