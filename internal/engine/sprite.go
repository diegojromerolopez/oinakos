package engine

import (
	"image"
	"io/fs"
	"log"
	"os"

	// Register PNG decoder
	_ "image/png"
)

// DecodeSpriteRaw opens a file (preferring local oinakos/ override if present),
// decodes it as an image.Image, and optionally applies the green-screen transparentizer.
// It is thread-safe and safe to call from background goroutines as it does not rely on Ebiten context.
func DecodeSpriteRaw(assets fs.FS, path string, removeBg bool) (image.Image, error) {
	var f fs.File
	var err error

	// Check for local override in oinakos/assets folder
	localPath := "oinakos/" + path
	if _, statErr := os.Stat(localPath); statErr == nil {
		f, err = os.Open(localPath)
	} else {
		if assets != nil {
			f, err = assets.Open(path)
		}
		if f == nil {
			f, err = os.Open(path)
		}
	}

	if err != nil {
		if !os.IsNotExist(err) && err != fs.ErrNotExist {
			log.Printf("Warning: failed to open raw sprite '%s': %v", path, err)
		}
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Warning: failed to decode raw sprite '%s': %v", path, err)
		return nil, err
	}

	if removeBg {
		img = Transparentize(img)
	}

	return img, nil
}
