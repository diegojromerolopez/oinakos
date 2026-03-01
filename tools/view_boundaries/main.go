// Package main provides a standalone tool for visualizing entity boundaries
// (obstacles, NPCs, and the main character) and their collision footprints
// in isometric space. It supports interactive vertex editing for obstacles.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"

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
)

var (
	backgroundColor = color.RGBA{30, 30, 30, 255}
	footprintColor  = color.RGBA{255, 0, 0, 255}
	vertexColor     = color.RGBA{255, 255, 0, 255}
	hoverColor      = color.RGBA{0, 255, 255, 255}
	buttonColor     = color.RGBA{60, 60, 60, 255}
)

// ─── Entity abstraction ───────────────────────────────────────────────────────

// Entity is a common interface for anything that can be drawn and has a footprint.
type Entity interface {
	GetFootprint() engine.Polygon
	GetLabel() string
}

// obstacleEntity wraps a game.Obstacle and supports vertex editing.
type obstacleEntity struct {
	obs      *game.Obstacle
	yamlPath string
}

func (e *obstacleEntity) GetFootprint() engine.Polygon { return e.obs.GetFootprint() }
func (e *obstacleEntity) GetLabel() string {
	return fmt.Sprintf("Obstacle: %s (%s)", e.obs.Archetype.ID, e.obs.Archetype.Type)
}

// npcEntity wraps a game.NPC.
type npcEntity struct {
	npc *game.NPC
}

func (e *npcEntity) GetFootprint() engine.Polygon { return e.npc.GetFootprint() }
func (e *npcEntity) GetLabel() string {
	return fmt.Sprintf("NPC: %s (%s)", e.npc.Archetype.ID, e.npc.Archetype.Name)
}

// characterEntity wraps a game.MainCharacter.
type characterEntity struct {
	mc *game.MainCharacter
}

func (e *characterEntity) GetFootprint() engine.Polygon { return e.mc.GetFootprint() }
func (e *characterEntity) GetLabel() string {
	return fmt.Sprintf("Character: %s", e.mc.Config.Name)
}

// ─── Viewer ───────────────────────────────────────────────────────────────────

// Viewer implements ebiten.Game for an interactive entity boundary visualizer.
type Viewer struct {
	entity      Entity
	graphics    engine.Graphics
	width       int
	height      int
	camX        float64
	camY        float64
	draggingIdx int
	hoverIdx    int
	addBtnRect  image.Rectangle
}

// NewViewer creates a new Viewer for the given entity.
func NewViewer(e Entity, g engine.Graphics, w, h int) *Viewer {
	return &Viewer{
		entity:      e,
		graphics:    g,
		width:       w,
		height:      h,
		draggingIdx: -1,
		hoverIdx:    -1,
		addBtnRect:  image.Rect(10, 60, 110, 90),
	}
}

// Update handles input and camera movement. Vertex editing is only enabled for
// obstacles.
func (v *Viewer) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	// Camera pan
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

	// Vertex editing is only available for obstacles.
	oe, isObstacle := v.entity.(*obstacleEntity)
	if !isObstacle {
		return nil
	}

	mx, my := ebiten.CursorPosition()
	mpos := image.Point{mx, my}
	baseX, baseY := float64(v.width)/2, float64(v.height)*0.6
	offsetX := baseX - v.camX
	offsetY := baseY - v.camY

	// Auto-initialize empty footprints from the current fallback so we can edit them.
	if len(oe.obs.Archetype.Footprint) == 0 {
		poly := oe.obs.GetFootprint()
		for _, p := range poly.Points {
			oe.obs.Archetype.Footprint = append(oe.obs.Archetype.Footprint, game.FootprintPoint{X: p.X, Y: p.Y})
		}
	}

	v.hoverIdx = -1
	for i, p := range oe.obs.Archetype.Footprint {
		ix, iy := engine.CartesianToIso(p.X, p.Y)
		dist := math.Hypot(float64(mx)-(ix+offsetX), float64(my)-(iy+offsetY))
		if dist < clickThreshold {
			v.hoverIdx = i
			break
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if mpos.In(v.addBtnRect) {
			v.addPoint(oe)
			return nil
		}
		if v.hoverIdx != -1 {
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				v.removePoint(oe, v.hoverIdx)
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
			oe.obs.Archetype.Footprint[v.draggingIdx].X = math.Round(cx*100) / 100
			oe.obs.Archetype.Footprint[v.draggingIdx].Y = math.Round(cy*100) / 100
		} else {
			v.draggingIdx = -1
			v.saveToYAML(oe)
		}
	}

	return nil
}

func (v *Viewer) addPoint(oe *obstacleEntity) {
	newP := game.FootprintPoint{}
	if len(oe.obs.Archetype.Footprint) > 0 {
		last := oe.obs.Archetype.Footprint[len(oe.obs.Archetype.Footprint)-1]
		newP.X = last.X + 0.5
		newP.Y = last.Y + 0.5
	}
	oe.obs.Archetype.Footprint = append(oe.obs.Archetype.Footprint, newP)
	v.saveToYAML(oe)
}

func (v *Viewer) removePoint(oe *obstacleEntity, idx int) {
	fp := oe.obs.Archetype.Footprint
	if len(fp) <= 3 {
		log.Println("Cannot remove: polygon must have at least 3 vertices.")
		return
	}
	oe.obs.Archetype.Footprint = append(fp[:idx], fp[idx+1:]...)
	v.saveToYAML(oe)
}

func (v *Viewer) saveToYAML(oe *obstacleEntity) {
	data, err := os.ReadFile(oe.yamlPath)
	if err != nil {
		log.Printf("failed to read yaml: %v", err)
		return
	}
	var m yaml.Node
	if err := yaml.Unmarshal(data, &m); err != nil {
		log.Printf("failed to unmarshal yaml: %v", err)
		return
	}
	fpData, _ := yaml.Marshal(oe.obs.Archetype.Footprint)
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
	f, err := os.Create(oe.yamlPath)
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
	log.Println("Footprint saved.")
}

// Draw renders the entity sprite (if any), its footprint, and the UI.
func (v *Viewer) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	eImg := engine.NewEbitenImageWrapper(screen)

	baseX := float64(v.width) / 2
	baseY := float64(v.height) * 0.6
	offsetX := baseX - v.camX
	offsetY := baseY - v.camY

	// Render entity sprite depending on type.
	switch e := v.entity.(type) {
	case *obstacleEntity:
		e.obs.Draw(eImg, v.graphics, offsetX, offsetY)
	case *npcEntity:
		e.npc.Draw(eImg, nil, v.graphics, offsetX, offsetY)
	case *characterEntity:
		e.mc.Draw(eImg, offsetX, offsetY)
	}

	// Render footprint polygon.
	poly := v.entity.GetFootprint()
	isoPoints := make([]engine.Point, len(poly.Points))
	for i, p := range poly.Points {
		ix, iy := engine.CartesianToIso(p.X, p.Y)
		isoPoints[i] = engine.Point{X: ix + offsetX, Y: iy + offsetY}
	}
	v.graphics.DrawPolygon(eImg, isoPoints, footprintColor, polyLineWidth)

	// Render vertices + coordinate labels.
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

	v.drawUI(eImg)
}

func (v *Viewer) drawUI(screen engine.Image) {
	v.graphics.DebugPrintAt(screen, v.entity.GetLabel(), 10, 10)
	v.graphics.DebugPrintAt(screen, fmt.Sprintf("Camera: (%.1f, %.1f)", v.camX, v.camY), 10, 25)

	// Editing controls are only shown for obstacles.
	if _, isObstacle := v.entity.(*obstacleEntity); isObstacle {
		v.graphics.DrawFilledRect(screen, float32(v.addBtnRect.Min.X), float32(v.addBtnRect.Min.Y), float32(v.addBtnRect.Dx()), float32(v.addBtnRect.Dy()), buttonColor, false)
		v.graphics.DebugPrintAt(screen, " ADD POINT ", v.addBtnRect.Min.X+5, v.addBtnRect.Min.Y+8)
		v.graphics.DebugPrintAt(screen, "Drag (Move) | Shift+Click (Remove) | ADD POINT button | Arrows (Cam) | ESC (Exit)", 10, v.height-20)
	} else {
		v.graphics.DebugPrintAt(screen, "Arrows (Pan Camera) | ESC (Exit)", 10, v.height-20)
	}
}

// Layout returns the logical screen size.
func (v *Viewer) Layout(_, _ int) (int, int) {
	return v.width, v.height
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	obstacleID := flag.String("obstacle", "", "ID of the obstacle to view (e.g. warehouse)")
	npcID := flag.String("npc", "", "ID of the NPC archetype to view (e.g. orc_male)")
	characterName := flag.String("character", "", "Load the main character (value is ignored, any string works)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s --obstacle <id> | --npc <id> | --character <name>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	count := 0
	if *obstacleID != "" {
		count++
	}
	if *npcID != "" {
		count++
	}
	if *characterName != "" {
		count++
	}
	if count != 1 {
		flag.Usage()
		os.Exit(1)
	}

	assets := os.DirFS(".")
	graphics := engine.NewEbitenGraphics()

	var entity Entity
	var title string

	switch {
	case *obstacleID != "":
		reg := game.NewObstacleRegistry()
		if err := reg.LoadAll(assets); err != nil {
			log.Fatalf("failed to load obstacles: %v", err)
		}
		arch, ok := reg.Archetypes[*obstacleID]
		if !ok {
			log.Fatalf("obstacle '%s' not found", *obstacleID)
		}
		reg.LoadAssets(assets, graphics)
		obs := game.NewObstacle(0, 0, arch)
		entity = &obstacleEntity{
			obs:      obs,
			yamlPath: filepath.Join("data/obstacles", *obstacleID+".yaml"),
		}
		title = fmt.Sprintf("Boundaries: obstacle/%s", *obstacleID)

	case *npcID != "":
		reg := game.NewArchetypeRegistry()
		if err := reg.LoadAll(assets); err != nil {
			log.Fatalf("failed to load archetypes: %v", err)
		}
		arch, ok := reg.Archetypes[*npcID]
		if !ok {
			log.Fatalf("npc archetype '%s' not found. Available: %v", *npcID, reg.IDs)
		}
		reg.LoadAssets(assets, graphics)
		npc := game.NewNPC(0, 0, arch, 1)
		entity = &npcEntity{npc: npc}
		title = fmt.Sprintf("Boundaries: npc/%s", *npcID)

	case *characterName != "":
		config, err := game.LoadMainCharacterConfig(assets)
		if err != nil {
			log.Printf("Warning: %v. Using defaults.", err)
		}
		// Load character assets
		if config.AssetDir != "" {
			config.StaticImage = graphics.LoadSprite(assets, filepath.Join(config.AssetDir, "static.png"), true)
		}
		mc := game.NewMainCharacter(0, 0, config)
		entity = &characterEntity{mc: mc}
		title = "Boundaries: character/main"
	}

	viewer := NewViewer(entity, graphics, defaultScreenWidth, defaultScreenHeight)

	ebiten.SetWindowTitle(title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(viewer); err != nil {
		log.Fatal(err)
	}
}
