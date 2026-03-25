//go:build !js || !wasm || test
package game

import (
	"testing"
)

func TestGame_AIModelGetSet(t *testing.T) {
	g := &Game{
		settings: DefaultSettings(),
	}

	providers := []string{"openai", "claude", "gemini", "ollama (local)", "ollama (service)", "mistral", "huggingface"}
	for _, p := range providers {
		g.settings.AIProvider = p
		original := g.getAIModelForCurrentProvider()
		
		newModel := "test-model-" + p
		g.setAIModelForCurrentProvider(newModel)
		
		got := g.getAIModelForCurrentProvider()
		if got != newModel {
			t.Errorf("Provider %s: expected %s, got %s", p, newModel, got)
		}
		
		// Restore
		g.setAIModelForCurrentProvider(original)
	}
}

func TestGame_InitAIManager_None(t *testing.T) {
	g := &Game{
		settings: DefaultSettings(),
	}
	g.settings.AIProvider = "none"
	g.initAIManager()
	
	if g.aiManager != nil {
		t.Error("Expected aiManager to be nil when provider is 'none'")
	}
}

func TestGame_InitAIManager_Providers(t *testing.T) {
	g := &Game{
		settings: DefaultSettings(),
	}
	providers := []string{"openai", "claude", "gemini", "ollama (local)", "ollama (service)", "mistral", "huggingface"}
	for _, p := range providers {
		t.Run(p, func(t *testing.T) {
			g.settings.AIProvider = p
			g.initAIManager()
			if g.aiManager == nil {
				t.Errorf("Expected aiManager to be initialized for %s", p)
			}
		})
	}

	// Test default / invalid
	t.Run("default_noop", func(t *testing.T) {
		g.settings.AIProvider = "invalid-provider"
		g.initAIManager()
		if g.aiManager == nil {
			t.Error("Expected aiManager to be initialized (even with default noop)")
		}
	})
}
