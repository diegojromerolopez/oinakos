package game
import "testing"
import "fmt"

func TestSwitch(t *testing.T) {
	ctx := NewTestContext()
	mc := NewCharacter(100, 100, nil, 1, true, nil)
	ctx.World.PlayableCharacter = mc

	fighter := NewCharacter(0, 0, &EntityConfig{ID: "fighter"}, 1, false, nil)
	fighter.Behavior = BehaviorNpcFighter
	fighter.Alignment = AlignmentEnemy

	victim1 := NewCharacter(2, 0, &EntityConfig{ID: "v1"}, 1, false, nil)
	victim1.Alignment = AlignmentAlly
	victim1.State.HealthPoints, victim1.Name = 100, "victim1"

	victim2 := NewCharacter(5, 0, &EntityConfig{ID: "v2"}, 1, false, nil)
	victim2.Alignment = AlignmentAlly
	victim2.State.HealthPoints, victim2.Name = 100, "victim2"

	ctx.World.Characters = []*Character{fighter, victim1, victim2}
	fighter.State.HealthPoints = 100
	mc.X, mc.Y = 100, 100

	fighter.Update(ctx)
	fmt.Printf("DEBUG1 Target: %v\n", fighter.TargetActor != nil)

	victim1.ActionState = ActorDead
	fmt.Printf("before Update 2: v1 Alive=%v, TargetActor nil=%v\n", victim1.IsAlive(), fighter.TargetActor == nil)
	
	_, _, ht, _ := fighter.findTarget(mc, ctx.World.Characters, 141.4)
	fmt.Printf("before Update 2 findTarget: ht=%v -> target nil? %v (Target points to v1? %v)\n", ht, fighter.TargetActor == nil, fighter.TargetActor == &victim1.Actor)

	fighter.Update(ctx)
	fmt.Printf("DEBUG2 Target nil? %v\n", fighter.TargetActor == nil)
}
