package game

import (
	"encoding/json"
	"os"
)

type SaveData struct {
	PlayerX   float64 `json:"player_x"`
	PlayerY   float64 `json:"player_y"`
	Kills     int     `json:"kills"`
	XP        int     `json:"xp"`
	Health    int     `json:"health"`
	MaxHealth int     `json:"max_health"`
	PlayTime  float64 `json:"play_time"`
}

func (g *Game) Save(path string) error {
	data := SaveData{
		PlayerX:   g.player.X,
		PlayerY:   g.player.Y,
		Kills:     g.player.Kills,
		XP:        g.player.XP,
		Health:    g.player.Health,
		MaxHealth: g.player.MaxHealth,
		PlayTime:  g.playTime,
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

func (g *Game) Load(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var data SaveData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}

	g.player.X = data.PlayerX
	g.player.Y = data.PlayerY
	g.player.Kills = data.Kills
	g.player.XP = data.XP
	g.player.Health = data.Health
	g.player.MaxHealth = data.MaxHealth
	g.playTime = data.PlayTime

	// Reset player state if they were dead?
	if g.player.Health > 0 {
		g.player.State = StateIdle
		g.isGameOver = false
	}

	return nil
}
