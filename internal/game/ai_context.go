package game

import (
	"encoding/json"
	"fmt"
	"math"
)

type NPCContext struct {
	Name             string  `json:"name"`
	Archetype        string  `json:"archetype"`
	HealthPct        int     `json:"health_pct"`
	Alignment        string  `json:"alignment"`
	DistanceToPlayer float64 `json:"distance_to_player"`
	Behavior         string  `json:"behavior"`
}

type PlayerContext struct {
	Name      string `json:"name"`
	HealthPct int    `json:"health_pct"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Kills     int    `json:"kills"`
	Level     int    `json:"level"`
}

type WorldContext struct {
	MapID      string       `json:"map_id"`
	Player     PlayerContext `json:"player"`
	FocusNPC   *NPCContext   `json:"focus_npc,omitempty"`
	NearbyNPCs []NPCContext  `json:"nearby_npcs"`
}

func BuildWorldContext(g *Game, focusNPC *NPC) string {
	wc := WorldContext{
		MapID: g.currentMapType.ID,
		Player: PlayerContext{
			Name:      g.playableCharacter.Name,
			HealthPct: int(float64(g.playableCharacter.Health) / float64(g.playableCharacter.MaxHealth) * 100),
			X:         g.playableCharacter.X,
			Y:         g.playableCharacter.Y,
			Kills:     g.playableCharacter.Kills,
			Level:     g.playableCharacter.Level,
		},
		NearbyNPCs: []NPCContext{},
	}

	for _, n := range g.npcs {
		if !n.IsAlive() {
			continue
		}

		dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
		
		// Only include NPCs within a reasonable range (e.g., 20 units)
		if dist > 20.0 {
			continue
		}

		ctx := NPCContext{
			Name:             n.Name,
			Archetype:        n.Archetype.ID,
			HealthPct:        int(float64(n.Health) / float64(n.MaxHealth) * 100),
			Alignment:        fmt.Sprint(n.Alignment),
			DistanceToPlayer: dist,
			Behavior:         fmt.Sprint(n.Behavior),
		}

		if focusNPC != nil && n == focusNPC {
			wc.FocusNPC = &ctx
		} else {
			wc.NearbyNPCs = append(wc.NearbyNPCs, ctx)
		}
	}

	data, _ := json.Marshal(wc)
	return string(data)
}
