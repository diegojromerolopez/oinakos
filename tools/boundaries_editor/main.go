package main

import (
	"flag"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"oinakos/internal/engine"
	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func findAssetsRoot() string {
	dir, _ := os.Getwd()
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "assets")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "oinakos", "assets")); err == nil {
			return filepath.Join(dir, "oinakos")
		}
		dir = filepath.Dir(dir)
	}
	return "."
}

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
		autoBtnRect:   image.Rect(sidebarWidth+120, 60, sidebarWidth+240, 90),
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
			if !v.InsertPointOnSegment(ee, mx, my, offsetX, offsetY) {
				cx, cy := engine.IsoToCartesian(float64(mx)-offsetX, float64(my)-offsetY)
				*ee.Footprint = append(*ee.Footprint, game.FootprintPoint{X: math.Round(cx*100)/100, Y: math.Round(cy*100)/100})
				v.saveToYAML(ee)
			}
			return nil
		}
		if image.Pt(mx, my).In(v.addBtnRect) {
			v.addPoint(ee)
			return nil
		}
		if image.Pt(mx, my).In(v.autoBtnRect) {
			v.AutoPerimeter(ee)
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
	root := findAssetsRoot()
	localAssets := os.DirFS(root)
	graphics := engine.NewEbitenGraphics()
	var entities []*EditorEntity

	// Load Obstacles
	obsReg := game.NewObstacleRegistry()
	if err := obsReg.LoadAll(localAssets); err == nil {
		obsReg.LoadAssets(localAssets, graphics, nil, nil)
		for _, id := range obsReg.IDs {
			arch := obsReg.Archetypes[id]
			obs := game.NewObstacle("editor_preview", 0, 0, arch)
			var img engine.Image
			if arch.Image != nil {
				img = arch.Image
			}
			entities = append(entities, &EditorEntity{
				ID: id, Type: "Obstacle", Image: img, Footprint: &arch.Footprint,
				YamlPath: filepath.Join(root, "data", "obstacles", id+".yaml"),
				DrawMain: func(screen engine.Image, g engine.Graphics, ox, oy float64) { obs.Draw(screen, g, ox, oy) },
			})
		}
	}

	// Load Objects
	objReg := game.NewObjectRegistry()
	if err := objReg.LoadAll(localAssets); err == nil {
		objReg.LoadAssets(localAssets, graphics, nil, nil)
		for _, id := range objReg.IDs {
			cfg := objReg.Objects[id]
			item := game.NewItemInstance(id, cfg, 0, 0)
			var img engine.Image
			if cfg.Sprite != nil {
				img = cfg.Sprite
			}
			entities = append(entities, &EditorEntity{
				ID: id, Type: "Object", Image: img, Footprint: &cfg.Footprint,
				YamlPath: filepath.Join(root, "data", "objects", id+".yaml"),
				DrawMain: func(screen engine.Image, _ engine.Graphics, ox, oy float64) {
					item.Draw(screen, ox, oy)
				},
			})
		}
	}

	sort.Slice(entities, func(i, j int) bool {
		if entities[i].Type != entities[j].Type {
			return entities[i].Type < entities[j].Type
		}
		return entities[i].ID < entities[j].ID
	})

	var targetID string
	var targetType string
	flag.StringVar(&targetID, "obstacle", "", "ID of the obstacle to select")
	flag.StringVar(&targetID, "object", "", "ID of the object to select")
	flag.Parse()

	if targetID == "" {
		// Try to find if any other flag was used (for backward compatibility or convenience)
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "obstacle" || f.Name == "object" {
				targetID = f.Value.String()
				targetType = strings.Title(f.Name)
			}
		})
	}

	selectedIndex := 0
	if targetID != "" {
		for i, e := range entities {
			if e.ID == targetID && (targetType == "" || e.Type == targetType) {
				selectedIndex = i
				break
			}
		}
	}

	viewer := NewViewer(entities, graphics, engine.NewEbitenInput(), defaultScreenWidth, defaultScreenHeight)
	viewer.selectedIndex = selectedIndex
	ebiten.SetWindowTitle("Oinakos Boundary Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(viewer); err != nil {
		log.Fatal(err)
	}
}
