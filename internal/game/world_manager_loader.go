package game

import (
	"fmt"
	"image"
	"io/fs"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
	"oinakos/internal/engine"
)

func (wm *WorldManager) LoadMapAssets() {
	startTime := time.Now()
	g := wm.game
	if g.assets == nil { return }
	
	ls := &LoadingState{Progress: &g.LoadingProgress}
	characterFilter := make(map[string]bool)
	obstacleFilter := make(map[string]bool)
	configs := make(map[string]*EntityConfig)

	addConfig := func(conf *EntityConfig) {
		if conf == nil { return }
		configs[conf.ID] = conf
		characterFilter[conf.ID] = true
		if conf.Archetype != "" {
			if arch, ok := g.archetypeRegistry.Archetypes[conf.Archetype]; ok { configs[arch.ID] = arch }
		}
	}

	if g.playableCharacter.Config != nil { addConfig(g.playableCharacter.Config) }
	for _, n := range g.characters { if n.Config != nil { addConfig(n.Config) } }
	for _, s := range g.currentMapType.Spawns { if arch, ok := g.archetypeRegistry.Archetypes[s.Archetype]; ok { addConfig(arch) } }
	for _, o := range g.obstacles { obstacleFilter[o.Archetype.ID] = true }

	var audioJobs []*AudioLoadJob
	for _, conf := range configs {
		if conf.AudioDir == "" { continue }
		entries, err := fs.ReadDir(g.assets, conf.AudioDir)
		if err != nil { continue }
		for _, e := range entries {
			if e.IsDir() { continue }
			lowerName := strings.ToLower(e.Name())
			if !strings.HasSuffix(lowerName, ".mp3") && !strings.HasSuffix(lowerName, ".wav") { continue }
			key := conf.ID + "/" + e.Name()[:len(e.Name())-4]
			if engine.GlobalAudio != nil && engine.GlobalAudio.HasSound(key) { continue }
			audioJobs = append(audioJobs, &AudioLoadJob{Name: key, Path: conf.AudioDir + "/" + e.Name()})
		}
	}

	charImgCount := g.characterRegistry.CountAssets(g.assets, g.archetypeRegistry, characterFilter)
	obsImgCount := g.obstacleRegistry.CountAssets(obstacleFilter)
	ls.Total = int32((len(audioJobs) + charImgCount + obsImgCount) * 2)
	atomic.StoreInt32(&g.LoadingProgress, 0)
	if len(audioJobs) > 0 { loadAudioParallel(g.assets, audioJobs, ls) }
	g.characterRegistry.LoadAssets(g.assets, g.Graphics, g.archetypeRegistry, characterFilter, ls)
	g.obstacleRegistry.LoadAssets(g.assets, g.Graphics, obstacleFilter, ls)
	log.Printf("LoadMapAssets: processed %d assets in %v", len(audioJobs), time.Since(startTime))
}

func (wm *WorldManager) LoadMapLevel() {
	g := wm.game
	if atomic.LoadInt32(&g.LoadingProgress) < 1000 { return }
	atomic.StoreInt32(&g.LoadingProgress, 0)
	g.LoadingMessage = "Loading Map..."
	if g.isCampaign { g.LoadingMessage = "Loading Campaign..." }
	startTime := time.Now()
	defer func() {
		atomic.StoreInt32(&g.LoadingProgress, 1000)
		log.Printf("LoadMapLevel: complete in %v", time.Since(startTime))
	}()

	if g.isCampaign && g.currentCampaign != nil && g.campaignIndex < len(g.currentCampaign.Maps) {
		mapID := g.currentCampaign.Maps[g.campaignIndex]
		if m, ok := g.mapTypeRegistry.Types[mapID]; ok { g.currentMapType = *m }
	}
	if g.initialMapID != "" && g.mapLevel == 1 && !g.isCampaign {
		if m, ok := g.mapTypeRegistry.Types[g.initialMapID]; ok { g.currentMapType = *m }
	}
	if g.currentMapType.ID == "" && len(g.mapTypeRegistry.IDs) > 0 {
		g.currentMapType = *g.mapTypeRegistry.Types[g.mapTypeRegistry.IDs[0]]
	}
	g.currentMapType.SeedMinerals(int64(g.mapLevel) + time.Now().UnixNano())
	g.playTime = 0; g.characters = make([]*Character, 0); g.obstacles = make([]*Obstacle, 0); g.floatingTexts = make([]*FloatingText, 0)
	
	weatherStr := strings.ToLower(g.currentMapType.Weather)
	if weatherStr == "random" {
		r := rand.Float64()
		if r < 0.2 { weatherStr = "rain" } else if r < 0.3 { weatherStr = "snow" } else if r < 0.35 { weatherStr = "storm" } else { weatherStr = "clear" }
	}
	g.World.State.Weather = WeatherClear; g.World.State.Intensity = 0.5
	switch weatherStr {
	case "rain": g.World.State.Weather = WeatherRain
	case "snow": g.World.State.Weather = WeatherSnow
		if g.currentMapType.FloorTile != "snow.png" && g.currentMapType.FloorTile != "snow_2.png" {
			if rand.Float64() < 0.5 { g.currentMapType.FloorTile = "snow.png" } else { g.currentMapType.FloorTile = "snow_2.png" }
		}
	case "storm": g.World.State.Weather = WeatherStorm
	}
	g.particles = nil
	
	if g.World != nil {
		g.World.Items = make([]*ItemInstance, 0)
		if g.Registries != nil && g.Registries.Objects != nil {
			spawnedUnique := make(map[string]bool)
			for _, o := range g.currentMapType.Objects {
				if cfg, ok := g.Registries.Objects.Objects[o.ID]; ok {
					if cfg.Unique && spawnedUnique[cfg.ID] { continue }
					safeX, safeY := findSafePosition(o.X, o.Y, engine.Circle{Radius: 0.5}, g.obstacles)
					it := &ItemInstance{ID: fmt.Sprintf("%s_%d", cfg.ID, rand.Int()), Config: cfg, Resistance: cfg.Resistance, X: safeX, Y: safeY, Pickable: true}
					g.World.Items = append(g.World.Items, it)
					if cfg.Unique { spawnedUnique[cfg.ID] = true }
				}
			}
		}
	}
	g.playableCharacter.MapKills = make(map[string]int); g.mapWonMenuIndex = 0; g.ExploredTiles = make(map[image.Point]bool)
	if g.currentMapType.Player != nil { g.playableCharacter.X, g.playableCharacter.Y = g.currentMapType.Player.X, g.currentMapType.Player.Y }
	g.playableCharacter.LoadEquipment(g.Registries.Objects)
	pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y); g.camera.SnapTo(pIsoX, pIsoY)
	g.currentMapType.TargetTime *= float64(g.mapLevel)
	newKills := make(map[string]int)
	for npcID, target := range g.currentMapType.TargetKills { newKills[npcID] = target * g.mapLevel }
	g.currentMapType.TargetKills = newKills

	switch g.currentMapType.Type {
	case ObjKillVIP:
		if len(g.archetypeRegistry.IDs) > 0 {
			vipID := g.archetypeRegistry.IDs[rand.Intn(len(g.archetypeRegistry.IDs))]
			vipConfig := g.archetypeRegistry.Archetypes[vipID]
			tpX, tpY := g.playableCharacter.X + (rand.Float64()*40 - 20), g.playableCharacter.Y + (rand.Float64()*40 - 20)
			vip := NewCharacter(tpX, tpY, vipConfig, g.mapLevel*2, false, g.Registries.Objects)
			vip.IsTarget = true; g.characters = append(g.characters, vip); g.currentMapType.TargetPoint = engine.Point{X: tpX, Y: tpY}
		}
	case ObjReachPortal, ObjReachZone, ObjReachBuilding:
		if g.currentMapType.TargetPointRaw != nil { g.currentMapType.TargetPoint = engine.Point{X: g.currentMapType.TargetPointRaw.X, Y: g.currentMapType.TargetPointRaw.Y}
		} else { g.currentMapType.TargetPoint = engine.Point{X: g.playableCharacter.X + (rand.Float64()*80-40), Y: g.playableCharacter.Y + (rand.Float64()*80-40)} }
		if g.currentMapType.Type == ObjReachBuilding { if cfg, ok := g.obstacleRegistry.Archetypes["warehouse"]; ok { g.obstacles = append(g.obstacles, NewObstacle("target_warehouse", g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y, cfg)) } }
	case ObjProtectNPC:
		if cfg, ok := g.archetypeRegistry.Archetypes["magi_male"]; ok {
			escort := NewCharacter(g.playableCharacter.X+2, g.playableCharacter.Y+2, cfg, g.mapLevel, false, g.Registries.Objects)
			escort.MustSurvive = true; g.characters = append([]*Character{escort}, g.characters...)
		}
	case ObjDestroyBuilding:
		if cfg, ok := g.obstacleRegistry.Archetypes["warehouse"]; ok {
			targetObs := NewObstacle("target_building", g.playableCharacter.X+40, g.playableCharacter.Y+40, cfg)
			g.obstacles = append(g.obstacles, targetObs); g.currentMapType.TargetObstacle = targetObs
		}
	case ObjSimulation:
		wg := NewWorldGenerator(g, time.Now().UnixNano())
		wg.GenerateVillage(g.playableCharacter.X+30, g.playableCharacter.Y+30, 15)
	}

	allInhabs := append(append(g.currentMapType.Inhabitants, g.currentMapType.Fauna...), g.currentMapType.Characters...)
	for _, ps := range allInhabs {
		var config *EntityConfig; var ok bool
		id := ps.NPC; if ps.NPCID != "" { id = ps.NPCID }
		if id != "" { config, ok = g.characterRegistry.Characters[id] } else if ps.Archetype != "" { config, ok = g.archetypeRegistry.Archetypes[ps.Archetype] }
		
		if !ok && IsDebugEnabled() { DebugLog("NPC-LOAD-FAIL: Could not find config for %s/%s (Archetype:%s, NPC:%s)", ps.Name, ps.ID, ps.Archetype, ps.NPC) }

		if ok {
			npc := NewCharacter(ps.X, ps.Y, config, g.mapLevel, false, g.Registries.Objects)
			npc.Alignment, npc.MustSurvive, npc.IsTarget = ps.Alignment, ps.MustSurvive, ps.IsTarget
			if ps.Name != "" { npc.Name = ps.Name }
			if ps.State == "dead" { npc.ActionState = ActorDead }
			switch ps.Behavior {
			case "wander": npc.Behavior = BehaviorWander
			case "patrol": npc.Behavior = BehaviorPatrol
			case "hunter": npc.Behavior = BehaviorKnightHunter
			case "fighter": npc.Behavior = BehaviorNpcFighter
			case "chaotic": npc.Behavior = BehaviorChaotic
			case "escort": npc.Behavior = BehaviorEscort
			case "trader": npc.Behavior = BehaviorTrader
			case "hauler": npc.Behavior = BehaviorHauler
			case "lumberjack": npc.Behavior = BehaviorLumberjack
			case "behavior_criminal": npc.Behavior = BehaviorCriminal
			}
			g.characters = append(g.characters, npc)
			if IsDebugEnabled() && len(g.characters) < 10 { DebugLog("NPC-LOADED: %s (%s)", npc.Name, ps.Archetype) }
		}
	}

	for i, po := range g.currentMapType.Obstacles {
		if po.Disabled { continue }
		if config, ok := g.obstacleRegistry.Archetypes[po.Archetype]; ok {
			px, py := 0.0, 0.0; if po.X != nil { px = *po.X }; if po.Y != nil { py = *po.Y }
			id := po.ID
			if id == "" { id = fmt.Sprintf("%s_%d_%d", po.Archetype, int(px), int(py)) }
			g.obstacles = append(g.obstacles, NewObstacle(id, px, py, config))
		}
		if i%10 == 0 { runtime.Gosched() }
	}

	for _, n := range g.characters {
		for i := 0; i < 50; i++ { if !n.checkCollisionAt(n.X, n.Y, g.obstacles) { break }; n.X += 0.5; n.Y += 0.5 }
	}
	for i := 0; i < 50; i++ { if !g.playableCharacter.checkCollisionAt(g.playableCharacter.X, g.playableCharacter.Y, g.obstacles) { break }; g.playableCharacter.X += 0.5; g.playableCharacter.Y += 0.5 }

	g.characters = append(g.characters, g.playableCharacter)
	
	if g.World != nil {
		g.World.Characters = g.characters
		g.World.Obstacles = g.obstacles
	}

	DebugLog("Map Load Complete: %s (Pop: %d, Obs: %d)", g.currentMapType.Name, len(g.characters), len(g.obstacles))
	wm.LoadMapAssets()
}
