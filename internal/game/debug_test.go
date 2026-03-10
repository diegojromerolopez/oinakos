package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDebugMode(t *testing.T) {
	// Use a local directory for testing
	testDir := "test_oinakos"
	SetOinakosDir(testDir)
	defer os.RemoveAll(testDir)

	// Start disabled
	SetDebugMode(false)
	if IsDebugEnabled() {
		t.Error("Debug should be disabled")
	}

	SetDebugMode(true)
	if !IsDebugEnabled() {
		t.Error("Debug should be enabled")
	}

	// Verify file was created
	logPath := filepath.Join(testDir, "debug.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("debug.log file should have been created at %s", logPath)
	}

	DebugLog("Test message: %s", "hello")

	SetDebugMode(false)
	if IsDebugEnabled() {
		t.Error("Debug should be disabled again")
	}

	// Clean up
	os.Remove("oinakos/debug.log")
}
