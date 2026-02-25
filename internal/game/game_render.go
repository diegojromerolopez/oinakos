package game

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"log"
	"math"
	"path"
	"sort"

	"oinakos/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// GameRenderer handles the Ebiten-dependent rendering of the game.
type GameRenderer struct {
	game         *Game
	renderer     *engine.Renderer
	graphics     engine.Graphics
	grassSprite  engine.Image
	portalSprite engine.Image
	crownSprite  engine.Image
	zoneSprite   engine.Image
	emptyImage   engine.Image
}

func NewGameRenderer(g *Game, assets fs.FS, graphics engine.Graphics) *GameRenderer {
	gr := &GameRenderer{
		game:         g,
		renderer:     engine.NewRenderer(),
		graphics:     graphics,
		grassSprite:  graphics.LoadSprite(assets, "assets/images/environment/grass_tile.png", true),
		portalSprite: graphics.LoadSprite(assets, "assets/images/scenario/portal.png", true),
		crownSprite:  graphics.LoadSprite(assets, "assets/images/scenario/crown.png", true),
		zoneSprite:   graphics.LoadSprite(assets, "assets/images/scenario/zone_marker.png", true),
	}
	gr.emptyImage = graphics.NewImage(3, 3)
	gr.emptyImage.Fill(color.White)
	return gr
}

func (gr *GameRenderer) LoadAssets(assets fs.FS) {
	gr.game.archetypeRegistry.LoadAssets(assets, gr.graphics)
	gr.game.obstacleRegistry.LoadAssets(assets, gr.graphics)

	// Load player assets
	mc := gr.game.mainCharacter
	if mc != nil && mc.Config != nil {
		imgDir := "assets/images/characters/main"
		mc.Config.StaticImage = gr.graphics.LoadSprite(assets, path.Join(imgDir, mc.Config.Sprites.Static), true)
		mc.Config.CorpseImage = gr.graphics.LoadSprite(assets, path.Join(imgDir, mc.Config.Sprites.Corpse), true)
		if mc.Config.Sprites.Attack != "" {
			mc.Config.AttackImage = gr.graphics.LoadSprite(assets, path.Join(imgDir, mc.Config.Sprites.Attack), true)
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

	if gr.grassSprite != nil {
		gr.renderer.DrawInfiniteGrass(screen, offsetX, offsetY, gr.grassSprite)
	}

	// Collect all drawable entities for Y-sorting
	type drawTask struct {
		y    float64
		draw func()
	}
	tasks := make([]drawTask, 0, len(g.obstacles)+len(g.npcs)+1)

	// Add obstacles
	for _, o := range g.obstacles {
		isoX, isoY := engine.CartesianToIso(o.X, o.Y)
		drawX := isoX + offsetX
		drawY := isoY + offsetY
		// Aggressive Culling for obstacles
		if drawX < -256 || drawX > float64(g.width)+256 || drawY < -256 || drawY > float64(g.height)+256 {
			continue
		}

		obj := o // local copy
		tasks = append(tasks, drawTask{
			y: obj.X + obj.Y,
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
		tasks = append(tasks, drawTask{
			y: npc.X + npc.Y,
			draw: func() {
				npc.Draw(screen, gr.graphics, gr.graphics, offsetX, offsetY)
			},
		})
	}

	// Add mainCharacter
	tasks = append(tasks, drawTask{
		y: g.mainCharacter.X + g.mainCharacter.Y,
		draw: func() {
			g.mainCharacter.Draw(screen, offsetX, offsetY)
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

	// Execute draw tasks
	for _, t := range tasks {
		t.draw()
	}

	// Draw floating texts (always on top of entities)
	for _, ft := range g.floatingTexts {
		ft.Draw(screen, gr.graphics, offsetX, offsetY)
	}

	// Draw map target points for navigation objectives
	switch g.currentMapType.Type {
	case ObjReachPortal:
		tasks = append(tasks, drawTask{
			y: g.currentMapType.TargetPoint.X + g.currentMapType.TargetPoint.Y,
			draw: func() {
				isoX, isoY := engine.CartesianToIso(g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y)
				op := engine.NewDrawImageOptions()
				op.Translate(isoX+offsetX-32, isoY+offsetY-16)
				screen.DrawImage(gr.portalSprite, op)
			},
		})
	case ObjReachZone, ObjProtectNPC:
		// Draw zone under everything, no need to y-sort.
		isoX, isoY := engine.CartesianToIso(g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y)
		op := engine.NewDrawImageOptions()
		op.Translate(isoX+offsetX-float64(gr.zoneSprite.Bounds().Dx())/2, isoY+offsetY-float64(gr.zoneSprite.Bounds().Dy())/2)
		// Draw as background element
		screen.DrawImage(gr.zoneSprite, op)
	case ObjReachBuilding:
		// Draw light column on top of the target point
		tasks = append(tasks, drawTask{
			y: g.currentMapType.TargetPoint.X + g.currentMapType.TargetPoint.Y,
			draw: func() {
				isoX, isoY := engine.CartesianToIso(g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y)
				op := engine.NewDrawImageOptions()
				op.Scale(2, 8)
				op.Translate(isoX+offsetX-16, isoY+offsetY-128) // Adjust for scaled size
				op.SetColorScale(0.5, 0.5, 0.5, 0.5)            // Alpha scaling
				screen.DrawImage(gr.crownSprite, op)
			},
		})
	case ObjKillVIP:
		if len(g.npcs) > 0 && g.npcs[0].IsAlive() {
			tasks = append(tasks, drawTask{
				y: g.npcs[0].X + g.npcs[0].Y + 0.1, // always slightly in front of the NPC
				draw: func() {
					isoX, isoY := engine.CartesianToIso(g.npcs[0].X, g.npcs[0].Y)
					op := engine.NewDrawImageOptions()
					op.Translate(isoX+offsetX-16, isoY+offsetY-60) // Float above head
					screen.DrawImage(gr.crownSprite, op)
				},
			})
		}
	}

	// Draw crown for boss NPCs
	for _, n := range gr.game.npcs {
		if n.IsBoss && n.IsAlive() {
			isoX, isoY := engine.CartesianToIso(n.X, n.Y)
			if gr.crownSprite != nil {
				op := engine.NewDrawImageOptions()
				op.Translate(isoX+offsetX-16, isoY+offsetY-40)
				screen.DrawImage(gr.crownSprite, op)
			}
		}
	}

	if g.isGameOver {
		gr.drawGameOver(screen)
	} else if g.isMapWon {
		gr.drawMapWon(screen)
	} else if g.isPaused {
		gr.drawPauseMenu(screen)
	} else {
		gr.drawHUD(screen)
	}
}

func (gr *GameRenderer) drawPauseMenu(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)
	gr.graphics.DebugPrintAt(screen, "GAME PAUSED", g.width/2-40, g.height/2-30)
	gr.graphics.DebugPrintAt(screen, "Press S to SAVE and QUIT", g.width/2-70, g.height/2)
	gr.graphics.DebugPrintAt(screen, "Press any other key to RESUME", g.width/2-95, g.height/2+20)
}

func (gr *GameRenderer) drawGameOver(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 180}, false)
	minutes := int(g.playTime) / 60
	seconds := int(g.playTime) % 60
	gr.graphics.DebugPrintAt(screen, "GAME OVER", g.width/2-30, g.height/2-45)
	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("Kills: %d", g.mainCharacter.Kills), g.width/2-25, g.height/2-15)
	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("Time: %02d:%02d", minutes, seconds), g.width/2-30, g.height/2+15)
	gr.graphics.DebugPrintAt(screen, "Press ESC to exit, or ENTER to restart", g.width/2-110, g.height/2+45)
}

func (gr *GameRenderer) drawMapWon(screen engine.Image) {
	g := gr.game
	gr.graphics.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{20, 60, 20, 200}, false)
	mapKillTotal := 0
	for _, k := range g.mainCharacter.MapKills {
		mapKillTotal += k
	}
	gr.graphics.DebugPrintAt(screen, "MAP WON!", g.width/2-30, g.height/2-45)
	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("Map Kills: %d", mapKillTotal), g.width/2-40, g.height/2-15)
	gr.graphics.DebugPrintAt(screen, "Press ENTER to continue, ESC to quit", g.width/2-110, g.height/2+45)
}

func (gr *GameRenderer) drawHUD(screen engine.Image) {
	g := gr.game
	// Use DrawFilledRect instead of NewImage every frame to avoid Metal leaks
	gr.graphics.DrawFilledRect(screen, 10, 10, 350, 150, color.RGBA{0, 0, 0, 180}, false)

	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("HP: %d/%d", g.mainCharacter.Health, g.mainCharacter.MaxHealth), 20, 20)

	// Health bar background
	gr.graphics.DrawFilledRect(screen, 100, 22, 200, 10, color.RGBA{100, 0, 0, 255}, false)

	healthPct := float64(g.mainCharacter.Health) / float64(g.mainCharacter.MaxHealth)
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

	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("LVL: %d  XP: %d", g.mainCharacter.Level, g.mainCharacter.XP), 20, 45)
	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("OBJ: %s", g.currentMapType.Description), 20, 57)
	minutes := int(g.playTime) / 60
	seconds := int(g.playTime) % 60
	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("POS %.1f,%.1f  KILLS: %d  XP: %d  LVL: %d", g.mainCharacter.X, g.mainCharacter.Y, g.mainCharacter.Kills, g.mainCharacter.XP, g.mainCharacter.Level), 20, 77)
	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("ATK: %d  DEF: %d  SHIELD: %d", g.mainCharacter.GetTotalAttack(), g.mainCharacter.GetTotalDefense(), g.mainCharacter.GetTotalProtection()), 20, 92)
	weaponText := fmt.Sprintf("WEAPON: %s (%d-%d)", g.mainCharacter.Weapon.Name, g.mainCharacter.Weapon.MinDamage, g.mainCharacter.Weapon.MaxDamage)
	if g.mainCharacter.Weapon.Bonus > 0 {
		weaponText += fmt.Sprintf(" +%d", g.mainCharacter.Weapon.Bonus)
	}
	gr.graphics.DebugPrintAt(screen, weaponText, 20, 107)
	gr.graphics.DebugPrintAt(screen, fmt.Sprintf("TIME: %02d:%02d", minutes, seconds), 20, 122)

	var tx, ty float64
	hasTarget := false
	switch g.currentMapType.Type {
	case ObjKillVIP:
		if len(g.npcs) > 0 && g.npcs[0].IsAlive() {
			tx, ty = g.npcs[0].X, g.npcs[0].Y
			hasTarget = true
		}
	case ObjProtectNPC:
		if len(g.npcs) > 0 && g.npcs[0].IsAlive() {
			tx, ty = g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y
			hasTarget = true
		}
	case ObjDestroyBuilding, ObjReachBuilding, ObjReachPortal, ObjReachZone:
		tx, ty = g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y
		hasTarget = true
	}

	if hasTarget {
		dx := tx - g.mainCharacter.X
		dy := ty - g.mainCharacter.Y
		isoDx := dx - dy
		isoDy := (dx + dy) * 0.5
		angle := math.Atan2(isoDy, isoDx)
		arrowX := float32(g.width - 50)
		arrowY := float32(50)
		size := float32(20.0)
		var path vector.Path
		path.MoveTo(size, 0)
		path.LineTo(-size, -size*0.6)
		path.LineTo(-size*0.5, 0)
		path.LineTo(-size, size*0.6)
		path.Close()
		opArr := &ebiten.DrawTrianglesOptions{}
		opArr.FillRule = ebiten.EvenOdd
		cosA := float32(math.Cos(angle))
		sinA := float32(math.Sin(angle))
		var vs []ebiten.Vertex
		var is []uint16
		vs, is = path.AppendVerticesAndIndicesForFilling(vs, is)
		for i := range vs {
			rx := vs[i].DstX*cosA - vs[i].DstY*sinA
			ry := vs[i].DstX*sinA + vs[i].DstY*cosA
			vs[i].DstX = rx + arrowX
			vs[i].DstY = ry + arrowY
			vs[i].SrcX = 0
			vs[i].SrcY = 0
			vs[i].ColorR = 1.0
			vs[i].ColorG = 0.2
			vs[i].ColorB = 0.2
			vs[i].ColorA = 1.0
		}

		var evs []engine.Vertex
		for _, v := range vs {
			evs = append(evs, engine.Vertex{
				DstX:   v.DstX,
				DstY:   v.DstY,
				SrcX:   v.SrcX,
				SrcY:   v.SrcY,
				ColorR: v.ColorR,
				ColorG: v.ColorG,
				ColorB: v.ColorB,
				ColorA: v.ColorA,
			})
		}
		opTri := &engine.DrawTrianglesOptions{
			FillRule: engine.FillRuleEvenOdd,
		}

		screen.DrawTriangles(evs, is, gr.emptyImage.SubImage(image.Rect(1, 1, 2, 2)), opTri)
		gr.graphics.DebugPrintAt(screen, "OBJ", int(arrowX)-10, int(arrowY)+25)
	}
}
