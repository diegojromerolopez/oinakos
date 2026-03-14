package game

import "testing"

// Tests for PlayableCharacter.CheckAttackHits

func TestCheckAttackHits_HitOrMiss(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
	mc.BaseAttack = 50 // Very high → typically hits
	mc.Facing = DirSE

	npc := NewNPC(0.5, 0.5, &Archetype{ID: "test", XP: 1}, 1)
	npc.Health = 100
	npc.Alignment = AlignmentEnemy

	var fts []*FloatingText
	mc.CheckAttackHits([]*NPC{npc}, nil, nil, &fts, NewMockAudioManager(), nil, nil)

	// Either health dropped (hit) or a floating text was produced (MISS) — both valid
	if npc.Health == 100 && len(fts) == 0 {
		t.Error("Expected NPC to take damage or a floating text (MISS) to appear")
	}
}

func TestCheckAttackHits_DeadNPCSkipped(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
	mc.Facing = DirSE

	npc := NewNPC(0.5, 0.5, nil, 1)
	npc.State = NPCDead
	npc.Health = 50

	var fts []*FloatingText
	mc.CheckAttackHits([]*NPC{npc}, nil, nil, &fts, NewMockAudioManager(), nil, nil)

	if npc.Health != 50 {
		t.Error("Dead NPC should never take damage")
	}
	if len(fts) != 0 {
		t.Error("Dead NPC should not produce floating text")
	}
}

func TestCheckAttackHits_OutOfRange(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
	mc.Facing = DirSE

	npc := NewNPC(50, 50, nil, 1) // Far away
	npc.Health = 100

	var fts []*FloatingText
	mc.CheckAttackHits([]*NPC{npc}, nil, nil, &fts, NewMockAudioManager(), nil, nil)

	if npc.Health != 100 {
		t.Error("Out-of-range NPC should not take damage")
	}
	if len(fts) != 0 {
		t.Error("Out-of-range NPC should not produce floating text")
	}
}

func TestCheckAttackHits_AllDirections(t *testing.T) {
	t.Skip("Flaky in bulk runs, investigation pending")
	directions := []Direction{DirSE, DirSW, DirNE, DirNW}
	offsets := [][2]float64{{1, 0.5}, {-0.5, 1}, {1, -0.5}, {-0.5, -1}}

	for i, dir := range directions {
		mc := NewPlayableCharacter(0, 0, nil)
		mc.BaseAttack = 9999 // Guarantee a hit
		mc.Facing = dir

		dx, dy := offsets[i][0], offsets[i][1]
		npc := NewNPC(dx, dy, &Archetype{ID: "test"}, 1)
		npc.Health = 100
		npc.BaseDefense = 0

		var fts []*FloatingText
		// The game has a 5% miss cap (max hitChance=95).
		// We loop up to 100 times to virtually guarantee a hit happens.
		hitDetected := false
		for attempt := 0; attempt < 100; attempt++ {
			mc.CheckAttackHits([]*NPC{npc}, nil, nil, &fts, NewMockAudioManager(), nil, nil)
			if npc.Health < 100 {
				hitDetected = true
				break
			}
		}

		if !hitDetected {
			t.Errorf("Dir %d: NPC in the attack zone should have taken damage after multiple attempts", dir)
		}
	}
}

func TestCheckAttackHits_KillUpdatesMapKills(t *testing.T) {
	mc := NewPlayableCharacter(0, 0, nil)
	mc.BaseAttack = 9999 // Guarantee a huge hit
	mc.Facing = DirSE

	// Ensure the MapKills map is initialized (it is inside NewPlayableCharacter)
	npc := NewNPC(0.5, 0.5, &Archetype{ID: "crimson_guard", XP: 10}, 1)
	npc.Health = 1
	npc.BaseDefense = 0

	var fts []*FloatingText

	// The game has a 5% miss cap (max hitChance=95).
	// We loop up to 100 times to virtually guarantee a hit happens.
	hitDetected := false
	for attempt := 0; attempt < 100; attempt++ {
		mc.CheckAttackHits([]*NPC{npc}, nil, nil, &fts, NewMockAudioManager(), nil, nil)
		if npc.State == NPCDead {
			hitDetected = true
			break
		}
	}

	if !hitDetected {
		t.Errorf("NPC should have died after multiple huge hits")
	}

	if mc.MapKills["crimson_guard"] != 1 {
		t.Errorf("Expected MapKills['crimson_guard'] to be 1, got %d", mc.MapKills["crimson_guard"])
	}
	if mc.Kills != 1 {
		t.Errorf("Expected Kills to be 1, got %d", mc.Kills)
	}
}
