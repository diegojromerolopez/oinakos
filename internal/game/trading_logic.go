package game

import (
	"fmt"
	"math"
)

// BuyItem moves an item from the trader to the player if they have enough denarii.
func (g *Game) BuyItem(index int) {
	if g.ActiveTrader == nil || index < 0 || index >= len(g.ActiveTrader.Inventory) {
		return
	}
	pc := g.playableCharacter
	trader := g.ActiveTrader
	item := trader.Inventory[index]

	price := g.CalculateTradePrice(item, &pc.Actor, &trader.Actor, true)
	if pc.Denarii >= price {
		pc.Denarii -= price
		trader.Denarii += price

		// Move item to player's inventory
		pc.Inventory = append(pc.Inventory, item)
		// Remove from trader's inventory
		trader.Inventory = append(trader.Inventory[:index], trader.Inventory[index+1:]...)

		if g.audio != nil {
			g.audio.PlayRandomSound("pickup")
		}
		
		// Trade builds reputation
		trader.ModifySentiment(pc.Name, 2.0)
		trader.AddMemory(trader.Tick, "trade", pc.Name, 2.0)
		
		g.LogEvent(fmt.Sprintf("Bought %s for %d denarii.", item.Config.Name, price), LogInfo)
	} else {
		g.LogEvent("Not enough denarii!", LogInfo)
	}
}

// SellItem moves an item from the player to the trader and gives them denarii.
func (g *Game) SellItem(index int) {
	if g.ActiveTrader == nil || index < 0 || index >= len(g.playableCharacter.Inventory) {
		return
	}
	pc := g.playableCharacter
	trader := g.ActiveTrader
	item := pc.Inventory[index]

	price := g.CalculateTradePrice(item, &trader.Actor, &pc.Actor, false)
	// Trader needs enough money to buy
	if trader.Denarii < price {
		g.LogEvent("Trader cannot afford that!", LogInfo)
		return
	}

	// Player gets money
	pc.Denarii += price
	// Trader pays the player
	trader.Denarii -= price

	// Move item to trader's inventory
	trader.Inventory = append(trader.Inventory, item)
	// Remove from player's inventory
	pc.Inventory = append(pc.Inventory[:index], pc.Inventory[index+1:]...)

	if g.audio != nil {
		g.audio.PlayRandomSound("pickup")
	}
	
	// Trade builds reputation
	trader.ModifySentiment(pc.Name, 1.0) // Selling gives less because merchant takes risk
	trader.AddMemory(trader.Tick, "trade", pc.Name, 1.0)
	
	g.LogEvent(fmt.Sprintf("Sold %s for %d denarii.", item.Config.Name, price), LogInfo)
}

func (g *Game) CalculateTradePrice(item *ItemInstance, buyer, seller *Actor, isBuyingFromTrader bool) int {
	if item == nil || item.Config == nil { return 0 }
	base := float64(item.Config.Value)
	
	// 1. Sentiment Multiplier
	sentiment := 0.0
	if buyer != nil && seller != nil {
		if s, ok := seller.Relationships[buyer.Name]; ok {
			sentiment = s
		}
	}
	// Scale: -100 (hate) -> 1.5x price, 100 (love) -> 0.5x price
	sentimentMult := 1.0 - (sentiment / 200.0)
	
	// 2. Scarcity Multiplier (Seller's stock)
	stockCount := 0
	if seller != nil {
		for _, it := range seller.Inventory {
			if it != nil && it.Config != nil && it.Config.ID == item.Config.ID {
				stockCount++
			}
		}
	}
	// More stock -> lower price. 1 item = 1.0x, 10 items = ~0.5x
	stockMult := 1.1 / (1.0 + float64(stockCount)*0.1)
	
	final := base * sentimentMult * stockMult
	
	// Incapacitated characters can be looted for free
	if seller != nil && seller.ActionState == ActorIncapacitated {
		return 0
	}

	// Merchant Margin
	if isBuyingFromTrader {
		final *= 1.25 // Buy at a premium
	} else {
		final *= 0.75 // Sell at a discount
	}
	
	val := int(math.Round(final))
	if val < 1 { val = 1 }
	return val
}

// BuyLodging allows a character to pay for a stay at the inn.
func (g *Game) BuyLodging() {
	if g.ActiveTrader == nil {
		return
	}
	pc := g.playableCharacter
	// 5 Denarii for 12 hours of rest rights
	price := 5
	if pc.Denarii >= price {
		pc.Denarii -= price
		g.ActiveTrader.Denarii += price
		pc.LodgingTicks += 8640 // 12 hours
		
		if g.audio != nil {
			g.audio.PlayRandomSound("pickup")
		}
		
		g.LogEvent("Paid 5 Denarii for a night at the inn.", LogInfo)
		g.AddFloatingText("Paid Lodging", pc.X, pc.Y, ColorHeal)
	} else {
		g.LogEvent("Not enough Denarii for the inn!", LogInfo)
	}
}
