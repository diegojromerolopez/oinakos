package engine

import (
	"testing"
)

type CapturingMockImage struct {
	MockImage
	LastVertices []Vertex
	LastIndices []uint16
}

func (m *CapturingMockImage) DrawTriangles(vertices []Vertex, indices []uint16, src Image, options *DrawTrianglesOptions) {
	m.LastVertices = append(m.LastVertices, vertices...)
	m.LastIndices = append(m.LastIndices, indices...)
}

func TestRenderer_DrawTileMap(t *testing.T) {
	r := NewRenderer()
	
	// Create a screen (800x600)
	screen := &CapturingMockImage{
		MockImage: MockImage{W: 800, H: 600},
	}
	
	// Mock tile sprite
	tileSprite := &MockImage{W: 64, H: 32}
	
	getTile := func(x, y int) Image {
		return tileSprite
	}
	
	getZ := func(x, y int) float64 {
		return 0.0
	}
	
	// Draw at (0, 0) offset
	r.DrawTileMap(screen, 0, 0, getTile, getZ)
	
	if len(screen.LastVertices) == 0 {
		t.Errorf("Expected Vertices to be drawn, got 0")
	}
	
	// Test nil tile
	screenCount := len(screen.LastVertices)
	getNilTile := func(x, y int) Image { return nil }
	r.DrawTileMap(screen, 0, 0, getNilTile, getZ)
	if len(screen.LastVertices) != screenCount {
		t.Errorf("Expected no change in vertices count when getTile returns nil")
	}
	
	r.DrawTileMap(screen, 0, 0, nil, nil)
}

func TestRenderer_ShadingByElevation(t *testing.T) {
	r := NewRenderer()
	screen := &CapturingMockImage{
		MockImage: MockImage{W: 100, H: 100},
	}
	tileSprite := &MockImage{W: 10, H: 10}
	
	// High elevation (should be bright)
	r.DrawTileMap(screen, 0, 0, 
		func(x, y int) Image { if x == 0 && y == 0 { return tileSprite } ; return nil },
		func(x, y int) float64 { return 10.0 }) // 1.0 + 10.0*0.1 = 2.0 (clamped to 1.2)
	
	highShade := screen.LastVertices[0].ColorR
	if highShade != 1.2 {
		t.Errorf("Expected high elevation shade clamped to 1.2, got %v", highShade)
	}
	
	// Low elevation (should be dark)
	screen.LastVertices = nil
	r.DrawTileMap(screen, 0, 0, 
		func(x, y int) Image { if x == 0 && y == 0 { return tileSprite } ; return nil },
		func(x, y int) float64 { return -10.0 }) // 1.0 - 1.0 = 0.0 (clamped to 0.3)
		
	lowShade := screen.LastVertices[0].ColorR
	if lowShade != 0.3 {
		t.Errorf("Expected low elevation shade clamped to 0.3, got %v", lowShade)
	}
}
