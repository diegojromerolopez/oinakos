package game

import (
	"testing"
	"oinakos/internal/engine"
	"testing/fstest"
)

func TestSettings_Keymap(t *testing.T) {
	s := DefaultSettings()
	s.Keymap = map[string]string{
		"move_up": "UP",
		"custom":  "SPACE",
	}

	tests := []struct {
		action string
		want   engine.Key
	}{
		{"move_up", engine.KeyUp},
		{"custom", engine.KeySpace},
		{"move_down", engine.KeyS}, // From defaults
		{"nonexistent", engine.KeyW}, // Fallback from NameToKey defaulting to W
	}

	for _, tt := range tests {
		if got := s.GetKey(tt.action); got != tt.want {
			t.Errorf("GetKey(%q) = %v, want %v", tt.action, got, tt.want)
		}
	}
}

func TestNameToKey(t *testing.T) {
	tests := []struct {
		name string
		want engine.Key
	}{
		{"W", engine.KeyW},
		{"A", engine.KeyA},
		{"S", engine.KeyS},
		{"D", engine.KeyD},
		{"UP", engine.KeyUp},
		{"DOWN", engine.KeyDown},
		{"LEFT", engine.KeyLeft},
		{"RIGHT", engine.KeyRight},
		{"SPACE", engine.KeySpace},
		{"ENTER", engine.KeyEnter},
		{"ESC", engine.KeyEscape},
		{"TAB", engine.KeyTab},
		{"I", engine.KeyI},
		{"F", engine.KeyF},
		{"R", engine.KeyR},
		{"C", engine.KeyC},
		{"V", engine.KeyV},
		{"Q", engine.KeyQ},
		{"E", engine.KeyE},
		{"B", engine.KeyB},
		{"T", engine.KeyT},
		{"F9", engine.KeyF9},
		{"F12", engine.KeyF9},
		{"UNKNOWN", engine.KeyW}, // Fallback
	}

	for _, tt := range tests {
		if got := NameToKey(tt.name); got != tt.want {
			t.Errorf("NameToKey(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestKeyToName(t *testing.T) {
	tests := []struct {
		key  engine.Key
		want string
	}{
		{engine.KeyW, "W"},
		{engine.KeyA, "A"},
		{engine.KeyS, "S"},
		{engine.KeyD, "D"},
		{engine.KeyUp, "UP"},
		{engine.KeyDown, "DOWN"},
		{engine.KeyLeft, "LEFT"},
		{engine.KeyRight, "RIGHT"},
		{engine.KeySpace, "SPACE"},
		{engine.KeyEnter, "ENTER"},
		{engine.KeyEscape, "ESC"},
		{engine.KeyTab, "TAB"},
		{engine.KeyI, "I"},
		{engine.KeyF, "F"},
		{engine.KeyR, "R"},
		{engine.KeyC, "C"},
		{engine.KeyV, "V"},
		{engine.KeyQ, "Q"},
		{engine.KeyE, "E"},
		{engine.KeyB, "B"},
		{engine.KeyT, "T"},
		{engine.KeyF9, "F9"},
		{engine.Key(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := KeyToName(tt.key); got != tt.want {
			t.Errorf("KeyToName(%v) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestDiscoverFonts(t *testing.T) {
	mockFS := fstest.MapFS{
		"assets/fonts/medieval.ttf":      &fstest.MapFile{Data: []byte("font")},
		"assets/fonts/modern.ttf":        &fstest.MapFile{Data: []byte("font")},
		"assets/fonts/readme.txt":         &fstest.MapFile{Data: []byte("not a font")},
		"assets/fonts/subdir/hidden.ttf": &fstest.MapFile{Data: []byte("hidden")},
	}

	fonts := DiscoverFonts(mockFS)
	
	// Should contain medieval, modern, and default
	expected := map[string]bool{
		"medieval": true,
		"modern":   true,
		"default":  true,
	}

	if len(fonts) != 3 {
		t.Errorf("expected 3 fonts, got %v: %v", len(fonts), fonts)
	}

	for _, f := range fonts {
		if !expected[f] {
			t.Errorf("unexpected font discovered: %s", f)
		}
	}
}

func TestDiscoverFonts_Error(t *testing.T) {
	mockFS := fstest.MapFS{} // Empty FS, no assets/fonts dir
	fonts := DiscoverFonts(mockFS)
	
	if len(fonts) != 1 || fonts[0] != "default" {
		t.Errorf("expected only default font on error, got %v", fonts)
	}
}

func TestSetFontOptions(t *testing.T) {
	oldOptions := FontOptions
	defer func() { FontOptions = oldOptions }()

	newOptions := []string{"test_font", "another_font"}
	SetFontOptions(newOptions)

	if len(FontOptions) != 2 || FontOptions[0] != "test_font" {
		t.Errorf("expected test_font, got %v", FontOptions)
	}
}
