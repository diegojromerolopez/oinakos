package game

import (
	"fmt"
	"image/color"
	"io/fs"
	"log"
	"math"
	"path"
	"sort"

	"oinakos/internal/engine"
	"strings"
)

// GameRenderer handles the Ebiten-dependent rendering of the game.
type GameRenderer struct {
	game          *Game
	renderer      *engine.Renderer
	graphics      engine.Graphics
	grassSprite   engine.Image
	emptyImage    engine.Image
	lastFloorPath string
	PaletteShader engine.Shader

	tileCache  map[string]engine.Image
	coordCache map[string]string
}

func NewGameRenderer(g *Game, assets fs.FS, graphics engine.Graphics) *GameRenderer {
	gr := &GameRenderer{
		game:        g,
		renderer:    engine.NewRenderer(),
		graphics:    graphics,
		grassSprite: graphics.LoadSprite(assets, "assets/images/floors/grass.png", true),
	}
	gr.emptyImage = graphics.NewImage(3, 3)
	gr.emptyImage.Fill(color.White)
	gr.tileCache = make(map[string]engine.Image)
	gr.coordCache = make(map[string]string)
	return gr
}

func (gr *GameRenderer) LoadAssets(assets fs.FS) {
	gr.game.archetypeRegistry.LoadAssets(assets, gr.graphics)
	gr.game.playableCharacterRegistry.LoadAssets(assets, gr.graphics)
	gr.game.obstacleRegistry.LoadAssets(assets, gr.graphics)
	gr.game.npcRegistry.LoadAssets(assets, gr.graphics, gr.game.archetypeRegistry)

	var err error
	gr.PaletteShader, err = gr.graphics.NewShader(paletteSwapShaderSource)
	if err != nil {
		log.Printf("Error building palette shader: %v", err)
	}

	// Load player assets
	mc := gr.game.playableCharacter
	if mc != nil && mc.Config != nil {
		if mc.Config.AssetDir == "" {
			mc.Config.AssetDir = "assets/images/characters/oinakos"
		}
		imgDir := mc.Config.AssetDir
		staticPath := path.Join(imgDir, "static.png")
		if _, err := fs.Stat(assets, staticPath); err == nil {
			mc.Config.StaticImage = gr.graphics.LoadSprite(assets, staticPath, true)
		}

		backPath := path.Join(imgDir, "back.png")
		if _, err := fs.Stat(assets, backPath); err == nil {
			mc.Config.BackImage = gr.graphics.LoadSprite(assets, backPath, true)
		}

		corpsePath := path.Join(imgDir, "corpse.png")
		if _, err := fs.Stat(assets, corpsePath); err == nil {
			mc.Config.CorpseImage = gr.graphics.LoadSprite(assets, corpsePath, true)
		}

		attackPath := path.Join(imgDir, "attack.png")
		if _, err := fs.Stat(assets, attackPath); err == nil {
			mc.Config.AttackImage = gr.graphics.LoadSprite(assets, attackPath, true)
		}
		attack1Path := path.Join(imgDir, "attack1.png")
		if _, err := fs.Stat(assets, attack1Path); err == nil {
			mc.Config.Attack1Image = gr.graphics.LoadSprite(assets, attack1Path, true)
		}
		attack2Path := path.Join(imgDir, "attack2.png")
		if _, err := fs.Stat(assets, attack2Path); err == nil {
			mc.Config.Attack2Image = gr.graphics.LoadSprite(assets, attack2Path, true)
		}

		hitPath := path.Join(imgDir, "hit.png")
		if _, err := fs.Stat(assets, hitPath); err == nil {
			mc.Config.HitImage = gr.graphics.LoadSprite(assets, hitPath, true)
		}
		hit1Path := path.Join(imgDir, "hit1.png")
		if _, err := fs.Stat(assets, hit1Path); err == nil {
			mc.Config.Hit1Image = gr.graphics.LoadSprite(assets, hit1Path, true)
		}
		hit2Path := path.Join(imgDir, "hit2.png")
		if _, err := fs.Stat(assets, hit2Path); err == nil {
			mc.Config.Hit2Image = gr.graphics.LoadSprite(assets, hit2Path, true)
		}
	}
}

func (gr *GameRenderer) Draw(screen engine.Image) {
	if screen == nil {
		log.Println("GameRenderer.Draw called with nil screen!")
		return
	}
	// log.Printf("GameRenderer.Draw Frame") // No longer needed
	g := gr.game
	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)

	if g.isMainMenu {
		gr.drawMainMenu(screen)
		return
	}

	if g.isCharacterSelect {
		gr.drawCharacterSelect(screen)
		return
	}

	if g.isCampaignSelect {
		gr.drawCampaignSelect(screen)
		return
	}

	if g.isSettingsScreen {
		gr.drawSettingsScreen(screen)
		return
	}

	if g.currentMapType.FloorTile != "" && g.currentMapType.FloorTile != gr.lastFloorPath {
		// Clear caches when map floor base changes
		gr.coordCache = make(map[string]string)
		gr.lastFloorPath = g.currentMapType.FloorTile
	}

	gr.renderer.DrawTileMap(screen, offsetX, offsetY, func(x, y int) engine.Image {
		return gr.getTileAt(x, y)
	})

	// Collect all drawable entities for Y-sorting
	type drawTask struct {
		y    float64
		draw func()
	}
	tasks := make([]drawTask, 0, len(g.obstacles)+len(g.npcs)+1)

	// Add obstacles
	for _, o := range g.obstacles {
		img, _ := o.Archetype.Image.(engine.Image)
		if img == nil {
			continue
		}
		sw, sh := img.Size()

		isoX, isoY := engine.CartesianToIso(o.X, o.Y)
		drawX := isoX + offsetX
		drawY := isoY + offsetY

		// Dynamic culling: use sprite dimensions to ensure large buildings don't flicker
		marginW := float64(sw)
		marginH := float64(sh)
		if drawX < -marginW || drawX > float64(g.width)+marginW || drawY < -marginH || drawY > float64(g.height)+marginH {
			continue
		}

		obj := o // local copy
		sortY := obj.X + obj.Y
		if obj.Archetype.Type == "static" || obj.Archetype.Type == "well" {
			sortY += 2.0
		} else {
			// Center of footprint
			p := obj.GetFootprint()
			if len(p.Points) > 0 {
				var minX, maxX, minY, maxY float64
				for i, pt := range p.Points {
					if i == 0 || pt.X < minX {
						minX = pt.X
					}
					if i == 0 || pt.X > maxX {
						maxX = pt.X
					}
					if i == 0 || pt.Y < minY {
						minY = pt.Y
					}
					if i == 0 || pt.Y > maxY {
						maxY = pt.Y
					}
				}
				// The footprint is already absolute, so we convert back to world "depth"
				sortY = (minX + maxX + minY + maxY) / 2
			}
		}

		tasks = append(tasks, drawTask{
			y: sortY,
			draw: func() {
				obj.Draw(screen, gr.graphics, offsetX, offsetY)
			},
		})
	}

	// Add NPCs
	for _, n := range g.npcs {
		isoX, isoY := engine.CartesianToIso(n.X, n.Y)
		drawX := isoX + offsetX
		drawY := isoY + offsetY
		// Culling margin
		if drawX < -256 || drawX > float64(g.width)+256 || drawY < -256 || drawY > float64(g.height)+256 {
			continue
		}

		npc := n // local copy
		sortY := npc.X + npc.Y
		if npc.State == NPCDead {
			sortY -= 100.0 // Push corpses behind a lot of things
		}
		tasks = append(tasks, drawTask{
			y: sortY,
			draw: func() {
				npc.Draw(screen, gr.graphics, gr.graphics, gr.PaletteShader, offsetX, offsetY)
			},
		})
	}

	// Add playableCharacter
	mcSortY := g.playableCharacter.X + g.playableCharacter.Y
	if g.playableCharacter.State == StateDead {
		mcSortY -= 100.0
	}
	tasks = append(tasks, drawTask{
		y: mcSortY,
		draw: func() {
			g.playableCharacter.Draw(screen, gr.graphics, gr.graphics, offsetX, offsetY)
		},
	})

	// Add projectiles
	for _, p := range g.projectiles {
		proj := p // local copy
		tasks = append(tasks, drawTask{
			y: proj.X + proj.Y,
			draw: func() {
				proj.Draw(screen, gr.graphics, offsetX, offsetY)
			},
		})
	}

	// Sort tasks by Y coordinate to achieve correct depth rendering
	// Because of isometric projection, higher X+Y means further "down" on the screen
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].y < tasks[j].y
	})

	// Execute draw tasks (everything including buildings, projectiles, etc.)
	for _, t := range tasks {
		t.draw()
	}

	// Draw debug information if enabled
	if g.debug || g.showBoundaries {
		gr.drawDebug(screen, offsetX, offsetY)
	}

	// Draw floating texts (always on top of entities)
	for _, ft := range g.floatingTexts {
		ft.Draw(screen, gr.graphics, offsetX, offsetY)
	}

	if g.isGameWon {
		gr.drawGameWon(screen)
	} else if g.isGameOver {
		gr.drawGameOver(screen)
	} else if g.isMapWon {
		gr.drawMapWon(screen)
	} else if g.isPaused {
		gr.drawPauseMenu(screen)
	} else {
		gr.drawHUD(screen)
		gr.drawHoverInfo(screen)
	}
}

func (gr *GameRenderer) drawPauseMenu(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)

	title := "GAME PAUSED"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-50, color.White, 32)

	msg1 := "Press S to SAVE and QUIT"
	msg2 := "Press any other key to RESUME"
	mw1, _ := gr.graphics.MeasureText(msg1, 18)
	mw2, _ := gr.graphics.MeasureText(msg2, 18)
	gr.graphics.DrawTextAt(screen, msg1, (g.width-int(mw1))/2, g.height/2, color.White, 18)
	gr.graphics.DrawTextAt(screen, msg2, (g.width-int(mw2))/2, g.height/2+30, color.White, 18)
}

func (gr *GameRenderer) drawGameOver(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)
	minutes := int(g.playTime) / 60
	seconds := int(g.playTime) % 60

	title := "GAME OVER"
	tw, _ := gr.graphics.MeasureText(title, 48)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-80, color.White, 48)

	kills := fmt.Sprintf("Kills: %d", g.playableCharacter.Kills)
	time := fmt.Sprintf("Time: %02d:%02d", minutes, seconds)
	msg := "Press ESC to exit, or ENTER to restart"

	kw, _ := gr.graphics.MeasureText(kills, 20)
	tmw, _ := gr.graphics.MeasureText(time, 20)
	mw, _ := gr.graphics.MeasureText(msg, 16)

	gr.graphics.DrawTextAt(screen, kills, (g.width-int(kw))/2, g.height/2-15, color.White, 20)
	gr.graphics.DrawTextAt(screen, time, (g.width-int(tmw))/2, g.height/2+20, color.White, 20)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(mw))/2, g.height/2+60, color.White, 16)
}

func (gr *GameRenderer) drawMapWon(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{20, 60, 20, 200}, false)
	mapKillTotal := 0
	for _, k := range g.playableCharacter.MapKills {
		mapKillTotal += k
	}

	title := "MAP WON!"
	tw, _ := gr.graphics.MeasureText(title, 48)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, g.height/2-80, color.White, 48)

	kills := fmt.Sprintf("Map Kills: %d", mapKillTotal)
	msg := "Press ENTER to continue, ESC to quit"

	kw, _ := gr.graphics.MeasureText(kills, 20)
	mw, _ := gr.graphics.MeasureText(msg, 16)

	gr.graphics.DrawTextAt(screen, kills, (g.width-int(kw))/2, g.height/2-15, color.White, 20)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(mw))/2, g.height/2+60, color.White, 16)
}

func (gr *GameRenderer) drawHUD(screen engine.Image) {
	g := gr.game
	// Use DrawFilledRect instead of NewImage every frame to avoid Metal leaks
	gr.graphics.DrawFilledRect(screen, 10, 10, 350, 150, color.RGBA{0, 0, 0, 180}, false)

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("HP: %d/%d", g.playableCharacter.Health, g.playableCharacter.MaxHealth), 20, 20, color.White, 16)

	// Health bar background
	gr.graphics.DrawFilledRect(screen, 100, 22, 200, 10, color.RGBA{100, 0, 0, 255}, false)

	healthPct := float64(g.playableCharacter.Health) / float64(g.playableCharacter.MaxHealth)
	if healthPct > 0 {
		var healthColor color.RGBA
		if healthPct > 0.7 {
			healthColor = color.RGBA{0, 200, 0, 255}
		} else if healthPct > 0.5 {
			healthColor = color.RGBA{200, 200, 0, 255}
		} else {
			healthColor = color.RGBA{200, 0, 0, 255}
		}
		gr.graphics.DrawFilledRect(screen, 100, 22, float32(200*healthPct), 10, healthColor, false)
	}

	gr.graphics.DrawTextAt(screen, fmt.Sprintf("LVL: %d  XP: %d", g.playableCharacter.Level, g.playableCharacter.XP), 20, 45, color.White, 14)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("OBJ: %s", g.currentMapType.Description), 20, 60, color.White, 12)

	minutes := int(g.playTime) / 60
	seconds := int(g.playTime) % 60
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("POS %.1f,%.1f  KILLS: %d  XP: %d  LVL: %d", g.playableCharacter.X, g.playableCharacter.Y, g.playableCharacter.Kills, g.playableCharacter.XP, g.playableCharacter.Level), 20, 80, color.White, 12)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("ATK: %d  DEF: %d  SHIELD: %d", g.playableCharacter.GetTotalAttack(), g.playableCharacter.GetTotalDefense(), g.playableCharacter.GetTotalProtection()), 20, 95, color.White, 12)

	weaponText := fmt.Sprintf("WEAPON: %s (%d-%d)", g.playableCharacter.Weapon.Name, g.playableCharacter.Weapon.MinDamage, g.playableCharacter.Weapon.MaxDamage)
	if g.playableCharacter.Weapon.Bonus > 0 {
		weaponText += fmt.Sprintf(" +%d", g.playableCharacter.Weapon.Bonus)
	}
	gr.graphics.DrawTextAt(screen, weaponText, 20, 110, color.White, 12)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("TIME: %02d:%02d", minutes, seconds), 20, 125, color.White, 12)

	// Menu Button (Top-Right)
	gr.graphics.DrawFilledRect(screen, float32(g.width-110), 20, 100, 30, color.RGBA{50, 50, 50, 200}, false)
	gr.graphics.DrawTextAt(screen, "MENU", g.width-85, 28, color.White, 16)

	// Map Name below Menu
	mapTitle := strings.ToUpper(g.currentMapType.Name)
	if g.isCampaign && g.currentCampaign != nil {
		mapTitle = strings.ToUpper(g.currentCampaign.Name)
	}
	mtw, _ := gr.graphics.MeasureText(mapTitle, 16)
	gr.graphics.DrawTextAt(screen, mapTitle, g.width-int(mtw)-20, 60, color.RGBA{218, 165, 32, 255}, 16)

	// Menu Overlay
	if g.isMenuOpen {
		mw, mh := 400, 280
		mx, my := (g.width-mw)/2, (g.height-mh)/2
		// Border & Backdrop
		gr.graphics.DrawFilledRect(screen, float32(mx-2), float32(my-2), float32(mw+4), float32(mh+4), color.RGBA{218, 165, 32, 255}, false)
		gr.graphics.DrawFilledRect(screen, float32(mx), float32(my), float32(mw), float32(mh), color.RGBA{0, 0, 0, 240}, false)

		menuTitle := "GAME MENU"
		mtw, _ := gr.graphics.MeasureText(menuTitle, 24)
		gr.graphics.DrawTextAt(screen, menuTitle, mx+(mw-int(mtw))/2, my+30, color.RGBA{218, 165, 32, 255}, 24)

		options := []string{"Resume", "Quicksave (Q)", "Load", "Settings", "Quit"}
		for i, opt := range options {
			var clr color.Color = color.White
			prefix := "  "
			if g.menuIndex == i {
				clr = color.RGBA{255, 255, 0, 255}
				prefix = "> "
			}
			gr.graphics.DrawTextAt(screen, prefix+opt, mx+100, my+80+i*35, clr, 18)
		}
		instr := "Press ENTER to select"
		itw, _ := gr.graphics.MeasureText(instr, 14)
		gr.graphics.DrawTextAt(screen, instr, mx+(mw-int(itw))/2, my+mh-30, color.RGBA{136, 136, 136, 255}, 14)
	}

	// Save Message (Bottom Center)
	if g.saveMessageTimer > 0 {
		msg := g.saveMessage
		sttw, _ := gr.graphics.MeasureText(msg, 18)
		gr.graphics.DrawTextAt(screen, msg, (g.width-int(sttw))/2, g.height-40, color.RGBA{218, 165, 32, 255}, 18)
	}
}

func (gr *GameRenderer) drawDebug(screen engine.Image, offsetX, offsetY float64) {
	red := color.RGBA{255, 0, 0, 255}
	green := color.RGBA{0, 255, 0, 255}
	cyan := color.RGBA{0, 255, 255, 255}

	// Helper to draw a Cartesian polygon in Isometric space
	drawPolygon := func(poly engine.Polygon, clr color.Color) {
		isoPoints := make([]engine.Point, len(poly.Points))
		for i, p := range poly.Points {
			ix, iy := engine.CartesianToIso(p.X, p.Y)
			isoPoints[i] = engine.Point{X: ix + offsetX, Y: iy + offsetY}
		}
		gr.graphics.DrawPolygon(screen, isoPoints, clr, 1.0)
	}

	// Obstacles: Cyan
	for _, o := range gr.game.obstacles {
		drawPolygon(o.GetFootprint(), cyan)
	}

	// NPCs
	for _, n := range gr.game.npcs {
		clr := red // Default: Enemy
		if n.Alignment == AlignmentAlly {
			clr = green
		} else if n.Alignment == AlignmentNeutral {
			clr = cyan
		}
		drawPolygon(n.GetFootprint(), clr)
	}

	// Player: Green
	drawPolygon(gr.game.playableCharacter.GetFootprint(), green)
}

func (gr *GameRenderer) drawHoverInfo(screen engine.Image) {
	g := gr.game
	mx, my := g.input.MousePosition()
	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)

	// Check NPCs
	for _, n := range g.npcs {
		if !n.IsAlive() {
			continue
		}
		isoX, isoY := engine.CartesianToIso(n.X, n.Y)
		scrX, scrY := isoX+offsetX, isoY+offsetY

		// Radius check for hover (offset by head height roughly)
		dist := math.Sqrt(math.Pow(float64(mx)-scrX, 2) + math.Pow(float64(my)-scrY+40, 2))
		if dist < 40 {
			if n.Archetype != nil && n.Archetype.Description != "" {
				gr.drawInfoBox(screen, n.Name, n.Archetype.Description, mx, my)
				return // Only one info box at a time
			}
		}
	}
}

func (gr *GameRenderer) drawInfoBox(screen engine.Image, title, desc string, x, y int) {
	// Draw a dark translucent box
	boxW, boxH := 300.0, 160.0
	bx, by := float32(x+20), float32(y+20)

	// Keep on screen
	if float64(bx)+boxW > float64(gr.game.width) {
		bx = float32(float64(x) - boxW - 20)
	}
	if float64(by)+boxH > float64(gr.game.height) {
		by = float32(float64(y) - boxH - 20)
	}

	gr.graphics.DrawFilledRect(screen, bx-2, by-2, float32(boxW+4), float32(boxH+4), color.RGBA{218, 165, 32, 255}, false)
	gr.graphics.DrawFilledRect(screen, bx, by, float32(boxW), float32(boxH), color.RGBA{0, 0, 0, 240}, false)
	gr.graphics.DebugPrintAt(screen, title, int(bx)+10, int(by)+10, color.RGBA{218, 165, 32, 255})

	// Wrap text manually for DebugPrint (it doesn't wrap)
	words := strings.Fields(desc)
	line := ""
	lineNum := 0
	for _, w := range words {
		if len(line)+len(w) > 35 {
			gr.graphics.DebugPrintAt(screen, line, int(bx)+10, int(by)+35+lineNum*15, color.White)
			line = w + " "
			lineNum++
		} else {
			line += w + " "
		}
	}
	gr.graphics.DebugPrintAt(screen, line, int(bx)+10, int(by)+35+lineNum*15, color.White)
}
func (gr *GameRenderer) drawCampaignSelect(screen engine.Image) {
	g := gr.game
	// Black background
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	// Title
	title := "OINAKOS: SELECT YOUR JOURNEY"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 50, color.RGBA{218, 165, 32, 255}, 32)

	col1X := 100
	col2X := g.width / 2

	gr.graphics.DrawTextAt(screen, "--- CAMPAIGNS ---", col1X-20, 100, color.RGBA{150, 150, 150, 255}, 18)
	gr.graphics.DrawTextAt(screen, "--- MAPS ---", col2X-20, 100, color.RGBA{150, 150, 150, 255}, 18)

	nC := len(g.campaignRegistry.IDs)
	nM := len(g.mapTypeRegistry.IDs)
	y := 130

	// Draw campaigns in the first column
	for i, id := range g.campaignRegistry.IDs {
		camp := g.campaignRegistry.Campaigns[id]
		var clr color.Color = color.White
		prefix := "  "
		if g.campaignMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}
		gr.graphics.DrawTextAt(screen, prefix+camp.Name, col1X, y+i*30, clr, 16)
	}

	// Draw maps in the second column, possibly wrapping around if there are too many
	for i, id := range g.mapTypeRegistry.IDs {
		m := g.mapTypeRegistry.Types[id]
		var clr color.Color = color.White
		prefix := "  "
		idx := nC + i
		if g.campaignMenuIndex == idx {
			clr = color.RGBA{150, 255, 150, 255}
			prefix = "> "
		}

		// Calculate row and column wrapping (max ~15 items per column)
		colOffset := col2X
		rowOffset := i
		if i > 15 {
			colOffset += 250 // Shift to a third column if there are tons of maps
			rowOffset = i - 16
		}

		gr.graphics.DrawTextAt(screen, prefix+m.Name, colOffset, y+rowOffset*30, clr, 16)
	}

	// Quit option at bottom
	var clr color.Color = color.White
	prefix := "  "
	if g.campaignMenuIndex == nC+nM {
		clr = color.RGBA{255, 0, 0, 255}
		prefix = "> "
	}
	quitText := prefix + "QUIT"
	qw, _ := gr.graphics.MeasureText(quitText, 24)
	gr.graphics.DrawTextAt(screen, quitText, (g.width-int(qw))/2, g.height-90, clr, 24)

	msg := "Press UP/DOWN to navigate, ENTER to begin."
	mw, _ := gr.graphics.MeasureText(msg, 14)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(mw))/2, g.height-50, color.RGBA{136, 136, 136, 255}, 14)
}

func (gr *GameRenderer) drawMainMenu(screen engine.Image) {
	g := gr.game
	// Black background
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	// Large Title
	title := "OINAKOS"
	tw, _ := gr.graphics.MeasureText(title, 60)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 150, color.RGBA{218, 165, 32, 255}, 60)

	subtitle := "A KNIGHT'S PATH"
	stw, _ := gr.graphics.MeasureText(subtitle, 24)
	gr.graphics.DrawTextAt(screen, subtitle, (g.width-int(stw))/2, 220, color.RGBA{150, 150, 150, 255}, 24)

	options := []string{"NEW GAME", "LOAD GAME", "SETTINGS", "QUIT"}

	for i, opt := range options {
		var clr color.Color = color.White
		prefix := "  "
		if g.mainMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}
		label := prefix + opt
		lw, _ := gr.graphics.MeasureText(label, 32)
		gr.graphics.DrawTextAt(screen, label, (g.width-int(lw))/2, 350+i*60, clr, 32)
	}

	gr.graphics.DrawTextAt(screen, "v0.9.0-alpha", 20, g.height-30, color.RGBA{80, 80, 80, 255}, 14)
}

func (gr *GameRenderer) drawSettingsScreen(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	title := "SETTINGS"
	tw, _ := gr.graphics.MeasureText(title, 40)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 100, color.RGBA{218, 165, 32, 255}, 40)

	rows := []string{"Font Style", "Sound Effects", "Save and Back"}
	for i, row := range rows {
		var clr color.Color = color.White
		prefix := "  "
		if g.settingsMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}

		label := prefix + row
		if i == 0 {
			label += fmt.Sprintf(": [%s]", strings.ToUpper(FontOptions[g.settingsFontIndex]))
		} else if i == 1 {
			label += fmt.Sprintf(": [%s]", strings.ToUpper(FrequencyOptions[g.settingsAudioIndex]))
		}

		lw, _ := gr.graphics.MeasureText(label, 24)
		gr.graphics.DrawTextAt(screen, label, (g.width-int(lw))/2, 250+i*60, clr, 24)
	}

	hint := "UP/DOWN to navigate, LEFT/RIGHT to change value, ENTER to confirm."
	hw, _ := gr.graphics.MeasureText(hint, 14)
	gr.graphics.DrawTextAt(screen, hint, (g.width-int(hw))/2, g.height-100, color.RGBA{136, 136, 136, 255}, 14)
}

func (gr *GameRenderer) drawCharacterSelect(screen engine.Image) {
	g := gr.game
	// Black background
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.Black, false)

	// Title
	title := "OINAKOS: CHOOSE YOUR HERO"
	tw, _ := gr.graphics.MeasureText(title, 32)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 50, color.RGBA{218, 165, 32, 255}, 32)

	mx, my := g.input.MousePosition()
	hoverIndex := -1

	for i, id := range g.playableCharacterRegistry.IDs {
		char := g.playableCharacterRegistry.Characters[id]
		var clr color.Color = color.White
		prefix := "  "

		// Mouse hover detection: X in [100, 400], Y in [130+i*30-5, 130+i*30+25]
		if mx >= 100 && mx <= 400 && my >= 130+i*30-5 && my <= 130+i*30+25 {
			hoverIndex = i
		}

		if g.characterMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "

			// Draw full info for selected hero
			gr.drawHeroPreview(screen, char, g.width/2+50, 130)
		}
		gr.graphics.DrawTextAt(screen, prefix+char.Name, 100, 130+i*35, clr, 18)
	}
	// Back button
	var clrBack color.Color = color.White
	prefixBack := "  "
	if g.characterMenuIndex == len(g.playableCharacterRegistry.IDs) {
		clrBack = color.RGBA{255, 255, 0, 255}
		prefixBack = "> "
	}
	gr.graphics.DrawTextAt(screen, prefixBack+"BACK", 100, 130+len(g.playableCharacterRegistry.IDs)*35+20, clrBack, 18)

	// If hovering but not selecting via keyboard, we can optionally show preview or update index.
	// But usually we want mouse and keyboard to stay in sync.
	if hoverIndex != -1 {
		// We don't update g.characterMenuIndex here directly, but handle it in Game.Update
		// However, for visual feedback we could highlight it.
	}

	msg := "Press UP/DOWN to navigate, ENTER to select hero."
	mw, _ := gr.graphics.MeasureText(msg, 14)
	gr.graphics.DrawTextAt(screen, msg, (g.width-int(mw))/2, g.height-50, color.RGBA{136, 136, 136, 255}, 14)
}

func (gr *GameRenderer) drawHeroPreview(screen engine.Image, char *EntityConfig, x, y int) {
	gr.graphics.DrawTextAt(screen, "--- HERO PROFILE ---", x, y, color.RGBA{218, 165, 32, 255}, 20)

	// Portrait
	if char.StaticImage != nil {
		img := char.StaticImage.(engine.Image)
		op := engine.NewDrawImageOptions()
		op.Scale(1.5, 1.5)
		op.Translate(float64(x), float64(y+30))
		screen.DrawImage(img, op)
	}

	statsX := x + 180
	statsY := y + 40
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Health:  %d", char.Stats.HealthMin), statsX, statsY, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Attack:  %d", char.Stats.BaseAttack), statsX, statsY+25, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Defense: %d", char.Stats.BaseDefense), statsX, statsY+50, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Speed:   %.2f", char.Stats.Speed), statsX, statsY+75, color.White, 16)
	gr.graphics.DrawTextAt(screen, fmt.Sprintf("Weapon:  %s", char.WeaponName), statsX, statsY+100, color.White, 16)

	// Description
	gr.graphics.DrawTextAt(screen, "--- BIOGRAPHY ---", x, y+230, color.RGBA{218, 165, 32, 255}, 20)
	words := strings.Fields(char.Description)
	line := ""
	lineNum := 0
	for _, w := range words {
		if len(line)+len(w) > 40 {
			gr.graphics.DrawTextAt(screen, line, x, y+260+lineNum*20, color.White, 14)
			line = w + " "
			lineNum++
		} else {
			line += w + " "
		}
	}
	gr.graphics.DrawTextAt(screen, line, x, y+260+lineNum*20, color.White, 14)
}

func (gr *GameRenderer) drawGameWon(screen engine.Image) {
	g := gr.game
	// Semi-transparent overlay
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 200}, false)

	title := "YOU WIN!"
	if g.isCampaign {
		title = "CAMPAIGN COMPLETED: YOU WIN!"
	}
	tw, _ := gr.graphics.MeasureText(title, 40)
	gr.graphics.DrawTextAt(screen, title, (g.width-int(tw))/2, 100, color.RGBA{218, 165, 32, 255}, 40)

	options := []string{"Replay", "Quit"}
	for i, opt := range options {
		var clr color.Color = color.White
		prefix := "  "
		if g.mapWonMenuIndex == i {
			clr = color.RGBA{255, 255, 0, 255}
			prefix = "> "
		}
		gr.graphics.DebugPrintAt(screen, prefix+opt, g.width/2-40, 200+i*40, clr)
	}
}

func (gr *GameRenderer) getTileAt(x, y int) engine.Image {
	key := fmt.Sprintf("%d_%d", x, y)

	if tilePath, ok := gr.coordCache[key]; ok {
		return gr.tileCache[tilePath]
	}

	resolvedTile := gr.game.currentMapType.FloorTile
	highestPriority := -1

	for _, zone := range gr.game.currentMapType.FloorZones {
		if zone.Priority > highestPriority {
			if zone.GetPolygon().Contains(float64(x), float64(y)) {
				resolvedTile = zone.Tile
				highestPriority = zone.Priority
			}
		}
	}

	gr.coordCache[key] = resolvedTile

	if _, ok := gr.tileCache[resolvedTile]; !ok {
		floorPath := path.Join("assets/images/floors", resolvedTile)
		loaded := gr.graphics.LoadSprite(gr.game.assets, floorPath, true)
		gr.tileCache[resolvedTile] = loaded
	}

	return gr.tileCache[resolvedTile]
}
