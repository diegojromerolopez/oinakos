//go:build !test

package engine

import (
	"io"
	"io/fs"
	"log"
	"os"
	"sync"

	"math/rand"

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
	mu           sync.RWMutex
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
	m.mu.RLock()
	if _, ok := m.sounds[name]; ok {
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()

	data, err := DecodeAudioRaw(m.fs, path)
	if err != nil {
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

	m.mu.Lock()
	m.sounds[name] = p
	m.mu.Unlock()
}

// DecodeAudioRaw performs file reading and decoding into raw PCM data.
// This can be safely called from background goroutines.
func DecodeAudioRaw(assets fs.FS, path string) ([]byte, error) {
	var f fs.File
	var err error

	// Check for local override in oinakos/assets folder
	localPath := "oinakos/" + path
	if _, statErr := os.Stat(localPath); statErr == nil {
		f, err = os.Open(localPath)
	} else {
		// Try embedded FS
		if assets != nil {
			f, err = assets.Open(path)
		}
		// Fallback to direct os.Open if not in embedded FS or not found
		if f == nil {
			f, err = os.Open(path)
		}
	}

	if err != nil {
		log.Printf("Warning: could not open sound file %s: %v", path, err)
		return nil, err
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
			return nil, io.ErrUnexpectedEOF
		}
	}

	if err != nil {
		log.Printf("Warning: could not decode sound file %s: %v", path, err)
		return nil, err
	}

	data, err := io.ReadAll(d)
	if err != nil {
		log.Printf("Warning: could not read sound data %s: %v", path, err)
		return nil, err
	}
	return data, nil
}

func (m *AudioManager) Play(name string) {
	m.mu.RLock()
	p, ok := m.sounds[name]
	m.mu.RUnlock()
	if !ok {
		return
	}
	if p.IsPlaying() {
		p.Rewind()
	} else {
		p.Play()
	}
}

func (m *AudioManager) getMatchingKeys(prefix string) []string {
	var matches []string
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k := range m.sounds {
		// Key matches prefix exactly, or key starts with prefix + "_" or "/" (e.g. attack_1, orc/hit)
		if k == prefix || (len(k) > len(prefix) && k[:len(prefix)] == prefix && (k[len(prefix)] == '_' || k[len(prefix)] == '/')) {
			matches = append(matches, k)
		}
	}
	return matches
}

func (m *AudioManager) PlayRandom(prefix string) {
	keys := m.getMatchingKeys(prefix)
	if len(keys) == 0 {
		return
	}
	// Note: We use a simple rand here. In a real engine we might want
	// a deterministic or seeded one, but for audio variety this is fine.
	key := keys[rand.Intn(len(keys))]
	m.Play(key)
}

func PlaySound(name string) {
	if GlobalAudio != nil {
		GlobalAudio.Play(name)
	}
}

func PlayRandomSound(prefix string) {
	if GlobalAudio != nil {
		GlobalAudio.PlayRandom(prefix)
	}
}
