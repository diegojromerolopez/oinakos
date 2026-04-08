package game

import (
	"context"
	"testing"
)

func TestAIManager_RequestDecision(t *testing.T) {
	provider := &NoopAIProvider{}
	m := NewAIManager(provider)
	
	m.RequestDecision(context.Background(), "npc1", "world state", []string{"option1"})
	
	if len(m.pendingDecisions) != 1 {
		t.Fatalf("Expected 1 pending decision, got %d", len(m.pendingDecisions))
	}
	
	applied := m.Poll()
	if len(applied) != 1 {
		t.Errorf("Expected 1 applied decision after Poll, got %d. Noop should be immediate.", len(applied))
	}
}
