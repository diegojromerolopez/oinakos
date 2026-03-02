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

func (e *EbitenInput) IsKeyPressed(key Key) bool {
	return ebiten.IsKeyPressed(toEbitenKey(key))
}

func (e *EbitenInput) IsKeyJustPressed(key Key) bool {
	return inpututil.IsKeyJustPressed(toEbitenKey(key))
}

func (e *EbitenInput) AppendJustPressedKeys(keys []Key) []Key {
	ebKeys := inpututil.AppendJustPressedKeys(nil)
	for _, k := range ebKeys {
		ek := fromEbitenKey(k)
		if ek != -1 {
			keys = append(keys, ek)
		}
	}
	return keys
}

func (e *EbitenInput) MousePosition() (x, y int) {
	return ebiten.CursorPosition()
}

func (e *EbitenInput) IsMouseButtonJustPressed(button MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(toEbitenMouseButton(button))
}

func toEbitenMouseButton(button MouseButton) ebiten.MouseButton {
	switch button {
	case MouseButtonLeft:
		return ebiten.MouseButtonLeft
	case MouseButtonRight:
		return ebiten.MouseButtonRight
	}
	return ebiten.MouseButtonLeft
}

func NewEbitenInput() *EbitenInput {
	return &EbitenInput{}
}

func toEbitenKey(key Key) ebiten.Key {
	switch key {
	case KeyW:
		return ebiten.KeyW
	case KeyA:
		return ebiten.KeyA
	case KeyS:
		return ebiten.KeyS
	case KeyD:
		return ebiten.KeyD
	case KeySpace:
		return ebiten.KeySpace
	case KeyEscape:
		return ebiten.KeyEscape
	case KeyEnter:
		return ebiten.KeyEnter
	case KeyUp:
		return ebiten.KeyUp
	case KeyDown:
		return ebiten.KeyDown
	case KeyLeft:
		return ebiten.KeyLeft
	case KeyRight:
		return ebiten.KeyRight
	case KeyF9:
		return ebiten.KeyF9
	case KeyTab:
		return ebiten.KeyTab
	case KeyQ:
		return ebiten.KeyQ
	}
	return -1
}

func fromEbitenKey(key ebiten.Key) Key {
	switch key {
	case ebiten.KeyW:
		return KeyW
	case ebiten.KeyA:
		return KeyA
	case ebiten.KeyS:
		return KeyS
	case ebiten.KeyD:
		return KeyD
	case ebiten.KeySpace:
		return KeySpace
	case ebiten.KeyEscape:
		return KeyEscape
	case ebiten.KeyEnter:
		return KeyEnter
	case ebiten.KeyUp:
		return KeyUp
	case ebiten.KeyDown:
		return KeyDown
	case ebiten.KeyLeft:
		return KeyLeft
	case ebiten.KeyRight:
		return KeyRight
	case ebiten.KeyF9:
		return KeyF9
	case ebiten.KeyTab:
		return KeyTab
	case ebiten.KeyQ:
		return KeyQ
	}
	return -1
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

func (e *EbitenGraphics) DebugPrintAt(screen Image, str string, x, y int, clr color.Color) {
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

func (e *EbitenGraphics) DrawLine(screen Image, x1, y1, x2, y2 float32, clr color.Color, width float32) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if ok && wrapper != nil && wrapper.img != nil {
		vector.StrokeLine(wrapper.img, x1, y1, x2, y2, width, clr, true)
	}
}

func (e *EbitenGraphics) DrawPolygon(screen Image, points []Point, clr color.Color, width float32) {
	if len(points) < 2 {
		return
	}
	wrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || wrapper == nil || wrapper.img == nil {
		return
	}
	// We use ebitenutil.DrawLine for simplicity across versions
	for i := 0; i < len(points); i++ {
		p1 := points[i]
		p2 := points[(i+1)%len(points)]
		ebitenutil.DrawLine(wrapper.img, float64(p1.X), float64(p1.Y), float64(p2.X), float64(p2.Y), clr)
	}
}

type EbitenShaderWrapper struct {
	shader *ebiten.Shader
}

func (e *EbitenGraphics) NewShader(src []byte) (Shader, error) {
	s, err := ebiten.NewShader(src)
	if err != nil {
		return nil, err
	}
	return &EbitenShaderWrapper{shader: s}, nil
}

func (e *EbitenGraphics) DrawImageWithShader(screen Image, img Image, shader Shader, uniforms map[string]interface{}, options *DrawImageOptions) {
	screenWrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || screenWrapper == nil || screenWrapper.img == nil {
		return
	}
	imgWrapper, ok := img.(*EbitenImageWrapper)
	if !ok || imgWrapper == nil || imgWrapper.img == nil {
		return
	}
	shaderWrapper, ok := shader.(*EbitenShaderWrapper)
	if !ok || shaderWrapper == nil || shaderWrapper.shader == nil {
		return
	}

	w, h := imgWrapper.img.Size()
	op := &ebiten.DrawRectShaderOptions{
		Uniforms: uniforms,
	}
	if options != nil {
		op.GeoM = toEbitenGeoM(options.GeoM)
		op.ColorScale = toEbitenColorScale(options.ColorScale)
	}
	op.Images[0] = imgWrapper.img

	screenWrapper.img.DrawRectShader(w, h, shaderWrapper.shader, op)
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
