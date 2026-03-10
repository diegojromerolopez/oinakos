package game

import (
	"fmt"
	"image"
	"io/fs"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
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
	width, height     int
	mainCharacter     *MainCharacter
	playerConfig      *EntityConfig
	obstacles         []*Obstacle
	npcs              []*NPC
	projectiles       []*Projectile
	isGameOver        bool
	isMapWon          bool
	isGameWon         bool // For completing entire campaign or single map
	mapWonMenuIndex   int  // 0: Continue/Replay, 1: Quit
	isPaused          bool
	currentMapType    MapType
	mapLevel          int       // Current level (for scaling)
	currentCampaign   *Campaign // If playing a campaign
	campaignIndex     int       // Progress in campaign Maps
	isCampaign        bool      // True if playing a campaign
	isMainMenu        bool      // True if showing main menu
	mainMenuIndex     int       // Index for main menu
	isSettingsScreen  bool      // True if showing settings screen
	settingsMenuIndex int       // Index for settings menu
	isCampaignSelect  bool      // True if showing campaign picker
	campaignMenuIndex int       // Index of selected campaign
	initialMapTypeID  string
	debug             bool

	generatedChunks map[image.Point]bool
	npcSpawnTimer   int
	playTime        float64

	camera *engine.Camera
	assets fs.FS

	floatingTexts             []*FloatingText
	archetypeRegistry         *ArchetypeRegistry
	playableCharacterRegistry *PlayableCharacterRegistry
	mapTypeRegistry           *MapTypeRegistry
	campaignRegistry          *CampaignRegistry
	obstacleRegistry          *ObstacleRegistry
	initialMapID              string
	initialHeroID             string
	lastSavePath              string
	input                     engine.Input
	showBoundaries            bool
	audio                     AudioManager
	npcRegistry               *NPCRegistry

	isMenuOpen       bool
	menuIndex        int // 0: Resume, 1: Quicksave, 2: Load, 3: Quit
	loadDialogActive bool
	loadPathInput    string

	isCharacterSelect  bool
	characterMenuIndex int
	saveMessage        string
	saveMessageTimer   int // Ticks to show the message

	settings *Settings

	isSettingsFromPause bool
}

func NewGame(assets fs.FS, initialMapID, initialMapTypeID, initialHeroID string, input engine.Input, audio AudioManager, debug bool) *Game {
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

	playableCharacterRegistry := NewPlayableCharacterRegistry()
	playableCharacterRegistry.LoadAll(assets)

	mapTypeRegistry := NewMapTypeRegistry()
	mapTypeRegistry.LoadAll(assets)

	campaignRegistry := NewCampaignRegistry()
	campaignRegistry.LoadAll(assets)

	obstacleRegistry := NewObstacleRegistry()
	obstacleRegistry.LoadAll(assets)

	npcRegistry := NewNPCRegistry()
	npcRegistry.LoadAll(assets)

	var selectedMapType MapType
	if m, ok := mapTypeRegistry.Types["safe_zone"]; ok {
		selectedMapType = *m
	} else if len(mapTypeRegistry.IDs) > 0 {
		selectedMapType = *mapTypeRegistry.Types[mapTypeRegistry.IDs[0]]
	}

	g := &Game{
		width:                     1280,
		height:                    720,
		mainCharacter:             mainCharacter,
		camera:                    engine.NewCamera(pIsoX, pIsoY),
		assets:                    assets,
		generatedChunks:           make(map[image.Point]bool),
		npcSpawnTimer:             0,
		archetypeRegistry:         archetypeRegistry,
		mapTypeRegistry:           mapTypeRegistry,
		campaignRegistry:          campaignRegistry,
		obstacleRegistry:          obstacleRegistry,
		npcRegistry:               npcRegistry,
		playableCharacterRegistry: playableCharacterRegistry,
		currentMapType:            selectedMapType,
		mapLevel:                  1,
		initialMapID:              initialMapID,
		initialMapTypeID:          initialMapTypeID,
		initialHeroID:             initialHeroID,
		input:                     input,
		audio:                     audio,
		debug:                     debug,
	}

	g.settings = LoadSettings()
	if audio != nil {
		audio.SetProbability(g.settings.GetSoundProbability())
	}

	SetDebugMode(debug)
	DebugLog("Game Initialized | MapID: %s | MapTypeID: %s", initialMapID, initialMapTypeID)

	// Default to Main Menu for new games
	g.isMainMenu = true

	// WASM Auto-resumption from localStorage
	if g.isWasm() {
		if err := g.Load("quicksave"); err == nil {
			log.Printf("Successfully Resumed Game from Browser Storage")
			g.isMainMenu = false
			g.isCharacterSelect = false
			return g
		}
	}

	if g.initialHeroID != "" {
		if config, ok := g.playableCharacterRegistry.Characters[g.initialHeroID]; ok {
			g.mainCharacter.Config = config
			g.mainCharacter.Health = config.Stats.HealthMin
			g.mainCharacter.MaxHealth = config.Stats.HealthMin
			g.mainCharacter.Speed = config.Stats.Speed
			g.mainCharacter.BaseAttack = config.Stats.BaseAttack
			g.mainCharacter.BaseDefense = config.Stats.BaseDefense
			g.mainCharacter.Weapon = config.Weapon
			g.isCharacterSelect = false
			log.Printf("Using initial hero: %s", g.initialHeroID)
		} else {
			log.Printf("Warning: initial hero %s not found in registry", g.initialHeroID)
		}
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
				log.Printf("Loaded map instance: %s | Closing Menu", path)
				instanceLoaded = true
				g.isMainMenu = false
				// If hero wasn't selected via flag, and we don't have one from save, go to select
				if g.initialHeroID == "" && g.mainCharacter.Config == nil {
					g.isCharacterSelect = true
				} else {
					g.isCharacterSelect = false
				}
				break
			} else {
				log.Printf("DEBUG: Failed to load candidate %s: %v", path, err)
			}
		}

		if !instanceLoaded {
			// Fallback: Check if it's a map type ID
			if m, ok := g.mapTypeRegistry.Types[g.initialMapID]; ok {
				g.currentMapType = *m
				g.isMainMenu = false
				g.isCharacterSelect = true
				log.Printf("Using initial map type: %s", g.initialMapID)
			} else {
				log.Printf("Warning: requested map %s not found", g.initialMapID)
			}
		}
	} else if g.initialMapTypeID != "" {
		if m, ok := g.mapTypeRegistry.Types[g.initialMapTypeID]; ok {
			g.currentMapType = *m
			g.isMainMenu = false
			g.isCharacterSelect = true
			log.Printf("Using initial map type target: %s", g.initialMapTypeID)
		}
	}

	// Initial generation around mainCharacter
	g.updateChunks()

	// Spawn NPCs if not loaded from instance and not in menu
	if !instanceLoaded && !g.isMainMenu {
		g.npcs = make([]*NPC, 0)
		g.loadMapLevel()
	}

	return g
}

func (g *Game) loadMapLevel() {
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
		// Fallback picker
		if len(g.mapTypeRegistry.IDs) > 0 {
			g.currentMapType = *g.mapTypeRegistry.Types[g.mapTypeRegistry.IDs[0]]
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

	// Initial player position
	if g.currentMapType.Player != nil {
		g.mainCharacter.X = g.currentMapType.Player.X
		g.mainCharacter.Y = g.currentMapType.Player.Y
	}

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
		if g.currentMapType.TargetPointRaw != nil {
			g.currentMapType.TargetPoint = engine.Point{
				X: g.currentMapType.TargetPointRaw.X,
				Y: g.currentMapType.TargetPointRaw.Y,
			}
		} else {
			// Fallback: random point at a safe distance
			offX := rand.Float64()*60 - 30
			offY := rand.Float64()*60 - 30
			if offX > -10 && offX < 10 {
				offX += 20
			}
			if offY > -10 && offY < 10 {
				offY += 20
			}
			g.currentMapType.TargetPoint = engine.Point{
				X: g.mainCharacter.X + offX,
				Y: g.mainCharacter.Y + offY,
			}
		}
	case ObjReachBuilding:
		if g.currentMapType.TargetPointRaw != nil {
			g.currentMapType.TargetPoint = engine.Point{
				X: g.currentMapType.TargetPointRaw.X,
				Y: g.currentMapType.TargetPointRaw.Y,
			}
		} else {
			offX := rand.Float64()*50 - 25
			offY := rand.Float64()*50 - 25
			if offX > -10 && offX < 10 {
				offX += 20
			}
			if offY > -10 && offY < 10 {
				offY += 20
			}
			g.currentMapType.TargetPoint = engine.Point{
				X: g.mainCharacter.X + offX,
				Y: g.mainCharacter.Y + offY,
			}
		}
		// Spawn a building at the target
		if config, ok := g.obstacleRegistry.Archetypes["warehouse"]; ok {
			g.obstacles = append(g.obstacles, NewObstacle("target_warehouse", g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y, config))
		} else {
			log.Println("WARNING: Warehouse config not found for ObjReachBuilding!")
		}
	case ObjProtectNPC:
		if g.currentMapType.TargetPointRaw != nil {
			g.currentMapType.TargetPoint = engine.Point{
				X: g.currentMapType.TargetPointRaw.X,
				Y: g.currentMapType.TargetPointRaw.Y,
			}
		} else {
			offX := rand.Float64()*80 - 40
			offY := rand.Float64()*80 - 40
			if offX > -20 && offX < 20 {
				offX += 40
			}
			if offY > -20 && offY < 20 {
				offY += 40
			}
			g.currentMapType.TargetPoint = engine.Point{
				X: g.mainCharacter.X + offX,
				Y: g.mainCharacter.Y + offY,
			}
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
			targetObs := NewObstacle("target_building", g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y, config)
			g.obstacles = append(g.obstacles, targetObs)
			g.currentMapType.TargetObstacle = targetObs
		} else {
			log.Println("WARNING: house_burned config not found for ObjDestroyBuilding!")
		}
	}

	// Spawn Inhabitants (Corpses, specific encounter targets, etc.)
	allInhabs := append(g.currentMapType.Inhabitants, g.currentMapType.NPCs...)
	for _, ps := range allInhabs {
		var config *EntityConfig
		var ok bool

		id := ps.NPC
		if ps.NPCID != "" {
			id = ps.NPCID
		}

		if id != "" {
			config, ok = g.npcRegistry.NPCs[id]
		} else {
			arch := ps.Archetype
			if ps.ArchetypeID != "" {
				arch = ps.ArchetypeID
			}
			if arch != "" {
				config, ok = g.archetypeRegistry.Archetypes[arch]
				id = arch
			}
		}

		if ok {
			// If this inhabitant is the character the player selected, we'll swap positions
			if id != "" && g.mainCharacter.Config != nil && id == g.mainCharacter.Config.ID {
				g.mainCharacter.X = ps.X
				g.mainCharacter.Y = ps.Y
				// We don't spawn the NPC instance if the player IS that NPC
				continue
			}

			npc := NewNPC(ps.X, ps.Y, config, g.mapLevel)
			npc.Alignment = ps.Alignment
			npc.MustSurvive = ps.MustSurvive
			if ps.Name != "" {
				npc.Name = ps.Name
			}
			if ps.State == "dead" {
				npc.Health = 0
				npc.State = NPCDead
			}
			g.npcs = append(g.npcs, npc)
		} else {
			log.Printf("WARNING: Inhabitant archetype/NPC not found: %s", id)
		}
	}

	// Spawn PreSpawn Obstacles — MUST BE LOADED BEFORE NPC/PC SAFE SPAWN
	for _, po := range g.currentMapType.Obstacles {
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

	DebugLog("Starting Map Level %d: %s at safe pos %.2f,%.2f", g.mapLevel, g.currentMapType.Name, g.mainCharacter.X, g.mainCharacter.Y)
}

func (g *Game) Update() error {
	if g.isMainMenu {
		nOptions := 4
		if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
			g.mainMenuIndex--
			if g.mainMenuIndex < 0 {
				g.mainMenuIndex = nOptions - 1
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
			g.mainMenuIndex++
			if g.mainMenuIndex >= nOptions {
				g.mainMenuIndex = 0
			}
		}

		mx, my := g.input.MousePosition()
		hoverIndex := -1
		// Center alignment horizontal
		centerX := g.width / 2
		for i := 0; i < nOptions; i++ {
			itemY := 300 + i*50
			if mx >= centerX-100 && mx <= centerX+100 && my >= itemY-20 && my <= itemY+20 {
				hoverIndex = i
			}
		}
		if hoverIndex != -1 {
			g.mainMenuIndex = hoverIndex
		}

		handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

		if handleSelect {
			switch g.mainMenuIndex {
			case 0: // New Game
				g.isMainMenu = false
				g.isCharacterSelect = true
			case 1: // Load Game
				g.loadDialogActive = true
			case 2: // Settings
				// Load or create default settings
				g.settings = LoadSettings()
				g.settings.Save() // Ensure folder/file exists
				// Find current index
				for idx, val := range FrequencyOptions {
					if val == g.settings.SoundFrequency {
						g.settingsMenuIndex = idx
						break
					}
				}
				g.isMainMenu = false
				g.isSettingsScreen = true
			case 3: // Quit
				if !g.isWasm() {
					os.Exit(0)
				}
			}
		}

		if g.loadDialogActive {
			path := g.openFilePicker()
			if path != "" {
				if err := g.Load(path); err == nil {
					g.isMainMenu = false
					g.isCharacterSelect = false
				} else {
					log.Printf("Failed to load map: %v", err)
				}
			}
			g.loadDialogActive = false
		}

		return nil
	}

	if g.isSettingsScreen {
		nO := len(FrequencyOptions)
		if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
			g.settingsMenuIndex--
			if g.settingsMenuIndex < 0 {
				g.settingsMenuIndex = nO - 1
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
			g.settingsMenuIndex++
			if g.settingsMenuIndex >= nO {
				g.settingsMenuIndex = 0
			}
		}

		// Mouse selection
		mx, my := g.input.MousePosition()
		hoverIdx := -1
		centerX := g.width / 2
		for i := 0; i < nO; i++ {
			itemY := 300 + i*30
			if mx >= centerX-150 && mx <= centerX+150 && my >= itemY-15 && my <= itemY+15 {
				hoverIdx = i
			}
		}
		if hoverIdx != -1 {
			g.settingsMenuIndex = hoverIdx
		}

		if g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIdx != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft)) {
			// Save setting
			g.settings.SoundFrequency = FrequencyOptions[g.settingsMenuIndex]
			g.settings.Save()
			// Update probability
			if g.audio != nil {
				g.audio.SetProbability(g.settings.GetSoundProbability())
			}
			g.isSettingsScreen = false
			if g.isSettingsFromPause {
				g.isMenuOpen = true
				g.isSettingsFromPause = false
			} else {
				g.isMainMenu = true
			}
		}

		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			g.isSettingsScreen = false
			if g.isSettingsFromPause {
				g.isMenuOpen = true
				g.isSettingsFromPause = false
			} else {
				g.isMainMenu = true
			}
		}
		return nil
	}

	if g.isCharacterSelect {
		nChars := len(g.playableCharacterRegistry.IDs)
		if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
			g.characterMenuIndex--
			if g.characterMenuIndex < 0 {
				g.characterMenuIndex = nChars - 1
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
			g.characterMenuIndex++
			if g.characterMenuIndex >= nChars {
				g.characterMenuIndex = 0
			}
		}

		mx, my := g.input.MousePosition()
		hoverIndex := -1
		for i := 0; i < nChars; i++ {
			if mx >= 100 && mx <= 400 && my >= 130+i*30-5 && my <= 130+i*30+25 {
				hoverIndex = i
			}
		}
		if hoverIndex != -1 {
			g.characterMenuIndex = hoverIndex
		}

		handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

		if handleSelect {
			charID := g.playableCharacterRegistry.IDs[g.characterMenuIndex]
			config := g.playableCharacterRegistry.Characters[charID]
			g.mainCharacter.Config = config
			g.mainCharacter.Health = config.Stats.HealthMin
			g.mainCharacter.MaxHealth = config.Stats.HealthMin
			g.mainCharacter.Speed = config.Stats.Speed
			g.mainCharacter.BaseAttack = config.Stats.BaseAttack
			g.mainCharacter.BaseDefense = config.Stats.BaseDefense
			g.mainCharacter.Weapon = config.Weapon

			g.isCharacterSelect = false
			// If we didn't start with a specific map from flags, go to campaign/map select
			if g.initialMapID == "" && g.initialMapTypeID == "" {
				g.isCampaignSelect = true
			} else {
				g.loadMapLevel()
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			if !g.isWasm() {
				os.Exit(0)
			}
		}
		return nil
	}

	if g.isCampaignSelect {
		nC := len(g.campaignRegistry.IDs)
		nM := len(g.mapTypeRegistry.IDs)

		if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
			g.campaignMenuIndex--
			if g.campaignMenuIndex < 0 {
				g.campaignMenuIndex = nC + nM
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
			g.campaignMenuIndex++
			if g.campaignMenuIndex > nC+nM {
				g.campaignMenuIndex = 0
			}
		}

		// Handle Mouse Hover & Click
		mx, my := g.input.MousePosition()
		hoverIndex := -1
		col1X := 100
		col2X := g.width / 2

		for i := 0; i < nC; i++ {
			cy := 130 + i*25
			if mx >= col1X && mx <= col1X+300 && my >= cy && my <= cy+20 {
				hoverIndex = i
			}
		}
		for i := 0; i < nM; i++ {
			colOffset := col2X
			rowOffset := i
			if i > 15 {
				colOffset += 250 // Shift to a third column if there are tons of maps
				rowOffset = i - 16
			}
			cy := 130 + rowOffset*25
			if mx >= colOffset && mx <= colOffset+300 && my >= cy && my <= cy+20 {
				hoverIndex = nC + i
			}
		}

		quitText := "  QUIT"
		quitW := len(quitText) * 7
		qx, qy := (g.width-quitW)/2, g.height-90
		if mx >= qx && mx <= qx+300 && my >= qy && my <= qy+20 {
			hoverIndex = nC + nM
		}

		if hoverIndex != -1 {
			// Update menu index to what the mouse is hovering over
			g.campaignMenuIndex = hoverIndex
		}

		handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

		if handleSelect {
			if g.campaignMenuIndex < nC {
				// Selected a campaign
				camID := g.campaignRegistry.IDs[g.campaignMenuIndex]
				g.currentCampaign = g.campaignRegistry.Campaigns[camID]
				g.isCampaign = true
				g.campaignIndex = 0
				g.isCampaignSelect = false
				g.loadMapLevel()
			} else if g.campaignMenuIndex < nC+nM {
				// Selected an individual map
				mapID := g.mapTypeRegistry.IDs[g.campaignMenuIndex-nC]
				g.currentMapType = *g.mapTypeRegistry.Types[mapID]
				g.isCampaign = false
				g.isCampaignSelect = false
				g.initialMapID = mapID
				g.loadMapLevel()
			} else {
				// Quit button
				os.Exit(0)
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			os.Exit(0)
		}
		return nil
	}

	if g.isGameWon {
		if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
			g.mapWonMenuIndex = 0
		}
		if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
			g.mapWonMenuIndex = 1
		}
		if g.input.IsKeyJustPressed(engine.KeyEnter) {
			if g.mapWonMenuIndex == 0 { // Replay
				*g = *NewGame(g.assets, g.initialMapID, g.initialMapTypeID, g.initialHeroID, g.input, g.audio, g.debug)
			} else { // Quit
				os.Exit(0)
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			os.Exit(0)
		}
		return nil
	}

	// Handle debug boundaries toggle
	if g.input.IsKeyJustPressed(engine.KeyTab) {
		g.showBoundaries = !g.showBoundaries
		g.debug = g.showBoundaries
		SetDebugMode(g.debug)
	}

	if g.isGameOver {
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			os.Exit(0)
		}
		if g.input.IsKeyJustPressed(engine.KeyEnter) {
			*g = *NewGame(g.assets, g.initialMapID, g.initialMapTypeID, g.initialHeroID, g.input, g.audio, g.debug)
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
				// Advance
				if g.isCampaign && g.currentCampaign != nil {
					g.campaignIndex++
					if g.campaignIndex >= len(g.currentCampaign.Maps) {
						g.isGameWon = true
						g.isMapWon = false
					} else {
						g.loadMapLevel()
						g.isMapWon = false
					}
				} else {
					// Single map win
					g.isGameWon = true
					g.isMapWon = false
				}
			} else if g.mapWonMenuIndex == WinMenuQuit {
				os.Exit(0)
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			os.Exit(0)
		}
		return nil
	}

	// Menu Button Detection (Top-Right)
	if !g.isMenuOpen && !g.isGameOver && !g.isMapWon {
		mx, my := g.input.MousePosition()
		// Button is at Width-110 to Width-10, Y: 40 to 70
		if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
			if mx >= g.width-110 && mx <= g.width-10 && my >= 40 && my <= 70 {
				g.isMenuOpen = true
			}
		}
	}

	if g.loadDialogActive {
		path := g.openFilePicker()
		if path != "" {
			if err := g.Load(path); err != nil {
				log.Printf("Failed to load map: %v", err)
			}
		}
		g.loadDialogActive = false
		return nil
	}

	if g.isMenuOpen {
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			g.isMenuOpen = false
			return nil
		}
		if g.input.IsKeyJustPressed(engine.KeyUp) || g.input.IsKeyJustPressed(engine.KeyW) {
			g.menuIndex--
			if g.menuIndex < 0 {
				g.menuIndex = 4
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyDown) || g.input.IsKeyJustPressed(engine.KeyS) {
			g.menuIndex++
			if g.menuIndex > 4 {
				g.menuIndex = 0
			}
		}

		mw, mh := 400, 280
		bx, by := (g.width-mw)/2, (g.height-mh)/2
		mx, my := g.input.MousePosition()

		hoverIndex := -1
		for i := 0; i < 5; i++ {
			itemY := by + 70 + i*35
			// Rough box: width ~150px, height ~20px
			if mx >= bx+100 && mx <= bx+250 && my >= itemY-10 && my <= itemY+20 {
				hoverIndex = i
			}
		}

		if hoverIndex != -1 {
			g.menuIndex = hoverIndex
		}

		handleSelect := g.input.IsKeyJustPressed(engine.KeyEnter) || (hoverIndex != -1 && g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft))

		if handleSelect {
			switch g.menuIndex {
			case 0: // Resume
				g.isMenuOpen = false
			case 1: // Quicksave
				g.performQuicksave()
				g.isMenuOpen = false
			case 2: // Load
				g.loadDialogActive = true
				g.isMenuOpen = false
			case 3: // Settings
				// Load or create default settings
				g.settings = LoadSettings()
				g.settings.Save() // Ensure folder/file exists
				// Find current index
				for idx, val := range FrequencyOptions {
					if val == g.settings.SoundFrequency {
						g.settingsMenuIndex = idx
						break
					}
				}
				g.isSettingsFromPause = true
				g.isMenuOpen = false
				g.isSettingsScreen = true
			case 4: // Quit
				os.Exit(0)
			}
		}
		return nil
	}

	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isMenuOpen = true
		return nil
	}

	// Handle 'Q' key for QuickSave
	if g.input.IsKeyJustPressed(engine.KeyQ) {
		g.performQuicksave()
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
		p.Update(g.mainCharacter, g.obstacles, &g.floatingTexts, g.audio)
		if p.Alive {
			activeProjectiles = append(activeProjectiles, p)
		}
	}
	g.projectiles = activeProjectiles

	if !g.isPaused && !g.isGameOver {
		g.playTime += 1.0 / 60.0
		if g.saveMessageTimer > 0 {
			g.saveMessageTimer--
		}
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

		DebugLog("[REALTIME] Player Pos: X=%.2f, Y=%.2f | Status: %s | Nearest: %s (Dist: %.2f)",
			g.mainCharacter.X, g.mainCharacter.Y, status, nearestName, nearestDist)
	}
	// Write to a dedicated file for the agent to poll
	os.WriteFile("/tmp/oinakos_pos.txt", []byte(fmt.Sprintf("%.2f,%.2f", g.mainCharacter.X, g.mainCharacter.Y)), 0644)

	// Dynamic spawning
	g.updateNPCSpawning()

	for _, o := range g.obstacles {
		o.Update()
	}

	g.updateProximityEffects()

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
		if !g.isMapWon {
			DebugLog("Objective Completed! Level %d cleared. Objective: %v", g.mapLevel, g.currentMapType.Type)
		}
		g.isMapWon = true
		return nil
	}

	if g.isGameOver && g.mainCharacter.IsAlive() == false {
		// Only log if it just happened
		// We need to check if we already logged it, or just log here once per game over
	}

	// Update all obstacles (important for animation and cooldowns)
	aliveObstacles := make([]*Obstacle, 0, len(g.obstacles))
	for _, o := range g.obstacles {
		o.Update()
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
		n.Update(g.mainCharacter, g.obstacles, g.npcs, &g.projectiles, &g.floatingTexts, g.currentMapType.MapWidth, g.currentMapType.MapHeight, g.audio)
		if n.MustSurvive && !n.IsAlive() {
			if !g.isGameOver {
				DebugLog("CRITICAL FAILURE: [%s] was killed! Quest Failed.", n.Name)
			}
			g.isGameOver = true
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
			g.obstacles = append(g.obstacles, NewObstacle(fmt.Sprintf("gen_%s_%.1f_%.1f", ot, tx, ty), tx, ty, config))
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
			g.obstacles = append(g.obstacles, NewObstacle(fmt.Sprintf("gen_%s_%.1f_%.1f", ot, wx, wy), wx, wy, config))
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
					g.obstacles = append(g.obstacles, NewObstacle(fmt.Sprintf("gen_tree_palm_%.0f_%.0f", bx, by), bx, by, config))
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
				g.obstacles = append(g.obstacles, NewObstacle(fmt.Sprintf("gen_well_%.1f_%.1f", wx, wy), wx, wy, config))
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
			g.obstacles = append(g.obstacles, NewObstacle(fmt.Sprintf("gen_%s_%.1f_%.1f", ot, bx, by), bx, by, config))
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
		if len(g.npcs) != len(activeNPCs) {
			DebugLog("Culled %d far-away NPCs", len(g.npcs)-len(activeNPCs))
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

	// 5% chance to check for elite variants in npcRegistry
	if rand.Float64() < 0.05 {
		var variants []*EntityConfig
		for _, v := range g.npcRegistry.NPCs {
			if v.ArchetypeID == sc.Archetype && !v.Unique {
				variants = append(variants, v)
			}
		}
		if len(variants) > 0 {
			npcConfig = variants[rand.Intn(len(variants))]
		}
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
	DebugLog("Dynamic Spawn: [%s] at (%.2f, %.2f) | Alignment: %v", npc.Name, npc.X, npc.Y, npc.Alignment)
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

	// 5% chance to check for elite variants in npcRegistry
	if rand.Float64() < 0.05 {
		var variants []*EntityConfig
		for _, v := range g.npcRegistry.NPCs {
			if v.ArchetypeID == sc.Archetype && !v.Unique {
				variants = append(variants, v)
			}
		}
		if len(variants) > 0 {
			npcConfig = variants[rand.Intn(len(variants))]
		}
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
	DebugLog("Dynamic Edge Spawn: [%s] at (%.2f, %.2f) | Alignment: %v", npc.Name, npc.X, npc.Y, npc.Alignment)
}

func (g *Game) openFilePicker() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("osascript", "-e", `POSIX path of (choose file with prompt "Select an .oinakos.yaml save file:" of type {"oinakos.yaml", "yaml"})`)
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	return ""
}
func (g *Game) updateProximityEffects() {
	if g.isPaused || g.isGameOver || g.isMapWon {
		return
	}

	for _, o := range g.obstacles {
		if !o.Alive || o.Archetype == nil {
			continue
		}

		// Process Entities: MainCharacter and NPCs
		entities := make([]interface{}, 0, len(g.npcs)+1)
		entities = append(entities, g.mainCharacter)
		for _, n := range g.npcs {
			if n.IsAlive() {
				entities = append(entities, n)
			}
		}

		for _, entity := range entities {
			var ex, ey float64
			var eFootprint engine.Polygon
			var isMC bool

			switch e := entity.(type) {
			case *MainCharacter:
				ex, ey = e.X, e.Y
				eFootprint = e.GetFootprint()
				isMC = true
			case *NPC:
				ex, ey = e.X, e.Y
				eFootprint = e.GetFootprint()
			default:
				continue
			}

			for _, action := range o.Archetype.Actions {
				inRange := false
				if action.Aura > 0 {
					dist := math.Sqrt(math.Pow(ex-o.X, 2) + math.Pow(ey-o.Y, 2))
					if dist <= action.Aura {
						inRange = true
					}
				} else {
					if engine.CheckCollision(eFootprint, o.GetFootprint()) {
						inRange = true
					}
				}

				if !inRange {
					continue
				}

				if action.Type == ActionHarm {
					if o.EffectTimers[entity] <= 0 {
						// Apply Damage
						switch e := entity.(type) {
						case *MainCharacter:
							e.TakeDamage(action.Amount, g.audio)
						case *NPC:
							e.TakeDamage(action.Amount, nil, nil, g.audio, g.npcs)
						}
						o.EffectTimers[entity] = 60
						g.floatingTexts = append(g.floatingTexts, &FloatingText{
							Text:  fmt.Sprintf("-%d", action.Amount),
							X:     ex,
							Y:     ey,
							Life:  45,
							Color: ColorHarm,
						})
					}
				} else if action.Type == ActionHeal && !action.RequiresInteraction {
					allowed := true
					if action.AlignmentLimit != "" && action.AlignmentLimit != "all" {
						var alignment Alignment
						if isMC {
							alignment = AlignmentAlly
						} else {
							alignment = entity.(*NPC).Alignment
						}
						if action.AlignmentLimit == "enemy" && alignment != AlignmentEnemy {
							allowed = false
						}
						if action.AlignmentLimit == "ally" && alignment != AlignmentAlly {
							allowed = false
						}
					}

					if allowed && o.EffectTimers[entity] <= 0 {
						switch e := entity.(type) {
						case *MainCharacter:
							e.Heal(action.Amount)
						case *NPC:
							e.Heal(action.Amount)
						}
						o.EffectTimers[entity] = 60
						g.floatingTexts = append(g.floatingTexts, &FloatingText{
							Text:  fmt.Sprintf("+%d", action.Amount),
							X:     ex,
							Y:     ey,
							Life:  45,
							Color: ColorHeal,
						})
					}
				}
			}
		}
	}
}
