package game

import (
	"fmt"
	"math/rand"
	"strings"
)

// Tesserae (Roman Dice Game)
// Typically played with three cubical dice.
// High rolls win. Venus throw (three different numbers sum to highest) or similar.
// To keep it simple and high-fidelity:
// We'll roll 3 dice (1-6). 
// 18 = Venus (High Win, 5x payout)
// 14+ = Win (2x payout)
// <14 = Loss

func (a *Actor) PlayTesserae(ctx *SystemContext, stake int) {
	if a.Denarii < stake {
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("[%s] cannot afford the %d denarii stake at the Fortune Home", a.Name, stake), LogNPC) }
		return
	}

	a.Denarii -= stake
	a.LastGamblingTick = a.Tick

	// Roll 3 dice
	d1 := rand.Intn(6) + 1
	d2 := rand.Intn(6) + 1
	d3 := rand.Intn(6) + 1
	total := d1 + d2 + d3

	resultStr := fmt.Sprintf("[%s] rolls the Tesserae: %d, %d, %d (Total: %d)", a.Name, d1, d2, d3, total)
	
	winAmount := 0
	winType := "LOSS"

	if total == 18 {
		winType = "VENUS (JACKPOT)"
		winAmount = stake * 5
	} else if total >= 14 {
		winType = "WIN"
		winAmount = stake * 2
	}

	if winAmount > 0 {
		a.Denarii += winAmount
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s -> %s! Wins %d Denarii.", resultStr, winType, winAmount), LogNPC) }
	} else {
		if ctx.Log != nil { ctx.Log(fmt.Sprintf("%s -> %s.", resultStr, winType), LogNPC) }
	}
}

func (g *Game) IsNearFortuneHome(p *Character) bool {
	for _, o := range g.obstacles {
		if strings.Contains(strings.ToLower(o.ID), "fortune_home") {
			dist := p.DistanceToObject(o)
			if dist < 5.0 { return true }
		}
	}
	return false
}
