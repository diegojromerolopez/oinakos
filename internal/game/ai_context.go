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
	Alignment        string   `json:"alignment"`
	DistanceToPlayer float64  `json:"distance_to_player"`
	Behavior         string   `json:"behavior"`
	Denarii          int           `json:"denarii"`
	Hunger           int           `json:"hunger"`
	Thirst           int           `json:"thirst"`
	Fatigue          int           `json:"fatigue"`
	Inventory        []string      `json:"inventory"`
	Memories         []MemoryEvent `json:"memories"`
}

type PlayerContext struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Alignment   string `json:"alignment"`
	HealthPct   int    `json:"health_pct"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Kills       int      `json:"kills"`
	Level       int      `json:"level"`
	Denarii     int      `json:"denarii"`
	Inventory   []string `json:"inventory"`
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

func BuildWorldContext(g *Game, focusNPC *Character) string {
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
			HealthPct:   int(float64(g.playableCharacter.State.HealthPoints) / float64(g.playableCharacter.State.MaxHealthPoints) * 100),
			X:           g.playableCharacter.X,
			Y:           g.playableCharacter.Y,
			Kills:       g.playableCharacter.Kills,
			Level:       g.playableCharacter.Level,
			Denarii:     g.playableCharacter.Denarii,
			Inventory:   g.playableCharacter.GetInventoryNames(),
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

	var nearestNPC *Character
	minDist := 999.0

	for _, n := range g.World.Characters {
		if !n.IsAlive() || n.IsPlayerControlled {
			continue
		}

		dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
		
		if focusNPC != nil && n == focusNPC {
			ctx := NPCContext{
				Name:             n.Name,
				Description:      n.Config.Description,
				Archetype:        n.Config.Archetype,
				HealthPct:        int(float64(n.State.HealthPoints) / float64(n.State.MaxHealthPoints) * 100),
				Alignment:        fmt.Sprint(n.Alignment),
				DistanceToPlayer: dist,
				Behavior:         fmt.Sprint(n.Behavior),
				Denarii:          n.Denarii,
				Hunger:           int(n.State.Hunger),
				Thirst:           int(n.State.Thirst),
				Fatigue:          int(n.State.Fatigue),
				Inventory:        n.GetInventoryNames(),
				Memories:         n.Memories,
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
			Archetype:        nearestNPC.Config.Archetype,
			HealthPct:        int(float64(nearestNPC.State.HealthPoints) / float64(nearestNPC.State.MaxHealthPoints) * 100),
			Alignment:        fmt.Sprint(nearestNPC.Alignment),
			DistanceToPlayer: minDist,
			Behavior:         fmt.Sprint(nearestNPC.Behavior),
			Denarii:          nearestNPC.Denarii,
			Hunger:           int(nearestNPC.State.Hunger),
			Thirst:           int(nearestNPC.State.Thirst),
			Fatigue:          int(nearestNPC.State.Fatigue),
			Inventory:        nearestNPC.GetInventoryNames(),
			Memories:         nearestNPC.Memories,
		}
		wc.NearbyNPCs = append(wc.NearbyNPCs, ctx)
	}

	data, _ := json.Marshal(wc)
	return string(data)
}
