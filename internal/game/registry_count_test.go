package game

import (
	"os"
	"testing"
)

func TestCharacterRegistryCount(t *testing.T) {
	// Go tests run in the package directory, so we need to go up to the repo root
	assets := os.DirFS("../..")
	reg := NewCharacterRegistry()
	err := reg.LoadAll(assets)
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	t.Logf("Loaded %d character IDs", len(reg.IDs))
	for _, id := range reg.IDs {
		char := reg.Characters[id]
		if char == nil {
			t.Errorf("Character %s not found in map!", id)
			continue
		}
		if char.ID != id {
			t.Errorf("Mismatch for %s: map entry has ID %s", id, char.ID)
		}
		t.Logf(" - Found ID: %s (Map ID: %s)", id, char.ID)
	}

	if len(reg.IDs) == 0 {
		t.Errorf("Registry is empty!")
	}
}
