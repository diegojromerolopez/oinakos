package game

import (
	"oinakos/internal/engine"
)

// SystemContext bundles external dependencies needed for game logic updates.
type SystemContext struct {
	World      *World
	Input      engine.Input
	Audio      AudioManager
	Registries *RegistryContainer
	Log        func(string, LogCategory)
}
