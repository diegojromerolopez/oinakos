package game

import (
	"fmt"
	"image/color"
	"strings"

	"oinakos/internal/engine"
)

func (gr *GameRenderer) drawBookOverlay(screen engine.Image) {
	g, book := gr.game, gr.game.ActiveBook
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 200}, false)
	dW, dH := 600, 400
	dX, dY := (g.width-dW)/2, (g.height-dH)/2
	gr.graphics.DrawFilledRect(screen, float32(dX-2), float32(dY-2), float32(dW+4), float32(dH+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(dX), float32(dY), float32(dW), float32(dH), color.RGBA{20, 20, 25, 255}, false)
	tw, _ := gr.graphics.MeasureText(book.Config.Name, 24)
	gr.graphics.DrawTextAt(screen, book.Config.Name, dX+(dW-int(tw))/2, dY+40, color.RGBA{218, 165, 32, 255}, 24)
	gr.drawWrappedText(screen, book.Config.Content, dX+40, dY+80, dW-80, color.White, 16, dY+dH-60)
	gr.graphics.DrawTextAt(screen, "[Click or ESC to close]", dX+(dW-150)/2, dY+dH-20, color.RGBA{150, 150, 150, 255}, 14)
}

func (gr *GameRenderer) drawTradeScreen(screen engine.Image) {
	g := gr.game
	if g.ActiveTrader == nil {
		return
	}
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 200}, false)

	dialogW, dialogH := 900, 600
	dialogX, dialogY := (g.width-dialogW)/2, (g.height-dialogH)/2
	gold := color.RGBA{218, 165, 32, 255}

	// Double Border
	gr.graphics.DrawFilledRect(screen, float32(dialogX-4), float32(dialogY-4), float32(dialogW+8), float32(dialogH+8), color.RGBA{50, 50, 50, 255}, false)
	gr.graphics.DrawFilledRect(screen, float32(dialogX-2), float32(dialogY-2), float32(dialogW+4), float32(dialogH+4), gold, false)
	gr.graphics.DrawFilledRect(screen, float32(dialogX), float32(dialogY), float32(dialogW), float32(dialogH), color.RGBA{10, 10, 12, 250}, false)

	// Title
	title := fmt.Sprintf("TRADING WITH %s", strings.ToUpper(g.ActiveTrader.Name))
	tw, _ := gr.graphics.MeasureText(title, 28)
	gr.graphics.DrawTextAt(screen, title, dialogX+(dialogW-int(tw))/2, dialogY+45, gold, 28)

	// Purses
	pPurse := fmt.Sprintf("Your Denarii: %d", g.playableCharacter.Denarii)
	tPurse := fmt.Sprintf("Trader Denarii: %d", g.ActiveTrader.Denarii)
	gr.graphics.DrawTextAt(screen, pPurse, dialogX+50, dialogY+85, color.White, 16)
	gr.graphics.DrawTextAt(screen, tPurse, dialogX+dialogW/2+50, dialogY+85, color.White, 16)

	// Columns
	listStartY := dialogY + 120
	columnW := (dialogW / 2) - 40

	// Player Items (Left) - Seller is Player, Buyer is Trader
	gr.drawInventoryColumn(screen, dialogX+20, listStartY, columnW, g.playableCharacter.Inventory, "SELL (Click to Sell)", &g.ActiveTrader.Actor, &g.playableCharacter.Actor, false)

	// Trader Items (Right) - Seller is Trader, Buyer is Player
	gr.drawInventoryColumn(screen, dialogX+dialogW/2+20, listStartY, columnW, g.ActiveTrader.Inventory, "BUY (Click to Buy)", &g.playableCharacter.Actor, &g.ActiveTrader.Actor, true)

	// Footer
	gr.graphics.DrawTextAt(screen, "[CLOSE]", dialogX+dialogW-100, dialogY+35, gold, 18)
}

func (gr *GameRenderer) drawInventoryColumn(screen engine.Image, x, y, w int, items []*ItemInstance, subtitle string, buyer, seller *Actor, isBuyingFromTrader bool) {
	gr.graphics.DrawTextAt(screen, subtitle, x+10, y-15, color.RGBA{150, 150, 150, 255}, 14)
	
	for i, it := range items {
		if i >= 10 { break } 
		itemY := y + i*40
		
		gr.graphics.DrawFilledRect(screen, float32(x), float32(itemY), float32(w), 38, color.RGBA{40, 40, 45, 150}, false)
		
		if it == nil || it.Config == nil { continue }
		
		if it.Config.Sprite != nil {
			_, sh := it.Config.Sprite.Size()
			scale := 32.0 / float64(sh)
			op := engine.NewDrawImageOptions()
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(x)+5, float64(itemY)+3)
			screen.DrawImage(it.Config.Sprite, op)
		}
		
		name := it.Config.Name
		if len(name) > 18 { name = name[:16] + ".." }
		gr.graphics.DrawTextAt(screen, name, x+45, itemY+25, color.White, 16)
		
		// Dynamic Cost
		val := gr.game.CalculateTradePrice(it, buyer, seller, isBuyingFromTrader)
		cost := fmt.Sprintf("%d den", val)
		cw, _ := gr.graphics.MeasureText(cost, 14)
		gr.graphics.DrawTextAt(screen, cost, x+w-int(cw)-10, itemY+25, color.RGBA{218, 165, 32, 255}, 14)
	}
}
