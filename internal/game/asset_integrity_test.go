package game

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAssetIntegrity_Archetypes(t *testing.T) {
	root := findProjectRoot()
	baseDirs := []string{"data/archetypes", "data/animals"}

	for _, baseDir := range baseDirs {
		dir := filepath.Join(root, baseDir)
		if _, err := os.Stat(dir); err != nil {
			continue
		}

		err := filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
			if err != nil { return err }
			if info.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
				return nil
			}

			t.Run(fpath, func(t *testing.T) {
				data, err := os.ReadFile(fpath)
				if err != nil { t.Fatalf("Failed to read file: %v", err) }

				var cfg EntityConfig
				if err := yaml.Unmarshal(data, &cfg); err != nil {
					t.Errorf("Failed to unmarshal %s: %v", fpath, err)
					return
				}

				relP, _ := filepath.Rel(dir, fpath)
				subDir := filepath.Dir(relP)
				if subDir == "." { subDir = "" }
				varN := strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))

				cat := "archetypes"
				if baseDir == "data/animals" { cat = "animals" }

				assetDir := path.Join(root, "assets/images", cat, subDir, varN)
				
				// Minimal required image
				staticPath := path.Join(assetDir, "static.png")
				if _, err := os.Stat(staticPath); err != nil {
					t.Errorf("Missing minimal required image: %s", staticPath)
				}

				// Check models if AssetDir exists
				modelDir := path.Join(assetDir, "models")
				if entries, err := os.ReadDir(modelDir); err == nil {
					for _, entry := range entries {
						if entry.IsDir() {
							mStaticPath := path.Join(modelDir, entry.Name(), "static.png")
							if _, err := os.Stat(mStaticPath); err != nil {
								t.Errorf("Model %s missing static.png: %s", entry.Name(), mStaticPath)
							}
						}
					}
				}
			})
			return nil
		})
		if err != nil { t.Fatalf("Walk failed: %v", err) }
	}
}

func TestAssetIntegrity_Objects(t *testing.T) {
	root := findProjectRoot()
	dir := filepath.Join(root, "data/objects")

	if _, err := os.Stat(dir); err != nil {
		t.Skip("Objects directory not found")
		return
	}

	err := filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
			return nil
		}

		t.Run(fpath, func(t *testing.T) {
			data, err := os.ReadFile(fpath)
			if err != nil { t.Fatalf("Failed to read file: %v", err) }

			var cfg ObjectConfig
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				t.Errorf("Failed to unmarshal %s: %v", fpath, err)
				return
			}

			if cfg.ID == "" {
				cfg.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
			}

			imagePath := path.Join(root, "assets/images/objects", cfg.ID+".png")
			if _, err := os.Stat(imagePath); err != nil {
				t.Errorf("Missing asset for object %s: expected %s", cfg.ID, imagePath)
			}
		})
		return nil
	})
	if err != nil { t.Fatalf("Walk failed: %v", err) }
}

func TestAssetIntegrity_Obstacles(t *testing.T) {
	root := findProjectRoot()
	dir := filepath.Join(root, "data/obstacles")

	if _, err := os.Stat(dir); err != nil {
		t.Skip("Obstacles directory not found")
		return
	}

	err := filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
			return nil
		}

		t.Run(fpath, func(t *testing.T) {
			data, err := os.ReadFile(fpath)
			if err != nil { t.Fatalf("Failed to read file: %v", err) }

			var cfg ObstacleArchetype
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				t.Errorf("Failed to unmarshal %s: %v", fpath, err)
				return
			}

			if cfg.ID == "" {
				cfg.ID = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
			}

			folderPath := path.Join(root, "assets/images/obstacles", cfg.ID)
			info, err := os.Stat(folderPath)
			if err == nil && info.IsDir() {
				// Folder exists, must have at least one valid state image
				// The engine uses closed.png as default if it exists.
				// If not, it fallback to <id>.png or fails.
				closedPath := path.Join(folderPath, "closed.png")
				imagePath := path.Join(root, "assets/images/obstacles", cfg.ID+".png")
				readyPath := path.Join(folderPath, "ready.png")
				growingPath := path.Join(folderPath, "growing.png")
				
				_, errClosed := os.Stat(closedPath)
				_, errImg := os.Stat(imagePath)
				_, errReady := os.Stat(readyPath)
				_, errGrowing := os.Stat(growingPath)
				
				if errClosed != nil && errImg != nil && errReady != nil && errGrowing != nil {
					t.Errorf("Missing default image for obstacle %s (checked %s, %s, %s and %s)", cfg.ID, closedPath, imagePath, readyPath, growingPath)
				}
			} else {
				imagePath := path.Join(root, "assets/images/obstacles", cfg.ID+".png")
				if _, err := os.Stat(imagePath); err != nil {
					t.Errorf("Missing image: %s", imagePath)
				}
			}
		})
		return nil
	})
	if err != nil { t.Fatalf("Walk failed: %v", err) }
}
