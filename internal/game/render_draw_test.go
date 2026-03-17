package game

import (
	"image/color"
	"testing"

	"oinakos/internal/engine"
)

func TestMainCharacterDraw(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
	graphics := &engine.MockGraphics{}
	screen := graphics.NewImage(100, 100)

	sprite := graphics.NewImage(32, 32)
	mc.Config = &EntityConfig{}
	mc.Config.StaticImage = sprite
	mc.Config.AttackImage = sprite
	mc.Config.CorpseImage = sprite

	// Test drawing in various states
	states := []ActorState{ActorIdle, ActorWalking, ActorAttacking, ActorDead}
	for _, s := range states {
		mc.State = s
		mc.Facing = DirSE // Trigger flip
		mc.Tick = 10
		mc.Draw(screen, graphics, graphics, 0, 0)

		mc.Facing = DirNW // No flip
		mc.Draw(screen, graphics, graphics, 0, 0)
	}
}

func TestNPCDraw(t *testing.T) {
	n := NewNPC(0, 0, nil, 1)
	graphics := &engine.MockGraphics{}
	screen := graphics.NewImage(100, 100)
	sprite := graphics.NewImage(32, 32)

	// Manually set sprites since we are in headless test
	n.Archetype = &Archetype{}
	n.Archetype.StaticImage = sprite
	n.Archetype.CorpseImage = sprite
	n.Archetype.AttackImage = sprite

	states := []ActorState{ActorIdle, ActorWalking, ActorAttacking, ActorDead}
	for _, s := range states {
		n.State = s
		n.Facing = DirSE
		n.Draw(screen, graphics, graphics, nil, 0, 0)

		n.Facing = DirNW
		n.Draw(screen, graphics, graphics, nil, 0, 0)
	}
}

func TestObstacleDraw(t *testing.T) {
	graphics := &engine.MockGraphics{}
	screen := graphics.NewImage(100, 100)
	sprite := graphics.NewImage(32, 32)

	config := &ObstacleArchetype{
		ID:    "tree",
		Image: sprite,
	}
	o := NewObstacle("tree_instance", 0, 0, config)
	o.Draw(screen, graphics, 0, 0)

	o.Alive = false
	o.Draw(screen, graphics, 0, 0)

	// Test without image
	o.Archetype.Image = nil
	o.Draw(screen, graphics, 0, 0)
}

func TestProjectileDraw(t *testing.T) {
	graphics := &engine.MockGraphics{}
	screen := graphics.NewImage(100, 100)
	p := NewProjectile(0, 0, 1, 0, 1.0, 10, true, 100.0)
	p.Draw(screen, graphics, 0, 0)
}

func TestFloatingTextDraw(t *testing.T) {
	graphics := &engine.MockGraphics{}
	screen := graphics.NewImage(100, 100)
	ft := &FloatingText{
		Text:  "Hello",
		X:     0,
		Y:     0,
		Life:  30,
		Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
	}

	// Test full opacity
	ft.Draw(screen, graphics, 0, 0)

	// Test fade out branch
	ft.Life = 10
	ft.Draw(screen, graphics, 0, 0)
}
