//go:build !test

package engine

import (
	"bytes"
	"image"
	"image/color"
	"io/fs"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// EbitenGraphics implements engine.Graphics using Ebiten's utility functions.
type EbitenGraphics struct {
	whiteImage  *ebiten.Image
	circleImage *ebiten.Image
	debugTxtImg *ebiten.Image
	fontSource  *text.GoTextFaceSource
}

func NewEbitenGraphics() *EbitenGraphics {
	whiteImg := ebiten.NewImage(3, 3)
	whiteImg.Fill(color.White)

	circleImg := ebiten.NewImage(128, 128)
	vector.DrawFilledCircle(circleImg, 64, 64, 64, color.White, true)

	return &EbitenGraphics{
		whiteImage:  whiteImg,
		circleImage: circleImg,
		debugTxtImg: ebiten.NewImage(256, 32),
	}
}

func (e *EbitenGraphics) NewImage(width, height int) Image {
	return &EbitenImageWrapper{img: ebiten.NewImage(width, height)}
}

func (e *EbitenGraphics) NewImageFromImage(img image.Image) Image {
	if img == nil { return nil }
	return &EbitenImageWrapper{img: ebiten.NewImageFromImage(img)}
}

func (e *EbitenGraphics) DebugPrintAt(screen Image, str string, x, y int, clr color.Color) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || wrapper == nil || wrapper.img == nil { return }

	r, g, b, a := clr.RGBA()
	if r == 0xffff && g == 0xffff && b == 0xffff && a == 0xffff {
		ebitenutil.DebugPrintAt(wrapper.img, str, x, y)
		return
	}

	neededW := len(str) * 6
	if neededW > e.debugTxtImg.Bounds().Dx() {
		ebitenutil.DebugPrintAt(wrapper.img, str, x, y)
		return
	}

	e.debugTxtImg.Clear()
	ebitenutil.DebugPrint(e.debugTxtImg, str)

	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)

	sub := e.debugTxtImg.SubImage(image.Rect(0, 0, neededW, 16)).(*ebiten.Image)
	wrapper.img.DrawImage(sub, &op)
}

func (e *EbitenGraphics) LoadFont(assets fs.FS, path string) error {
	if path == "" {
		e.fontSource = nil
		return nil
	}
	data, err := fs.ReadFile(assets, path)
	if err != nil {
		data, err = os.ReadFile(path)
	}
	if err != nil { return err }

	s, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil { return err }
	e.fontSource = s
	return nil
}

func (e *EbitenGraphics) DrawTextAt(screen Image, str string, x, y int, clr color.Color, size float64) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || wrapper == nil || wrapper.img == nil { return }

	if e.fontSource == nil {
		e.DebugPrintAt(screen, str, x, y, clr)
		return
	}

	face := &text.GoTextFace{ Source: e.fontSource, Size: size }
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(wrapper.img, str, face, op)
}

func (e *EbitenGraphics) MeasureText(str string, size float64) (float64, float64) {
	if e.fontSource == nil { return float64(len(str) * 6), 16 }
	face := &text.GoTextFace{ Source: e.fontSource, Size: size }
	return text.Measure(str, face, 0)
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

func (e *EbitenGraphics) DrawFilledEllipse(screen Image, x, y, rx, ry float32, clr color.Color, antiAlias bool) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || wrapper == nil || wrapper.img == nil || e.circleImage == nil { return }

	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(rx)/64.0, float64(ry)/64.0)
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)
	wrapper.img.DrawImage(e.circleImage, &op)
}

func (e *EbitenGraphics) DrawEllipse(screen Image, x, y, rx, ry float32, clr color.Color, width float32, antiAlias bool) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || wrapper == nil || wrapper.img == nil { return }

	const segments = 32
	for i := 0; i < segments; i++ {
		a1 := float64(i) * 2 * math.Pi / segments
		a2 := float64(i+1) * 2 * math.Pi / segments
		x1 := x + float32(math.Cos(a1))*rx
		y1 := y + float32(math.Sin(a1))*ry
		x2 := x + float32(math.Cos(a2))*rx
		y2 := y + float32(math.Sin(a2))*ry
		vector.StrokeLine(wrapper.img, x1, y1, x2, y2, width, clr, antiAlias)
	}
}

func (e *EbitenGraphics) DrawLine(screen Image, x1, y1, x2, y2 float32, clr color.Color, width float32) {
	wrapper, ok := screen.(*EbitenImageWrapper)
	if ok && wrapper != nil && wrapper.img != nil {
		vector.StrokeLine(wrapper.img, x1, y1, x2, y2, width, clr, true)
	}
}

func (e *EbitenGraphics) DrawPolygon(screen Image, points []Point, clr color.Color, width float32) {
	if len(points) < 2 { return }
	wrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || wrapper == nil || wrapper.img == nil { return }
	for i := 0; i < len(points); i++ {
		p1, p2 := points[i], points[(i+1)%len(points)]
		ebitenutil.DrawLine(wrapper.img, float64(p1.X), float64(p1.Y), float64(p2.X), float64(p2.Y), clr)
	}
}

type EbitenShaderWrapper struct {
	shader *ebiten.Shader
}

func (e *EbitenGraphics) NewShader(src []byte) (Shader, error) {
	s, err := ebiten.NewShader(src)
	if err != nil { return nil, err }
	return &EbitenShaderWrapper{shader: s}, nil
}

func (e *EbitenGraphics) DrawFilledPolygon(screen Image, points []Point, clr color.Color, antiAlias bool) {
	if len(points) < 3 { return }
	wrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || wrapper == nil || wrapper.img == nil { return }

	var path vector.Path
	path.MoveTo(float32(points[0].X), float32(points[0].Y))
	for i := 1; i < len(points); i++ {
		path.LineTo(float32(points[i].X), float32(points[i].Y))
	}
	path.Close()

	vertices, indices := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vertices {
		r32, g32, b32, a32 := clr.RGBA()
		vertices[i].ColorR = float32(r32) / 0xffff
		vertices[i].ColorG = float32(g32) / 0xffff
		vertices[i].ColorB = float32(b32) / 0xffff
		vertices[i].ColorA = float32(a32) / 0xffff
	}

	var op ebiten.DrawTrianglesOptions
	if antiAlias { op.FillRule = ebiten.FillRuleEvenOdd }
	wrapper.img.DrawTriangles(vertices, indices, e.whiteImage, &op)
}

func (e *EbitenGraphics) DrawImageWithShader(screen Image, img Image, shader Shader, uniforms map[string]interface{}, options *DrawImageOptions) {
	screenWrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || screenWrapper == nil || screenWrapper.img == nil { return }
	imgWrapper, ok := img.(*EbitenImageWrapper)
	if !ok || imgWrapper == nil || imgWrapper.img == nil { return }
	shaderWrapper, ok := shader.(*EbitenShaderWrapper)
	if !ok || shaderWrapper == nil || shaderWrapper.shader == nil { return }

	w, h := imgWrapper.img.Size()
	op := &ebiten.DrawRectShaderOptions{ Uniforms: uniforms }
	if options != nil {
		op.GeoM = toEbitenGeoM(options.GeoM)
		op.ColorScale = toEbitenColorScale(options.ColorScale)
	}
	op.Images[0] = imgWrapper.img
	screenWrapper.img.DrawRectShader(w, h, shaderWrapper.shader, op)
}

func (e *EbitenGraphics) DrawTriangles(screen Image, vertices []Vertex, indices []uint16, src Image, options *DrawTrianglesOptions) {
	screenWrapper, ok := screen.(*EbitenImageWrapper)
	if !ok || screenWrapper == nil || screenWrapper.img == nil { return }
	srcWrapper, ok := src.(*EbitenImageWrapper)
	if !ok || srcWrapper == nil || srcWrapper.img == nil { return }

	evs := make([]ebiten.Vertex, len(vertices))
	for i, v := range vertices {
		evs[i] = ebiten.Vertex{
			DstX: v.DstX, DstY: v.DstY, SrcX: v.SrcX, SrcY: v.SrcY,
			ColorR: v.ColorR, ColorG: v.ColorG, ColorB: v.ColorB, ColorA: v.ColorA,
		}
	}

	op := &ebiten.DrawTrianglesOptions{}
	if options != nil { op.FillRule = ebiten.FillRule(options.FillRule) }
	screenWrapper.img.DrawTriangles(evs, indices, srcWrapper.img, op)
}
