package engine

import "image"

// DrawImageOptions abstracts ebiten.DrawImageOptions
type DrawImageOptions struct {
	GeoM       Matrix
	ColorScale ColorScale
	Blend      Blend
}

type Blend int

const (
	BlendSourceOver Blend = iota
	BlendDestinationOut
	BlendDestinationIn
)

type ColorScale struct {
	R, G, B, A float32
}

func NewDrawImageOptions() *DrawImageOptions {
	return &DrawImageOptions{
		GeoM: Matrix{
			m: [2][3]float64{
				{1, 0, 0},
				{0, 1, 0},
			},
		},
		ColorScale: ColorScale{1, 1, 1, 1},
		Blend:      BlendSourceOver,
	}
}

func (d *DrawImageOptions) SetColorScale(r, g, b, a float32) {
	d.ColorScale = ColorScale{r, g, b, a}
}

func (d *DrawImageOptions) Scale(x, y float64) {
	d.GeoM.Scale(x, y)
}

func (d *DrawImageOptions) Translate(tx, ty float64) {
	d.GeoM.Translate(tx, ty)
}

// Matrix abstracts ebiten.GeoM
type Matrix struct {
	m [2][3]float64
}

func NewMatrix() Matrix {
	return Matrix{
		m: [2][3]float64{
			{1, 0, 0},
			{0, 1, 0},
		},
	}
}

func (m *Matrix) Scale(x, y float64) {
	m.m[0][0] *= x
	m.m[0][1] *= x
	m.m[0][2] *= x
	m.m[1][0] *= y
	m.m[1][1] *= y
	m.m[1][2] *= y
}

func (m *Matrix) Reset() {
	m.m[0][0] = 1
	m.m[0][1] = 0
	m.m[0][2] = 0
	m.m[1][0] = 0
	m.m[1][1] = 1
	m.m[1][2] = 0
}

func (m *Matrix) Translate(tx, ty float64) {
	m.m[0][2] += tx
	m.m[1][2] += ty
}

func (m *Matrix) SetElement(row, col int, val float64) {
	m.m[row][col] = val
}

// Image abstracts ebiten.Image for rendering without Ebiten import dependencies
type Image interface {
	Size() (int, int)
	Bounds() image.Rectangle
	DrawImage(img Image, options *DrawImageOptions)
	DrawTriangles(vertices []Vertex, indices []uint16, src Image, options *DrawTrianglesOptions)
	SubImage(r image.Rectangle) Image
	Clear()
	Fill(clr interface{}) // color.Color
}

// Vertex matches ebiten.Vertex
type Vertex struct {
	DstX, DstY float32
	SrcX, SrcY float32
	ColorR     float32
	ColorG     float32
	ColorB     float32
	ColorA     float32
}

type DrawTrianglesOptions struct {
	FillRule FillRule
}

type FillRule int

const (
	FillRuleEvenOdd FillRule = iota
	FillRuleNonZero
)
