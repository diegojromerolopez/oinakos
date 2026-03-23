package game

import (
	"testing"
)

func TestGame_Dialogue(t *testing.T) {
	g := setupTestGame()
	npc := NewCharacter(10, 0, nil, 1, false, nil)
	npc.Name = "TestNPC"
	npc.Config = &EntityConfig{
		Dialogues: &DialogueRoot{
			StartScenarios: []StartScenario{
				{Text: "Hello!", Choices: []Choice{{Text: "Hi", Next: "node1"}}},
			},
			Nodes: map[string]*DialogueNode{
				"node1": {Text: "How are you?", Choices: []Choice{{Text: "Fine", Next: "exit"}}},
			},
		},
	}
	
	g.InitiateDialogue(npc)
	if g.ActiveDialogue == nil {
		t.Fatalf("Dialogue failed to initiate")
	}
	
	if g.ActiveDialogue.CurrentText != "Hello!" {
		t.Errorf("Unexpected dialogue text: %s", g.ActiveDialogue.CurrentText)
	}
	
	// Choose "Hi"
	g.ActiveDialogue.SelectedChoice = 0
	g.AdvanceDialogue()
	
	if g.ActiveDialogue.CurrentText != "How are you?" {
		t.Errorf("Dialogue did not advance to node1. got: %s", g.ActiveDialogue.CurrentText)
	}
	
	// Exit dialogue
	g.ActiveDialogue.SelectedChoice = 0
	g.AdvanceDialogue()
	
	if g.ActiveDialogue != nil {
		t.Error("Dialogue should have ended")
	}
}

func TestGame_DialogueEffects(t *testing.T) {
	g := setupTestGame()
	npc := NewCharacter(10, 0, nil, 1, false, nil)
	npc.Alignment = AlignmentNeutral
	
	effect := DialogueEffect{Type: "change_alignment", Value: "enemy"}
	g.ApplyDialogueEffect(npc, effect)
	
	if npc.Alignment != AlignmentEnemy {
		t.Errorf("Dialogue effect failed to change alignment. got %v", npc.Alignment)
	}
	
	effect = DialogueEffect{Type: "change_behavior", Value: "patrol"}
	g.ApplyDialogueEffect(npc, effect)
	if npc.Behavior != BehaviorPatrol {
		t.Errorf("Dialogue effect failed to change behavior. got %v", npc.Behavior)
	}
}

func TestGame_LogEvents(t *testing.T) {
	g := setupTestGame()
	g.LogEvent("Test message", LogInfo)
	
	if len(g.EventLog) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(g.EventLog))
	}
	
	if g.EventLog[0].Text != "Test message" {
		t.Errorf("Log message mismatch. got %s", g.EventLog[0].Text)
	}
}
