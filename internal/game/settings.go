package game

import (
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
	SoundFrequency string `yaml:"sound_frequency"`
	Font           string `yaml:"font"`
}

var FrequencyOptions = []string{"never", "rare", "infrequent", "half the time", "frequent", "always"}
var FontOptions = []string{"medieval", "modern_antiqua", "uncial_antiqua", "glass_antiqua", "kings", "eagle_lake", "default"}

func SetFontOptions(fonts []string) {
	FontOptions = fonts
}

func DefaultSettings() *Settings {
	return &Settings{
		SoundFrequency: "rare",
		Font:           "medieval",
	}
}

func (s *Settings) GetSoundProbability() float64 {
	switch s.SoundFrequency {
	case "never":
		return 0.0
	case "rare":
		return 0.1
	case "infrequent":
		return 0.2
	case "half the time":
		return 0.5
	case "frequent":
		return 0.7
	case "always":
		return 1.0
	default:
		return 0.1
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
	filename := getSettingsPath()
	data, err := os.ReadFile(filename)
	if err != nil {
		return DefaultSettings()
	}

	var s Settings
	err = yaml.Unmarshal(data, &s)
	if err != nil {
		log.Printf("Warning: failed to unmarshal settings: %v", err)
		return DefaultSettings()
	}

	return &s
}

func (s *Settings) Save() error {
	filename := getSettingsPath()
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	log.Printf("Settings saved to: %s", filename)
	return os.WriteFile(filename, data, 0644)
}
