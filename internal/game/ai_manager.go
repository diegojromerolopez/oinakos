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
