package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"oinakos/internal/engine"
	"oinakos/internal/game"
	"gopkg.in/yaml.v3"
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

func (m *MapEditor) loadLibrary() {
	root := findAssetsRoot()
	assets := os.DirFS(root)

	obsReg := game.NewObstacleRegistry()
	if err := obsReg.LoadAll(assets); err == nil {
		obsReg.LoadAssets(assets, m.Graphics, nil, nil)
		for _, id := range obsReg.IDs {
			arch := obsReg.Archetypes[id]
			var img engine.Image
			if arch.Image != nil {
				img = arch.Image
			}
			m.Library = append(m.Library, &EditorItem{
				ID:        id,
				Type:      "obstacle",
				Image:     img,
				Archetype: arch,
			})
		}
	}

	npcReg := game.NewArchetypeRegistry()
	if err := npcReg.LoadAll(assets); err == nil {
		npcReg.LoadAssets(assets, m.Graphics, nil, nil)
		for _, id := range npcReg.IDs {
			arch := npcReg.Archetypes[id]
			var img engine.Image
			if arch.StaticImage != nil {
				img = arch.StaticImage
			}
			m.Library = append(m.Library, &EditorItem{
				ID:        id,
				Type:      "npc",
				Image:     img,
				Archetype: arch,
			})
		}
	}

	sort.Slice(m.Library, func(i, j int) bool {
		return m.Library[i].ID < m.Library[j].ID
	})
}

func (m *MapEditor) loadFloors() {
	root := findAssetsRoot()
	assets := os.DirFS(root)
	floorDirPath := filepath.Join(root, "assets", "images", "floors")
	files, err := os.ReadDir(floorDirPath)
	if err != nil {
		log.Printf("Failed to list floors in %s: %v", floorDirPath, err)
		return
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".png") {
			name := f.Name()
			m.Floors = append(m.Floors, name)
			// For fs.FS we must use path.Join to ensure forward slashes on all platforms
			tex := m.Graphics.LoadSprite(assets, "assets/images/floors/"+name, true)
			m.FloorImages[name] = tex
		}
	}
	sort.Strings(m.Floors)
	for i, f := range m.Floors {
		if f == "grass.png" {
			m.FloorIdx = i
			break
		}
	}
}

func (m *MapEditor) initializeMap() {
	if m.InName == "" { return }
	root := findAssetsRoot()
	mapsDir := filepath.Join(root, "data", "maps")
	os.MkdirAll(mapsDir, 0755)
	m.Filename = filepath.Join(mapsDir, m.InName+".yaml")

	width, height := 0, 0
	fmt.Sscanf(m.InWidth, "%d", &width)
	fmt.Sscanf(m.InHeight, "%d", &height)

	m.MapData = &game.SaveData{}
	m.MapData.Map.ID = m.InName
	m.MapData.Map.WidthPixels = width
	m.MapData.Map.HeightPixels = height
	if m.FloorIdx < len(m.Floors) {
		m.MapData.Map.FloorTile = m.Floors[m.FloorIdx]
	}

	m.MapData.Player = game.PlayerSaveData{
		X: 0, Y: 0,
		State: game.State{
			HealthPoints:    100,
			MaxHealthPoints: 100,
		},
		Level: 1, 
		PrimaryAttributes: game.PrimaryAttributes{
			Strength:  50,
			Dexterity: 50,
			Health:    50,
			Intellect: 50,
			Wisdom:    50,
		},
	}

	m.Mode = "EDITOR"
	m.saveMap()
}

func (m *MapEditor) saveMap() {
	if m.Filename == "" { return }
	f, err := os.Create(m.Filename)
	if err != nil {
		log.Printf("Save failed: %v", err)
		return
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	enc.Encode(m.MapData)
}

func (m *MapEditor) placeItem(mx, my int) {
	worldX := float64(mx) - (sidebarWidth + (screenWidth-2*sidebarWidth)/2) + m.CamX
	worldY := float64(my) - (screenHeight / 2) + m.CamY
	cx, cy := engine.IsoToCartesian(worldX, worldY)

	if m.PendingItem.Type == "obstacle" {
		m.MapData.Obstacles = append(m.MapData.Obstacles, game.ObstacleSaveData{
			Archetype: m.PendingItem.ID, X: &cx, Y: &cy,
		})
	} else {
		m.MapData.Characters = append(m.MapData.Characters, game.NPCSaveData{
			Archetype: m.PendingItem.ID, X: cx, Y: cy, Level: 1, Behavior: "wander",
		})
	}
	m.saveMap()
}

func (m *MapEditor) pickAt(mx, my int) int {
	worldX := float64(mx) - (sidebarWidth + (screenWidth-2*sidebarWidth)/2) + m.CamX
	worldY := float64(my) - (screenHeight / 2) + m.CamY
	cx, cy := engine.IsoToCartesian(worldX, worldY)

	for i, npc := range m.MapData.Characters {
		if math.Hypot(cx-npc.X, cy-npc.Y) < 1.0 { return i | (1 << 30) }
	}
	for i, obs := range m.MapData.Obstacles {
		if obs.X != nil && obs.Y != nil && math.Hypot(cx-*obs.X, cy-*obs.Y) < 1.5 { return i }
	}
	return -1
}

func (m *MapEditor) selectElement(val int) {
	if val == -1 { m.Selection = nil; return }
	isNPC := (val & (1 << 30)) != 0
	idx := val &^ (1 << 30)

	if isNPC {
		data := m.MapData.Characters[idx]
		m.Selection = &MapElement{
			ID: fmt.Sprintf("npc_%d", idx), X: data.X, Y: data.Y, Item: m.findItem(data.Archetype, "npc"),
		}
	} else {
		data := m.MapData.Obstacles[idx]
		m.Selection = &MapElement{
			ID: fmt.Sprintf("obs_%d", idx), X: *data.X, Y: *data.Y, Item: m.findItem(data.Archetype, "obstacle"),
		}
	}
}

func (m *MapEditor) findItem(id, itype string) *EditorItem {
	for _, it := range m.Library {
		if it.ID == id && it.Type == itype { return it }
	}
	return nil
}

func (m *MapEditor) removeSelection() {
	if m.Selection == nil { return }
	parts := strings.Split(m.Selection.ID, "_")
	var idx int
	fmt.Sscanf(parts[1], "%d", &idx)

	if strings.HasPrefix(m.Selection.ID, "npc_") {
		if idx < len(m.MapData.Characters) {
			m.MapData.Characters = append(m.MapData.Characters[:idx], m.MapData.Characters[idx+1:]...)
		}
	} else {
		if idx < len(m.MapData.Obstacles) {
			m.MapData.Obstacles = append(m.MapData.Obstacles[:idx], m.MapData.Obstacles[idx+1:]...)
		}
	}
	m.Selection = nil
	m.saveMap()
}

func (m *MapEditor) syncToSaveData() {
	if m.Selection == nil { return }
	parts := strings.Split(m.Selection.ID, "_")
	var idx int
	fmt.Sscanf(parts[1], "%d", &idx)

	if strings.HasPrefix(m.Selection.ID, "npc_") {
		if idx < len(m.MapData.Characters) {
			m.MapData.Characters[idx].X = m.Selection.X
			m.MapData.Characters[idx].Y = m.Selection.Y
		}
	} else {
		if idx < len(m.MapData.Obstacles) {
			*m.MapData.Obstacles[idx].X = m.Selection.X
			*m.MapData.Obstacles[idx].Y = m.Selection.Y
		}
	}
	m.saveMap()
}

func (m *MapEditor) handleElevationClick(mx, my int, increase bool) {
	worldX := float64(mx) - (sidebarWidth + (screenWidth-2*sidebarWidth)/2) + m.CamX
	worldY := float64(my) - (screenHeight / 2) + m.CamY
	cx, cy := engine.IsoToCartesian(worldX, worldY)
	
	gridX := int(math.Floor(cx))
	gridY := int(math.Floor(cy))
	key := fmt.Sprintf("%d,%d", gridX, gridY)

	if m.MapData.Map.Heightmap == nil {
		m.MapData.Map.Heightmap = make(map[string]float64)
	}

	if m.ElevationTool == "brush" {
		val := m.MapData.Map.Heightmap[key]
		if increase {
			val += 0.5
		} else {
			val -= 0.5
		}
		if val == 0 {
			delete(m.MapData.Map.Heightmap, key)
		} else {
			m.MapData.Map.Heightmap[key] = val
		}
	} else if m.ElevationTool == "flatten" || m.ElevationTool == "slope" {
		pt := engine.Point{X: float64(gridX), Y: float64(gridY)}
		if increase {
			if m.ElevationP1 == nil {
				m.ElevationP1 = &pt
			} else {
				p2 := pt
				if m.ElevationTool == "flatten" {
					p1Key := fmt.Sprintf("%d,%d", int(m.ElevationP1.X), int(m.ElevationP1.Y))
					targetZ := m.MapData.Map.Heightmap[p1Key]
					
					minX := int(math.Min(m.ElevationP1.X, p2.X))
					maxX := int(math.Max(m.ElevationP1.X, p2.X))
					minY := int(math.Min(m.ElevationP1.Y, p2.Y))
					maxY := int(math.Max(m.ElevationP1.Y, p2.Y))
					
					for y := minY; y <= maxY; y++ {
						for x := minX; x <= maxX; x++ {
							k := fmt.Sprintf("%d,%d", x, y)
							if targetZ == 0 {
								delete(m.MapData.Map.Heightmap, k)
							} else {
								m.MapData.Map.Heightmap[k] = targetZ
							}
						}
					}
				} else if m.ElevationTool == "slope" {
					p1Key := fmt.Sprintf("%d,%d", int(m.ElevationP1.X), int(m.ElevationP1.Y))
					z1 := m.MapData.Map.Heightmap[p1Key]
					p2Key := fmt.Sprintf("%d,%d", int(p2.X), int(p2.Y))
					z2 := m.MapData.Map.Heightmap[p2Key]
					
					steps := math.Max(math.Abs(p2.X-m.ElevationP1.X), math.Abs(p2.Y-m.ElevationP1.Y))
					if steps > 0 {
						dx := (p2.X - m.ElevationP1.X) / steps
						dy := (p2.Y - m.ElevationP1.Y) / steps
						dz := (z2 - z1) / steps
						
						for i := 0; i <= int(steps); i++ {
							currX := int(math.Round(m.ElevationP1.X + float64(i)*dx))
							currY := int(math.Round(m.ElevationP1.Y + float64(i)*dy))
							currZ := z1 + float64(i)*dz
							
							k := fmt.Sprintf("%d,%d", currX, currY)
							if currZ == 0 {
								delete(m.MapData.Map.Heightmap, k)
							} else {
								m.MapData.Map.Heightmap[k] = currZ
							}
						}
					}
				}
				m.ElevationP1 = nil
			}
		} else {
			m.ElevationP1 = nil
		}
	}
	m.saveMap()
}
