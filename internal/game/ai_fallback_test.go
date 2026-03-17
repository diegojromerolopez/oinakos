//go:build test
package game

import (
	"context"
	"fmt"
	"testing"
	"strings"
)

type FailingProvider struct{}
func (f *FailingProvider) Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) <-chan AIResponse {
	ch := make(chan AIResponse, 1)
	ch <- AIResponse{Err: fmt.Errorf("simulated failure")}
	return ch
}
func (f *FailingProvider) Decide(ctx context.Context, situation string, options []string) <-chan AIDecision {
	ch := make(chan AIDecision, 1)
	ch <- AIDecision{Err: fmt.Errorf("simulated failure")}
	return ch
}
func (f *FailingProvider) ListModels(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("simulated failure")
}

func TestFallbackAIProvider(t *testing.T) {
	primary := &FailingProvider{}
	fallback := &NativeAIProvider{}
	provider := NewFallbackAIProvider(primary, fallback)

	ctx := context.Background()
	options := []string{"attack", "flee", "wander"}
	situation := "{\"player\": {\"name\": \"Hero\", \"health_pct\": 10}, \"nearby_npcs\": [{\"alignment\": \"enemy\"}]}"

	resCh := provider.Decide(ctx, situation, options)
	res := <-resCh

	if res.Err != nil {
		t.Errorf("Expected fallback to succeed, but got error: %v", res.Err)
	}
	if !strings.Contains(strings.ToLower(res.ChosenOption), "flee") {
		t.Errorf("Expected choice to be flee at low health, got: %s", res.ChosenOption)
	}
}

func TestNativeAIProviderHeuristics(t *testing.T) {
	p := &NativeAIProvider{}
	options := []string{"attack", "flee", "wander"}

	tests := []struct {
		name      string
		situation string
		expected  string
	}{
		{
			"Low health flee",
			"{\"player\": {\"health_pct\": 15}, \"nearby_npcs\": [{\"alignment\": \"enemy\"}]}",
			"flee",
		},
		{
			"Healthy attack",
			"{\"player\": {\"health_pct\": 80}, \"nearby_npcs\": [{\"alignment\": \"enemy\"}]}",
			"attack",
		},
		{
			"No enemies wander",
			"{\"player\": {\"health_pct\": 80}, \"nearby_npcs\": []}",
			"wander",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resCh := p.Decide(context.Background(), tt.situation, options)
			res := <-resCh
			if !strings.Contains(strings.ToLower(res.ChosenOption), tt.expected) {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, res.ChosenOption)
			}
		})
	}
}
