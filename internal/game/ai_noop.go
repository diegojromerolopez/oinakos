package game

import (
	"context"
)

// NoopAIProvider is a fallback provider that does nothing.
type NoopAIProvider struct{}

func (n *NoopAIProvider) Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) <-chan AIResponse {
	ch := make(chan AIResponse, 1)
	ch <- AIResponse{Text: "..."} // Or just empty
	return ch
}

func (n *NoopAIProvider) Decide(ctx context.Context, situation string, options []string) <-chan AIDecision {
	ch := make(chan AIDecision, 1)
	// Return the first option as a safe default if available
	choice := ""
	if len(options) > 0 {
		choice = options[0]
	}
	ch <- AIDecision{ChosenOption: choice}
	return ch
}

func (n *NoopAIProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"noop"}, nil
}
