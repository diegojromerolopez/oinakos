package main

import (
	"fmt"
	"image/color"
	"math"
	"oinakos/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
)

func (v *Viewer) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	eImg := engine.NewEbitenImageWrapper(screen)

	baseX, baseY := sidebarWidth+float64(v.width-sidebarWidth)/2, float64(v.height)*0.6
	offsetX := baseX - v.camX
	offsetY := baseY - v.camY

	if v.selectedIndex >= 0 && v.selectedIndex < len(v.entities) {
		ee := v.entities[v.selectedIndex]
		ee.DrawMain(eImg, v.graphics, offsetX, offsetY)

		poly := ee.GetFootprint()
		isoPoints := make([]engine.Point, len(poly.Points))
		for i, p := range poly.Points {
			ix, iy := engine.CartesianToIso(p.X, p.Y)
			isoPoints[i] = engine.Point{X: ix + offsetX, Y: iy + offsetY}
		}
		v.graphics.DrawPolygon(eImg, isoPoints, footprintColor, polyLineWidth)

		for i, p := range isoPoints {
			c := vertexColor
			r := float32(vertexRadius)
			if i == v.hoverIdx || i == v.draggingIdx {
				c = hoverColor
				r *= 1.5
			}
			v.graphics.DrawFilledCircle(eImg, float32(p.X), float32(p.Y), r, c, true)
			cp := poly.Points[i]
			v.graphics.DebugPrintAt(eImg, fmt.Sprintf("(%.2f, %.2f)", cp.X, cp.Y), int(p.X)+5, int(p.Y)+5, color.White)
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
		if y+slotHeight < 0 || y > v.height { continue }
		if i == v.selectedIndex {
			v.graphics.DrawFilledRect(eImg, 2, float32(y+2), sidebarWidth-4, slotHeight-4, selectedColor, false)
			v.graphics.DrawFilledRect(eImg, 5, float32(y+5), sidebarWidth-10, slotHeight-10, sidebarColor, false)
		}
		if ee.Image != nil {
			sw, sh := ee.Image.Size()
			scale := float64(thumbnailSize) / math.Max(float64(sw), float64(sh))
			op := engine.NewDrawImageOptions()
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(10+float64(thumbnailSize-float64(sw)*scale)/2, float64(y)+10+float64(thumbnailSize-float64(sh)*scale)/2)
			eImg.DrawImage(ee.Image, op)
		}
		v.graphics.DebugPrintAt(eImg, ee.ID, thumbnailSize+20, y+25, color.White)
		v.graphics.DebugPrintAt(eImg, ee.Type, thumbnailSize+20, y+45, color.White)
	}
}

func (v *Viewer) drawUI(screen engine.Image, ee *EditorEntity) {
	title := fmt.Sprintf("[%s] %s", ee.Type, ee.ID)
	v.graphics.DebugPrintAt(screen, title, sidebarWidth+10, 10, color.White)
	v.graphics.DebugPrintAt(screen, fmt.Sprintf("Camera: (%.1f, %.1f)", v.camX, v.camY), sidebarWidth+10, 25, color.White)

	mx, my := v.input.MousePosition()
	v.graphics.DebugPrintAt(screen, fmt.Sprintf("Mouse: (%d, %d)", mx, my), sidebarWidth+10, 40, color.White)
	if v.hoverIdx != -1 {
		v.graphics.DebugPrintAt(screen, fmt.Sprintf("Hover: Vertex %d", v.hoverIdx), sidebarWidth+110, 40, color.White)
	}

	v.graphics.DrawFilledRect(screen, float32(v.addBtnRect.Min.X), float32(v.addBtnRect.Min.Y), float32(v.addBtnRect.Dx()), float32(v.addBtnRect.Dy()), buttonColor, false)
	v.graphics.DebugPrintAt(screen, " ADD POINT ", v.addBtnRect.Min.X+5, v.addBtnRect.Min.Y+8, color.White)
	v.graphics.DebugPrintAt(screen, "Drag (Move) | Shift+Click (Remove) | CTRL/CMD+Click (Add at Mouse) | ADD POINT button | Arrows (Cam) | Wheel (Sidebar) | ESC (Exit)", sidebarWidth+10, v.height-20, color.White)
}
