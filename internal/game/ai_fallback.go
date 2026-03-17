package game

import (
	"context"
)

// FallbackAIProvider wraps two providers and switches to the fallback if the primary fails.
type FallbackAIProvider struct {
	Primary  AIProvider
	Fallback AIProvider
}

func NewFallbackAIProvider(primary, fallback AIProvider) *FallbackAIProvider {
	return &FallbackAIProvider{
		Primary:  primary,
		Fallback: fallback,
	}
}

func (f *FallbackAIProvider) Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) <-chan AIResponse {
	ch := make(chan AIResponse, 1)
	go func() {
		resCh := f.Primary.Chat(ctx, systemPrompt, userMessage, history)
		res := <-resCh
		if res.Err != nil {
			DebugLog("[AI-FALLBACK] Primary Chat failed: %v. Using fallback AI.", res.Err)
			resFallbackCh := f.Fallback.Chat(ctx, systemPrompt, userMessage, history)
			ch <- <-resFallbackCh
		} else {
			ch <- res
		}
	}()
	return ch
}

func (f *FallbackAIProvider) Decide(ctx context.Context, situation string, options []string) <-chan AIDecision {
	ch := make(chan AIDecision, 1)
	go func() {
		resCh := f.Primary.Decide(ctx, situation, options)
		res := <-resCh
		if res.Err != nil {
			DebugLog("[AI-FALLBACK] Primary Decide failed for situation. Using fallback AI. Error: %v", res.Err)
			resFallbackCh := f.Fallback.Decide(ctx, situation, options)
			ch <- <-resFallbackCh
		} else {
			ch <- res
		}
	}()
	return ch
}

func (f *FallbackAIProvider) ListModels(ctx context.Context) ([]string, error) {
	models, err := f.Primary.ListModels(ctx)
	if err != nil {
		return f.Fallback.ListModels(ctx)
	}
	return models, nil
}
