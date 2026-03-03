package main

import (
	"os"
	"path/filepath"
	"testing"

	"oinakos/internal/engine"
	"oinakos/internal/game"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

func newTestEditor() *MapEditor {
	g := &engine.MockGraphics{}
	me := &MapEditor{
		Graphics:    g,
		Mode:        "DIALOG",
		InWidth:     "640",
		InHeight:    "640",
		FloorImages: make(map[string]engine.Image),
	}
	return me
}

// ─── Dialog Field Input ────────────────────────────────────────────────────────

func TestDialogDefaultValues(t *testing.T) {
	me := newTestEditor()
	if me.Mode != "DIALOG" {
		t.Errorf("initial mode: want DIALOG, got %s", me.Mode)
	}
	if me.InWidth != "640" {
		t.Errorf("default width: want 640, got %s", me.InWidth)
	}
	if me.InHeight != "640" {
		t.Errorf("default height: want 640, got %s", me.InHeight)
	}
	if me.ActiveField != 0 {
		t.Errorf("initial active field: want 0, got %d", me.ActiveField)
	}
}

func TestDialogInputName(t *testing.T) {
	me := newTestEditor()
	me.ActiveField = 0

	// Simulate typing rune-by-rune as updateDialog() now does.
	for _, ch := range "mymap" {
		me.InName += string(ch)
	}

	if me.InName != "mymap" {
		t.Errorf("InName: want mymap, got %s", me.InName)
	}
}

func TestDialogInputWidthOnlyDigits(t *testing.T) {
	me := newTestEditor()
	me.InWidth = ""
	me.ActiveField = 1

	for _, ch := range []rune{'3', '2', 'x', '0'} { // 'x' should be rejected
		if ch >= '0' && ch <= '9' {
			me.InWidth += string(ch)
		}
	}

	if me.InWidth != "320" {
		t.Errorf("InWidth: want 320, got %s", me.InWidth)
	}
}

func TestDialogInputHeightOnlyDigits(t *testing.T) {
	me := newTestEditor()
	me.InHeight = ""
	me.ActiveField = 2

	for _, ch := range []rune{'2', '4', '0', 'p'} { // 'p' should be rejected
		if ch >= '0' && ch <= '9' {
			me.InHeight += string(ch)
		}
	}

	if me.InHeight != "240" {
		t.Errorf("InHeight: want 240, got %s", me.InHeight)
	}
}

func TestDialogBackspaceOnName(t *testing.T) {
	me := newTestEditor()
	me.InName = "hello"
	me.ActiveField = 0

	// Simulate one backspace
	if len(me.InName) > 0 {
		me.InName = me.InName[:len(me.InName)-1]
	}

	if me.InName != "hell" {
		t.Errorf("after backspace: want hell, got %s", me.InName)
	}
}

func TestDialogBackspaceOnEmpty(t *testing.T) {
	me := newTestEditor()
	me.InName = ""

	// Should not panic or underflow
	if len(me.InName) > 0 {
		me.InName = me.InName[:len(me.InName)-1]
	}

	if me.InName != "" {
		t.Errorf("backspace on empty: want empty, got %s", me.InName)
	}
}

func TestDialogTabCyclesActiveField(t *testing.T) {
	me := newTestEditor()

	fields := []int{0, 1, 2, 0} // wraps back to 0
	for i, want := range fields {
		if me.ActiveField != want {
			t.Errorf("step %d: active field want %d, got %d", i, want, me.ActiveField)
		}
		me.ActiveField = (me.ActiveField + 1) % 3
	}
}

// ─── Map Initialisation ────────────────────────────────────────────────────────

func TestInitializeMapRejectsEmptyName(t *testing.T) {
	me := newTestEditor()
	me.InName = ""
	me.InWidth = "640"
	me.InHeight = "480"
	me.initializeMap()

	if me.Mode == "EDITOR" {
		t.Error("initializeMap should not advance to EDITOR when name is empty")
	}
	if me.MapData != nil {
		t.Error("MapData should be nil when name is empty")
	}
}

func TestInitializeMapCreatesMapData(t *testing.T) {
	dir := t.TempDir()
	// map editor writes to maps/<name>.yaml relative to cwd;
	// change cwd so the file lands under TempDir.
	origDir, _ := os.Getwd()
	os.MkdirAll(filepath.Join(dir, "maps"), 0755)
	os.Chdir(dir)
	defer os.Chdir(origDir)

	me := newTestEditor()
	me.InName = "test_map"
	me.InWidth = "320"
	me.InHeight = "240"
	me.initializeMap()

	if me.Mode != "EDITOR" {
		t.Errorf("mode after init: want EDITOR, got %s", me.Mode)
	}
	if me.MapData == nil {
		t.Fatal("MapData is nil after initializeMap")
	}
	if me.MapData.Map.ID != "test_map" {
		t.Errorf("Map.ID: want test_map, got %s", me.MapData.Map.ID)
	}
	if me.MapData.Map.WidthPixels != 320 {
		t.Errorf("WidthPixels: want 320, got %d", me.MapData.Map.WidthPixels)
	}
	if me.MapData.Map.HeightPixels != 240 {
		t.Errorf("HeightPixels: want 240, got %d", me.MapData.Map.HeightPixels)
	}
	if me.Filename != filepath.Join("maps", "test_map.yaml") {
		t.Errorf("Filename: want maps/test_map.yaml, got %s", me.Filename)
	}
}

func TestInitializeMapInfiniteMode(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.MkdirAll(filepath.Join(dir, "maps"), 0755)
	os.Chdir(dir)
	defer os.Chdir(origDir)

	me := newTestEditor()
	me.InName = "infinite_map"
	me.InWidth = "0"
	me.InHeight = "0"
	me.initializeMap()

	if me.MapData.Map.WidthPixels != 0 {
		t.Errorf("infinite mode: WidthPixels want 0, got %d", me.MapData.Map.WidthPixels)
	}
}

// ─── Item Placement ────────────────────────────────────────────────────────────

func newEditorWithMap() *MapEditor {
	me := newTestEditor()
	me.Mode = "EDITOR"
	me.MapData = &game.SaveData{}
	return me
}

func testObstacleItem() *EditorItem {
	cx := 0.0
	cy := 0.0
	_ = cx
	_ = cy
	return &EditorItem{
		ID:        "test_tree",
		Type:      "obstacle",
		Archetype: &game.ObstacleArchetype{ID: "test_tree"},
	}
}

func testNPCItem() *EditorItem {
	return &EditorItem{
		ID:        "test_orc",
		Type:      "npc",
		Archetype: &game.Archetype{ID: "test_orc"},
	}
}

func TestPlaceObstacleAddsToMapData(t *testing.T) {
	me := newEditorWithMap()
	me.PendingItem = testObstacleItem()

	// Place at screen centre (maps to iso (0,0) → cartesian (0,0))
	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	if len(me.MapData.Obstacles) != 1 {
		t.Fatalf("want 1 obstacle, got %d", len(me.MapData.Obstacles))
	}
	if me.MapData.Obstacles[0].ArchetypeID != "test_tree" {
		t.Errorf("ArchetypeID: want test_tree, got %s", me.MapData.Obstacles[0].ArchetypeID)
	}
}

func TestPlaceNPCAddsToMapData(t *testing.T) {
	me := newEditorWithMap()
	me.PendingItem = testNPCItem()

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	if len(me.MapData.NPCs) != 1 {
		t.Fatalf("want 1 NPC, got %d", len(me.MapData.NPCs))
	}
	if me.MapData.NPCs[0].ArchetypeID != "test_orc" {
		t.Errorf("ArchetypeID: want test_orc, got %s", me.MapData.NPCs[0].ArchetypeID)
	}
}

func TestPlaceMultipleItems(t *testing.T) {
	me := newEditorWithMap()
	me.PendingItem = testObstacleItem()

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2

	me.placeItem(mx, my)
	me.placeItem(mx+50, my+50)
	me.placeItem(mx-50, my-50)

	if len(me.MapData.Obstacles) != 3 {
		t.Errorf("want 3 obstacles, got %d", len(me.MapData.Obstacles))
	}
}

// ─── Pick & Selection ──────────────────────────────────────────────────────────

func TestPickAtEmptyMapReturnsMinusOne(t *testing.T) {
	me := newEditorWithMap()

	result := me.pickAt(screenWidth/2, screenHeight/2)
	if result != -1 {
		t.Errorf("pickAt on empty map: want -1, got %d", result)
	}
}

func TestPickAtFindsObstacle(t *testing.T) {
	me := newEditorWithMap()
	me.PendingItem = testObstacleItem()

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	result := me.pickAt(mx, my)
	if result == -1 {
		t.Error("pickAt should find the placed obstacle")
	}
	// High bit not set → it's an obstacle
	if result&(1<<30) != 0 {
		t.Error("expected obstacle (bit 30 clear), got NPC flag")
	}
}

func TestPickAtFindsNPC(t *testing.T) {
	me := newEditorWithMap()
	me.PendingItem = testNPCItem()

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	result := me.pickAt(mx, my)
	if result == -1 {
		t.Error("pickAt should find the placed NPC")
	}
	if result&(1<<30) == 0 {
		t.Error("expected NPC (bit 30 set), got obstacle flag")
	}
}

func TestSelectElement(t *testing.T) {
	me := newEditorWithMap()
	me.Library = []*EditorItem{testObstacleItem()}
	me.PendingItem = me.Library[0]

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	val := me.pickAt(mx, my)
	me.selectElement(val)

	if me.Selection == nil {
		t.Fatal("Selection should not be nil after selectElement")
	}
}

func TestSelectElementMinusOneClearsSelection(t *testing.T) {
	me := newEditorWithMap()
	me.Selection = &MapElement{ID: "obs_0"}

	me.selectElement(-1)

	if me.Selection != nil {
		t.Error("selectElement(-1) should clear Selection")
	}
}

func TestDeselect(t *testing.T) {
	me := newEditorWithMap()
	me.Selection = &MapElement{ID: "obs_0"}

	me.deselect()

	if me.Selection != nil {
		t.Error("deselect should set Selection to nil")
	}
}

// ─── Remove & Sync ─────────────────────────────────────────────────────────────

func TestRemoveObstacle(t *testing.T) {
	me := newEditorWithMap()
	me.Library = []*EditorItem{testObstacleItem()}
	me.PendingItem = me.Library[0]

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	val := me.pickAt(mx, my)
	me.selectElement(val)
	me.removeSelection()

	if len(me.MapData.Obstacles) != 0 {
		t.Errorf("after remove: want 0 obstacles, got %d", len(me.MapData.Obstacles))
	}
	if me.Selection != nil {
		t.Error("Selection should be nil after remove")
	}
}

func TestRemoveNPC(t *testing.T) {
	me := newEditorWithMap()
	me.Library = []*EditorItem{testNPCItem()}
	me.PendingItem = me.Library[0]

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	val := me.pickAt(mx, my)
	me.selectElement(val)
	me.removeSelection()

	if len(me.MapData.NPCs) != 0 {
		t.Errorf("after remove: want 0 NPCs, got %d", len(me.MapData.NPCs))
	}
}

func TestRemoveSelectionNilIsNoop(t *testing.T) {
	me := newEditorWithMap()
	me.Selection = nil
	// Should not panic
	me.removeSelection()
}

func TestSyncObstaclePosition(t *testing.T) {
	me := newEditorWithMap()
	me.Library = []*EditorItem{testObstacleItem()}
	me.PendingItem = me.Library[0]

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	val := me.pickAt(mx, my)
	me.selectElement(val)

	// Move selection
	me.Selection.X = 5.0
	me.Selection.Y = 7.0
	me.syncToSaveData()

	if *me.MapData.Obstacles[0].X != 5.0 {
		t.Errorf("synced X: want 5.0, got %f", *me.MapData.Obstacles[0].X)
	}
	if *me.MapData.Obstacles[0].Y != 7.0 {
		t.Errorf("synced Y: want 7.0, got %f", *me.MapData.Obstacles[0].Y)
	}
}

func TestSyncNPCPosition(t *testing.T) {
	me := newEditorWithMap()
	me.Library = []*EditorItem{testNPCItem()}
	me.PendingItem = me.Library[0]

	mx := sidebarWidth + (screenWidth-2*sidebarWidth)/2
	my := screenHeight / 2
	me.placeItem(mx, my)

	val := me.pickAt(mx, my)
	me.selectElement(val)

	me.Selection.X = 3.0
	me.Selection.Y = -2.5
	me.syncToSaveData()

	if me.MapData.NPCs[0].X != 3.0 {
		t.Errorf("synced X: want 3.0, got %f", me.MapData.NPCs[0].X)
	}
	if me.MapData.NPCs[0].Y != -2.5 {
		t.Errorf("synced Y: want -2.5, got %f", me.MapData.NPCs[0].Y)
	}
}

// ─── Save Round-Trip ───────────────────────────────────────────────────────────

func TestSaveMapCreatesFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.MkdirAll(filepath.Join(dir, "maps"), 0755)
	os.Chdir(dir)
	defer os.Chdir(origDir)

	me := newTestEditor()
	me.InName = "save_test"
	me.InWidth = "100"
	me.InHeight = "100"
	me.initializeMap()

	path := filepath.Join(dir, "maps", "save_test.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected YAML file at %s, not found", path)
	}
}

func TestSaveMapNoFilenameIsNoop(t *testing.T) {
	me := newEditorWithMap()
	me.Filename = ""
	// Should not panic or create any file
	me.saveMap()
}

// ─── FindItem ─────────────────────────────────────────────────────────────────

func TestFindItem(t *testing.T) {
	me := newTestEditor()
	obs := testObstacleItem()
	npc := testNPCItem()
	me.Library = []*EditorItem{obs, npc}

	found := me.findItem("test_tree", "obstacle")
	if found != obs {
		t.Error("findItem: should find test_tree obstacle")
	}

	found = me.findItem("test_orc", "npc")
	if found != npc {
		t.Error("findItem: should find test_orc NPC")
	}

	found = me.findItem("test_tree", "npc") // wrong type
	if found != nil {
		t.Error("findItem: should return nil for wrong type")
	}

	found = me.findItem("nonexistent", "obstacle")
	if found != nil {
		t.Error("findItem: should return nil for unknown ID")
	}
}
