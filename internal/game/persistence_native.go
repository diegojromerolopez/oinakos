//go:build !js || !wasm

package game

import "os"

func (g *Game) saveToLocalStorage(data []byte) error {
	return nil
}

func (g *Game) loadFromLocalStorage() ([]byte, error) {
	return nil, nil
}

func (g *Game) isWasm() bool {
	return false
}

func loadSettingsData() ([]byte, error) {
	return os.ReadFile(getSettingsPath())
}

func saveSettingsData(data []byte) error {
	return os.WriteFile(getSettingsPath(), data, 0644)
}

func (g *Game) CloseWindow() {
	os.Exit(0)
}
