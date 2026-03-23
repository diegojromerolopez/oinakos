package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	if s.SoundFrequency != "rare" {
		t.Errorf("expected rare, got %s", s.SoundFrequency)
	}
	if s.TalkingFrequency != "infrequent" {
		t.Errorf("expected infrequent, got %s", s.TalkingFrequency)
	}
	if s.Font != "medieval" {
		t.Errorf("expected medieval, got %s", s.Font)
	}
	if s.FogOfWar != "none" {
		t.Errorf("expected none, got %s", s.FogOfWar)
	}
}

func TestGetFrequencyProb(t *testing.T) {
	s := &Settings{}
	tests := []struct {
		freq string
		want float64
	}{
		{"never", 0.0},
		{"rare", 0.05},
		{"infrequent", 0.15},
		{"half the time", 0.4},
		{"frequent", 0.7},
		{"always", 1.0},
		{"unknown", 0.1},
	}

	for _, tt := range tests {
		if got := s.getFrequencyProb(tt.freq); got != tt.want {
			t.Errorf("getFrequencyProb(%q) = %v, want %v", tt.freq, got, tt.want)
		}
	}
}

func TestGetOinakosDir(t *testing.T) {
	// Test override
	tmpDir, err := os.MkdirTemp("", "oinakos_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	overridePath := filepath.Join(tmpDir, "override")
	SetOinakosDir(overridePath)
	
	got := GetOinakosDir()
	if got != overridePath {
		t.Errorf("expected %s, got %s", overridePath, got)
	}

	// Ensure it was created
	if _, err := os.Stat(overridePath); os.IsNotExist(err) {
		t.Error("override path was not created")
	}

	// Reset override for other tests
	SetOinakosDir("")
}

func TestSettings_GetProbabilities(t *testing.T) {
	s := &Settings{
		SoundFrequency:   "always",
		TalkingFrequency: "never",
	}
	if got := s.GetSoundProbability(); got != 1.0 {
		t.Errorf("expected 1.0, got %v", got)
	}
	if got := s.GetTalkingProbability(); got != 0.0 {
		t.Errorf("expected 0.0, got %v", got)
	}
}
