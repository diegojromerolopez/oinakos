package game

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestNativeAIProvider_Coverage(t *testing.T) {
	p := &NativeAIProvider{}
	ctx := context.Background()

	t.Run("Chat greeting", func(t *testing.T) {
		resCh := p.Chat(ctx, "", "Hello there", nil)
		res := <-resCh
		if res.Text != "Greetings. I am ready to act." {
			t.Errorf("unexpected chat response: %s", res.Text)
		}
	})

	t.Run("Decide with valid context", func(t *testing.T) {
		world := WorldContext{
			Player: PlayerContext{Name: "Boris", HealthPct: 100},
			NearbyNPCs: []NPCContext{
				{Name: "Enemy", Alignment: "enemy"},
			},
		}
		data, _ := json.Marshal(world)
		
		options := []string{"attack", "flee", "wander"}
		decCh := p.Decide(ctx, string(data), options)
		dec := <-decCh
		
		if dec.ChosenOption != "attack" {
			t.Errorf("expected native provider to choose attack when enemy is nearby, got %s", dec.ChosenOption)
		}
	})

	t.Run("Decide under pressure (low health)", func(t *testing.T) {
		world := WorldContext{
			Player: PlayerContext{Name: "Boris", HealthPct: 10},
			NearbyNPCs: []NPCContext{
				{Name: "Scary Ogre", Alignment: "enemy"},
			},
		}
		data, _ := json.Marshal(world)
		
		options := []string{"attack", "flee", "objective"}
		decCh := p.Decide(ctx, string(data), options)
		dec := <-decCh
		
		if dec.ChosenOption != "flee" {
			t.Errorf("expected native provider to choose flee when health is low, got %s", dec.ChosenOption)
		}
	})
	
	t.Run("Decision fallback for bad JSON", func(t *testing.T) {
		options := []string{"wander", "idle"}
		decCh := p.Decide(ctx, "invalid json", options)
		dec := <-decCh
		if dec.ChosenOption != "wander" {
			t.Errorf("expected fallback to first option for invalid JSON, got %s", dec.ChosenOption)
		}
	})
}

func TestAIManager_Poll(t *testing.T) {
	p := &NativeAIProvider{}
	m := NewAIManager(p)
	ctx := context.Background()

	m.RequestDecision(ctx, "npc1", "{}", []string{"idle"})
	
	// Native provider returns immediately on channel
	applied := m.Poll()
	if len(applied) != 1 || applied[0].NPCID != "npc1" {
		t.Errorf("expected 1 applied decision for npc1, got %v", applied)
	}
	
	// Subsequent poll should be empty
	applied = m.Poll()
	if len(applied) != 0 {
		t.Error("expected empty poll after results were consumed")
	}
}

func TestAgentBridgeAIProvider_Coverage(t *testing.T) {
	p := &AgentBridgeAIProvider{}
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Errorf("ListModels failed: %v", err)
	}
	if len(models) == 0 {
		t.Error("expected at least one model from bridge provider")
	}
	
	ch := p.Chat(context.Background(), "", "msg", nil)
	select {
	case res := <-ch:
		if res.Text == "" { t.Error("empty response from bridge") }
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for bridge chat")
	}
}

func TestAIProvider_InterfaceConsistency(t *testing.T) {
	native := &NativeAIProvider{}
	providers := []AIProvider{
		native,
		&FallbackAIProvider{Primary: native, Fallback: native},
		&NoopAIProvider{},
		&AgentBridgeAIProvider{},
	}
	
	for _, p := range providers {
		models, err := p.ListModels(context.Background())
		if err != nil {
			t.Errorf("provider failed ListModels: %v", err)
		}
		if len(models) == 0 { }
	}
}
