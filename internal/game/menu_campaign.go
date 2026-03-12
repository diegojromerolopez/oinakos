package game

import (
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updateCampaignSelect() error {
	g := mh.game
	nC := len(g.campaignRegistry.IDs)
	nM := len(g.mapTypeRegistry.IDs)

	if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
		g.campaignMenuIndex--
		if g.campaignMenuIndex < 0 {
			g.campaignMenuIndex = nC + nM
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
		g.campaignMenuIndex++
		if g.campaignMenuIndex > nC+nM {
			g.campaignMenuIndex = 0
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
		if g.campaignMenuIndex < nC {
			g.campaignMenuIndex += nC
			if g.campaignMenuIndex > nC+nM-1 {
				g.campaignMenuIndex = nC + nM - 1
			}
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
		if g.campaignMenuIndex >= nC && g.campaignMenuIndex < nC+nM {
			g.campaignMenuIndex -= nC
			if g.campaignMenuIndex < 0 {
				g.campaignMenuIndex = 0
			}
		}
	}

	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my

	hoverIndex := -1
	col1X := 100
	col2X := g.width / 2

	for i := 0; i < nC; i++ {
		cy := 130 + i*30
		if mx >= col1X && mx <= col1X+300 && my >= cy-15 && my <= cy+15 {
			hoverIndex = i
		}
	}
	for i := 0; i < nM; i++ {
		colOffset := col2X
		rowOffset := i
		if i > 15 {
			colOffset += 250
			rowOffset = i - 16
		}
		cy := 130 + rowOffset*30
		if mx >= colOffset && mx <= colOffset+300 && my >= cy-15 && my <= cy+15 {
			hoverIndex = nC + i
		}
	}
	if mx >= g.width/2-100 && mx <= g.width/2+100 && my >= g.height-110 && my <= g.height-70 {
		hoverIndex = nC + nM
	}

	if hoverIndex != -1 && mouseMoved {
		g.campaignMenuIndex = hoverIndex
	}

	quitText := "  QUIT"
	quitW := len(quitText) * 7
	qx, qy := (g.width-quitW)/2, g.height-90
	if mx >= qx && mx <= qx+300 && my >= qy && my <= qy+20 {
		hoverIndex = nC + nM
	}

	if hoverIndex != -1 {
		g.campaignMenuIndex = hoverIndex
	}

	handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

	if handleSelect {
		if g.campaignMenuIndex < nC {
			camID := g.campaignRegistry.IDs[g.campaignMenuIndex]
			g.currentCampaign = g.campaignRegistry.Campaigns[camID]
			g.isCampaign = true
			g.campaignIndex = 0
			g.isCampaignSelect = false
			g.worldManager.LoadMapLevel()
		} else if g.campaignMenuIndex < nC+nM {
			mapID := g.mapTypeRegistry.IDs[g.campaignMenuIndex-nC]
			g.currentMapType = *g.mapTypeRegistry.Types[mapID]
			g.isCampaign = false
			g.isCampaignSelect = false
			g.initialMapID = mapID
			g.worldManager.LoadMapLevel()
		} else {
			g.CloseWindow()
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isQuitConfirmationOpen = true
		g.quitConfirmationIndex = 1
	}
	return nil
}
