package game

import (
	"fmt"
	"math"
	"oinakos/internal/engine"
	"gopkg.in/yaml.v3"
)

type ObjectiveType int
const ( ObjKillVIP ObjectiveType = iota; ObjReachPortal; ObjSurvive; ObjReachZone; ObjKillCount; ObjReachBuilding; ObjProtectNPC; ObjPacifist; ObjDestroyBuilding; ObjSandbox; ObjSimulation )
func (t ObjectiveType) String() string { 
	names := []string{"kill_vip", "reach_portal", "survive", "reach_zone", "kill_count", "reach_building", "protect_npc", "pacifist", "destroy_building", "sandbox", "simulation"}
	if int(t) < len(names) { return names[t] }; return "unknown"
}
func (t *ObjectiveType) UnmarshalYAML(v *yaml.Node) error {
	var s string; if err := v.Decode(&s); err == nil {
		m := map[string]ObjectiveType{"kill_vip": ObjKillVIP, "reach_portal": ObjReachPortal, "survive": ObjSurvive, "reach_zone": ObjReachZone, "kill_count": ObjKillCount, "reach_building": ObjReachBuilding, "protect_npc": ObjProtectNPC, "pacifist": ObjPacifist, "destroy_building": ObjDestroyBuilding, "sandbox": ObjSandbox, "simulation": ObjSimulation}
		if typ, ok := m[s]; ok { *t = typ; return nil }
	}
	var i int; if err := v.Decode(&i); err == nil { *t = ObjectiveType(i); return nil }; return fmt.Errorf("unknown objective type")
}

type Inhabitant struct {
	ID          string    `yaml:"id,omitempty"`
	Name        string    `yaml:"name,omitempty"`
	Archetype   string    `yaml:"archetype,omitempty"`
	NPC         string    `yaml:"npc,omitempty"`
	NPCID       string    `yaml:"npc_id,omitempty"`
	X           float64   `yaml:"x"`
	Y           float64   `yaml:"y"`
	State       string    `yaml:"state,omitempty"`
	Alignment   Alignment `yaml:"alignment"`
	Behavior    string    `yaml:"behavior,omitempty"`
	MustSurvive bool      `yaml:"must_survive,omitempty"`
	IsTarget    bool      `yaml:"is_target,omitempty"`
}
type PreSpawnObstacle struct { ID string `yaml:"id"`; Archetype string `yaml:"archetype"`; Actions *ActionConfig `yaml:"actions,omitempty"`; Weapon WeaponConfig `yaml:"weapon"`; CollisionRadius float64 `yaml:"collision_radius,omitempty"`; X *float64 `yaml:"x,omitempty"`; Y *float64 `yaml:"y,omitempty"`; Disabled bool `yaml:"disabled,omitempty"` }

type SpawnConfig struct {
	Archetype   string    `yaml:"archetype"`
	Alignment   Alignment `yaml:"alignment"`
	Probability float64   `yaml:"probability"`
	Frequency   float64   `yaml:"frequency"`
	X           *float64  `yaml:"x,omitempty"`
	Y           *float64  `yaml:"y,omitempty"`
	Timer       int       `yaml:"-"`
}

type FloorZone struct {
	Name      string           `yaml:"name"`
	Tile      string           `yaml:"tile"`
	Priority  int              `yaml:"priority"`
	Perimeter []FootprintPoint `yaml:"perimeter"`
	Type      string           `yaml:"type,omitempty"`
	Accepts   []string         `yaml:"accepts,omitempty"`
	Polygon   engine.Polygon   `yaml:"-"`
	MinX, MaxX, MinY, MaxY float64 `yaml:"-"`; AABBCalculated bool `yaml:"-"`
}
func (fz *FloorZone) GetPolygon() engine.Polygon {
	if len(fz.Polygon.Points) > 0 { return fz.Polygon }
	pts := make([]engine.Point, len(fz.Perimeter))
	if len(fz.Perimeter) > 0 { fz.MinX, fz.MaxX, fz.MinY, fz.MaxY = fz.Perimeter[0].X, fz.Perimeter[0].X, fz.Perimeter[0].Y, fz.Perimeter[0].Y }
	for i, pt := range fz.Perimeter { pts[i] = engine.Point{X: pt.X, Y: pt.Y}; if pt.X < fz.MinX { fz.MinX = pt.X }; if pt.X > fz.MaxX { fz.MaxX = pt.X }; if pt.Y < fz.MinY { fz.MinY = pt.Y }; if pt.Y > fz.MaxY { fz.MaxY = pt.Y } }
	fz.AABBCalculated, fz.Polygon = true, engine.Polygon{Points: pts}; return fz.Polygon
}
func (fz *FloorZone) Contains(x, y float64) bool { if !fz.GetPolygon().Contains(x, y) { return false }; if fz.AABBCalculated { if x < fz.MinX || x > fz.MaxX || y < fz.MinY || y > fz.MaxY { return false } }; return fz.Polygon.Contains(x, y) }
type TargetPointConfig struct { X float64 `yaml:"x"`; Y float64 `yaml:"y"` }
func (fz *FloorZone) DistanceTo(x, y float64) float64 { fz.GetPolygon(); cx, cy := (fz.MinX+fz.MaxX)*0.5, (fz.MinY+fz.MaxY)*0.5; return math.Sqrt(math.Pow(x-cx, 2) + math.Pow(y-cy, 2)) }
