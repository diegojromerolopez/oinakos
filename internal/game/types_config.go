package game

import (
	"fmt"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type ObstacleType string
const ( TypeBuilding ObstacleType = "building"; TypeTree ObstacleType = "tree"; TypeRock ObstacleType = "rock"; TypeResource ObstacleType = "resource"; TypeBush ObstacleType = "bush" )

type FootprintPoint struct { X float64 `yaml:"x"`; Y float64 `yaml:"y"` }
func (p FootprintPoint) MarshalYAML() (any, error) {
	f := func(v float64) string { s := fmt.Sprintf("%g", v); if !stringsContains(s, ".") && !stringsContains(s, "e") { return s + ".0" }; return s }
	return &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{{Kind: yaml.ScalarNode, Value: "x"}, {Kind: yaml.ScalarNode, Value: f(p.X)}, {Kind: yaml.ScalarNode, Value: "y"}, {Kind: yaml.ScalarNode, Value: f(p.Y)}}}, nil
}
func stringsContains(s, substr string) bool { for i := 0; i < len(s)-len(substr)+1; i++ { if s[i:i+len(substr)] == substr { return true } }; return false }

type ObjectConfig struct {
	ID string `yaml:"id"`; Name string `yaml:"name"`; Description string `yaml:"description,omitempty"`; Weight float64 `yaml:"weight"`; Type string `yaml:"type"`; Category string `yaml:"category,omitempty"`; Value int `yaml:"value"`; Unique bool `yaml:"unique,omitempty"`; Resistance int `yaml:"resistance,omitempty"`; Content string `yaml:"content,omitempty"`; Consumable bool `yaml:"consumable,omitempty"`; Combat *Weapon `yaml:"combat,omitempty"`; Slot string `yaml:"slot,omitempty"`; Effects map[string]StatEffect `yaml:"effects,omitempty"`; Hunger float64 `yaml:"hunger,omitempty"`; Thirst float64 `yaml:"thirst,omitempty"`; Fatigue float64 `yaml:"fatigue,omitempty"`; Energy float64 `yaml:"energy,omitempty"`; ClearSick bool `yaml:"clear_sick,omitempty"`; MaxHours float64 `yaml:"max_hours,omitempty"`; MaxLiquid float64 `yaml:"max_liquid,omitempty"`; Refillable bool `yaml:"refillable,omitempty"`; IsAlcoholic bool `yaml:"is_alcoholic,omitempty"`
	LightRadius float64 `yaml:"light_radius,omitempty"`; IsTorch bool `yaml:"is_torch,omitempty"`
	AssetDir string `yaml:"-"`; Sprite engine.Image `yaml:"-"`; Footprint []FootprintPoint `yaml:"footprint,omitempty"`
}

type WeaponConfig struct { ID string; Inline *Weapon }
func (w *WeaponConfig) UnmarshalYAML(v *yaml.Node) error {
	var s string; if err := v.Decode(&s); err == nil { w.ID = s; return nil }
	var i Weapon; if err := v.Decode(&i); err == nil { w.Inline = &i; return nil }
	if v.Kind == yaml.ScalarNode && v.Value == "" { return nil }; return fmt.Errorf("invalid weapon")
}
func (w *WeaponConfig) IsEmpty() bool { return w.ID == "" && w.Inline == nil }
func (w *WeaponConfig) Resolve(reg *ObjectRegistry) *Weapon { if w.Inline != nil { return w.Inline }; if w.ID != "" && reg != nil { if obj, ok := reg.Objects[w.ID]; ok { return obj.Combat } }; return nil }
