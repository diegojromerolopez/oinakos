//go:build test || headless

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
func PlayLoop(name string)        {}
func StopSound(name string)        {}
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
func (m *AudioManager) PlayLoop(name string)        {}
func (m *AudioManager) Stop(name string)            {}

var GlobalAudio *AudioManager

// DecodeAudioRaw stub for test builds.
func DecodeAudioRaw(assets fs.FS, path string) ([]byte, error) {
	return nil, nil
}

// Stubs for types used in main.go and tools to allow compilation with -tags test
type EbitenGraphics struct{ MockGraphics }

func NewEbitenGraphics() *EbitenGraphics { return &EbitenGraphics{} }

type EbitenInput struct{ MockInput }

func NewEbitenInput() *EbitenInput { return &EbitenInput{} }

type EbitenImageWrapper struct{ *MockImage }

func NewEbitenImageWrapper(img interface{}) *EbitenImageWrapper {
	return &EbitenImageWrapper{NewMockImage(0, 0)}
}
func (w *EbitenImageWrapper) UpdateRaw(img interface{}) {}
func (w *EbitenImageWrapper) GetRaw() interface{}       { return nil }
