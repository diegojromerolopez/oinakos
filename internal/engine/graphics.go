package engine

import (
	"image"
	"image/color"
	"io/fs"
)

// Graphics is a high-level interface for graphics operations and image creation.
type Graphics interface {
	VectorRenderer
	TextRenderer
	NewImage(width, height int) Image
	NewImageFromImage(img image.Image) Image
	LoadSprite(assets fs.FS, path string, removeBg bool) Image
}

// VectorRenderer defines basic primitive drawing operations.
type VectorRenderer interface {
	DrawFilledRect(screen Image, x, y, width, height float32, clr color.Color, antiAlias bool)
	DrawFilledCircle(screen Image, x, y, radius float32, clr color.Color, antiAlias bool)
}

// TextRenderer defines text drawing operations.
type TextRenderer interface {
	DebugPrintAt(screen Image, str string, x, y int)
}
