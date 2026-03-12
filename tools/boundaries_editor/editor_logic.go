package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"oinakos/internal/game"
)

func (v *Viewer) addPoint(ee *EditorEntity) {
	newP := game.FootprintPoint{}
	if len(*ee.Footprint) > 0 {
		last := (*ee.Footprint)[len(*ee.Footprint)-1]
		newP.X = last.X + 0.5
		newP.Y = last.Y + 0.5
	}
	*ee.Footprint = append(*ee.Footprint, newP)
	v.saveToYAML(ee)
}

func (v *Viewer) removePoint(ee *EditorEntity, idx int) {
	fp := *ee.Footprint
	if len(fp) <= 3 {
		log.Println("Cannot remove: polygon must have at least 3 vertices.")
		return
	}
	*ee.Footprint = append(fp[:idx], fp[idx+1:]...)
	v.saveToYAML(ee)
}

func (v *Viewer) saveToYAML(ee *EditorEntity) {
	if ee.YamlPath == "" { return }
	data, err := os.ReadFile(ee.YamlPath)
	if err != nil {
		log.Printf("failed to read yaml: %v", err)
		return
	}
	var m yaml.Node
	if err := yaml.Unmarshal(data, &m); err != nil {
		log.Printf("failed to unmarshal yaml: %v", err)
		return
	}
	fpData, _ := yaml.Marshal(*ee.Footprint)
	var fpNode yaml.Node
	yaml.Unmarshal(fpData, &fpNode)
	if m.Content[0].Kind == yaml.MappingNode {
		found := false
		for i := 0; i < len(m.Content[0].Content); i += 2 {
			if m.Content[0].Content[i].Value == "footprint" {
				m.Content[0].Content[i+1] = fpNode.Content[0]
				found = true
				break
			}
		}
		if !found {
			m.Content[0].Content = append(m.Content[0].Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "footprint"},
				fpNode.Content[0],
			)
		}
	}
	f, err := os.Create(ee.YamlPath)
	if err != nil {
		log.Printf("failed to write yaml: %v", err)
		return
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	enc.Encode(&m)
	log.Println("Footprint saved to", ee.YamlPath)
}

func findArchetypeYAML(id string) string {
	baseDir := "data/archetypes"
	var found string
	localBaseDir := filepath.Join("oinakos", baseDir)
	if _, statErr := os.Stat(localBaseDir); statErr == nil {
		filepath.WalkDir(localBaseDir, func(fpath string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() { return nil }
			if filepath.Ext(fpath) == ".yaml" || filepath.Ext(fpath) == ".yml" {
				data, err := os.ReadFile(fpath)
				if err == nil && containsID(data, id) {
					found = fpath
					return filepath.SkipAll
				}
			}
			return nil
		})
	}
	if found != "" { return found }
	filepath.WalkDir(baseDir, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() { return nil }
		if filepath.Ext(fpath) == ".yaml" || filepath.Ext(fpath) == ".yml" {
			data, err := os.ReadFile(fpath)
			if err == nil && containsID(data, id) {
				found = fpath
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

func containsID(data []byte, id string) bool {
	var m struct { ID string `yaml:"id"` }
	yaml.Unmarshal(data, &m)
	return m.ID == id
}
