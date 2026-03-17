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
	for pt := range g.ExploredTiles {
		data.Map.ExploredTiles = append(data.Map.ExploredTiles, pt)
	}

	data.Player = PlayerSaveData{
		ArchetypeID: g.playableCharacter.Config.ID,
		X:           g.playableCharacter.X,
		Y:           g.playableCharacter.Y,
		Health:      g.playableCharacter.Health,
		MaxHealth:   g.playableCharacter.MaxHealth,
		XP:          g.playableCharacter.XP,
		Level:       g.playableCharacter.Level,
		Kills:       g.playableCharacter.Kills,
		MapKills:    g.playableCharacter.MapKills,
		BaseAttack:  g.playableCharacter.BaseAttack,
		BaseDefense: g.playableCharacter.BaseDefense,
		Inventory:   []string{},
		Slots:       make(map[string]string),
	}
	for _, item := range g.playableCharacter.Inventory {
		data.Player.Inventory = append(data.Player.Inventory, item.ID)
	}
	for slot, item := range g.playableCharacter.Slots {
		if item != nil {
			data.Player.Slots[slot] = item.ID
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
			Health:      n.Health,
			MaxHealth:   n.MaxHealth,
			Level:       n.Level,
			Behavior:    behaviorStr,
			Name:        n.Name,
			Alignment:   n.Alignment,
			Group:       n.Group,
			LeaderID:    n.LeaderID,
			MustSurvive: n.MustSurvive,
			BaseAttack:  n.BaseAttack,
			BaseDefense: n.BaseDefense,
			Inventory:   []string{},
			Slots:       make(map[string]string),
		}
		for _, item := range n.Inventory {
			npcSave.Inventory = append(npcSave.Inventory, item.ID)
		}
		for slot, item := range n.Slots {
			if item != nil {
				npcSave.Slots[slot] = item.ID
			}
		}
		if n.Config != nil {
			if n.Config.Unique {
				npcSave.NPCID = n.Config.ID
			} else {
				npcSave.ArchetypeID = n.Config.ID
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
			ArchetypeID:   o.Archetype.ID,
			X:             &xVal,
			Y:             &yVal,
			Health:        o.Health,
			CooldownTicks: o.CooldownTicks,
		})
	}

	for _, it := range g.World.Items {
		data.Items = append(data.Items, ItemSaveData{
			ID: it.ID,
			X:  it.X,
			Y:  it.Y,
		})
	}

	return yaml.Marshal(data)
}
