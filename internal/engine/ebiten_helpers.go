//go:build !test

package engine

import (
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

func LoadSprite(assets fs.FS, path string, removeBg bool) Image {
	img, err := DecodeSpriteRaw(assets, path, removeBg)
	if err != nil || img == nil {
		return nil
	}
	return &EbitenImageWrapper{img: ebiten.NewImageFromImage(img)}
}

func toEbitenGeoM(m Matrix) ebiten.GeoM {
	var g ebiten.GeoM
	g.SetElement(0, 0, m.m[0][0])
	g.SetElement(0, 1, m.m[0][1])
	g.SetElement(0, 2, m.m[0][2])
	g.SetElement(1, 0, m.m[1][0])
	g.SetElement(1, 1, m.m[1][1])
	g.SetElement(1, 2, m.m[1][2])
	return g
}

func toEbitenColorScale(c ColorScale) ebiten.ColorScale {
	var g ebiten.ColorScale
	g.Scale(c.R, c.G, c.B, c.A)
	return g
}
