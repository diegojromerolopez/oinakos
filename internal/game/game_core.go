package game

import (
	"image"
	"io/fs"
	"sync/atomic"

	"oinakos/internal/engine"
)

var GlobalHeadlessLogger func(string)

func SetGlobalHeadlessLogger(f func(string)) {
	GlobalHeadlessLogger = f
}


const (
	WinMenuContinue = 0
	WinMenuQuit     = 1
)

type Game struct {
	width, height     int
	playableCharacter *Character
	playerConfig      *EntityConfig
	obstacles         []*Obstacle
	characters        []*Character
	projectiles       []*Projectile
	isGameOver        bool
	isMapWon          bool
	isGameWon         bool
	mapWonMenuIndex   int
	isPaused          bool
	currentMapType    MapType
	mapLevel          int
	currentCampaign   *Campaign
	campaignIndex     int
	isCampaign        bool
	isMainMenu        bool
	mainMenuIndex     int
	isAboutScreen     bool
	isSettingsScreen  bool
	isKeymapScreen    bool
	isInventoryOpen   bool
	ActiveBook        *ItemInstance
	isCampaignSelect  bool
	campaignMenuIndex int
	keymapSelectedIndex int
	remappingAction   string
	initialMapTypeID  string
	debug             bool

	generatedChunks map[image.Point]bool
	npcSpawnTimer   int
	playTime        float64
	Tick            int

	camera *engine.Camera
	assets fs.FS

	floatingTexts             []*FloatingText
	archetypeRegistry         *ArchetypeRegistry
	characterRegistry         *CharacterRegistry
	mapTypeRegistry           *MapTypeRegistry
	campaignRegistry          *CampaignRegistry
	obstacleRegistry          *ObstacleRegistry
	initialMapID              string
	initialHeroID             string
	lastSavePath              string
	input                     engine.Input
	showBoundaries            bool
	audio                     AudioManager

	isMenuOpen       bool
	menuIndex        int
	loadDialogActive bool
	loadPathInput    string

	isCharacterSelect  bool
	characterMenuIndex int
	saveMessage        string
	saveMessageTimer   int

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
	quitConfirmationIndex  int

	World      *World
	Registries *RegistryContainer

	menuHandler      *MenuHandler
	worldManager     *WorldManager
	mechanicsManager *MechanicsManager

	LoadingProgress int32
	LoadingMessage  string
	Version         string

	ActiveDialogue *DialogueState
	EventLog       []LogEntry
	LogScrollOffset int
	IsDraggingLog   bool
	LogUIState      DialogueUIState

	ActiveTrader    *Character
	isTradeOpen     bool

	aiManager *AIManager
	Graphics  engine.Graphics
	silhouetteBuffer engine.Image

	availableModels   []string
	isFetchingModels  bool

	pinnedCharacter  *Character
	pinnedUIX, pinnedUIY int
	isDraggingPinnedUI   bool
	dragPinnedOffsetX, dragPinnedOffsetY int
	particles        []*Particle
	deathReason      string
	simulationMode   bool
	isHUDVisible     bool
}

func (g *Game) ToggleHUD() {
	g.isHUDVisible = !g.isHUDVisible
}

func (g *Game) SetSimulationMode(enabled bool) {
	g.simulationMode = enabled
	if enabled {
		atomic.StoreInt32(&g.LoadingProgress, 1000)
	}
	if g.settings != nil {
		g.settings.AISimulationMode = enabled
	}
}

func (g *Game) SetOnFontUpdate(cb func(string)) {
	g.onFontUpdate = cb
}

func (g *Game) GetSilhouetteBuffer() engine.Image {
	return g.silhouetteBuffer
}

func (g *Game) UpdateFont() {
	if g.onFontUpdate != nil && g.settings != nil {
		g.onFontUpdate(g.settings.Font)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.width, g.height
}

func (g *Game) GetContext() *SystemContext {
	return &SystemContext{
		World:      g.World,
		Input:      g.input,
		Audio:      g.audio,
		Registries: g.Registries,
		Log:        g.LogEvent,
		AIManager:  g.aiManager,
		Weather:    g.World.State.Weather,
		Intensity:  g.World.State.Intensity,
		Settings:   g.settings,
	}
}

func (g *Game) BypassMenu() {
	g.isMainMenu = false
	g.isCharacterSelect = false
	g.isCampaignSelect = false
	g.isQuitConfirmationOpen = false
	g.isAboutScreen = false
	g.isSettingsScreen = false
}

func (g *Game) GetAIManager() *AIManager {
	return g.aiManager
}

func (g *Game) SetAIManager(m *AIManager) {
	g.aiManager = m
}

var isTestingEnvironment = false

func SetTestingMode(active bool) {
	isTestingEnvironment = active
}

func (g *Game) GetDeathReason() string {
	return g.deathReason
}

func (g *Game) IsGameOver() bool {
	return g.isGameOver
}

func (g *Game) GetPlayableCharacter() *Character {
	return g.playableCharacter
}

func (g *Game) SetCurrentMapType(m *MapType) {
	if m == nil { return }
	g.currentMapType = *m
	if g.World != nil {
		g.World.CurrentMapType = &g.currentMapType
	}
}

func (g *Game) TriggerMapLoad() {
	if g.worldManager != nil {
		g.worldManager.LoadMapLevel()
	}
}
