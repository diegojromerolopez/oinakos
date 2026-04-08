package game

import (
	"testing"
)

func TestNewItemInstance(t *testing.T) {
	config := &ObjectConfig{
		ID:         "sword",
		Resistance: 10,
	}
	
	it := NewItemInstance("sword_1", config, 5, 5)
	
	if it.ID != "sword_1" {
		t.Errorf("ID: got %s, want sword_1", it.ID)
	}
	if it.Resistance != 10 {
		t.Errorf("Resistance: got %d, want 10", it.Resistance)
	}
	if it.X != 5 || it.Y != 5 {
		t.Errorf("Position: got (%f,%f), want (5,5)", it.X, it.Y)
	}
	if !it.Pickable {
		t.Error("expected item to be pickable")
	}
}

func TestItemInstance_GetFootprint(t *testing.T) {
	config := &ObjectConfig{
		Footprint: []FootprintPoint{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 0, Y: 1},
		},
	}
	
	it := NewItemInstance("item", config, 10, 10)
	fp := it.GetFootprint()
	
	if len(fp.Points) != 3 {
		t.Errorf("expected 3 points, got %d", len(fp.Points))
	}
	
	// Transformed position
	if fp.Points[0].X != 10 || fp.Points[0].Y != 10 {
		t.Errorf("expected point[0] at (10,10), got (%f,%f)", fp.Points[0].X, fp.Points[0].Y)
	}
	
	// Default footprint
	it2 := NewItemInstance("item2", nil, 0, 0)
	fp2 := it2.GetFootprint()
	if len(fp2.Points) != 4 {
		t.Errorf("expected 4 points in default footprint, got %d", len(fp2.Points))
	}
}
