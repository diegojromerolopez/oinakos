package game

import "image"

type ItemInstanceSaveData struct {
	ID         string  `yaml:"id"`
	Resistance int     `yaml:"resistance,omitempty"`
	X          float64 `yaml:"x,omitempty"`
	Y          float64 `yaml:"y,omitempty"`
}

type PlayerSaveData struct {
	Archetype string                           `yaml:"archetype"`
	X           float64                          `yaml:"x"`
	Y           float64                          `yaml:"y"`
	State `yaml:"state,inline"`
	XP          int                              `yaml:"xp"`
	Level       int                              `yaml:"level"`
	Kills       int                              `yaml:"kills"`
	MapKills    map[string]int                   `yaml:"map_kills"`
	
	PrimaryAttributes `yaml:",inline"`

	BaseAttack  int                              `yaml:"base_attack"`
	BaseDefense int                              `yaml:"base_defense"`
	BaseProtection int                           `yaml:"base_protection"`
	Submission     map[string]float64            `yaml:"submission,omitempty"`
	Denarii     int                              `yaml:"denarii"`
	Weapon      *Weapon                          `yaml:"weapon"`
	Inventory   []ItemInstanceSaveData           `yaml:"inventory,omitempty"`
	Slots       map[string]ItemInstanceSaveData  `yaml:"slots,omitempty"`
	Trauma      PhysicalTrauma                   `yaml:"trauma,omitempty"`
	SelectedModel string                        `yaml:"selected_model,omitempty"`
}

type NPCSaveData struct {
	Archetype string                           `yaml:"archetype,omitempty"`
	NPCID       string                           `yaml:"npc_id,omitempty"`
	X           float64                          `yaml:"x"`
	Y           float64                          `yaml:"y"`
	State `yaml:"state,inline"`
	Level       int                              `yaml:"level"`
	Behavior    string                           `yaml:"behavior"`
	Name        string                           `yaml:"name,omitempty"`
	Alignment   Alignment                        `yaml:"alignment,omitempty"`
	Group       string                           `yaml:"group,omitempty"`
	LeaderID    string                           `yaml:"leader_id,omitempty"`
	MustSurvive bool                             `yaml:"must_survive,omitempty"`
	
	PrimaryAttributes `yaml:",inline"`

	BaseAttack  int                              `yaml:"base_attack,omitempty"`
	BaseDefense int                              `yaml:"base_defense,omitempty"`
	BaseProtection int                           `yaml:"base_protection,omitempty"`
	Submission     map[string]float64            `yaml:"submission,omitempty"`
	Denarii     int                              `yaml:"denarii,omitempty"`
	Inventory   []ItemInstanceSaveData           `yaml:"inventory,omitempty"`
	Slots       map[string]ItemInstanceSaveData  `yaml:"slots,omitempty"`
	Trauma      PhysicalTrauma                   `yaml:"trauma,omitempty"`
	SelectedModel string                        `yaml:"selected_model,omitempty"`
}

type ObstacleSaveData struct {
	ID            string   `yaml:"id,omitempty"`
	Archetype   string   `yaml:"archetype"`
	X             *float64 `yaml:"x,omitempty"`
	Y             *float64 `yaml:"y,omitempty"`
	HealthPoints    int      `yaml:"health_points,omitempty"`
	CooldownTicks int      `yaml:"cooldown_ticks,omitempty"`
	Disabled      bool     `yaml:"disabled,omitempty"`
}

type ItemSaveData = ItemInstanceSaveData

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
		Heightmap     map[string]float64 `yaml:"heightmap,omitempty"`
	} `yaml:"map"`
	Player    PlayerSaveData     `yaml:"player"`
	Characters []NPCSaveData      `yaml:"characters"`
	Obstacles []ObstacleSaveData `yaml:"obstacles"`
	Items     []ItemSaveData     `yaml:"items,omitempty"`
}
