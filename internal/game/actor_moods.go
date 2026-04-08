package game

import (
	"fmt"
	"math/rand"
	"strings"
)

type MoodType int

const (
	MoodNeutral MoodType = iota
	MoodStrange          // Fecund inspiration (Dwarf Fortress style)
	MoodMelancholy
	MoodBerserk
)

func (a *Actor) updateMood(ctx *SystemContext) {
	if !a.IsAlive() || a.Config == nil || a.Config.IsAnimal { return }

	// 1. Chance to enter Strange Mood (applies to neutral NPCs)
	if a.Mood == MoodNeutral && a.Alignment == AlignmentNeutral && rand.Intn(100000) < 5 {
		a.Mood = MoodStrange
		a.WorkTicks = 0 // Used as a timer for the mood
		if ctx.Log != nil {
			ctx.Log(fmt.Sprintf("%s has entered a STRANGE MOOD!", a.Name), LogNPC)
		}
	}

	if a.Mood == MoodStrange {
		a.updateStrangeMood(ctx)
	}
}

func (a *Actor) updateStrangeMood(ctx *SystemContext) {
	a.WorkTicks++
	
	// NPCs in strange mood seek materials: Wood, Bone, Stone
	a.LastAIReasoning = "I must create... I need materials!"
	
	// Timeout (e.g. 10 minutes real time = 36,000 ticks)
	if a.WorkTicks > 36000 {
		a.Mood = MoodMelancholy
		if ctx.Log != nil {
			ctx.Log(fmt.Sprintf("%s has fallen into MELANCHOLY after failing a mood.", a.Name), LogNPC)
		}
		return
	}

	// Check if NPC has materials (simplified: 1 wood, 1 bone)
	hasWood := false
	hasBone := false
	for _, it := range a.Inventory {
		if strings.Contains(strings.ToLower(it.Config.ID), "wood") { hasWood = true }
		if strings.Contains(strings.ToLower(it.Config.ID), "bone") { hasBone = true }
	}

	if hasWood && hasBone {
		a.createArtifact(ctx)
	}
}

func (a *Actor) createArtifact(ctx *SystemContext) {
	a.Mood = MoodNeutral
	a.WorkTicks = 0
	
	// Create a Legendary Artifact
	artifactName := fmt.Sprintf("The Legendary %s of %s", "Relic", a.Name)
	
	// Spawn a high-stat item
	baseID := "iron_sword"
	if rand.Float64() < 0.5 { baseID = "bone_amulet" }
	
	config := ctx.Registries.Objects.Get(baseID)
	if config != nil {
		inst := &ItemInstance{
			ID:         fmt.Sprintf("artifact_%d", rand.Int()),
			Config:     config,
			Resistance: config.Resistance * 2, // Twice as durable
			Pickable:   true,
		}
		
		a.Inventory = append(a.Inventory, inst)
		
		if ctx.Log != nil {
			ctx.Log(fmt.Sprintf("%s has created %s!", a.Name, artifactName), LogNPC)
		}
		
		// Visual feedback
		ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{
			Text: "!!! ARTIFACT !!!", X: a.X, Y: a.Y - 1, Life: 120, Color: ColorHeal,
		})
	}
}
