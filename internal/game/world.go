package game

import (
	"image"
)

// World holds all live game entities and spatial data.
type World struct {
	PlayableCharacter *PlayableCharacter
	NPCs              []*NPC
	Obstacles         []*Obstacle
	Projectiles       []*Projectile
	FloatingTexts     []*FloatingText
	
	CurrentMapType    *MapType
	ExploredTiles     map[image.Point]bool
	PlayTime          float64
}

func NewWorld() *World {
	return &World{
		NPCs:          make([]*NPC, 0),
		Obstacles:     make([]*Obstacle, 0),
		Projectiles:   make([]*Projectile, 0),
		FloatingTexts: make([]*FloatingText, 0),
		ExploredTiles: make(map[image.Point]bool),
	}
}
