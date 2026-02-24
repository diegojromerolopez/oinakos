package game

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"oinakos/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	width, height int
	renderer      *engine.Renderer
	player        *Player
	playerConfig  *EntityConfig
	obstacles     []*Obstacle
	npcs          []*NPC
	isGameOver    bool
	isPaused      bool
	currentMap    MapData
	mapLevel      int

	generatedChunks map[image.Point]bool
	npcSpawnTimer   int
	playTime        float64

	// Assets
	rockSprite        *ebiten.Image
	treeSprite        *ebiten.Image
	tree2Sprite       *ebiten.Image
	tree3Sprite       *ebiten.Image
	tree4Sprite       *ebiten.Image
	tree5Sprite       *ebiten.Image
	tree6Sprite       *ebiten.Image
	tree7Sprite       *ebiten.Image
	bushSprite        *ebiten.Image
	grassSprite       *ebiten.Image
	houseSprite       *ebiten.Image
	house2Sprite      *ebiten.Image
	house3Sprite      *ebiten.Image
	houseBurnedSprite *ebiten.Image
	templeSprite      *ebiten.Image
	smitherySprite    *ebiten.Image
	farmSprite        *ebiten.Image
	warehouseSprite   *ebiten.Image
	rock2Sprite       *ebiten.Image
	rock3Sprite       *ebiten.Image
	rock4Sprite       *ebiten.Image
	rock5Sprite       *ebiten.Image
	bush2Sprite       *ebiten.Image
	bush3Sprite       *ebiten.Image
	bush4Sprite       *ebiten.Image
	bush5Sprite       *ebiten.Image

	camera *engine.Camera
	assets fs.FS

	floatingTexts    []*FloatingText
	npcRegistry      *NPCRegistry
	mapRegistry      *MapRegistry
	obstacleRegistry *ObstacleRegistry
	portalSprite     *ebiten.Image
	crownSprite      *ebiten.Image
	zoneSprite       *ebiten.Image
}

func loadSprite(assets fs.FS, path string, removeBg bool) *ebiten.Image {
	var f fs.File
	var err error

	if assets != nil {
		f, err = assets.Open(path)
	}

	if err != nil || assets == nil {
		f, err = os.Open(path)
	}

	if err != nil {
		log.Printf("Warning: failed to open sprite '%s': %v", path, err)
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Warning: failed to decode sprite '%s': %v", path, err)
		return nil
	}

	if removeBg {
		img = engine.Transparentize(img)
	}

	return ebiten.NewImageFromImage(img)
}

func loadSpriteWithFallback(assets fs.FS, path string, removeBg bool, fallback *ebiten.Image) *ebiten.Image {
	img := loadSprite(assets, path, removeBg)
	if img == nil {
		return fallback
	}
	return img
}

func NewGame(assets fs.FS) *Game {
	rand.Seed(time.Now().UnixNano())

	engine.InitAudio(assets)
	// Load sounds (placeholders for now)
	// Load sounds
	engine.GlobalAudio.LoadSound("knight_attack", "assets/audio/knight/knight_attack.wav")
	engine.GlobalAudio.LoadSound("knight_hit", "assets/audio/knight/knight_hit.wav")
	engine.GlobalAudio.LoadSound("orc_hit", "assets/audio/orc/orc_hit.wav")
	engine.GlobalAudio.LoadSound("demon_hit", "assets/audio/demon/demon_hit.wav")
	engine.GlobalAudio.LoadSound("peasant_hit", "assets/audio/peasant/peasant_hit.wav")

	// Menace lines
	engine.GlobalAudio.LoadSound("orc_menace_1", "assets/audio/orc/orc_menace_1.wav")
	engine.GlobalAudio.LoadSound("orc_menace_2", "assets/audio/orc/orc_menace_2.wav")
	engine.GlobalAudio.LoadSound("orc_menace_3", "assets/audio/orc/orc_menace_3.wav")
	engine.GlobalAudio.LoadSound("orc_menace_4", "assets/audio/orc/orc_menace_4.wav")
	engine.GlobalAudio.LoadSound("orc_menace_5", "assets/audio/orc/orc_menace_5.wav")

	engine.GlobalAudio.LoadSound("demon_menace_1", "assets/audio/demon/demon_menace_1.wav")
	engine.GlobalAudio.LoadSound("demon_menace_2", "assets/audio/demon/demon_menace_2.wav")
	engine.GlobalAudio.LoadSound("demon_menace_3", "assets/audio/demon/demon_menace_3.wav")
	engine.GlobalAudio.LoadSound("demon_menace_4", "assets/audio/demon/demon_menace_4.wav")
	engine.GlobalAudio.LoadSound("demon_menace_5", "assets/audio/demon/demon_menace_5.wav")

	engine.GlobalAudio.LoadSound("peasant_menace_1", "assets/audio/peasant/peasant_menace_1.wav")
	engine.GlobalAudio.LoadSound("peasant_menace_2", "assets/audio/peasant/peasant_menace_2.wav")
	engine.GlobalAudio.LoadSound("peasant_menace_3", "assets/audio/peasant/peasant_menace_3.wav")
	engine.GlobalAudio.LoadSound("peasant_menace_4", "assets/audio/peasant/peasant_menace_4.wav")
	engine.GlobalAudio.LoadSound("peasant_menace_5", "assets/audio/peasant/peasant_menace_5.wav")

	// Death cries
	engine.GlobalAudio.LoadSound("knight_death", "assets/audio/knight/knight_death.wav")
	engine.GlobalAudio.LoadSound("orc_death", "assets/audio/orc/orc_death.wav")
	engine.GlobalAudio.LoadSound("demon_death", "assets/audio/demon/demon_death.wav")
	engine.GlobalAudio.LoadSound("peasant_death", "assets/audio/peasant/peasant_death.wav")

	// Load player config
	pConfig, err := LoadPlayerConfig(assets)
	if err != nil {
		log.Printf("Error loading player config: %v", err)
	}

	player := NewPlayer(0, 0, pConfig)
	pIsoX, pIsoY := engine.CartesianToIso(player.X, player.Y)

	// Select map based on difficulty level (default 1)
	mapLevel := 1
	var selectedMap MapData

	mapRegistry := NewMapRegistry()
	if err := mapRegistry.LoadAll(assets); err != nil {
		log.Printf("Error loading Map configs: %v", err)
	}

	if len(mapRegistry.IDs) > 0 {
		selectedMap = *mapRegistry.Maps[mapRegistry.IDs[0]]
	}

	// Load base tree sprites for fallbacks
	t1 := loadSprite(assets, "assets/images/environment/tree1.png", true)
	t2 := loadSprite(assets, "assets/images/environment/tree2.png", true)

	g := &Game{
		width:             1280,
		height:            720,
		renderer:          engine.NewRenderer(),
		player:            player,
		rockSprite:        loadSprite(assets, "assets/images/environment/rock1.png", true),
		treeSprite:        t1,
		tree2Sprite:       t2,
		tree3Sprite:       loadSpriteWithFallback(assets, "assets/images/environment/tree3.png", true, t1),
		tree4Sprite:       loadSpriteWithFallback(assets, "assets/images/environment/tree4.png", true, t2),
		tree5Sprite:       loadSpriteWithFallback(assets, "assets/images/environment/tree5.png", true, t1),
		tree6Sprite:       loadSpriteWithFallback(assets, "assets/images/environment/tree6.png", true, t2),
		tree7Sprite:       loadSpriteWithFallback(assets, "assets/images/environment/tree7.png", true, t1),
		bushSprite:        loadSprite(assets, "assets/images/environment/bush1.png", true),
		grassSprite:       loadSprite(assets, "assets/images/environment/grass_tile.png", true),
		houseSprite:       loadSprite(assets, "assets/images/environment/house1.png", true),
		house2Sprite:      loadSprite(assets, "assets/images/environment/house2.png", true),
		house3Sprite:      loadSprite(assets, "assets/images/environment/house3.png", true),
		houseBurnedSprite: loadSprite(assets, "assets/images/environment/house1.png", true), // Using house1 as placeholder due to quota
		templeSprite:      loadSprite(assets, "assets/images/environment/temple.png", true),
		smitherySprite:    loadSprite(assets, "assets/images/environment/smithery.png", true),
		farmSprite:        loadSprite(assets, "assets/images/environment/farm.png", true),
		warehouseSprite:   loadSprite(assets, "assets/images/environment/warehouse.png", true),
		rock2Sprite:       loadSprite(assets, "assets/images/environment/rock2.png", true),
		rock3Sprite:       loadSprite(assets, "assets/images/environment/rock3.png", true),
		rock4Sprite:       loadSprite(assets, "assets/images/environment/rock4.png", true),
		rock5Sprite:       loadSprite(assets, "assets/images/environment/rock5.png", true),
		bush2Sprite:       loadSprite(assets, "assets/images/environment/bush2.png", true),
		bush3Sprite:       loadSprite(assets, "assets/images/environment/bush3.png", true),
		bush4Sprite:       loadSprite(assets, "assets/images/environment/bush4.png", true),
		bush5Sprite:       loadSprite(assets, "assets/images/environment/bush5.png", true),
		portalSprite:      loadSprite(assets, "assets/images/scenario/portal.png", true),
		crownSprite:       loadSprite(assets, "assets/images/scenario/crown.png", true),
		zoneSprite:        loadSprite(assets, "assets/images/scenario/zone_marker.png", true),
		camera:            engine.NewCamera(pIsoX, pIsoY),
		assets:            assets,
		generatedChunks:   make(map[image.Point]bool),
		npcSpawnTimer:     0,
		npcRegistry:       NewNPCRegistry(),
		mapRegistry:       mapRegistry,
		obstacleRegistry:  NewObstacleRegistry(),
		currentMap:        selectedMap,
		mapLevel:          mapLevel,
	}

	if err := g.npcRegistry.LoadAll(assets); err != nil {
		log.Printf("Error loading NPC configs: %v", err)
	}
	if err := g.obstacleRegistry.LoadAll(assets); err != nil {
		log.Printf("Error loading Obstacle configs: %v", err)
	}

	// Initial generation around player
	g.updateChunks()

	// Spawn NPCs
	g.npcs = make([]*NPC, 0)

	g.loadMapLevel()

	return g
}

func (g *Game) loadMapLevel() {
	// Pick a map, generally increasing difficulty matching the level
	availableMaps := make([]MapData, 0)
	for _, id := range g.mapRegistry.IDs {
		m := g.mapRegistry.Maps[id]
		// allow slightly higher or lower difficulty maps
		if m.Difficulty <= g.mapLevel+1 {
			availableMaps = append(availableMaps, *m)
		}
	}

	if len(availableMaps) > 0 {
		g.currentMap = availableMaps[rand.Intn(len(availableMaps))]
	} else if len(g.mapRegistry.IDs) > 0 {
		g.currentMap = *g.mapRegistry.Maps[g.mapRegistry.IDs[0]] // Fallback
	}

	// Reset map-specific state
	g.playTime = 0
	g.npcs = make([]*NPC, 0)
	g.currentMap.StartTime = 0

	// Apply Difficulty Multipliers
	g.currentMap.TargetTime *= float64(g.mapLevel)

	newKills := make(map[string]int)
	for npcID, target := range g.currentMap.TargetKills {
		newKills[npcID] = target * g.mapLevel
	}
	g.currentMap.TargetKills = newKills

	// Spawn map targets
	switch g.currentMap.Type {
	case ObjKillVIP:
		if len(g.npcRegistry.IDs) > 0 {
			vipID := g.npcRegistry.IDs[rand.Intn(len(g.npcRegistry.IDs))]
			vipConfig := g.npcRegistry.Configs[vipID]
			// Spawn far away
			tpX := g.player.X + (rand.Float64()*40 - 20)
			tpY := g.player.Y + (rand.Float64()*40 - 20)
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
			g.currentMap.TargetPoint = engine.Point{X: tpX, Y: tpY} // Not used directly, but stores initial pos

			// We need a way to track the exact instance. In a real ECS we'd use ID.
			// For now, we tag it via config pointer address since we don't have unique IDs per instance yet.
			// Actually, let's just make the *first* NPC the VIP for simplicity if it's a VIP map.
		}
	case ObjReachPortal, ObjReachZone:
		g.currentMap.TargetPoint = engine.Point{
			X: g.player.X + (rand.Float64()*60 - 30),
			Y: g.player.Y + (rand.Float64()*60 - 30),
		}
		if g.currentMap.TargetPoint.X > -10 && g.currentMap.TargetPoint.X < 10 {
			g.currentMap.TargetPoint.X += 20
		}
		if g.currentMap.TargetPoint.Y > -10 && g.currentMap.TargetPoint.Y < 10 {
			g.currentMap.TargetPoint.Y += 20
		}
	case ObjReachBuilding:
		// Pick a random direction
		g.currentMap.TargetPoint = engine.Point{
			X: g.player.X + (rand.Float64()*50 - 25),
			Y: g.player.Y + (rand.Float64()*50 - 25),
		}
		if g.currentMap.TargetPoint.X > -10 && g.currentMap.TargetPoint.X < 10 {
			g.currentMap.TargetPoint.X += 20
		}
		if g.currentMap.TargetPoint.Y > -10 && g.currentMap.TargetPoint.Y < 10 {
			g.currentMap.TargetPoint.Y += 20
		}
		// We'll spawn a building there in the update loop or rely on chunks. Let's force a building spawn.
		if config, ok := g.obstacleRegistry.Configs["warehouse"]; ok {
			g.obstacles = append(g.obstacles, NewObstacle(g.currentMap.TargetPoint.X, g.currentMap.TargetPoint.Y, config))
		} else {
			log.Println("WARNING: Warehouse config not found for ObjReachBuilding!")
		}
	case ObjProtectNPC:
		// Target point far away
		g.currentMap.TargetPoint = engine.Point{
			X: g.player.X + (rand.Float64()*80 - 40),
			Y: g.player.Y + (rand.Float64()*80 - 40),
		}
		if g.currentMap.TargetPoint.X > -20 && g.currentMap.TargetPoint.X < 20 {
			g.currentMap.TargetPoint.X += 40
		}
		if g.currentMap.TargetPoint.Y > -20 && g.currentMap.TargetPoint.Y < 20 {
			g.currentMap.TargetPoint.Y += 40
		}

		// Spawn Escort right next to player
		if config, ok := g.npcRegistry.Configs["escort"]; ok {
			escort := NewNPC(g.player.X+2, g.player.Y+2, config, g.mapLevel)
			g.npcs = append([]*NPC{escort}, g.npcs...) // Prepend so it's always index 0 for easy tracking
		} else {
			log.Println("WARNING: Escort config not found!")
		}
	case ObjDestroyBuilding:
		g.currentMap.TargetPoint = engine.Point{
			X: g.player.X + (rand.Float64()*80 - 40),
			Y: g.player.Y + (rand.Float64()*80 - 40),
		}
		if g.currentMap.TargetPoint.X > -20 && g.currentMap.TargetPoint.X < 20 {
			g.currentMap.TargetPoint.X += 40
		}
		if g.currentMap.TargetPoint.Y > -20 && g.currentMap.TargetPoint.Y < 20 {
			g.currentMap.TargetPoint.Y += 40
		}

		// Spawn a target building like a warehouse or farm
		if config, ok := g.obstacleRegistry.Configs["house_burned"]; ok {
			targetObs := NewObstacle(g.currentMap.TargetPoint.X, g.currentMap.TargetPoint.Y, config)
			g.obstacles = append(g.obstacles, targetObs)
			g.currentMap.TargetObstacle = targetObs
		} else {
			log.Println("WARNING: house_burned config not found for ObjDestroyBuilding!")
		}
	}

	log.Printf("Starting Map Level %d: %s", g.mapLevel, g.currentMap.Name)
}

func (g *Game) Update() error {
	if g.isGameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			os.Exit(0)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			*g = *NewGame(g.assets)
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.isPaused {
			os.Exit(0)
		}
		g.isPaused = true
		return nil
	}

	if g.isPaused {
		// If any other key is pressed, resume
		keys := inpututil.AppendJustPressedKeys(nil)
		if len(keys) > 0 {
			g.isPaused = false
		}
		return nil
	}

	// Handle Save/Load
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		if err := g.Save("save.json"); err == nil {
			log.Println("Game saved to save.json")
		} else {
			log.Printf("Failed to save game: %v", err)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		if err := g.Load("save.json"); err == nil {
			log.Println("Game loaded from save.json")
		} else {
			log.Printf("Failed to load game: %v", err)
		}
	}

	// Check and generate new chunks
	g.updateChunks()

	// Handle NPC spawning
	g.updateNPCSpawning()

	if !g.isPaused && !g.isGameOver {
		g.playTime += 1.0 / 60.0
	}

	g.player.Update(g.obstacles, g.npcs, &g.floatingTexts)

	g.player.Update(g.obstacles, g.npcs, &g.floatingTexts)

	// Check Win Conditions
	mapWon := false
	switch g.currentMap.Type {
	case ObjKillCount:
		// Check all specific NPC targets
		won := true
		for npcID, targetAmount := range g.currentMap.TargetKills {
			if g.player.MapKills[npcID] < targetAmount {
				won = false
				break
			}
		}
		if len(g.currentMap.TargetKills) > 0 && won {
			mapWon = true
		}
	case ObjSurvive:
		if g.playTime >= g.currentMap.TargetTime {
			mapWon = true
		}
	case ObjReachPortal, ObjReachZone, ObjReachBuilding:
		// Check distance
		dx := g.player.X - g.currentMap.TargetPoint.X
		dy := g.player.Y - g.currentMap.TargetPoint.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		radius := g.currentMap.TargetRadius
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
				dx := escort.X - g.currentMap.TargetPoint.X
				dy := escort.Y - g.currentMap.TargetPoint.Y
				dist := math.Sqrt(dx*dx + dy*dy)

				radius := g.currentMap.TargetRadius
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
		if g.playTime >= g.currentMap.TargetTime {
			mapWon = true
		}
		// Failure condition: Don't kill anything
		for _, kills := range g.player.MapKills {
			if kills > 0 {
				g.isGameOver = true
				break
			}
		}
	case ObjDestroyBuilding:
		if g.currentMap.TargetObstacle != nil {
			if !g.currentMap.TargetObstacle.Alive {
				mapWon = true
			}
		}
	}

	if mapWon && !g.isGameOver && g.player.IsAlive() {
		// Transition to next map
		g.mapLevel++
		// Heal player slightly as reward
		g.player.Health += g.player.MaxHealth / 2
		if g.player.Health > g.player.MaxHealth {
			g.player.Health = g.player.MaxHealth
		}
		g.loadMapLevel()
		return nil // skip rest of update this frame
	}

	// Filter dead obstacles
	aliveObstacles := make([]*Obstacle, 0, len(g.obstacles))
	for _, o := range g.obstacles {
		if o.Alive {
			aliveObstacles = append(aliveObstacles, o)
		}
	}
	g.obstacles = aliveObstacles

	// Check if player died
	if !g.player.IsAlive() {
		g.isGameOver = true
	}

	// Update all NPCs
	for _, n := range g.npcs {
		if n.IsAlive() {
			n.Update(g.player, g.obstacles, g.npcs, &g.floatingTexts)
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
	pIsoX, pIsoY := engine.CartesianToIso(g.player.X, g.player.Y)
	g.camera.Follow(pIsoX, pIsoY, 0.1)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)

	g.renderer.DrawInfiniteGrass(screen, offsetX, offsetY, g.grassSprite)

	// Collect all drawable entities for Y-sorting
	type drawTask struct {
		y    float64
		draw func()
	}
	tasks := make([]drawTask, 0)

	// Add obstacles
	for _, o := range g.obstacles {
		obj := o // local copy
		tasks = append(tasks, drawTask{
			y: obj.X + obj.Y,
			draw: func() {
				obj.Draw(screen, offsetX, offsetY)
			},
		})
	}

	// Add NPCs
	for _, n := range g.npcs {
		npc := n // local copy
		tasks = append(tasks, drawTask{
			y: npc.X + npc.Y,
			draw: func() {
				npc.Draw(screen, offsetX, offsetY)
			},
		})
	}

	// Add player
	tasks = append(tasks, drawTask{
		y: g.player.X + g.player.Y,
		draw: func() {
			g.player.Draw(screen, offsetX, offsetY)
		},
	})

	// Sort tasks by Y (back to front)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].y < tasks[j].y
	})

	// Execute draw tasks
	for _, t := range tasks {
		t.draw()
	}

	// Draw floating texts (always on top of entities)
	for _, ft := range g.floatingTexts {
		ft.Draw(screen, offsetX, offsetY)
	}

	// Draw map target points for navigation objectives
	switch g.currentMap.Type {
	case ObjReachPortal:
		tasks = append(tasks, drawTask{
			y: g.currentMap.TargetPoint.X + g.currentMap.TargetPoint.Y,
			draw: func() {
				op := &ebiten.DrawImageOptions{}
				cx, cy := engine.CartesianToIso(g.currentMap.TargetPoint.X, g.currentMap.TargetPoint.Y)
				op.GeoM.Translate(cx-float64(g.portalSprite.Bounds().Dx())/2, cy-float64(g.portalSprite.Bounds().Dy()))
				op.GeoM.Translate(offsetX, offsetY)
				screen.DrawImage(g.portalSprite, op)
			},
		})
	case ObjReachZone, ObjProtectNPC:
		// Draw zone under everything, no need to y-sort.
		op := &ebiten.DrawImageOptions{}
		cx, cy := engine.CartesianToIso(g.currentMap.TargetPoint.X, g.currentMap.TargetPoint.Y)
		op.GeoM.Translate(cx-float64(g.zoneSprite.Bounds().Dx())/2, cy-float64(g.zoneSprite.Bounds().Dy())/2)
		op.GeoM.Translate(offsetX, offsetY)
		// Draw as background element
		screen.DrawImage(g.zoneSprite, op)
	case ObjReachBuilding:
		// Draw light column on top of the target point
		tasks = append(tasks, drawTask{
			y: g.currentMap.TargetPoint.X + g.currentMap.TargetPoint.Y,
			draw: func() {
				// We don't have the light column loaded yet, let's just use the crown scaled up as a placeholder
				// or just rely on the building being spawned there.
				op := &ebiten.DrawImageOptions{}
				cx, cy := engine.CartesianToIso(g.currentMap.TargetPoint.X, g.currentMap.TargetPoint.Y)
				op.GeoM.Scale(2, 8)
				op.GeoM.Translate(cx-16, cy-128) // Adjust for scaled size
				op.GeoM.Translate(offsetX, offsetY)
				// Reusing crown with alpha for column effect temporarily pending sprite load
				op.ColorScale.ScaleAlpha(0.5)
				screen.DrawImage(g.crownSprite, op)
			},
		})
	case ObjKillVIP:
		if len(g.npcs) > 0 && g.npcs[0].IsAlive() {
			tasks = append(tasks, drawTask{
				y: g.npcs[0].X + g.npcs[0].Y + 0.1, // always slightly in front of the NPC
				draw: func() {
					op := &ebiten.DrawImageOptions{}
					cx, cy := engine.CartesianToIso(g.npcs[0].X, g.npcs[0].Y)
					op.GeoM.Translate(cx-float64(g.crownSprite.Bounds().Dx())/2, cy-60) // Float above head
					op.GeoM.Translate(offsetX, offsetY)
					screen.DrawImage(g.crownSprite, op)
				},
			})
		}
	}

	if g.isGameOver {
		g.drawGameOver(screen)
	} else if g.isPaused {
		g.drawPauseMenu(screen)
	} else {
		g.drawHUD(screen)
	}
}

func (g *Game) drawPauseMenu(screen *ebiten.Image) {
	// Semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)

	msg1 := "GAME PAUSED"
	msg2 := "Press ESC again to EXIT"
	msg3 := "Press any other key to RESUME"

	ebitenutil.DebugPrintAt(screen, msg1, g.width/2-40, g.height/2-30)
	ebitenutil.DebugPrintAt(screen, msg2, g.width/2-80, g.height/2)
	ebitenutil.DebugPrintAt(screen, msg3, g.width/2-95, g.height/2+20)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)

	msg1 := "GAME OVER"
	msgKills := fmt.Sprintf("Final Kills: %d", g.player.Kills)
	msgMap := fmt.Sprintf("Maps Completed: %d (Reached Lvl %d)", g.mapLevel-1, g.mapLevel)
	minutes := int(g.playTime) / 60
	seconds := int(g.playTime) % 60
	msgTime := fmt.Sprintf("Total Survival Time: %02d:%02d", minutes, seconds)
	msg2 := "Press ESC to exit, or ENTER to restart"

	ebitenutil.DebugPrintAt(screen, msg1, g.width/2-30, g.height/2-40)
	ebitenutil.DebugPrintAt(screen, msgMap, g.width/2-80, g.height/2-20)
	ebitenutil.DebugPrintAt(screen, msgKills, g.width/2-45, g.height/2)
	ebitenutil.DebugPrintAt(screen, msgTime, g.width/2-75, g.height/2+20)
	ebitenutil.DebugPrintAt(screen, msg2, g.width/2-110, g.height/2+45)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Semi-transparent background for better readability
	hudBg := ebiten.NewImage(350, 150)
	hudBg.Fill(color.RGBA{0, 0, 0, 180})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(10, 10)
	screen.DrawImage(hudBg, op)

	// Health bar
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %d/%d", g.player.Health, g.player.MaxHealth), 20, 20)
	healthBarBg := ebiten.NewImage(200, 10)
	healthBarBg.Fill(color.RGBA{100, 0, 0, 255})
	opBg := &ebiten.DrawImageOptions{}
	opBg.GeoM.Translate(100, 22)
	screen.DrawImage(healthBarBg, opBg)

	healthPct := float64(g.player.Health) / float64(g.player.MaxHealth)
	if healthPct > 0 {
		healthBarFg := ebiten.NewImage(int(200*healthPct), 10)
		healthBarFg.Fill(color.RGBA{255, 0, 0, 255})
		screen.DrawImage(healthBarFg, opBg)
	}

	// Map Objective & Progress
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("MAP [%d]: %s", g.mapLevel, g.currentMap.Name), 20, 42)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("OBJ: %s", g.currentMap.Description), 20, 57)

	minutes := int(g.playTime) / 60
	seconds := int(g.playTime) % 60

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("POS %.1f,%.1f  KILLS: %d  XP: %d  LVL: %d", g.player.X, g.player.Y, g.player.Kills, g.player.XP, g.player.Level), 20, 77)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ATK: %d  DEF: %d  SHIELD: %d", g.player.GetTotalAttack(), g.player.GetTotalDefense(), g.player.GetTotalProtection()), 20, 92)

	weaponText := fmt.Sprintf("WEAPON: %s (%d-%d)", g.player.Weapon.Name, g.player.Weapon.MinDamage, g.player.Weapon.MaxDamage)
	if g.player.Weapon.Bonus > 0 {
		weaponText += fmt.Sprintf(" +%d", g.player.Weapon.Bonus)
	}
	ebitenutil.DebugPrintAt(screen, weaponText, 20, 107)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TIME: %02d:%02d", minutes, seconds), 20, 122)

	// Goal Navigation Arrow
	var tx, ty float64
	hasTarget := false

	switch g.currentMap.Type {
	case ObjKillVIP:
		if len(g.npcs) > 0 && g.npcs[0].IsAlive() {
			tx, ty = g.npcs[0].X, g.npcs[0].Y
			hasTarget = true
		}
	case ObjProtectNPC:
		if len(g.npcs) > 0 && g.npcs[0].IsAlive() {
			tx, ty = g.currentMap.TargetPoint.X, g.currentMap.TargetPoint.Y
			hasTarget = true
		}
	case ObjDestroyBuilding, ObjReachBuilding, ObjReachPortal, ObjReachZone:
		tx, ty = g.currentMap.TargetPoint.X, g.currentMap.TargetPoint.Y
		hasTarget = true
	}

	if hasTarget {
		// Calculate angle in cartesian world space
		dx := tx - g.player.X
		dy := ty - g.player.Y
		// We want to project the angle to the screen, so we need to account for isometric projection
		// In iso, moving dx, dy changes screen by:
		// scrX = (dx - dy) * tileWidth/2
		// scrY = (dx + dy) * tileHeight/2
		isoDx := dx - dy
		isoDy := (dx + dy) * 0.5 // iso ratio
		angle := math.Atan2(isoDy, isoDx)

		arrowX := float32(g.width - 50)
		arrowY := float32(50)
		size := float32(20.0)

		// Create a triangle pointing right
		var path vector.Path
		path.MoveTo(size, 0)
		path.LineTo(-size, -size*0.6)
		path.LineTo(-size*0.5, 0)
		path.LineTo(-size, size*0.6)
		path.Close()

		opArr := &ebiten.DrawTrianglesOptions{}
		opArr.FillRule = ebiten.EvenOdd

		// We do math.Cos and math.Sin for manual rotation because vector.Path doesn't have an easy native transform before 2.5
		cosA := float32(math.Cos(angle))
		sinA := float32(math.Sin(angle))

		var vs []ebiten.Vertex
		var is []uint16

		vs, is = path.AppendVerticesAndIndicesForFilling(vs, is)
		for i := range vs {
			// Rotate
			rx := vs[i].DstX*cosA - vs[i].DstY*sinA
			ry := vs[i].DstX*sinA + vs[i].DstY*cosA
			vs[i].DstX = rx + arrowX
			vs[i].DstY = ry + arrowY
			vs[i].SrcX = 1
			vs[i].SrcY = 1
			vs[i].ColorR = 1.0
			vs[i].ColorG = 0.2
			vs[i].ColorB = 0.2
			vs[i].ColorA = 1.0
		}

		// Create a 1x1 white image for the vertices
		whiteImage := ebiten.NewImage(3, 3)
		whiteImage.Fill(color.White)
		screen.DrawTriangles(vs, is, whiteImage, opArr)

		ebitenutil.DebugPrintAt(screen, "OBJ", int(arrowX)-10, int(arrowY)+25)
	}
}

func (g *Game) updateChunks() {
	const chunkSize = 10
	cpX := int(math.Floor(g.player.X / float64(chunkSize)))
	cpY := int(math.Floor(g.player.Y / float64(chunkSize)))

	// Check 9x9 grid around player (radius 4)
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
			config := g.obstacleRegistry.Configs[ot]
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
			config := g.obstacleRegistry.Configs[ot]
			g.obstacles = append(g.obstacles, NewObstacle(wx, wy, config))
			wx += r.Float64()*1.2 - 0.6
			wy += r.Float64()*1.2 - 0.6
		}
	}

	// 3. Buildings
	if r.Float64() < 0.1 {
		bx := startX + r.Float64()*chunkSize
		by := startY + r.Float64()*chunkSize
		// Don't spawn on top of player's initial position
		if math.Abs(bx) > 3 || math.Abs(by) > 3 {
			types := []string{"house1", "house2", "house3", "house_burned", "temple", "smithery", "farm", "warehouse"}
			ot := types[r.Intn(len(types))]
			config := g.obstacleRegistry.Configs[ot]
			g.obstacles = append(g.obstacles, NewObstacle(bx, by, config))
		}
	}

	// 4. Random bushes
	for i := 0; i < 2; i++ {
		if r.Float64() < 0.4 {
			bx := startX + r.Float64()*chunkSize
			by := startY + r.Float64()*chunkSize
			types := []string{"bush1", "bush2", "bush3", "bush4", "bush5"}
			ot := types[r.Intn(len(types))]
			config := g.obstacleRegistry.Configs[ot]
			g.obstacles = append(g.obstacles, NewObstacle(bx, by, config))
		}
	}
}

func (g *Game) updateNPCSpawning() {
	g.npcSpawnTimer++

	// Base spawn rate is every 180 ticks (3 seconds)
	// Difficulty modifier reduces this time (makes them spawn faster)
	// Example: level 1 = 180, level 2 = 150, level 3 = 120... (clamp to minimum 30)
	spawnThreshold := 180 - (g.currentMap.Difficulty * 20) - (g.mapLevel * 10)
	if spawnThreshold < 30 {
		spawnThreshold = 30
	}
	if g.currentMap.SpawnFreq > 0 {
		spawnThreshold = int(g.currentMap.SpawnFreq * 60)
	}

	if g.npcSpawnTimer >= spawnThreshold {
		g.npcSpawnTimer = 0
		maxNPCs := 15 + (g.mapLevel * 5) // More max enemies in higher levels
		if maxNPCs > 50 {
			maxNPCs = 50
		}

		if len(g.npcs) < maxNPCs {
			spawnAmount := 1
			if g.currentMap.SpawnAmount > 0 {
				spawnAmount = g.currentMap.SpawnAmount
			}

			for i := 0; i < spawnAmount; i++ {
				if g.currentMap.Type == ObjDestroyBuilding && g.currentMap.TargetObstacle != nil && g.currentMap.TargetObstacle.Alive {
					g.spawnNPCNear(g.currentMap.TargetObstacle.X, g.currentMap.TargetObstacle.Y)
				} else {
					g.spawnNPCAtEdges()
				}
			}
		}
	}

	// Periodic cleanup of far away NPCs
	if g.npcSpawnTimer == 90 {
		activeNPCs := make([]*NPC, 0)
		for _, n := range g.npcs {
			dist := math.Sqrt(math.Pow(n.X-g.player.X, 2) + math.Pow(n.Y-g.player.Y, 2))
			if n.IsAlive() {
				if dist < 40 {
					activeNPCs = append(activeNPCs, n)
				}
			} else {
				if dist < 25 {
					activeNPCs = append(activeNPCs, n)
				}
			}
		}
		g.npcs = activeNPCs
	}
}

func (g *Game) pickNPCIDToSpawn() string {
	if len(g.currentMap.SpawnWeights) > 0 {
		totalWeight := 0
		for _, w := range g.currentMap.SpawnWeights {
			totalWeight += w
		}
		if totalWeight > 0 {
			roll := rand.Intn(totalWeight)
			curr := 0
			for id, w := range g.currentMap.SpawnWeights {
				curr += w
				if roll < curr {
					if _, ok := g.npcRegistry.Configs[id]; ok {
						return id
					}
				}
			}
		}
	}
	// Fallback
	return g.npcRegistry.IDs[rand.Intn(len(g.npcRegistry.IDs))]
}

func (g *Game) spawnNPCNear(x, y float64) {
	if len(g.npcRegistry.IDs) == 0 {
		return
	}
	const spawnRadius = 2.0
	angle := rand.Float64() * 2 * math.Pi
	ex := x + math.Cos(angle)*spawnRadius
	ey := y + math.Sin(angle)*spawnRadius

	npcID := g.pickNPCIDToSpawn()
	npcConfig := g.npcRegistry.Configs[npcID]
	g.npcs = append(g.npcs, NewNPC(ex, ey, npcConfig, g.mapLevel))
}

func (g *Game) spawnNPCAtEdges() {
	if len(g.npcRegistry.IDs) == 0 {
		return
	}

	const spawnDist = 30.0
	angle := rand.Float64() * 2 * math.Pi
	ex := g.player.X + math.Cos(angle)*spawnDist
	ey := g.player.Y + math.Sin(angle)*spawnDist

	// Pick a random NPC config based on weights
	npcID := g.pickNPCIDToSpawn()
	npcConfig := g.npcRegistry.Configs[npcID]

	g.npcs = append(g.npcs, NewNPC(ex, ey, npcConfig, g.mapLevel))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.width, g.height
}
