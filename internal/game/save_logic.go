package game

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func (g *Game) Save(fpath string) error {
	bytes, err := g.serialize()
	if err != nil {
		DebugLog("Failed to serialize for save to %s: %v", fpath, err)
		return err
	}
	err = os.WriteFile(fpath, bytes, 0644)
	if err == nil {
		DebugLog("Game Successfully Saved to %s | NPCs: %d | Obstacles: %d", fpath, len(g.characters), len(g.obstacles))
	}
	return err
}

func (g *Game) performQuicksave() {
	if g.isWasm() {
		data, err := g.serialize()
		if err != nil {
			log.Printf("Failed to serialize game: %v", err)
			return
		}
		if err := g.saveToLocalStorage(data); err == nil {
			g.saveMessage = "Saved in Browser Storage"
			g.saveMessageTimer = 300
		}
		return
	}

	oinakosDir := GetOinakosDir()
	saveDir := filepath.Join(oinakosDir, "saves")
	if err := os.MkdirAll(saveDir, 0755); err == nil {
		savePath := filepath.Join(saveDir, fmt.Sprintf("quicksave-%s.oinakos.yaml", time.Now().Format("2006-01-02T150405")))
		if err := g.Save(savePath); err == nil {
			log.Printf("Game quicksaved: %s", savePath)
			g.lastSavePath = savePath
			absPath, err := filepath.Abs(savePath)
			if err != nil {
				absPath = savePath
			}
			g.saveMessage = "Saved in " + absPath
			g.saveMessageTimer = 300
		} else {
			log.Printf("Failed to quicksave: %v", err)
		}
	} else {
		log.Printf("Failed to create saves directory: %v", err)
	}
}

func (g *Game) serialize() ([]byte, error) {
	data := SaveData{}
	data.Map.ID = g.currentMapType.ID
	data.Map.WidthPixels = g.currentMapType.WidthPixels
	data.Map.HeightPixels = g.currentMapType.HeightPixels
	data.Map.Level = g.mapLevel
	data.Map.PlayTime = g.playTime
	data.Map.Heightmap = g.currentMapType.Heightmap
	for pt := range g.ExploredTiles {
		data.Map.ExploredTiles = append(data.Map.ExploredTiles, pt)
	}

	data.Player = PlayerSaveData{
		Archetype: g.playableCharacter.Config.ID,
		X:           g.playableCharacter.X,
		Y:           g.playableCharacter.Y,
		TemporalState: g.playableCharacter.TemporalState,
		XP:          g.playableCharacter.XP,
		Level:       g.playableCharacter.Level,
		Kills:       g.playableCharacter.Kills,
		MapKills:    g.playableCharacter.MapKills,
		
		PrimaryAttributes: g.playableCharacter.PrimaryAttributes,

		BaseAttack:  g.playableCharacter.BaseAttack,
		BaseDefense: g.playableCharacter.BaseDefense,
		BaseProtection: g.playableCharacter.BaseProtection,
		Submission:     g.playableCharacter.Submission,
		Denarii:     g.playableCharacter.Denarii,
		Inventory:   []ItemInstanceSaveData{},
		Slots:       make(map[string]ItemInstanceSaveData),
		SelectedModel: g.playableCharacter.SelectedModel,
	}
	for _, item := range g.playableCharacter.Inventory {
		if item == nil || item.Config == nil { continue }
		data.Player.Inventory = append(data.Player.Inventory, ItemInstanceSaveData{ID: item.Config.ID, Resistance: item.Resistance})
	}
	for slot, item := range g.playableCharacter.Slots {
		if item != nil && item.Config != nil {
			data.Player.Slots[slot] = ItemInstanceSaveData{ID: item.Config.ID, Resistance: item.Resistance}
		}
	}
	if g.playableCharacter.Weapon != nil {
		data.Player.Weapon = g.playableCharacter.Weapon
	}


	for _, n := range g.characters {
		if n.Config == nil {
			continue
		}
		behaviorStr := ""
		switch n.Behavior {
		case BehaviorWander:
			behaviorStr = "wander"
		case BehaviorPatrol:
			behaviorStr = "patrol"
		case BehaviorKnightHunter:
			behaviorStr = "hunter"
		case BehaviorNpcFighter:
			behaviorStr = "fighter"
		case BehaviorChaotic:
			behaviorStr = "chaotic"
		}

		npcSave := NPCSaveData{
			X:           n.X,
			Y:           n.Y,
			TemporalState: n.TemporalState,
			Level:       n.Level,
			Behavior:    behaviorStr,
			Name:        n.Name,
			Alignment:   n.Alignment,
			Group:       n.Group,
			LeaderID:    n.LeaderID,
			MustSurvive: n.MustSurvive,

			PrimaryAttributes: n.PrimaryAttributes,

			BaseAttack:  n.BaseAttack,
			BaseDefense: n.BaseDefense,
			BaseProtection: n.BaseProtection,
			Submission:     n.Submission,
			Denarii:     n.Denarii,
			SelectedModel: n.SelectedModel,
			Inventory:   []ItemInstanceSaveData{},
			Slots:       make(map[string]ItemInstanceSaveData),
		}
		for _, item := range n.Inventory {
			if item == nil || item.Config == nil { continue }
			npcSave.Inventory = append(npcSave.Inventory, ItemInstanceSaveData{ID: item.Config.ID, Resistance: item.Resistance})
		}
		for slot, item := range n.Slots {
			if item != nil && item.Config != nil {
				npcSave.Slots[slot] = ItemInstanceSaveData{ID: item.Config.ID, Resistance: item.Resistance}
			}
		}
		if n.Config != nil {
			if n.Config.Unique {
				npcSave.NPCID = n.Config.ID
			} else {
				npcSave.Archetype = n.Config.ID
			}
		}
		data.Characters = append(data.Characters, npcSave)
	}

	for _, o := range g.obstacles {
		if o.Archetype == nil {
			continue
		}
		xVal, yVal := o.X, o.Y
		data.Obstacles = append(data.Obstacles, ObstacleSaveData{
			ID:            o.ID,
			Archetype:   o.Archetype.ID,
			X:             &xVal,
			Y:             &yVal,
			HealthPoints:  o.HealthPoints,
			CooldownTicks: o.CooldownTicks,
		})
	}

	for _, it := range g.World.Items {
		if it == nil { continue }
		data.Items = append(data.Items, ItemSaveData{
			ID:         it.ID,
			Resistance: it.Resistance,
			X:          it.X,
			Y:          it.Y,
		})
	}

	return yaml.Marshal(data)
}
