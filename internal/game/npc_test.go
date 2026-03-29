package game

import (
	"math"
	"math/rand"
	"oinakos/internal/engine"
	"testing"
)

func TestNPCCalculateStat(t *testing.T) {
	n := &Character{}
	if res := n.Actor.calculateStat(10, 1); res != 10 {
		t.Errorf("calculateStat(10, 1): got %d, want 10", res)
	}
	if res := n.Actor.calculateStat(10, 10); res != 35 {
		t.Errorf("calculateStat(10, 10): got %d, want 35", res)
	}
}

func TestNPCGetters(t *testing.T) {
	n := &Character{Actor: Actor{
		Level: 1,
		Config: &EntityConfig{
			Attributes: PrimaryAttributeConfig{
				Strength: IntInterval{Min: 5, Max: 5}, Dexterity: IntInterval{Min: 2, Max: 2}, Health: IntInterval{Min: 1, Max: 1},
			},
		},
		PrimaryAttributes: PrimaryAttributes{
			Strength: 5, Dexterity: 2, Health: 1,
		},
		AgeTicks: 25.0 * float64(TicksPerYear),
	}}
	n.SyncStats(NewObjectRegistry())

	if n.GetTotalAttack() != 10 {
		t.Errorf("GetTotalAttack: got %d, want 10", n.GetTotalAttack())
	}
	if n.GetTotalDefense() != 4 {
		t.Errorf("GetTotalDefense: got %d, want 4", n.GetTotalDefense())
	}
	if n.GetTotalProtection() != 0 {
		t.Errorf("GetTotalProtection: got %d, want 0", n.GetTotalProtection())
	}
}

func TestNPCTakeDamage(t *testing.T) {
	ctx := NewTestContext()
	n := &Character{Actor: Actor{
		State: State{
			HealthPoints:    100,
			MaxHealthPoints: 100,
		},
		Config: &EntityConfig{ID: "test"},
	}}
	n.TakeDamage(10, nil, ctx)
	if n.State.HealthPoints != 90 {
		t.Errorf("Health after damage: got %d, want 90", n.State.HealthPoints)
	}
	if !n.IsAlive() {
		t.Error("NPC should still be alive")
	}

	n.TakeDamage(90, nil, ctx)
	if n.State.HealthPoints != 0 || n.ActionState != ActorIncapacitated {
		t.Errorf("Health after fatal damage: got %d, state=%v, want 0, state=ActorIncapacitated", n.State.HealthPoints, n.ActionState)
	}
	if !n.IsAlive() {
		t.Error("NPC should still be 'alive' (incapacitated) at 0 HP")
	}

	// Damage reaching death threshold (-10% of 100 = -10)
	n.TakeDamage(10, nil, ctx)
	if n.State.HealthPoints != -10 || n.ActionState != ActorDead {
		t.Errorf("Health after irremediable damage: got %d, state=%v, want -10, state=ActorDead", n.State.HealthPoints, n.ActionState)
	}
	if n.IsAlive() {
		t.Error("NPC should be truly dead at -10 HP")
	}
}

func TestNPCIsAlive(t *testing.T) {
	n1 := &Character{Actor: Actor{ActionState: ActorIdle}}
	if !n1.IsAlive() {
		t.Error("Expected NPC with ActionState=ActorIdle to be alive")
	}
	n2 := &Character{Actor: Actor{ActionState: ActorDead}}
	if n2.IsAlive() {
		t.Error("Expected NPC with ActionState=ActorDead to be dead")
	}
}

func TestNewCharacter(t *testing.T) {
	arch := &EntityConfig{
		ID:   "orc",
		Name: "Orc",
		Attributes: PrimaryAttributeConfig{
			Strength:  IntInterval{Min: 5, Max: 5},
			Dexterity: IntInterval{Min: 2, Max: 2},
			Health:    IntInterval{Min: 2, Max: 2},
			Intellect: IntInterval{Min: 2, Max: 2},
			Wisdom:    IntInterval{Min: 2, Max: 2},
		},
		Stats: EntityStatsConfig{
			HealthMin: IntInterval{Min: 100, Max: 100},
			Speed:     FloatInterval{Min: 0.5, Max: 0.5},
			Age:       AgeConfig{Current: FloatInterval{Mean: 25.0, SD: 0.0, Mode: "normal"}, Rate: 1.0},
		},
	}
	n := NewCharacter(10, 20, arch, 1, false, nil)
	if n.X != 10 || n.Y != 20 {
		t.Errorf("Position: got (%v, %v), want (10, 20)", n.X, n.Y)
	}
	if n.BaseAttack != 10 {
		t.Errorf("BaseAttack: got %d, want 10", n.BaseAttack)
	}
}

func TestNPCCollisionCircle(t *testing.T) {
	n := NewCharacter(10, 10, nil, 1, false, nil)
	c := n.GetCollisionCircle()
	if c.Radius <= 0 {
		t.Error("NPC Collision circle should have radius > 0")
	}
}

func TestNPCAllyFollowing(t *testing.T) {
	ctx := NewTestContext()
	n := NewCharacter(0, 0, nil, 1, false, nil)
	n.Alignment = AlignmentAlly
	n.Behavior = BehaviorWander // Fix flakiness: avoid default random behavior clearing TargetActor
	mc := &Character{Actor: Actor{X: 10, Y: 10, ActionState: ActorIdle}, IsPlayerControlled: true}
	ctx.World.PlayableCharacter = mc
	ctx.World.Characters = []*Character{n}

	// First update should set target to player because they are far away (dist 14.14 > 8.0)
	n.Update(ctx)

	if n.TargetActor != &mc.Actor {
		t.Errorf("Expected ally NPC to target player for rejoining, got %v", n.TargetActor)
	}
	if n.ActionState != ActorWalking {
		t.Errorf("Expected ally NPC to be walking, got %v", n.ActionState)
	}
}

func TestNPCCollision(t *testing.T) {
	arch := &EntityConfig{ID: "test_npc"}
	n := NewCharacter(10, 10, arch, 1, false, nil)
	n.AgeTicks = 25.0 * float64(TicksPerYear)
	n.SyncStats(nil)
	// Must provide non-nil archetype with Passable=false for collision to work
	colArch := &ObstacleArchetype{ID: "collider", Passable: false}
	obs := []*Obstacle{NewObstacle("test_npc_collider", 10.5, 10.5, colArch)}
	if !n.checkCollisionAt(10.5, 10.5, obs) {
		t.Error("Expected collision at 10.5, 10.5")
	}
}

func TestNPCUpdate_Behaviors(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(10, 10, nil, 1, true, nil)
	mc.State.HealthPoints = 100
	ctx.World.PlayableCharacter = mc

	n := NewCharacter(0, 0, nil, 1, false, nil)
	n.State.HealthPoints = 100
	n.Speed = 1.0
	ctx.World.Characters = []*Character{n}

	// 1. BehaviorKnightHunter (moves towards MC)
	n.Behavior = BehaviorKnightHunter
	n.X, n.Y = 0, 0
	n.Update(ctx)
	if n.X == 0 && n.Y == 0 {
		t.Error("BehaviorKnightHunter did not move")
	}
	if n.ActionState != ActorWalking {
		t.Error("BehaviorKnightHunter failed state transition")
	}
	if n.TargetActor != &mc.Actor {
		t.Error("TargetActor not set to player for BehaviorKnightHunter")
	}

	// 2. BehaviorPatrol (moves towards patrol end, then back)
	n.Behavior = BehaviorPatrol
	n.TargetActor = nil
	n.X, n.Y = 0, 0
	n.PatrolStartX, n.PatrolStartY = 0, 0
	n.PatrolEndX, n.PatrolEndY = 10, 0
	n.PatrolHeading = true
	// Force it to reach the end
	mc.X, mc.Y = 100, 100 // Move player far away so patrol continues
	n.X = 9.9
	n.Update(ctx)
	if n.PatrolHeading != false {
		t.Error("BehaviorPatrol should bounce back at end")
	}

	// 3. BehaviorWander (random movement)
	n.Behavior = BehaviorWander
	n.TargetActor = nil
	ctx.World.PlayableCharacter = nil // Clear player so it wanders
	n.X, n.Y = 0, 0
	n.Tick = 119 // trigger wander pick
	n.Update(ctx)
	if n.WanderDirX == 0 && n.WanderDirY == 0 {
		t.Error("BehaviorWander should set new direction")
	}

	// 4. BehaviorNpcFighter (targets nearest living NPC except self)
	n.Behavior = BehaviorNpcFighter
	n.TargetActor = nil
	targetNPC := NewCharacter(5, 5, nil, 1, false, nil)
	targetNPC.State.HealthPoints = 100
	targetNPC.Alignment = AlignmentAlly
	deadNPC := NewCharacter(2, 2, nil, 1, false, nil)
	deadNPC.State.HealthPoints = 0
	deadNPC.ActionState = ActorDead
	ctx.World.Characters = []*Character{n, deadNPC, targetNPC}
	n.X, n.Y = 0, 0
	n.Update(ctx)
	if n.TargetActor != &targetNPC.Actor {
		t.Errorf("BehaviorNpcFighter did not acquire nearest alive NPC. Got %v", n.TargetActor)
	}

	// 5. BehaviorChaotic (targets closest between MC or NPC)
	n.Behavior = BehaviorChaotic
	n.TargetActor = nil
	ctx.World.PlayableCharacter = mc // RESTORE PLAYER
	mc.X, mc.Y = 20, 20             // Far
	mc.State.HealthPoints = 100
	targetNPC.X, targetNPC.Y = 5, 5 // Near
	targetNPC.State.HealthPoints = 100
	n.X, n.Y = 0, 0
	n.Update(ctx)
	if n.TargetActor != &targetNPC.Actor {
		t.Error("BehaviorChaotic should pick the closer NPC over the Player")
	}

	// Swap distances to test MC priority
	n.TargetActor = nil               // reset
	mc.X, mc.Y = 5, 5                 // Near
	targetNPC.X, targetNPC.Y = 20, 20 // Far
	n.X, n.Y = 0, 0
	n.Update(ctx)
	if n.TargetActor != &mc.Actor {
		t.Error("BehaviorChaotic should pick the closer Player over the NPC")
	}
}

func TestNPC_MeleeAttack(t *testing.T) {
	rand.Seed(1) // Ensure deterministic attack rolls so hit guarantees do not flip on the 5% margin within CI
	ctx := NewTestContext()
	mc := NewCharacter(0.5, 0, nil, 1, true, nil) // Very close
	ctx.World.PlayableCharacter = mc

	arch := &EntityConfig{Stats: EntityStatsConfig{
		HealthMin:      IntInterval{Min: 50, Max: 50},
		BaseAttack:     IntInterval{Min: 1000, Max: 1000},
		BaseDefense:    IntInterval{Min: 5, Max: 5},
		AttackRange:    FloatInterval{Min: 1.0, Max: 1.0},
		AttackCooldown: IntInterval{Min: 60, Max: 60},
		Speed:          FloatInterval{Min: 1.0, Max: 1.0},
	}, Behavior: "hunter"}
	n := NewCharacter(0, 0, arch, 1, false, nil)
	n.TargetActor = &mc.Actor
	n.Weapon = &Weapon{Name: "TestWeapon", Damage: Damage{Min: 10, Max: 10}}
	n.AttackTimer = 60 // Ready to attack
	ctx.World.Characters = []*Character{n}

	// Loop until a hit connects (due to built-in 5% miss chance RNG)
	startHealth := mc.State.HealthPoints
	for i := 0; i < 100; i++ {
		n.AttackTimer = 60
		n.Update(ctx)
		if mc.State.HealthPoints < startHealth {
			break
		}
	}

	if n.ActionState != ActorAttacking {
		t.Error("NPC should transition to Attacking state")
	}
	if mc.State.HealthPoints >= startHealth {
		t.Error("MC should have taken damage from guaranteed hit test after multiple attempts")
	}

	// Test NPC vs NPC attack
	n.TargetActor = nil
	targetNPC := NewCharacter(0.5, 0, nil, 1, false, nil)
	targetNPC.Alignment = AlignmentAlly
	n.TargetActor = &targetNPC.Actor
	n.AttackTimer = 60
	ctx.World.Characters = []*Character{n, targetNPC}
	startNpcHealth := targetNPC.State.HealthPoints
	for i := 0; i < 100; i++ {
		n.AttackTimer = 60
		n.Update(ctx)
		if targetNPC.State.HealthPoints < startNpcHealth {
			break
		}
	}

	if targetNPC.State.HealthPoints >= startNpcHealth {
		t.Error("Target NPC should have taken damage after multiple attempts")
	}
}

func TestNPC_RangedAttack(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(4, 0, nil, 1, true, nil) // Within ranged attack
	mc.State.HealthPoints = 100
	ctx.World.PlayableCharacter = mc

	arch := &EntityConfig{
		Attributes: PrimaryAttributeConfig{
			Strength: IntInterval{Min: 100, Max: 100}, Dexterity: IntInterval{Min: 100, Max: 100}, Health: IntInterval{Min: 100, Max: 100}, Intellect: IntInterval{Min: 100, Max: 100}, Wisdom: IntInterval{Min: 100, Max: 100},
		},
		Stats: EntityStatsConfig{
			AttackRange:    FloatInterval{Min: 5.0, Max: 5.0},
			AttackCooldown: IntInterval{Min: 60, Max: 60},
		}, Behavior: "hunter"}
	n := NewCharacter(0, 0, arch, 1, false, nil)
	n.State.HealthPoints = 100
	n.TargetActor = &mc.Actor
	n.Slots = make(map[string]*ItemInstance)
	weaponConfig := &ObjectConfig{
		Slot: "weapon",
		Combat: &Weapon{Name: "Bow", Type: "ranged", MaxDistance: "5.0", Damage: Damage{Min: 3, Max: 6}},
	}
	n.Slots["weapon"] = NewItemInstance("bow", weaponConfig, 0, 0)
	n.UpdateEffects()
	n.AttackTimer = 60 // Ready to attack

	ctx.World.Characters = []*Character{n}

	n.Update(ctx)
	if n.ActionState != ActorAttacking {
		t.Error("Ranged NPC should transition to Attacking state")
	}
	if len(ctx.World.Projectiles) == 0 {
		t.Error("Projectile should have been spawned")
	}

	// Test kiting behavior (too close)
	mc.X, mc.Y = 1, 0 // Inside minimum range
	n.X, n.Y = 0, 0
	n.ActionState = ActorIdle // Force idle to allow movement
	n.Update(ctx)

	if math.Sqrt(math.Pow(n.X, 2)+math.Pow(n.Y, 2)) == 0 {
		t.Error("Ranged NPC should kite away when player is too close")
	}
	if n.ActionState != ActorWalking {
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
	var staticImg engine.Image = engine.NewMockImage(10, 10)
	var attackImg engine.Image = engine.NewMockImage(10, 10)
	var corpseImg engine.Image = engine.NewMockImage(10, 10)

	n := NewCharacter(0, 0, &EntityConfig{
		StaticImage: staticImg,
		AttackImage: attackImg,
		CorpseImage: corpseImg,
	}, 1, false, nil)

	// 1. Attack WITH image
	n.ActionState = ActorAttacking
	track1 := &trackingImage{}
	n.Draw(track1, nil, nil, nil, 0, 0, true)
	if len(track1.drawnImages) != 1 || track1.drawnImages[0] != attackImg {
		t.Error("NPCDraw: failed to use AttackImage during attack")
	}

	// 2. Attack WITHOUT image (should fallback to static)
	n2 := NewCharacter(0, 0, &EntityConfig{
		StaticImage: staticImg,
	}, 1, false, nil)
	n2.ActionState = ActorAttacking
	track2 := &trackingImage{}
	n2.Draw(track2, nil, nil, nil, 0, 0, true)
	if len(track2.drawnImages) != 1 || track2.drawnImages[0] != staticImg {
		t.Error("NPCDraw: failed to fallback to StaticImage when AttackImage is missing")
	}

	// 3. Dead WITH image
	n3 := NewCharacter(0, 0, &EntityConfig{
		StaticImage: staticImg,
		CorpseImage: corpseImg,
	}, 1, false, nil)
	n3.ActionState = ActorDead
	track3 := &trackingImage{}
	n3.Draw(track3, nil, nil, nil, 0, 0, true)
	if len(track3.drawnImages) != 1 || track3.drawnImages[0] != corpseImg {
		t.Error("NPCDraw: failed to use CorpseImage during death")
	}

	// 4. Dead WITHOUT image (should draw nothing)
	n4 := NewCharacter(0, 0, &EntityConfig{
		StaticImage: staticImg,
	}, 1, false, nil)
	n4.ActionState = ActorDead
	track4 := &trackingImage{}
	n4.Draw(track4, nil, nil, nil, 0, 0, true)
	if len(track4.drawnImages) != 0 {
		t.Error("NPCDraw: should not draw anything when CorpseImage is missing")
	}
}
