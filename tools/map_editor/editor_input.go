package main

import (
	"oinakos/internal/engine"
	"github.com/hajimehoshi/ebiten/v2"
)

func (m *MapEditor) updateDialog() error {
	for _, ch := range m.Input.AppendInputChars(nil) {
		switch m.ActiveField {
		case 0: m.InName += string(ch)
		case 1: if ch >= '0' && ch <= '9' { m.InWidth += string(ch) }
		case 2: if ch >= '0' && ch <= '9' { m.InHeight += string(ch) }
		}
	}

	if m.Input.IsKeyJustPressed(engine.KeyBackspace) {
		switch m.ActiveField {
		case 0: if len(m.InName) > 0 { m.InName = m.InName[:len(m.InName)-1] }
		case 1: if len(m.InWidth) > 0 { m.InWidth = m.InWidth[:len(m.InWidth)-1] }
		case 2: if len(m.InHeight) > 0 { m.InHeight = m.InHeight[:len(m.InHeight)-1] }
		}
	}

	if m.Input.IsKeyJustPressed(engine.KeyTab) { m.ActiveField = (m.ActiveField + 1) % 3 }
	if m.Input.IsKeyJustPressed(engine.KeyEnter) { m.initializeMap() }
	if m.Input.IsKeyJustPressed(engine.KeyEscape) { return ebiten.Termination }
	return nil
}

func (m *MapEditor) updateEditor() error {
	mx, my := m.Input.MousePosition()

	if mx < sidebarWidth {
		_, wy := m.Input.Wheel()
		m.ScrollL += int(wy * 30)
		if m.ScrollL > 0 { m.ScrollL = 0 }
		if minL := -(len(m.Library) - 1) * slotHeight; m.ScrollL < minL { m.ScrollL = minL }
		if m.Input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
			idx := (my - m.ScrollL) / slotHeight
			if idx >= 0 && idx < len(m.Library) { m.PendingItem = m.Library[idx] }
		}
	}

	if mx > screenWidth-sidebarWidth {
		_, wy := m.Input.Wheel()
		m.ScrollR += int(wy * 30)
		if m.ScrollR > 0 { m.ScrollR = 0 }
		if minR := -(len(m.Floors) - 1) * slotHeight; m.ScrollR < minR { m.ScrollR = minR }
		if m.Input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
			idx := (my - m.ScrollR) / slotHeight
			if idx >= 0 && idx < len(m.Floors) {
				m.FloorIdx = idx
				m.MapData.Map.FloorTile = m.Floors[idx]
				m.saveMap()
			}
		}
	}

	if mx >= sidebarWidth && mx <= screenWidth-sidebarWidth {
		if m.Input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
			found := m.pickAt(mx, my)
			if found != -1 { m.selectElement(found); m.PendingItem = nil } else {
				if m.PendingItem != nil { m.placeItem(mx, my) } else { m.deselect() }
			}
		}

		if m.Selection != nil {
			moved := false
			if m.Input.IsKeyJustPressed(engine.KeyUp) { m.Selection.Y -= 1.0; moved = true }
			if m.Input.IsKeyJustPressed(engine.KeyDown) { m.Selection.Y += 1.0; moved = true }
			if m.Input.IsKeyJustPressed(engine.KeyLeft) { m.Selection.X -= 1.0; moved = true }
			if m.Input.IsKeyJustPressed(engine.KeyRight) { m.Selection.X += 1.0; moved = true }
			if m.Input.IsKeyJustPressed(engine.KeyDelete) || m.Input.IsKeyJustPressed(engine.KeyBackspace) { m.removeSelection() } else if moved { m.syncToSaveData() }
		} else {
			if m.Input.IsKeyPressed(engine.KeyUp) { m.CamY -= 5 }
			if m.Input.IsKeyPressed(engine.KeyDown) { m.CamY += 5 }
			if m.Input.IsKeyPressed(engine.KeyLeft) { m.CamX -= 5 }
			if m.Input.IsKeyPressed(engine.KeyRight) { m.CamX += 5 }
		}
	}

	if m.Input.IsKeyJustPressed(engine.KeyEscape) { return ebiten.Termination }
	return nil
}

func (m *MapEditor) deselect() {
	m.Selection = nil
}
