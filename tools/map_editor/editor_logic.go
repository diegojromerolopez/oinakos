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

func (m *MapEditor) loadLibrary() {
	assets := os.DirFS(".")

	obsReg := game.NewObstacleRegistry()
	if err := obsReg.LoadAll(assets); err == nil {
		obsReg.LoadAssets(assets, m.Graphics, nil)
		for _, id := range obsReg.IDs {
			arch := obsReg.Archetypes[id]
			var img engine.Image
			if arch.Image != nil {
				img = arch.Image.(engine.Image)
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
		npcReg.LoadAssets(assets, m.Graphics, nil)
		for _, id := range npcReg.IDs {
			arch := npcReg.Archetypes[id]
			var img engine.Image
			if arch.StaticImage != nil {
				img = arch.StaticImage.(engine.Image)
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
	assets := os.DirFS(".")
	files, err := os.ReadDir("assets/images/floors")
	if err != nil {
		log.Printf("Failed to list floors: %v", err)
		return
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".png") {
			name := f.Name()
			m.Floors = append(m.Floors, name)
			tex := m.Graphics.LoadSprite(assets, filepath.Join("assets/images/floors", name), true)
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
	const mapsDir = "oinakos/data/maps"
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
		X: 0, Y: 0, Health: 100, MaxHealth: 100,
		Level: 1, BaseAttack: 10, BaseDefense: 5,
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
			ArchetypeID: m.PendingItem.ID, X: &cx, Y: &cy,
		})
	} else {
		m.MapData.NPCs = append(m.MapData.NPCs, game.NPCSaveData{
			ArchetypeID: m.PendingItem.ID, X: cx, Y: cy, Level: 1, Behavior: "wander",
		})
	}
	m.saveMap()
}

func (m *MapEditor) pickAt(mx, my int) int {
	worldX := float64(mx) - (sidebarWidth + (screenWidth-2*sidebarWidth)/2) + m.CamX
	worldY := float64(my) - (screenHeight / 2) + m.CamY
	cx, cy := engine.IsoToCartesian(worldX, worldY)

	for i, npc := range m.MapData.NPCs {
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
		data := m.MapData.NPCs[idx]
		m.Selection = &MapElement{
			ID: fmt.Sprintf("npc_%d", idx), X: data.X, Y: data.Y, Item: m.findItem(data.ArchetypeID, "npc"),
		}
	} else {
		data := m.MapData.Obstacles[idx]
		m.Selection = &MapElement{
			ID: fmt.Sprintf("obs_%d", idx), X: *data.X, Y: *data.Y, Item: m.findItem(data.ArchetypeID, "obstacle"),
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
		if idx < len(m.MapData.NPCs) {
			m.MapData.NPCs = append(m.MapData.NPCs[:idx], m.MapData.NPCs[idx+1:]...)
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
		if idx < len(m.MapData.NPCs) {
			m.MapData.NPCs[idx].X = m.Selection.X
			m.MapData.NPCs[idx].Y = m.Selection.Y
		}
	} else {
		if idx < len(m.MapData.Obstacles) {
			*m.MapData.Obstacles[idx].X = m.Selection.X
			*m.MapData.Obstacles[idx].Y = m.Selection.Y
		}
	}
	m.saveMap()
}
