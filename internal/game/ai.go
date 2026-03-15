package game

import (
	"context"
)

// AIProvider is the interface for all AI backends.
type AIProvider interface {
	// Chat sends a conversation and returns the NPC's reply.
	Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) <-chan AIResponse
	
	// Decide asks the AI to pick an action from a list of options.
	Decide(ctx context.Context, situation string, options []string) <-chan AIDecision
}

type ChatMessage struct {
	Role    string // "user" | "assistant" | "system"
	Content string
}

type AIResponse struct {
	Text string
	Err  error
}

type AIDecision struct {
	ChosenOption string
	Reasoning    string // optional, for debug logging
	Err          error
}
