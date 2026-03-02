package game

import "testing"

// Tests for MainCharacter.CheckAttackHits

func TestCheckAttackHits_HitOrMiss(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	mc.BaseAttack = 50 // Very high → typically hits
	mc.Facing = DirSE

	npc := NewNPC(0.5, 0.5, &Archetype{ID: "test", XP: 1}, 1)
	npc.Health = 100
	npc.Alignment = AlignmentEnemy

	var fts []*FloatingText
	mc.CheckAttackHits([]*NPC{npc}, nil, &fts, NewMockAudioManager())

	// Either health dropped (hit) or a floating text was produced (MISS) — both valid
	if npc.Health == 100 && len(fts) == 0 {
		t.Error("Expected NPC to take damage or a floating text (MISS) to appear")
	}
}

func TestCheckAttackHits_DeadNPCSkipped(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	mc.Facing = DirSE

	npc := NewNPC(0.5, 0.5, nil, 1)
	npc.State = NPCDead
	npc.Health = 50

	var fts []*FloatingText
	mc.CheckAttackHits([]*NPC{npc}, nil, &fts, NewMockAudioManager())

	if npc.Health != 50 {
		t.Error("Dead NPC should never take damage")
	}
	if len(fts) != 0 {
		t.Error("Dead NPC should not produce floating text")
	}
}

func TestCheckAttackHits_OutOfRange(t *testing.T) {
	mc := NewMainCharacter(0, 0, nil)
	mc.Facing = DirSE

	npc := NewNPC(50, 50, nil, 1) // Far away
	npc.Health = 100

	var fts []*FloatingText
	mc.CheckAttackHits([]*NPC{npc}, nil, &fts, NewMockAudioManager())

	if npc.Health != 100 {
		t.Error("Out-of-range NPC should not take damage")
	}
	if len(fts) != 0 {
		t.Error("Out-of-range NPC should not produce floating text")
	}
}

func TestCheckAttackHits_AllDirections(t *testing.T) {
	directions := []Direction{DirSE, DirSW, DirNE, DirNW}
	offsets := [][2]float64{{1, 0.5}, {-0.5, 1}, {1, -0.5}, {-0.5, -1}}

	for i, dir := range directions {
		mc := NewMainCharacter(0, 0, nil)
		mc.BaseAttack = 9999 // Guarantee a hit
		mc.Facing = dir

		dx, dy := offsets[i][0], offsets[i][1]
		npc := NewNPC(dx, dy, &Archetype{ID: "test"}, 1)
		npc.Health = 100
		npc.BaseDefense = 0

		var fts []*FloatingText
		mc.CheckAttackHits([]*NPC{npc}, nil, &fts, NewMockAudioManager())

		if npc.Health == 100 {
			t.Errorf("Dir %d: NPC in the attack zone should have taken damage", dir)
		}
	}
}
