package game

import (
	"fmt"
	"image"
	"io/fs"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"oinakos/internal/engine"
)

const (
	WinMenuContinue = 0
	WinMenuQuit     = 1
)

type Game struct {
	width, height    int
	mainCharacter    *MainCharacter
	playerConfig     *EntityConfig
	obstacles        []*Obstacle
	npcs             []*NPC
	projectiles      []*Projectile
	isGameOver       bool
	isMapWon         bool
	mapWonMenuIndex  int // 0: Continue, 1: Quit
	isPaused         bool
	currentMapType   MapType
	mapLevel         int
	initialMapTypeID string
	debug            bool

	generatedChunks map[image.Point]bool
	npcSpawnTimer   int
	playTime        float64

	camera *engine.Camera
	assets fs.FS

	floatingTexts     []*FloatingText
	archetypeRegistry *ArchetypeRegistry
	mapTypeRegistry   *MapTypeRegistry
	obstacleRegistry  *ObstacleRegistry
	initialMapID      string
	lastSavePath      string
	input             engine.Input
	audio             AudioManager
}

func NewGame(assets fs.FS, initialMapID, initialMapTypeID string, input engine.Input, audio AudioManager, debug bool) *Game {
	rand.Seed(time.Now().UnixNano())

	// Load mainCharacter config
	pConfig, err := LoadMainCharacterConfig(assets)
	if err != nil {
		log.Printf("Warning: failed to load main character: %v. Using default values.", err)
	}

	mainCharacter := NewMainCharacter(0, 0, pConfig)
	pIsoX, pIsoY := engine.CartesianToIso(mainCharacter.X, mainCharacter.Y)

	// Registries
	archetypeRegistry := NewArchetypeRegistry()
	archetypeRegistry.LoadAll(assets)

	mapTypeRegistry := NewMapTypeRegistry()
	mapTypeRegistry.LoadAll(assets)

	obstacleRegistry := NewObstacleRegistry()
	obstacleRegistry.LoadAll(assets)

	var selectedMapType MapType
	if len(mapTypeRegistry.IDs) > 0 {
		selectedMapType = *mapTypeRegistry.Types[mapTypeRegistry.IDs[0]]
	}

	g := &Game{
		width:             1280,
		height:            720,
		mainCharacter:     mainCharacter,
		camera:            engine.NewCamera(pIsoX, pIsoY),
		assets:            assets,
		generatedChunks:   make(map[image.Point]bool),
		npcSpawnTimer:     0,
		archetypeRegistry: archetypeRegistry,
		mapTypeRegistry:   mapTypeRegistry,
		obstacleRegistry:  obstacleRegistry,
		currentMapType:    selectedMapType,
		mapLevel:          1,
		initialMapID:      initialMapID,
		initialMapTypeID:  initialMapTypeID,
		input:             input,
		audio:             audio,
		debug:             debug,
	}

	// Trigger map loading if requested
	instanceLoaded := false
	if g.initialMapID != "" {
		// Define search candidates in order of priority
		candidates := []string{
			g.initialMapID, // 1. As provided
			fmt.Sprintf("data/maps/%s", g.initialMapID), // 2. Inside data/maps/
		}

		// If no extension, add .yaml and .yml variants
		if !strings.HasSuffix(g.initialMapID, ".yaml") && !strings.HasSuffix(g.initialMapID, ".yml") {
			candidates = append(candidates,
				g.initialMapID+".yaml",
				g.initialMapID+".yml",
				fmt.Sprintf("data/maps/%s.yaml", g.initialMapID),
				fmt.Sprintf("data/maps/%s.yml", g.initialMapID),
			)
		}

		// Try each candidate
		for _, path := range candidates {
			if err := g.Load(path); err == nil {
				log.Printf("Loaded map instance: %s", path)
				instanceLoaded = true
				break
			}
		}

		if !instanceLoaded {
			// Fallback: Check if it's a map type ID
			if m, ok := g.mapTypeRegistry.Types[g.initialMapID]; ok {
				g.currentMapType = *m
				log.Printf("Using initial map type: %s", g.initialMapID)
			} else {
				log.Printf("Warning: requested map %s not found", g.initialMapID)
			}
		}
	} else if g.initialMapTypeID != "" {
		if m, ok := g.mapTypeRegistry.Types[g.initialMapTypeID]; ok {
			g.currentMapType = *m
			log.Printf("Using initial map type target: %s", g.initialMapTypeID)
		}
	}

	// Initial generation around mainCharacter
	g.updateChunks()

	// Spawn NPCs if not loaded from instance
	if !instanceLoaded {
		g.npcs = make([]*NPC, 0)
		g.loadMapLevel()
	}

	return g
}

func (g *Game) loadMapLevel() {
	// Pick a map, generally increasing difficulty matching the level
	availableMaps := make([]MapType, 0)
	for _, id := range g.mapTypeRegistry.IDs {
		m := g.mapTypeRegistry.Types[id]
		// allow slightly higher or lower difficulty maps
		if m.Difficulty <= g.mapLevel+1 {
			availableMaps = append(availableMaps, *m)
		}
	}

	if g.initialMapID != "" && g.mapLevel == 1 {
		if m, ok := g.mapTypeRegistry.Types[g.initialMapID]; ok {
			g.currentMapType = *m
			log.Printf("Using initial map selection: %s", g.initialMapID)
		} else {
			log.Printf("Warning: requested initial map %s not found", g.initialMapID)
			// fallback below
		}
	}

	if g.currentMapType.ID == "" {
		if len(availableMaps) > 0 {
			g.currentMapType = availableMaps[rand.Intn(len(availableMaps))]
		} else if len(g.mapTypeRegistry.IDs) > 0 {
			g.currentMapType = *g.mapTypeRegistry.Types[g.mapTypeRegistry.IDs[0]] // Fallback
		}
	}

	// Reset map-specific state
	g.playTime = 0
	g.npcs = make([]*NPC, 0)
	g.obstacles = make([]*Obstacle, 0)
	g.floatingTexts = make([]*FloatingText, 0)
	g.currentMapType.StartTime = 0
	g.mainCharacter.MapKills = make(map[string]int) // reset per-map kills
	g.mapWonMenuIndex = 0

	// Camera Snap to player
	pIsoX, pIsoY := engine.CartesianToIso(g.mainCharacter.X, g.mainCharacter.Y)
	g.camera.SnapTo(pIsoX, pIsoY)

	// Apply Difficulty Multipliers
	g.currentMapType.TargetTime *= float64(g.mapLevel)

	newKills := make(map[string]int)
	for npcID, target := range g.currentMapType.TargetKills {
		newKills[npcID] = target * g.mapLevel
	}
	g.currentMapType.TargetKills = newKills

	// Spawn map targets
	switch g.currentMapType.Type {
	case ObjKillVIP:
		if len(g.archetypeRegistry.IDs) > 0 {
			vipID := g.archetypeRegistry.IDs[rand.Intn(len(g.archetypeRegistry.IDs))]
			vipConfig := g.archetypeRegistry.Archetypes[vipID]
			// Spawn far away
			tpX := g.mainCharacter.X + (rand.Float64()*40 - 20)
			tpY := g.mainCharacter.Y + (rand.Float64()*40 - 20)
			if tpX > -5 && tpX < 5 {
				tpX += 10
			}
			if tpY > -5 && tpY < 5 {
				tpY += 10
			}

			vip := NewNPC(tpX, tpY, vipConfig, g.mapLevel*2)
			// Boost VIP stats based on level
			vip.MaxHealth *= g.mapLevel * 2
			vip.Health = vip.MaxHealth
			vip.BaseAttack += g.mapLevel * 2
			g.npcs = append(g.npcs, vip)

			// HACK: we use TargetNPC here to mark the VIP
			// A better approach would be an interface, but we reuse existing structures
			g.currentMapType.TargetPoint = engine.Point{X: tpX, Y: tpY} // Not used directly, but stores initial pos

			// We need a way to track the exact instance. In a real ECS we'd use ID.
			// For now, we tag it via config pointer address since we don't have unique IDs per instance yet.
			// Actually, let's just make the *first* NPC the VIP for simplicity if it's a VIP map.
		}
	case ObjReachPortal, ObjReachZone:
		g.currentMapType.TargetPoint = engine.Point{
			X: g.mainCharacter.X + (rand.Float64()*60 - 30),
			Y: g.mainCharacter.Y + (rand.Float64()*60 - 30),
		}
		if g.currentMapType.TargetPoint.X > -10 && g.currentMapType.TargetPoint.X < 10 {
			g.currentMapType.TargetPoint.X += 20
		}
		if g.currentMapType.TargetPoint.Y > -10 && g.currentMapType.TargetPoint.Y < 10 {
			g.currentMapType.TargetPoint.Y += 20
		}
	case ObjReachBuilding:
		// Pick a random direction
		g.currentMapType.TargetPoint = engine.Point{
			X: g.mainCharacter.X + (rand.Float64()*50 - 25),
			Y: g.mainCharacter.Y + (rand.Float64()*50 - 25),
		}
		if g.currentMapType.TargetPoint.X > -10 && g.currentMapType.TargetPoint.X < 10 {
			g.currentMapType.TargetPoint.X += 20
		}
		if g.currentMapType.TargetPoint.Y > -10 && g.currentMapType.TargetPoint.Y < 10 {
			g.currentMapType.TargetPoint.Y += 20
		}
		// We'll spawn a building there in the update loop or rely on chunks. Let's force a building spawn.
		if config, ok := g.obstacleRegistry.Archetypes["warehouse"]; ok {
			g.obstacles = append(g.obstacles, NewObstacle(g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y, config))
		} else {
			log.Println("WARNING: Warehouse config not found for ObjReachBuilding!")
		}
	case ObjProtectNPC:
		// Target point far away
		g.currentMapType.TargetPoint = engine.Point{
			X: g.mainCharacter.X + (rand.Float64()*80 - 40),
			Y: g.mainCharacter.Y + (rand.Float64()*80 - 40),
		}
		if g.currentMapType.TargetPoint.X > -20 && g.currentMapType.TargetPoint.X < 20 {
			g.currentMapType.TargetPoint.X += 40
		}
		if g.currentMapType.TargetPoint.Y > -20 && g.currentMapType.TargetPoint.Y < 20 {
			g.currentMapType.TargetPoint.Y += 40
		}

		// Spawn Escort right next to mainCharacter
		if config, ok := g.archetypeRegistry.Archetypes["magi_male"]; ok {
			escort := NewNPC(g.mainCharacter.X+2, g.mainCharacter.Y+2, config, g.mapLevel)
			g.npcs = append([]*NPC{escort}, g.npcs...) // Prepend so it's always index 0 for easy tracking
		} else {
			log.Println("WARNING: Magi config (magi_male) not found!")
		}
	case ObjDestroyBuilding:
		g.currentMapType.TargetPoint = engine.Point{
			X: g.mainCharacter.X + (rand.Float64()*80 - 40),
			Y: g.mainCharacter.Y + (rand.Float64()*80 - 40),
		}
		if g.currentMapType.TargetPoint.X > -20 && g.currentMapType.TargetPoint.X < 20 {
			g.currentMapType.TargetPoint.X += 40
		}
		if g.currentMapType.TargetPoint.Y > -20 && g.currentMapType.TargetPoint.Y < 20 {
			g.currentMapType.TargetPoint.Y += 40
		}

		// Spawn a target building like a warehouse or farm
		if config, ok := g.obstacleRegistry.Archetypes["house_burned"]; ok {
			targetObs := NewObstacle(g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y, config)
			g.obstacles = append(g.obstacles, targetObs)
			g.currentMapType.TargetObstacle = targetObs
		} else {
			log.Println("WARNING: house_burned config not found for ObjDestroyBuilding!")
		}
	}

	// Spawn Inhabitants (Corpses, specific encounter targets, etc.)
	for _, ps := range g.currentMapType.Inhabitants {
		if config, ok := g.archetypeRegistry.Archetypes[ps.Archetype]; ok {
			npc := NewNPC(ps.X, ps.Y, config, g.mapLevel)
			npc.Alignment = ps.Alignment
			if ps.Name != "" {
				npc.Name = ps.Name
			}
			if ps.State == "dead" {
				npc.Health = 0
				npc.State = NPCDead
			} else {
				// Default to alive/active
				npc.State = NPCIdle
			}
			g.npcs = append(g.npcs, npc)
		} else {
			log.Printf("WARNING: Inhabitant archetype not found: %s", ps.Archetype)
		}
	}

	// Spawn PreSpawn Obstacles — MUST BE LOADED BEFORE NPC/PC SAFE SPAWN
	for _, po := range g.currentMapType.Obstacles {
		if po.Disabled {
			continue
		}
		if config, ok := g.obstacleRegistry.Archetypes[po.Archetype]; ok {
			px, py := 0.0, 0.0
			if po.X != nil {
				px = *po.X
			}
			if po.Y != nil {
				py = *po.Y
			}
			g.obstacles = append(g.obstacles, NewObstacle(px, py, config))
		} else {
			log.Printf("WARNING: PreSpawn obstacle archetype not found: %s", po.Archetype)
		}
	}

	// Ensure NPCs start in a safe spot (not inside my new huge building shadows)
	for _, n := range g.npcs {
		if !n.IsAlive() {
			continue
		}
		const maxTries = 500
		for i := 0; i < maxTries; i++ {
			if !n.checkCollisionAt(n.X, n.Y, g.obstacles) {
				break
			}
			// Push southward/outward more aggressively
			n.X += 0.5
			n.Y += 0.5
		}
	}

	// Ensure MainCharacter starts in a safe (non-colliding) spot
	const maxTries = 500
	radius := 1.0
	for i := 0; i < maxTries; i++ {
		if !g.mainCharacter.checkCollisionAt(g.mainCharacter.X, g.mainCharacter.Y, g.obstacles) {
			break // Safe!
		}
		// Spiraling outward search for a safe spot — larger steps for huge building shadows
		angle := float64(i) * 0.3
		dist := radius + (float64(i) * 0.2)
		g.mainCharacter.X += math.Cos(angle) * dist
		g.mainCharacter.Y += math.Sin(angle) * dist
	}

	log.Printf("Starting Map Level %d: %s at safe pos %.2f,%.2f", g.mapLevel, g.currentMapType.Name, g.mainCharacter.X, g.mainCharacter.Y)
}

func (g *Game) Update() error {
	if g.isGameOver {
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			os.Exit(0)
		}
		if g.input.IsKeyJustPressed(engine.KeyEnter) {
			*g = *NewGame(g.assets, g.initialMapID, g.initialMapTypeID, g.input, g.audio, g.debug)
		}
		return nil
	}

	if g.isMapWon {
		// Player beat the map — wait for ENTER or ESC or Up/Down
		if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
			g.mapWonMenuIndex = 0
		}
		if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
			g.mapWonMenuIndex = 1
		}

		if g.input.IsKeyJustPressed(engine.KeyEnter) {
			if g.mapWonMenuIndex == WinMenuContinue {
				// Advance to next map
				g.mapLevel++
				// Heal as reward
				g.mainCharacter.Health += g.mainCharacter.MaxHealth / 2
				if g.mainCharacter.Health > g.mainCharacter.MaxHealth {
					g.mainCharacter.Health = g.mainCharacter.MaxHealth
				}
				g.isMapWon = false
				g.loadMapLevel()
			} else if g.mapWonMenuIndex == WinMenuQuit {
				os.Exit(0)
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			os.Exit(0)
		}
		return nil
	}

	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		if g.isPaused {
			os.Exit(0)
		}
		g.isPaused = true
		return nil
	}

	if g.isPaused {
		// If any other key is pressed, resume
		keys := g.input.AppendJustPressedKeys(nil)
		if len(keys) > 0 {
			g.isPaused = false
		}
		return nil
	}

	// Handle Save (S) / Load (F9)
	// Manual input checks should ideally be decoupled, but for now we keep them
	// using engine abstractions if possible or just leaving them for later refactor
	/*
		if g.input.IsKeyJustPressed("S") {
			if err := os.MkdirAll("quicksaves", 0755); err == nil {
				savePath := fmt.Sprintf("quicksaves/save_%s.yaml", time.Now().Format("20060102_150405"))
				if err := g.Save(savePath); err == nil {
					log.Printf("Game saved: %s", savePath)
					g.lastSavePath = savePath
				} else {
					log.Printf("Failed to save: %v", err)
				}
			}
		}
		if g.input.IsKeyJustPressed("F9") {
			if g.lastSavePath != "" {
				if err := g.Load(g.lastSavePath); err == nil {
					log.Printf("Game loaded: %s", g.lastSavePath)
				} else {
					log.Printf("Failed to load: %v", err)
				}
			}
		}
	*/

	// Check and generate new chunks
	g.updateChunks()

	// Handle NPC spawning
	g.updateNPCSpawning()

	// Handle projectiles
	activeProjectiles := []*Projectile{}
	for _, p := range g.projectiles {
		p.Update(g.mainCharacter, g.obstacles, &g.floatingTexts)
		if p.Alive {
			activeProjectiles = append(activeProjectiles, p)
		}
	}
	g.projectiles = activeProjectiles

	if !g.isPaused && !g.isGameOver {
		g.playTime += 1.0 / 60.0
	}

	g.mainCharacter.Update(g.input, g.audio, g.obstacles, g.npcs, &g.floatingTexts, g.currentMapType.MapWidth, g.currentMapType.MapHeight)

	// Real-time position tracking for the USER and Agent
	if g.mainCharacter.Tick%30 == 0 {
		isIllegal := g.mainCharacter.checkCollisionAt(g.mainCharacter.X, g.mainCharacter.Y, g.obstacles)
		status := "OK"
		if isIllegal {
			status = "ILLEGAL POSITION (INSIDE BUILDING)"
		}

		// Find nearest building
		nearestDist := 999.0
		nearestName := "None"
		for _, o := range g.obstacles {
			dist := math.Sqrt(math.Pow(g.mainCharacter.X-o.X, 2) + math.Pow(g.mainCharacter.Y-o.Y, 2))
			if dist < nearestDist {
				nearestDist = dist
				if o.Archetype != nil {
					nearestName = o.Archetype.Name
				}
			}
		}

		log.Printf("[REALTIME] Player Pos: X=%.2f, Y=%.2f | Status: %s | Nearest: %s (Dist: %.2f)",
			g.mainCharacter.X, g.mainCharacter.Y, status, nearestName, nearestDist)
	}
	// Write to a dedicated file for the agent to poll
	os.WriteFile("/tmp/oinakos_pos.txt", []byte(fmt.Sprintf("%.2f,%.2f", g.mainCharacter.X, g.mainCharacter.Y)), 0644)

	// Dynamic spawning
	g.updateNPCSpawning()

	for _, o := range g.obstacles {
		o.Update()
	}

	// Check Win Conditions
	mapWon := false
	switch g.currentMapType.Type {
	case ObjKillCount:
		// Count kills made in this map session
		mapKillTotal := 0
		for _, v := range g.mainCharacter.MapKills {
			mapKillTotal += v
		}

		won := false
		hasTarget := false

		// Check per-NPC-type targets
		if len(g.currentMapType.TargetKills) > 0 {
			hasTarget = true
			allMet := true
			for npcID, targetAmount := range g.currentMapType.TargetKills {
				if g.mainCharacter.MapKills[npcID] < targetAmount {
					allMet = false
					break
				}
			}
			if allMet {
				won = true
			}
		}

		// Check total kill-count target (uses per-map kills, not career kills)
		if g.currentMapType.TargetKillCount > 0 {
			hasTarget = true
			if mapKillTotal >= g.currentMapType.TargetKillCount {
				won = true
			} else {
				won = false
			}
		}

		if hasTarget && won {
			mapWon = true
		}
	case ObjSurvive:
		if g.playTime >= g.currentMapType.TargetTime {
			mapWon = true
		}
	case ObjReachPortal, ObjReachZone, ObjReachBuilding:
		// Check distance
		dx := g.mainCharacter.X - g.currentMapType.TargetPoint.X
		dy := g.mainCharacter.Y - g.currentMapType.TargetPoint.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		radius := g.currentMapType.TargetRadius
		if radius <= 0 {
			radius = 2.0
		} // default

		if dist < radius {
			mapWon = true
		}
	case ObjProtectNPC:
		if len(g.npcs) > 0 {
			escort := g.npcs[0] // Assumed index 0 from NewGame
			if !escort.IsAlive() {
				// Escort died, game over
				g.isGameOver = true
			} else {
				// Check distance to target
				dx := escort.X - g.currentMapType.TargetPoint.X
				dy := escort.Y - g.currentMapType.TargetPoint.Y
				dist := math.Sqrt(dx*dx + dy*dy)

				radius := g.currentMapType.TargetRadius
				if radius <= 0 {
					radius = 5.0
				}
				if dist < radius {
					mapWon = true
				}
			}
		}
	case ObjKillVIP:
		// For simplicity, let's check if the first NPC (the VIP) is dead.
		if len(g.npcs) > 0 {
			if !g.npcs[0].IsAlive() {
				mapWon = true
			}
		} else {
			// If no NPCs, maybe we killed them all
			if g.playTime > 2 { // give it a sec to spawn
				mapWon = true
			}
		}
	case ObjPacifist:
		if g.playTime >= g.currentMapType.TargetTime {
			mapWon = true
		}
		// Failure condition: Don't kill anything
		for _, kills := range g.mainCharacter.MapKills {
			if kills > 0 {
				g.isGameOver = true
				break
			}
		}
	case ObjDestroyBuilding:
		if g.currentMapType.TargetObstacle != nil {
			if !g.currentMapType.TargetObstacle.Alive {
				mapWon = true
			}
		}
	}

	if mapWon && !g.isGameOver && g.mainCharacter.IsAlive() {
		// Show win dialog, don't auto-advance
		g.isMapWon = true
		return nil
	}

	// Filter dead obstacles
	aliveObstacles := make([]*Obstacle, 0, len(g.obstacles))
	for _, o := range g.obstacles {
		if o.Alive {
			aliveObstacles = append(aliveObstacles, o)
		}
	}
	g.obstacles = aliveObstacles

	// Check if mainCharacter died
	if !g.mainCharacter.IsAlive() {
		g.isGameOver = true
	}

	// Update all NPCs (keep corpses indefinitely per user request)
	for _, n := range g.npcs {
		if n.IsAlive() {
			n.Update(g.mainCharacter, g.obstacles, g.npcs, &g.projectiles, &g.floatingTexts, g.currentMapType.MapWidth, g.currentMapType.MapHeight, g.audio)
		}
	}

	// Update floating texts
	activeTexts := make([]*FloatingText, 0)
	for _, ft := range g.floatingTexts {
		if ft.Update() {
			activeTexts = append(activeTexts, ft)
		}
	}
	g.floatingTexts = activeTexts

	// Camera follows mainCharacter
	pIsoX, pIsoY := engine.CartesianToIso(g.mainCharacter.X, g.mainCharacter.Y)
	g.camera.Follow(pIsoX, pIsoY, 0.1)

	// Final safety check: ensure player is not stuck in a newly loaded obstacle
	for i := 0; i < 50; i++ {
		if !g.mainCharacter.checkCollisionAt(g.mainCharacter.X, g.mainCharacter.Y, g.obstacles) {
			break
		}
		g.mainCharacter.X += rand.Float64()*2 - 1
		g.mainCharacter.Y += rand.Float64()*2 - 1
		// Update camera to match new position
		ncX, ncY := engine.CartesianToIso(g.mainCharacter.X, g.mainCharacter.Y)
		g.camera.SnapTo(ncX, ncY)
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.width, g.height
}

func (g *Game) updateChunks() {
	// Procedural spawning disabled per user request: "obstacles MUST NOT spawn at all"
	/*
		const chunkSize = 10
		cpX := int(math.Floor(g.mainCharacter.X / float64(chunkSize)))
		cpY := int(math.Floor(g.mainCharacter.Y / float64(chunkSize)))

		// Check 9x9 grid around mainCharacter (radius 4)
		// With chunkSize=10, this covers ±40 tiles, well beyond renderer's 25-tile limit.
		for dy := -4; dy <= 4; dy++ {
			for dx := -4; dx <= 4; dx++ {
				chunk := image.Point{cpX + dx, cpY + dy}
				if !g.generatedChunks[chunk] {
					g.spawnObstaclesInChunk(chunk.X, chunk.Y)
					g.generatedChunks[chunk] = true
				}
			}
		}
	*/
}

func (g *Game) spawnObstaclesInChunk(cx, cy int) {
	const chunkSize = 10
	startX := float64(cx * chunkSize)
	startY := float64(cy * chunkSize)

	// Seed based on chunk coordinates for stability
	r := rand.New(rand.NewSource(int64(cx*1000 + cy)))

	if len(g.obstacleRegistry.IDs) == 0 {
		return // Do not spawn anything if registries aren't loaded
	}

	// 1. Forest patch
	if r.Float64() < 0.25 {
		centerX := startX + r.Float64()*chunkSize
		centerY := startY + r.Float64()*chunkSize
		for i := 0; i < 6; i++ {
			tx := centerX + r.NormFloat64()*1.2
			ty := centerY + r.NormFloat64()*1.2
			types := []string{"tree1", "tree2", "tree3", "tree4", "tree5", "tree6", "tree7"}
			ot := types[r.Intn(len(types))]
			config := g.obstacleRegistry.Archetypes[ot]
			g.obstacles = append(g.obstacles, NewObstacle(tx, ty, config))
		}
	}

	// 2. Continuous Stones
	if r.Float64() < 0.2 {
		wx := startX + r.Float64()*chunkSize
		wy := startY + r.Float64()*chunkSize
		for i := 0; i < 4; i++ {
			types := []string{"rock1", "rock2", "rock3", "rock4", "rock5"}
			ot := types[r.Intn(len(types))]
			config := g.obstacleRegistry.Archetypes[ot]
			g.obstacles = append(g.obstacles, NewObstacle(wx, wy, config))
			wx += r.Float64()*1.2 - 0.6
			wy += r.Float64()*1.2 - 0.6
		}
	}

	// 3. Buildings
	if r.Float64() < 0.08 {
		bx := startX + r.Float64()*chunkSize
		by := startY + r.Float64()*chunkSize
		// Don't spawn on top of mainCharacter's initial position
		if math.Abs(bx) > 5 || math.Abs(by) > 5 {
			// Ensure buildings are far from each other
			tooClose := false
			const minBuildingDist = 16.0 // ~1024 pixels (16 * 64)
			for _, obs := range g.obstacles {
				if obs.Archetype == nil {
					continue
				}
				// Simple check for building types
				id := obs.Archetype.ID
				isBuilding := strings.HasPrefix(id, "house") || id == "farm" || id == "smithery" || id == "temple" || id == "warehouse" || id == "house_burned"
				if isBuilding {
					distSq := (obs.X-bx)*(obs.X-bx) + (obs.Y-by)*(obs.Y-by)
					if distSq < minBuildingDist*minBuildingDist {
						tooClose = true
						break
					}
				}
			}

			if !tooClose {
				types := []string{"house1", "house2", "house3", "farm", "smithery", "temple", "warehouse", "house_burned"}
				ot := types[r.Intn(len(types))]
				if config, ok := g.obstacleRegistry.Archetypes[ot]; ok {
					g.obstacles = append(g.obstacles, NewObstacle(bx, by, config))
				}
			}
		}
	}

	// 4. Healing Wells
	if r.Float64() < 0.05 {
		wx := startX + r.Float64()*chunkSize
		wy := startY + r.Float64()*chunkSize
		if math.Abs(wx) > 2 || math.Abs(wy) > 2 {
			if config, ok := g.obstacleRegistry.Archetypes["well"]; ok {
				g.obstacles = append(g.obstacles, NewObstacle(wx, wy, config))
			}
		}
	}

	// 4. Random bushes
	for i := 0; i < 2; i++ {
		if r.Float64() < 0.4 {
			bx := startX + r.Float64()*chunkSize
			by := startY + r.Float64()*chunkSize
			types := []string{"bush1", "bush2", "bush3", "bush4", "bush5"}
			ot := types[r.Intn(len(types))]
			config := g.obstacleRegistry.Archetypes[ot]
			g.obstacles = append(g.obstacles, NewObstacle(bx, by, config))
		}
	}
}

func (g *Game) updateNPCSpawning() {
	// 1. Process individual spawn configurations
	if len(g.currentMapType.Spawns) > 0 {
		for i := range g.currentMapType.Spawns {
			s := &g.currentMapType.Spawns[i]
			if s.Frequency <= 0 {
				continue // Skip if no frequency set
			}

			s.Timer++
			// Convert frequency from seconds to ticks (60fps)
			threshold := int(s.Frequency * 60)
			if s.Timer >= threshold {
				s.Timer = 0

				// Check individual probability
				if rand.Float64() <= s.Probability {
					// Maximum NPC limit check (stay under 100 for performance)
					if len(g.npcs) < 100 {
						if s.X != nil && s.Y != nil {
							g.spawnNPCNearPosition(*s.X, *s.Y, s)
						} else {
							g.spawnNPCAtMapEdges(s)
						}
					}
				}
			}
		}
	}

	// 2. Periodic cleanup of far away NPCs
	// We check every 5 seconds (300 ticks) roughly
	g.npcSpawnTimer++
	if g.npcSpawnTimer >= 300 {
		g.npcSpawnTimer = 0
		activeNPCs := make([]*NPC, 0)
		for _, n := range g.npcs {
			dist := math.Sqrt(math.Pow(n.X-g.mainCharacter.X, 2) + math.Pow(n.Y-g.mainCharacter.Y, 2))
			// Only cull if it's far away AND not a corpse
			// Corpses remain forever per user rule
			if n.IsAlive() {
				if dist < 40 {
					activeNPCs = append(activeNPCs, n)
				}
			} else {
				activeNPCs = append(activeNPCs, n)
			}
		}
		g.npcs = activeNPCs
	}
}

func (g *Game) spawnNPCNearPosition(x, y float64, sc *SpawnConfig) {
	if len(g.archetypeRegistry.IDs) == 0 || sc == nil {
		return
	}
	const spawnRadius = 2.0
	angle := rand.Float64() * 2 * math.Pi
	ex := x + math.Cos(angle)*spawnRadius
	ey := y + math.Sin(angle)*spawnRadius

	npcConfig := g.archetypeRegistry.Archetypes[sc.Archetype]
	if npcConfig == nil {
		return
	}
	npc := NewNPC(ex, ey, npcConfig, g.mapLevel)
	npc.Alignment = sc.Alignment

	// Collision retry
	for i := 0; i < 10; i++ {
		collides := false
		for _, o := range g.obstacles {
			if o.Alive && engine.CheckCollision(npc.GetFootprint(), o.GetFootprint()) {
				collides = true
				break
			}
		}
		if !collides {
			break
		}
		angle := rand.Float64() * 2 * math.Pi
		npc.X = x + math.Cos(angle)*(spawnRadius+rand.Float64())
		npc.Y = y + math.Sin(angle)*(spawnRadius+rand.Float64())
	}
	g.npcs = append(g.npcs, npc)
}

func (g *Game) spawnNPCAtMapEdges(sc *SpawnConfig) {
	if len(g.archetypeRegistry.IDs) == 0 || sc == nil {
		return
	}

	const spawnDist = 30.0
	angle := rand.Float64() * 2 * math.Pi
	ex := g.mainCharacter.X + math.Cos(angle)*spawnDist
	ey := g.mainCharacter.Y + math.Sin(angle)*spawnDist

	npcConfig := g.archetypeRegistry.Archetypes[sc.Archetype]
	if npcConfig == nil {
		return
	}
	npc := NewNPC(ex, ey, npcConfig, g.mapLevel)
	npc.Alignment = sc.Alignment

	// Edges usually clear but let's check anyway
	for i := 0; i < 10; i++ {
		collides := false
		for _, o := range g.obstacles {
			if o.Alive && engine.CheckCollision(npc.GetFootprint(), o.GetFootprint()) {
				collides = true
				break
			}
		}
		if !collides {
			break
		}
		angle := rand.Float64() * 2 * math.Pi
		npc.X = g.mainCharacter.X + math.Cos(angle)*(spawnDist+rand.Float64()*2)
		npc.Y = g.mainCharacter.Y + math.Sin(angle)*(spawnDist+rand.Float64()*2)
	}

	g.npcs = append(g.npcs, npc)
}
