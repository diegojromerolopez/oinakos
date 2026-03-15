package game

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// debugEnabled is the global state for debug logging.
var debugEnabled bool
var debugLogger *log.Logger
var debugFile *os.File

// SetDebugMode toggles the global debug state and initializes/closes the debug log file.
func SetDebugMode(enabled bool) {
	debugEnabled = enabled
	if enabled {
		if debugLogger == nil {
			initDebugFile()
		}
		DebugLog("Debug mode enabled")
	} else {
		if debugLogger != nil {
			DebugLog("Debug mode disabled")
			// We keep the file open until the process ends or we want to rotate,
			// but for simplicity we'll just keep the logger instance.
		}
	}
}

func initDebugFile() {
	oinakosDir := GetOinakosDir()
	if err := os.MkdirAll(oinakosDir, 0755); err != nil {
		log.Printf("Warning: failed to create oinakos directory for debug log: %v", err)
		return
	}

	logPath := filepath.Join(oinakosDir, "debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: failed to open debug log file: %v", err)
		return
	}
	debugFile = f
	debugLogger = log.New(f, "", log.LstdFlags)
}

// DebugLog prints a formatted message to the standard logger and the debug file if debug mode is enabled.
func DebugLog(format string, v ...any) {
	if debugEnabled {
		msg := fmt.Sprintf(format, v...)
		// Also print to standard log for convenience in dev console
		log.Printf("[DEBUG] %s", msg)
		if debugLogger != nil {
			debugLogger.Printf("[DEBUG] %s", msg)
		}
	}
}

// IsDebugEnabled returns the current debug state.
func IsDebugEnabled() bool {
	return debugEnabled
}
