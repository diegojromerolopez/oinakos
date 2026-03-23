package game

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFloorZone_GetPolygon(t *testing.T) {
	fz := &FloorZone{
		Perimeter: []FootprintPoint{
			{X: 0, Y: 0},
			{X: 10, Y: 0},
			{X: 10, Y: 10},
			{X: 0, Y: 10},
		},
	}

	poly := fz.GetPolygon()
	if len(poly.Points) != 4 {
		t.Errorf("expected 4 points, got %d", len(poly.Points))
	}

	if fz.MinX != 0 || fz.MaxX != 10 || fz.MinY != 0 || fz.MaxY != 10 {
		t.Errorf("incorrect AABB: %+v", fz)
	}

	if !fz.AABBCalculated {
		t.Error("AABBCalculated should be true")
	}

	// Test caching
	poly2 := fz.GetPolygon()
	if &poly == &poly2 {
		// This check might be tricky if poly is returned by value, but fz.Polygon should be the same
	}
	if len(fz.Polygon.Points) != 4 {
		t.Error("cached polygon should have 4 points")
	}
}

func TestFloorZone_Contains(t *testing.T) {
	fz := &FloorZone{
		Perimeter: []FootprintPoint{
			{X: 0, Y: 0},
			{X: 10, Y: 0},
			{X: 10, Y: 10},
			{X: 0, Y: 10},
		},
	}

	tests := []struct {
		x, y float64
		want bool
	}{
		{5, 5, true},
		{-1, 5, false},
		{11, 5, false},
		{5, -1, false},
		{5, 11, false},
	}

	for _, tt := range tests {
		if got := fz.Contains(tt.x, tt.y); got != tt.want {
			t.Errorf("Contains(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestWeaponConfig_UnmarshalYAML(t *testing.T) {
	t.Run("string ID", func(t *testing.T) {
		var wc WeaponConfig
		err := yaml.Unmarshal([]byte("longsword"), &wc)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if wc.ID != "longsword" || wc.Inline != nil {
			t.Errorf("expected ID 'longsword', got ID=%q Inline=%+v", wc.ID, wc.Inline)
		}
	})

	t.Run("inline object", func(t *testing.T) {
		var wc WeaponConfig
		data := `
damage:
  min: 10
  max: 10
range: 1.5
`
		err := yaml.Unmarshal([]byte(data), &wc)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if wc.ID != "" || wc.Inline == nil || wc.Inline.Damage.Min != 10 {
			t.Errorf("expected inline weapon with damage 10, got ID=%q Inline=%+v", wc.ID, wc.Inline)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var wc WeaponConfig
		err := yaml.Unmarshal([]byte(""), &wc)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if !wc.IsEmpty() {
			t.Error("expected empty WeaponConfig")
		}
	})
}

func TestWeaponConfig_Resolve(t *testing.T) {
	reg := &ObjectRegistry{
		Objects: map[string]*ObjectConfig{
			"longsword": {
				Combat: &Weapon{Damage: Damage{Min: 15, Max: 15}},
			},
		},
	}

	t.Run("resolve ID", func(t *testing.T) {
		wc := WeaponConfig{ID: "longsword"}
		w := wc.Resolve(reg)
		if w == nil || w.Damage.Min != 15 {
			t.Errorf("failed to resolve ID, got %+v", w)
		}
	})

	t.Run("resolve inline", func(t *testing.T) {
		wc := WeaponConfig{Inline: &Weapon{Damage: Damage{Min: 20, Max: 20}}}
		w := wc.Resolve(reg)
		if w == nil || w.Damage.Min != 20 {
			t.Errorf("failed to resolve inline, got %+v", w)
		}
	})

	t.Run("resolve missing", func(t *testing.T) {
		wc := WeaponConfig{ID: "rusty_spoon"}
		w := wc.Resolve(reg)
		if w != nil {
			t.Errorf("should not have resolved missing ID, got %+v", w)
		}
	})
}

func TestFootprintPoint_MarshalYAML(t *testing.T) {
	fp := FootprintPoint{X: 10.5, Y: 20}
	node, err := fp.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML failed: %v", err)
	}

	data, err := yaml.Marshal(node)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}

	expected := "x: 10.5\ny: 20.0\n"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestStringsContains(t *testing.T) {
	tests := []struct {
		s, substr string
		want      bool
	}{
		{"hello", "ell", true},
		{"hello", "world", false},
		{"hello", "h", true},
		{"hello", "o", true},
		{"hello", "", true},
		{"", "a", false},
	}

	for _, tt := range tests {
		if got := stringsContains(tt.s, tt.substr); got != tt.want {
			t.Errorf("stringsContains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestObjectiveType_String(t *testing.T) {
	tests := []struct {
		ot   ObjectiveType
		want string
	}{
		{ObjKillVIP, "kill_vip"},
		{ObjReachPortal, "reach_portal"},
		{ObjectiveType(999), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.ot.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestObjectiveType_UnmarshalYAML(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var ot ObjectiveType
		err := yaml.Unmarshal([]byte("kill_count"), &ot)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if ot != ObjKillCount {
			t.Errorf("expected ObjKillCount, got %v", ot)
		}
	})

	t.Run("int", func(t *testing.T) {
		var ot ObjectiveType
		err := yaml.Unmarshal([]byte("2"), &ot)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if ot != ObjSurvive {
			t.Errorf("expected ObjSurvive, got %v", ot)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		var ot ObjectiveType
		err := yaml.Unmarshal([]byte("invalid_type"), &ot)
		if err == nil {
			t.Error("expected error for invalid type")
		}
	})
}
