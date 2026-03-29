package game

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)


func TestAgentBridgeAIProvider(t *testing.T) {
	provider := NewAgentBridgeAIProvider()
	statePath := provider.StatePath
	decisionPath := provider.DecisionPath

	// Cleanup
	os.Remove(statePath)
	os.Remove(decisionPath)
	defer os.Remove(statePath)
	defer os.Remove(decisionPath)


	situation := "Test situation"
	options := []string{"option1", "option2"}

	// Start a goroutine to simulate the external agent
	go func() {
		// Wait for state file
		for i := 0; i < 100; i++ {
			if _, err := os.Stat(statePath); err == nil {
				// Read state
				data, _ := os.ReadFile(statePath)
				var state map[string]any
				json.Unmarshal(data, &state)

				// Write decision
				decision := AIDecision{
					ChosenOption: "option2",
					Reasoning:    "Test reasoning",
				}
				decData, _ := json.Marshal(decision)
				os.WriteFile(decisionPath, decData, 0644)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	ctx := context.Background()
	decChan := provider.Decide(ctx, situation, options)
	
	var decision AIDecision
	select {
	case decision = <-decChan:
	case <-time.After(1 * time.Second):
		t.Fatalf("Timed out waiting for decision")
	}

	if decision.Err != nil {
		t.Fatalf("Decide failed: %v", decision.Err)
	}

	if decision.ChosenOption != "option2" {
		t.Errorf("Expected option2, got %s", decision.ChosenOption)
	}
	if decision.Reasoning != "Test reasoning" {
		t.Errorf("Expected 'Test reasoning', got %s", decision.Reasoning)
	}

	// Test timeout
	os.Remove(statePath)
	os.Remove(decisionPath)
	
	provider.Chat(ctx, "system", "user", nil)
	models, err := provider.ListModels(ctx)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 1 || models[0] != "agent-bridge" {
		t.Errorf("Unexpected models: %v", models)
	}
}
