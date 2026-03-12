package main

import (
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"

	"oinakos/internal/engine"
	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func NewViewer(entities []*EditorEntity, g engine.Graphics, input engine.Input, w, h int) *Viewer {
	return &Viewer{
		entities:      entities,
		selectedIndex: 0,
		graphics:      g,
		input:         input,
		width:         w,
		height:        h,
		draggingIdx:   -1,
		hoverIdx:      -1,
		addBtnRect:    image.Rect(sidebarWidth+10, 60, sidebarWidth+110, 90),
	}
}

func (v *Viewer) Update() error {
	if v.input.IsKeyPressed(engine.KeyEscape) || v.input.IsKeyPressed(engine.KeyQ) {
		return ebiten.Termination
	}

	mx, my := v.input.MousePosition()
	if v.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) && mx < sidebarWidth {
		slotIdx := (my - v.scrollOffset) / slotHeight
		if slotIdx >= 0 && slotIdx < len(v.entities) {
			v.selectedIndex = slotIdx
			v.draggingIdx = -1
			v.hoverIdx = -1
			v.camX, v.camY = 0, 0
		}
	}

	_, wheelY := v.input.Wheel()
	if mx < sidebarWidth {
		v.scrollOffset += int(wheelY * 20)
		if v.scrollOffset > 0 { v.scrollOffset = 0 }
		maxScroll := -(len(v.entities)*slotHeight - v.height)
		if maxScroll < 0 && v.scrollOffset < maxScroll { v.scrollOffset = maxScroll } else if maxScroll >= 0 { v.scrollOffset = 0 }
	}

	if mx >= sidebarWidth {
		if v.input.IsKeyPressed(engine.KeyUp) { v.camY -= cameraSpeed }
		if v.input.IsKeyPressed(engine.KeyDown) { v.camY += cameraSpeed }
		if v.input.IsKeyPressed(engine.KeyLeft) { v.camX -= cameraSpeed }
		if v.input.IsKeyPressed(engine.KeyRight) { v.camX += cameraSpeed }
	}

	if v.selectedIndex < 0 || v.selectedIndex >= len(v.entities) { return nil }

	baseX, baseY := sidebarWidth+float64(v.width-sidebarWidth)/2, float64(v.height)*0.6
	offsetX := baseX - v.camX
	offsetY := baseY - v.camY
	ee := v.entities[v.selectedIndex]

	if ee.Footprint != nil && len(*ee.Footprint) == 0 {
		poly := ee.GetFootprint()
		for _, p := range poly.Points {
			*ee.Footprint = append(*ee.Footprint, game.FootprintPoint{X: p.X, Y: p.Y})
		}
	}

	v.hoverIdx = -1
	if ee.Footprint != nil {
		for i, p := range *ee.Footprint {
			ix, iy := engine.CartesianToIso(p.X, p.Y)
			px, py := ix+offsetX, iy+offsetY
			if math.Hypot(float64(mx)-px, float64(my)-py) < clickThreshold {
				v.hoverIdx = i
				break
			}
		}
	}

	if v.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) && mx >= sidebarWidth {
		if v.input.IsKeyPressed(engine.KeyControl) || v.input.IsKeyPressed(engine.KeyMeta) {
			cx, cy := engine.IsoToCartesian(float64(mx)-offsetX, float64(my)-offsetY)
			*ee.Footprint = append(*ee.Footprint, game.FootprintPoint{X: math.Round(cx*100)/100, Y: math.Round(cy*100)/100})
			v.saveToYAML(ee)
			return nil
		}
		if image.Pt(mx, my).In(v.addBtnRect) {
			v.addPoint(ee)
			return nil
		}
		if v.hoverIdx != -1 {
			if v.input.IsKeyPressed(engine.KeyShift) {
				v.removePoint(ee, v.hoverIdx)
				return nil
			}
			v.draggingIdx = v.hoverIdx
		}
	}

	if v.draggingIdx != -1 {
		if v.input.IsMouseButtonPressed(engine.MouseButtonLeft) {
			cx, cy := engine.IsoToCartesian(float64(mx)-offsetX, float64(my)-offsetY)
			(*ee.Footprint)[v.draggingIdx].X = math.Round(cx*100) / 100
			(*ee.Footprint)[v.draggingIdx].Y = math.Round(cy*100) / 100
		} else {
			v.draggingIdx = -1
			v.saveToYAML(ee)
		}
	}
	return nil
}

func (v *Viewer) Layout(_, _ int) (int, int) {
	return v.width, v.height
}

func main() {
	assets := os.DirFS(".")
	graphics := engine.NewEbitenGraphics()
	var entities []*EditorEntity

	obsReg := game.NewObstacleRegistry()
	if err := obsReg.LoadAll(assets); err == nil {
		obsReg.LoadAssets(assets, graphics, nil)
		for _, id := range obsReg.IDs {
			arch := obsReg.Archetypes[id]
			obs := game.NewObstacle("editor_preview", 0, 0, arch)
			var img engine.Image
			if arch.Image != nil { img = arch.Image.(engine.Image) }
			entities = append(entities, &EditorEntity{
				ID: id, Type: "Obstacle", Image: img, Footprint: &arch.Footprint,
				YamlPath: filepath.Join("data/obstacles", id+".yaml"),
				DrawMain: func(screen engine.Image, g engine.Graphics, ox, oy float64) { obs.Draw(screen, g, ox, oy) },
			})
		}
	}

	npcReg := game.NewArchetypeRegistry()
	if err := npcReg.LoadAll(assets); err == nil {
		npcReg.LoadAssets(assets, graphics, nil)
		for _, id := range npcReg.IDs {
			arch := npcReg.Archetypes[id]
			npc := game.NewNPC(0, 0, arch, 1)
			var img engine.Image
			if arch.StaticImage != nil { img = arch.StaticImage.(engine.Image) }
			entities = append(entities, &EditorEntity{
				ID: id, Type: "NPC", Image: img, Footprint: &arch.Footprint, YamlPath: findArchetypeYAML(id),
				DrawMain: func(screen engine.Image, g engine.Graphics, ox, oy float64) { npc.Draw(screen, g, g, nil, ox, oy) },
			})
		}
	}

	mcConfig, err := game.LoadPlayableCharacterConfig(assets)
	if err == nil {
		mcConfig.AssetDir = "assets/images/characters/oinakos"
		mcConfig.StaticImage = graphics.LoadSprite(assets, filepath.Join(mcConfig.AssetDir, "static.png"), true)
		mcConfig.CorpseImage = graphics.LoadSprite(assets, filepath.Join(mcConfig.AssetDir, "corpse.png"), true)
		mcConfig.AttackImage = graphics.LoadSprite(assets, filepath.Join(mcConfig.AssetDir, "attack.png"), true)
		mc := game.NewPlayableCharacter(0, 0, mcConfig)
		var img engine.Image
		if mcConfig.StaticImage != nil { img = mcConfig.StaticImage.(engine.Image) }
		entities = append(entities, &EditorEntity{
			ID: "playable_character", Type: "Character", Image: img, Footprint: &mcConfig.Footprint,
			YamlPath: "data/characters/oinakos.yaml",
			DrawMain: func(screen engine.Image, g engine.Graphics, ox, oy float64) { mc.Draw(screen, g, g, ox, oy) },
		})
	}

	sort.Slice(entities, func(i, j int) bool {
		if entities[i].Type != entities[j].Type { return entities[i].Type < entities[j].Type }
		return entities[i].ID < entities[j].ID
	})

	viewer := NewViewer(entities, graphics, engine.NewEbitenInput(), defaultScreenWidth, defaultScreenHeight)
	ebiten.SetWindowTitle("Oinakos Boundary Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(viewer); err != nil { log.Fatal(err) }
}
