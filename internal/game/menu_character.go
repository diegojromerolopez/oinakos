package game

import (
	"oinakos/internal/engine"
)

func (mh *MenuHandler) updateCharacterSelect() error {
	g := mh.game
	nChars := len(g.playableCharacterRegistry.IDs)
	nOptions := nChars + 1 // +1 for "Back"
	if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyLeft) || g.input.IsKeyJustPressed(engine.KeyA) {
		g.characterMenuIndex--
		if g.characterMenuIndex < 0 {
			g.characterMenuIndex = nOptions - 1
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyRight) || g.input.IsKeyJustPressed(engine.KeyD) {
		g.characterMenuIndex++
		if g.characterMenuIndex >= nOptions {
			g.characterMenuIndex = 0
		}
	}

	mx, my := g.input.MousePosition()
	mouseMoved := mx != g.lastMouseX || my != g.lastMouseY
	g.lastMouseX, g.lastMouseY = mx, my

	hoverIndex := -1
	for i := 0; i < nChars; i++ {
		itemY := 130 + i*35
		if mx >= 100 && mx <= 600 && my >= itemY-15 && my <= itemY+15 {
			hoverIndex = i
		}
	}
	backY := 130 + nChars*35 + 20
	if mx >= 100 && mx <= 400 && my >= backY-15 && my <= backY+15 {
		hoverIndex = nChars
	}
	if g.characterMenuIndex < nChars {
		px, py := g.width/2+50, 130+30
		if mx >= px && mx <= px+240 && my >= py && my <= py+240 {
			hoverIndex = g.characterMenuIndex
		}
	}

	isClick := g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft)
	if hoverIndex != -1 && (mouseMoved || isClick) {
		g.characterMenuIndex = hoverIndex
	}

	handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && isClick)

	if handleSelect {
		if g.characterMenuIndex < nChars {
			charID := g.playableCharacterRegistry.IDs[g.characterMenuIndex]
			config := g.playableCharacterRegistry.Characters[charID]
			g.initialHeroID = charID
			g.playableCharacter.Config = config
			g.playableCharacter.Health = config.Stats.HealthMin
			g.playableCharacter.MaxHealth = config.Stats.HealthMin
			g.playableCharacter.Speed = config.Stats.Speed
			g.playableCharacter.BaseAttack = config.Stats.BaseAttack
			g.playableCharacter.BaseDefense = config.Stats.BaseDefense
			g.playableCharacter.Weapon = config.Weapon
			g.playableCharacter.Name = config.Name

			g.isCharacterSelect = false
			if g.initialMapID == "" && g.initialMapTypeID == "" {
				g.isCampaignSelect = true
			} else {
				go g.worldManager.LoadMapLevel()
			}
		} else {
			g.isCharacterSelect = false
			g.isMainMenu = true
			g.characterMenuIndex = 0
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isCharacterSelect = false
		g.isMainMenu = true
		g.characterMenuIndex = 0
	}
	return nil
}
