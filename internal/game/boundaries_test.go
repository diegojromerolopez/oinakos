package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestPlayableCharacterBoundaries(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
	mc.Speed = 1.0

	// Map 10x10 -> halfW=5, halfH=5
	mapW, mapH := 10.0, 10.0

	// 1. Try to move beyond positive X boundary
	mc.X, mc.Y = 4.5, 0
	input := &MockInput{PressedKeys: []engine.Key{engine.KeyD}} // Move Right (+X)
	mc.Update(input, nil, nil, nil, nil, nil, mapW, mapH, nil, nil)

	if mc.X > 5.0 {
		t.Errorf("PlayableCharacter moved beyond +X boundary: got %f, want <= 5.0", mc.X)
	}

	// 2. Try to move beyond negative X boundary
	mc.X, mc.Y = -4.5, 0
	input = &MockInput{PressedKeys: []engine.Key{engine.KeyA}} // Move Left (-X)
	mc.Update(input, nil, nil, nil, nil, nil, mapW, mapH, nil, nil)

	if mc.X < -5.0 {
		t.Errorf("PlayableCharacter moved beyond -X boundary: got %f, want >= -5.0", mc.X)
	}

	// 3. Try to move beyond positive Y boundary
	mc.X, mc.Y = 0, 4.5
	input = &MockInput{PressedKeys: []engine.Key{engine.KeyS}} // Move Down (+Y)
	mc.Update(input, nil, nil, nil, nil, nil, mapW, mapH, nil, nil)

	if mc.Y > 5.0 {
		t.Errorf("PlayableCharacter moved beyond +Y boundary: got %f, want <= 5.0", mc.Y)
	}

	// 4. Try to move beyond negative Y boundary
	mc.X, mc.Y = 0, -4.5
	input = &MockInput{PressedKeys: []engine.Key{engine.KeyW}} // Move Up (-Y)
	mc.Update(input, nil, nil, nil, nil, nil, mapW, mapH, nil, nil)

	if mc.Y < -5.0 {
		t.Errorf("PlayableCharacter moved beyond -Y boundary: got %f, want >= -5.0", mc.Y)
	}

	// 5. Teleport outside and check if it clamps anyway
	mc.X, mc.Y = 100, 100
	mc.Update(nil, nil, nil, nil, nil, nil, mapW, mapH, nil, nil)
	if mc.X > 5.0 || mc.Y > 5.0 {
		t.Errorf("Teleported PlayableCharacter still outside boundaries after Update: got (%f, %f), want (<=5.0, <=5.0)", mc.X, mc.Y)
	}
}

func TestNPCBoundaries(t *testing.T) {
	n := NewNPC(0, 0, nil, 1)
	n.Speed = 1.0
	n.Behavior = BehaviorWander

	// Map 10x10 -> halfW=5, halfH=5
	mapW, mapH := 10.0, 10.0

	// 1. NPC wandering beyond positive X
	n.X, n.Y = 4.5, 0
	n.WanderDirX = 1.0
	n.WanderDirY = 0.0
	// Force wandering logic
	n.Tick = 1 // Not a 120 multiple so it doesn't pick new dir
	n.Update(nil, nil, nil, []*NPC{n}, nil, nil, mapW, mapH, nil, nil, nil)

	if n.X > 5.0 {
		t.Errorf("NPC moved beyond +X boundary: got %f, want <= 5.0", n.X)
	}

	// 2. NPC wandering beyond negative X
	n.X, n.Y = -4.5, 0
	n.WanderDirX = -1.0
	n.WanderDirY = 0.0
	n.Update(nil, nil, nil, []*NPC{n}, nil, nil, mapW, mapH, nil, nil, nil)

	if n.X < -5.0 {
		t.Errorf("NPC moved beyond -X boundary: got %f, want >= -5.0", n.X)
	}

	// 3. NPC wandering beyond positive Y
	n.X, n.Y = 0, 4.5
	n.WanderDirX = 0.0
	n.WanderDirY = 1.0
	n.Update(nil, nil, nil, []*NPC{n}, nil, nil, mapW, mapH, nil, nil, nil)

	if n.Y > 5.0 {
		t.Errorf("NPC moved beyond +Y boundary: got %f, want <= 5.0", n.Y)
	}

	// 4. NPC wandering beyond negative Y
	n.X, n.Y = 0, -4.5
	n.WanderDirX = 0.0
	n.WanderDirY = -1.0
	n.Update(nil, nil, nil, []*NPC{n}, nil, nil, mapW, mapH, nil, nil, nil)

	if n.Y < -5.0 {
		t.Errorf("NPC moved beyond -Y boundary: got %f, want >= -5.0", n.Y)
	}

	// 5. Teleport outside and check if it clamps anyway
	n.X, n.Y = 100, 100
	n.Update(nil, nil, nil, []*NPC{n}, nil, nil, mapW, mapH, nil, nil, nil)
	if n.X > 5.0 || n.Y > 5.0 {
		t.Errorf("Teleported NPC still outside boundaries after Update: got (%f, %f), want (<=5.0, <=5.0)", n.X, n.Y)
	}
}

// MockInput for testing
type MockInput struct {
	engine.Input
	PressedKeys []engine.Key
}

func (m *MockInput) IsKeyPressed(k engine.Key) bool {
	for _, pk := range m.PressedKeys {
		if pk == k {
			return true
		}
	}
	return false
}

func (m *MockInput) IsMouseButtonPressed(button engine.MouseButton) bool {
	return false
}

func (m *MockInput) IsMouseButtonJustPressed(button engine.MouseButton) bool {
	return false
}

func (m *MockInput) Wheel() (x, y float64) {
	return 0, 0
}

func (m *MockInput) AppendInputChars(chars []rune) []rune {
	return chars
}

func (m *MockInput) MousePosition() (x, y int) {
	return 0, 0
}

func (m *MockInput) AppendJustPressedKeys(keys []engine.Key) []engine.Key {
	return keys
}

func (m *MockInput) IsKeyJustPressed(k engine.Key) bool {
	return false
}
