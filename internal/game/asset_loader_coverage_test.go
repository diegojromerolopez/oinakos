package game

import (
	"oinakos/internal/engine"
	"testing"
	"testing/fstest"
)

func TestAssetLoader_Coverage(t *testing.T) {
	g := setupTestGame()
	g.characterRegistry.IDs = []string{"id1"}
	g.characterRegistry.Characters["id1"] = &EntityConfig{ID: "id1"}
	
	// Create a dummy FS with some file
	fsys := fstest.MapFS{
		"assets/images/characters/id1/static.png": {Data: []byte("fake image")},
	}
	
	gr := NewGameRenderer(g, fsys, &engine.MockGraphics{})
	gr.LoadAssets(fsys)
	
	// Test npc registries loading assets
	g.characterRegistry.LoadAssets(fsys, &engine.MockGraphics{}, g.archetypeRegistry, nil, nil)
	g.archetypeRegistry.LoadAssets(fsys, &engine.MockGraphics{}, nil, nil)
	g.obstacleRegistry.LoadAssets(fsys, &engine.MockGraphics{}, nil, nil)
}
