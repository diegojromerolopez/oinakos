package game

import (
	"image"
	"image/color"
	"io/fs"
	"oinakos/internal/engine"
	"testing"
	"testing/fstest"
)

// MockGraphicsInheritance implements the methods needed for LoadAssets
type MockGraphicsInheritance struct {
	LoadedPaths []string
}

func (m *MockGraphicsInheritance) LoadSprite(assets fs.FS, path string, removeBg bool) engine.Image {
	m.LoadedPaths = append(m.LoadedPaths, path)
	return nil
}
func (m *MockGraphicsInheritance) NewImage(w, h int) engine.Image                 { return nil }
func (m *MockGraphicsInheritance) NewImageFromImage(img image.Image) engine.Image { return nil }
func (m *MockGraphicsInheritance) DrawPolygon(screen engine.Image, points []engine.Point, clr color.Color, width float32) {
}
func (m *MockGraphicsInheritance) NewShader(src []byte) (engine.Shader, error) { return nil, nil }
func (m *MockGraphicsInheritance) DrawImageWithShader(screen engine.Image, img engine.Image, shader engine.Shader, uniforms map[string]interface{}, options *engine.DrawImageOptions) {
}

func (m *MockGraphicsInheritance) DrawFilledRect(screen engine.Image, x, y, width, height float32, clr color.Color, antiAlias bool) {
}
func (m *MockGraphicsInheritance) DrawFilledCircle(screen engine.Image, x, y, radius float32, clr color.Color, antiAlias bool) {
}
func (m *MockGraphicsInheritance) DrawFilledEllipse(screen engine.Image, x, y, rx, ry float32, clr color.Color, antiAlias bool) {
}
func (m *MockGraphicsInheritance) DrawEllipse(screen engine.Image, x, y, rx, ry float32, clr color.Color, width float32, antiAlias bool) {
}
func (m *MockGraphicsInheritance) DrawLine(screen engine.Image, x1, y1, x2, y2 float32, clr color.Color, width float32) {
}

func (m *MockGraphicsInheritance) DebugPrintAt(screen engine.Image, str string, x, y int, clr color.Color) {
}
func (m *MockGraphicsInheritance) LoadFont(assets fs.FS, path string) error { return nil }
func (m *MockGraphicsInheritance) DrawTextAt(screen engine.Image, str string, x, y int, clr color.Color, size float64) {
}
func (m *MockGraphicsInheritance) MeasureText(str string, size float64) (float64, float64) {
	return float64(len(str)) * size * 0.5, size
}

func TestNPC_InheritanceAndSoundID(t *testing.T) {
	// Setup a mock filesystem
	fsys := fstest.MapFS{
		"data/npcs/crimson_guard.yaml": {
			Data: []byte(`
id: crimson_guard
archetype: man_at_arms
gender: male
`),
		},
		"assets/audio/npcs/crimson_guard/hit.wav":          {Data: []byte("wav data")},
		"assets/audio/archetypes/man_at_arms/male/hit.wav": {Data: []byte("wav data")},
	}

	// 1. Create registries
	archReg := NewArchetypeRegistry()
	npcReg := NewNPCRegistry()

	// 2. Add the base archetype manually
	archReg.Archetypes["man_at_arms_male"] = &EntityConfig{
		ID: "man_at_arms_male",
		Stats: struct {
			HealthMin       int     `yaml:"health_min"`
			HealthMax       int     `yaml:"health_max"`
			Speed           float64 `yaml:"speed"`
			BaseAttack      int     `yaml:"base_attack"`
			BaseDefense     int     `yaml:"base_defense"`
			AttackCooldown  int     `yaml:"attack_cooldown"`
			AttackRange     float64 `yaml:"attack_range"`
			ProjectileSpeed float64 `yaml:"projectile_speed"`
		}{HealthMin: 100},
	}

	// 3. Load the NPC from YAML
	configs := map[string]*EntityConfig{}
	configs["crimson_guard"] = &EntityConfig{
		ID:          "crimson_guard",
		ArchetypeID: "man_at_arms",
		Gender:      "male",
		AudioDir:    "assets/audio/npcs/crimson_guard",
	}
	npcReg.NPCs = configs

	// 4. Run LoadAssets to exercise inheritance logic
	graphics := &MockGraphicsInheritance{}
	npcReg.LoadAssets(fsys, graphics, archReg)

	npc := npcReg.NPCs["crimson_guard"]

	// VERIFY: Did it find the correct SoundID?
	// It should favor local audio (crimson_guard/hit.wav exists)
	if npc.SoundID != "crimson_guard" {
		t.Errorf("Expected SoundID 'crimson_guard' due to local override, got %q", npc.SoundID)
	}

	// VERIFY: Did it inherit stats from the composite ID?
	if npc.Stats.HealthMin != 100 {
		t.Errorf("Expected inherited health 100, got %d", npc.Stats.HealthMin)
	}

	// TEST 2: Inherited audio (no local wav)
	npc2 := &EntityConfig{
		ID:          "golden_guard",
		ArchetypeID: "man_at_arms",
		Gender:      "male",
		AudioDir:    "assets/audio/npcs/golden_guard", // Empty in mock FS
	}
	npcReg.NPCs["golden_guard"] = npc2
	npcReg.LoadAssets(fsys, graphics, archReg)

	if npc2.SoundID != "man_at_arms_male" {
		t.Errorf("Expected SoundID 'man_at_arms_male' from inheritance, got %q", npc2.SoundID)
	}
}

func TestNPC_GenderFallback(t *testing.T) {
	// Verify that if gender is explicitly "none", it doesn't try to append gender to lookup if it's already there
	archReg := NewArchetypeRegistry()
	archReg.Archetypes["virculus"] = &EntityConfig{ID: "virculus"}

	npcReg := NewNPCRegistry()
	npcReg.NPCs["virculus"] = &EntityConfig{
		ID:          "virculus",
		ArchetypeID: "virculus",
		Gender:      "none",
	}

	npcReg.LoadAssets(fstest.MapFS{}, &MockGraphicsInheritance{}, archReg)

	if npcReg.NPCs["virculus"].SoundID != "virculus" {
		t.Errorf("Expected SoundID 'virculus', got %q", npcReg.NPCs["virculus"].SoundID)
	}
}
