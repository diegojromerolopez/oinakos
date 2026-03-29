package game

import (
	"testing"
)

func TestChronos_TimeProgression(t *testing.T) {
	g := setupTestGame()
	s := &g.World.State
	
	// Start at 12:00
	s.Hour = 12
	s.Ticks = 719
	
	// One more tick should trigger next hour
	g.updateWorldState()
	if s.Hour != 13 {
		t.Errorf("Expected hour 13, got %d", s.Hour)
	}
	if s.Ticks != 0 {
		t.Errorf("Expected ticks reset to 0, got %d", s.Ticks)
	}
}

func TestChronos_SeasonProgression(t *testing.T) {
	g := setupTestGame()
	s := &g.World.State
	
	// Season Logic: 3 months per season. 4 days per month.
	// 12 days per season.
	
	// Set to end of Winter (Month 2, Day 4 / 4, Hour 23, Ticks 719)
	// Winter is Months 12, 1, 2
	s.Season = SeasonWinter
	s.Month = 2
	s.Day = 8
	s.Hour = 23
	s.Ticks = 719
	
	g.updateWorldState()
	
	if s.Month != 3 {
		t.Errorf("Expected Month 3, got %d", s.Month)
	}
	if s.Season != SeasonSpring {
		t.Errorf("Expected Season Spring after Winter, got %s", s.Season)
	}
}

func TestChronos_DiurnalTemperature(t *testing.T) {
	g := setupTestGame()
	s := &g.World.State
	
	// Noon (Hour 12) should be warmer than Midnight (Hour 0)
	s.Season = SeasonSpring
	s.Hour = 12
	g.updateWorldState()
	tempNoon := s.Temperature
	
	s.Hour = 0
	g.updateWorldState()
	tempMidnight := s.Temperature
	
	if tempNoon <= tempMidnight {
		t.Errorf("Noon temperature (%.1f) should be greater than Midnight temperature (%.1f)", tempNoon, tempMidnight)
	}
}
