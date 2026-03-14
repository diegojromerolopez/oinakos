//go:build test

package engine

import "io/fs"

// LoadSprite stub for headless test builds.
func LoadSprite(assets fs.FS, path string, removeBg bool) Image {
	return nil // tests that care should mock this via the Graphics interface
}

// PlaySound stub for headless test builds.
func PlaySound(name string) {}

// PlayRandomSound stub for headless test builds.
func PlayRandomSound(prefix string) {}

// InitAudio stub for headless test builds.
func InitAudio(assets fs.FS) {
	GlobalAudio = &AudioManager{}
}

// AudioManager minimal stub for test builds.
type AudioManager struct{}

func (m *AudioManager) LoadSound(name, path string) {}
func (m *AudioManager) LoadSoundFromBytes(name string, data []byte) {}
func (m *AudioManager) HasSound(name string) bool { return false }
func (m *AudioManager) PlayRandom(prefix string)    {}
func (m *AudioManager) Play(name string)            {}

var GlobalAudio *AudioManager

// DecodeAudioRaw stub for test builds.
func DecodeAudioRaw(assets fs.FS, path string) ([]byte, error) {
	return nil, nil
}
