package game

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestSettings_Advanced(t *testing.T) {
	tmpDir := t.TempDir()
	SetOinakosDir(tmpDir)
	defer SetOinakosDir("")
	
	// Top-level functions
	DiscoverFonts(fstest.MapFS{})
	
	oldFonts := FontOptions
	SetFontOptions([]string{"OpenSans"})
	if FontOptions[0] != "OpenSans" { t.Error("Failed to set font options") }
	FontOptions = oldFonts // Restore
	
	// Method
	s := DefaultSettings()
	err := s.Save()
	if err != nil { t.Errorf("Failed to save settings: %v", err) }
	
	// Check file exists
	setPath := filepath.Join(tmpDir, "settings.yml")
	if _, err := os.Stat(setPath); os.IsNotExist(err) {
		t.Error("Settings file not created")
	}
}
