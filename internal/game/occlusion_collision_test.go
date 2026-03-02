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
		ID:   "smithery",
		Type: "static",
		Footprint: []FootprintPoint{
			{X: -3.0, Y: -3.0}, {X: 3.0, Y: -3.0},
			{X: 3.0, Y: 3.0}, {X: -3.0, Y: 3.0},
		},
	}
	obsConfig.Image = graphics.NewImage(128, 128)

	building := NewObstacle("test_building", 0, 15, obsConfig)
	obstacles := []*Obstacle{building}

	// 1. Test Collision: Should collide with base
	collidesAtBase := mc.checkCollisionAt(0, 15, obstacles)
	if !collidesAtBase {
		t.Fatalf("Should have collided with building at Cartesian base (0, 15)")
	}

	// 2. Test "Relaxed Behind": Should NOT collide far behind (North) anymore
	collidesBehind := mc.checkCollisionAt(0, 5, obstacles) // 10 units North of center
	if collidesBehind {
		t.Errorf("Relaxed Behind failed: character should be allowed to go behind (North) the building base")
	}
}
