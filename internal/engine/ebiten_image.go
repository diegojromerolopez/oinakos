//go:build !test

package engine

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// EbitenImageWrapper implements Image using an actual *ebiten.Image
type EbitenImageWrapper struct {
	img *ebiten.Image
}

func NewEbitenImageWrapper(img *ebiten.Image) *EbitenImageWrapper {
	return &EbitenImageWrapper{img: img}
}

func (w *EbitenImageWrapper) Size() (int, int) {
	if w.img == nil {
		return 0, 0
	}
	return w.img.Size()
}

func (w *EbitenImageWrapper) Bounds() image.Rectangle {
	if w.img == nil {
		return image.Rectangle{}
	}
	return w.img.Bounds()
}

func (w *EbitenImageWrapper) DrawImage(img Image, options *DrawImageOptions) {
	if img == nil {
		return
	}
	wrapper, ok := img.(*EbitenImageWrapper)
	if !ok || wrapper == nil || wrapper.img == nil {
		return
	}

	var op ebiten.DrawImageOptions
	if options != nil {
		scaleX := options.GeoM.m[0][0]
		scaleY := options.GeoM.m[1][1]
		tx := options.GeoM.m[0][2]
		ty := options.GeoM.m[1][2]

		op.GeoM.Scale(scaleX, scaleY)
		op.GeoM.Translate(tx, ty)

		switch options.Blend {
		case BlendSourceOver:
			op.Blend = ebiten.BlendSourceOver
		case BlendDestinationOut:
			op.Blend = ebiten.BlendDestinationOut
		case BlendDestinationIn:
			op.Blend = ebiten.BlendDestinationIn
		}
	}
	w.img.DrawImage(wrapper.img, &op)
}

func (w *EbitenImageWrapper) DrawTriangles(vertices []Vertex, indices []uint16, src Image, options *DrawTrianglesOptions) {
	if w.img == nil || src == nil {
		return
	}
	srcWrapper, ok := src.(*EbitenImageWrapper)
	if !ok || srcWrapper == nil || srcWrapper.img == nil {
		return
	}

	var staticVs [4]ebiten.Vertex
	var evs []ebiten.Vertex
	if len(vertices) <= 4 {
		evs = staticVs[:len(vertices)]
	} else {
		evs = make([]ebiten.Vertex, len(vertices))
	}

	for i, v := range vertices {
		evs[i].DstX = v.DstX
		evs[i].DstY = v.DstY
		evs[i].SrcX = v.SrcX
		evs[i].SrcY = v.SrcY
		evs[i].ColorR = v.ColorR
		evs[i].ColorG = v.ColorG
		evs[i].ColorB = v.ColorB
		evs[i].ColorA = v.ColorA
	}

	var op ebiten.DrawTrianglesOptions
	if options != nil {
		op.FillRule = ebiten.FillRule(options.FillRule)
	}

	w.img.DrawTriangles(evs, indices, srcWrapper.img, &op)
}

func (w *EbitenImageWrapper) SubImage(r image.Rectangle) Image {
	sub := w.img.SubImage(r)
	if sub == nil {
		return nil
	}
	return &EbitenImageWrapper{img: sub.(*ebiten.Image)}
}

func (w *EbitenImageWrapper) Clear() {
	if w.img != nil {
		w.img.Clear()
	}
}

func (w *EbitenImageWrapper) Fill(clr color.Color) {
	if w.img != nil {
		w.img.Fill(clr)
	}
}

func (w *EbitenImageWrapper) UpdateRaw(img *ebiten.Image) {
	w.img = img
}

func (w *EbitenImageWrapper) GetRaw() *ebiten.Image {
	return w.img
}
