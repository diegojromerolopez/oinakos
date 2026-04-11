package game

import (
	"testing"
)

func TestDebtMechanicsFull(t *testing.T) {
	ctx := NewTestContext()
	ctx.Settings.AdultMode = true
	g := ctx.World.Game
	
	t.Run("Loan Interest Calculation", func(t *testing.T) {
		p := g.World.PlayableCharacter
		p.Debts = nil
		p.Denarii = 0
		
		npc := &Character{Actor: Actor{UID: "banker_test"}}
		g.ApplyDialogueEffect(npc, DialogueEffect{Type: "borrow_100"})
		if len(p.Debts) != 1 || p.Debts[0].Amount != 120 {
			t.Errorf("Expected 120 debt for 100 loan, got %+v", p.Debts)
		}
	})

	t.Run("Credit Limit Enforcement", func(t *testing.T) {
		p := g.World.PlayableCharacter
		p.Debts = nil
		
		banker1 := &Character{Actor: Actor{UID: "banker1"}}
		banker2 := &Character{Actor: Actor{UID: "banker2"}}
		banker3 := &Character{Actor: Actor{UID: "banker3"}}
		
		g.ApplyDialogueEffect(banker1, DialogueEffect{Type: "borrow_100"})
		g.ApplyDialogueEffect(banker2, DialogueEffect{Type: "borrow_100"})
		
		if len(p.Debts) != 2 {
			t.Errorf("Should have 2 loans")
		}
		
		g.ApplyDialogueEffect(banker3, DialogueEffect{Type: "borrow_100"})
		if len(p.Debts) > 2 {
			t.Errorf("Should NOT be able to take a 3rd loan")
		}
	})

	t.Run("Partial Payment and Default", func(t *testing.T) {
		debtor := NewCharacter(0, 0, &EntityConfig{ID: "peasant"}, 25, false, nil)
		debtor.Debts = []Loan{{LenderUID: "banker_uid", Amount: 100, Deadline: 1000}}
		debtor.Denarii = 60
		
		g.playableCharacter = debtor
		g.World.PlayableCharacter = debtor
		banker := &Character{Actor: Actor{UID: "banker_uid"}}
		g.ApplyDialogueEffect(banker, DialogueEffect{Type: "repay_50"})
		
		if len(debtor.Debts) != 1 || debtor.Debts[0].Amount != 50 {
			t.Errorf("Partial payment failed, debt is %d", debtor.Debts[0].Amount)
		}
		
		ctx.World.State.Ticks = 1001
		debtor.updateAI(ctx)
		
		if debtor.Behavior != BehaviorSlave {
			t.Errorf("Debtor should be enslaved if any amount remains at deadline")
		}
	})

	t.Run("Full Prepayment Success", func(t *testing.T) {
		p := NewCharacter(0, 0, &EntityConfig{ID: "peasant"}, 25, true, nil)
		g.playableCharacter = p
		g.World.PlayableCharacter = p
		p.Debts = []Loan{{LenderUID: "banker_uid", Amount: 100, Deadline: 1000}}
		p.Denarii = 150
		
		banker := &Character{Actor: Actor{UID: "banker_uid"}}
		g.ApplyDialogueEffect(banker, DialogueEffect{Type: "repay_debt"})
		
		if len(p.Debts) != 0 {
			t.Errorf("Full repayment failed")
		}
		
		ctx.World.State.Ticks = 1001
		p.updateAI(ctx)
		
		if p.Behavior == BehaviorSlave {
			t.Errorf("Should not be enslaved if debt was cleared early")
		}
	})

	t.Run("Generational Debt Inheritance", func(t *testing.T) {
		parent := NewCharacter(0, 0, &EntityConfig{ID: "parent"}, 25, false, nil)
		parent.Name = "ParentName"
		parent.Debts = []Loan{{LenderUID: "guild1", Amount: 1000, Deadline: 5000}}
		
		child := NewCharacter(1, 1, &EntityConfig{ID: "child"}, 25, false, nil)
		child.ParentID = parent.Name
		
		ctx.World.Characters = append(ctx.World.Characters, parent, child)
		parent.die(nil, ctx)
		
		if len(child.Debts) != 1 || child.Debts[0].Amount != 1000 {
			t.Errorf("Child did not inherit parent's ledger")
		}
	})
}
