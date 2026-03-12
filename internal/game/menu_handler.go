package game

type MenuHandler struct {
	game *Game
}

func NewMenuHandler(g *Game) *MenuHandler {
	return &MenuHandler{game: g}
}

func (mh *MenuHandler) Update() error {
	g := mh.game
	if g.isQuitConfirmationOpen {
		return mh.updateQuitConfirmation()
	}

	if g.isMainMenu {
		return mh.updateMainMenu()
	}

	if g.isSettingsScreen {
		return mh.updateSettingsScreen()
	}

	if g.isCharacterSelect {
		return mh.updateCharacterSelect()
	}

	if g.isCampaignSelect {
		return mh.updateCampaignSelect()
	}

	if g.isGameWon {
		return mh.updateGameWon()
	}

	if g.isGameOver {
		return mh.updateGameOver()
	}

	if g.isMapWon {
		return mh.updateMapWon()
	}

	if g.isMenuOpen {
		return mh.updatePauseMenu()
	}

	return nil
}
