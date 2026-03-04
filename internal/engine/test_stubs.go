//go:build test

package engine

import "io/fs"

// LoadSprite stub for headless test builds.
func LoadSprite(assets fs.FS, path string, removeBg bool) Image {
	return nil // tests that care should mock this via the Graphics interface
}

// PlaySound stub for headless test builds.
func PlaySound(name string) {}

// InitAudio stub for headless test builds.
func InitAudio(assets fs.FS) {
	GlobalAudio = &AudioManager{}
}

// AudioManager minimal stub for test builds.
type AudioManager struct{}

var GlobalAudio *AudioManager
