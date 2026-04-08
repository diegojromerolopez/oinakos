package game

import (
	"testing"
)

func TestCharacter_ActionsCoverage(t *testing.T) {
	ctx := NewTestContext()
	g := ctx.World.Game
	if g == nil {
		g = &Game{
			World:      ctx.World,
			Registries: ctx.Registries,
		}
		ctx.World.Game = g
	}
	p := NewCharacter(0, 0, nil, 1, true, nil)
	g.playableCharacter = p
	
	target := NewCharacter(1, 1, nil, 1, false, nil)
	target.Name = "Stultus"
	ctx.World.Characters = append(ctx.World.Characters, target)
	
	// 1. Dialogue triggers (Game methods)
	target.Config = &EntityConfig{
		Dialogues: &DialogueRoot{
			StartScenarios: []StartScenario{
				{Weight: 1.0, Text: "Hello", Choices: []Choice{{Text: "Bye", Next: "exit"}}},
			},
		},
	}
	g.InitiateDialogue(target)
	if g.ActiveDialogue == nil { t.Error("Dialogue should be active") }
	g.AdvanceDialogue()
	if g.ActiveDialogue != nil { t.Error("Dialogue should be closed after exit choice") }

	// 2. Character methods
	p.Rest(ctx)
	target.Config.Stats.IsMilkable = true
	target.RawStats.IsMilkable = true
	if ctx.Registries.Objects == nil { ctx.Registries.Objects = NewObjectRegistry() }
	ctx.Registries.Objects.Objects["milk"] = &ObjectConfig{Name: "Milk"}
	p.Milk(&target.Actor, ctx)

	// 3. Attack hits variants
	p.ActionState = ActorChopping
	p.CheckAttackHits(ctx, "")
	
	p.ActionState = ActorDigging
	p.CheckAttackHits(ctx, "")
	
	p.ActionState = ActorForaging
	p.CheckAttackHits(ctx, "")

	// 4. Movement
	p.MoveTo(ctx, 10, 10)
	p.ExecutePathTo(ctx, 5, 5)
}
