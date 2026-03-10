package engine

import (
	"image"
	"image/color"
	"io/fs"
)

// MockImage is a headless implementation of Image.
type MockImage struct {
	W, H int
}

func NewMockImage(w, h int) *MockImage {
	return &MockImage{W: w, H: h}
}

func (m *MockImage) Size() (int, int)                               { return m.W, m.H }
func (m *MockImage) Bounds() image.Rectangle                        { return image.Rect(0, 0, m.W, m.H) }
func (m *MockImage) DrawImage(img Image, options *DrawImageOptions) {}
func (m *MockImage) DrawTriangles(vertices []Vertex, indices []uint16, src Image, options *DrawTrianglesOptions) {
}

func (m *MockImage) SubImage(r image.Rectangle) Image {
	return &MockImage{W: r.Dx(), H: r.Dy()}
}
func (m *MockImage) Clear()               {}
func (m *MockImage) Fill(clr interface{}) {}

// MockInput is a mock for engine.Input.
type MockInput struct {
	PressedKeys     map[Key]bool
	JustPressedKeys map[Key]bool
}

func NewMockInput() *MockInput {
	return &MockInput{
		PressedKeys:     make(map[Key]bool),
		JustPressedKeys: make(map[Key]bool),
	}
}

func (m *MockInput) IsKeyPressed(key Key) bool {
	return m.PressedKeys[key]
}

func (m *MockInput) IsKeyJustPressed(key Key) bool {
	return m.JustPressedKeys[key]
}

func (m *MockInput) AppendJustPressedKeys(keys []Key) []Key {
	for k, v := range m.JustPressedKeys {
		if v {
			keys = append(keys, k)
		}
	}
	return keys
}

func (m *MockInput) MousePosition() (x, y int) {
	return 0, 0
}

func (m *MockInput) AppendInputChars(chars []rune) []rune {
	return chars
}

func (m *MockInput) IsMouseButtonPressed(button MouseButton) bool {
	return false
}

func (m *MockInput) IsMouseButtonJustPressed(button MouseButton) bool {
	return false
}

func (m *MockInput) Wheel() (x, y float64) {
	return 0, 0
}

// MockGraphics is a mock for engine.Graphics.
type MockGraphics struct{}

func (m *MockGraphics) NewImage(width, height int) Image {
	return &MockImage{W: width, H: height}
}

func (m *MockGraphics) NewImageFromImage(img image.Image) Image {
	if img == nil {
		return nil
	}
	b := img.Bounds()
	return &MockImage{W: b.Dx(), H: b.Dy()}
}

func (m *MockGraphics) DebugPrintAt(screen Image, str string, x, y int, clr color.Color) {}
func (m *MockGraphics) LoadFont(assets fs.FS, path string) error                       { return nil }
func (m *MockGraphics) DrawTextAt(screen Image, str string, x, y int, clr color.Color, size float64) {
}

func (m *MockGraphics) MeasureText(str string, size float64) (float64, float64) {
	return float64(len(str)) * size * 0.5, size
}
func (m *MockGraphics) DrawFilledRect(screen Image, x, y, width, height float32, clr color.Color, antiAlias bool) {
}
func (m *MockGraphics) DrawFilledCircle(screen Image, x, y, radius float32, clr color.Color, antiAlias bool) {
}
func (m *MockGraphics) DrawFilledEllipse(screen Image, x, y, rx, ry float32, clr color.Color, antiAlias bool) {
}
func (m *MockGraphics) DrawEllipse(screen Image, x, y, rx, ry float32, clr color.Color, width float32, antiAlias bool) {
}
func (m *MockGraphics) DrawTriangles(screen Image, vertices []Vertex, indices []uint16, src Image, options *DrawTrianglesOptions) {
}

func (m *MockGraphics) NewShader(src []byte) (Shader, error) {
	return nil, nil
}

func (m *MockGraphics) DrawImageWithShader(screen Image, img Image, shader Shader, uniforms map[string]interface{}, options *DrawImageOptions) {
}

func (m *MockGraphics) LoadSprite(assets fs.FS, path string, removeBg bool) Image {
	return &MockImage{W: 160, H: 160}
}

func (m *MockGraphics) DrawLine(screen Image, x1, y1, x2, y2 float32, clr color.Color, width float32) {
}

func (m *MockGraphics) DrawPolygon(screen Image, points []Point, clr color.Color, width float32) {
}
