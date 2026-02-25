package game

import (
	"embed"
	"oinakos/internal/engine"
	"os"
	"strings"
	"testing"
)

func TestAssetLoading(t *testing.T) {
	var assets embed.FS

	// Test engine.LoadSprite with nil assets and non-existent file (should fallback to os and fail/log)
	img := engine.LoadSprite(assets, "non_existent.png", true)
	if img != nil {
		t.Error("Expected nil image for non-existent file")
	}

}

func TestArchetypeLoadingErrorPaths(t *testing.T) {
	var assets embed.FS
	r := NewArchetypeRegistry()
	err := r.LoadAll(assets)
	if err != nil && !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("LoadAll(nil) should not error, got %v", err)
	}
}

func TestMapTypeLoadingErrorPaths(t *testing.T) {
	var assets embed.FS
	r := NewMapTypeRegistry()
	err := r.LoadAll(assets)
	if err != nil && !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("LoadAll(nil) should not error, got %v", err)
	}
}

func TestObstacleLoadingErrorPaths(t *testing.T) {
	var assets embed.FS
	r := NewObstacleRegistry()
	err := r.LoadAll(assets)
	if err != nil && !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("LoadAll(nil) should error or handle missing dir gracefully, got %v", err)
	}
}
