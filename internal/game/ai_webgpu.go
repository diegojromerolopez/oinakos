//go:build js && wasm

package game

import (
	"context"
	"syscall/js"
)

type WebGPUAIProvider struct {
	llm js.Value
}

func NewWebGPUAIProvider() *WebGPUAIProvider {
	return &WebGPUAIProvider{
		llm: js.Global().Get("oinakosWebLLM"),
	}
}

func (p *WebGPUAIProvider) Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) <-chan AIResponse {
	ch := make(chan AIResponse, 1)
	
	if p.llm.IsUndefined() || p.llm.IsNull() {
		ch <- AIResponse{Err: context.Canceled} // Or a specific error
		return ch
	}

	// Prepare history for JS
	jsHistory := js.Global().Get("Array").New()
	for _, msg := range history {
		m := js.Global().Get("Object").New()
		m.Set("role", msg.Role)
		m.Set("content", msg.Content)
		jsHistory.Call("push", m)
	}

	go func() {
		promise := p.llm.Call("chat", systemPrompt, userMessage, jsHistory)
		
		success := js.FuncOf(func(this js.Value, args []js.Value) any {
			ch <- AIResponse{Text: args[0].String()}
			return nil
		})
		defer success.Release()

		failure := js.FuncOf(func(this js.Value, args []js.Value) any {
			ch <- AIResponse{Err: context.Canceled}
			return nil
		})
		defer failure.Release()

		promise.Call("then", success, failure)
	}()

	return ch
}

func (p *WebGPUAIProvider) Decide(ctx context.Context, situation string, options []string) <-chan AIDecision {
	ch := make(chan AIDecision, 1)
	
	if p.llm.IsUndefined() || p.llm.IsNull() {
		ch <- AIDecision{ChosenOption: options[0]} // Fallback
		return ch
	}

	// Prepare options for JS
	jsOptions := js.Global().Get("Array").New()
	for _, opt := range options {
		jsOptions.Call("push", opt)
	}

	go func() {
		promise := p.llm.Call("decide", situation, jsOptions)
		
		success := js.FuncOf(func(this js.Value, args []js.Value) any {
			res := args[0]
			ch <- AIDecision{
				ChosenOption: res.Get("choice").String(),
				Reasoning:    res.Get("reasoning").String(),
			}
			return nil
		})
		defer success.Release()

		failure := js.FuncOf(func(this js.Value, args []js.Value) any {
			ch <- AIDecision{ChosenOption: options[0]}
			return nil
		})
		defer failure.Release()

		promise.Call("then", success, failure)
	}()

	return ch
}

func (p *WebGPUAIProvider) ListModels(ctx context.Context) ([]string, error) {
	// For now, return a static list or call JS
	return []string{"webgpu-local"}, nil
}
