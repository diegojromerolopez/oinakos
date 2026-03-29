package game

import (
	"math"
	"math/rand"
	"oinakos/internal/engine"
)

func (wm *WorldManager) UpdateNPCSpawning() {
	g := wm.game
	if len(g.currentMapType.Spawns) > 0 {
		for i := range g.currentMapType.Spawns {
			s := &g.currentMapType.Spawns[i]
			if s.Frequency <= 0 { continue }
			s.Timer++
			if s.Timer >= int(s.Frequency*60) {
				s.Timer = 0
				if rand.Float64() <= s.Probability {
					if len(g.characters) < 100 {
						if s.X != nil && s.Y != nil { wm.spawnNPCNearPosition(*s.X, *s.Y, s) } else { wm.spawnNPCAtMapEdges(s) }
					}
				}
			}
		}
	}

	g.npcSpawnTimer++
	if g.npcSpawnTimer >= 300 {
		g.npcSpawnTimer = 0
		wm.AssignPersonalChests()
		activeNPCs := make([]*Character, 0)
		for _, n := range g.characters {
			dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
			if n.IsAlive() {
				if dist < 40 { activeNPCs = append(activeNPCs, n) }
			} else { activeNPCs = append(activeNPCs, n) }
		}
		g.characters = activeNPCs
	}
}

func (wm *WorldManager) spawnNPCNearPosition(x, y float64, sc *SpawnConfig) {
	g := wm.game
	if len(g.archetypeRegistry.IDs) == 0 || sc == nil { return }
	npcConfig := g.archetypeRegistry.Archetypes[sc.Archetype]
	if npcConfig == nil { return }
	if rand.Float64() < 0.05 {
		var variants []*EntityConfig
		for _, v := range g.characterRegistry.Characters {
			if v.Archetype == sc.Archetype && !v.Unique { variants = append(variants, v) }
		}
		if len(variants) > 0 { npcConfig = variants[rand.Intn(len(variants))] }
	}
	npc := NewCharacter(x, y, npcConfig, g.mapLevel, false, g.Registries.Objects)
	npc.Alignment = sc.Alignment
	for i := 0; i < 10; i++ {
		collides := false
		for _, o := range g.obstacles {
			if o.Alive && engine.CheckCirclePolygonCollision(npc.GetCollisionCircle(), o.GetFootprint()) { collides = true; break }
		}
		if !collides { break }
		angle := rand.Float64() * 2 * math.Pi
		npc.X = x + math.Cos(angle)*(2.0+rand.Float64()); npc.Y = y + math.Sin(angle)*(2.0+rand.Float64())
	}
	npc.LoadEquipment(g.Registries.Objects)
	g.characters = append(g.characters, npc)
}

func (wm *WorldManager) spawnNPCAtMapEdges(sc *SpawnConfig) {
	g := wm.game
	if len(g.archetypeRegistry.IDs) == 0 || sc == nil { return }
	npcConfig := g.archetypeRegistry.Archetypes[sc.Archetype]
	if npcConfig == nil { return }
	if rand.Float64() < 0.05 {
		var variants []*EntityConfig
		for _, v := range g.characterRegistry.Characters {
			if v.Archetype == sc.Archetype && !v.Unique { variants = append(variants, v) }
		}
		if len(variants) > 0 { npcConfig = variants[rand.Intn(len(variants))] }
	}
	angle := rand.Float64() * 2 * math.Pi
	npc := NewCharacter(g.playableCharacter.X+math.Cos(angle)*30, g.playableCharacter.Y+math.Sin(angle)*30, npcConfig, g.mapLevel, false, g.Registries.Objects)
	npc.Alignment = sc.Alignment
	for i := 0; i < 10; i++ {
		collides := false
		for _, o := range g.obstacles {
			if o.Alive && engine.CheckCirclePolygonCollision(npc.GetCollisionCircle(), o.GetFootprint()) { collides = true; break }
		}
		if !collides { break }
		angle := rand.Float64() * 2 * math.Pi
		npc.X = g.playableCharacter.X + math.Cos(angle)*(30.0+rand.Float64()*2)
		npc.Y = g.playableCharacter.Y + math.Sin(angle)*(30.0+rand.Float64()*2)
	}
	npc.LoadEquipment(g.Registries.Objects)
	g.characters = append(g.characters, npc)
}

func (wm *WorldManager) AssignPersonalChests() {
	g := wm.game
	assignedChests := make(map[string]bool)
	for _, c := range g.characters {
		if c.OwnedChestID != "" { assignedChests[c.OwnedChestID] = true }
	}

	for _, c := range g.characters {
		if c.OwnedChestID != "" || !c.IsAlive() || c.Alignment != AlignmentNeutral { continue }
		
		// Find nearest unassigned chest
		for _, o := range g.obstacles {
			if !o.Alive || o.Archetype == nil || o.Archetype.ID != "personal_chest" { continue }
			if assignedChests[o.ID] { continue }
			
			dist := math.Sqrt(math.Pow(c.X-o.X, 2) + math.Pow(c.Y-o.Y, 2))
			if dist < 40.0 { 
				c.OwnedChestID = o.ID
				assignedChests[o.ID] = true
				DebugLog("NPC [%s] has claimed chest [%s]", c.Name, c.OwnedChestID)
				break 
			}
		}
	}
}
