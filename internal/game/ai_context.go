package game

import (
	"encoding/json"
	"fmt"
	"math"
)

type NPCContext struct {
	Name             string  `json:"name"`
	Description      string  `json:"description,omitempty"`
	Archetype        string  `json:"archetype"`
	HealthPct        int     `json:"health_pct"`
	Alignment        string  `json:"alignment"`
	DistanceToPlayer float64 `json:"distance_to_player"`
	Behavior         string  `json:"behavior"`
}

type PlayerContext struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Alignment   string `json:"alignment"`
	HealthPct   int    `json:"health_pct"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Kills       int    `json:"kills"`
	Level       int    `json:"level"`
}

type MapMission struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	TargetName  string  `json:"target_name,omitempty"`
	TargetX     float64 `json:"target_x,omitempty"`
	TargetY     float64 `json:"target_y,omitempty"`
}

type WorldContext struct {
	MapID      string       `json:"map_id"`
	Mission    MapMission   `json:"mission"`
	Player     PlayerContext `json:"player"`
	FocusNPC   *NPCContext   `json:"focus_npc,omitempty"`
	NearbyNPCs []NPCContext  `json:"nearby_npcs"`
}

func BuildWorldContext(g *Game, focusNPC *NPC) string {
	wc := WorldContext{
		MapID: g.currentMapType.ID,
		Mission: MapMission{
			Type:        g.currentMapType.Type.String(),
			Description: g.currentMapType.Description,
		},
		Player: PlayerContext{
			Name:        g.playableCharacter.Name,
			Description: g.playableCharacter.Config.Description,
			Alignment:   fmt.Sprint(g.playableCharacter.Alignment),
			HealthPct:   int(float64(g.playableCharacter.Health) / float64(g.playableCharacter.MaxHealth) * 100),
			X:           g.playableCharacter.X,
			Y:           g.playableCharacter.Y,
			Kills:       g.playableCharacter.Kills,
			Level:       g.playableCharacter.Level,
		},
		NearbyNPCs: []NPCContext{},
	}

	if g.currentMapType.TargetNPC != nil {
		wc.Mission.TargetName = g.currentMapType.TargetNPC.Name
	}
	if g.currentMapType.TargetPoint.X != 0 || g.currentMapType.TargetPoint.Y != 0 {
		wc.Mission.TargetX = g.currentMapType.TargetPoint.X
		wc.Mission.TargetY = g.currentMapType.TargetPoint.Y
	}

	var nearestNPC *NPC
	minDist := 999.0

	for _, n := range g.npcs {
		if !n.IsAlive() {
			continue
		}

		dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
		
		if focusNPC != nil && n == focusNPC {
			ctx := NPCContext{
				Name:             n.Name,
				Description:      n.Config.Description,
				Archetype:        n.Archetype.ID,
				HealthPct:        int(float64(n.Health) / float64(n.MaxHealth) * 100),
				Alignment:        fmt.Sprint(n.Alignment),
				DistanceToPlayer: dist,
				Behavior:         fmt.Sprint(n.Behavior),
			}
			wc.FocusNPC = &ctx
			continue
		}

		if dist < 20.0 && dist < minDist {
			minDist = dist
			nearestNPC = n
		}
	}

	if nearestNPC != nil {
		ctx := NPCContext{
			Name:             nearestNPC.Name,
			Description:      nearestNPC.Config.Description,
			Archetype:        nearestNPC.Archetype.ID,
			HealthPct:        int(float64(nearestNPC.Health) / float64(nearestNPC.MaxHealth) * 100),
			Alignment:        fmt.Sprint(nearestNPC.Alignment),
			DistanceToPlayer: minDist,
			Behavior:         fmt.Sprint(nearestNPC.Behavior),
		}
		wc.NearbyNPCs = append(wc.NearbyNPCs, ctx)
	}

	data, _ := json.Marshal(wc)
	return string(data)
}
