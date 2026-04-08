package engine

import (
	_ "image/jpeg"
	_ "image/png"
)

type Renderer struct {
	grassOptions *DrawImageOptions
}

func NewRenderer() *Renderer {
	return &Renderer{
		grassOptions: NewDrawImageOptions(),
	}
}

func (r *Renderer) DrawTileMap(screen Image, offsetX, offsetY float64, getTile func(x, y int) Image, getZ func(x, y int) float64) {
	if getTile == nil || getZ == nil {
		return
	}

	screenWidth, screenHeight := screen.Size()

	// Convert camera center back to Cartesian to find the visible range
	camIsoX := float64(screenWidth)/2 - offsetX
	camIsoY := float64(screenHeight)/2 - offsetY
	camX, camY := IsoToCartesian(camIsoX, camIsoY)

	dim := 25
	minX := int(camX) - dim
	maxX := int(camX) + dim
	minY := int(camY) - dim
	maxY := int(camY) + dim

	var vertices []Vertex
	var indices []uint16

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			tileSprite := getTile(x, y)
			if tileSprite == nil {
				continue
			}

			// Corner elevations
			z00 := getZ(x, y)
			z10 := getZ(x+1, y)
			z01 := getZ(x, y+1)
			z11 := getZ(x+1, y+1)

			// Screen coordinates for the 4 corners
			p00X, p00Y := CartesianToIsoZ(float64(x), float64(y), z00)
			p10X, p10Y := CartesianToIsoZ(float64(x+1), float64(y), z10)
			p01X, p01Y := CartesianToIsoZ(float64(x), float64(y+1), z01)
			p11X, p11Y := CartesianToIsoZ(float64(x+1), float64(y+1), z11)

			tw, th := tileSprite.Size()
			ftw, fth := float32(tw), float32(th)

			// Add Offset
			p00X += offsetX; p00Y += offsetY
			p10X += offsetX; p10Y += offsetY
			p01X += offsetX; p01Y += offsetY
			p11X += offsetX; p11Y += offsetY

			// Simple culling (only if all points out)
			// (skipped for brevity/simplicity in this pass, but should be added for performance)

			// Shading per corner
			getShade := func(z float64) float32 {
				s := float32(1.0 + (z * 0.1))
				if s < 0.3 { s = 0.3 }; if s > 1.2 { s = 1.2 }
				return s
			}

			s00 := getShade(z00)
			s10 := getShade(z10)
			s01 := getShade(z01)
			s11 := getShade(z11)

			baseIdx := uint16(len(vertices))
			vertices = append(vertices,
				Vertex{DstX: float32(p00X), DstY: float32(p00Y), SrcX: 0, SrcY: 0, ColorR: s00, ColorG: s00, ColorB: s00, ColorA: 1},
				Vertex{DstX: float32(p10X), DstY: float32(p10Y), SrcX: ftw, SrcY: 0, ColorR: s10, ColorG: s10, ColorB: s10, ColorA: 1},
				Vertex{DstX: float32(p01X), DstY: float32(p01Y), SrcX: 0, SrcY: fth, ColorR: s01, ColorG: s01, ColorB: s01, ColorA: 1},
				Vertex{DstX: float32(p11X), DstY: float32(p11Y), SrcX: ftw, SrcY: fth, ColorR: s11, ColorG: s11, ColorB: s11, ColorA: 1},
			)
			indices = append(indices,
				baseIdx+0, baseIdx+1, baseIdx+2,
				baseIdx+1, baseIdx+2, baseIdx+3,
			)

			// To avoid huge buffers, we draw per-tile if sprite changes or every N tiles
			// But since we use getTile(x,y), the sprite might change every tile.
			// Ideally we batch by sprite. But for now, let's just DrawTriangles per tile.
			screen.DrawTriangles(vertices, indices, tileSprite, nil)
			vertices = vertices[:0]
			indices = indices[:0]
		}
	}
}
