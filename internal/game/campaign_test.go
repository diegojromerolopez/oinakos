package game

import (
	"testing"

	"oinakos/internal/engine"
)

func TestCampaignProgression(t *testing.T) {
	// Setup mock assets with a simple campaign
	// In a real test we'd mock the FS better, but for this project we'll rely on the logic itself.

	g := NewGame(nil, "", "", engine.NewMockInput(), &DefaultAudioManager{}, false)

	// Manually inject a campaign
	camp := &Campaign{
		ID:   "test_campaign",
		Name: "Test Campaign",
		Maps: []string{"safe_zone", "ancient_tavern"},
	}
	g.campaignRegistry.Campaigns[camp.ID] = camp
	g.campaignRegistry.IDs = append(g.campaignRegistry.IDs, camp.ID)

	// Simulate selecting the campaign
	g.isCampaignSelect = true
	g.campaignMenuIndex = 0

	input := g.input.(*engine.MockInput)
	input.JustPressedKeys[engine.KeyEnter] = true
	g.Update()
	input.JustPressedKeys[engine.KeyEnter] = false

	if g.isCampaignSelect {
		t.Error("isCampaignSelect should be false after selecting")
	}
	if !g.isCampaign {
		t.Error("isCampaign should be true")
	}
	if g.currentCampaign == nil || g.currentCampaign.ID != "test_campaign" {
		t.Error("currentCampaign not set correctly")
	}

	// Check if first map is loaded (assuming registries are populated enough)
	// Even if it fails to find the map in a nil FS, we can check the index.
	if g.campaignIndex != 0 {
		t.Errorf("Expected campaignIndex 0, got %d", g.campaignIndex)
	}

	// Simulate winning the first map and moving to the next
	g.isMapWon = true
	g.mapWonMenuIndex = WinMenuContinue

	input.JustPressedKeys[engine.KeyEnter] = true
	g.Update() // This should increment campaignIndex
	input.JustPressedKeys[engine.KeyEnter] = false

	if g.campaignIndex != 1 {
		t.Errorf("Expected campaignIndex 1 after win, got %d", g.campaignIndex)
	}
	if g.isMapWon {
		t.Error("isMapWon should be reset after continue")
	}

	// Simulate winning the last map in the campaign
	g.isMapWon = true
	g.mapWonMenuIndex = WinMenuContinue

	input.JustPressedKeys[engine.KeyEnter] = true
	g.Update() // This should set isGameWon since index reaches length
	input.JustPressedKeys[engine.KeyEnter] = false

	if !g.isGameWon {
		t.Error("isGameWon should be true after completing last map")
	}
}

func TestGameWon_ReplayClearsState(t *testing.T) {
	g := NewGame(nil, "", "", engine.NewMockInput(), &DefaultAudioManager{}, false)
	g.isGameWon = true
	g.isCampaign = false
	g.mapWonMenuIndex = WinMenuContinue

	input := g.input.(*engine.MockInput)
	input.JustPressedKeys[engine.KeyEnter] = true
	g.Update()

	if g.isGameWon {
		t.Error("isGameWon should be cleared after selecting Replay")
	}
}

func TestGameWon_MenuNavigatesDown(t *testing.T) {
	g := NewGame(nil, "", "", engine.NewMockInput(), &DefaultAudioManager{}, false)
	g.isGameWon = true
	g.mapWonMenuIndex = 0

	input := g.input.(*engine.MockInput)
	input.JustPressedKeys[engine.KeyDown] = true
	g.Update()

	if g.mapWonMenuIndex != 1 {
		t.Errorf("Expected mapWonMenuIndex=1 after Down, got %d", g.mapWonMenuIndex)
	}
}

func TestGameWon_MenuNavigatesUp(t *testing.T) {
	g := NewGame(nil, "", "", engine.NewMockInput(), &DefaultAudioManager{}, false)
	g.isGameWon = true
	g.mapWonMenuIndex = 1 // Start on Quit

	input := g.input.(*engine.MockInput)
	input.JustPressedKeys[engine.KeyUp] = true
	g.Update()

	if g.mapWonMenuIndex != 0 {
		t.Errorf("Expected mapWonMenuIndex=0 after Up, got %d", g.mapWonMenuIndex)
	}
}
