//go:build !js || !wasm

package game

import (
	"context"
	"strings"
	"time"
)

func (g *Game) initAIManager() {
	if g.settings.AIProvider == "none" {
		DebugLog("AI Manager NOT initialized (provider set to 'none')")
		return
	}

	var provider AIProvider
	model := g.settings.AIModelOverride
	
	switch g.settings.AIProvider {
	case "openai":
		if model == "" { model = "gpt-4o-mini" }
		provider = NewOpenAIProvider(g.settings.OpenAIApiKey, g.settings.AIBaseURL, model)
	case "claude":
		if model == "" { model = "claude-3-5-sonnet-20240620" }
		url := g.settings.AIBaseURL
		provider = NewOpenAIProvider(g.settings.ClaudeApiKey, url, model)
	case "gemini":
		if model == "" { model = "gemini-1.5-flash" }
		url := g.settings.AIBaseURL
		if url == "" {
			url = "https://generativelanguage.googleapis.com/v1beta/openai/"
		}
		provider = NewOpenAIProvider(g.settings.GeminiApiKey, url, model)
	case "ollama":
		if model == "" { model = "llama3" }
		url := g.settings.AIBaseURL
		if url == "" {
			url = "http://localhost:11434/v1"
		}
		provider = NewOpenAIProvider("ollama", url, model)
	case "mistral":
		if model == "" { model = "mistral-small-latest" }
		url := g.settings.AIBaseURL
		if url == "" {
			url = "https://api.mistral.ai/v1"
		}
		provider = NewOpenAIProvider(g.settings.MistralApiKey, url, model)
	case "huggingface":
		if model == "" { model = "meta-llama/Llama-3.1-8B-Instruct" }
		url := g.settings.AIBaseURL
		if url == "" {
			url = "https://api-inference.huggingface.co/v1"
		}
		provider = NewOpenAIProvider(g.settings.HuggingFaceApiKey, url, model)
	default:
		provider = &NoopAIProvider{}
	}

	g.aiManager = NewAIManager(NewFallbackAIProvider(provider, &NativeAIProvider{}))

	// Fetch the actual URL used (since NewOpenAIProvider handles defaults)
	actualURL := ""
	if op, ok := provider.(*OpenAIProvider); ok {
		actualURL = op.BaseURL
	}

	// Model discovery and validation
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		availableModels, err := provider.ListModels(ctx)
		if err == nil {
			DebugLog("Available models for %s: %v", g.settings.AIProvider, availableModels)
			resolvedModel := model
			found := false
			for _, m := range availableModels {
				if m == model {
					found = true
					break
				}
				if g.settings.AIProvider == "gemini" && m == "models/"+model {
					resolvedModel = m
					found = true
					break
				}
			}
			if found && resolvedModel != model {
				DebugLog("Resolved model ID for %s: %s -> %s", g.settings.AIProvider, model, resolvedModel)
				if op, ok := provider.(*OpenAIProvider); ok {
					op.Model = resolvedModel
				}
			} else if !found && len(availableModels) > 0 {
				fallback := availableModels[0]
				for _, m := range availableModels {
					if strings.Contains(strings.ToLower(m), "flash") {
						fallback = m
						break
					}
				}
				DebugLog("Warning: configured model %s not found. Auto-selecting best fallback: %s", model, fallback)
				if op, ok := provider.(*OpenAIProvider); ok {
					op.Model = fallback
				}
			}
		} else {
			DebugLog("Error listing models for %s: %v", g.settings.AIProvider, err)
		}
	}()

	DebugLog("AI Manager initialized with provider: %s (model: %s, url: %s)", g.settings.AIProvider, model, actualURL)
}
