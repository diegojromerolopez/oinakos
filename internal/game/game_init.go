package game

import (
	"fmt"
	"image"
	"io/fs"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"oinakos/internal/engine"
)

func NewGame(assets fs.FS, graphics engine.Graphics, initialMapID, initialMapTypeID, initialHeroID string, input engine.Input, audio AudioManager, debug bool, version string) *Game {
	rand.Seed(time.Now().UnixNano())

	// Load playableCharacter config
	pConfig, err := LoadPlayableCharacterConfig(assets)
	if err != nil {
		log.Printf("Warning: failed to load playable character: %v. Using default values.", err)
	}

	// Registries
	archetypeRegistry := NewArchetypeRegistry()
	archetypeRegistry.LoadAll(assets)

	characterRegistry := NewCharacterRegistry()
	characterRegistry.LoadAll(assets)

	mapTypeRegistry := NewMapTypeRegistry()
	mapTypeRegistry.LoadAll(assets)

	campaignRegistry := NewCampaignRegistry()
	campaignRegistry.LoadAll(assets)

	obstacleRegistry := NewObstacleRegistry()
	obstacleRegistry.LoadAll(assets)

	objectRegistry := NewObjectRegistry()
	objectRegistry.LoadAll(assets)

	// Process inheritance AFTER all registries are loaded
	characterRegistry.ProcessInheritance(archetypeRegistry)

	var selectedMapType MapType
	if m, ok := mapTypeRegistry.Types["safe_zone"]; ok {
		selectedMapType = *m
	} else if len(mapTypeRegistry.IDs) > 0 {
		selectedMapType = *mapTypeRegistry.Types[mapTypeRegistry.IDs[0]]
	}

	// Create playable character placeholder
	playableCharacter := NewCharacter(0, 0, pConfig, 1, true, objectRegistry)
	pIsoX, pIsoY := engine.CartesianToIso(playableCharacter.X, playableCharacter.Y)

	g := &Game{
		width:             1280,
		height:            720,
		camera:            engine.NewCamera(pIsoX, pIsoY),
		assets:            assets,
		input:             input,
		audio:             audio,
		debug:             debug,
		initialMapID:     initialMapID,
		initialMapTypeID: initialMapTypeID,
		initialHeroID:    initialHeroID,
		LoadingProgress:  1000,
		Version:          version,
		Graphics:         graphics,
		silhouetteBuffer: graphics.NewImage(1280, 720),
		mapLevel:         1,
	}

	g.Registries = &RegistryContainer{
		Archetypes: archetypeRegistry,
		Characters: characterRegistry,
		Maps:       mapTypeRegistry,
		Campaigns:  campaignRegistry,
		Obstacles:  obstacleRegistry,
		Objects:    objectRegistry,
	}

	g.World = &World{
		PlayableCharacter: playableCharacter,
		Characters:        nil,
		Obstacles:         nil,
		Projectiles:       nil,
		FloatingTexts:     nil,
		CurrentMapType:    &selectedMapType,
		ExploredTiles:     make(map[image.Point]bool),
		Items:             nil,
	}

	g.playableCharacter = playableCharacter
	g.archetypeRegistry = archetypeRegistry
	g.mapTypeRegistry = mapTypeRegistry
	g.campaignRegistry = campaignRegistry
	g.obstacleRegistry = obstacleRegistry
	g.characterRegistry = characterRegistry
	g.currentMapType = selectedMapType
	g.ExploredTiles = g.World.ExploredTiles
	g.generatedChunks = make(map[image.Point]bool)
	g.World.Game = g

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
		if config, ok := g.characterRegistry.Characters[g.initialHeroID]; ok {
			g.playableCharacter.Config = config
			rolledStats := config.Stats.Roll()
			rolledAttrs := config.Attributes.Roll()
			g.playableCharacter.RawStats = rolledStats
			g.playableCharacter.PrimaryAttributes = rolledAttrs
			
			// Use robust HP initialization (mirrors NewCharacter)
			maxHP := rolledStats.HealthMin
			if maxHP <= 0 {
				maxHP = rolledAttrs.Health * 10
			}
			if maxHP < 100 { maxHP = 100 }

			g.playableCharacter.TemporalState.MaxHealthPoints = maxHP
			g.playableCharacter.TemporalState.HealthPoints = maxHP
			
			g.playableCharacter.Speed = rolledStats.Speed
			g.playableCharacter.BaseAttack = rolledStats.BaseAttack
			g.playableCharacter.BaseDefense = rolledStats.BaseDefense
			g.playableCharacter.Weapon = config.Weapon.Resolve(g.Registries.Objects)
			g.playableCharacter.Name = config.Name
			g.isCharacterSelect = false
			log.Printf("Using initial hero: %s (HP=%d)", g.initialHeroID, maxHP)
		} else {
			log.Printf("Warning: initial hero %s not found in registry", g.initialHeroID)
		}
	}

	// Trigger map loading if requested
	instanceLoaded := false
	if g.initialMapID != "" {
		// Define search candidates in order of priority
		candidates := []string{
			g.initialMapID,
			fmt.Sprintf("data/maps/%s", g.initialMapID),
		}

		if !strings.HasSuffix(g.initialMapID, ".yaml") && !strings.HasSuffix(g.initialMapID, ".yml") {
			candidates = append(candidates,
				g.initialMapID+".yaml",
				g.initialMapID+".yml",
				fmt.Sprintf("data/maps/%s.yaml", g.initialMapID),
				fmt.Sprintf("data/maps/%s.yml", g.initialMapID),
			)
		}

		for _, path := range candidates {
			if err := g.Load(path); err == nil {
				log.Printf("Loaded map instance: %s | Closing Menu", path)
				instanceLoaded = true
				g.isMainMenu = false
				if g.initialHeroID == "" && g.playableCharacter.Config == nil {
					g.isCharacterSelect = true
				} else {
					g.isCharacterSelect = false
				}
				break
			}
		}

		if !instanceLoaded {
			if m, ok := g.mapTypeRegistry.Types[g.initialMapID]; ok {
				g.currentMapType = *m
				g.isMainMenu = false
				g.isCharacterSelect = true
				log.Printf("Using initial map type: %s", g.initialMapID)
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

	g.menuHandler = NewMenuHandler(g)
	g.LogEvent("Welcome to Oinakos.", LogInfo)
	g.LogEvent(fmt.Sprintf("Playing as %s. Follow your destiny.", g.playableCharacter.Name), LogInfo)
	
	g.worldManager = NewWorldManager(g)
	g.mechanicsManager = NewMechanicsManager(g)
	g.initAIManager()
	g.worldManager.UpdateChunks()

	if !instanceLoaded && !g.isMainMenu {
		g.characters = make([]*Character, 0)
		if isTestingEnvironment {
			g.worldManager.LoadMapLevel()
		} else {
			go g.worldManager.LoadMapLevel()
		}
	}

	if input != nil {
		input.SetCursorMode(engine.CursorModeHidden)
	}

	return g
}
