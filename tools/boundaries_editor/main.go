// Package main provides a standalone tool (boundaries_editor) for visualizing
// entity boundaries (obstacles, NPCs, and the main character) and their
// collision footprints in isometric space. It supports interactive vertex editing.
package main

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"

	"oinakos/internal/engine"
	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"gopkg.in/yaml.v3"
)

// ─── Constants ───────────────────────────────────────────────────────────────

const (
	defaultScreenWidth  = 1280
	defaultScreenHeight = 720
	cameraSpeed         = 5.0
	vertexRadius        = 3.0
	polyLineWidth       = 2.0
	clickThreshold      = 10.0
	sidebarWidth        = 240
	thumbnailSize       = 64
	slotHeight          = 80
)

var (
	backgroundColor = color.RGBA{20, 20, 20, 255}
	sidebarColor    = color.RGBA{40, 40, 40, 255}
	footprintColor  = color.RGBA{0, 255, 255, 255} // Cyan
	vertexColor     = color.RGBA{255, 255, 0, 255}
	hoverColor      = color.RGBA{255, 255, 255, 255}
	selectedColor   = color.RGBA{0, 255, 0, 255} // Green border
	buttonColor     = color.RGBA{60, 60, 60, 255}
)

// ─── EditorEntity ────────────────────────────────────────────────────────────

type EditorEntity struct {
	ID        string
	Type      string
	Image     engine.Image
	Footprint *[]game.FootprintPoint
	YamlPath  string
	DrawMain  func(screen engine.Image, graphics engine.Graphics, offsetX, offsetY float64)
}

func (e *EditorEntity) GetFootprint() engine.Polygon {
	if e.Footprint == nil || len(*e.Footprint) == 0 {
		// Fallback for missing footprint
		return engine.Polygon{Points: []engine.Point{
			{X: -0.15, Y: -0.15}, {X: 0.15, Y: -0.15},
			{X: 0.15, Y: 0.15}, {X: -0.15, Y: 0.15},
		}}
	}
	poly := engine.Polygon{Points: make([]engine.Point, len(*e.Footprint))}
	for i, p := range *e.Footprint {
		poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
	}
	return poly
}

// ─── Viewer ───────────────────────────────────────────────────────────────────

type Viewer struct {
	entities      []*EditorEntity
	selectedIndex int
	graphics      engine.Graphics
	width, height int
	camX, camY    float64
	draggingIdx   int
	hoverIdx      int
	scrollOffset  int
	addBtnRect    image.Rectangle
}

func NewViewer(entities []*EditorEntity, g engine.Graphics, w, h int) *Viewer {
	return &Viewer{
		entities:      entities,
		selectedIndex: 0,
		graphics:      g,
		width:         w,
		height:        h,
		draggingIdx:   -1,
		hoverIdx:      -1,
		addBtnRect:    image.Rect(sidebarWidth+10, 60, sidebarWidth+110, 90),
	}
}

func (v *Viewer) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	// Sidebar interaction
	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && mx < sidebarWidth {
		slotIdx := (my - v.scrollOffset) / slotHeight
		if slotIdx >= 0 && slotIdx < len(v.entities) {
			v.selectedIndex = slotIdx
			v.draggingIdx = -1
			v.hoverIdx = -1
			v.camX, v.camY = 0, 0
		}
	}

	// Scroll sidebar
	_, wheelY := ebiten.Wheel()
	if mx < sidebarWidth {
		v.scrollOffset += int(wheelY * 20)
		if v.scrollOffset > 0 {
			v.scrollOffset = 0
		}
		maxScroll := -(len(v.entities)*slotHeight - v.height)
		if maxScroll < 0 && v.scrollOffset < maxScroll {
			v.scrollOffset = maxScroll
		} else if maxScroll >= 0 {
			v.scrollOffset = 0
		}
	}

	// Camera pan in main view
	if mx >= sidebarWidth {
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			v.camY -= cameraSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			v.camY += cameraSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			v.camX -= cameraSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyRight) {
			v.camX += cameraSpeed
		}
	}

	if v.selectedIndex < 0 || v.selectedIndex >= len(v.entities) {
		return nil
	}

	mx, my = ebiten.CursorPosition()
	mpos := image.Point{mx, my}
	baseX, baseY := sidebarWidth+float64(v.width-sidebarWidth)/2, float64(v.height)*0.6
	offsetX := baseX - v.camX
	offsetY := baseY - v.camY

	ee := v.entities[v.selectedIndex]

	// Auto-initialize empty footprints
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
			dist := math.Hypot(float64(mx)-px, float64(my)-py)
			if dist < clickThreshold {
				v.hoverIdx = i
				break
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && mx >= sidebarWidth {
		// New Shortcut: CTRL or CMD + Click to add point at mouse position
		if ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta) {
			worldX := float64(mx) - offsetX
			worldY := float64(my) - offsetY
			cx, cy := engine.IsoToCartesian(worldX, worldY)
			newP := game.FootprintPoint{
				X: math.Round(cx*100) / 100,
				Y: math.Round(cy*100) / 100,
			}
			*ee.Footprint = append(*ee.Footprint, newP)
			v.saveToYAML(ee)
			return nil
		}

		if mpos.In(v.addBtnRect) {
			v.addPoint(ee)
			return nil
		}
		if v.hoverIdx != -1 {
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				v.removePoint(ee, v.hoverIdx)
				return nil
			}
			v.draggingIdx = v.hoverIdx
		}
	}

	if v.draggingIdx != -1 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			worldX := float64(mx) - offsetX
			worldY := float64(my) - offsetY
			cx, cy := engine.IsoToCartesian(worldX, worldY)
			(*ee.Footprint)[v.draggingIdx].X = math.Round(cx*100) / 100
			(*ee.Footprint)[v.draggingIdx].Y = math.Round(cy*100) / 100
		} else {
			v.draggingIdx = -1
			v.saveToYAML(ee)
		}
	}

	return nil
}

func (v *Viewer) addPoint(ee *EditorEntity) {
	newP := game.FootprintPoint{}
	if len(*ee.Footprint) > 0 {
		last := (*ee.Footprint)[len(*ee.Footprint)-1]
		newP.X = last.X + 0.5
		newP.Y = last.Y + 0.5
	}
	*ee.Footprint = append(*ee.Footprint, newP)
	v.saveToYAML(ee)
}

func (v *Viewer) removePoint(ee *EditorEntity, idx int) {
	fp := *ee.Footprint
	if len(fp) <= 3 {
		log.Println("Cannot remove: polygon must have at least 3 vertices.")
		return
	}
	*ee.Footprint = append(fp[:idx], fp[idx+1:]...)
	v.saveToYAML(ee)
}

func (v *Viewer) saveToYAML(ee *EditorEntity) {
	if ee.YamlPath == "" {
		return
	}
	data, err := os.ReadFile(ee.YamlPath)
	if err != nil {
		log.Printf("failed to read yaml: %v", err)
		return
	}
	var m yaml.Node
	if err := yaml.Unmarshal(data, &m); err != nil {
		log.Printf("failed to unmarshal yaml: %v", err)
		return
	}
	fpData, _ := yaml.Marshal(*ee.Footprint)
	var fpNode yaml.Node
	if err := yaml.Unmarshal(fpData, &fpNode); err != nil {
		log.Printf("failed to unmarshal footprint: %v", err)
		return
	}
	if m.Content[0].Kind == yaml.MappingNode {
		found := false
		for i := 0; i < len(m.Content[0].Content); i += 2 {
			if m.Content[0].Content[i].Value == "footprint" {
				m.Content[0].Content[i+1] = fpNode.Content[0]
				found = true
				break
			}
		}
		if !found {
			m.Content[0].Content = append(m.Content[0].Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "footprint"},
				fpNode.Content[0],
			)
		}
	}
	f, err := os.Create(ee.YamlPath)
	if err != nil {
		log.Printf("failed to write yaml: %v", err)
		return
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(&m); err != nil {
		log.Printf("failed to encode yaml: %v", err)
		return
	}
	log.Println("Footprint saved to", ee.YamlPath)
}

func (v *Viewer) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	eImg := engine.NewEbitenImageWrapper(screen)

	baseX, baseY := sidebarWidth+float64(v.width-sidebarWidth)/2, float64(v.height)*0.6
	offsetX := baseX - v.camX
	offsetY := baseY - v.camY

	if v.selectedIndex >= 0 && v.selectedIndex < len(v.entities) {
		ee := v.entities[v.selectedIndex]
		ee.DrawMain(eImg, v.graphics, offsetX, offsetY)

		// Render footprint polygon
		poly := ee.GetFootprint()
		isoPoints := make([]engine.Point, len(poly.Points))
		for i, p := range poly.Points {
			ix, iy := engine.CartesianToIso(p.X, p.Y)
			isoPoints[i] = engine.Point{X: ix + offsetX, Y: iy + offsetY}
		}
		v.graphics.DrawPolygon(eImg, isoPoints, footprintColor, polyLineWidth)

		// Render vertices
		for i, p := range isoPoints {
			c := vertexColor
			r := float32(vertexRadius)
			if i == v.hoverIdx || i == v.draggingIdx {
				c = hoverColor
				r *= 1.5
			}
			v.graphics.DrawFilledCircle(eImg, float32(p.X), float32(p.Y), r, c, true)
			cp := poly.Points[i]
			v.graphics.DebugPrintAt(eImg, fmt.Sprintf("(%.2f, %.2f)", cp.X, cp.Y), int(p.X)+5, int(p.Y)+5)
		}

		v.drawUI(eImg, ee)
	}

	v.drawSidebar(screen)
}

func (v *Viewer) drawSidebar(screen *ebiten.Image) {
	eImg := engine.NewEbitenImageWrapper(screen)
	v.graphics.DrawFilledRect(eImg, 0, 0, sidebarWidth, float32(v.height), sidebarColor, false)

	for i, ee := range v.entities {
		y := v.scrollOffset + i*slotHeight
		if y+slotHeight < 0 || y > v.height {
			continue
		}

		// Selection border
		if i == v.selectedIndex {
			v.graphics.DrawFilledRect(eImg, 2, float32(y+2), sidebarWidth-4, slotHeight-4, selectedColor, false)
			v.graphics.DrawFilledRect(eImg, 5, float32(y+5), sidebarWidth-10, slotHeight-10, sidebarColor, false)
		}

		// Thumbnail sprite
		if ee.Image != nil {
			sw, sh := ee.Image.Size()
			scale := float64(thumbnailSize) / math.Max(float64(sw), float64(sh))
			op := engine.NewDrawImageOptions()
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(10+float64(thumbnailSize-float64(sw)*scale)/2, float64(y)+10+float64(thumbnailSize-float64(sh)*scale)/2)
			eImg.DrawImage(ee.Image, op)
		}

		// Label
		v.graphics.DebugPrintAt(eImg, ee.ID, thumbnailSize+20, y+25)
		v.graphics.DebugPrintAt(eImg, ee.Type, thumbnailSize+20, y+45)
	}
}

func (v *Viewer) drawUI(screen engine.Image, ee *EditorEntity) {
	title := fmt.Sprintf("[%s] %s", ee.Type, ee.ID)
	v.graphics.DebugPrintAt(screen, title, sidebarWidth+10, 10)
	v.graphics.DebugPrintAt(screen, fmt.Sprintf("Camera: (%.1f, %.1f)", v.camX, v.camY), sidebarWidth+10, 25)

	mx, my := ebiten.CursorPosition()
	v.graphics.DebugPrintAt(screen, fmt.Sprintf("Mouse: (%d, %d)", mx, my), sidebarWidth+10, 40)
	if v.hoverIdx != -1 {
		v.graphics.DebugPrintAt(screen, fmt.Sprintf("Hover: Vertex %d", v.hoverIdx), sidebarWidth+110, 40)
	}

	v.graphics.DrawFilledRect(screen, float32(v.addBtnRect.Min.X), float32(v.addBtnRect.Min.Y), float32(v.addBtnRect.Dx()), float32(v.addBtnRect.Dy()), buttonColor, false)
	v.graphics.DebugPrintAt(screen, " ADD POINT ", v.addBtnRect.Min.X+5, v.addBtnRect.Min.Y+8)
	v.graphics.DebugPrintAt(screen, "Drag (Move) | Shift+Click (Remove) | CTRL/CMD+Click (Add at Mouse) | ADD POINT button | Arrows (Cam) | Wheel (Sidebar) | ESC (Exit)", sidebarWidth+10, v.height-20)
}

func (v *Viewer) Layout(_, _ int) (int, int) {
	return v.width, v.height
}

func main() {
	assets := os.DirFS(".")
	graphics := engine.NewEbitenGraphics()

	var entities []*EditorEntity

	// 1. Obstacles
	obsReg := game.NewObstacleRegistry()
	if err := obsReg.LoadAll(assets); err == nil {
		obsReg.LoadAssets(assets, graphics)
		for _, id := range obsReg.IDs {
			arch := obsReg.Archetypes[id]
			obs := game.NewObstacle(0, 0, arch)
			var img engine.Image
			if arch.Image != nil {
				img = arch.Image.(engine.Image)
			}
			entities = append(entities, &EditorEntity{
				ID:        id,
				Type:      "Obstacle",
				Image:     img,
				Footprint: &arch.Footprint,
				YamlPath:  filepath.Join("data/obstacles", id+".yaml"),
				DrawMain: func(screen engine.Image, g engine.Graphics, offsetX, offsetY float64) {
					obs.Draw(screen, g, offsetX, offsetY)
				},
			})
		}
	}

	// 2. NPC Archetypes
	npcReg := game.NewArchetypeRegistry()
	if err := npcReg.LoadAll(assets); err == nil {
		npcReg.LoadAssets(assets, graphics)
		for _, id := range npcReg.IDs {
			arch := npcReg.Archetypes[id]
			npc := game.NewNPC(0, 0, arch, 1)
			var img engine.Image
			if arch.StaticImage != nil {
				img = arch.StaticImage.(engine.Image)
			}
			entities = append(entities, &EditorEntity{
				ID:        id,
				Type:      "NPC",
				Image:     img,
				Footprint: &arch.Footprint,
				YamlPath:  findArchetypeYAML(id),
				DrawMain: func(screen engine.Image, g engine.Graphics, offsetX, offsetY float64) {
					npc.Draw(screen, nil, g, offsetX, offsetY)
				},
			})
		}
	}

	// 3. Main Character
	mcConfig, err := game.LoadMainCharacterConfig(assets)
	if err == nil {
		// Override AssetDir for tool to find images
		mcConfig.AssetDir = "assets/images/characters/main"
		mcConfig.StaticImage = graphics.LoadSprite(assets, filepath.Join(mcConfig.AssetDir, "static.png"), true)
		mcConfig.CorpseImage = graphics.LoadSprite(assets, filepath.Join(mcConfig.AssetDir, "corpse.png"), true)
		mcConfig.AttackImage = graphics.LoadSprite(assets, filepath.Join(mcConfig.AssetDir, "attack.png"), true)

		mc := game.NewMainCharacter(0, 0, mcConfig)
		var img engine.Image
		if mcConfig.StaticImage != nil {
			img = mcConfig.StaticImage.(engine.Image)
		}
		entities = append(entities, &EditorEntity{
			ID:        "main_character",
			Type:      "Character",
			Image:     img,
			Footprint: &mcConfig.Footprint,
			YamlPath:  "data/characters/main/character.yaml",
			DrawMain: func(screen engine.Image, g engine.Graphics, offsetX, offsetY float64) {
				mc.Draw(screen, offsetX, offsetY)
			},
		})
	}

	// Sort entities by ID for convenience
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].Type != entities[j].Type {
			return entities[i].Type < entities[j].Type
		}
		return entities[i].ID < entities[j].ID
	})

	viewer := NewViewer(entities, graphics, defaultScreenWidth, defaultScreenHeight)

	ebiten.SetWindowTitle("Oinakos Boundary Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(viewer); err != nil {
		log.Fatal(err)
	}
}

// findArchetypeYAML recursively searches data/archetypes for the given ID.
func findArchetypeYAML(id string) string {
	baseDir := "data/archetypes"
	var found string
	filepath.WalkDir(baseDir, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(fpath) == ".yaml" || filepath.Ext(fpath) == ".yml" {
			data, err := os.ReadFile(fpath)
			if err == nil && containsID(data, id) {
				found = fpath
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

func containsID(data []byte, id string) bool {
	var m struct {
		ID string `yaml:"id"`
	}
	yaml.Unmarshal(data, &m)
	return m.ID == id
}
