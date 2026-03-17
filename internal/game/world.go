package game

import (
	"image"
)

// World holds all live game entities and spatial data.
type World struct {
	PlayableCharacter *Character
	Characters        []*Character
	Obstacles         []*Obstacle
	Projectiles       []*Projectile
	FloatingTexts     []*FloatingText
	
	CurrentMapType    *MapType
	ExploredTiles     map[image.Point]bool
	PlayTime          float64
	Game              *Game
	Items             []*ItemInstance
}

func NewWorld() *World {
	return &World{
		Characters:    make([]*Character, 0),
		Obstacles:     make([]*Obstacle, 0),
		Projectiles:   make([]*Projectile, 0),
		FloatingTexts: make([]*FloatingText, 0),
		ExploredTiles: make(map[image.Point]bool),
	}
}
