package game

import (
	"fmt"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type ObstacleActionType string

const (
	ActionHarm ObstacleActionType = "harm"
	ActionHeal ObstacleActionType = "heal"
)

type ObstacleActionConfig struct {
	Type                ObstacleActionType `yaml:"type"`
	Amount              int                `yaml:"amount"`
	Aura                float64            `yaml:"aura"`
	AlignmentLimit      string             `yaml:"alignment_limit"`      // "all", "ally", "enemy"
	RequiresInteraction bool               `yaml:"requires_interaction"` // e.g. the Well
}

type ObstacleType string

const (
	TypeBuilding ObstacleType = "building"
	TypeTree     ObstacleType = "tree"
	TypeRock     ObstacleType = "rock"
	TypeResource ObstacleType = "resource"
	TypeBush     ObstacleType = "bush"
)

type ObjectiveType int

const (
	ObjKillVIP ObjectiveType = iota
	ObjReachPortal
	ObjSurvive
	ObjReachZone
	ObjKillCount
	ObjReachBuilding
	ObjProtectNPC
	ObjPacifist
	ObjDestroyBuilding
)

func (t ObjectiveType) String() string {
	switch t {
	case ObjKillVIP:
		return "kill_vip"
	case ObjReachPortal:
		return "reach_portal"
	case ObjSurvive:
		return "survive"
	case ObjReachZone:
		return "reach_zone"
	case ObjKillCount:
		return "kill_count"
	case ObjReachBuilding:
		return "reach_building"
	case ObjProtectNPC:
		return "protect_npc"
	case ObjPacifist:
		return "pacifist"
	case ObjDestroyBuilding:
		return "destroy_building"
	}
	return "unknown"
}

func (t *ObjectiveType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err == nil {
		switch s {
		case "kill_vip":
			*t = ObjKillVIP
			return nil
		case "reach_portal":
			*t = ObjReachPortal
			return nil
		case "survive":
			*t = ObjSurvive
			return nil
		case "reach_zone":
			*t = ObjReachZone
			return nil
		case "kill_count":
			*t = ObjKillCount
			return nil
		case "reach_building":
			*t = ObjReachBuilding
			return nil
		case "protect_npc":
			*t = ObjProtectNPC
			return nil
		case "pacifist":
			*t = ObjPacifist
			return nil
		case "destroy_building":
			*t = ObjDestroyBuilding
			return nil
		}
	}

	var i int
	if err := value.Decode(&i); err == nil {
		*t = ObjectiveType(i)
		return nil
	}

	return fmt.Errorf("unknown objective type: %v", value.Value)
}

func (a *Alignment) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err == nil {
		switch s {
		case "enemy":
			*a = AlignmentEnemy
			return nil
		case "neutral":
			*a = AlignmentNeutral
			return nil
		case "ally":
			*a = AlignmentAlly
			return nil
		}
	}
	var i int
	if err := value.Decode(&i); err == nil {
		*a = Alignment(i)
		return nil
	}
	return fmt.Errorf("unknown alignment: %v", value.Value)
}

type Inhabitant struct {
	ID          string    `yaml:"id,omitempty"` // For internal mapping if needed
	Name        string    `yaml:"name,omitempty"`
	Archetype   string    `yaml:"archetype,omitempty"`
	ArchetypeID string    `yaml:"archetype_id,omitempty"`
	NPC         string    `yaml:"npc,omitempty"`
	NPCID       string    `yaml:"npc_id,omitempty"`
	X           float64   `yaml:"x"`
	Y           float64   `yaml:"y"`
	State       string    `yaml:"state,omitempty"` // e.g. "dead", empty means alive
	Alignment   Alignment `yaml:"alignment"`
	MustSurvive bool      `yaml:"must_survive,omitempty"`
}

type PreSpawnObstacle struct {
	ID          string   `yaml:"id"`
	Archetype   string   `yaml:"archetype"`
	ArchetypeID string   `yaml:"archetype_id,omitempty"`
	X           *float64 `yaml:"x,omitempty"`
	Y           *float64 `yaml:"y,omitempty"`
	Disabled    bool     `yaml:"disabled,omitempty"`
}

type SpawnConfig struct {
	Archetype   string    `yaml:"archetype"`
	Alignment   Alignment `yaml:"alignment"`
	Probability float64   `yaml:"probability"` // 0.0 to 1.0
	Frequency   float64   `yaml:"frequency"`   // seconds
	X           *float64  `yaml:"x,omitempty"`
	Y           *float64  `yaml:"y,omitempty"`

	Timer int `yaml:"-"` // Internal tick counter
}

type TargetPointConfig struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}

type FootprintPoint struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}

// MarshalYAML emits the point as a plain YAML mapping without quoting the 'y'
func (p FootprintPoint) MarshalYAML() (interface{}, error) {
	format := func(f float64) string {
		s := fmt.Sprintf("%g", f)
		if !stringsContains(s, ".") && !stringsContains(s, "e") {
			return s + ".0"
		}
		return s
	}
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "x"},
			{Kind: yaml.ScalarNode, Value: format(p.X)},
			{Kind: yaml.ScalarNode, Value: "y"},
			{Kind: yaml.ScalarNode, Value: format(p.Y)},
		},
	}, nil
}

// Helper because we can't import strings in types_config easily without polluting
func stringsContains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type FloorZone struct {
	Name      string           `yaml:"name"`
	Tile      string           `yaml:"tile"`
	Priority  int              `yaml:"priority"`
	Perimeter []FootprintPoint `yaml:"perimeter"`
	Polygon   engine.Polygon   `yaml:"-"`

	MinX float64 `yaml:"-"`
	MaxX float64 `yaml:"-"`
	MinY float64 `yaml:"-"`
	MaxY float64 `yaml:"-"`
	AABBCalculated bool `yaml:"-"`
}

func (fz *FloorZone) GetPolygon() engine.Polygon {
	if len(fz.Polygon.Points) > 0 {
		return fz.Polygon
	}
	pts := make([]engine.Point, len(fz.Perimeter))
	
	if len(fz.Perimeter) > 0 {
		fz.MinX, fz.MaxX = fz.Perimeter[0].X, fz.Perimeter[0].X
		fz.MinY, fz.MaxY = fz.Perimeter[0].Y, fz.Perimeter[0].Y
	}

	for i, pt := range fz.Perimeter {
		pts[i] = engine.Point{X: pt.X, Y: pt.Y}
		if pt.X < fz.MinX { fz.MinX = pt.X }
		if pt.X > fz.MaxX { fz.MaxX = pt.X }
		if pt.Y < fz.MinY { fz.MinY = pt.Y }
		if pt.Y > fz.MaxY { fz.MaxY = pt.Y }
	}
	
	fz.AABBCalculated = true
	fz.Polygon = engine.Polygon{Points: pts}
	return fz.Polygon
}

func (fz *FloorZone) Contains(x, y float64) bool {
	poly := fz.GetPolygon()
	if fz.AABBCalculated {
		if x < fz.MinX || x > fz.MaxX || y < fz.MinY || y > fz.MaxY {
			return false
		}
	}
	return poly.Contains(x, y)
}

type ActionConfig struct {
	OnKill []KillAction `yaml:"on_kill"`
}

type KillAction struct {
	Type        string       `yaml:"type"`
	Probability float64      `yaml:"probability"`
	Effect      ActionEffect `yaml:"effect"`
}

type ActionEffect struct {
	Victim   *VictimEffect   `yaml:"victim,omitempty"`
	Attacker *AttackerEffect `yaml:"attacker,omitempty"`
}

type VictimEffect struct {
	Transform   string `yaml:"transform,omitempty"`
	Alignment   string `yaml:"alignment,omitempty"`
	SpawnCorpse *bool  `yaml:"spawn_corpse,omitempty"` // Pointer to distinguish between false and omission
	CorpseImage string `yaml:"corpse_image,omitempty"`
}

type AttackerEffect struct {
	Heal int `yaml:"heal,omitempty"`
}
