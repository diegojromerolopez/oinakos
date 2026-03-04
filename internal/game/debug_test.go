package game

import (
	"os"
	"testing"
)

func TestDebugMode(t *testing.T) {
	// Start disabled
	SetDebugMode(false)
	if IsDebugEnabled() {
		t.Error("Debug should be disabled")
	}

	SetDebugMode(true)
	if !IsDebugEnabled() {
		t.Error("Debug should be enabled")
	}

	// Verify file was created (if native)
	if _, err := os.Stat("oinakos/debug.log"); os.IsNotExist(err) {
		t.Error("debug.log file should have been created")
	}

	DebugLog("Test message: %s", "hello")

	SetDebugMode(false)
	if IsDebugEnabled() {
		t.Error("Debug should be disabled again")
	}

	// Clean up
	os.Remove("oinakos/debug.log")
}
