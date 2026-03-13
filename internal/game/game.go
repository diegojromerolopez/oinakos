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

	ActiveDialogue *DialogueState
	EventLog       []LogEntry
	LogScrollOffset int
	IsDraggingLog   bool
	LogUIState      DialogueUIState
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
	g.LogEvent("Welcome to Oinakos.", LogInfo)
	g.LogEvent(fmt.Sprintf("Playing as %s. Follow your destiny.", g.playableCharacter.Name), LogInfo)
	
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

// DestroyProgress resets the game state to its initial menu-ready state.
func (g *Game) DestroyProgress() {
	g.isMainMenu = true
	g.isCharacterSelect = false
	g.isCampaignSelect = false
	g.isGameOver = false
	g.isMapWon = false
	g.isGameWon = false
	g.isPaused = false
	g.isMenuOpen = false
	g.isQuitConfirmationOpen = false

	g.LoadingProgress = 1000
	g.LoadingMessage = ""

	g.mainMenuIndex = 0
	g.characterMenuIndex = 0
	g.campaignMenuIndex = 0
	g.menuIndex = 0
	g.mapWonMenuIndex = 0

	g.currentCampaign = nil
	g.campaignIndex = 0
	g.isCampaign = false
	g.mapLevel = 1
	g.playTime = 0
	g.npcSpawnTimer = 0

	g.generatedChunks = make(map[image.Point]bool)
	g.ExploredTiles = make(map[image.Point]bool)

	g.npcs = nil
	g.obstacles = nil
	g.projectiles = nil
	g.floatingTexts = nil
	g.ActiveDialogue = nil
	g.EventLog = nil

	g.initialMapID = ""
	g.initialMapTypeID = ""
	g.initialHeroID = ""
	g.lastSavePath = ""
	g.saveMessage = ""
	g.saveMessageTimer = 0

	// Load default character config to reset hero selection
	pConfig, err := LoadPlayableCharacterConfig(g.assets)
	if err != nil {
		log.Printf("Warning: failed to reload playable character config: %v", err)
	}
	g.playableCharacter = NewPlayableCharacter(0, 0, pConfig)

	// Reset camera
	pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
	g.camera = engine.NewCamera(pIsoX, pIsoY)

	// Reset current MapType to safe zone default
	if m, ok := g.mapTypeRegistry.Types["safe_zone"]; ok {
		g.currentMapType = *m
	} else if len(g.mapTypeRegistry.IDs) > 0 {
		g.currentMapType = *g.mapTypeRegistry.Types[g.mapTypeRegistry.IDs[0]]
	}
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

	// 0. Update Dialogue & Log UI toggles
	// We use 'L' or some other key to minimize/maximize logic if needed,
	// but the plan says it can be maximized/minimized.
	// Let's use 'TAB' for debug and maybe 'BACKSPACE' for UI toggle if not in dialogue.
	if g.ActiveDialogue != nil {
		if g.input.IsKeyJustPressed(engine.KeyW) || g.input.IsKeyJustPressed(engine.KeyUp) {
			g.ActiveDialogue.SelectedChoice--
			if g.ActiveDialogue.SelectedChoice < 0 {
				g.ActiveDialogue.SelectedChoice = len(g.ActiveDialogue.Choices) - 1
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyS) || g.input.IsKeyJustPressed(engine.KeyDown) {
			g.ActiveDialogue.SelectedChoice++
			if g.ActiveDialogue.SelectedChoice >= len(g.ActiveDialogue.Choices) {
				g.ActiveDialogue.SelectedChoice = 0
			}
		}
		if g.input.IsKeyJustPressed(engine.KeyEnter) {
			g.AdvanceDialogue()
		}
		if g.input.IsKeyJustPressed(engine.KeyEscape) || g.input.IsKeyJustPressed(engine.KeyBackspace) {
			g.CloseDialogue()
			return nil // Return early to prevent the menu from opening in the same frame
		}
	} else {
		// Log UI minimizing toggle? For now let's just use clicks and proximity.
		g.handleDialogueInput()
		g.handleDialogueProximity()
		g.handleLogScrolling()

		// 2. Gameplay input (Pause/QuickSave)
		if g.input.IsKeyJustPressed(engine.KeyEscape) {
			g.isMenuOpen = true
			return nil
		}
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
	g.playableCharacter.Update(g.input, g.audio, g.obstacles, g.npcs, &g.floatingTexts, g.currentMapType.MapWidth, g.currentMapType.MapHeight, g.archetypeRegistry, g.LogEvent)

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
			n.Update(g.playableCharacter, g.obstacles, g.npcs, &g.projectiles, &g.floatingTexts, g.currentMapType.MapWidth, g.currentMapType.MapHeight, g.audio, g.LogEvent, g.archetypeRegistry)
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
func (g *Game) LogEvent(text string, category LogCategory) {
	entry := LogEntry{
		Text:     text,
		Ticks:    0,
		Category: category,
	}
	if g.playableCharacter != nil {
		entry.Ticks = g.playableCharacter.Tick
	}
	g.EventLog = append(g.EventLog, entry)
	g.LogScrollOffset = 0

	if len(g.EventLog) > 100 {
		g.EventLog = g.EventLog[len(g.EventLog)-100:]
	}
}

func (g *Game) handleDialogueInput() {
	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		mx, my := g.input.MousePosition()
		
		// Check for maximize/minimize by clicking anywhere in the box
		isDialogue := g.ActiveDialogue != nil
		boxH := 300
		if isDialogue {
			if g.ActiveDialogue.UIState == DialogueMaximized {
				boxH = 600
			}
		} else {
			boxH = 60
		}
		
		bx, by := 10, g.height-boxH-10
		boxW := g.width - 20
		
		// Ignore dialogue toggle if clicking on the scrollbar area (right 20px of the box)
		if mx >= bx && mx <= bx+boxW-20 && my >= by && my <= by+boxH {
			g.ToggleDialogueSize()
			return
		}

		offX, offY := g.camera.GetOffsets(g.width, g.height)
		isoX := float64(mx) - offX
		isoY := float64(my) - offY
		cartX, cartY := engine.IsoToCartesian(isoX, isoY)

		for _, n := range g.npcs {
			if !n.IsAlive() || n.Alignment == AlignmentEnemy {
				continue
			}
			if n.GetFootprint().Contains(cartX, cartY) {
				g.InitiateDialogue(n)
				break
			}
		}
	}
}

func (g *Game) handleLogScrolling() {
	mx, my := g.input.MousePosition()
	isDialogue := g.ActiveDialogue != nil
	boxH := 300
	if isDialogue && g.ActiveDialogue.UIState == DialogueMaximized {
		boxH = 600
	} else if !isDialogue {
		boxH = 60
	}
	
	bx, by := 10, g.height-boxH-10
	boxW := g.width - 20
	
	// Track coordinates from renderer
	sbX := bx + boxW - 10
	sbTrackY := by + 5
	sbTrackH := boxH - 10

	// Handle Drag Start
	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		// A bit wider hit area for the scrollbar (10px instead of 4px)
		if mx >= sbX-5 && mx <= sbX+10 && my >= sbTrackY && my <= sbTrackY+sbTrackH {
			g.IsDraggingLog = true
		}
	}

	// Handle Drag Process
	if g.IsDraggingLog {
		if !g.input.IsMouseButtonPressed(engine.MouseButtonLeft) {
			g.IsDraggingLog = false
		} else {
			// Calculate ratio within track
			ratio := 1.0 - float32(my-sbTrackY)/float32(sbTrackH)
			if ratio < 0 { ratio = 0 }
			if ratio > 1 { ratio = 1 }
			
			var maxLogEntries int
			if isDialogue {
				maxLogEntries = 1
			} else {
				maxLogEntries = 2
			}
			
			maxOffset := len(g.EventLog) - maxLogEntries
			if maxOffset < 0 { maxOffset = 0 }
			
			g.LogScrollOffset = int(float32(maxOffset) * ratio)
		}
	}

	// Handle Wheel Scrolling
	_, wheelY := g.input.Wheel()
	if wheelY != 0 {
		if mx >= bx && mx <= bx+boxW && my >= by && my <= by+boxH {
			g.LogScrollOffset -= int(wheelY)
			if g.LogScrollOffset < 0 {
				g.LogScrollOffset = 0
			}
			maxScroll := len(g.EventLog) - 1
			if maxScroll < 0 {
				maxScroll = 0
			}
			if g.LogScrollOffset > maxScroll {
				g.LogScrollOffset = maxScroll
			}
		}
	}
}

func (g *Game) ToggleDialogueSize() {
	if g.ActiveDialogue != nil {
		if g.ActiveDialogue.UIState == DialogueMinimized {
			g.ActiveDialogue.UIState = DialogueMaximized
		} else {
			g.ActiveDialogue.UIState = DialogueMinimized
		}
		return
	}
	
	// Toggle standard Log UI if no dialogue
	if g.LogUIState == DialogueMinimized {
		g.LogUIState = DialogueMaximized
	} else {
		g.LogUIState = DialogueMinimized
	}
}

func (g *Game) handleDialogueProximity() {
	if g.ActiveDialogue != nil {
		return
	}
	for _, n := range g.npcs {
		if !n.IsAlive() || n.Alignment == AlignmentEnemy || n.HasInitiatedDialogue {
			continue
		}
		if n.Archetype != nil && n.Archetype.Dialogues != nil {
			for _, s := range n.Archetype.Dialogues.StartScenarios {
				if s.AutoInitiate {
					dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
					if dist < s.ProximityRange {
						g.InitiateDialogue(n)
						n.HasInitiatedDialogue = true
						break
					}
				}
			}
		}
	}
}

func (g *Game) InitiateDialogue(npc *NPC) {
	DebugLog("InitiateDialogue: Attempting to talk to NPC %s (Alignment: %v)", npc.Name, npc.Alignment)
	if npc.Archetype == nil {
		DebugLog("InitiateDialogue FAILED: NPC %s has no archetype", npc.Name)
		return
	}
	if npc.Archetype.Dialogues == nil {
		DebugLog("InitiateDialogue FAILED: NPC %s (Archetype: %s) has no dialogues block", npc.Name, npc.Archetype.ID)
		return
	}

	dr := npc.Archetype.Dialogues
	DebugLog("InitiateDialogue: Found %d start scenarios for %s", len(dr.StartScenarios), npc.Name)
	greeting := dr.PickGreeting()
	g.LogEvent(fmt.Sprintf("%s: %s", g.playableCharacter.Name, greeting), LogPlayer)

	start := dr.PickStart()
	if start == nil {
		return
	}

	g.ActiveDialogue = &DialogueState{
		SpeakerNPC:   npc,
		CurrentText:  start.Text,
		Choices:      start.Choices,
		IsActive:     true,
		UIState:      DialogueMaximized,
	}

	if start.Next != "" {
		if node, ok := dr.Nodes[start.Next]; ok {
			g.ActiveDialogue.Choices = node.Choices
		}
	}

	g.LogEvent(fmt.Sprintf("%s: %s", npc.Name, start.Text), LogNPC)
}

func (g *Game) AdvanceDialogue() {
	if g.ActiveDialogue == nil || len(g.ActiveDialogue.Choices) == 0 {
		g.CloseDialogue()
		return
	}

	choice := g.ActiveDialogue.Choices[g.ActiveDialogue.SelectedChoice]
	g.LogEvent(fmt.Sprintf("%s: %s", g.playableCharacter.Name, choice.Text), LogPlayer)

	// Apply effects
	for _, effect := range choice.Effects {
		g.ApplyDialogueEffect(g.ActiveDialogue.SpeakerNPC, effect)
	}

	if choice.Next == "" || choice.Next == "exit" {
		g.CloseDialogue()
		return
	}

	dr := g.ActiveDialogue.SpeakerNPC.Archetype.Dialogues
	if node, ok := dr.Nodes[choice.Next]; ok {
		g.ActiveDialogue.CurrentText = node.Text
		g.ActiveDialogue.Choices = node.Choices
		g.ActiveDialogue.SelectedChoice = 0
		g.LogEvent(fmt.Sprintf("%s: %s", g.ActiveDialogue.SpeakerNPC.Name, node.Text), LogNPC)
	} else {
		g.CloseDialogue()
	}
}

func (g *Game) CloseDialogue() {
	g.ActiveDialogue = nil
}

func (g *Game) ApplyDialogueEffect(npc *NPC, effect DialogueEffect) {
	switch effect.Type {
	case "change_alignment":
		switch effect.Value {
		case "enemy":
			npc.Alignment = AlignmentEnemy
			npc.TargetActor = &g.playableCharacter.Actor
			npc.Behavior = BehaviorKnightHunter
		case "neutral":
			npc.Alignment = AlignmentNeutral
			npc.TargetActor = nil
			npc.Behavior = BehaviorWander
		case "ally":
			npc.Alignment = AlignmentAlly
			npc.TargetActor = nil
			npc.Behavior = BehaviorEscort // Fixed from BehaviorFollow
		}
	case "change_behavior":
		switch effect.Value {
		case "flee":
			npc.Behavior = BehaviorFlee
		case "wander":
			npc.Behavior = BehaviorWander
		case "patrol":
			npc.Behavior = BehaviorPatrol
		case "follow":
			npc.Behavior = BehaviorEscort // Close enough for now
		}
	}
}
