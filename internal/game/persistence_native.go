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

func (g *Game) CloseWindow() {
	os.Exit(0)
}
