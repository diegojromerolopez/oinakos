//go:build js && wasm

package game

import (
	"syscall/js"
)

func (g *Game) saveToLocalStorage(data []byte) error {
	js.Global().Get("localStorage").Call("setItem", "quicksave", string(data))
	return nil
}

func (g *Game) loadFromLocalStorage() ([]byte, error) {
	val := js.Global().Get("localStorage").Call("getItem", "quicksave")
	if val.IsNull() {
		return nil, nil // Return nil, nil to indicate no save data found
	}
	return []byte(val.String()), nil
}

func (g *Game) isWasm() bool {
	return true
}

func (g *Game) CloseWindow() {
	js.Global().Get("window").Call("close")
}
