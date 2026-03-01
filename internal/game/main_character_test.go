package game

import (
	"oinakos/internal/engine"
	"testing"
	"testing/fstest"
)

func TestMainCharacterStats(t *testing.T) {
	mc := &MainCharacter{
		BaseAttack:  10,
		BaseDefense: 5,
		Level:       1,
	}

	if att := mc.GetTotalAttack(); att != 10 {
		t.Errorf("GetTotalAttack(Level 1): got %d, want 10", att)
	}

	mc.Level = 10
	if att := mc.GetTotalAttack(); att != 43 {
		t.Errorf("GetTotalAttack(Level 10): got %d, want 43", att)
	}
}

func TestMainCharacterXPAndLevelUp(t *testing.T) {
	mc := &MainCharacter{
		Level: 1,
		XP:    0,
	}

	mc.AddXP(50)
	if mc.XP != 50 {
		t.Errorf("XP after 50: got %d, want 50", mc.XP)
	}
	if mc.Level != 1 {
		t.Errorf("Level after 50 XP: got %d, want 1", mc.Level)
	}

	mc.AddXP(100)
	if mc.Level != 2 {
		t.Errorf("Level after 150 XP: got %d, want 2", mc.Level)
	}
	if mc.XP != 150 {
		t.Errorf("XP after Level Up: got %d, want 150", mc.XP)
	}
}

func TestMainCharacterTakeDamage(t *testing.T) {
	mc := &MainCharacter{Health: 100, MaxHealth: 100}
	mc.TakeDamage(20, nil)
	if mc.Health != 80 {
		t.Errorf("Health after damage: got %d, want 80", mc.Health)
	}
	mc.TakeDamage(100, nil)
	if mc.Health != 0 {
		t.Errorf("Health after lethal damage: got %d, want 0", mc.Health)
	}
	if mc.IsAlive() {
		t.Error("Character should be dead")
	}
}

func TestMainCharacterGetters(t *testing.T) {
	mc := &MainCharacter{BaseDefense: 10}
	if mc.GetTotalDefense() != 10 {
		t.Errorf("GetTotalDefense: got %d, want 10", mc.GetTotalDefense())
	}
	if mc.GetTotalProtection() != 0 {
		t.Errorf("GetTotalProtection: got %d, want 0", mc.GetTotalProtection())
	}

	mc.EquippedArmor = map[ArmorSlot]*Armor{SlotBody: {Protection: 5}}
	if mc.GetTotalProtection() != 5 {
		t.Errorf("GetTotalProtection with armor: got %d, want 5", mc.GetTotalProtection())
	}
}

func TestMainCharacterCheckAttackHits(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	mc.Weapon = &Weapon{MinDamage: 10, MaxDamage: 10}
	mc.Facing = DirSE

	npc := &NPC{X: 1, Y: 0.5, State: NPCIdle}
	npcs := []*NPC{npc}
	fts := &[]*FloatingText{}
	mc.CheckAttackHits(npcs, nil, fts, nil)
}

func TestMainCharacterFootprint(t *testing.T) {
	mc := NewMainCharacter(10, 10, nil)
	fp := mc.GetFootprint()
	if len(fp.Points) == 0 {
		t.Error("Footprint should have points")
	}
}

func TestMainCharacterCollision(t *testing.T) {
	mc := NewMainCharacter(10, 10, nil)
	colliders := []*Obstacle{NewObstacle(10.5, 10.5, nil)}
	if !mc.checkCollisionAt(10.5, 10.5, colliders) {
		t.Error("Expected collision at 10.5, 10.5")
	}
	if mc.checkCollisionAt(20, 20, colliders) {
		t.Error("Expected no collision at 20, 20")
	}
}

func TestLoadPlayerImage(t *testing.T) {
	// Need a valid 1x1 PNG data string
	pngData := "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xb4\x00\x00\x00\x00IEND\xaeB`\x82"
	mockFS := fstest.MapFS{
		"test.png": {Data: []byte(pngData)},
	}

	_, err := loadPlayerImage(mockFS, "test.png")
	if err != nil {
		t.Errorf("Expected nil error loading from fs: %v", err)
	}

	// Test missing file
	_, err = loadPlayerImage(mockFS, "missing.png")
	if err == nil {
		t.Errorf("Expected error loading missing file")
	}
}

func TestMainCharacterUpdate_Full(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	mc.Health = mc.MaxHealth
	mockInput := NewMockInputManager()
	fts := &[]*FloatingText{}

	// Update when dead
	mc.State = StateDead
	mc.Update(mockInput, nil, nil, nil, fts, 100, 100)
	if mc.State != StateDead {
		t.Error("Dead mc should stay dead")
	}

	// Update drinking
	mc.State = StateDrinking
	mc.Tick = 0
	mc.Update(mockInput, nil, nil, nil, fts, 100, 100)
	if mc.State != StateDrinking {
		t.Error("Should stay drinking")
	}
	mc.Tick = 60
	mc.Update(mockInput, nil, nil, nil, fts, 100, 100)
	if mc.State != StateIdle {
		t.Error("Should be idle after drink timer")
	}

	// Update attacking
	mc.State = StateAttacking
	mc.Tick = 14
	mc.Update(mockInput, nil, nil, nil, fts, 100, 100)
	if mc.Tick != 15 {
		t.Error("Tick should advance")
	}
	mc.Tick = 30
	mc.Update(mockInput, nil, nil, nil, fts, 100, 100)
	if mc.State != StateIdle {
		t.Error("Should be idle after attack anim")
	}

	// Movement Input checks
	mc.State = StateIdle
	mockInput.PressedKeys[engine.KeyW] = true
	mockInput.PressedKeys[engine.KeyD] = true
	mc.Update(mockInput, nil, nil, nil, fts, 100, 100)
	if mc.State != StateWalking {
		t.Error("Should be walking on input")
	}
	if mc.Facing != DirNE {
		t.Errorf("Expected Facing DirNE, got %v", mc.Facing)
	}

	mc.X = 0
	mc.Y = 0
	delete(mockInput.PressedKeys, engine.KeyW)
	mockInput.PressedKeys[engine.KeyS] = true
	mc.Update(mockInput, nil, nil, nil, fts, 100, 100)
	if mc.Facing != DirSE {
		t.Errorf("Expected Facing DirSE, got %v", mc.Facing)
	}

	// Test clamp boundaries
	mc.X = 1000
	mc.Y = 1000
	mockInput.PressedKeys[engine.KeyD] = true // Move right edge
	mc.Update(mockInput, nil, nil, nil, fts, 100, 100)
	if mc.X > 50 || mc.Y > 50 {
		t.Error("Position not clamped correctly")
	}
}
