package game

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// NativeAIProvider is a rule-based AI that makes decisions locally.
type NativeAIProvider struct{}

func (p *NativeAIProvider) Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) <-chan AIResponse {
	ch := make(chan AIResponse, 1)
	// Simple template-based response
	response := "I am processing the current situation."
	if strings.Contains(strings.ToLower(userMessage), "hello") {
		response = "Greetings. I am ready to act."
	}
	ch <- AIResponse{Text: response}
	return ch
}

func (p *NativeAIProvider) Decide(ctx context.Context, situation string, options []string) <-chan AIDecision {
	ch := make(chan AIDecision, 1)
	
	var world WorldContext
	if err := json.Unmarshal([]byte(situation), &world); err != nil {
		// Fallback to first option if situation parsing fails
		choice := ""
		if len(options) > 0 { choice = options[0] }
		ch <- AIDecision{ChosenOption: choice, Reasoning: "Failed to parse world context"}
		return ch
	}

	choice := p.pickBestOption(world, options)
	ch <- AIDecision{
		ChosenOption: choice,
		Reasoning:    fmt.Sprintf("Rule-based decision for %s", world.Player.Name),
	}
	return ch
}

func (p *NativeAIProvider) pickBestOption(world WorldContext, options []string) string {
	hasOption := func(opt string) bool {
		for _, o := range options {
			if strings.Contains(strings.ToLower(o), opt) {
				return true
			}
		}
		return false
	}

	getOption := func(opt string) string {
		for _, o := range options {
			if strings.Contains(strings.ToLower(o), opt) {
				return o
			}
		}
		return options[0]
	}

	// 1. High-priority: Flee if low health and enemies nearby
	if world.Player.HealthPct <= 20 && len(world.NearbyNPCs) > 0 {
		isEnemyNearby := false
		for _, n := range world.NearbyNPCs {
			if n.Alignment == "enemy" || n.Alignment == "3" { // AlignmentEnemy = 3
				isEnemyNearby = true
				break
			}
		}
		if isEnemyNearby && hasOption("flee") {
			return getOption("flee")
		}
	}

	// 2. Combat: Attack if enemies are nearby
	if len(world.NearbyNPCs) > 0 {
		for _, n := range world.NearbyNPCs {
			if n.Alignment == "enemy" || n.Alignment == "3" {
				if hasOption("attack") {
					return getOption("attack")
				}
			}
		}
	}

	// 3. Objective: Move to target if set
	if (world.Mission.TargetX != 0 || world.Mission.TargetY != 0) && hasOption("objective") {
		return getOption("objective")
	}

	// 4. Default: Wander or Idle
	if hasOption("wander") {
		return getOption("wander")
	}
	if hasOption("idle") {
		return getOption("idle")
	}

	if len(options) > 0 {
		return options[0]
	}
	return "idle"
}

func (p *NativeAIProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"native-simple"}, nil
}
