//go:build js && wasm

package game

import (
	"log"
)

func (g *Game) initAIManager() {
	if g.settings.AIProvider == "none" {
		DebugLog("AI Manager NOT initialized (provider set to 'none')")
		return
	}

	// In WASM, we only support WebGPU LLM.
	// If the user has any other provider selected, we override it to WebGPU.
	if g.settings.AIProvider != "webgpu" {
		log.Printf("WASM detected: Overriding AI provider %s to webgpu (local LLM)", g.settings.AIProvider)
		g.settings.AIProvider = "webgpu"
	}

	provider := NewWebGPUAIProvider()
	g.aiManager = NewAIManager(NewFallbackAIProvider(provider, &NativeAIProvider{}))

	DebugLog("AI Manager initialized with WebGPU provider")
}
