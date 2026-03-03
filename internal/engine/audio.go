package engine

import (
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	sampleRate = 44100
)

type AudioManager struct {
	audioContext *audio.Context
	sounds       map[string]*audio.Player
	fs           fs.FS
}

var GlobalAudio *AudioManager

func InitAudio(assets fs.FS) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Warning: failed to initialize audio context (likely headless environment): %v", r)
			GlobalAudio = &AudioManager{
				sounds: make(map[string]*audio.Player),
				fs:     assets,
			}
		}
	}()

	GlobalAudio = &AudioManager{
		audioContext: audio.NewContext(sampleRate),
		sounds:       make(map[string]*audio.Player),
		fs:           assets,
	}
}

func (m *AudioManager) LoadSound(name, path string) {
	var f fs.File
	var err error

	// Check for local override in oinakos/assets folder
	localPath := "oinakos/" + path
	if _, statErr := os.Stat(localPath); statErr == nil {
		f, err = os.Open(localPath)
	} else {
		// Try embedded FS
		if m.fs != nil {
			f, err = m.fs.Open(path)
		}
		// Fallback to direct os.Open if not in embedded FS or not found
		if f == nil {
			f, err = os.Open(path)
		}
	}

	if err != nil {
		log.Printf("Warning: could not open sound file %s: %v", path, err)
		return
	}
	defer f.Close()

	var d io.Reader
	if len(path) > 4 {
		ext := path[len(path)-3:]
		switch ext {
		case "wav":
			d, err = wav.DecodeWithSampleRate(sampleRate, f)
		case "mp3":
			d, err = mp3.DecodeWithSampleRate(sampleRate, f)
		default:
			log.Printf("Warning: unsupported audio format: %s", ext)
			return
		}
	}

	if err != nil {
		log.Printf("Warning: could not decode sound file %s: %v", path, err)
		return
	}

	data, err := io.ReadAll(d)
	if err != nil {
		log.Printf("Warning: could not read sound data %s: %v", path, err)
		return
	}

	if m.audioContext == nil {
		log.Printf("Warning: audio context is nil, skipping player creation for %s", path)
		return
	}

	p := m.audioContext.NewPlayerFromBytes(data)
	if p == nil {
		log.Printf("Warning: failed to create audio player for %s", path)
		return
	}

	m.sounds[name] = p
}

func (m *AudioManager) Play(name string) {
	p, ok := m.sounds[name]
	if !ok {
		return
	}
	if p.IsPlaying() {
		p.Rewind()
	} else {
		p.Play()
	}
}

func PlaySound(name string) {
	if GlobalAudio != nil {
		GlobalAudio.Play(name)
	}
}
