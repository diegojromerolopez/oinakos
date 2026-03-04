package game

import (
	"testing"
	"testing/fstest"

	"gopkg.in/yaml.v3"
)

func TestIsWell(t *testing.T) {
	cases := []struct {
		name         string
		cooldownTime float64
		want         bool
	}{
		{"no cooldown", 0, false},
		{"negative cooldown", -1, false},
		{"has cooldown", 30, true},
		{"small cooldown", 0.1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := &ObstacleArchetype{CooldownTime: tc.cooldownTime}
			if got := a.IsWell(); got != tc.want {
				t.Errorf("IsWell() = %v, want %v (cooldown=%f)", got, tc.want, tc.cooldownTime)
			}
		})
	}
}

func TestGetWeaponByName_AllBranches(t *testing.T) {
	cases := []struct {
		name     string
		wantNil  bool
		wantName string
	}{
		{"Tizon", false, "Tizon"},
		{"Orcish Axe", false, "Orcish Axe"},
		{"Iron Broadsword", false, "Iron Broadsword"},
		{"Fists", false, "Fists"},
		{"Cleaver", false, "Cleaver"},
		{"Trident", false, "Trident"},
		{"Whip", false, "Whip"},
		{"Bow", false, "Bow"},
		{"Dagger", false, "Dagger"},
		{"Gilded Pitchfork", false, "Gilded Pitchfork"},
		{"Shouts", false, "Shouts"},
		{"unknown_weapon", false, "Fists"}, // default fallback
		{"", false, "Fists"},               // empty string fallback
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetWeaponByName(tc.name)
			if got == nil {
				t.Fatal("GetWeaponByName returned nil")
			}
			if got.Name != tc.wantName {
				t.Errorf("got weapon %q, want %q", got.Name, tc.wantName)
			}
		})
	}
}

func TestObjectiveTypeUnmarshalYAML_AllTypes(t *testing.T) {
	cases := []struct {
		yaml string
		want ObjectiveType
	}{
		{"kill_vip", ObjKillVIP},
		{"reach_portal", ObjReachPortal},
		{"survive", ObjSurvive},
		{"reach_zone", ObjReachZone},
		{"kill_count", ObjKillCount},
		{"reach_building", ObjReachBuilding},
		{"protect_npc", ObjProtectNPC},
		{"pacifist", ObjPacifist},
		{"destroy_building", ObjDestroyBuilding},
	}
	for _, tc := range cases {
		t.Run(tc.yaml, func(t *testing.T) {
			type wrapper struct {
				T ObjectiveType `yaml:"t"`
			}
			data := []byte("t: " + tc.yaml)
			var w wrapper
			if err := yaml.Unmarshal(data, &w); err != nil {
				t.Fatalf("unmarshal error for %q: %v", tc.yaml, err)
			}
			if w.T != tc.want {
				t.Errorf("got %d, want %d", w.T, tc.want)
			}
		})
	}
}

func TestObjectiveTypeUnmarshalYAML_Invalid(t *testing.T) {
	type wrapper struct {
		T ObjectiveType `yaml:"t"`
	}
	var w wrapper
	err := yaml.Unmarshal([]byte("t: not_a_valid_objective"), &w)
	if err == nil {
		t.Error("expected error for unknown objective type, got nil")
	}
}

func TestFootprintPointMarshalYAML(t *testing.T) {
	cases := []struct {
		x, y float64
	}{
		{1.0, 2.0},
		{0.0, 0.0},
		{-3.5, 7.25},
		{2, 3},       // integers should get .0
		{1e10, 5e-5}, // scientific notation
	}
	for _, tc := range cases {
		p := FootprintPoint{X: tc.x, Y: tc.y}
		out, err := p.MarshalYAML()
		if err != nil {
			t.Errorf("MarshalYAML(%v) error: %v", p, err)
		}
		if out == nil {
			t.Errorf("MarshalYAML(%v) returned nil", p)
		}
	}
}
func TestObjectiveTypeString(t *testing.T) {
	cases := []struct {
		t    ObjectiveType
		want string
	}{
		{ObjKillVIP, "kill_vip"},
		{ObjReachPortal, "reach_portal"},
		{ObjSurvive, "survive"},
		{ObjReachZone, "reach_zone"},
		{ObjKillCount, "kill_count"},
		{ObjReachBuilding, "reach_building"},
		{ObjProtectNPC, "protect_npc"},
		{ObjPacifist, "pacifist"},
		{ObjDestroyBuilding, "destroy_building"},
	}
	for _, tc := range cases {
		if got := tc.t.String(); got != tc.want {
			t.Errorf("ObjectiveType(%d).String() = %q, want %q", int(tc.t), got, tc.want)
		}
	}
}

func TestForEachYAML_ErrorPath(t *testing.T) {
	// 1. Assets is nil, should skip first walk
	err := forEachYAML(nil, "something_non_existent", func(fpath string, data []byte) error { return nil })
	if err != nil {
		t.Errorf("expected no error from forEachYAML when assets=nil, got %v", err)
	}
}

func TestMapTypeRegistry_LoadAll_Empty(t *testing.T) {
	r := NewMapTypeRegistry()
	mockFS := fstest.MapFS{} // empty FS
	err := r.LoadAll(mockFS)
	if err != nil {
		t.Errorf("LoadAll(empty) should not error, got %v", err)
	}
}

func TestCampaignRegistry_LoadAll_Empty(t *testing.T) {
	r := NewCampaignRegistry()
	mockFS := fstest.MapFS{} // empty FS
	err := r.LoadAll(mockFS)
	if err != nil {
		t.Errorf("LoadAll(empty) should not error, got %v", err)
	}
}
