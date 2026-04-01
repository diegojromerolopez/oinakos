package game

import (
	"fmt"
	"image"
	"io/fs"
	"log"
	"oinakos/internal/engine"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

func (g *Game) Load(fpath string) error {
	var bytes []byte
	var err error

	DebugLog("Attempting to load: %s", fpath)
	if g.isWasm() && (fpath == "" || fpath == "quicksave") {
		bytes, err = g.loadFromLocalStorage()
		if err == nil && bytes != nil {
			DebugLog("Loaded from Browser Storage")
			return g.unmarshal(bytes, fpath)
		}
	}

	if g.assets != nil {
		cleanPath := path.Clean(fpath)
		bytes, err = fs.ReadFile(g.assets, cleanPath)
	}

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

	var raw map[string]any
	yaml.Unmarshal(bytes, &raw)
	if _, isTemplate := raw["width_px"]; isTemplate {
		return fmt.Errorf("file appears to be a map template (width_px at top level), not a save file")
	}
	if _, isTemplate := raw["floor_tile"]; isTemplate {
		return fmt.Errorf("file appears to be a map template (floor_tile at top level), not a save file")
	}

	g.mapLevel = data.Map.Level
	g.playTime = data.Map.PlayTime

	sanitizePlayerSaveData(&data.Player, fpath)

	if m, ok := g.mapTypeRegistry.Types[data.Map.ID]; ok {
		g.currentMapType = *m
	} else {
		log.Printf("Warning: saved map type ID %s not found in registry", data.Map.ID)
	}

	ov := data.Map.Overrides
	if ov.TargetKillCount > 0 { g.currentMapType.TargetKillCount = ov.TargetKillCount }
	if ov.TargetTime > 0 { g.currentMapType.TargetTime = ov.TargetTime }
	if ov.Difficulty > 0 { g.currentMapType.Difficulty = ov.Difficulty }
	if len(ov.TargetKills) > 0 { g.currentMapType.TargetKills = ov.TargetKills }
	if ov.Name != "" { g.currentMapType.Name = ov.Name }
	if ov.Description != "" { g.currentMapType.Description = ov.Description }
	if data.Map.FloorTile != "" { g.currentMapType.FloorTile = data.Map.FloorTile }
	if len(data.Map.FloorZones) > 0 { g.currentMapType.FloorZones = data.Map.FloorZones }

	if data.Map.Heightmap != nil {
		g.currentMapType.Heightmap = data.Map.Heightmap
	} else if g.currentMapType.Heightmap == nil {
		g.currentMapType.Heightmap = make(map[string]float64)
	}

	g.ExploredTiles = make(map[image.Point]bool)
	for _, pt := range data.Map.ExploredTiles {
		g.ExploredTiles[pt] = true
	}

	if data.Player.Archetype != "" {
		if config, ok := g.characterRegistry.Characters[data.Player.Archetype]; ok {
			g.playableCharacter.Config = config
			g.playableCharacter.Name = config.Name
			g.isCharacterSelect = false
			g.isMainMenu = false
		} else {
			log.Printf("Warning: saved character archetype ID %s not found in registry", data.Player.Archetype)
		}
	}

	g.playableCharacter.X = data.Player.X
	g.playableCharacter.Y = data.Player.Y
	g.playableCharacter.State.HealthPoints = data.Player.HealthPoints
	g.playableCharacter.XP = data.Player.XP
	g.playableCharacter.Level = data.Player.Level
	g.playableCharacter.Kills = data.Player.Kills
	g.playableCharacter.SelectedModel = data.Player.SelectedModel
	
	// Load inventory
	g.playableCharacter.Inventory = nil
	for _, itemData := range data.Player.Inventory {
		if config, ok := g.Registries.Objects.Objects[itemData.ID]; ok {
			instance := NewItemInstance(itemData.ID, config, g.playableCharacter.X, g.playableCharacter.Y)
			if itemData.Resistance > 0 {
				instance.Resistance = itemData.Resistance
			}
			g.playableCharacter.Inventory = append(g.playableCharacter.Inventory, instance)
		}
	}
	// Load slots
	g.playableCharacter.Slots = make(map[string]*ItemInstance)
	for slot, itemData := range data.Player.Slots {
		if config, ok := g.Registries.Objects.Objects[itemData.ID]; ok {
			instance := NewItemInstance(itemData.ID, config, g.playableCharacter.X, g.playableCharacter.Y)
			if itemData.Resistance > 0 {
				instance.Resistance = itemData.Resistance
			}
			g.playableCharacter.Slots[slot] = instance
		}
	}
	g.playableCharacter.UpdateEffects()

	g.playableCharacter.State.MaxHealthPoints = data.Player.MaxHealthPoints
	g.playableCharacter.BaseAttack = data.Player.BaseAttack
	g.playableCharacter.BaseDefense = data.Player.BaseDefense
	g.playableCharacter.BaseProtection = data.Player.BaseProtection
	g.playableCharacter.PrimaryAttributes.Strength = data.Player.Strength
	g.playableCharacter.PrimaryAttributes.Dexterity = data.Player.Dexterity
	g.playableCharacter.PrimaryAttributes.Health = data.Player.Health
	g.playableCharacter.PrimaryAttributes.Intellect = data.Player.Intellect
	g.playableCharacter.PrimaryAttributes.Wisdom = data.Player.Wisdom
	g.playableCharacter.Submission = data.Player.Submission
	if g.playableCharacter.Submission == nil { g.playableCharacter.Submission = make(map[string]float64) }
	
	// Hunger, Thirst, Fatigue (with fallback for old saves)
	g.playableCharacter.State.Hunger = data.Player.Hunger
	g.playableCharacter.State.Thirst = data.Player.Thirst
	g.playableCharacter.State.Fatigue = data.Player.Fatigue
	if g.playableCharacter.State.Hunger == 0 && g.playableCharacter.State.Thirst == 0 && g.playableCharacter.State.Fatigue == 0 {
		g.playableCharacter.State.Hunger = 0
		g.playableCharacter.State.Thirst = 0
		g.playableCharacter.State.Fatigue = 0
	}

	if data.Player.Weapon != nil {
		g.playableCharacter.Weapon = data.Player.Weapon
	}
	g.playableCharacter.Denarii = data.Player.Denarii


	// Load floor items
	g.World.Items = nil
	for _, itemData := range data.Items {
		if config, ok := g.Registries.Objects.Objects[itemData.ID]; ok {
			instance := NewItemInstance(itemData.ID, config, itemData.X, itemData.Y)
			if itemData.Resistance > 0 {
				instance.Resistance = itemData.Resistance
			}
			g.World.Items = append(g.World.Items, instance)
		}
	}

	g.playableCharacter.MapKills = data.Player.MapKills
	if g.playableCharacter.MapKills == nil {
		g.playableCharacter.MapKills = make(map[string]int)
	}


	if g.playableCharacter.State.HealthPoints > 0 {
		g.playableCharacter.ActionState = ActorIdle
		g.isGameOver = false
	} else {
		g.playableCharacter.ActionState = ActorDead
		g.isGameOver = true
	}

	g.mapWonMenuIndex = WinMenuContinue
	pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
	g.camera.SnapTo(pIsoX, pIsoY)

	g.characters = nil
	for i, nData := range data.Characters {
		sanitizeNPCSaveData(&nData, i, fpath)
		id := nData.Archetype
		if id == "" { id = nData.NPCID }

		config, ok := g.archetypeRegistry.Archetypes[id]
		if !ok { config, ok = g.characterRegistry.Characters[id] }

		if !ok {
			log.Printf("Warning: saved NPC/Archetype ID %s not found in any registry", nData.Archetype)
			continue
		}
		n := NewCharacter(nData.X, nData.Y, config, nData.Level, false, g.Registries.Objects)
		n.State.HealthPoints = nData.HealthPoints
		n.State.MaxHealthPoints = nData.MaxHealthPoints
		if nData.Name != "" { n.Name = nData.Name }
		if nData.BaseAttack > 0 { n.BaseAttack = nData.BaseAttack }
		if nData.BaseDefense > 0 { n.BaseDefense = nData.BaseDefense }
		if nData.BaseProtection > 0 { n.BaseProtection = nData.BaseProtection }
		if nData.Strength > 0 { n.PrimaryAttributes.Strength = nData.Strength }
		if nData.Dexterity > 0 { n.PrimaryAttributes.Dexterity = nData.Dexterity }
		if nData.Health > 0 { n.PrimaryAttributes.Health = nData.Health }
		if nData.Intellect > 0 { n.PrimaryAttributes.Intellect = nData.Intellect }
		if nData.Wisdom > 0 { n.PrimaryAttributes.Wisdom = nData.Wisdom }
		n.Submission = nData.Submission
		if n.Submission == nil { n.Submission = make(map[string]float64) }
		
		n.State.Hunger = nData.Hunger
		n.State.Thirst = nData.Thirst
		n.State.Fatigue = nData.Fatigue
		if n.State.Hunger == 0 && n.State.Thirst == 0 && n.State.Fatigue == 0 {
			n.State.Hunger = 0
			n.State.Thirst = 0
			n.State.Fatigue = 0
		}

		n.Denarii = nData.Denarii
		n.SelectedModel = nData.SelectedModel

		switch nData.Behavior {
		case "wander": n.Behavior = BehaviorWander
		case "patrol": n.Behavior = BehaviorPatrol
		case "hunter": n.Behavior = BehaviorKnightHunter
		case "fighter": n.Behavior = BehaviorNpcFighter
		case "chaotic": n.Behavior = BehaviorChaotic
		case "escort": n.Behavior = BehaviorEscort
		case "trader": n.Behavior = BehaviorTrader
		case "hauler": n.Behavior = BehaviorHauler
		case "lumberjack": n.Behavior = BehaviorLumberjack
		}

		if nData.Alignment != 0 { n.Alignment = nData.Alignment }
		if nData.Group != "" { n.Group = nData.Group }
		if nData.LeaderID != "" { n.LeaderID = nData.LeaderID }
		if nData.MustSurvive { n.MustSurvive = nData.MustSurvive }

		// Load inventory and slots
		n.Inventory = nil
		for _, itemData := range nData.Inventory {
			if config, ok := g.Registries.Objects.Objects[itemData.ID]; ok {
				instance := NewItemInstance(itemData.ID, config, n.X, n.Y)
				if itemData.Resistance > 0 {
					instance.Resistance = itemData.Resistance
				}
				n.Inventory = append(n.Inventory, instance)
			}
		}
		n.Slots = make(map[string]*ItemInstance)
		for slot, itemData := range nData.Slots {
			if config, ok := g.Registries.Objects.Objects[itemData.ID]; ok {
				instance := NewItemInstance(itemData.ID, config, n.X, n.Y)
				if itemData.Resistance > 0 {
					instance.Resistance = itemData.Resistance
				}
				n.Slots[slot] = instance
			}
		}
		
		if n.State.HealthPoints <= 0 { n.ActionState = ActorDead }
		n.UpdateEffects()
		g.characters = append(g.characters, n)
	}

	g.obstacles = nil
	preSpawns := make(map[string]PreSpawnObstacle)
	for _, ps := range g.currentMapType.Obstacles {
		if ps.ID != "" { preSpawns[ps.ID] = ps }
	}
	handledPreSpawns := make(map[string]bool)

	for _, oData := range data.Obstacles {
		var base *PreSpawnObstacle
		if oData.ID != "" {
			if ps, ok := preSpawns[oData.ID]; ok {
				base = &ps
				handledPreSpawns[oData.ID] = true
			}
		}
		if oData.Disabled { continue }

		archID := oData.Archetype
		if archID == "" && base != nil { archID = base.Archetype }

		config, ok := g.obstacleRegistry.Archetypes[archID]
		if !ok { continue }

		px, py := 0.0, 0.0
		if oData.X != nil { px = *oData.X } else if base != nil && base.X != nil { px = *base.X }
		if oData.Y != nil { py = *oData.Y } else if base != nil && base.Y != nil { py = *base.Y }

		o := NewObstacle(oData.ID, px, py, config)
		if oData.HealthPoints > 0 || oData.X != nil { o.HealthPoints = oData.HealthPoints } else if config.HealthPoints > 0 { o.HealthPoints = config.HealthPoints }
		o.CooldownTicks = oData.CooldownTicks
		if o.HealthPoints <= 0 && config.HealthPoints > 0 { o.Alive = false }
		g.obstacles = append(g.obstacles, o)
	}

	for _, ps := range g.currentMapType.Obstacles {
		if !handledPreSpawns[ps.ID] && !ps.Disabled {
			if config, ok := g.obstacleRegistry.Archetypes[ps.Archetype]; ok {
				px, py := 0.0, 0.0
				if ps.X != nil { px = *ps.X }
				if ps.Y != nil { py = *ps.Y }
				g.obstacles = append(g.obstacles, NewObstacle(ps.ID, px, py, config))
			}
		}
	}
	return nil
}
