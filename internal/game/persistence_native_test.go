//go:build !js || !wasm

package game

import "testing"

func TestPersistenceNative(t *testing.T) {
	g := &Game{}
	if g.isWasm() {
		t.Error("isWasm() should be false in native build")
	}
	if err := g.saveToLocalStorage(nil); err != nil {
		t.Errorf("saveToLocalStorage should not fail, got %v", err)
	}
	data, err := g.loadFromLocalStorage()
	if err != nil || data != nil {
		t.Errorf("loadFromLocalStorage should return nil, nil; got %v, %v", data, err)
	}
}
