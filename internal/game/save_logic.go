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
		DebugLog("Game Successfully Saved to %s | NPCs: %d | Obstacles: %d", fpath, len(g.npcs), len(g.obstacles))
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
	}
	if g.playableCharacter.Weapon != nil {
		data.Player.Weapon = g.playableCharacter.Weapon.Name
	}

	for _, n := range g.npcs {
		if n.Archetype == nil {
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
		}
		if n.Archetype != nil {
			if n.Archetype.Unique {
				npcSave.NPCID = n.Archetype.ID
			} else {
				npcSave.ArchetypeID = n.Archetype.ID
			}
		}
		data.NPCs = append(data.NPCs, npcSave)
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

	return yaml.Marshal(data)
}
