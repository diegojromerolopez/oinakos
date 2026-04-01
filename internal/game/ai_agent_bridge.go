package game

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// AgentBridgeAIProvider allows an external agent (like Antigravity) to control the game 
// by reading/writing state and decision files.
type AgentBridgeAIProvider struct {
	StatePath    string
	DecisionPath string
}

func NewAgentBridgeAIProvider() *AgentBridgeAIProvider {
	dir := filepath.Join(GetOinakosDir(), "headless")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0755)
	}
	return &AgentBridgeAIProvider{
		StatePath:    filepath.Join(dir, "output.json"), // Game output (state)
		DecisionPath: filepath.Join(dir, "input.json"),  // Game input (decision)
	}
}


func (p *AgentBridgeAIProvider) Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) <-chan AIResponse {
	ch := make(chan AIResponse, 1)
	// For now, chat is not bridged, but we could do it similarly.
	ch <- AIResponse{Text: "Agent Bridge active. Please use the decision file."}
	return ch
}

func (p *AgentBridgeAIProvider) Decide(ctx context.Context, situation string, options []string) <-chan AIDecision {
	ch := make(chan AIDecision, 1)

	// 1. Write the state to file
	state := struct {
		Situation string   `json:"situation"`
		Options   []string `json:"options"`
		Timestamp string   `json:"timestamp"`
	}{
		Situation: situation,
		Options:   options,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	data, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(p.StatePath, data, 0644)

	// 2. Poll for decision
	go func() {
		defer os.Remove(p.StatePath)
		
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timeout:
				// Fallback if agent doesn't respond
				ch <- AIDecision{ChosenOption: options[0], Reasoning: "Agent timed out."}
				return
			case <-ticker.C:
				if _, err := os.Stat(p.DecisionPath); err == nil {
					data, err := os.ReadFile(p.DecisionPath)
					if err == nil {
						var dec AIDecision
						if err := json.Unmarshal(data, &dec); err == nil {
							_ = os.Remove(p.DecisionPath)
							ch <- dec
							return
						}
					}
				}
			}
		}
	}()

	return ch
}

func (p *AgentBridgeAIProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"agent-bridge"}, nil
}
