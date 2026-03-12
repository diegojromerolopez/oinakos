package game

import (
	"image"
	"io/fs"
	"log"
	"math"
	"math/rand"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"oinakos/internal/engine"
)

type WorldManager struct {
	game *Game
}

func NewWorldManager(g *Game) *WorldManager {
	return &WorldManager{game: g}
}

func (wm *WorldManager) LoadMapAudio() {
	startTime := time.Now()
	g := wm.game
	configs := make(map[string]*EntityConfig)

	collect := func(id string, isNPC bool) {
		if id == "" {
			return
		}
		if isNPC {
			if conf, ok := g.npcRegistry.NPCs[id]; ok {
				configs[conf.ID] = conf
				if conf.ArchetypeID != "" {
					if arch, ok := g.archetypeRegistry.Archetypes[conf.ArchetypeID]; ok {
						configs[arch.ID] = arch
					}
				}
			}
		}
		if arch, ok := g.archetypeRegistry.Archetypes[id]; ok {
			configs[arch.ID] = arch
		}
	}

	if g.playableCharacter.Config != nil {
		configs[g.playableCharacter.Config.ID] = g.playableCharacter.Config
	}

	allInhabs := append(g.currentMapType.Inhabitants, g.currentMapType.NPCs...)
	for _, ps := range allInhabs {
		if ps.NPCID != "" {
			collect(ps.NPCID, true)
		} else if ps.NPC != "" {
			collect(ps.NPC, true)
		} else if ps.ArchetypeID != "" {
			collect(ps.ArchetypeID, false)
		} else if ps.Archetype != "" {
			collect(ps.Archetype, false)
		}
	}

	for _, s := range g.currentMapType.Spawns {
		collect(s.Archetype, false)
	}

	for _, n := range g.npcs {
		if n.Config != nil {
			configs[n.Config.ID] = n.Config
		}
	}

	var jobs []*AudioLoadJob
	for _, conf := range configs {
		if conf.AudioDir == "" {
			continue
		}
		entries, err := fs.ReadDir(g.assets, conf.AudioDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			lowerName := strings.ToLower(e.Name())
			if !strings.HasSuffix(lowerName, ".mp3") && !strings.HasSuffix(lowerName, ".wav") {
				continue
			}
			extLen := 4
			if strings.HasSuffix(lowerName, ".mp3") {
				extLen = 4
			}
			stem := e.Name()[:len(e.Name())-extLen]
			key := conf.ID + "/" + stem
			if engine.GlobalAudio != nil && engine.GlobalAudio.HasSound(key) {
				continue
			}
			jobs = append(jobs, &AudioLoadJob{
				Name: key,
				Path: conf.AudioDir + "/" + e.Name(),
			})
		}
	}

	if len(jobs) > 0 {
		DebugLog("Parallel Loading %d audio files for map...", len(jobs))
		loadAudioParallel(g.assets, jobs, &g.LoadingProgress)
	}
	log.Printf("LoadMapAudio: processed %d jobs in %v", len(jobs), time.Since(startTime))
}

func (wm *WorldManager) LoadMapLevel() {
	g := wm.game
	if atomic.LoadInt32(&g.LoadingProgress) < 1000 {
		return // Already loading
	}
	atomic.StoreInt32(&g.LoadingProgress, 0)
	g.LoadingMessage = "Loading Map..."
	if g.isCampaign {
		g.LoadingMessage = "Loading Campaign..."
	}
	log.Printf("Starting Async Map Load: %s", g.LoadingMessage)

	startTime := time.Now()
	go func() {
		defer func() {
			atomic.StoreInt32(&g.LoadingProgress, 1000)
			log.Printf("Async Map Load Complete. Total Time: %v", time.Since(startTime))
		}()

		if g.isCampaign && g.currentCampaign != nil {
		if g.campaignIndex < len(g.currentCampaign.Maps) {
			mapID := g.currentCampaign.Maps[g.campaignIndex]
			if m, ok := g.mapTypeRegistry.Types[mapID]; ok {
				g.currentMapType = *m
				log.Printf("Loading Campaign Map [%d/%d]: %s", g.campaignIndex+1, len(g.currentCampaign.Maps), mapID)
			}
		}
	}

	if g.initialMapID != "" && g.mapLevel == 1 && !g.isCampaign {
		if m, ok := g.mapTypeRegistry.Types[g.initialMapID]; ok {
			g.currentMapType = *m
			log.Printf("Using initial map selection: %s", g.initialMapID)
		}
	}

	if g.currentMapType.ID == "" {
		if len(g.mapTypeRegistry.IDs) > 0 {
			g.currentMapType = *g.mapTypeRegistry.Types[g.mapTypeRegistry.IDs[0]]
		}
	}

	g.playTime = 0
	g.npcs = make([]*NPC, 0)
	g.obstacles = make([]*Obstacle, 0)
	g.floatingTexts = make([]*FloatingText, 0)
	g.currentMapType.StartTime = 0
	g.playableCharacter.MapKills = make(map[string]int)
	g.mapWonMenuIndex = 0
	g.ExploredTiles = make(map[image.Point]bool)

	if g.currentMapType.Player != nil {
		g.playableCharacter.X = g.currentMapType.Player.X
		g.playableCharacter.Y = g.currentMapType.Player.Y
	}

	pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
	g.camera.SnapTo(pIsoX, pIsoY)

	g.currentMapType.TargetTime *= float64(g.mapLevel)
	newKills := make(map[string]int)
	for npcID, target := range g.currentMapType.TargetKills {
		newKills[npcID] = target * g.mapLevel
	}
	g.currentMapType.TargetKills = newKills

	switch g.currentMapType.Type {
	case ObjKillVIP:
		if len(g.archetypeRegistry.IDs) > 0 {
			vipID := g.archetypeRegistry.IDs[rand.Intn(len(g.archetypeRegistry.IDs))]
			vipConfig := g.archetypeRegistry.Archetypes[vipID]
			tpX := g.playableCharacter.X + (rand.Float64()*40 - 20)
			tpY := g.playableCharacter.Y + (rand.Float64()*40 - 20)
			if tpX > -5 && tpX < 5 { tpX += 10 }
			if tpY > -5 && tpY < 5 { tpY += 10 }
			vip := NewNPC(tpX, tpY, vipConfig, g.mapLevel*2)
			vip.MaxHealth *= g.mapLevel * 2
			vip.Health = vip.MaxHealth
			vip.BaseAttack += g.mapLevel * 2
			g.npcs = append(g.npcs, vip)
			g.currentMapType.TargetPoint = engine.Point{X: tpX, Y: tpY}
		}
	case ObjReachPortal, ObjReachZone:
		if g.currentMapType.TargetPointRaw != nil {
			g.currentMapType.TargetPoint = engine.Point{X: g.currentMapType.TargetPointRaw.X, Y: g.currentMapType.TargetPointRaw.Y}
		} else {
			offX := rand.Float64()*60 - 30
			offY := rand.Float64()*60 - 30
			if offX > -10 && offX < 10 { offX += 20 }
			if offY > -10 && offY < 10 { offY += 20 }
			g.currentMapType.TargetPoint = engine.Point{X: g.playableCharacter.X + offX, Y: g.playableCharacter.Y + offY}
		}
	case ObjReachBuilding:
		if g.currentMapType.TargetPointRaw != nil {
			g.currentMapType.TargetPoint = engine.Point{X: g.currentMapType.TargetPointRaw.X, Y: g.currentMapType.TargetPointRaw.Y}
		} else {
			offX := rand.Float64()*50 - 25
			offY := rand.Float64()*50 - 25
			if offX > -10 && offX < 10 { offX += 20 }
			if offY > -10 && offY < 10 { offY += 20 }
			g.currentMapType.TargetPoint = engine.Point{X: g.playableCharacter.X + offX, Y: g.playableCharacter.Y + offY}
		}
		if config, ok := g.obstacleRegistry.Archetypes["warehouse"]; ok {
			g.obstacles = append(g.obstacles, NewObstacle("target_warehouse", g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y, config))
		}
	case ObjProtectNPC:
		if g.currentMapType.TargetPointRaw != nil {
			g.currentMapType.TargetPoint = engine.Point{X: g.currentMapType.TargetPointRaw.X, Y: g.currentMapType.TargetPointRaw.Y}
		} else {
			offX := rand.Float64()*80 - 40
			offY := rand.Float64()*80 - 40
			if offX > -20 && offX < 20 { offX += 40 }
			if offY > -20 && offY < 20 { offY += 40 }
			g.currentMapType.TargetPoint = engine.Point{X: g.playableCharacter.X + offX, Y: g.playableCharacter.Y + offY}
		}
		if config, ok := g.archetypeRegistry.Archetypes["magi_male"]; ok {
			escort := NewNPC(g.playableCharacter.X+2, g.playableCharacter.Y+2, config, g.mapLevel)
			g.npcs = append([]*NPC{escort}, g.npcs...)
		}
	case ObjDestroyBuilding:
		var targetObs *Obstacle
		for _, o := range g.obstacles {
			if o.ID == "target_building" {
				targetObs = o
				break
			}
		}
		if targetObs != nil {
			g.currentMapType.TargetObstacle = targetObs
			g.currentMapType.TargetPoint = engine.Point{X: targetObs.X, Y: targetObs.Y}
		} else {
			if g.currentMapType.TargetPointRaw != nil {
				g.currentMapType.TargetPoint = engine.Point{X: g.currentMapType.TargetPointRaw.X, Y: g.currentMapType.TargetPointRaw.Y}
			} else {
				g.currentMapType.TargetPoint = engine.Point{
					X: g.playableCharacter.X + (rand.Float64()*80 - 40),
					Y: g.playableCharacter.Y + (rand.Float64()*80 - 40),
				}
				if g.currentMapType.TargetPoint.X > -20 && g.currentMapType.TargetPoint.X < 20 { g.currentMapType.TargetPoint.X += 40 }
				if g.currentMapType.TargetPoint.Y > -20 && g.currentMapType.TargetPoint.Y < 20 { g.currentMapType.TargetPoint.Y += 40 }
			}
			if config, ok := g.obstacleRegistry.Archetypes["warehouse"]; ok {
				targetObs = NewObstacle("target_building", g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y, config)
				g.obstacles = append(g.obstacles, targetObs)
				g.currentMapType.TargetObstacle = targetObs
			}
		}
	}

	allInhabs := append(g.currentMapType.Inhabitants, g.currentMapType.NPCs...)
	for _, ps := range allInhabs {
		var config *EntityConfig
		var ok bool
		id := ps.NPC
		if ps.NPCID != "" { id = ps.NPCID }
		if id != "" { config, ok = g.npcRegistry.NPCs[id] } else {
			arch := ps.Archetype
			if ps.ArchetypeID != "" { arch = ps.ArchetypeID }
			if arch != "" { config, ok = g.archetypeRegistry.Archetypes[arch]; id = arch }
		}
		if ok {
			if id != "" && g.playableCharacter.Config != nil && id == g.playableCharacter.Config.ID {
				g.playableCharacter.X, g.playableCharacter.Y = ps.X, ps.Y
				continue
			}
			npc := NewNPC(ps.X, ps.Y, config, g.mapLevel)
			npc.Alignment, npc.MustSurvive = ps.Alignment, ps.MustSurvive
			if ps.Name != "" { npc.Name = ps.Name }
			if ps.State == "dead" { npc.Health, npc.State = 0, NPCDead }
			g.npcs = append(g.npcs, npc)
		}
	}

	for i, po := range g.currentMapType.Obstacles {
		if po.Disabled {
			continue
		}
		arch := po.Archetype
		if po.ArchetypeID != "" {
			arch = po.ArchetypeID
		}
		if config, ok := g.obstacleRegistry.Archetypes[arch]; ok {
			px, py := 0.0, 0.0
			if po.X != nil {
				px = *po.X
			}
			if po.Y != nil {
				py = *po.Y
			}
			g.obstacles = append(g.obstacles, NewObstacle(po.ID, px, py, config))
		}
		if i%10 == 0 {
			runtime.Gosched()
		}
	}

	for _, n := range g.npcs {
		if !n.IsAlive() {
			continue
		}
		for i := 0; i < 500; i++ {
			if !n.checkCollisionAt(n.X, n.Y, g.obstacles) {
				break
			}
			n.X += 0.5
			n.Y += 0.5
			if i%50 == 0 {
				runtime.Gosched()
			}
		}
	}

	for i := 0; i < 500; i++ {
		if !g.playableCharacter.checkCollisionAt(g.playableCharacter.X, g.playableCharacter.Y, g.obstacles) { break }
		angle := float64(i) * 0.3; dist := 1.0 + (float64(i) * 0.2)
		g.playableCharacter.X += math.Cos(angle) * dist
		g.playableCharacter.Y += math.Sin(angle) * dist
		if i%25 == 0 {
			runtime.Gosched()
		}
	}

		DebugLog("Starting Map Level %d: %s at safe pos %.2f,%.2f", g.mapLevel, g.currentMapType.Name, g.playableCharacter.X, g.playableCharacter.Y)
		wm.LoadMapAudio()
	}()
}

func (wm *WorldManager) UpdateChunks() {
	// Procedural spawning disabled
}

func (wm *WorldManager) UpdateNPCSpawning() {
	g := wm.game
	if len(g.currentMapType.Spawns) > 0 {
		for i := range g.currentMapType.Spawns {
			s := &g.currentMapType.Spawns[i]
			if s.Frequency <= 0 { continue }
			s.Timer++
			if s.Timer >= int(s.Frequency*60) {
				s.Timer = 0
				if rand.Float64() <= s.Probability {
					if len(g.npcs) < 100 {
						if s.X != nil && s.Y != nil { wm.spawnNPCNearPosition(*s.X, *s.Y, s) } else { wm.spawnNPCAtMapEdges(s) }
					}
				}
			}
		}
	}

	g.npcSpawnTimer++
	if g.npcSpawnTimer >= 300 {
		g.npcSpawnTimer = 0
		activeNPCs := make([]*NPC, 0)
		for _, n := range g.npcs {
			dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
			if n.IsAlive() {
				if dist < 40 { activeNPCs = append(activeNPCs, n) }
			} else { activeNPCs = append(activeNPCs, n) }
		}
		g.npcs = activeNPCs
	}
}

func (wm *WorldManager) spawnNPCNearPosition(x, y float64, sc *SpawnConfig) {
	g := wm.game
	if len(g.archetypeRegistry.IDs) == 0 || sc == nil { return }
	npcConfig := g.archetypeRegistry.Archetypes[sc.Archetype]
	if npcConfig == nil { return }
	if rand.Float64() < 0.05 {
		var variants []*EntityConfig
		for _, v := range g.npcRegistry.NPCs {
			if v.ArchetypeID == sc.Archetype && !v.Unique { variants = append(variants, v) }
		}
		if len(variants) > 0 { npcConfig = variants[rand.Intn(len(variants))] }
	}
	npc := NewNPC(x, y, npcConfig, g.mapLevel)
	npc.Alignment = sc.Alignment
	for i := 0; i < 10; i++ {
		collides := false
		for _, o := range g.obstacles {
			if o.Alive && engine.CheckCollision(npc.GetFootprint(), o.GetFootprint()) { collides = true; break }
		}
		if !collides { break }
		angle := rand.Float64() * 2 * math.Pi
		npc.X = x + math.Cos(angle)*(2.0+rand.Float64()); npc.Y = y + math.Sin(angle)*(2.0+rand.Float64())
	}
	g.npcs = append(g.npcs, npc)
}

func (wm *WorldManager) spawnNPCAtMapEdges(sc *SpawnConfig) {
	g := wm.game
	if len(g.archetypeRegistry.IDs) == 0 || sc == nil { return }
	npcConfig := g.archetypeRegistry.Archetypes[sc.Archetype]
	if npcConfig == nil { return }
	if rand.Float64() < 0.05 {
		var variants []*EntityConfig
		for _, v := range g.npcRegistry.NPCs {
			if v.ArchetypeID == sc.Archetype && !v.Unique { variants = append(variants, v) }
		}
		if len(variants) > 0 { npcConfig = variants[rand.Intn(len(variants))] }
	}
	angle := rand.Float64() * 2 * math.Pi
	npc := NewNPC(g.playableCharacter.X+math.Cos(angle)*30, g.playableCharacter.Y+math.Sin(angle)*30, npcConfig, g.mapLevel)
	npc.Alignment = sc.Alignment
	for i := 0; i < 10; i++ {
		collides := false
		for _, o := range g.obstacles {
			if o.Alive && engine.CheckCollision(npc.GetFootprint(), o.GetFootprint()) { collides = true; break }
		}
		if !collides { break }
		angle := rand.Float64() * 2 * math.Pi
		npc.X = g.playableCharacter.X + math.Cos(angle)*(30.0+rand.Float64()*2)
		npc.Y = g.playableCharacter.Y + math.Sin(angle)*(30.0+rand.Float64()*2)
	}
	g.npcs = append(g.npcs, npc)
}
