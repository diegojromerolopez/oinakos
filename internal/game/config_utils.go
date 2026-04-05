package game

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

// forEachYAML iterates over YAML files in a base directory from both the embedded FS
// and the local oinakos/data override directory.
func forEachYAML(assets fs.FS, baseDir string, callback func(fpath string, data []byte) error) error {
	// 1. Embedded assets
	if assets != nil {
		fs.WalkDir(assets, baseDir, func(fpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
				return nil
			}
			data, err := fs.ReadFile(assets, fpath)
			if err == nil {
				callback(fpath, data)
			}
			return nil
		})
	}

	// 2. Local data overrides (with path discovery)
	foundPath := baseDir
	if _, err := os.Stat(foundPath); err != nil {
		// Try walking up to 3 levels up to find the project root
		candidate := filepath.Join("../..", baseDir)
		if _, err := os.Stat(candidate); err == nil {
			foundPath = candidate
		} else {
			candidate = filepath.Join("..", baseDir)
			if _, err := os.Stat(candidate); err == nil {
				foundPath = candidate
			}
		}
	}

	if _, err := os.Stat(foundPath); err == nil {
		filepath.WalkDir(foundPath, func(fpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
				return nil
			}
			data, err := os.ReadFile(fpath)
			if err == nil {
				callback(fpath, data)
			}
			return nil
		})
	}

	// 3. Optional 'oinakos/' prefixed local data...
	localBaseDir := filepath.Join("oinakos", baseDir)
	if _, err := os.Stat(localBaseDir); err != nil {
		candidate := filepath.Join("../..", "oinakos", baseDir)
		if _, err := os.Stat(candidate); err == nil {
			localBaseDir = candidate
		} else {
			candidate = filepath.Join("..", "oinakos", baseDir)
			if _, err := os.Stat(candidate); err == nil {
				localBaseDir = candidate
			}
		}
	}
	if _, err := os.Stat(localBaseDir); err == nil {
		filepath.WalkDir(localBaseDir, func(fpath string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || (filepath.Ext(fpath) != ".yaml" && filepath.Ext(fpath) != ".yml") {
				return nil
			}
			data, err := os.ReadFile(fpath)
			if err == nil {
				callback(fpath, data)
			}
			return nil
		})
	}
	return nil
}

func LoadPlayableCharacterConfig(assets fs.FS) (*EntityConfig, error) {
	const configPath = "data/characters/oinakos.yaml"
	localPath := filepath.Join("oinakos", configPath)

	var data []byte
	var err error

	// Try local override first
	if _, errStat := os.Stat(localPath); errStat == nil {
		data, err = os.ReadFile(localPath)
	}
	if data == nil && assets != nil {
		data, err = fs.ReadFile(assets, configPath)
	}
	if err != nil {
		// Fallback to direct OS read of regular path
		data, err = os.ReadFile(configPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read playable character config: %w", err)
	}

	var config EntityConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal playable character config %s: %w", configPath, err)
	}

	sanitizeEntityConfig(&config, configPath)

	// config.Weapon is now auto-loaded by YAML


	return &config, nil
}


