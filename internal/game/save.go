package game

import (
	"fmt"
	"io/fs"
	"log"
	"oinakos/internal/engine"
	"os"
	"path"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

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
	ArchetypeID string  `yaml:"archetype_id"`
	X           float64 `yaml:"x"`
	Y           float64 `yaml:"y"`
	Health      int     `yaml:"health"`
	MaxHealth   int     `yaml:"max_health"`
	Level       int     `yaml:"level"`
	Behavior    string  `yaml:"behavior"`
	Name        string  `yaml:"name,omitempty"`
	BaseAttack  int     `yaml:"base_attack,omitempty"`
	BaseDefense int     `yaml:"base_defense,omitempty"`
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
		// Optional overrides — any non-zero value here replaces the map_type equivalent
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
		FloorTile string `yaml:"floor_tile,omitempty"`
	} `yaml:"map"`
	Player    PlayerSaveData     `yaml:"player"`
	NPCs      []NPCSaveData      `yaml:"npcs"`
	Obstacles []ObstacleSaveData `yaml:"obstacles"`
}

func (g *Game) Save(fpath string) error {
	bytes, err := g.serialize()
	if err != nil {
		DebugLog("Failed to serialize for save to %s: %v", fpath, err)
		return err
	}
	err = os.WriteFile(fpath, bytes, 0644)
	if err == nil {
		DebugLog("Game Successfully Saved to %s | NPCs: %d | Obstacles: %d", fpath, len(g.npcs), len(g.obstacles))
	}
	return err
}

func (g *Game) performQuicksave() {
	if g.isWasm() {
		data, err := g.serialize()
		if err != nil {
			log.Printf("Failed to serialize game: %v", err)
			return
		}
		if err := g.saveToLocalStorage(data); err == nil {
			g.saveMessage = "Saved in Browser Storage"
			g.saveMessageTimer = 300
		}
		return
	}

	oinakosDir := GetOinakosDir()
	saveDir := filepath.Join(oinakosDir, "saves")
	if err := os.MkdirAll(saveDir, 0755); err == nil {
		savePath := filepath.Join(saveDir, fmt.Sprintf("quicksave-%s.oinakos.yaml", time.Now().Format("2006-01-02T150405")))
		if err := g.Save(savePath); err == nil {
			log.Printf("Game quicksaved: %s", savePath)
			g.lastSavePath = savePath
			absPath, err := filepath.Abs(savePath)
			if err != nil {
				absPath = savePath // Fallback
			}
			g.saveMessage = "Saved in " + absPath
			g.saveMessageTimer = 300 // 5 seconds at 60fps
		} else {
			log.Printf("Failed to quicksave: %v", err)
		}
	} else {
		log.Printf("Failed to create saves directory: %v", err)
	}
}

func (g *Game) serialize() ([]byte, error) {
	data := SaveData{}
	data.Map.ID = g.currentMapType.ID
	data.Map.WidthPixels = g.currentMapType.WidthPixels
	data.Map.HeightPixels = g.currentMapType.HeightPixels
	data.Map.Level = g.mapLevel
	data.Map.PlayTime = g.playTime

	data.Player = PlayerSaveData{
		ArchetypeID: g.mainCharacter.Config.ID,
		X:           g.mainCharacter.X,
		Y:           g.mainCharacter.Y,
		Health:      g.mainCharacter.Health,
		MaxHealth:   g.mainCharacter.MaxHealth,
		XP:          g.mainCharacter.XP,
		Level:       g.mainCharacter.Level,
		Kills:       g.mainCharacter.Kills,
		MapKills:    g.mainCharacter.MapKills,
		BaseAttack:  g.mainCharacter.BaseAttack,
		BaseDefense: g.mainCharacter.BaseDefense,
	}
	if g.mainCharacter.Weapon != nil {
		data.Player.Weapon = g.mainCharacter.Weapon.Name
	}

	for _, n := range g.npcs {
		if n.Archetype == nil {
			continue
		}
		behaviorStr := ""
		switch n.Behavior {
		case BehaviorWander:
			behaviorStr = "wander"
		case BehaviorPatrol:
			behaviorStr = "patrol"
		case BehaviorKnightHunter:
			behaviorStr = "hunter"
		case BehaviorNpcFighter:
			behaviorStr = "fighter"
		case BehaviorChaotic:
			behaviorStr = "chaotic"
		}

		data.NPCs = append(data.NPCs, NPCSaveData{
			ArchetypeID: n.Archetype.ID,
			X:           n.X,
			Y:           n.Y,
			Health:      n.Health,
			MaxHealth:   n.MaxHealth,
			Level:       n.Level,
			Behavior:    behaviorStr,
		})
	}

	for _, o := range g.obstacles {
		if o.Archetype == nil {
			continue
		}
		xVal, yVal := o.X, o.Y
		data.Obstacles = append(data.Obstacles, ObstacleSaveData{
			ID:            o.ID,
			ArchetypeID:   o.Archetype.ID,
			X:             &xVal,
			Y:             &yVal,
			Health:        o.Health,
			CooldownTicks: o.CooldownTicks,
		})
	}

	return yaml.Marshal(data)
}

func (g *Game) Load(fpath string) error {
	var bytes []byte
	var err error

	DebugLog("Attempting to load: %s", fpath)
	// WASM LocalStorage check
	if g.isWasm() && (fpath == "" || fpath == "quicksave") {
		bytes, err = g.loadFromLocalStorage()
		if err == nil && bytes != nil {
			DebugLog("Loaded from Browser Storage")
			return g.unmarshal(bytes, fpath)
		}
	}

	// Try reading from embedded assets first (for map instances in WASM)
	if g.assets != nil {
		// Ensure path is forward-slashes for embed.FS
		cleanPath := path.Clean(fpath)
		bytes, err = fs.ReadFile(g.assets, cleanPath)
	}

	// Fallback to real filesystem (for native saves)
	if err != nil || g.assets == nil {
		bytes, err = os.ReadFile(fpath)
	}

	if err != nil {
		DebugLog("Failed to read save file %s: %v", fpath, err)
		return fmt.Errorf("failed to read save file %s: %w", fpath, err)
	}

	return g.unmarshal(bytes, fpath)
}

func (g *Game) unmarshal(bytes []byte, fpath string) error {
	var data SaveData
	if err := yaml.Unmarshal(bytes, &data); err != nil {
		DebugLog("Failed to unmarshal save data: %v", err)
		return fmt.Errorf("failed to unmarshal save data: %w", err)
	}

	// Restore Map Level and PlayTime
	g.mapLevel = data.Map.Level
	g.playTime = data.Map.PlayTime

	// Sanitize player data from save
	sanitizePlayerSaveData(&data.Player, fpath)

	// Try to restore the map config if it exists
	if m, ok := g.mapTypeRegistry.Types[data.Map.ID]; ok {
		g.currentMapType = *m
	} else {
		log.Printf("Warning: saved map type ID %s not found in registry", data.Map.ID)
	}

	// Apply per-instance overrides on top of the loaded map type
	ov := data.Map.Overrides
	if ov.TargetKillCount > 0 {
		g.currentMapType.TargetKillCount = ov.TargetKillCount
	}
	if ov.TargetTime > 0 {
		g.currentMapType.TargetTime = ov.TargetTime
	}
	if ov.Difficulty > 0 {
		g.currentMapType.Difficulty = ov.Difficulty
	}
	if len(ov.TargetKills) > 0 {
		g.currentMapType.TargetKills = ov.TargetKills
	}
	if ov.Name != "" {
		g.currentMapType.Name = ov.Name
	}
	if ov.Description != "" {
		g.currentMapType.Description = ov.Description
	}
	if data.Map.FloorTile != "" {
		g.currentMapType.FloorTile = data.Map.FloorTile
	}

	// Restore Player
	if data.Player.ArchetypeID != "" {
		if config, ok := g.playableCharacterRegistry.Characters[data.Player.ArchetypeID]; ok {
			g.mainCharacter.Config = config
			g.isCharacterSelect = false
			g.isMainMenu = false
			// Note: We might need to reload assets for this config if not already loaded
		} else {
			log.Printf("Warning: saved character archetype ID %s not found in registry", data.Player.ArchetypeID)
		}
	}

	g.mainCharacter.X = data.Player.X
	g.mainCharacter.Y = data.Player.Y
	g.mainCharacter.Health = data.Player.Health
	g.mainCharacter.MaxHealth = data.Player.MaxHealth
	g.mainCharacter.XP = data.Player.XP
	g.mainCharacter.Level = data.Player.Level
	g.mainCharacter.Kills = data.Player.Kills
	g.mainCharacter.MapKills = data.Player.MapKills
	if g.mainCharacter.MapKills == nil {
		g.mainCharacter.MapKills = make(map[string]int)
	}
	g.mainCharacter.BaseAttack = data.Player.BaseAttack
	g.mainCharacter.BaseDefense = data.Player.BaseDefense
	if data.Player.Weapon != "" {
		g.mainCharacter.Weapon = GetWeaponByName(data.Player.Weapon)
	}

	if g.mainCharacter.Health > 0 {
		g.mainCharacter.State = StateIdle
		g.isGameOver = false
	} else {
		g.mainCharacter.State = StateDead
		g.isGameOver = true
	}

	g.mapWonMenuIndex = WinMenuContinue

	// Snap camera
	pIsoX, pIsoY := engine.CartesianToIso(g.mainCharacter.X, g.mainCharacter.Y)
	g.camera.SnapTo(pIsoX, pIsoY)

	// Restore NPCs
	g.npcs = nil
	for i, nData := range data.NPCs {
		// Sanitize NPC data
		sanitizeNPCSaveData(&nData, i, fpath)

		config, ok := g.archetypeRegistry.Archetypes[nData.ArchetypeID]
		if !ok {
			// Fallback: check NPC registry for unique/named NPCs
			config, ok = g.npcRegistry.NPCs[nData.ArchetypeID]
			if !ok {
				var archIDs []string
				for id := range g.archetypeRegistry.Archetypes {
					archIDs = append(archIDs, id)
				}
				var npcIDs []string
				for id := range g.npcRegistry.NPCs {
					npcIDs = append(npcIDs, id)
				}
				log.Printf("DEBUG: Lookup failed for %q. Archetypes Loaded: %v, NPCs Loaded: %v", nData.ArchetypeID, archIDs, npcIDs)
			}
		}

		if !ok {
			log.Printf("Warning: saved NPC/Archetype ID %s not found in any registry", nData.ArchetypeID)
			continue
		}
		n := NewNPC(nData.X, nData.Y, config, nData.Level)
		n.Health = nData.Health
		n.MaxHealth = nData.MaxHealth
		if nData.Name != "" {
			n.Name = nData.Name
		}
		if nData.BaseAttack > 0 {
			n.BaseAttack = nData.BaseAttack
		}
		if nData.BaseDefense > 0 {
			n.BaseDefense = nData.BaseDefense
		}

		switch nData.Behavior {
		case "wander":
			n.Behavior = BehaviorWander
		case "patrol":
			n.Behavior = BehaviorPatrol
		case "hunter":
			n.Behavior = BehaviorKnightHunter
		case "fighter":
			n.Behavior = BehaviorNpcFighter
		case "chaotic":
			n.Behavior = BehaviorChaotic
		}

		if n.Health <= 0 {
			n.State = NPCDead
		}
		g.npcs = append(g.npcs, n)
	}

	// Restore Obstacles with merging logic
	g.obstacles = nil

	// Index pre-spawn obstacles from the base map type by ID
	preSpawns := make(map[string]PreSpawnObstacle)
	for _, ps := range g.currentMapType.Obstacles {
		if ps.ID != "" {
			preSpawns[ps.ID] = ps
		}
	}

	// Track which pre-spawns were handled by overrides in the save file
	handledPreSpawns := make(map[string]bool)

	for _, oData := range data.Obstacles {
		// Matching logic: if save data has an ID, look for a matching pre-spawn
		var base *PreSpawnObstacle
		if oData.ID != "" {
			if ps, ok := preSpawns[oData.ID]; ok {
				base = &ps
				handledPreSpawns[oData.ID] = true
			}
		}

		if oData.Disabled {
			continue
		}

		// Determine archetype
		archID := oData.ArchetypeID
		if archID == "" && base != nil {
			archID = base.Archetype
		}

		config, ok := g.obstacleRegistry.Archetypes[archID]
		if !ok {
			log.Printf("Warning: saved obstacle archetype ID %s not found", archID)
			continue
		}

		// Determine Position
		px, py := 0.0, 0.0
		if oData.X != nil {
			px = *oData.X
		} else if base != nil && base.X != nil {
			px = *base.X
		}

		if oData.Y != nil {
			py = *oData.Y
		} else if base != nil && base.Y != nil {
			py = *base.Y
		}

		o := NewObstacle(oData.ID, px, py, config)
		if oData.Health > 0 || oData.X != nil { // Use save data health if provided
			o.Health = oData.Health
		} else if config.Health > 0 {
			o.Health = config.Health
		}

		o.CooldownTicks = oData.CooldownTicks
		if o.Health <= 0 && config.Health > 0 {
			o.Alive = false
		}
		g.obstacles = append(g.obstacles, o)
	}

	// Add any pre-spawns that were NOT explicitly overriden or disabled in the save file
	for _, ps := range g.currentMapType.Obstacles {
		if !handledPreSpawns[ps.ID] && !ps.Disabled {
			if config, ok := g.obstacleRegistry.Archetypes[ps.Archetype]; ok {
				px, py := 0.0, 0.0
				if ps.X != nil {
					px = *ps.X
				}
				if ps.Y != nil {
					py = *ps.Y
				}
				g.obstacles = append(g.obstacles, NewObstacle(ps.ID, px, py, config))
			}
		}
	}

	DebugLog("Game Successfully Unmarshaled: %s | Level: %d | NPCs: %d | Obstacles: %d", fpath, g.mapLevel, len(g.npcs), len(g.obstacles))
	return nil
}
