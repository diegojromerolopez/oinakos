package game

import (
	"image"
	"log"

	"oinakos/internal/engine"
)

// DestroyProgress resets the game state to its initial menu-ready state.
func (g *Game) DestroyProgress() {
	g.isMainMenu = true
	g.isCharacterSelect = false
	g.isCampaignSelect = false
	g.isGameOver = false
	g.deathReason = ""
	g.isMapWon = false
	g.isGameWon = false
	g.isPaused = false
	g.isMenuOpen = false
	g.isInventoryOpen = false
	g.isQuitConfirmationOpen = false
	g.isAboutScreen = false

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

	g.characters = nil
	g.obstacles = nil
	g.projectiles = nil
	g.floatingTexts = nil
	g.ActiveDialogue = nil
	g.EventLog = nil
	g.LogScrollOffset = 0

	g.initialMapID = ""
	g.initialMapTypeID = ""
	g.initialHeroID = ""
	g.lastSavePath = ""
	g.saveMessage = ""
	g.saveMessageTimer = 0

	// Reset World
	if g.World != nil {
		g.World.Characters = nil
		g.World.Obstacles = nil
		g.World.Projectiles = nil
		g.World.FloatingTexts = nil
		g.World.ExploredTiles = make(map[image.Point]bool)
		g.World.PlayTime = 0
	}

	// Load default character config to reset hero selection
	pConfig, err := LoadPlayableCharacterConfig(g.assets)
	if err != nil {
		log.Printf("Warning: failed to reload playable character config: %v", err)
	}
	g.World.PlayableCharacter = NewCharacter(0, 0, pConfig, 1, true, g.Registries.Objects)
	g.playableCharacter = g.World.PlayableCharacter
	g.characters = nil
	g.World.Characters = nil

	// Reset camera
	pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
	g.camera = engine.NewCamera(pIsoX, pIsoY)

	// Reset current MapType to safe zone default
	if m, ok := g.mapTypeRegistry.Types["safe_zone"]; ok {
		g.currentMapType = *m
	} else if len(g.mapTypeRegistry.IDs) > 0 {
		g.currentMapType = *g.mapTypeRegistry.Types[g.mapTypeRegistry.IDs[0]]
	}
	if g.World != nil {
		g.World.CurrentMapType = &g.currentMapType
	}
}

// Restart resets the current map/session while preserving the chosen hero and campaign context.
func (g *Game) Restart() {
	g.isMainMenu = false
	g.isCharacterSelect = false
	g.isCampaignSelect = false
	g.isGameOver = false
	g.deathReason = ""
	g.isMapWon = false
	g.isGameWon = false
	g.isPaused = false
	g.isMenuOpen = false
	g.isInventoryOpen = false
	g.isQuitConfirmationOpen = false

	g.LoadingProgress = 1000
	g.LoadingMessage = ""

	g.playTime = 0
	g.npcSpawnTimer = 0

	g.generatedChunks = make(map[image.Point]bool)
	g.ExploredTiles = make(map[image.Point]bool)

	// Synchronously clear world lists to prevent old entities from being processed
	g.characters = nil
	g.obstacles = nil
	g.projectiles = nil
	g.floatingTexts = nil
	g.ActiveDialogue = nil
	g.EventLog = nil
	g.LogScrollOffset = 0

	g.saveMessage = ""
	g.saveMessageTimer = 0

	// Reset World
	if g.World != nil {
		g.World.Characters = nil
		g.World.Obstacles = nil
		g.World.Projectiles = nil
		g.World.FloatingTexts = nil
		g.World.ExploredTiles = make(map[image.Point]bool)
		g.World.PlayTime = 0
	}

	// Reload current MapType to reset any mutated fields (like TargetKills/StartTime)
	if g.currentMapType.ID != "" {
		if m, ok := g.mapTypeRegistry.Types[g.currentMapType.ID]; ok {
			g.currentMapType = *m
		}
	}

	// Re-initialize playableCharacter using initialHeroID (if set)
	if g.initialHeroID != "" {
		if config, ok := g.characterRegistry.Characters[g.initialHeroID]; ok {
			g.playableCharacter.Config = config
			rolledStats := config.Stats.Roll()
			rolledAttrs := config.Attributes.Roll()
			g.playableCharacter.RawStats = rolledStats
			g.playableCharacter.PrimaryAttributes = rolledAttrs
			g.playableCharacter.State.MaxHealthPoints = rolledStats.HealthMin
			g.playableCharacter.State.HealthPoints = rolledStats.HealthMin
			g.playableCharacter.Speed = rolledStats.Speed
			g.playableCharacter.BaseAttack = rolledStats.BaseAttack
			g.playableCharacter.BaseDefense = rolledStats.BaseDefense
			g.playableCharacter.Weapon = config.Weapon.Resolve(g.Registries.Objects)
			g.playableCharacter.ActionState = ActorIdle
			g.playableCharacter.Tick = 0
			g.playableCharacter.DeadTimer = 0
			g.playableCharacter.HitTimer = 0
			g.playableCharacter.Kills = 0
			g.playableCharacter.MapKills = make(map[string]int)
			g.playableCharacter.XP = 0
			g.playableCharacter.Level = 1
			g.playableCharacter.Name = config.Name
			
			if g.currentMapType.Player != nil {
				g.playableCharacter.X = g.currentMapType.Player.X
				g.playableCharacter.Y = g.currentMapType.Player.Y
			} else {
				g.playableCharacter.X, g.playableCharacter.Y = 0, 0
			}
		}
	} else {
		pConfig, _ := LoadPlayableCharacterConfig(g.assets)
		g.playableCharacter = NewCharacter(0, 0, pConfig, 1, true, g.Registries.Objects)
	}

	if g.World != nil {
		g.World.PlayableCharacter = g.playableCharacter
		g.World.ExploredTiles = g.ExploredTiles
		g.World.CurrentMapType = &g.currentMapType
	}

	// Reset camera
	pIsoX, pIsoY := engine.CartesianToIso(g.playableCharacter.X, g.playableCharacter.Y)
	g.camera.SnapTo(pIsoX, pIsoY)

	// Trigger map loading asynchronously
	go g.worldManager.LoadMapLevel()
}
