package game

import (
	"math/rand"
	"oinakos/internal/engine"
)

// AudioManager defines an interface for all audio operations to allow mocking.
type AudioManager interface {
	PlaySound(name string)
	PlayRandomSound(prefix string)
	SetProbability(prob float64)
}

// DefaultAudioManager uses the actual engine audio functions with a probability filter.
type DefaultAudioManager struct {
	probability float64
}

func (d *DefaultAudioManager) SetProbability(prob float64) {
	d.probability = prob
}

func (d *DefaultAudioManager) PlaySound(name string) {
	if rand.Float64() < d.probability {
		engine.PlaySound(name)
	}
}

func (d *DefaultAudioManager) PlayRandomSound(prefix string) {
	if rand.Float64() < d.probability {
		engine.PlayRandomSound(prefix)
	}
}

// MockAudioManager can be used in headless tests to prevent Ebiten panics.
type MockAudioManager struct {
	PlayedSounds []string
	probability  float64
}

func NewMockAudioManager() *MockAudioManager {
	return &MockAudioManager{
		PlayedSounds: make([]string, 0),
		probability:  1.0, // Default to 1.0 for tests
	}
}

func (m *MockAudioManager) SetProbability(prob float64) {
	m.probability = prob
}

func (m *MockAudioManager) PlaySound(name string) {
	if rand.Float64() < m.probability {
		m.PlayedSounds = append(m.PlayedSounds, name)
	}
}

func (m *MockAudioManager) PlayRandomSound(prefix string) {
	if rand.Float64() < m.probability {
		m.PlayedSounds = append(m.PlayedSounds, "RANDOM:"+prefix)
	}
}
