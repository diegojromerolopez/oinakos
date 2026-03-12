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
	"sync/atomic"
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
	playableCharacter     *PlayableCharacter
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
	settingsFontIndex  int
	settingsAudioIndex int
	settingsFogIndex   int
	settingsMenuIndex  int

	onFontUpdate func(fontName string)

	lastMouseX, lastMouseY int
	isSettingsFromPause    bool

	ExploredTiles map[image.Point]bool

	isQuitConfirmationOpen bool
	quitConfirmationIndex  int // 0: Yes, 1: No

	menuHandler      *MenuHandler
	worldManager     *WorldManager
	mechanicsManager *MechanicsManager

	LoadingProgress int32 // 0 to 1000 representing 0.0 to 1.0
	LoadingMessage  string
}

func (g *Game) SetOnFontUpdate(cb func(string)) {
	g.onFontUpdate = cb
}

func (g *Game) UpdateFont() {
	if g.onFontUpdate != nil && g.settings != nil {
		g.onFontUpdate(g.settings.Font)
	}
}

func NewGame(assets fs.FS, initialMapID, initialMapTypeID, initialHeroID string, input engine.Input, audio AudioManager, debug bool) *Game {
	rand.Seed(time.Now().UnixNano())

	// Load playableCharacter config
	pConfig, err := LoadPlayableCharacterConfig(assets)
	if err != nil {
		log.Printf("Warning: failed to load playable character: %v. Using default values.", err)
	}

	playableCharacter := NewPlayableCharacter(0, 0, pConfig)
	pIsoX, pIsoY := engine.CartesianToIso(playableCharacter.X, playableCharacter.Y)

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
		playableCharacter:             playableCharacter,
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
		ExploredTiles:             make(map[image.Point]bool),
		LoadingProgress:           1000,
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
			g.playableCharacter.Config = config
			g.playableCharacter.Health = config.Stats.HealthMin
			g.playableCharacter.MaxHealth = config.Stats.HealthMin
			g.playableCharacter.Speed = config.Stats.Speed
			g.playableCharacter.BaseAttack = config.Stats.BaseAttack
			g.playableCharacter.BaseDefense = config.Stats.BaseDefense
			g.playableCharacter.Weapon = config.Weapon
			g.playableCharacter.Name = config.Name
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
				if g.initialHeroID == "" && g.playableCharacter.Config == nil {
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

	// Initialize Managers
	g.menuHandler = NewMenuHandler(g)
	g.worldManager = NewWorldManager(g)
	g.mechanicsManager = NewMechanicsManager(g)

	// Initial generation around playableCharacter
	g.worldManager.UpdateChunks()

	// Spawn NPCs if not loaded from instance and not in menu
	if !instanceLoaded && !g.isMainMenu {
		g.npcs = make([]*NPC, 0)
		if isTestingEnvironment {
			g.worldManager.LoadMapLevel()
		} else {
			go g.worldManager.LoadMapLevel()
		}
	}

	return g
}



func (g *Game) Update() error {
	if atomic.LoadInt32(&g.LoadingProgress) < 1000 {
		return nil
	}
	if g.input.IsKeyJustPressed(engine.KeyTab) {
		g.showBoundaries = !g.showBoundaries
		g.debug = g.showBoundaries
		SetDebugMode(g.debug)
	}

	// 1. Check if we are in a menu state
	if g.isQuitConfirmationOpen || g.isMainMenu || g.isSettingsScreen ||
		g.isCharacterSelect || g.isCampaignSelect || g.isGameWon ||
		g.isGameOver || g.isMapWon || g.isMenuOpen || g.loadDialogActive {
		return g.menuHandler.Update()
	}

	// 2. Gameplay input (Pause/QuickSave)
	if g.input.IsKeyJustPressed(engine.KeyEscape) {
		g.isMenuOpen = true
		return nil
	}
	if g.input.IsKeyJustPressed(engine.KeyQ) {
		g.performQuicksave()
	}

	// 3. World and Game Logic
	g.mechanicsManager.UpdateFogOfWar()
	g.worldManager.UpdateChunks()
	g.worldManager.UpdateNPCSpawning()

	// Update projectiles
	activeProjectiles := []*Projectile{}
	for _, p := range g.projectiles {
		p.Update(g.playableCharacter, g.obstacles, &g.floatingTexts, g.audio)
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

	g.playableCharacter.CurrentTile = g.currentMapType.GetTileAt(g.playableCharacter.X, g.playableCharacter.Y)
	g.playableCharacter.Update(g.input, g.audio, g.obstacles, g.npcs, &g.floatingTexts, g.currentMapType.MapWidth, g.currentMapType.MapHeight)

	// Position tracking and logging
	if g.playableCharacter.Tick%30 == 0 {
		g.logRealtimePosition()
	}
	os.WriteFile("/tmp/oinakos_pos.txt", []byte(fmt.Sprintf("%.2f,%.2f", g.playableCharacter.X, g.playableCharacter.Y)), 0644)

	for _, o := range g.obstacles {
		o.Update()
	}

	g.mechanicsManager.UpdateProximityEffects()

	// Check Win/Loss Conditions
	if g.mechanicsManager.CheckWinConditions() {
		if !g.isMapWon {
			DebugLog("Objective Completed! Level %d cleared. Objective: %v", g.mapLevel, g.currentMapType.Type)
		}
		g.isMapWon = true
		return nil
	}

	if !g.playableCharacter.IsAlive() {
		g.isGameOver = true
	}

	// Update all obstacles
	aliveObstacles := make([]*Obstacle, 0, len(g.obstacles))
	for _, o := range g.obstacles {
		o.Update()
		if o.Alive {
			aliveObstacles = append(aliveObstacles, o)
		}
	}
	g.obstacles = aliveObstacles

	// Update all NPCs
	for _, n := range g.npcs {
		n.CurrentTile = g.currentMapType.GetTileAt(n.X, n.Y)
		n.Update(g.playableCharacter, g.obstacles, g.npcs, &g.projectiles, &g.floatingTexts, g.currentMapType.MapWidth, g.currentMapType.MapHeight, g.audio)
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

	// Camera follows player
	pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
	g.camera.Follow(pIsoX, pIsoY, 0.1)

	// Final safety check
	g.ensurePlayerNotStuck()

	return nil
}

func (g *Game) logRealtimePosition() {
	isIllegal := g.playableCharacter.checkCollisionAt(g.playableCharacter.X, g.playableCharacter.Y, g.obstacles)
	status := "OK"
	if isIllegal {
		status = "ILLEGAL POSITION (INSIDE BUILDING)"
	}
	nearestDist := 999.0
	nearestName := "None"
	for _, o := range g.obstacles {
		dist := math.Sqrt(math.Pow(g.playableCharacter.X-o.X, 2) + math.Pow(g.playableCharacter.Y-o.Y, 2))
		if dist < nearestDist {
			nearestDist = dist
			if o.Archetype != nil {
				nearestName = o.Archetype.Name
			}
		}
	}
	DebugLog("[REALTIME] Player Pos: X=%.2f, Y=%.2f | Status: %s | Nearest: %s (Dist: %.2f)",
		g.playableCharacter.X, g.playableCharacter.Y, status, nearestName, nearestDist)
}

func (g *Game) ensurePlayerNotStuck() {
	for i := 0; i < 50; i++ {
		if !g.playableCharacter.checkCollisionAt(g.playableCharacter.X, g.playableCharacter.Y, g.obstacles) {
			break
		}
		g.playableCharacter.X += rand.Float64()*2 - 1
		g.playableCharacter.Y += rand.Float64()*2 - 1
		ncX, ncY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
		g.camera.SnapTo(ncX, ncY)
	}
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

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.width, g.height
}

// Extracted methods removed to reduce file size.
// functionality moved to MenuHandler, WorldManager, and MechanicsManager.
