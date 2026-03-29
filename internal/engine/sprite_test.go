package engine

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/fstest"
)

func TestDecodeSpriteRaw(t *testing.T) {
	// Create a simple red 2x2 image
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatal(err)
	}
	
	// Mock FS
	assets := fstest.MapFS{
		"test_sprite.png": {Data: buf.Bytes()},
	}

	decoded, err := DecodeSpriteRaw(assets, "test_sprite.png", false)
	if err != nil {
		t.Errorf("Expected successful decode, got error: %v", err)
	}
	if decoded == nil {
		t.Fatal("Expected non-nil image")
	}
	if decoded.Bounds().Dx() != 2 || decoded.Bounds().Dy() != 2 {
		t.Errorf("Expected 2x2 image, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
	
	// Test error
	_, err = DecodeSpriteRaw(assets, "missing.png", false)
	if err == nil {
		t.Errorf("Expected error for missing file, got nil")
	}
}

func TestDecodeSpriteRaw_Transparentize(t *testing.T) {
	// Create a simple image with 00FF00
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	
	assets := fstest.MapFS{
		"green.png": {Data: buf.Bytes()},
	}

	decoded, err := DecodeSpriteRaw(assets, "green.png", true)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	
	// Check one pixel
	c := decoded.At(0, 0)
	_, _, _, a := c.RGBA()
	if a != 0 {
		t.Errorf("Expected green pixel to be transparent, got alpha %v", a)
	}
}
