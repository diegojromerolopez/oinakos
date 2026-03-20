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

	// Check if WebGPU is available (WASM-only check)
	if !HasWebGPU() {
		log.Printf("WebGPU not supported in this browser. Falling back to native AI provider for WASM.")
		g.settings.AIProvider = "native"
		g.aiManager = NewAIManager(&NativeAIProvider{})
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

