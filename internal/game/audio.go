package game

import "oinakos/internal/engine"

// AudioManager defines an interface for all audio operations to allow mocking.
type AudioManager interface {
	PlaySound(name string)
}

// DefaultAudioManager uses the actual engine audio functions.
type DefaultAudioManager struct{}

func (d *DefaultAudioManager) PlaySound(name string) {
	engine.PlaySound(name)
}

// MockAudioManager can be used in headless tests to prevent Ebiten panics.
type MockAudioManager struct {
	PlayedSounds []string
}

func NewMockAudioManager() *MockAudioManager {
	return &MockAudioManager{
		PlayedSounds: make([]string, 0),
	}
}

func (m *MockAudioManager) PlaySound(name string) {
	m.PlayedSounds = append(m.PlayedSounds, name)
}
