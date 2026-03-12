package game

import "image"

type PlayerSaveData struct {
	ArchetypeID string         `yaml:"archetype_id"`
	X           float64        `yaml:"x"`
	Y           float64        `yaml:"y"`
	Health      int            `yaml:"health"`
	MaxHealth   int            `yaml:"max_health"`
	XP          int            `yaml:"xp"`
	Level       int            `yaml:"level"`
	Kills       int            `yaml:"kills"`
	MapKills    map[string]int `yaml:"map_kills"`
	BaseAttack  int            `yaml:"base_attack"`
	BaseDefense int            `yaml:"base_defense"`
	Weapon      string         `yaml:"weapon"`
}

type NPCSaveData struct {
	ArchetypeID string    `yaml:"archetype_id,omitempty"`
	NPCID       string    `yaml:"npc_id,omitempty"`
	X           float64 `yaml:"x"`
	Y           float64 `yaml:"y"`
	Health      int     `yaml:"health"`
	MaxHealth   int     `yaml:"max_health"`
	Level       int     `yaml:"level"`
	Behavior    string    `yaml:"behavior"`
	Name        string    `yaml:"name,omitempty"`
	Alignment   Alignment `yaml:"alignment,omitempty"`
	Group       string    `yaml:"group,omitempty"`
	LeaderID    string    `yaml:"leader_id,omitempty"`
	MustSurvive bool      `yaml:"must_survive,omitempty"`
	BaseAttack  int       `yaml:"base_attack,omitempty"`
	BaseDefense int       `yaml:"base_defense,omitempty"`
}

type ObstacleSaveData struct {
	ID            string   `yaml:"id,omitempty"`
	ArchetypeID   string   `yaml:"archetype_id"`
	X             *float64 `yaml:"x,omitempty"`
	Y             *float64 `yaml:"y,omitempty"`
	Health        int      `yaml:"health,omitempty"`
	CooldownTicks int      `yaml:"cooldown_ticks,omitempty"`
	Disabled      bool     `yaml:"disabled,omitempty"`
}

type SaveData struct {
	Map struct {
		ID           string  `yaml:"id"`
		WidthPixels  int     `yaml:"width_px"`
		HeightPixels int     `yaml:"height_px"`
		Level        int     `yaml:"level"`
		PlayTime     float64 `yaml:"play_time"`
		Overrides struct {
			TargetKillCount int            `yaml:"target_kill_count,omitempty"`
			TargetTime      float64        `yaml:"target_time,omitempty"`
			Difficulty      int            `yaml:"difficulty,omitempty"`
			SpawnFrequency  float64        `yaml:"spawn_frequency,omitempty"`
			SpawnAmount     int            `yaml:"spawn_amount,omitempty"`
			TargetKills     map[string]int `yaml:"target_kills,omitempty"`
			Name            string         `yaml:"name,omitempty"`
			Description     string         `yaml:"description,omitempty"`
		} `yaml:"overrides,omitempty"`
		FloorTile  string       `yaml:"floor_tile,omitempty"`
		FloorZones []*FloorZone `yaml:"floor_zones,omitempty"`
		ExploredTiles []image.Point `yaml:"explored_tiles,omitempty"`
	} `yaml:"map"`
	Player    PlayerSaveData     `yaml:"player"`
	NPCs      []NPCSaveData      `yaml:"npcs"`
	Obstacles []ObstacleSaveData `yaml:"obstacles"`
}
