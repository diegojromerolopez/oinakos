package game

import (
	"image/color"
	"io/fs"
	"log"
	"path"
	"time"
	"runtime"
	"sort"

	"oinakos/internal/engine"
	"sync/atomic"
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
	start := time.Now()
	g := gr.game
	g.LoadingMessage = "Initializing Oinakos..."
	atomic.StoreInt32(&g.LoadingProgress, 50)
	
	s := time.Now()
	// Always load all archetypes (small 10 items)
	// Passing nil permitList loads all that aren't already loaded
	g.archetypeRegistry.LoadAssets(assets, gr.graphics, nil, &g.LoadingProgress)
	log.Printf("[LOADER] Archetypes loaded in %v", time.Since(s))
	
	s = time.Now()
	atomic.StoreInt32(&g.LoadingProgress, 300)
	g.LoadingMessage = "Preparing Heroes..."
	// Load all playable characters (17 items)
	g.characterRegistry.LoadAssets(assets, gr.graphics, g.archetypeRegistry, nil, &g.LoadingProgress)
	log.Printf("[LOADER] Heroes loaded in %v", time.Since(s))
	
	s = time.Now()
	atomic.StoreInt32(&g.LoadingProgress, 500)
	g.LoadingMessage = "Loading Environment..."
	// Only load obstacles that are in the map or are wells
	obstacleFilter := make(map[string]bool)
	if g.World != nil {
		for _, o := range g.obstacles {
			obstacleFilter[o.Archetype.ID] = true
		}
	}
	g.obstacleRegistry.LoadAssets(assets, gr.graphics, obstacleFilter, &g.LoadingProgress)
	log.Printf("[LOADER] Obstacles loaded in %v", time.Since(s))
	
	s = time.Now()
	atomic.StoreInt32(&g.LoadingProgress, 700)
	g.LoadingMessage = "Populating World..."
	// Character Registry already loaded all playable. 
	// Now load NPCs present in map (characters list)
	npcFilter := make(map[string]bool)
	if g.World != nil {
		for _, n := range g.characters {
			if n.Config != nil {
				npcFilter[n.Config.ID] = true
			}
		}
	}
	g.characterRegistry.LoadAssets(assets, gr.graphics, g.archetypeRegistry, npcFilter, &g.LoadingProgress)
	log.Printf("[LOADER] World NPCs loaded in %v", time.Since(s))
	
	s = time.Now()
	atomic.StoreInt32(&g.LoadingProgress, 800)
	g.LoadingMessage = "Geeking out on Loot..."
	// TODO: Filter object assets too?
	g.Registries.Objects.LoadAssets(assets, gr.graphics, nil, &g.LoadingProgress)
	log.Printf("[LOADER] Objects loaded in %v", time.Since(s))
	
	s = time.Now()
	atomic.StoreInt32(&g.LoadingProgress, 950)
	g.LoadingMessage = "Finalizing Graphics..."
	runtime.Gosched()

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
		addJob := func(filename string, target *engine.Image) {
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
	loadSpritesParallel(assets, jobs, gr.graphics, &g.LoadingProgress)
}
	}
	log.Printf("[LOADER] Total initialization finished in %v", time.Since(start))
	atomic.StoreInt32(&g.LoadingProgress, 1000)
}

func (gr *GameRenderer) Draw(screen engine.Image) {
	if screen == nil {
		log.Println("GameRenderer.Draw called with nil screen!")
		return
	}
	g := gr.game
	if atomic.LoadInt32(&g.LoadingProgress) < 1000 {
		gr.drawLoadingProgress(screen)
		return
	}
	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)

	if g.isMainMenu {
		gr.drawMainMenu(screen)
	} else if g.isCharacterSelect {
		gr.drawCharacterSelect(screen)
	} else if g.isCampaignSelect {
		gr.drawCampaignSelect(screen)
	} else if g.isAboutScreen {
		gr.drawAboutScreen(screen)
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

		// 1. Dedicated Object Pass (Floor Items)
		if g.World != nil {
			for _, it := range g.World.Items {
				if it == nil {
					continue
				}
				if g.debug {
					// Draw green silhouette
					isoX, isoY := engine.CartesianToIso(it.X, it.Y)
					if it.Config != nil && it.Config.Sprite != nil {
						w, h := it.Config.Sprite.Size()
						op := engine.NewDrawImageOptions()
						tx := isoX + offsetX - float64(w)/2
						ty := isoY + offsetY - float64(h)*0.9 // Grounded anchoring to match item_instance.go
						op.Translate(tx, ty)
						op.SetColorScale(0, 10, 0, 1) // Pure green silhouette
						screen.DrawImage(it.Config.Sprite, op)
					}
				} else {
					it.Draw(screen, offsetX, offsetY)
				}
			}
		}

		// 2. Sorted Entity Pass (Actors, Obstacles, Projectiles)
		type drawTask struct {
			y    float64
			draw func()
		}
		tasks := make([]drawTask, 0, len(g.obstacles)+len(g.characters)+2)


		for _, o := range g.obstacles {
			img := o.Archetype.Image
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
			sortY := o.GetSortY()

			tasks = append(tasks, drawTask{
				y: sortY,
				draw: func() {
					obj.Draw(screen, gr.graphics, offsetX, offsetY)
				},
			})
		}

		for _, n := range g.characters {
			isoX, isoY := engine.CartesianToIso(n.X, n.Y)
			drawX := isoX + offsetX
			drawY := isoY + offsetY
			if drawX < -256 || drawX > float64(g.width)+256 || drawY < -256 || drawY > float64(g.height)+256 {
				continue
			}

			npc := n
			sortY := n.GetSortY()
			tasks = append(tasks, drawTask{
				y: sortY,
				draw: func() {
					DrawActor(&npc.Actor, screen, gr.graphics, gr.graphics, gr.PaletteShader, offsetX, offsetY, npc.IsPlayerControlled)
				},
			})
		}

		mcSortY := g.playableCharacter.GetSortY()
		tasks = append(tasks, drawTask{
			y: mcSortY,
			draw: func() {
				DrawActor(&g.playableCharacter.Actor, screen, gr.graphics, gr.graphics, gr.PaletteShader, offsetX, offsetY, true)
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

		// UI Pass: Draw character and NPC UI on top (always visible)
		for _, n := range g.characters {
			isoX, isoY := engine.CartesianToIso(n.X, n.Y)
			drawX := isoX + offsetX
			drawY := isoY + offsetY
			if drawX < -256 || drawX > float64(g.width)+256 || drawY < -256 || drawY > float64(g.height)+256 {
				continue
			}
			DrawActorUI(g, &n.Actor, screen, gr.graphics, gr.graphics, offsetX, offsetY, n.IsPlayerControlled, g.debug)
		}
		DrawActorUI(g, &g.playableCharacter.Actor, screen, gr.graphics, gr.graphics, offsetX, offsetY, true, g.debug)

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
		} else if g.isInventoryOpen {
			gr.drawInventoryScreen(screen)
		} else {
			gr.drawFog(screen)
			gr.drawWeather(screen)
			gr.drawHUD(screen)
			gr.drawDialogueBox(screen)
			gr.drawHoverInfo(screen)
		}
	}

	if g.isQuitConfirmationOpen {
		gr.drawQuitConfirmation(screen)
	}
}

func (gr *GameRenderer) drawWeather(screen engine.Image) {
	g := gr.game
	if g.CurrentWeather == WeatherClear {
		return
	}

	// 1. World Overlay (Tint)
	if g.CurrentWeather == WeatherRain || g.CurrentWeather == WeatherStorm {
		// Slight grey/blue tint
		gr.emptyImage.Fill(color.NRGBA{50, 50, 80, 40}) // Semi-transparent tint
		op := engine.NewDrawImageOptions()
		op.Scale(float64(g.width)/3, float64(g.height)/3) // emptyImage is 3x3
		screen.DrawImage(gr.emptyImage, op)
	}

	// 2. Lightning Flash
	if g.CurrentWeather == WeatherStorm && g.Tick%600 < 5 {
		// Flash for 5 frames every 10 seconds approx
		alpha := uint8(100 - (g.Tick%600)*20)
		gr.emptyImage.Fill(color.NRGBA{255, 255, 255, alpha})
		op := engine.NewDrawImageOptions()
		op.Scale(float64(g.width)/3, float64(g.height)/3)
		screen.DrawImage(gr.emptyImage, op)
	}

	// 3. Particles
	for _, p := range g.particles {
		op := engine.NewDrawImageOptions()
		op.Translate(p.X, p.Y)

		c := color.NRGBA{255, 255, 255, 200}
		if p.Type == ParticleRain {
			c = color.NRGBA{150, 150, 255, 180}
			// Draw rain as a line
			gr.emptyImage.Fill(c)
			op.Scale(0.5, 4.0) // Thin and tall
		} else {
			// Draw snow as a dot
			gr.emptyImage.Fill(c)
			op.Scale(1.0, 1.0)
		}
		screen.DrawImage(gr.emptyImage, op)
	}
}
