package game

import "oinakos/internal/engine"

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

	if g.isAboutScreen {
		return mh.updateAboutScreen()
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

	if g.isInventoryOpen {
		return mh.updateInventoryScreen()
	}

	return nil
}

func (mh *MenuHandler) updateInventoryScreen() error {
	g := mh.game
	
	// Close Active Book Overlay
	if g.ActiveBook != nil {
		if g.input.IsKeyJustPressed(engine.KeyEscape) || g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
			g.ActiveBook = nil
		}
		// Block inventory actions while reading
		return nil
	}

	// Close via Escape ONLY (game.go handles 'I' key toggle)
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isInventoryOpen = false
		return nil
	}

	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		mx, my := g.input.MousePosition()
		
		dialogW, dialogH := 800, 500
		dialogX := (g.width - dialogW) / 2
		dialogY := (g.height - dialogH) / 2
		
		// 1. Close Button
		closeW, closeH := 100, 30
		closeX := dialogX + (dialogW - closeW) / 2
		closeY := dialogY + dialogH - 50
		if mx >= closeX && mx <= closeX+closeW && my >= closeY && my <= closeY+closeH {
			g.isInventoryOpen = false
			return nil
		}

		// 2. Paper Doll Slots
		dollCenterX := dialogX + 200
		dollCenterY := dialogY + 250
		slots := []struct {
			id string
			x  int
			y  int
		}{
			{"head", dollCenterX, dollCenterY - 120},
			{"shield", dollCenterX - 100, dollCenterY - 30},
			{"body", dollCenterX, dollCenterY - 30},
			{"weapon", dollCenterX + 100, dollCenterY - 30},
			{"ring1", dollCenterX - 100, dollCenterY + 40},
			{"ring2", dollCenterX + 100, dollCenterY + 40},
			{"legs", dollCenterX, dollCenterY + 90},
		}

		pc := g.playableCharacter
		for _, s := range slots {
			// Button is drawn at: sx+25, sy-20, 12x12
			btnX, btnY, btnW, btnH := s.x+25, s.y-20, 12, 12
			if mx >= btnX && mx <= btnX+btnW && my >= btnY && my <= btnY+btnH {
				if _, hasObj := pc.Slots[s.id]; hasObj {
					g.DropEquippedItem(&pc.Actor, s.id)
					return nil // Handled click
				}
			}
		}

		// 3. Backpack (Right Side)
		listStartX := dialogX + 400
		listStartY := dialogY + 80
		listW := dialogW - 420
		
		for i := 0; i < len(pc.Inventory); i++ {
			itemY := listStartY + 20 + i*40
			item := pc.Inventory[i]
			
			// Drop button (X) at listStartX+listW-30
			btnX, btnY, btnW, btnH := listStartX+listW-30, itemY+10, 15, 15
			if mx >= btnX && mx <= btnX+btnW && my >= btnY && my <= btnY+btnH {
				g.TryDrop(&pc.Actor, i)
				return nil // Handled click
			}

			// Read button (R) at listStartX+listW-50 if it has content
			if item.Content != "" {
				rBtnX, rBtnY, rBtnW, rBtnH := listStartX+listW-50, itemY+10, 15, 15
				if mx >= rBtnX && mx <= rBtnX+rBtnW && my >= rBtnY && my <= rBtnY+rBtnH {
					g.ActiveBook = item
					if item.Consumable {
						pc.Actor.ApplyPermanentEffects(item)
						// Remove from inventory
						pc.Inventory = append(pc.Inventory[:i], pc.Inventory[i+1:]...)
						pc.Actor.UpdateEffects()
					}
					return nil // Handled click
				}
			}
		}
	}

	return nil
}

