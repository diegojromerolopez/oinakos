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

	if g.isKeymapScreen {
		return mh.updateKeymapScreen()
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

	if g.isTradeOpen {
		return mh.updateTradeScreen()
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

	if g.input.IsKeyJustPressed(g.settings.GetKey("eat")) {
		pc := g.playableCharacter
		for i, item := range pc.Inventory {
			if item != nil && item.Config != nil && item.Config.Type == "consumable" {
				ctx := &SystemContext{
					World: g.World, Input: g.input, Audio: g.audio, Registries: g.Registries, 
					Log: g.LogEvent, AIManager: g.aiManager, Weather: g.World.State.Weather, Intensity: g.World.State.Intensity, Settings: g.settings,
				}
				if pc.Actor.ConsumeItem(item, ctx) {
					pc.Inventory = append(pc.Inventory[:i], pc.Inventory[i+1:]...)
					pc.Actor.UpdateEffects()
					if g.audio != nil { g.audio.PlayRandomSound("pickup") }
					return nil
				}
			}
		}
	}

	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		mx, my := g.input.MousePosition()
		
		dialogW, dialogH := 900, 600
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
		dollCenterX := dialogX + 220
		dollCenterY := dialogY + 300
		slots := []struct {
			id string
			x  int
			y  int
		}{
			{"head", dollCenterX, dollCenterY - 140},
			{"shield", dollCenterX - 110, dollCenterY - 40},
			{"body", dollCenterX, dollCenterY - 40},
			{"weapon", dollCenterX + 110, dollCenterY - 40},
			{"ring1", dollCenterX - 110, dollCenterY + 50},
			{"ring2", dollCenterX + 110, dollCenterY + 50},
			{"legs", dollCenterX, dollCenterY + 110},
		}

		pc := g.playableCharacter
		for _, s := range slots {
			// [X] Label hitbox: sx+30, sy-15 size 30x20
			btnX, btnY, btnW, btnH := s.x+30, s.y-25, 30, 25
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
			
			// Drop button [X] at listStartX+listW-35, it's 16pt text. Use approx 35x25 hitbox
			btnX, btnY, btnW, btnH := listStartX+listW-40, itemY+5, 40, 25
			if mx >= btnX && mx <= btnX+btnW && my >= btnY && my <= btnY+btnH {
				g.TryDrop(&pc.Actor, i)
				return nil // Handled click
			}

			// Read button [R] at listStartX+listW-75. Use approx 35x25 hitbox
			if item.Config != nil && item.Config.Content != "" {
				rBtnX, rBtnY, rBtnW, rBtnH := listStartX+listW-80, itemY+5, 40, 25
				if mx >= rBtnX && mx <= rBtnX+rBtnW && my >= rBtnY && my <= rBtnY+rBtnH {
					g.ActiveBook = item
					if item.Config.Consumable {
						pc.Actor.ApplyPermanentEffects(item.Config)
						// Remove from inventory
						pc.Inventory = append(pc.Inventory[:i], pc.Inventory[i+1:]...)
						pc.Actor.UpdateEffects()
					}
					return nil // Handled click
				}
			}

			// Eat button [E] at listStartX+listW-115.
			if item.Config != nil && item.Config.Type == "consumable" {
				uBtnX, uBtnY, uBtnW, uBtnH := listStartX+listW-120, itemY+5, 40, 25
				if mx >= uBtnX && mx <= uBtnX+uBtnW && my >= uBtnY && my <= uBtnY+uBtnH {
					ctx := &SystemContext{
						World:      g.World,
						Input:      g.input,
						Audio:      g.audio,
						Registries: g.Registries,
						Log:        g.LogEvent,
						AIManager:  g.aiManager,
						Weather:    g.World.State.Weather,
						Intensity:  g.World.State.Intensity,
					}
					if pc.Actor.ConsumeItem(item, ctx) {
						// Remove from inventory
						pc.Inventory = append(pc.Inventory[:i], pc.Inventory[i+1:]...)
						pc.Actor.UpdateEffects()
						if g.audio != nil { g.audio.PlayRandomSound("pickup") }
					}
					return nil // Handled click
				}
			}
		}
	}

	return nil
}

func (mh *MenuHandler) updateKeymapScreen() error {
	g := mh.game

	// If remapping an action, wait for any keypress
	if g.remappingAction != "" {
		keys := []engine.Key{}
		keys = g.input.AppendJustPressedKeys(keys)
		if len(keys) > 0 {
			newKey := keys[0]
			if newKey != engine.KeyEscape {
				g.settings.Keymap[g.remappingAction] = KeyToName(newKey)
				g.settings.Save()
			}
			g.remappingAction = ""
		}
		return nil
	}

	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isKeymapScreen = false
		g.isSettingsScreen = true
		return nil
	}

	if g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyUp) {
		g.keymapSelectedIndex--
		if g.keymapSelectedIndex < 0 { g.keymapSelectedIndex = len(RemappableActions) - 1 }
	}
	if g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyDown) {
		g.keymapSelectedIndex++
		if g.keymapSelectedIndex >= len(RemappableActions) { g.keymapSelectedIndex = 0 }
	}

	if g.input.IsKeyJustPressed(engine.KeyEnter) || g.input.IsKeyJustPressed(engine.KeySpace) {
		g.remappingAction = RemappableActions[g.keymapSelectedIndex].ID
	}

	return nil
}

func (mh *MenuHandler) updateDialogueScreen() error {
	g := mh.game
	if g.ActiveDialogue == nil { return nil }
	
	if g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyUp) {
		g.ActiveDialogue.SelectedChoice--
		if g.ActiveDialogue.SelectedChoice < 0 {
			g.ActiveDialogue.SelectedChoice = len(g.ActiveDialogue.Choices) - 1
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyDown) {
		g.ActiveDialogue.SelectedChoice++
		if g.ActiveDialogue.SelectedChoice >= len(g.ActiveDialogue.Choices) {
			g.ActiveDialogue.SelectedChoice = 0
		}
	}
	if g.input.IsKeyJustPressed(engine.KeyEnter) { g.AdvanceDialogue() }
	if g.input.IsKeyJustPressed(engine.KeyEscape) || g.input.IsKeyJustPressed(engine.KeyBackspace) {
		g.CloseDialogue()
	}
	return nil
}

