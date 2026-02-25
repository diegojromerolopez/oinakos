package engine

import (
	"image"
	"image/color"
	"io/fs"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

	// Use a stack-allocated ebiten.DrawImageOptions to avoid heap allocation
	var op ebiten.DrawImageOptions
	if options != nil {
		scaleX := options.GeoM.m[0][0]
		scaleY := options.GeoM.m[1][1]
		tx := options.GeoM.m[0][2]
		ty := options.GeoM.m[1][2]

		// If there is a flip/scale, ensure both exist in the matrix representation properly
		op.GeoM.Scale(scaleX, scaleY)
		op.GeoM.Translate(tx, ty)
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

	// Use a small stack-allocated buffer for vertices to avoid heap allocation for common cases (3-4 vertices)
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

func (w *EbitenImageWrapper) Fill(clr interface{}) {
	c, ok := clr.(color.Color)
	if ok && w.img != nil {
		w.img.Fill(c)
	}
}

func (w *EbitenImageWrapper) UpdateRaw(img *ebiten.Image) {
	w.img = img
}

// GetRaw returns the underlying Ebiten image (useful for main.go)
func (w *EbitenImageWrapper) GetRaw() *ebiten.Image {
	return w.img
}

// EbitenInput implements engine.Input using the actual ebiten package.
type EbitenInput struct{}

func (e *EbitenInput) IsKeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

func (e *EbitenInput) IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}

func (e *EbitenInput) AppendJustPressedKeys(keys []ebiten.Key) []ebiten.Key {
	return inpututil.AppendJustPressedKeys(keys)
}

func NewEbitenInput() *EbitenInput {
	return &EbitenInput{}
}

// EbitenGraphics implements engine.Graphics using Ebiten's utility functions.
type EbitenGraphics struct {
	whiteImage *ebiten.Image
}

func NewEbitenGraphics() *EbitenGraphics {
	whiteImg := ebiten.NewImage(3, 3)
	whiteImg.Fill(color.White)
	return &EbitenGraphics{
		whiteImage: whiteImg,
	}
}

func (e *EbitenGraphics) NewImage(width, height int) Image {
	return &EbitenImageWrapper{img: ebiten.NewImage(width, height)}
}

func (e *EbitenGraphics) NewImageFromImage(img image.Image) Image {
	if img == nil {
		return nil
	}
	return &EbitenImageWrapper{img: ebiten.NewImageFromImage(img)}
}

func (e *EbitenGraphics) DebugPrintAt(screen Image, str string, x, y int) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if ok && wrapper != nil && wrapper.img != nil {
		ebitenutil.DebugPrintAt(wrapper.img, str, x, y)
	}
}

func (e *EbitenGraphics) LoadSprite(assets fs.FS, path string, removeBg bool) Image {
	return LoadSprite(assets, path, removeBg)
}

func (e *EbitenGraphics) DrawFilledRect(screen Image, x, y, width, height float32, clr color.Color, antiAlias bool) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if ok && wrapper != nil && wrapper.img != nil {
		if e.whiteImage == nil {
			vector.DrawFilledRect(wrapper.img, x, y, width, height, clr, antiAlias)
			return
		}
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(float64(width)/3.0, float64(height)/3.0)
		op.GeoM.Translate(float64(x), float64(y))
		op.ColorScale.ScaleWithColor(clr)
		wrapper.img.DrawImage(e.whiteImage, &op)
	}
}

func (e *EbitenGraphics) DrawFilledCircle(screen Image, x, y, radius float32, clr color.Color, antiAlias bool) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if ok && wrapper != nil && wrapper.img != nil {
		vector.DrawFilledCircle(wrapper.img, x, y, radius, clr, antiAlias)
	}
}

func (e *EbitenGraphics) DrawTriangles(screen Image, vertices []Vertex, indices []uint16, src Image, options *DrawTrianglesOptions) {
	screenWrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || screenWrapper == nil || screenWrapper.img == nil {
		return
	}
	srcWrapper, ok := src.(*EbitenImageWrapper)
	if !ok || srcWrapper == nil || srcWrapper.img == nil {
		return
	}

	evs := make([]ebiten.Vertex, len(vertices))
	for i, v := range vertices {
		evs[i] = ebiten.Vertex{
			DstX:   v.DstX,
			DstY:   v.DstY,
			SrcX:   v.SrcX,
			SrcY:   v.SrcY,
			ColorR: v.ColorR,
			ColorG: v.ColorG,
			ColorB: v.ColorB,
			ColorA: v.ColorA,
		}
	}

	op := &ebiten.DrawTrianglesOptions{}
	if options != nil {
		op.FillRule = ebiten.FillRule(options.FillRule)
	}

	screenWrapper.img.DrawTriangles(evs, indices, srcWrapper.img, op)
}

func LoadSprite(assets fs.FS, path string, removeBg bool) Image {
	var f fs.File
	var err error

	if assets != nil {
		f, err = assets.Open(path)
	}

	if err != nil || assets == nil {
		f, err = os.Open(path)
	}

	if err != nil {
		log.Printf("Warning: failed to open sprite '%s': %v", path, err)
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Warning: failed to decode sprite '%s': %v", path, err)
		return nil
	}

	if removeBg {
		img = Transparentize(img)
	}

	return &EbitenImageWrapper{img: ebiten.NewImageFromImage(img)}
}
