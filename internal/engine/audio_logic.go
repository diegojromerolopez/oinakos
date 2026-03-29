package engine

import (
	"math/rand"
	"strings"
	"sync"
)

// SharedAudioManager contains the non-Ebiten logic for sound management.
type SharedAudioManager struct {
	mu     sync.RWMutex
	sounds []string // Just the keys for logic testing
}

func (m *SharedAudioManager) getMatchingKeys(prefix string, availableKeys []string) []string {
	var matches []string
	for _, k := range availableKeys {
		// Key matches prefix exactly, or key starts with prefix + "_" or "/" (e.g. attack_1, orc/hit)
		if k == prefix || (len(k) > len(prefix) && strings.HasPrefix(k, prefix) && (k[len(prefix)] == '_' || k[len(prefix)] == '/')) {
			matches = append(matches, k)
		}
	}
	return matches
}

func (m *SharedAudioManager) PickRandom(prefix string, availableKeys []string) string {
	keys := m.getMatchingKeys(prefix, availableKeys)
	if len(keys) == 0 {
		return ""
	}
	return keys[rand.Intn(len(keys))]
}
