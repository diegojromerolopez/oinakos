package main

import (
	"oinakos/internal/engine"
	"oinakos/internal/game"
	"testing"
)

func TestNewMapEditor(t *testing.T) {
	g := &engine.MockGraphics{}
	in := engine.NewMockInput()
	me := NewMapEditor(g, in)

	if me == nil {
		t.Fatal("NewMapEditor returned nil")
	}
	if me.Mode != "DIALOG" {
		t.Errorf("Initial mode: got %s, want DIALOG", me.Mode)
	}
}

func TestMapEditor_InitializeMap(t *testing.T) {
	g := &engine.MockGraphics{}
	in := engine.NewMockInput()
	me := NewMapEditor(g, in)

	me.InName = "test_map"
	me.InWidth = "100"
	me.InHeight = "100"

	// We call initializeMap directly to check state change
	me.initializeMap()

	if me.Mode != "EDITOR" {
		t.Errorf("Expected mode EDITOR after init, got %s", me.Mode)
	}
	if me.MapData == nil {
		t.Fatal("MapData not initialized")
	}
	if me.MapData.Map.ID != "test_map" {
		t.Errorf("Map ID: got %s, want test_map", me.MapData.Map.ID)
	}
}

func TestMapEditor_Selection(t *testing.T) {
	g := &engine.MockGraphics{}
	in := engine.NewMockInput()
	me := NewMapEditor(g, in)

	// Mock data with one NPC
	me.MapData = &game.SaveData{
		NPCs: []game.NPCSaveData{
			{ArchetypeID: "peasant_male", X: 5.0, Y: 5.0},
		},
	}
	me.Mode = "EDITOR"

	// Select the first NPC (bit 30 set for NPC)
	me.selectElement(0 | (1 << 30))

	if me.Selection == nil {
		t.Fatal("Selection should not be nil")
	}
	if me.Selection.X != 5.0 || me.Selection.Y != 5.0 {
		t.Errorf("Selection position: got (%f, %f), want (5.0, 5.0)", me.Selection.X, me.Selection.Y)
	}
}
