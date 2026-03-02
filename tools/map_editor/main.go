package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"oinakos/internal/engine"
	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"gopkg.in/yaml.v3"
)

// ─── Constants & UI Settings ──────────────────────────────────────────────────

const (
	screenWidth  = 1280
	screenHeight = 720
	sidebarWidth = 220
	slotHeight   = 80
	thumbSize    = 60
)

var (
	colorBG      = color.RGBA{15, 15, 15, 255}
	colorSide    = color.RGBA{35, 35, 35, 255}
	colorText    = color.White
	colorSelect  = color.RGBA{0, 255, 0, 255}
	colorOutline = color.RGBA{50, 50, 50, 255}
	colorModal   = color.RGBA{25, 25, 25, 240}
)

// ─── Entity Definitions ────────────────────────────────────────────────────

type EditorItem struct {
	ID        string
	Type      string // "obstacle", "npc"
	Image     engine.Image
	Archetype interface{} // *game.ObstacleArchetype or *game.Archetype
}

type MapElement struct {
	Item     *EditorItem
	X, Y     float64 // Cartesian
	ID       string  // Instance ID
	Selected bool
}

// ─── Editor Application ──────────────────────────────────────────────────────

type MapEditor struct {
	// Persistence
	Filename string
	MapData  *game.SaveData

	// Assets
	Graphics    engine.Graphics
	Library     []*EditorItem
	Floors      []string
	FloorImages map[string]engine.Image

	// Selection
	PendingItem *EditorItem
	Selection   *MapElement
	FloorIdx    int

	// Camera & State
	CamX, CamY float64
	ScrollL    int
	ScrollR    int
	Mode       string // "DIALOG", "EDITOR"

	// Dialog Input
	InName      string
	InWidth     string
	InHeight    string
	ActiveField int // 0:Name, 1:W, 2:H
}

func NewMapEditor(g engine.Graphics) *MapEditor {
	me := &MapEditor{
		Graphics:    g,
		Mode:        "DIALOG",
		InWidth:     "640",
		InHeight:    "640",
		FloorImages: make(map[string]engine.Image),
	}
	me.loadLibrary()
	me.loadFloors()
	return me
}

func (m *MapEditor) loadLibrary() {
	assets := os.DirFS(".")

	// 1. Load Obstacles
	obsReg := game.NewObstacleRegistry()
	if err := obsReg.LoadAll(assets); err == nil {
		obsReg.LoadAssets(assets, m.Graphics)
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

	// 2. Load NPCs
	npcReg := game.NewArchetypeRegistry()
	if err := npcReg.LoadAll(assets); err == nil {
		npcReg.LoadAssets(assets, m.Graphics)
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

	// Sort alphabetically
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
			// Preload floor sprites
			tex := m.Graphics.LoadSprite(assets, filepath.Join("assets/images/floors", name), false)
			m.FloorImages[name] = tex
		}
	}
	sort.Strings(m.Floors)
	// Find grass index
	for i, f := range m.Floors {
		if f == "grass.png" {
			m.FloorIdx = i
			break
		}
	}
}

// ─── Ebiten Logic ───────────────────────────────────────────────────────────

func (m *MapEditor) Update() error {
	if m.Mode == "DIALOG" {
		return m.updateDialog()
	}
	return m.updateEditor()
}

func (m *MapEditor) updateDialog() error {
	for {
		chars := ebiten.AppendInputChars(nil)
		if len(chars) > 0 {
			switch m.ActiveField {
			case 0:
				m.InName += string(chars)
			case 1:
				if strings.Contains("0123456789", string(chars)) {
					m.InWidth += string(chars)
				}
			case 2:
				if strings.Contains("0123456789", string(chars)) {
					m.InHeight += string(chars)
				}
			}
		} else {
			break
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		switch m.ActiveField {
		case 0:
			if len(m.InName) > 0 {
				m.InName = m.InName[:len(m.InName)-1]
			}
		case 1:
			if len(m.InWidth) > 0 {
				m.InWidth = m.InWidth[:len(m.InWidth)-1]
			}
		case 2:
			if len(m.InHeight) > 0 {
				m.InHeight = m.InHeight[:len(m.InHeight)-1]
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		m.ActiveField = (m.ActiveField + 1) % 3
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		m.initializeMap()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	return nil
}

func (m *MapEditor) initializeMap() {
	if m.InName == "" {
		return
	}
	m.Filename = filepath.Join("data/maps", m.InName+".yaml")

	width := 0
	height := 0
	fmt.Sscanf(m.InWidth, "%d", &width)
	fmt.Sscanf(m.InHeight, "%d", &height)

	m.MapData = &game.SaveData{}
	m.MapData.Map.ID = m.InName
	m.MapData.Map.WidthPixels = width
	m.MapData.Map.HeightPixels = height
	if m.FloorIdx < len(m.Floors) {
		m.MapData.Map.FloorTile = m.Floors[m.FloorIdx]
	}

	// Create essential player start
	m.MapData.Player = game.PlayerSaveData{
		X: 0, Y: 0,
		Health: 100, MaxHealth: 100,
		Level: 1, BaseAttack: 10, BaseDefense: 5,
	}

	m.Mode = "EDITOR"
	m.saveMap()
}

func (m *MapEditor) saveMap() {
	if m.Filename == "" {
		return
	}
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

func (m *MapEditor) updateEditor() error {
	mx, my := ebiten.CursorPosition()

	// Scroll Left (Library)
	if mx < sidebarWidth {
		_, wy := ebiten.Wheel()
		m.ScrollL += int(wy * 30)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			idx := (my - m.ScrollL) / slotHeight
			if idx >= 0 && idx < len(m.Library) {
				m.PendingItem = m.Library[idx]
			}
		}
	}

	// Scroll Right (Floors)
	if mx > screenWidth-sidebarWidth {
		_, wy := ebiten.Wheel()
		m.ScrollR += int(wy * 30)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			idx := (my - m.ScrollR) / slotHeight
			if idx >= 0 && idx < len(m.Floors) {
				m.FloorIdx = idx
				m.MapData.Map.FloorTile = m.Floors[idx]
				m.saveMap()
			}
		}
	}

	// Main Editor View Interaction
	if mx >= sidebarWidth && mx <= screenWidth-sidebarWidth {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			// Click to select or place
			found := m.pickAt(mx, my)
			if found != -1 {
				m.selectElement(found)
				m.PendingItem = nil // Clear cursor item if we selected something on map
			} else {
				if m.PendingItem != nil {
					m.placeItem(mx, my)
				} else {
					m.deselect()
				}
			}
		}

		// Move selected item
		if m.Selection != nil {
			moved := false
			if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
				m.Selection.Y -= 1.0
				moved = true
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
				m.Selection.Y += 1.0
				moved = true
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				m.Selection.X -= 1.0
				moved = true
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				m.Selection.X += 1.0
				moved = true
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
				m.removeSelection()
			} else if moved {
				m.syncToSaveData()
			}
		} else {
			// Move camera if nothing selected
			if ebiten.IsKeyPressed(ebiten.KeyUp) {
				m.CamY -= 5
			}
			if ebiten.IsKeyPressed(ebiten.KeyDown) {
				m.CamY += 5
			}
			if ebiten.IsKeyPressed(ebiten.KeyLeft) {
				m.CamX -= 5
			}
			if ebiten.IsKeyPressed(ebiten.KeyRight) {
				m.CamX += 5
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	return nil
}

func (m *MapEditor) placeItem(mx, my int) {
	// Screen -> Iso -> Cartesian
	worldX := float64(mx) - (sidebarWidth + (screenWidth-2*sidebarWidth)/2) + m.CamX
	worldY := float64(my) - (screenHeight / 2) + m.CamY
	cx, cy := engine.IsoToCartesian(worldX, worldY)

	if m.PendingItem.Type == "obstacle" {
		m.MapData.Obstacles = append(m.MapData.Obstacles, game.ObstacleSaveData{
			ArchetypeID: m.PendingItem.ID,
			X:           &cx,
			Y:           &cy,
		})
	} else {
		m.MapData.NPCs = append(m.MapData.NPCs, game.NPCSaveData{
			ArchetypeID: m.PendingItem.ID,
			X:           cx,
			Y:           cy,
			Level:       1,
			Behavior:    "wander", // Default
		})
	}
	m.saveMap()
}

func (m *MapEditor) pickAt(mx, my int) int {
	worldX := float64(mx) - (sidebarWidth + (screenWidth-2*sidebarWidth)/2) + m.CamX
	worldY := float64(my) - (screenHeight / 2) + m.CamY
	cx, cy := engine.IsoToCartesian(worldX, worldY)

	// 1. Pick NPCs (Priority)
	for i, npc := range m.MapData.NPCs {
		dist := math.Hypot(cx-npc.X, cy-npc.Y)
		if dist < 1.0 {
			return i | (1 << 30) // Use high bit to signal NPC
		}
	}

	// 2. Pick Obstacles
	for i, obs := range m.MapData.Obstacles {
		if obs.X == nil || obs.Y == nil {
			continue
		}
		dist := math.Hypot(cx-*obs.X, cy-*obs.Y)
		if dist < 1.5 {
			return i
		}
	}
	return -1
}

func (m *MapEditor) selectElement(val int) {
	if val == -1 {
		m.Selection = nil
		return
	}
	isNPC := (val & (1 << 30)) != 0
	idx := val &^ (1 << 30)

	if isNPC {
		data := m.MapData.NPCs[idx]
		m.Selection = &MapElement{
			ID:   fmt.Sprintf("npc_%d", idx),
			X:    data.X,
			Y:    data.Y,
			Item: m.findItem(data.ArchetypeID, "npc"),
		}
	} else {
		data := m.MapData.Obstacles[idx]
		m.Selection = &MapElement{
			ID:   fmt.Sprintf("obs_%d", idx),
			X:    *data.X,
			Y:    *data.Y,
			Item: m.findItem(data.ArchetypeID, "obstacle"),
		}
	}
}

func (m *MapEditor) findItem(id, itype string) *EditorItem {
	for _, it := range m.Library {
		if it.ID == id && it.Type == itype {
			return it
		}
	}
	return nil
}

func (m *MapEditor) removeSelection() {
	if m.Selection == nil {
		return
	}

	isNPC := strings.HasPrefix(m.Selection.ID, "npc_")
	parts := strings.Split(m.Selection.ID, "_")
	var idx int
	fmt.Sscanf(parts[1], "%d", &idx)

	if isNPC {
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

func (m *MapEditor) deselect() {
	m.Selection = nil
}

func (m *MapEditor) syncToSaveData() {
	if m.Selection == nil {
		return
	}

	isNPC := strings.HasPrefix(m.Selection.ID, "npc_")
	parts := strings.Split(m.Selection.ID, "_")
	var idx int
	fmt.Sscanf(parts[1], "%d", &idx)

	if isNPC {
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

// ─── Drawing Logic ──────────────────────────────────────────────────────────

func (m *MapEditor) Draw(screen *ebiten.Image) {
	if m.Mode == "DIALOG" {
		m.drawDialog(screen)
		return
	}
	m.drawEditor(screen)
}

func (m *MapEditor) drawDialog(screen *ebiten.Image) {
	screen.Fill(colorBG)
	eImg := engine.NewEbitenImageWrapper(screen)

	// Draw Modal Box
	mvWidth, mvHeight := float32(400), float32(300)
	mvX, mvY := float32(screenWidth-400)/2, float32(screenHeight-300)/2
	m.Graphics.DrawFilledRect(eImg, mvX, mvY, mvWidth, mvHeight, colorModal, false)

	m.Graphics.DebugPrintAt(eImg, "--- NEW MAP CONFIG ---", int(mvX)+100, int(mvY)+20, colorText)

	m.drawField(eImg, "Map Name:", m.InName, mvX+40, mvY+70, m.ActiveField == 0)
	m.drawField(eImg, "Width (px):", m.InWidth, mvX+40, mvY+120, m.ActiveField == 1)
	m.drawField(eImg, "Height (px):", m.InHeight, mvX+40, mvY+170, m.ActiveField == 2)

	infiniteLabel := "Mode: FINITE (Set 0 for Infinite)"
	if m.InWidth == "0" || m.InWidth == "" {
		infiniteLabel = "Mode: INFINITE"
	}
	m.Graphics.DebugPrintAt(eImg, infiniteLabel, int(mvX)+60, int(mvY)+210, colorSelect)

	m.Graphics.DebugPrintAt(eImg, "[TAB] to switch fields | [ENTER] to OK", int(mvX)+60, int(mvY)+240, colorText)
	m.Graphics.DebugPrintAt(eImg, "ESC to Cancel and Exit", int(mvX)+110, int(mvY)+270, colorText)
}

func (m *MapEditor) drawField(eImg engine.Image, label, val string, x, y float32, active bool) {
	m.Graphics.DebugPrintAt(eImg, label, int(x), int(y), colorText)
	boxColor := colorOutline
	if active {
		boxColor = colorSelect
	}
	m.Graphics.DrawFilledRect(eImg, x+120, y-5, 200, 25, boxColor, false)
	m.Graphics.DrawFilledRect(eImg, x+122, y-3, 196, 21, colorBG, false)
	m.Graphics.DebugPrintAt(eImg, val, int(x)+125, int(y), colorText)
}

func (m *MapEditor) drawEditor(screen *ebiten.Image) {
	screen.Fill(colorBG)
	eImg := engine.NewEbitenImageWrapper(screen)

	// 1. Draw Map Viewport
	offsetX := sidebarWidth + float64(screenWidth-2*sidebarWidth)/2 - m.CamX
	offsetY := float64(screenHeight)/2 - m.CamY

	// Draw Floor Tile if selected
	if m.FloorIdx < len(m.Floors) {
		floorImg := m.FloorImages[m.Floors[m.FloorIdx]]
		if floorImg != nil {
			// Center it roughly
			op := engine.NewDrawImageOptions()
			sw, sh := floorImg.Size()
			op.GeoM.Translate(-float64(sw)/2, -float64(sh)/2)
			op.GeoM.Translate(offsetX, offsetY)
			eImg.DrawImage(floorImg, op)
		}
	}

	// Draw Obstacles from MapData
	for i, obsData := range m.MapData.Obstacles {
		if obsData.X == nil || obsData.Y == nil {
			continue
		}
		var arch *game.ObstacleArchetype
		for _, item := range m.Library {
			if item.ID == obsData.ArchetypeID {
				arch = item.Archetype.(*game.ObstacleArchetype)
				break
			}
		}
		if arch != nil {
			obs := game.NewObstacle("edit", *obsData.X, *obsData.Y, arch)
			obs.Draw(eImg, m.Graphics, offsetX, offsetY)
			if m.Selection != nil && m.Selection.ID == fmt.Sprintf("obs_%d", i) {
				ix, iy := engine.CartesianToIso(*obsData.X, *obsData.Y)
				m.Graphics.DrawFilledRect(eImg, float32(ix+offsetX-20), float32(iy+offsetY-20), 40, 40, colorSelect, true)
			}
		}
	}

	// Draw NPCs from MapData
	for i, npcData := range m.MapData.NPCs {
		var arch *game.Archetype
		for _, item := range m.Library {
			if item.ID == npcData.ArchetypeID {
				arch = item.Archetype.(*game.Archetype)
				break
			}
		}
		if arch != nil {
			npc := game.NewNPC(npcData.X, npcData.Y, arch, 1)
			npc.Draw(eImg, m.Graphics, m.Graphics, nil, offsetX, offsetY)
			if m.Selection != nil && m.Selection.ID == fmt.Sprintf("npc_%d", i) {
				ix, iy := engine.CartesianToIso(npcData.X, npcData.Y)
				m.Graphics.DrawFilledRect(eImg, float32(ix+offsetX-20), float32(iy+offsetY-20), 40, 40, colorSelect, true)
			}
		}
	}

	// 2. Sidebars
	m.Graphics.DrawFilledRect(eImg, 0, 0, sidebarWidth, screenHeight, colorSide, false)
	m.Graphics.DrawFilledRect(eImg, screenWidth-sidebarWidth, 0, sidebarWidth, screenHeight, colorSide, false)

	// Draw Left Sidebar (Library)
	for i, item := range m.Library {
		y := m.ScrollL + i*slotHeight
		if item.Image != nil {
			op := engine.NewDrawImageOptions()
			sw, sh := item.Image.Size()
			scale := float64(thumbSize) / math.Max(float64(sw), float64(sh))
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(10, float64(y+10))
			eImg.DrawImage(item.Image, op)
		}
		m.Graphics.DebugPrintAt(eImg, item.ID, 80, y+30, colorText)
		if m.PendingItem == item {
			m.Graphics.DrawFilledRect(eImg, 2, float32(y+2), sidebarWidth-4, slotHeight-4, colorSelect, true)
		}
	}

	// Draw Right Sidebar (Floors)
	for i, f := range m.Floors {
		y := m.ScrollR + i*slotHeight
		img := m.FloorImages[f]
		if img != nil {
			op := engine.NewDrawImageOptions()
			op.GeoM.Scale(0.1, 0.1) // Tiny preview
			op.GeoM.Translate(float64(screenWidth-sidebarWidth+10), float64(y+10))
			eImg.DrawImage(img, op)
		}
		m.Graphics.DebugPrintAt(eImg, f, screenWidth-sidebarWidth+80, y+30, colorText)
		if m.FloorIdx == i {
			m.Graphics.DrawFilledRect(eImg, float32(screenWidth-sidebarWidth+2), float32(y+2), sidebarWidth-4, slotHeight-4, colorSelect, true)
		}
	}

	// 3. UI Overlays
	m.Graphics.DebugPrintAt(eImg, "MAP EDITOR - "+m.InName, sidebarWidth+20, 20, colorText)
	if m.PendingItem != nil {
		m.Graphics.DebugPrintAt(eImg, "PLACING: "+m.PendingItem.ID, screenWidth/2-50, 50, colorSelect)
	}
}

func (m *MapEditor) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	graphics := engine.NewEbitenGraphics()
	editor := NewMapEditor(graphics)

	ebiten.SetWindowTitle("Oinakos Map Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(editor); err != nil {
		log.Fatal(err)
	}
}
