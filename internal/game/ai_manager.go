package game

import (
	"context"
	"sync"
)

type NPCConversation struct {
	NPCID      string
	History    []ChatMessage
	Pending    bool
	ResponseCh <-chan AIResponse
}

type PendingDecision struct {
	NPCID string
	ResCh <-chan AIDecision
}

type AIManager struct {
	provider         AIProvider
	conversations    map[string]*NPCConversation
	pendingDecisions []PendingDecision
	mu              sync.Mutex
}

func NewAIManager(provider AIProvider) *AIManager {
	return &AIManager{
		provider:      provider,
		conversations: make(map[string]*NPCConversation),
	}
}

type AppliedDecision struct {
	NPCID    string
	Decision AIDecision
}

func (m *AIManager) Poll() []AppliedDecision {
	m.mu.Lock()
	defer m.mu.Unlock()

	var applied []AppliedDecision
	remaining := m.pendingDecisions[:0]
	for _, pd := range m.pendingDecisions {
		select {
		case dec := <-pd.ResCh:
			applied = append(applied, AppliedDecision{NPCID: pd.NPCID, Decision: dec})
		default:
			remaining = append(remaining, pd)
		}
	}
	m.pendingDecisions = remaining
	return applied
}

func (m *AIManager) RequestDecision(ctx context.Context, npcID string, worldCtx string, options []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	resCh := m.provider.Decide(ctx, worldCtx, options)
	m.pendingDecisions = append(m.pendingDecisions, PendingDecision{NPCID: npcID, ResCh: resCh})
}

func (g *Game) getAIModelForCurrentProvider() string {
	if g.settings == nil { return "" }
	switch g.settings.AIProvider {
	case "openai":
		return g.settings.OpenAIModel
	case "claude":
		return g.settings.ClaudeModel
	case "gemini":
		return g.settings.GeminiModel
	case "ollama (local)":
		return g.settings.OllamaLocalModel
	case "ollama (service)":
		return g.settings.OllamaModel
	case "mistral":
		return g.settings.MistralModel
	case "huggingface":
		return g.settings.HuggingFaceModel
	case "webgpu":
		return g.settings.WebGPUModel
	}
	return ""
}

func (g *Game) setAIModelForCurrentProvider(model string) {
	if g.settings == nil { return }
	switch g.settings.AIProvider {
	case "openai":
		g.settings.OpenAIModel = model
	case "claude":
		g.settings.ClaudeModel = model
	case "gemini":
		g.settings.GeminiModel = model
	case "ollama (local)":
		g.settings.OllamaLocalModel = model
	case "ollama (service)":
		g.settings.OllamaModel = model
	case "mistral":
		g.settings.MistralModel = model
	case "huggingface":
		g.settings.HuggingFaceModel = model
	case "webgpu":
		g.settings.WebGPUModel = model
	}
}

