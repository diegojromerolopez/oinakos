package game

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAlignmentString(t *testing.T) {
	cases := []struct {
		alignment Alignment
		want      string
	}{
		{AlignmentEnemy, "hostile"},
		{AlignmentNeutral, "neutral"},
		{AlignmentAlly, "friendly"},
		{Alignment(99), "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.alignment.String(); got != tc.want {
				t.Errorf("Alignment(%d).String() = %q, want %q", int(tc.alignment), got, tc.want)
			}
		})
	}
}

func TestAlignmentUnmarshalYAML(t *testing.T) {
	cases := []struct {
		input string
		want  Alignment
	}{
		{"enemy", AlignmentEnemy},
		{"neutral", AlignmentNeutral},
		{"ally", AlignmentAlly},
		{"0", AlignmentNeutral},   // numeric form
		{"1", AlignmentFriendly}, // numeric form
		{"2", AlignmentHostile},    // numeric form
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			type wrapper struct {
				A Alignment `yaml:"a"`
			}
			data := []byte("a: " + tc.input)
			var w wrapper
			if err := yaml.Unmarshal(data, &w); err != nil {
				t.Fatalf("unmarshal error for %q: %v", tc.input, err)
			}
			if w.A != tc.want {
				t.Errorf("got %v, want %v", w.A, tc.want)
			}
		})
	}
}

func TestAlignmentUnmarshalYAML_Invalid(t *testing.T) {
	type wrapper struct {
		A Alignment `yaml:"a"`
	}
	var w wrapper
	err := yaml.Unmarshal([]byte("a: bogus_value"), &w)
	if err == nil {
		t.Error("expected error for unknown alignment string, got nil")
	}
}
