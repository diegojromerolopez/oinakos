package game

import (
	"testing"
	"oinakos/internal/engine"
)

func TestGame_CoreBasics(t *testing.T) {
	g := &Game{
		width: 800,
		height: 600,
	}

	// Test Layout
	sw, sh := g.Layout(1000, 1000)
	if sw != 800 || sh != 600 {
		t.Errorf("Layout returned %dx%d, want 800x600", sw, sh)
	}

	// Test SilhouetteBuffer
	sb := g.GetSilhouetteBuffer()
	if sb != nil {
		t.Error("expected nil silhouette buffer initially")
	}
	
	mockImg := &engine.MockImage{}
	g.silhouetteBuffer = mockImg
	if g.GetSilhouetteBuffer() != mockImg {
		t.Error("failed to retrieve silhouette buffer")
	}

	// Test Font Callback
	var updatedFont string
	g.SetOnFontUpdate(func(font string) {
		updatedFont = font
	})
	
	g.settings = &Settings{Font: "test-font"}
	g.UpdateFont()
	if updatedFont != "test-font" {
		t.Errorf("expected font 'test-font', got %q", updatedFont)
	}
}

func TestGame_GetContext(t *testing.T) {
	reg := NewRegistryContainer()
	mockInput := &engine.MockInput{}
	mockAudio := &MockAudioManager{}
	setts := DefaultSettings()
	
	g := &Game{
		World:      NewWorld(),
		input:      mockInput,
		audio:      mockAudio,
		Registries: reg,
		settings:   setts,
		aiManager:  &AIManager{},
	}
	g.World.State.Weather = WeatherRain
	g.World.State.Intensity = 0.8

	ctx := g.GetContext()
	if ctx.World != g.World { t.Error("mismatched world") }
	if ctx.Input != mockInput { t.Error("mismatched input") }
	if ctx.Audio != mockAudio { t.Error("mismatched audio") }
	if ctx.Registries != reg { t.Error("mismatched registries") }
	if ctx.Settings != setts { t.Error("mismatched settings") }
	if ctx.Weather != WeatherRain { t.Error("mismatched weather") }
	if ctx.Intensity != 0.8 { t.Error("mismatched intensity") }
}
