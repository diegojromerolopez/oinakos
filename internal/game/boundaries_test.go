package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestPlayableCharacterBoundaries(t *testing.T) {
	ctx := NewTestContext()
	mc := NewPlayableCharacter(0, 0, nil)
	mc.Speed = 1.0
	ctx.World.PlayableCharacter = mc

	// Map 10x10 -> halfW=5, halfH=5
	ctx.World.CurrentMapType = &MapType{MapWidth: 10.0, MapHeight: 10.0}

	mockInput := ctx.Input.(*MockInputManager)

	// 1. Try to move beyond positive X boundary
	mc.X, mc.Y = 4.5, 0
	mockInput.PressedKeys[engine.KeyD] = true // Move Right (+X)
	mc.Update(ctx)
	delete(mockInput.PressedKeys, engine.KeyD)

	if mc.X > 5.0 {
		t.Errorf("PlayableCharacter moved beyond +X boundary: got %f, want <= 5.0", mc.X)
	}

	// 2. Try to move beyond negative X boundary
	mc.X, mc.Y = -4.5, 0
	mockInput.PressedKeys[engine.KeyA] = true // Move Left (-X)
	mc.Update(ctx)
	delete(mockInput.PressedKeys, engine.KeyA)

	if mc.X < -5.0 {
		t.Errorf("PlayableCharacter moved beyond -X boundary: got %f, want >= -5.0", mc.X)
	}

	// 3. Try to move beyond positive Y boundary
	mc.X, mc.Y = 0, 4.5
	mockInput.PressedKeys[engine.KeyS] = true // Move Down (+Y)
	mc.Update(ctx)
	delete(mockInput.PressedKeys, engine.KeyS)

	if mc.Y > 5.0 {
		t.Errorf("PlayableCharacter moved beyond +Y boundary: got %f, want <= 5.0", mc.Y)
	}

	// 4. Try to move beyond negative Y boundary
	mc.X, mc.Y = 0, -4.5
	mockInput.PressedKeys[engine.KeyW] = true // Move Up (-Y)
	mc.Update(ctx)
	delete(mockInput.PressedKeys, engine.KeyW)

	if mc.Y < -5.0 {
		t.Errorf("PlayableCharacter moved beyond -Y boundary: got %f, want >= -5.0", mc.Y)
	}

	// 5. Teleport outside and check if it clamps anyway
	mc.X, mc.Y = 100, 100
	mc.Update(ctx)
	if mc.X > 5.0 || mc.Y > 5.0 {
		t.Errorf("Teleported PlayableCharacter still outside boundaries after Update: got (%f, %f), want (<=5.0, <=5.0)", mc.X, mc.Y)
	}
}

func TestNPCBoundaries(t *testing.T) {
	ctx := NewTestContext()
	n := NewNPC(0, 0, nil, 1)
	n.Speed = 1.0
	n.Behavior = BehaviorWander
	ctx.World.NPCs = []*NPC{n}

	// Map 10x10 -> halfW=5, halfH=5
	ctx.World.CurrentMapType = &MapType{MapWidth: 10.0, MapHeight: 10.0}

	// 1. NPC wandering beyond positive X
	n.X, n.Y = 4.5, 0
	n.WanderDirX = 1.0
	n.WanderDirY = 0.0
	// Force wandering logic
	n.Tick = 1 // Not a 120 multiple so it doesn't pick new dir
	n.Update(ctx)

	if n.X > 5.0 {
		t.Errorf("NPC moved beyond +X boundary: got %f, want <= 5.0", n.X)
	}

	// 2. NPC wandering beyond negative X
	n.X, n.Y = -4.5, 0
	n.WanderDirX = -1.0
	n.WanderDirY = 0.0
	n.Update(ctx)

	if n.X < -5.0 {
		t.Errorf("NPC moved beyond -X boundary: got %f, want >= -5.0", n.X)
	}

	// 3. NPC wandering beyond positive Y
	n.X, n.Y = 0, 4.5
	n.WanderDirX = 0.0
	n.WanderDirY = 1.0
	n.Update(ctx)

	if n.Y > 5.0 {
		t.Errorf("NPC moved beyond +Y boundary: got %f, want <= 5.0", n.Y)
	}

	// 4. NPC wandering beyond negative Y
	n.X, n.Y = 0, -4.5
	n.WanderDirX = 0.0
	n.WanderDirY = -1.0
	n.Update(ctx)

	if n.Y < -5.0 {
		t.Errorf("NPC moved beyond -Y boundary: got %f, want >= -5.0", n.Y)
	}

	// 5. Teleport outside and check if it clamps anyway
	n.X, n.Y = 100, 100
	n.Update(ctx)
	if n.X > 5.0 || n.Y > 5.0 {
		t.Errorf("Teleported NPC still outside boundaries after Update: got (%f, %f), want (<=5.0, <=5.0)", n.X, n.Y)
	}
}
