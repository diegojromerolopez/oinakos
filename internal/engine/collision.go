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

// Bounds returns the axis-aligned bounding box (minX, minY, maxX, maxY).
func (p Polygon) Bounds() (minX, minY, maxX, maxY float64) {
	if len(p.Points) == 0 {
		return 0, 0, 0, 0
	}
	minX, maxX = p.Points[0].X, p.Points[0].X
	minY, maxY = p.Points[0].Y, p.Points[0].Y
	for _, pt := range p.Points[1:] {
		if pt.X < minX {
			minX = pt.X
		} else if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		} else if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	return
}

// Contains returns true if the point (x, y) is inside the polygon using ray-casting.
func (p Polygon) Contains(x, y float64) bool {
	if len(p.Points) < 3 {
		return false
	}

	// Fast AABB check
	minX, minY, maxX, maxY := p.Bounds()
	if x < minX || x > maxX || y < minY || y > maxY {
		return false
	}

	inside := false
	j := len(p.Points) - 1
	for i := 0; i < len(p.Points); i++ {
		xi, yi := p.Points[i].X, p.Points[i].Y
		xj, yj := p.Points[j].X, p.Points[j].Y

		intersect := ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
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

// Transparentize returns a new image where the solid lime green (#00FF00) background is made transparent.
// This matches the project's standard for asset generation.
func Transparentize(img image.Image) image.Image {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			// PROJECT STANDARD: Solid lime green background (#00FF00) removal.
			// Ebiten/Go use 16-bit RGBA (0-65535).
			// We use a robust ratio-based approach to catch AI-generated lime.
			isLime := false
			if a > 0 {
				r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
				if float64(g8) > 160 && float64(g8) > float64(r8)*1.5 && float64(g8) > float64(b8)*1.5 {
					isLime = true
				}
			}

			if isLime {
				newImg.Set(x, y, color.Transparent)
				count++
			} else {
				newImg.Set(x, y, c)
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
