package engine

import (
	"image"
	"image/color"
	"math"
)

// Point represents a 2D point in world space.
type Point struct {
	X, Y float64
}

// Polygon represents a shape made of a list of points.
type Polygon struct {
	Points []Point
}

// Transformed returns a new polygon with all points translated by (dx, dy).
func (p Polygon) Transformed(dx, dy float64) Polygon {
	newPoints := make([]Point, len(p.Points))
	for i, pt := range p.Points {
		newPoints[i] = Point{X: pt.X + dx, Y: pt.Y + dy}
	}
	return Polygon{Points: newPoints}
}

// GetEdges returns the vectors representing each edge of the polygon.
func (p Polygon) GetEdges() []Point {
	edges := make([]Point, len(p.Points))
	for i := 0; i < len(p.Points); i++ {
		p1 := p.Points[i]
		p2 := p.Points[(i+1)%len(p.Points)]
		edges[i] = Point{X: p2.X - p1.X, Y: p2.Y - p1.Y}
	}
	return edges
}

// GetNormals returns the normal vectors for each edge.
func (p Polygon) GetNormals() []Point {
	edges := p.GetEdges()
	normals := make([]Point, len(edges))
	for i, edge := range edges {
		// Normal of (x, y) is (-y, x)
		normals[i] = Point{X: -edge.Y, Y: edge.X}
	}
	return normals
}

// Project projects the polygon onto an axis and returns the min/max values.
func (p Polygon) Project(axis Point) (float64, float64) {
	// Normalize axis
	mag := math.Sqrt(axis.X*axis.X + axis.Y*axis.Y)
	if mag == 0 {
		return 0, 0
	}
	normAxis := Point{X: axis.X / mag, Y: axis.Y / mag}

	min := (p.Points[0].X * normAxis.X) + (p.Points[0].Y * normAxis.Y)
	max := min

	for i := 1; i < len(p.Points); i++ {
		proj := (p.Points[i].X * normAxis.X) + (p.Points[i].Y * normAxis.Y)
		if proj < min {
			min = proj
		}
		if proj > max {
			max = proj
		}
	}

	return min, max
}

const ()

// Transparentize returns a new image where the background is made transparent.
// It uses a flood-fill approach from the edges to identify background pixels,
// sampling edge colors to handle checkered or solid backgrounds.
func Transparentize(img image.Image) image.Image {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(bounds)

	// Copy original image first
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newImg.Set(x, y, img.At(x, y))
		}
	}

	// backgroundMask marks pixels to be made transparent
	backgroundMask := make([][]bool, height)
	for i := range backgroundMask {
		backgroundMask[i] = make([]bool, width)
	}

	// Helper to calculate color distance
	colorDist := func(c1, c2 color.Color) float64 {
		r1, g1, b1, _ := c1.RGBA()
		r2, g2, b2, _ := c2.RGBA()
		dr := float64(int32(r1) - int32(r2))
		dg := float64(int32(g1) - int32(g2))
		db := float64(int32(b1) - int32(b2))
		return math.Sqrt(dr*dr + dg*dg + db*db)
	}

	// Sample edge colors to identify likely background colors
	bgSamples := []color.Color{}
	// Take samples from the corners and multiple points along edges
	for i := 0; i <= 4; i++ {
		f := float64(i) / 4.0
		samplePoints := []image.Point{
			{bounds.Min.X + int(f*float64(width-1)), bounds.Min.Y},
			{bounds.Min.X + int(f*float64(width-1)), bounds.Max.Y - 1},
			{bounds.Min.X, bounds.Min.Y + int(f*float64(height-1))},
			{bounds.Max.X - 1, bounds.Min.Y + int(f*float64(height-1))},
		}
		for _, p := range samplePoints {
			bgSamples = append(bgSamples, img.At(p.X, p.Y))
		}
	}

	// 10% tolerance for background matching
	tolerance := 0.10 * 65535.0 * math.Sqrt(3)

	isBackgroundAt := func(x, y int) bool {
		c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
		// Check if it matches any of our edge samples
		for _, bgc := range bgSamples {
			if colorDist(c, bgc) < tolerance {
				return true
			}
		}
		// Catch bright lime green (AI generated backgrounds)
		r, g, b, _ := c.RGBA()
		if g > uint32(200<<8) && r < uint32(150<<8) && b < uint32(150<<8) {
			return true
		}
		return false
	}

	// BFS for flood fill
	type pt struct{ x, y int }
	queue := []pt{}

	// Seed from all four edges
	for x := 0; x < width; x++ {
		queue = append(queue, pt{x, 0}, pt{x, height - 1})
	}
	for y := 1; y < height-1; y++ {
		queue = append(queue, pt{0, y}, pt{width - 1, y})
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.x < 0 || curr.x >= width || curr.y < 0 || curr.y >= height {
			continue
		}
		if backgroundMask[curr.y][curr.x] {
			continue
		}

		if isBackgroundAt(curr.x, curr.y) {
			backgroundMask[curr.y][curr.x] = true
			// Add neighbors
			queue = append(queue, pt{curr.x + 1, curr.y}, pt{curr.x - 1, curr.y}, pt{curr.x, curr.y + 1}, pt{curr.x, curr.y - 1})
		}
	}

	// Apply mask
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if backgroundMask[y][x] {
				newImg.Set(bounds.Min.X+x, bounds.Min.Y+y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return newImg
}

// CheckCollision returns true if two polygons intersect using SAT.
func CheckCollision(p1, p2 Polygon) bool {
	// Check axes of p1
	normals1 := p1.GetNormals()
	for _, axis := range normals1 {
		min1, max1 := p1.Project(axis)
		min2, max2 := p2.Project(axis)
		if max1 < min2 || max2 < min1 {
			return false // Gap found
		}
	}

	// Check axes of p2
	normals2 := p2.GetNormals()
	for _, axis := range normals2 {
		min1, max1 := p1.Project(axis)
		min2, max2 := p2.Project(axis)
		if max1 < min2 || max2 < min1 {
			return false // Gap found
		}
	}

	return true // No gaps found on any axis
}
