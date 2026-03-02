package game

import (
	"math"
	"math/rand"
	"oinakos/internal/engine"
	"testing"
)

func TestNPCCalculateStat(t *testing.T) {
	n := &NPC{}
	if res := n.calculateStat(10, 1); res != 10 {
		t.Errorf("calculateStat(10, 1): got %d, want 10", res)
	}
	if res := n.calculateStat(10, 10); res != 43 {
		t.Errorf("calculateStat(10, 10): got %d, want 43", res)
	}
}

func TestNPCGetters(t *testing.T) {
	n := &NPC{BaseAttack: 10, BaseDefense: 5}
	if n.GetTotalAttack() != 10 {
		t.Errorf("GetTotalAttack: got %d, want 10", n.GetTotalAttack())
	}
	if n.GetTotalDefense() != 5 {
		t.Errorf("GetTotalDefense: got %d, want 5", n.GetTotalDefense())
	}
	if n.GetTotalProtection() != 0 {
		t.Errorf("GetTotalProtection: got %d, want 0", n.GetTotalProtection())
	}
}

func TestNPCTakeDamage(t *testing.T) {
	n := &NPC{Health: 100, MaxHealth: 100}
	n.TakeDamage(10, nil, nil, nil)
	if n.Health != 90 {
		t.Errorf("Health after damage: got %d, want 90", n.Health)
	}
	if !n.IsAlive() {
		t.Error("NPC should still be alive")
	}

	n.TakeDamage(100, nil, nil, nil)
	if n.Health != -10 {
		t.Errorf("Health after lethal damage: got %d, want -10", n.Health)
	}
	if n.IsAlive() {
		t.Error("NPC should be dead")
	}
}

func TestNPCIsAlive(t *testing.T) {
	n1 := &NPC{State: NPCIdle}
	if !n1.IsAlive() {
		t.Error("Expected NPC with State=NPCIdle to be alive")
	}
	n2 := &NPC{State: NPCDead}
	if n2.IsAlive() {
		t.Error("Expected NPC with State=NPCDead to be dead")
	}
}

func TestNewNPC(t *testing.T) {
	arch := &Archetype{
		ID:   "orc",
		Name: "Orc",
		Stats: struct {
			HealthMin       int     `yaml:"health_min"`
			HealthMax       int     `yaml:"health_max"`
			Speed           float64 `yaml:"speed"`
			BaseAttack      int     `yaml:"base_attack"`
			BaseDefense     int     `yaml:"base_defense"`
			AttackCooldown  int     `yaml:"attack_cooldown"`
			AttackRange     float64 `yaml:"attack_range"`
			ProjectileSpeed float64 `yaml:"projectile_speed"`
		}{
			HealthMin:   50,
			HealthMax:   50,
			BaseAttack:  10,
			BaseDefense: 5,
		},
	}
	n := NewNPC(10, 20, arch, 1)
	if n.X != 10 || n.Y != 20 {
		t.Errorf("Position: got (%v, %v), want (10, 20)", n.X, n.Y)
	}
	if n.BaseAttack != 10 {
		t.Errorf("BaseAttack: got %d, want 10", n.BaseAttack)
	}
}

func TestNPCFootprint(t *testing.T) {
	n := NewNPC(10, 10, nil, 1)
	fp := n.GetFootprint()
	if len(fp.Points) == 0 {
		t.Error("NPC Footprint should have points")
	}
}

func TestNPCUpdateEscort(t *testing.T) {
	n := NewNPC(0, 0, &Archetype{ID: "escort"}, 1)
	mc := &MainCharacter{X: 10, Y: 10}
	n.updateEscort(mc, nil)
}

func TestNPCCollision(t *testing.T) {
	n := NewNPC(10, 10, nil, 1)
	obs := []*Obstacle{NewObstacle("test_npc_collider", 10.5, 10.5, nil)}
	if !n.checkCollisionAt(10.5, 10.5, obs) {
		t.Error("Expected collision at 10.5, 10.5")
	}
}

func TestNPCUpdate_Behaviors(t *testing.T) {
	mc := NewMainCharacter(10, 10, nil)
	fts := []*FloatingText{}
	projs := []*Projectile{}
	allNPCs := []*NPC{}

	n := NewNPC(0, 0, nil, 1)
	n.Speed = 1.0

	// 1. BehaviorKnightHunter (moves towards MC)
	n.Behavior = BehaviorKnightHunter
	n.X, n.Y = 0, 0
	n.Update(mc, nil, allNPCs, &projs, &fts, 100, 100, nil)
	if n.X == 0 && n.Y == 0 {
		t.Error("BehaviorKnightHunter did not move")
	}
	if n.State != NPCWalking {
		t.Error("BehaviorKnightHunter failed state transition")
	}
	if n.TargetPlayer != mc {
		t.Error("TargetPlayer not set for BehaviorKnightHunter")
	}

	// 2. BehaviorPatrol (moves towards patrol end, then back)
	n.Behavior = BehaviorPatrol
	n.TargetPlayer = nil
	n.X, n.Y = 0, 0
	n.PatrolStartX, n.PatrolStartY = 0, 0
	n.PatrolEndX, n.PatrolEndY = 10, 0
	n.PatrolHeading = true
	// Force it to reach the end
	n.X = 9.9
	n.Update(mc, nil, allNPCs, &projs, &fts, 100, 100, nil)
	if n.PatrolHeading != false {
		t.Error("BehaviorPatrol should bounce back at end")
	}

	// 3. BehaviorWander (random movement)
	n.Behavior = BehaviorWander
	n.TargetPlayer = nil
	n.X, n.Y = 0, 0
	n.Tick = 119 // trigger wander pick
	n.Update(mc, nil, allNPCs, &projs, &fts, 100, 100, nil)
	if n.WanderDirX == 0 && n.WanderDirY == 0 {
		t.Error("BehaviorWander should set new direction")
	}

	// 4. BehaviorNpcFighter (targets nearest living NPC except self)
	n.Behavior = BehaviorNpcFighter
	n.TargetPlayer = nil
	targetNPC := NewNPC(5, 5, nil, 1)
	deadNPC := NewNPC(2, 2, nil, 1)
	deadNPC.State = NPCDead
	allNPCs = []*NPC{n, deadNPC, targetNPC}
	n.X, n.Y = 0, 0
	n.Update(mc, nil, allNPCs, &projs, &fts, 100, 100, nil)
	if n.TargetNPC != targetNPC {
		t.Errorf("BehaviorNpcFighter did not acquire nearest alive NPC. Got %v", n.TargetNPC)
	}

	// 5. BehaviorChaotic (targets closest between MC or NPC)
	n.Behavior = BehaviorChaotic
	n.TargetPlayer = nil
	n.TargetNPC = nil
	mc.X, mc.Y = 20, 20             // Far
	targetNPC.X, targetNPC.Y = 5, 5 // Near
	n.X, n.Y = 0, 0
	n.Update(mc, nil, allNPCs, &projs, &fts, 100, 100, nil)
	if n.TargetNPC != targetNPC {
		t.Error("BehaviorChaotic should pick the closer NPC over the Player")
	}

	// Swap distances to test MC priority
	n.TargetNPC = nil                 // reset
	mc.X, mc.Y = 5, 5                 // Near
	targetNPC.X, targetNPC.Y = 20, 20 // Far
	n.X, n.Y = 0, 0
	n.Update(mc, nil, allNPCs, &projs, &fts, 100, 100, nil)
	if n.TargetPlayer != mc {
		t.Error("BehaviorChaotic should pick the closer Player over the NPC")
	}
}

func TestNPC_MeleeAttack(t *testing.T) {
	rand.Seed(1)                        // Ensure deterministic attack rolls so hit guarantees do not flip on the 5% margin within CI
	mc := NewMainCharacter(0.5, 0, nil) // Very close
	fts := []*FloatingText{}
	projs := []*Projectile{}

	arch := &Archetype{Stats: struct {
		HealthMin       int     `yaml:"health_min"`
		HealthMax       int     `yaml:"health_max"`
		Speed           float64 `yaml:"speed"`
		BaseAttack      int     `yaml:"base_attack"`
		BaseDefense     int     `yaml:"base_defense"`
		AttackCooldown  int     `yaml:"attack_cooldown"`
		AttackRange     float64 `yaml:"attack_range"`
		ProjectileSpeed float64 `yaml:"projectile_speed"`
	}{
		HealthMin:      50,
		HealthMax:      50,
		BaseAttack:     1000, // Guarantee hit
		BaseDefense:    5,
		AttackRange:    1.0,
		AttackCooldown: 60,
		Speed:          1.0,
	}, Behavior: "hunter"}
	n := NewNPC(0, 0, arch, 1)
	n.TargetPlayer = mc
	n.Weapon = &Weapon{MinDamage: 10, MaxDamage: 10}
	n.AttackTimer = 60 // Ready to attack

	// Loop until a hit connects (due to built-in 5% miss chance RNG)
	startHealth := mc.Health
	for i := 0; i < 100; i++ {
		n.AttackTimer = 60
		n.Update(mc, nil, []*NPC{n}, &projs, &fts, 100, 100, nil)
		if mc.Health < startHealth {
			break
		}
	}

	if n.State != NPCAttacking {
		t.Error("NPC should transition to Attacking state")
	}
	if mc.Health >= startHealth {
		t.Error("MC should have taken damage from guaranteed hit test after multiple attempts")
	}

	// Test NPC vs NPC attack
	n.TargetPlayer = nil
	targetNPC := NewNPC(0.5, 0, nil, 1)
	n.TargetNPC = targetNPC
	n.AttackTimer = 60
	startNpcHealth := targetNPC.Health
	for i := 0; i < 100; i++ {
		n.AttackTimer = 60
		n.Update(mc, nil, []*NPC{n, targetNPC}, &projs, &fts, 100, 100, nil)
		if targetNPC.Health < startNpcHealth {
			break
		}
	}

	if targetNPC.Health >= startNpcHealth {
		t.Error("Target NPC should have taken damage after multiple attempts")
	}
}

func TestNPC_RangedAttack(t *testing.T) {
	mc := NewMainCharacter(4, 0, nil) // Within ranged attack
	fts := []*FloatingText{}
	projs := []*Projectile{}

	arch := &Archetype{Stats: struct {
		HealthMin       int     `yaml:"health_min"`
		HealthMax       int     `yaml:"health_max"`
		Speed           float64 `yaml:"speed"`
		BaseAttack      int     `yaml:"base_attack"`
		BaseDefense     int     `yaml:"base_defense"`
		AttackCooldown  int     `yaml:"attack_cooldown"`
		AttackRange     float64 `yaml:"attack_range"`
		ProjectileSpeed float64 `yaml:"projectile_speed"`
	}{
		AttackRange:    5.0, // Ranged!
		AttackCooldown: 60,
		Speed:          1.0,
	}, Behavior: "hunter"}
	n := NewNPC(0, 0, arch, 1)
	n.TargetPlayer = mc
	n.AttackTimer = 60 // Ready to attack

	n.Update(mc, nil, []*NPC{n}, &projs, &fts, 100, 100, nil)

	if n.State != NPCAttacking {
		t.Error("Ranged NPC should transition to Attacking state")
	}
	if len(projs) == 0 {
		t.Error("Projectile should have been spawned")
	}

	// Test kiting behavior (too close)
	mc.X, mc.Y = 1, 0 // Inside minimum range
	n.X, n.Y = 0, 0
	n.Update(mc, nil, []*NPC{n}, &projs, &fts, 100, 100, nil)

	if math.Sqrt(math.Pow(n.X, 2)+math.Pow(n.Y, 2)) == 0 {
		t.Error("Ranged NPC should kite away when player is too close")
	}
	if n.State != NPCWalking {
		t.Error("Kiting NPC should be walking")
	}
}

type trackingImage struct {
	engine.Image
	drawnImages []engine.Image
}

func (t *trackingImage) DrawImage(img engine.Image, options *engine.DrawImageOptions) {
	t.drawnImages = append(t.drawnImages, img)
}

func TestNPCDraw_AttackAndDeadBehavior(t *testing.T) {
	staticImg := engine.NewMockImage(10, 10)
	attackImg := engine.NewMockImage(10, 10)
	corpseImg := engine.NewMockImage(10, 10)

	n := NewNPC(0, 0, &Archetype{
		StaticImage: staticImg,
		AttackImage: attackImg,
		CorpseImage: corpseImg,
	}, 1)

	// 1. Attack WITH image
	n.State = NPCAttacking
	track1 := &trackingImage{}
	n.Draw(track1, nil, nil, nil, 0, 0)
	if len(track1.drawnImages) != 1 || track1.drawnImages[0] != attackImg {
		t.Error("NPCDraw: failed to use AttackImage during attack")
	}

	// 2. Attack WITHOUT image (should fallback to static)
	n2 := NewNPC(0, 0, &Archetype{
		StaticImage: staticImg,
	}, 1)
	n2.State = NPCAttacking
	track2 := &trackingImage{}
	n2.Draw(track2, nil, nil, nil, 0, 0)
	if len(track2.drawnImages) != 1 || track2.drawnImages[0] != staticImg {
		t.Error("NPCDraw: failed to fallback to StaticImage when AttackImage is missing")
	}

	// 3. Dead WITH image
	n3 := NewNPC(0, 0, &Archetype{
		StaticImage: staticImg,
		CorpseImage: corpseImg,
	}, 1)
	n3.State = NPCDead
	track3 := &trackingImage{}
	n3.Draw(track3, nil, nil, nil, 0, 0)
	if len(track3.drawnImages) != 1 || track3.drawnImages[0] != corpseImg {
		t.Error("NPCDraw: failed to use CorpseImage during death")
	}

	// 4. Dead WITHOUT image (should draw nothing)
	n4 := NewNPC(0, 0, &Archetype{
		StaticImage: staticImg,
	}, 1)
	n4.State = NPCDead
	track4 := &trackingImage{}
	n4.Draw(track4, nil, nil, nil, 0, 0)
	if len(track4.drawnImages) != 0 {
		t.Error("NPCDraw: should not draw anything when CorpseImage is missing")
	}
}
