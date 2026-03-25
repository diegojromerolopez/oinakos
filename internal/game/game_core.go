package game

import (
	"image"
	"io/fs"

	"oinakos/internal/engine"
)

const (
	WinMenuContinue = 0
	WinMenuQuit     = 1
)

type WeatherType int

const (
	WeatherClear WeatherType = iota
	WeatherRain
	WeatherSnow
	WeatherStorm
)

type Game struct {
	width, height     int
	playableCharacter     *Character
	playerConfig      *EntityConfig
	obstacles         []*Obstacle
	characters        []*Character
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
	isAboutScreen     bool      // True if showing about screen
	isSettingsScreen  bool      // True if showing settings screen
	isKeymapScreen    bool      // True if showing keymap screen
	isInventoryOpen   bool      // True if showing inventory screen
	ActiveBook        *ItemInstance // True if currently reading a book
	isCampaignSelect  bool      // True if showing campaign picker
	campaignMenuIndex int       // Index of selected campaign
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

	World      *World
	Registries *RegistryContainer

	menuHandler      *MenuHandler
	worldManager     *WorldManager
	mechanicsManager *MechanicsManager

	LoadingProgress int32 // 0 to 1000 representing 0.0 to 1.0
	LoadingMessage  string
	Version         string

	ActiveDialogue *DialogueState
	EventLog       []LogEntry
	LogScrollOffset int
	IsDraggingLog   bool
	LogUIState      DialogueUIState

	aiManager *AIManager
	Graphics  engine.Graphics
	silhouetteBuffer engine.Image

	availableModels   []string
	isFetchingModels  bool

	CurrentWeather   WeatherType
	WeatherIntensity float64
	particles        []*Particle
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
