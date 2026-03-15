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

func (mm *MechanicsManager) UpdateFogOfWar(ctx *SystemContext) {
	if ctx.Registries != nil { // Settings check moved to ctx or world eventually
		// For now we keep it simple or pass it in. Assume logic stays here.
	}
	// Note: MechanicsManager will eventually go away or be truly stateless.
	// For now let's just use ctx.
	world := ctx.World
	radius := 8.0
	px, py := world.PlayableCharacter.X, world.PlayableCharacter.Y
	startX, endX := int(px-radius), int(px+radius)
	startY, endY := int(py-radius), int(py+radius)
	for x := startX; x <= endX; x++ {
		for y := startY; y <= endY; y++ {
			dx, dy := float64(x)-px, float64(y)-py
			if dx*dx+dy*dy <= radius*radius {
				world.ExploredTiles[image.Point{X: x, Y: y}] = true
			}
		}
	}
}

func (mm *MechanicsManager) UpdateProximityEffects(ctx *SystemContext) {
	world := ctx.World
	for _, o := range world.Obstacles {
		if !o.Alive || o.Archetype == nil { continue }
		entities := make([]ActorInterface, 0, len(world.NPCs)+1)
		entities = append(entities, world.PlayableCharacter)
		for _, n := range world.NPCs {
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
						if entity.GetActor().IsGiant() {
							continue
						}
						switch e := entity.(type) {
						case *PlayableCharacter: e.TakeDamage(action.Amount, ctx)
						case *NPC:
							e.TakeDamage(action.Amount, nil, ctx)
						}
						o.EffectTimers[entity] = 60
						world.FloatingTexts = append(world.FloatingTexts, &FloatingText{
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
						world.FloatingTexts = append(world.FloatingTexts, &FloatingText{
							Text: fmt.Sprintf("+%d", action.Amount), X: ex, Y: ey, Life: 45, Color: ColorHeal,
						})
					}
				}
			}
		}
	}
}

func (mm *MechanicsManager) CheckWinConditions(ctx *SystemContext) bool {
	world := ctx.World
	mapWon := false
	switch world.CurrentMapType.Type {
	case ObjKillCount:
		mapKillTotal := 0
		for _, v := range world.PlayableCharacter.MapKills { mapKillTotal += v }
		won, hasTarget := false, false
		if len(world.CurrentMapType.TargetKills) > 0 {
			hasTarget = true
			allMet := true
			for npcID, targetAmount := range world.CurrentMapType.TargetKills {
				if world.PlayableCharacter.MapKills[npcID] < targetAmount { allMet = false; break }
			}
			if allMet { won = true }
		}
		if world.CurrentMapType.TargetKillCount > 0 {
			hasTarget = true
			if mapKillTotal >= world.CurrentMapType.TargetKillCount { won = true } else { won = false }
		}
		if hasTarget && won { mapWon = true }
	case ObjSurvive:
		if world.CurrentMapType.TargetTime > 0 && world.PlayTime >= world.CurrentMapType.TargetTime {
			mapWon = true
		}
	case ObjReachPortal, ObjReachZone, ObjReachBuilding:
		dx, dy := world.PlayableCharacter.X-world.CurrentMapType.TargetPoint.X, world.PlayableCharacter.Y-world.CurrentMapType.TargetPoint.Y
		dist, radius := math.Sqrt(dx*dx+dy*dy), world.CurrentMapType.TargetRadius
		if radius <= 0 { radius = 2.0 }
		if dist < radius { mapWon = true }
	case ObjProtectNPC:
		if len(world.NPCs) > 0 {
			escort := world.NPCs[0]
			if !escort.IsAlive() { /* handle game over externally */ } else {
				dx, dy := escort.X-world.CurrentMapType.TargetPoint.X, escort.Y-world.CurrentMapType.TargetPoint.Y
				dist, radius := math.Sqrt(dx*dx+dy*dy), world.CurrentMapType.TargetRadius
				if radius <= 0 { radius = 5.0 }
				if dist < radius { mapWon = true }
			}
		}
	case ObjKillVIP: if len(world.NPCs) > 0 && !world.NPCs[0].IsAlive() { mapWon = true }
	case ObjPacifist:
		for _, kills := range world.PlayableCharacter.MapKills { if kills > 0 { /* handle game over externally */ break } }
	case ObjDestroyBuilding:
		if world.CurrentMapType.TargetObstacle != nil && !world.CurrentMapType.TargetObstacle.Alive { mapWon = true }
	}
	return mapWon
}
