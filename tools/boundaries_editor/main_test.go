package main

import (
	"os"
	"path/filepath"
	"testing"

	"oinakos/internal/engine"
	"oinakos/internal/game"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

func newTestViewer(entities []*EditorEntity) *Viewer {
	g := &engine.MockGraphics{}
	return NewViewer(entities, g, defaultScreenWidth, defaultScreenHeight)
}

func makeFootprint(points ...game.FootprintPoint) *[]game.FootprintPoint {
	fp := make([]game.FootprintPoint, len(points))
	copy(fp, points)
	return &fp
}

func squareFP() *[]game.FootprintPoint {
	return makeFootprint(
		game.FootprintPoint{X: -0.5, Y: -0.5},
		game.FootprintPoint{X: 0.5, Y: -0.5},
		game.FootprintPoint{X: 0.5, Y: 0.5},
		game.FootprintPoint{X: -0.5, Y: 0.5},
	)
}

func newEntity(id string, fp *[]game.FootprintPoint) *EditorEntity {
	return &EditorEntity{
		ID:        id,
		Type:      "Obstacle",
		Footprint: fp,
		DrawMain:  func(_ engine.Image, _ engine.Graphics, _, _ float64) {},
	}
}

// ─── EditorEntity.GetFootprint ────────────────────────────────────────────────

func TestGetFootprintWithPoints(t *testing.T) {
	fp := squareFP()
	ee := newEntity("tree", fp)

	poly := ee.GetFootprint()
	if len(poly.Points) != 4 {
		t.Fatalf("want 4 points, got %d", len(poly.Points))
	}
	if poly.Points[0].X != -0.5 || poly.Points[0].Y != -0.5 {
		t.Errorf("first point: want (-0.5,-0.5), got (%.2f,%.2f)", poly.Points[0].X, poly.Points[0].Y)
	}
}

func TestGetFootprintNilReturnsDefault(t *testing.T) {
	ee := &EditorEntity{
		ID:        "no_fp",
		Type:      "Obstacle",
		Footprint: nil,
		DrawMain:  func(_ engine.Image, _ engine.Graphics, _, _ float64) {},
	}
	poly := ee.GetFootprint()
	if len(poly.Points) != 4 {
		t.Fatalf("default footprint should have 4 points, got %d", len(poly.Points))
	}
}

func TestGetFootprintEmptyReturnsDefault(t *testing.T) {
	empty := makeFootprint()
	ee := newEntity("empty_fp", empty)

	poly := ee.GetFootprint()
	if len(poly.Points) != 4 {
		t.Fatalf("empty footprint should return default 4 points, got %d", len(poly.Points))
	}
}

func TestGetFootprintCoordinatePassthrough(t *testing.T) {
	fp := makeFootprint(
		game.FootprintPoint{X: 1.0, Y: 2.0},
		game.FootprintPoint{X: 3.0, Y: 4.0},
		game.FootprintPoint{X: 5.0, Y: 6.0},
	)
	ee := newEntity("coords", fp)
	poly := ee.GetFootprint()

	for i, want := range []engine.Point{{X: 1, Y: 2}, {X: 3, Y: 4}, {X: 5, Y: 6}} {
		if poly.Points[i].X != want.X || poly.Points[i].Y != want.Y {
			t.Errorf("point %d: want (%.1f,%.1f), got (%.1f,%.1f)", i, want.X, want.Y, poly.Points[i].X, poly.Points[i].Y)
		}
	}
}

// ─── Viewer.addPoint ──────────────────────────────────────────────────────────

func TestAddPointIncreasesCount(t *testing.T) {
	fp := squareFP()
	ee := newEntity("tree", fp)
	v := newTestViewer([]*EditorEntity{ee})

	before := len(*ee.Footprint)
	v.addPoint(ee)
	after := len(*ee.Footprint)

	if after != before+1 {
		t.Errorf("addPoint: want %d points, got %d", before+1, after)
	}
}

func TestAddPointOffsetFromLast(t *testing.T) {
	fp := squareFP()
	ee := newEntity("tree", fp)
	v := newTestViewer([]*EditorEntity{ee})

	last := (*ee.Footprint)[len(*ee.Footprint)-1]
	v.addPoint(ee)
	newPt := (*ee.Footprint)[len(*ee.Footprint)-1]

	if newPt.X != last.X+0.5 || newPt.Y != last.Y+0.5 {
		t.Errorf("new point offset: want (%.1f,%.1f), got (%.1f,%.1f)",
			last.X+0.5, last.Y+0.5, newPt.X, newPt.Y)
	}
}

func TestAddPointOnEmptyFootprint(t *testing.T) {
	empty := makeFootprint()
	ee := newEntity("empty", empty)
	v := newTestViewer([]*EditorEntity{ee})

	v.addPoint(ee)

	if len(*ee.Footprint) != 1 {
		t.Fatalf("want 1 point after add on empty, got %d", len(*ee.Footprint))
	}
	// Should be zero-value when list was empty
	pt := (*ee.Footprint)[0]
	if pt.X != 0 || pt.Y != 0 {
		t.Errorf("first point on empty footprint: want (0,0), got (%.2f,%.2f)", pt.X, pt.Y)
	}
}

// ─── Viewer.removePoint ───────────────────────────────────────────────────────

func TestRemovePointDecreasesCount(t *testing.T) {
	fp := squareFP() // 4 points
	ee := newEntity("tree", fp)
	v := newTestViewer([]*EditorEntity{ee})

	v.removePoint(ee, 0)

	if len(*ee.Footprint) != 3 {
		t.Errorf("after remove: want 3 points, got %d", len(*ee.Footprint))
	}
}

func TestRemovePointMiddle(t *testing.T) {
	fp := makeFootprint(
		game.FootprintPoint{X: 1, Y: 0},
		game.FootprintPoint{X: 2, Y: 0},
		game.FootprintPoint{X: 3, Y: 0},
		game.FootprintPoint{X: 4, Y: 0},
	)
	ee := newEntity("pts", fp)
	v := newTestViewer([]*EditorEntity{ee})

	v.removePoint(ee, 1) // remove X=2

	pts := *ee.Footprint
	if len(pts) != 3 {
		t.Fatalf("want 3 points, got %d", len(pts))
	}
	if pts[0].X != 1 || pts[1].X != 3 || pts[2].X != 4 {
		t.Errorf("remaining points wrong: %v", pts)
	}
}

func TestRemovePointRefusesBelow3(t *testing.T) {
	fp := makeFootprint(
		game.FootprintPoint{X: 0, Y: 0},
		game.FootprintPoint{X: 1, Y: 0},
		game.FootprintPoint{X: 0, Y: 1},
	)
	ee := newEntity("tri", fp)
	v := newTestViewer([]*EditorEntity{ee})

	v.removePoint(ee, 0) // should be refused

	if len(*ee.Footprint) != 3 {
		t.Errorf("removePoint should not go below 3 vertices, got %d", len(*ee.Footprint))
	}
}

func TestRemovePointLast(t *testing.T) {
	fp := squareFP()
	ee := newEntity("tree", fp)
	v := newTestViewer([]*EditorEntity{ee})

	last := (*fp)[3]
	v.removePoint(ee, 3)

	pts := *ee.Footprint
	if len(pts) != 3 {
		t.Fatalf("want 3, got %d", len(pts))
	}
	for _, p := range pts {
		if p.X == last.X && p.Y == last.Y {
			t.Error("last point should have been removed")
		}
	}
}

// ─── Viewer initialisation ────────────────────────────────────────────────────

func TestNewViewerDefaults(t *testing.T) {
	v := newTestViewer(nil)

	if v.selectedIndex != 0 {
		t.Errorf("selectedIndex: want 0, got %d", v.selectedIndex)
	}
	if v.draggingIdx != -1 {
		t.Errorf("draggingIdx: want -1, got %d", v.draggingIdx)
	}
	if v.hoverIdx != -1 {
		t.Errorf("hoverIdx: want -1, got %d", v.hoverIdx)
	}
	if v.width != defaultScreenWidth || v.height != defaultScreenHeight {
		t.Errorf("dimensions: want %dx%d, got %dx%d", defaultScreenWidth, defaultScreenHeight, v.width, v.height)
	}
}

func TestNewViewerSetsEntities(t *testing.T) {
	entities := []*EditorEntity{
		newEntity("a", squareFP()),
		newEntity("b", squareFP()),
	}
	v := newTestViewer(entities)

	if len(v.entities) != 2 {
		t.Errorf("want 2 entities, got %d", len(v.entities))
	}
}

// ─── saveToYAML ───────────────────────────────────────────────────────────────

func TestSaveToYAMLNoPathIsNoop(t *testing.T) {
	ee := newEntity("noop", squareFP())
	ee.YamlPath = ""
	v := newTestViewer([]*EditorEntity{ee})
	// Should not panic
	v.saveToYAML(ee)
}

func TestSaveToYAMLWritesFootprint(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "obstacle.yaml")

	// Write a minimal YAML with an existing footprint field
	initial := `id: test_rock
footprint:
  - x: -0.5
    y: -0.5
  - x: 0.5
    y: -0.5
  - x: 0.5
    y: 0.5
  - x: -0.5
    y: 0.5
`
	if err := os.WriteFile(yamlPath, []byte(initial), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	fp := makeFootprint(
		game.FootprintPoint{X: -1.0, Y: -1.0},
		game.FootprintPoint{X: 1.0, Y: -1.0},
		game.FootprintPoint{X: 1.0, Y: 1.0},
	)
	ee := newEntity("test_rock", fp)
	ee.YamlPath = yamlPath

	v := newTestViewer([]*EditorEntity{ee})
	v.saveToYAML(ee)

	// Re-read and check the new footprint is present
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	content := string(data)
	if !contains(content, "-1") {
		t.Errorf("saved YAML doesn't contain updated footprint value -1:\n%s", content)
	}
}

func TestSaveToYAMLAddsFootprintIfMissing(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "nofp.yaml")

	// YAML without footprint key
	if err := os.WriteFile(yamlPath, []byte("id: nofp\nhealth: 10\n"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	fp := makeFootprint(
		game.FootprintPoint{X: -0.25, Y: -0.25},
		game.FootprintPoint{X: 0.25, Y: -0.25},
		game.FootprintPoint{X: 0.25, Y: 0.25},
	)
	ee := newEntity("nofp", fp)
	ee.YamlPath = yamlPath

	v := newTestViewer([]*EditorEntity{ee})
	v.saveToYAML(ee)

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !contains(string(data), "footprint") {
		t.Errorf("YAML should now contain 'footprint' key:\n%s", string(data))
	}
}

func TestSaveToYAMLBadPathIsNoop(t *testing.T) {
	ee := newEntity("ghost", squareFP())
	ee.YamlPath = "/nonexistent/path/obstacle.yaml"
	v := newTestViewer([]*EditorEntity{ee})
	// Should log but not panic
	v.saveToYAML(ee)
}

// ─── containsID ───────────────────────────────────────────────────────────────

func TestContainsIDMatch(t *testing.T) {
	yaml := []byte("id: tree_oak\nhealth: 100\n")
	if !containsID(yaml, "tree_oak") {
		t.Error("containsID should return true for matching ID")
	}
}

func TestContainsIDNoMatch(t *testing.T) {
	yaml := []byte("id: rock1\nhealth: 100\n")
	if containsID(yaml, "tree_oak") {
		t.Error("containsID should return false for non-matching ID")
	}
}

func TestContainsIDPartialNoMatch(t *testing.T) {
	// "tree" is a prefix of "tree_oak" but shouldn't match "tree_oak" when looking for "tree"
	yaml := []byte("id: tree_oak\n")
	if containsID(yaml, "tree") {
		t.Error("containsID should not match partial IDs")
	}
}

func TestContainsIDInvalidYAML(t *testing.T) {
	// Malformed YAML should not panic, just return false
	if containsID([]byte("{{{invalid"), "anything") {
		t.Error("invalid YAML should return false")
	}
}

func TestContainsIDEmptyData(t *testing.T) {
	if containsID([]byte(""), "something") {
		t.Error("empty YAML should return false")
	}
}

// ─── Polygon round-trip (GetFootprint ↔ addPoint) ─────────────────────────────

func TestFootprintRoundTrip(t *testing.T) {
	fp := squareFP()
	ee := newEntity("rt", fp)
	v := newTestViewer([]*EditorEntity{ee})

	original := ee.GetFootprint()
	v.addPoint(ee)
	v.removePoint(ee, len(*ee.Footprint)-1) // remove what was just added
	result := ee.GetFootprint()

	if len(original.Points) != len(result.Points) {
		t.Fatalf("round-trip: length changed from %d to %d", len(original.Points), len(result.Points))
	}
	for i := range original.Points {
		if original.Points[i].X != result.Points[i].X || original.Points[i].Y != result.Points[i].Y {
			t.Errorf("point %d changed: (%.2f,%.2f) → (%.2f,%.2f)",
				i, original.Points[i].X, original.Points[i].Y,
				result.Points[i].X, result.Points[i].Y)
		}
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
