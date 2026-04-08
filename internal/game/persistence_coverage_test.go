package game

import (
	"testing"
)

func TestGame_PersistenceCoverage(t *testing.T) {
	g := setupTestGame()
	
	// 1. Serialize
	data, err := g.serialize()
	if err != nil { t.Fatalf("Failed to serialize: %v", err) }
	if len(data) == 0 { t.Error("Serialized data is empty") }
	
	// 2. Mock save logic (avoids actual file writes if we just test serialize)
	// But we can test Save to a temp file
	tmpPath := "test_save.oinakos.yaml"
	err = g.Save(tmpPath)
	if err == nil {
		// Clean up
		// os.Remove(tmpPath)
	}
}
