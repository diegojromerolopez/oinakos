package game

import (
	"fmt"
	"oinakos/internal/engine"
	"testing"
)

func TestMapType_SeedMinerals(t *testing.T) {
	m := &MapType{
		MapWidth:  100,
		MapHeight: 100,
	}
	
	m.SeedMinerals(42)
	
	if len(m.MineralMap) == 0 {
		t.Error("expected minerals to be seeded")
	}
	
	// Check if seeded minerals are within bounds
	for key := range m.MineralMap {
		var x, y int
		fmt.Sscanf(key, "%d,%d", &x, &y)
		if float64(x) < -55 || float64(x) > 55 || float64(y) < -55 || float64(y) > 55 {
			// A bit of leeway for radius
			t.Errorf("mineral at %d,%d is way out of bounds", x, y)
		}
	}
}

func TestMapType_GetTileAt(t *testing.T) {
	m := &MapType{
		FloorTile: "grass.png",
		FloorZones: []*FloorZone{
			{
				Tile:     "water.png",
				Priority: 10,
				Perimeter: []FootprintPoint{
					{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 5}, {X: 0, Y: 5},
				},
			},
			{
				Tile:     "sand.png",
				Priority: 5,
				Perimeter: []FootprintPoint{
					{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10},
				},
			},
		},
	}
	
	// Test default tile
	if got := m.GetTileAt(20, 20); got != "grass.png" {
		t.Errorf("expected grass.png, got %s", got)
	}
	
	// Test higher priority zone
	if got := m.GetTileAt(2, 2); got != "water.png" {
		t.Errorf("expected water.png, got %s", got)
	}
	
	// Test lower priority zone
	if got := m.GetTileAt(7, 7); got != "sand.png" {
		t.Errorf("expected sand.png, got %s", got)
	}
}

func TestMapType_GetElevationAt(t *testing.T) {
	m := &MapType{
		Heightmap: map[string]float64{
			"0,0": 0.0,
			"1,0": 10.0,
			"0,1": 10.0,
			"1,1": 20.0,
		},
	}
	
	// Test exact corners
	if got := m.GetElevationAt(0, 0); got != 0.0 {
		t.Errorf("expected 0.0 at (0,0), got %f", got)
	}
	if got := m.GetElevationAt(1, 1); got != 20.0 {
		t.Errorf("expected 20.0 at (1,1), got %f", got)
	}
	
	// Test interpolation
	if got := m.GetElevationAt(0.5, 0.5); got != 10.0 {
		t.Errorf("expected 10.0 at (0.5,0.5), got %f", got)
	}
	if got := m.GetElevationAt(0.5, 0); got != 5.0 {
		t.Errorf("expected 5.0 at (0.5,0), got %f", got)
	}
}

func TestMapType_Dig(t *testing.T) {
	m := &MapType{
		Heightmap: map[string]float64{
			"0,0": 10.0,
			"1,0": 10.0,
			"0,1": 10.0,
			"1,1": 10.0,
		},
	}
	
	m.Dig(0.5, 0.5, 2.0)
	
	// Check that height decreased
	z := m.GetElevationAt(0.5, 0.5)
	if z >= 10.0 {
		t.Errorf("expected elevation to decrease after digging, got %f", z)
	}
	
	if m.Heightmap["0,0"] >= 10.0 || m.Heightmap["1,1"] >= 10.0 {
		t.Error("expected corner heights in heightmap to decrease")
	}
}

func TestMapType_GetElevationAt_WithZones(t *testing.T) {
	m := &MapType{
		HeightZones: []*HeightZone{
			{
				Polygon: engine.Polygon{Points: []engine.Point{
					{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10},
				}},
				BaseZ:    5.0,
				Priority: 1,
			},
		},
	}
	
	// Test zone fallback when heightmap is empty
	if got := m.GetElevationAt(5, 5); got != 5.0 {
		t.Errorf("expected 5.0 from zone, got %f", got)
	}
	
	// Test out of zone
	if got := m.GetElevationAt(15, 15); got != 0.0 {
		t.Errorf("expected 0.0 outside zones, got %f", got)
	}
}
