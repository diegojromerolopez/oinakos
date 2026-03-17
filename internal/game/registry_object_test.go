package game

import (
	"os"
	"testing"
)

func TestObjectRegistryLoadAll(t *testing.T) {
	// Go tests run in the package directory, so we need to go up to the repo root
	assets := os.DirFS("../..")
	reg := NewObjectRegistry()
	err := reg.LoadAll(assets)
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	t.Logf("Loaded %d object IDs", len(reg.IDs))
	if len(reg.IDs) == 0 {
		t.Errorf("Registry is empty!")
	}

	foundIronSword := false
	for _, id := range reg.IDs {
		obj := reg.Objects[id]
		if obj == nil {
			t.Errorf("Object %s not found in map!", id)
			continue
		}
		if obj.ID != id {
			t.Errorf("Mismatch for %s: map entry has ID %s", id, obj.ID)
		}
		if id == "iron_sword" {
			foundIronSword = true
			if obj.Name != "Iron Sword" {
				t.Errorf("Expected name Iron Sword, got %s", obj.Name)
			}
			if obj.Weight != 4.0 {
				t.Errorf("Expected weight 4.0, got %f", obj.Weight)
			}
			if obj.Type != "weapon" {
				t.Errorf("Expected type weapon, got %s", obj.Type)
			}
			if obj.Combat == nil {
				t.Errorf("Expected combat section, got nil")
			} else {
				if obj.Combat.Name != "Iron Sword" {
					t.Errorf("Expected combat name Iron Sword, got %s", obj.Combat.Name)
				}
			}
		}
		t.Logf(" - Found ID: %s (Name: %s)", id, obj.Name)
	}

	if !foundIronSword {
		t.Errorf("iron_sword not found in registry")
	}
}
