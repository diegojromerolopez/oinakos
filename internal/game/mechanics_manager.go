package game

import (
	"fmt"
	"image"
	"math"
	"oinakos/internal/engine"
)

type MechanicsManager struct {
	game *Game
}

func NewMechanicsManager(g *Game) *MechanicsManager {
	return &MechanicsManager{game: g}
}

func (mm *MechanicsManager) UpdateFogOfWar() {
	g := mm.game
	if g.settings == nil || g.settings.FogOfWar == "none" {
		return
	}
	radius := 8.0
	px, py := g.playableCharacter.X, g.playableCharacter.Y
	startX, endX := int(px-radius), int(px+radius)
	startY, endY := int(py-radius), int(py+radius)
	for x := startX; x <= endX; x++ {
		for y := startY; y <= endY; y++ {
			dx, dy := float64(x)-px, float64(y)-py
			if dx*dx+dy*dy <= radius*radius {
				g.ExploredTiles[image.Point{X: x, Y: y}] = true
			}
		}
	}
}

func (mm *MechanicsManager) UpdateProximityEffects() {
	g := mm.game
	if g.isPaused || g.isGameOver || g.isMapWon {
		return
	}
	for _, o := range g.obstacles {
		if !o.Alive || o.Archetype == nil { continue }
		entities := make([]interface{}, 0, len(g.npcs)+1)
		entities = append(entities, g.playableCharacter)
		for _, n := range g.npcs {
			if n.IsAlive() { entities = append(entities, n) }
		}
		for _, entity := range entities {
			var ex, ey float64
			var eFootprint engine.Polygon
			var isMC bool
			switch e := entity.(type) {
			case *PlayableCharacter: ex, ey, eFootprint, isMC = e.X, e.Y, e.GetFootprint(), true
			case *NPC: ex, ey, eFootprint = e.X, e.Y, e.GetFootprint()
			default: continue
			}
			for _, action := range o.Archetype.Actions {
				inRange := false
				if action.Aura > 0 {
					dist := math.Sqrt(math.Pow(ex-o.X, 2) + math.Pow(ey-o.Y, 2))
					if dist <= action.Aura { inRange = true }
				} else {
					if engine.CheckCollision(eFootprint, o.GetFootprint()) { inRange = true }
				}
				if !inRange { continue }
				if action.Type == ActionHarm {
					if o.EffectTimers[entity] <= 0 {
						switch e := entity.(type) {
						case *PlayableCharacter: e.TakeDamage(action.Amount, g.audio)
						case *NPC: e.TakeDamage(action.Amount, nil, nil, g.audio, g.npcs)
						}
						o.EffectTimers[entity] = 60
						g.floatingTexts = append(g.floatingTexts, &FloatingText{
							Text: fmt.Sprintf("-%d", action.Amount), X: ex, Y: ey, Life: 45, Color: ColorHarm,
						})
					}
				} else if action.Type == ActionHeal && !action.RequiresInteraction {
					allowed := true
					if action.AlignmentLimit != "" && action.AlignmentLimit != "all" {
						var alignment Alignment
						if isMC { alignment = AlignmentAlly } else { alignment = entity.(*NPC).Alignment }
						if (action.AlignmentLimit == "enemy" && alignment != AlignmentEnemy) || (action.AlignmentLimit == "ally" && alignment != AlignmentAlly) {
							allowed = false
						}
					}
					if allowed && o.EffectTimers[entity] <= 0 {
						switch e := entity.(type) {
						case *PlayableCharacter: e.Heal(action.Amount)
						case *NPC: e.Heal(action.Amount)
						}
						o.EffectTimers[entity] = 60
						g.floatingTexts = append(g.floatingTexts, &FloatingText{
							Text: fmt.Sprintf("+%d", action.Amount), X: ex, Y: ey, Life: 45, Color: ColorHeal,
						})
					}
				}
			}
		}
	}
}

func (mm *MechanicsManager) CheckWinConditions() bool {
	g := mm.game
	mapWon := false
	switch g.currentMapType.Type {
	case ObjKillCount:
		mapKillTotal := 0
		for _, v := range g.playableCharacter.MapKills { mapKillTotal += v }
		won, hasTarget := false, false
		if len(g.currentMapType.TargetKills) > 0 {
			hasTarget = true
			allMet := true
			for npcID, targetAmount := range g.currentMapType.TargetKills {
				if g.playableCharacter.MapKills[npcID] < targetAmount { allMet = false; break }
			}
			if allMet { won = true }
		}
		if g.currentMapType.TargetKillCount > 0 {
			hasTarget = true
			if mapKillTotal >= g.currentMapType.TargetKillCount { won = true } else { won = false }
		}
		if hasTarget && won { mapWon = true }
	case ObjSurvive: if g.playTime >= g.currentMapType.TargetTime { mapWon = true }
	case ObjReachPortal, ObjReachZone, ObjReachBuilding:
		dx, dy := g.playableCharacter.X-g.currentMapType.TargetPoint.X, g.playableCharacter.Y-g.currentMapType.TargetPoint.Y
		dist, radius := math.Sqrt(dx*dx+dy*dy), g.currentMapType.TargetRadius
		if radius <= 0 { radius = 2.0 }
		if dist < radius { mapWon = true }
	case ObjProtectNPC:
		if len(g.npcs) > 0 {
			escort := g.npcs[0]
			if !escort.IsAlive() { g.isGameOver = true } else {
				dx, dy := escort.X-g.currentMapType.TargetPoint.X, escort.Y-g.currentMapType.TargetPoint.Y
				dist, radius := math.Sqrt(dx*dx+dy*dy), g.currentMapType.TargetRadius
				if radius <= 0 { radius = 5.0 }
				if dist < radius { mapWon = true }
			}
		}
	case ObjKillVIP: if len(g.npcs) > 0 && !g.npcs[0].IsAlive() { mapWon = true } else if len(g.npcs) == 0 && g.playTime > 2 { mapWon = true }
	case ObjPacifist:
		if g.playTime >= g.currentMapType.TargetTime { mapWon = true }
		for _, kills := range g.playableCharacter.MapKills { if kills > 0 { g.isGameOver = true; break } }
	case ObjDestroyBuilding:
		if g.currentMapType.TargetObstacle != nil && !g.currentMapType.TargetObstacle.Alive { mapWon = true }
	}
	return mapWon
}
