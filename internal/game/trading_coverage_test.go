package game

import (
	"testing"
)

func TestGame_TradingCoverage(t *testing.T) {
	ctx := NewTestContext()
	g := ctx.World.Game
	if g == nil {
		g = &Game{
			World:      ctx.World,
			Registries: ctx.Registries,
		}
		ctx.World.Game = g
	}
	pc := NewCharacter(0, 0, nil, 1, true, nil)
	pc.Name = "Player"
	pc.Denarii = 100
	g.playableCharacter = pc
	
	trader := NewCharacter(1, 1, nil, 1, false, nil)
	trader.Name = "Trader"
	trader.Denarii = 100
	g.ActiveTrader = trader
	
	item := &ItemInstance{Config: &ObjectConfig{ID: "sword", Name: "Sword", Value: 50}}
	trader.Inventory = append(trader.Inventory, item)
	
	// 1. BuyItem
	g.BuyItem(0)
	if len(pc.Inventory) == 0 { t.Error("Failed to buy item") }
	if pc.Denarii >= 100 { t.Error("Did not pay for item") }

	// 2. SellItem
	g.SellItem(0)
	if len(trader.Inventory) == 0 { t.Error("Failed to sell item") }
	if pc.Denarii <= 50 { t.Error("Did not get paid for item") }

	// 3. CalculateTradePrice - branches
	price := g.CalculateTradePrice(item, &pc.Actor, &trader.Actor, true)
	if price <= 0 { t.Error("Price should be positive") }
	
	// Sentiment effect
	trader.ModifySentiment(pc.Name, 100)
	friendlyPrice := g.CalculateTradePrice(item, &pc.Actor, &trader.Actor, true)
	if friendlyPrice >= price { t.Error("Friendly price should be lower") }
	
	// Incapacitated
	trader.ActionState = ActorIncapacitated
	lootedPrice := g.CalculateTradePrice(item, &pc.Actor, &trader.Actor, true)
	if lootedPrice != 0 { t.Errorf("Expected 0 price for incapacitated seller, got %d", lootedPrice) }
}
