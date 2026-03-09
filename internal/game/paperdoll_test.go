package game

import (
	"image/color"
	"io/fs"
	"oinakos/internal/engine"
	"testing"
)

type paperdollMockGraphics struct {
	engine.Graphics
}

func (m *paperdollMockGraphics) LoadSprite(assets fs.FS, path string, removeBg bool) engine.Image {
	return &paperdollMockImage{}
}

func (m *paperdollMockGraphics) DebugPrintAt(screen engine.Image, str string, x, y int, clr color.Color) {
}

type paperdollMockImage struct {
	engine.Image
}

func (m *paperdollMockImage) Size() (int, int) {
	return 160, 160
}

type emptyFS struct{ fs.FS }

func (emptyFS) Open(name string) (fs.File, error) { return nil, fs.ErrNotExist }

func TestPaperdollConfig(t *testing.T) {
	config := &EntityConfig{
		Engine:   "paperdoll",
		AssetDir: "assets/images/characters/conde_olinos/paperdoll",
	}

	graphics := &paperdollMockGraphics{}
	assets := emptyFS{}

	config.LoadPaperdoll(assets, graphics)

	if config.Paperdoll == nil {
		t.Fatal("Paperdoll assets should be initialized")
	}

	if config.Paperdoll.Layers == nil {
		t.Fatal("Paperdoll layers should be initialized")
	}
}
