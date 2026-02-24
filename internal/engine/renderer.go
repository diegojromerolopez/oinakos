package engine

import (
	_ "image/jpeg"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) DrawInfiniteGrass(screen *ebiten.Image, offsetX, offsetY float64, grassSprite *ebiten.Image) {
	if grassSprite == nil {
		return
	}

	screenWidth, screenHeight := screen.Size()

	// Convert camera center back to Cartesian to find the visible range
	// Camera offset = screen/2 - cameraPos
	// cameraPos = screen/2 - offset
	camIsoX := float64(screenWidth)/2 - offsetX
	camIsoY := float64(screenHeight)/2 - offsetY
	camX, camY := IsoToCartesian(camIsoX, camIsoY)

	// Determine range of tiles to draw.
	// Isometric views need a wider range due to the diamond shape.
	// 50x50 tiles roughly covers most reasonable screen sizes.
	dim := 25
	minX := int(camX) - dim
	maxX := int(camX) + dim
	minY := int(camY) - dim
	maxY := int(camY) + dim

	sw, sh := grassSprite.Size()
	scaleX, scaleY := 64.0/float64(sw), 64.0/float64(sh)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			isoX, isoY := CartesianToIso(float64(x), float64(y))

			drawX := isoX + offsetX
			drawY := isoY + offsetY

			// Simple culling
			if drawX < -128 || drawX > float64(screenWidth)+128 || drawY < -128 || drawY > float64(screenHeight)+128 {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(drawX-32, drawY)
			screen.DrawImage(grassSprite, op)
		}
	}
}
