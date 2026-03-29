package game

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"oinakos/internal/engine"

	"gopkg.in/yaml.v3"
)

func DiscoverFonts(assets fs.FS) []string {
	fonts := []string{}
	entries, err := fs.ReadDir(assets, "assets/fonts")
	if err != nil {
		return []string{"default"}
	}
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(name), ".ttf") {
			stem := name[:len(name)-4]
			fonts = append(fonts, stem)
		}
	}
	fonts = append(fonts, "default")
	return fonts
}

type Settings struct {
	SoundFrequency    string `yaml:"sound_frequency"`
	TalkingFrequency  string `yaml:"talking_frequency"`
	Font           string `yaml:"font"`
	FogOfWar       string `yaml:"fog_of_war"` // none | vision | exploration

	AIProvider        string `yaml:"ai_provider"`        // none | openai | claude | gemini | mistral | huggingface | ollama (local) | ollama | webgpu
	AISimulationMode  bool   `yaml:"ai_simulation_mode"`  // If true, the player character is also AI-controlled
	TimePace          string `yaml:"time_pace"`           // real | double | fast | standard | month | year

	// API Keys
	OpenAIApiKey     string `yaml:"openai_api_key"`
	ClaudeApiKey     string `yaml:"claude_api_key"`
	GeminiApiKey     string `yaml:"gemini_api_key"`
	MistralApiKey    string `yaml:"mistral_api_key"`
	HuggingFaceApiKey string `yaml:"huggingface_api_key"`

	// Models
	OpenAIModel      string `yaml:"openai_model"`
	ClaudeModel      string `yaml:"claude_model"`
	GeminiModel      string `yaml:"gemini_model"`
	MistralModel     string `yaml:"mistral_model"`
	HuggingFaceModel string `yaml:"huggingface_model"`
	OllamaModel      string `yaml:"ollama_model"`
	OllamaLocalModel string `yaml:"ollama_local_model"`
	WebGPUModel      string `yaml:"webgpu_model"`

	// Advanced
	AIBaseURL       string `yaml:"ai_base_url"`

	// Units
	Units string `yaml:"units"` // venburguian | si | imperial
	
	// Adult Mode
	AdultMode bool `yaml:"adult_mode"`

	// Keymap
	Keymap map[string]string `yaml:"keymap"`
}

var FrequencyOptions = []string{"never", "rare", "infrequent", "half the time", "frequent", "always"}
var FontOptions = []string{"medieval", "modern_antiqua", "uncial_antiqua", "glass_antiqua", "kings", "eagle_lake", "default"}
var FogOfWarOptions = []string{"none", "vision", "exploration"}
var AIProviderOptions = []string{"none", "openai", "claude", "gemini", "mistral", "huggingface", "ollama (local)", "ollama (service)"}
var UnitsOptions = []string{"venburguian", "si", "imperial"}

var TimePaceOptions = []string{"real", "double", "fast", "standard", "month", "year"}
var TimePaceLabels = map[string]string{
	"real":     "Real-Time (1:1)",
	"double":   "Double-Time (2x)",
	"fast":     "Fast-Forward (10x)",
	"standard": "Simulation Standard (Standard)",
	"month":    "Month Burst (Month/min)",
	"year":     "Year Burst (Year/min)",
}
var RemappableActions = []struct {
	ID   string
	Name string
}{
	{"move_up", "Move Up"},
	{"move_down", "Move Down"},
	{"move_left", "Move Left"},
	{"move_right", "Move Right"},
	{"attack", "Attack / Interact"},
	{"talk", "Speak to NPC"},
	{"chop", "Chop Wood (requires Axe)"},
	{"dig", "Dig Ground (requires Pike)"},
	{"forage", "Forage / Cook Food"},
	{"punch", "Punch (unarmed attack)"},
	{"slap", "Slap (make them submissive)"},
	{"inventory", "Open Inventory"},
	{"rest", "Rest (Sleep)"},
	{"eat", "Eat (while in inventory)"},
	{"debug", "Toggle Debug Collidables"},
	{"menu", "Pause Menu / Back"},
}

func SetFontOptions(fonts []string) {
	FontOptions = fonts
}

func DefaultSettings() *Settings {
	return &Settings{
		SoundFrequency:   "rare",
		TalkingFrequency: "infrequent",
		Font:             "medieval",
		FogOfWar:         "none",
		AIProvider:       "none",
		AISimulationMode: false,
		TimePace:         "standard",
		Units:            "venburguian",

		OpenAIModel:      "gpt-4o-mini",
		ClaudeModel:      "claude-3-5-sonnet-20240620",
		GeminiModel:      "gemini-1.5-flash",
		MistralModel:     "mistral-small-latest",
		HuggingFaceModel: "meta-llama/Llama-3.1-8B-Instruct",
		OllamaModel:      "llama3",
		OllamaLocalModel: "llama3",
		WebGPUModel:      "tiny-llama-1.1b",
		AdultMode:        true,
		Keymap: map[string]string{
			"move_up":    "W",
			"move_down":  "S",
			"move_left":  "A",
			"move_right": "D",
			"attack":     "SPACE",
			"punch":      "Q",
			"slap":       "B",
			"talk":       "T",
			"chop":       "C",
			"dig":        "V",
			"forage":     "F",
			"inventory":  "I",
			"rest":       "R",
			"eat":        "E",
			"debug":      "TAB",
			"menu":       "ESC",
		},
	}
}

func (s *Settings) GetDefaultModel(provider string) string {
	d := DefaultSettings()
	switch provider {
	case "openai":
		return d.OpenAIModel
	case "claude":
		return d.ClaudeModel
	case "gemini":
		return d.GeminiModel
	case "ollama (local)":
		return d.OllamaLocalModel
	case "ollama (service)":
		return d.OllamaModel
	case "mistral":
		return d.MistralModel
	case "huggingface":
		return d.HuggingFaceModel
	case "webgpu":
		return d.WebGPUModel
	}
	return ""
}

func (s *Settings) GetSoundProbability() float64 {
	return s.getFrequencyProb(s.SoundFrequency)
}

func (s *Settings) GetTalkingProbability() float64 {
	return s.getFrequencyProb(s.TalkingFrequency)
}

func (s *Settings) getFrequencyProb(freq string) float64 {
	switch freq {
	case "never":
		return 0.0
	case "rare":
		return 0.05
	case "infrequent":
		return 0.15
	case "half the time":
		return 0.4
	case "frequent":
		return 0.7
	case "always":
		return 1.0
	default:
		return 0.1
	}
}

func (s *Settings) FormatWeight(w float64) string {
	switch s.Units {
	case "si":
		return fmt.Sprintf("%.2f kg", w)
	case "imperial":
		return fmt.Sprintf("%.2f lb", w*2.20462)
	case "venburguian":
		return fmt.Sprintf("%.2f lib", w/0.329) // Roman Pound (Libra)
	default:
		return fmt.Sprintf("%.2f", w)
	}
}

func (s *Settings) FormatDistance(d float64) string {
	switch s.Units {
	case "si":
		return fmt.Sprintf("%.2f m", d*0.296)
	case "imperial":
		return fmt.Sprintf("%.2f ft", d*0.971)
	case "venburguian":
		if d >= 5000 {
			return fmt.Sprintf("%.1f millia", d/5000.0)
		}
		if d >= 5 {
			return fmt.Sprintf("%.1f passus", d/5.0)
		}
		return fmt.Sprintf("%.1f pedes", d)
	default:
		return fmt.Sprintf("%.2f units", d)
	}
}

var oinakosDirOverride string

func SetOinakosDir(path string) {
	oinakosDirOverride = path
}

func GetOinakosDir() string {
	if oinakosDirOverride != "" {
		// Ensure it exists if overridden
		if _, err := os.Stat(oinakosDirOverride); os.IsNotExist(err) {
			_ = os.MkdirAll(oinakosDirOverride, 0755)
		}
		return oinakosDirOverride
	}

	var baseDir string
	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = "."
		}
		baseDir = filepath.Join(baseDir, "oinakos")
	default: // linux, darwin
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		baseDir = filepath.Join(home, ".oinakos")
	}

	// Create dir if it doesn't exist
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		_ = os.MkdirAll(baseDir, 0755)
	}

	return baseDir
}

func getSettingsPath() string {
	return filepath.Join(GetOinakosDir(), "settings.yml")
}

func LoadSettings() *Settings {
	s := DefaultSettings()
	data, err := loadSettingsData()
	if err != nil {
		return s
	}

	err = yaml.Unmarshal(data, s)
	if err != nil {
		log.Printf("Warning: failed to unmarshal settings: %v", err)
		return DefaultSettings()
	}

	return s
}

func (s *Settings) Save() error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	return saveSettingsData(data)
}

func (s *Settings) GetKey(action string) engine.Key {
	if s.Keymap == nil {
		return DefaultSettings().GetKey(action)
	}
	name, ok := s.Keymap[action]
	if !ok {
		return NameToKey(DefaultSettings().Keymap[action])
	}
	return NameToKey(name)
}

func NameToKey(name string) engine.Key {
	switch strings.ToUpper(name) {
	case "W": return engine.KeyW
	case "A": return engine.KeyA
	case "S": return engine.KeyS
	case "D": return engine.KeyD
	case "UP": return engine.KeyUp
	case "DOWN": return engine.KeyDown
	case "LEFT": return engine.KeyLeft
	case "RIGHT": return engine.KeyRight
	case "SPACE": return engine.KeySpace
	case "ENTER": return engine.KeyEnter
	case "ESC": return engine.KeyEscape
	case "TAB": return engine.KeyTab
	case "I": return engine.KeyI
	case "F": return engine.KeyF
	case "R": return engine.KeyR
	case "C": return engine.KeyC
	case "V": return engine.KeyV
	case "Q": return engine.KeyQ
	case "E": return engine.KeyE
	case "B": return engine.KeyB
	case "T": return engine.KeyT
	case "F12": return engine.KeyF9 // We don't have F12 in engine.Key, so mapping F9 for now or just using F9
	case "F9": return engine.KeyF9
	}
	return engine.KeyW // Fallback
}

func KeyToName(key engine.Key) string {
	switch key {
	case engine.KeyW: return "W"
	case engine.KeyA: return "A"
	case engine.KeyS: return "S"
	case engine.KeyD: return "D"
	case engine.KeyUp: return "UP"
	case engine.KeyDown: return "DOWN"
	case engine.KeyLeft: return "LEFT"
	case engine.KeyRight: return "RIGHT"
	case engine.KeySpace: return "SPACE"
	case engine.KeyEnter: return "ENTER"
	case engine.KeyEscape: return "ESC"
	case engine.KeyTab: return "TAB"
	case engine.KeyI: return "I"
	case engine.KeyF: return "F"
	case engine.KeyR: return "R"
	case engine.KeyC: return "C"
	case engine.KeyV: return "V"
	case engine.KeyQ: return "Q"
	case engine.KeyE: return "E"
	case engine.KeyB: return "B"
	case engine.KeyT: return "T"
	case engine.KeyF9: return "F9"
	}
	return "UNKNOWN"
}
