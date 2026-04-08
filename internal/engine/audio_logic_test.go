package engine

import (
	"reflect"
	"testing"
)

func TestSharedAudioManager_GetMatchingKeys(t *testing.T) {
	m := &SharedAudioManager{}
	available := []string{"hit", "attack_1", "attack_2", "orc/hit", "archer/bow/hit", "other"}
	
	tests := []struct {
		prefix string
		want   []string
	}{
		{"hit", []string{"hit"}},
		{"attack", []string{"attack_1", "attack_2"}},
		{"orc", []string{"orc/hit"}},
		{"archer/bow", []string{"archer/bow/hit"}},
		{"missing", nil},
	}
	
	for _, tt := range tests {
		got := m.getMatchingKeys(tt.prefix, available)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("getMatchingKeys(%q) = %v, want %v", tt.prefix, got, tt.want)
		}
	}
}

func TestSharedAudioManager_PickRandom(t *testing.T) {
	m := &SharedAudioManager{}
	available := []string{"attack_1", "attack_2", "attack_3"}
	
	// Test picking from multiple
	results := make(map[string]bool)
	for i := 0; i < 100; i++ {
		res := m.PickRandom("attack", available)
		results[res] = true
	}
	
	if len(results) < 2 {
		t.Errorf("PickRandom should vary across multiple keys")
	}
	
	// Test single
	res := m.PickRandom("attack_1", available)
	if res != "attack_1" {
		t.Errorf("Expected attack_1, got %q", res)
	}
	
	// Test empty
	res = m.PickRandom("missing", available)
	if res != "" {
		t.Errorf("Expected empty string for missing prefix, got %q", res)
	}
}
