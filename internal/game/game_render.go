package game

import (
	"image/color"
	"io/fs"
	"log"
	"path"
	"sort"

	"oinakos/internal/engine"
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

	fogImage   engine.Image
	torchImage engine.Image
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

	gr.torchImage = generateTorchImage(graphics, 250)
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
		var jobs []*SpriteLoadJob
		addJob := func(filename string, target *interface{}) {
			if *target == nil {
				jobs = append(jobs, &SpriteLoadJob{
					Path: path.Join(imgDir, filename),
					Dest: target,
				})
			}
		}

		addJob("static.png", &mc.Config.StaticImage)
		addJob("back.png", &mc.Config.BackImage)
		addJob("corpse.png", &mc.Config.CorpseImage)
		addJob("attack.png", &mc.Config.AttackImage)
		addJob("attack1.png", &mc.Config.Attack1Image)
		addJob("attack2.png", &mc.Config.Attack2Image)
		addJob("hit.png", &mc.Config.HitImage)
		addJob("hit1.png", &mc.Config.Hit1Image)
		addJob("hit2.png", &mc.Config.Hit2Image)

		if len(jobs) > 0 {
			loadSpritesParallel(assets, jobs, gr.graphics)
		}
	}
}

func (gr *GameRenderer) Draw(screen engine.Image) {
	if screen == nil {
		log.Println("GameRenderer.Draw called with nil screen!")
		return
	}
	g := gr.game
	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)

	if g.isMainMenu {
		gr.drawMainMenu(screen)
	} else if g.isCharacterSelect {
		gr.drawCharacterSelect(screen)
	} else if g.isCampaignSelect {
		gr.drawCampaignSelect(screen)
	} else if g.isSettingsScreen {
		gr.drawSettingsScreen(screen)
	} else {
		if g.currentMapType.FloorTile != "" && g.currentMapType.FloorTile != gr.lastFloorPath {
			gr.coordCache = make(map[string]string)
			gr.lastFloorPath = g.currentMapType.FloorTile
		}

		gr.renderer.DrawTileMap(screen, offsetX, offsetY, func(x, y int) engine.Image {
			return gr.getTileAt(x, y)
		})

		type drawTask struct {
			y    float64
			draw func()
		}
		tasks := make([]drawTask, 0, len(g.obstacles)+len(g.npcs)+1)

		for _, o := range g.obstacles {
			img, _ := o.Archetype.Image.(engine.Image)
			if img == nil {
				continue
			}
			sw, sh := img.Size()

			isoX, isoY := engine.CartesianToIso(o.X, o.Y)
			drawX := isoX + offsetX
			drawY := isoY + offsetY

			marginW := float64(sw)
			marginH := float64(sh)
			if drawX < -marginW || drawX > float64(g.width)+marginW || drawY < -marginH || drawY > float64(g.height)+marginH {
				continue
			}

			obj := o
			sortY := obj.X + obj.Y
			if obj.Archetype.Type == "static" || obj.Archetype.Type == "well" {
				sortY += 2.0
			} else {
				p := obj.GetFootprint()
				if len(p.Points) > 0 {
					var minX, maxX, minY, maxY float64
					for i, pt := range p.Points {
						if i == 0 || pt.X < minX { minX = pt.X }
						if i == 0 || pt.X > maxX { maxX = pt.X }
						if i == 0 || pt.Y < minY { minY = pt.Y }
						if i == 0 || pt.Y > maxY { maxY = pt.Y }
					}
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

		for _, n := range g.npcs {
			isoX, isoY := engine.CartesianToIso(n.X, n.Y)
			drawX := isoX + offsetX
			drawY := isoY + offsetY
			if drawX < -256 || drawX > float64(g.width)+256 || drawY < -256 || drawY > float64(g.height)+256 {
				continue
			}

			npc := n
			sortY := npc.X + npc.Y
			if npc.State == NPCDead {
				sortY -= 100.0
			}
			tasks = append(tasks, drawTask{
				y: sortY,
				draw: func() {
					npc.Draw(screen, gr.graphics, gr.graphics, gr.PaletteShader, offsetX, offsetY)
				},
			})
		}

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

		for _, p := range g.projectiles {
			proj := p
			tasks = append(tasks, drawTask{
				y: proj.X + proj.Y,
				draw: func() {
					proj.Draw(screen, gr.graphics, offsetX, offsetY)
				},
			})
		}

		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].y < tasks[j].y
		})

		for _, t := range tasks {
			t.draw()
		}

		if g.debug || g.showBoundaries {
			gr.drawDebug(screen, offsetX, offsetY)
		}

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
			gr.drawFog(screen)
			gr.drawHUD(screen)
			gr.drawHoverInfo(screen)
		}
	}

	if g.isQuitConfirmationOpen {
		gr.drawQuitConfirmation(screen)
	}
}
