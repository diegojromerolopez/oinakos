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

func TestSettings_SaveLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oinakos_settings_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	SetOinakosDir(tmpDir)
	defer SetOinakosDir("")

	s := DefaultSettings()
	s.OpenAIModel = "gpt-test"
	s.Units = "si"

	err = s.Save()
	if err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	loaded := LoadSettings()
	if loaded.OpenAIModel != "gpt-test" {
		t.Errorf("expected gpt-test, got %s", loaded.OpenAIModel)
	}
	if loaded.Units != "si" {
		t.Errorf("expected si, got %s", loaded.Units)
	}
}

func TestSettings_GetDefaultModel(t *testing.T) {
	s := &Settings{}
	tests := []struct {
		provider string
		want     string
	}{
		{"openai", "gpt-4o-mini"},
		{"ollama (service)", "llama3"},
		{"ollama (local)", "llama3"},
		{"claude", "claude-3-5-sonnet-20240620"},
		{"gemini", "gemini-1.5-flash"},
		{"mistral", "mistral-small-latest"},
		{"huggingface", "meta-llama/Llama-3.1-8B-Instruct"},
		{"webgpu", "tiny-llama-1.1b"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		if got := s.GetDefaultModel(tt.provider); got != tt.want {
			t.Errorf("GetDefaultModel(%q) = %q, want %q", tt.provider, got, tt.want)
		}
	}
}

func TestSettings_GetFrequencyProb(t *testing.T) {
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
		got := s.getFrequencyProb(tt.freq)
		if got != tt.want {
			t.Errorf("getFrequencyProb(%s) = %v, want %v", tt.freq, got, tt.want)
		}
	}
}

func TestSettings_FormatWeight(t *testing.T) {
	s := &Settings{Units: "si"}
	if got := s.FormatWeight(10.5); got != "10.50 kg" {
		t.Errorf("SI: expected 10.50 kg, got %s", got)
	}

	s.Units = "imperial"
	// 10.5 * 2.20462 = 23.14851
	if got := s.FormatWeight(10.5); got != "23.15 lb" {
		t.Errorf("Imperial: expected 23.15 lb, got %s", got)
	}

	s.Units = "venburguian"
	// 10.5 / 0.329 = 31.9148...
	if got := s.FormatWeight(10.5); got != "31.91 lib" {
		t.Errorf("Venburguian: expected 31.91 lib, got %s", got)
	}
}

func TestSettings_FormatDistance(t *testing.T) {
	s := &Settings{Units: "si"}
	if got := s.FormatDistance(10.0); got != "2.96 m" {
		t.Errorf("SI: expected 2.96 m, got %s", got)
	}

	s.Units = "imperial"
	if got := s.FormatDistance(10.0); got != "9.71 ft" {
		t.Errorf("Imperial: expected 9.71 ft, got %s", got)
	}

	s.Units = "venburguian"
	if got := s.FormatDistance(4.0); got != "4.0 pedes" {
		t.Errorf("Venburguian small: expected 4.0 pedes, got %s", got)
	}
	if got := s.FormatDistance(10.0); got != "2.0 passus" {
		t.Errorf("Venburguian medium: expected 2.0 passus, got %s", got)
	}
	if got := s.FormatDistance(10000.0); got != "2.0 millia" {
		t.Errorf("Venburguian large: expected 2.0 millia, got %s", got)
	}
}

func TestSettings_FormatUnits_Unsupported(t *testing.T) {
	s := &Settings{Units: "unknown"}
	if got := s.FormatWeight(10.0); got != "10.00" {
		t.Errorf("Weight unknown: expected 10.00, got %s", got)
	}
	if got := s.FormatDistance(10.0); got != "10.00 units" {
		t.Errorf("Distance unknown: expected 10.00 units, got %s", got)
	}
}

func TestLoadSettings_MissingFile(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "oinakos_nonexistent_dir")
	SetOinakosDir(tmpDir)
	defer SetOinakosDir("")

	s := LoadSettings()
	if s == nil {
		t.Fatal("LoadSettings should never return nil")
	}
	// Should match defaults
	if s.Font != "medieval" {
		t.Errorf("Expected default font medieval, got %s", s.Font)
	}
}
