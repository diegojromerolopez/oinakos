//go:build !js || !wasm

package game

import (
	"context"
	"time"
)

func (g *Game) initAIManager() {
	if g.settings.AIProvider == "none" {
		DebugLog("AI Manager NOT initialized (provider set to 'none')")
		return
	}

	var provider AIProvider
	var model string

	switch g.settings.AIProvider {
	case "openai":
		model = g.settings.OpenAIModel
		provider = NewOpenAIProvider(g.settings.OpenAIApiKey, g.settings.AIBaseURL, model)
	case "claude":
		model = g.settings.ClaudeModel
		provider = NewOpenAIProvider(g.settings.ClaudeApiKey, g.settings.AIBaseURL, model)
	case "gemini":
		model = g.settings.GeminiModel
		url := g.settings.AIBaseURL
		if url == "" {
			url = "https://generativelanguage.googleapis.com/v1beta/openai/"
		}
		provider = NewOpenAIProvider(g.settings.GeminiApiKey, url, model)
	case "ollama (local)":
		model = g.settings.OllamaLocalModel
		url := g.settings.AIBaseURL
		if url == "" {
			url = "http://localhost:11434/v1"
		}
		provider = NewOpenAIProvider("ollama", url, model)
	case "ollama (service)":
		model = g.settings.OllamaModel
		url := g.settings.AIBaseURL
		if url == "" {
			url = "http://localhost:11434/v1"
		}
		provider = NewOpenAIProvider("ollama", url, model)
	case "mistral":
		model = g.settings.MistralModel
		url := g.settings.AIBaseURL
		if url == "" {
			url = "https://api.mistral.ai/v1"
		}
		provider = NewOpenAIProvider(g.settings.MistralApiKey, url, model)
	case "huggingface":
		model = g.settings.HuggingFaceModel
		url := g.settings.AIBaseURL
		if url == "" {
			url = "https://api-inference.huggingface.co/v1"
		}
		provider = NewOpenAIProvider(g.settings.HuggingFaceApiKey, url, model)
	case "bridge":
		provider = NewAgentBridgeAIProvider()
		model = "agent-bridge"
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
	g.availableModels = nil
	g.isFetchingModels = true
	go func() {
		defer func() { g.isFetchingModels = false }()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		availableModels, err := provider.ListModels(ctx)
		if err == nil {
			g.availableModels = availableModels
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
				g.setAIModelForCurrentProvider(resolvedModel)
				g.settings.Save()
			} else if !found && len(availableModels) > 0 {
				fallback := g.settings.GetDefaultModel(g.settings.AIProvider)
				DebugLog("Warning: configured model %s not found for %s. Auto-selecting default: %s", model, g.settings.AIProvider, fallback)
				if op, ok := provider.(*OpenAIProvider); ok {
					op.Model = fallback
				}
				g.setAIModelForCurrentProvider(fallback)
				g.settings.Save()
			}
		} else {
			DebugLog("Error listing models for %s: %v", g.settings.AIProvider, err)
		}
	}()

	DebugLog("AI Manager initialized with provider: %s (model: %s, url: %s)", g.settings.AIProvider, model, actualURL)
}

