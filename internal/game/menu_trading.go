package game

import "oinakos/internal/engine"

func (mh *MenuHandler) updateTradeScreen() error {
	g := mh.game
	if g.ActiveTrader == nil {
		g.isTradeOpen = false
		return nil
	}

	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isTradeOpen = false
		return nil
	}

	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		mx, my := g.input.MousePosition()

		// Match UI layout in drawTradeScreen (see menu_trading_render.go)
		dialogW, dialogH := 900, 600
		bx, by := (g.width-dialogW)/2, (g.height-dialogH)/2

		// Close button hitbox [CLOSE]
		if mx >= bx+dialogW-100 && mx <= bx+dialogW-20 && my >= by+20 && my <= by+55 {
			g.isTradeOpen = false
			return nil
		}

		// Left Column: Player Inventory
		pX, lY := bx+20, by+100
		for i := 0; i < len(g.playableCharacter.Inventory); i++ {
			itemY := lY + i*40
			if mx >= pX && mx <= pX+400 && my >= itemY && my <= itemY+40 {
				g.SellItem(i)
				return nil
			}
		}

		// Right Column: Trader Inventory
		tX := bx + dialogW/2 + 20
		for i := 0; i < len(g.ActiveTrader.Inventory); i++ {
			itemY := lY + i*40
			if mx >= tX && mx <= tX+400 && my >= itemY && my <= itemY+40 {
				g.BuyItem(i)
				return nil
			}
		}
	}

	return nil
}
