package game

import (
	"fmt"
	"image/color"
	"io/fs"
	"oinakos/internal/engine"
	"testing"
)

type mockGraphics struct {
	engine.Graphics
	drawCalls []string
}

func (m *mockGraphics) DebugPrintAt(screen engine.Image, str string, x, y int) {
	m.drawCalls = append(m.drawCalls, fmt.Sprintf("DebugPrint:%s at %d,%d", str, x, y))
}

func (m *mockGraphics) NewImage(w, h int) engine.Image {
	return &mockImage{w: w, h: h}
}

func (m *mockGraphics) LoadSprite(assets fs.FS, path string, removeBg bool) engine.Image {
	return &mockImage{w: 128, h: 128} // Fake building/character sprite
}

func (m *mockGraphics) DrawFilledRect(screen engine.Image, x, y, width, height float32, clr color.Color, antiAlias bool) {
}

type mockImage struct {
	engine.Image
	w, h int
}

func (m *mockImage) Size() (int, int) {
	return m.w, m.h
}

func (m *mockImage) DrawImage(img engine.Image, options *engine.DrawImageOptions) {
}

func TestOcclusionAndCollision(t *testing.T) {
	graphics := &mockGraphics{}
	mcConfig := &EntityConfig{ID: "player"}
	mcConfig.StaticImage = graphics.NewImage(32, 32)
	mcConfig.Weapon = WeaponTizon

	mc := NewMainCharacter(0, 14, mcConfig) // Move player "behind" the building

	obsConfig := &ObstacleArchetype{
		ID:              "smithery",
		Type:            "static",
		FootprintWidth:  6.0,
		FootprintHeight: 6.0,
	}
	obsConfig.Image = graphics.NewImage(128, 128)

	building := NewObstacle(0, 15, obsConfig)
	obstacles := []*Obstacle{building}

	// 1. Test Collision
	collides := mc.checkCollisionAt(0, 15, obstacles)
	if !collides {
		t.Errorf("Should have collided with building at Cartesian(0, 15)")
	}

	// 2. Test Occlusion Detection
	g := &Game{
		mainCharacter:  mc,
		obstacles:      obstacles,
		currentMapType: MapType{ID: "test"},
		camera:         engine.NewCamera(0, 0),
	}
	gr := &GameRenderer{
		game:     g,
		graphics: graphics,
	}

	// Check if character is correctly obscured using actual rendering logic
	isObscured := gr.isEntityObscured(mc.X, mc.Y)

	if !isObscured {
		t.Errorf("Occlusion failed: character should be obscured by the building based on isometric projections")
	}
}
