package game

import (
	"fmt"
	"math/rand"
)

// WorldGenerator handles seed-based procedural additions to the world.
type WorldGenerator struct {
	game *Game
	rnd  *rand.Rand
}

func NewWorldGenerator(g *Game, seed int64) *WorldGenerator {
	return &WorldGenerator{
		game: g,
		rnd:  rand.New(rand.NewSource(seed)),
	}
}

// GenerateVillage creates a cluster of buildings and NPCs at a given location.
func (wg *WorldGenerator) GenerateVillage(centerX, centerY float64, radius float64) {
	g := wg.game
	
	// 1. Houses
	houseConfig, hasHouse := g.obstacleRegistry.Archetypes["shop_tent"]
	if !hasHouse {
		houseConfig, hasHouse = g.obstacleRegistry.Archetypes["warehouse"]
	}
	
	if hasHouse {
		numHouses := 3 + wg.rnd.Intn(4)
		for i := 0; i < numHouses; i++ {
			hx := centerX + (wg.rnd.Float64()*radius*2 - radius)
			hy := centerY + (wg.rnd.Float64()*radius*2 - radius)
			
			// Slightly randomize house type if possible
			g.obstacles = append(g.obstacles, NewObstacle(fmt.Sprintf("v_house_%d", i), hx, hy, houseConfig))
		}
	}

	// 2. Villagers
	villagerArch, hasVillager := g.archetypeRegistry.Archetypes["vampire_male"] // Fallback
	if !hasVillager {
		if len(g.archetypeRegistry.IDs) > 0 {
			villagerArch = g.archetypeRegistry.Archetypes[g.archetypeRegistry.IDs[0]]
			hasVillager = true
		}
	}

	if hasVillager {
		numVillagers := 5 + wg.rnd.Intn(5)
		for i := 0; i < numVillagers; i++ {
			vx := centerX + (wg.rnd.Float64()*radius*2 - radius)
			vy := centerY + (wg.rnd.Float64()*radius*2 - radius)
			
			npc := NewCharacter(vx, vy, villagerArch, g.mapLevel, false, g.Registries.Objects)
			npc.Alignment = AlignmentNeutral
			npc.Behavior = BehaviorWander
			g.characters = append(g.characters, npc)
		}
	}
	
	// 3. Campfire
	fireConfig, hasFire := g.obstacleRegistry.Archetypes["campfire"]
	if hasFire {
		g.obstacles = append(g.obstacles, NewObstacle("v_campfire", centerX, centerY, fireConfig))
	}
}

// SeedResources clusters items like wood or minerals.
func (wg *WorldGenerator) SeedResources(centerX, centerY float64, resourceID string, count int) {
	g := wg.game
	if g.Registries == nil || g.Registries.Objects == nil { return }
	cfg, ok := g.Registries.Objects.Objects[resourceID]
	if !ok { return }

	for i := 0; i < count; i++ {
		rx := centerX + (wg.rnd.Float64()*10 - 5)
		ry := centerY + (wg.rnd.Float64()*10 - 5)
		
		it := &ItemInstance{
			ID:       fmt.Sprintf("res_%s_%d", resourceID, wg.rnd.Int()),
			Config:   cfg,
			X:        rx,
			Y:        ry,
			Pickable: true,
		}
		if g.World != nil {
			g.World.Items = append(g.World.Items, it)
		}
	}
}
