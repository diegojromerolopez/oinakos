package game

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
}

var FrequencyOptions = []string{"never", "rare", "infrequent", "half the time", "frequent", "always"}
var FontOptions = []string{"medieval", "modern_antiqua", "uncial_antiqua", "glass_antiqua", "kings", "eagle_lake", "default"}
var FogOfWarOptions = []string{"none", "vision", "exploration"}
var AIProviderOptions = []string{"none", "openai", "claude", "gemini", "mistral", "huggingface", "ollama (local)", "ollama (service)"}
var UnitsOptions = []string{"venburguian", "si", "imperial"}

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
		Units:            "venburguian",

		OpenAIModel:      "gpt-4o-mini",
		ClaudeModel:      "claude-3-5-sonnet-20240620",
		GeminiModel:      "gemini-1.5-flash",
		MistralModel:     "mistral-small-latest",
		HuggingFaceModel: "meta-llama/Llama-3.1-8B-Instruct",
		OllamaModel:      "llama3",
		OllamaLocalModel: "llama3",
		WebGPUModel:      "tiny-llama-1.1b",
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
