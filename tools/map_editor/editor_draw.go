package main

import (
	"fmt"
	"image/color"
	"math"
	"oinakos/internal/engine"
	"oinakos/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func (m *MapEditor) drawDialog(screen *ebiten.Image) {
	screen.Fill(colorBG)
	eImg := engine.NewEbitenImageWrapper(screen)

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
	if active { boxColor = colorSelect }
	m.Graphics.DrawFilledRect(eImg, x+120, y-5, 200, 25, boxColor, false)
	m.Graphics.DrawFilledRect(eImg, x+122, y-3, 196, 21, colorBG, false)
	m.Graphics.DebugPrintAt(eImg, val, int(x)+125, int(y), colorText)
}

func (m *MapEditor) drawOutlineRect(eImg engine.Image, x, y, w, h float32) {
	pts := []engine.Point{
		{X: float64(x), Y: float64(y)},
		{X: float64(x + w), Y: float64(y)},
		{X: float64(x + w), Y: float64(y + h)},
		{X: float64(x), Y: float64(y + h)},
	}
	m.Graphics.DrawPolygon(eImg, pts, colorSelect, 1.0)
}

func (m *MapEditor) drawEditor(screen *ebiten.Image) {
	screen.Fill(colorBG)
	eImg := engine.NewEbitenImageWrapper(screen)

	offsetX := sidebarWidth + float64(screenWidth-2*sidebarWidth)/2 - m.CamX
	offsetY := float64(screenHeight)/2 - m.CamY

	if m.FloorIdx < len(m.Floors) {
		floorImg := m.FloorImages[m.Floors[m.FloorIdx]]
		if floorImg != nil {
			m.Renderer.DrawTileMap(eImg, offsetX, offsetY, func(x, y int) engine.Image {
				return floorImg
			})
		}
	}

	for i, obsData := range m.MapData.Obstacles {
		if obsData.X == nil || obsData.Y == nil { continue }
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
				m.drawOutlineRect(eImg, float32(ix+offsetX-20), float32(iy+offsetY-20), 40, 40)
			}
		}
	}

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
				m.drawOutlineRect(eImg, float32(ix+offsetX-20), float32(iy+offsetY-20), 40, 40)
			}
		}
	}

	m.Graphics.DrawFilledRect(eImg, 0, 0, sidebarWidth, screenHeight, colorSide, false)
	m.Graphics.DrawFilledRect(eImg, screenWidth-sidebarWidth, 0, sidebarWidth, screenHeight, colorSide, false)

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
		if m.PendingItem == item { m.drawOutlineRect(eImg, 2, float32(y+2), sidebarWidth-4, slotHeight-4) }
	}

	for i, f := range m.Floors {
		y := m.ScrollR + i*slotHeight
		img := m.FloorImages[f]
		if img != nil {
			op := engine.NewDrawImageOptions()
			op.GeoM.Scale(0.1, 0.1)
			op.GeoM.Translate(float64(screenWidth-sidebarWidth+10), float64(y+10))
			eImg.DrawImage(img, op)
		}
		m.Graphics.DebugPrintAt(eImg, f, screenWidth-sidebarWidth+80, y+30, colorText)
		if m.FloorIdx == i {
			m.Graphics.DrawFilledRect(eImg, float32(screenWidth-sidebarWidth+2), float32(y+2), sidebarWidth-4, slotHeight-4, color.RGBA{60, 80, 60, 120}, false)
		}
	}

	m.Graphics.DebugPrintAt(eImg, "MAP EDITOR - "+m.InName, sidebarWidth+20, 20, colorText)
	if m.PendingItem != nil {
		m.Graphics.DebugPrintAt(eImg, "PLACING: "+m.PendingItem.ID, screenWidth/2-50, 50, colorSelect)
	}
}
