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
				return nil // Skip if not found in embedded
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

	// 2. Local oinakos/data overrides
	localBaseDir := filepath.Join("oinakos", baseDir)
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
	if assets == nil {
		return &EntityConfig{}, nil
	}
	const configPath = "data/characters/oinakos.yaml"
	localPath := filepath.Join("oinakos", configPath)

	var data []byte
	var err error

	// Try local override first
	if _, errStat := os.Stat(localPath); errStat == nil {
		data, err = os.ReadFile(localPath)
	}
	if data == nil {
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

	// Link weapon
	config.Weapon = GetWeaponByName(config.WeaponName)

	return &config, nil
}

func GetWeaponByName(name string) *Weapon {
	switch name {
	case "Tizon":
		return WeaponTizon
	case "Orcish Axe":
		return WeaponOrcishAxe
	case "Iron Broadsword":
		return WeaponIronBroadsword
	case "Fists":
		return WeaponFists
	case "Cleaver":
		return &Weapon{Name: "Cleaver", MinDamage: 3, MaxDamage: 7}
	case "Trident":
		return &Weapon{Name: "Trident", MinDamage: 6, MaxDamage: 12}
	case "Whip":
		return &Weapon{Name: "Whip", MinDamage: 4, MaxDamage: 9}
	case "Bow":
		return &Weapon{Name: "Bow", MinDamage: 3, MaxDamage: 6}
	case "Dagger":
		return &Weapon{Name: "Dagger", MinDamage: 2, MaxDamage: 5}
	case "Pitchfork":
		return &Weapon{Name: "Pitchfork", MinDamage: 4, MaxDamage: 10}
	case "Gilded Pitchfork":
		return &Weapon{Name: "Gilded Pitchfork", MinDamage: 8, MaxDamage: 18}
	case "Shouts":
		return &Weapon{Name: "Shouts", MinDamage: 10, MaxDamage: 50}
	default:
		// Return a basic unarmed weapon instead of nil to prevent panics
		return WeaponFists
	}
}
