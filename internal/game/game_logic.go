package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"oinakos/internal/engine"
)

func (g *Game) updateOcclusion() {
	g.handleOcclusionFeedback(&g.playableCharacter.Actor)
	for _, n := range g.characters {
		g.handleOcclusionFeedback(&n.Actor)
	}
}

func (g *Game) handleOcclusionFeedback(a *Actor) {
	if a == nil || !a.IsAlive() { return }

	sortY := GetActorSortY(a)
	isoX, isoY := engine.CartesianToIso(a.X, a.Y)

	isNowOccluded := false
	for _, o := range g.obstacles {
		if !o.Alive { continue }
		if GetObstacleSortY(o) > sortY {
			if IsPointCoveredByObstacle(o, isoX, isoY) {
				isNowOccluded = true
				break
			}
		}
	}
	a.IsOccluded = isNowOccluded
}

func (g *Game) logRealtimePosition() {
	isIllegal := g.playableCharacter.checkCollisionAt(g.playableCharacter.X, g.playableCharacter.Y, g.obstacles)
	status := "OK"
	if isIllegal { status = "ILLEGAL POSITION (INSIDE BUILDING)" }
	nearestDist := 999.0
	nearestName := "None"
	nearestX, nearestY := 0.0, 0.0
	for _, o := range g.obstacles {
		dist := math.Sqrt(math.Pow(g.playableCharacter.X-o.X, 2) + math.Pow(g.playableCharacter.Y-o.Y, 2))
		if dist < nearestDist {
			nearestDist = dist
			if o.Archetype != nil {
				nearestName = o.Archetype.Name
				nearestX, nearestY = o.X, o.Y
			}
		}
	}
	DebugLog("[REALTIME] Player Pos: X=%.2f, Y=%.2f | Status: %s | Nearest: %s (Dist: %.2f) at (%.2f, %.2f)",
		g.playableCharacter.X, g.playableCharacter.Y, status, nearestName, nearestDist, nearestX, nearestY)
}

func (g *Game) ensurePlayerNotStuck() {
	for i := 0; i < 50; i++ {
		if !g.playableCharacter.checkCollisionAt(g.playableCharacter.X, g.playableCharacter.Y, g.obstacles) {
			break
		}
		g.playableCharacter.X += rand.Float64()*2 - 1
		g.playableCharacter.Y += rand.Float64()*2 - 1
		ncX, ncY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
		g.camera.SnapTo(ncX, ncY)
	}
}

func (g *Game) tryUnloading() {
	pc := g.playableCharacter
	if pc == nil || pc.TemporalState.HealthPoints <= 0 { return }

	nearWarehouse := false
	for _, obj := range g.obstacles {
		if !obj.Alive || obj.Archetype == nil { continue }
		if obj.Archetype.ID == "warehouse" || obj.Archetype.ID == "smithery" {
			dx := pc.X - obj.X
			dy := pc.Y - obj.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 40.0 { nearWarehouse = true; break }
		}
	}
	if !nearWarehouse { return }

	unloadedCount := 0
	var newInventory []*ItemInstance
	for _, item := range pc.Inventory {
		if item == nil || item.Config == nil { continue }
		switch item.Config.ID {
		case "wood", "lumber", "lumber2", "iron_ore", "copper_ore", "gold_ore", "silver_ore":
			unloadedCount++
		default:
			newInventory = append(newInventory, item)
		}
	}

	if unloadedCount > 0 {
		pc.Inventory = newInventory
		if g.audio != nil { g.audio.PlayRandomSound("pickup") }
		g.World.FloatingTexts = append(g.World.FloatingTexts, &FloatingText{
			Text:  fmt.Sprintf("Unloaded %d Materials!", unloadedCount),
			X:     pc.X, Y: pc.Y, Life: 90, Color: color.RGBA{50, 255, 50, 255},
		})
		pc.UpdateEffects()
	} else {
		g.World.FloatingTexts = append(g.World.FloatingTexts, &FloatingText{
			Text: "No materials to unload", X: pc.X, Y: pc.Y, Life: 60, Color: color.RGBA{200, 200, 200, 255},
		})
	}
}
