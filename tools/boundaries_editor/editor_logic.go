package main

import (
	"image"
	"log"
	"math"
	"os"

	"gopkg.in/yaml.v3"
	"oinakos/internal/engine"
	"oinakos/internal/game"
)

func (v *Viewer) addPoint(ee *EditorEntity) {
	newP := game.FootprintPoint{}
	if len(*ee.Footprint) > 0 {
		last := (*ee.Footprint)[len(*ee.Footprint)-1]
		newP.X = last.X + 0.5
		newP.Y = last.Y + 0.5
	}
	*ee.Footprint = append(*ee.Footprint, newP)
	v.saveToYAML(ee)
}

func (v *Viewer) removePoint(ee *EditorEntity, idx int) {
	fp := *ee.Footprint
	if len(fp) <= 3 {
		log.Println("Cannot remove: polygon must have at least 3 vertices.")
		return
	}
	*ee.Footprint = append(fp[:idx], fp[idx+1:]...)
	v.saveToYAML(ee)
}

func (v *Viewer) saveToYAML(ee *EditorEntity) {
	if ee.YamlPath == "" { return }
	data, err := os.ReadFile(ee.YamlPath)
	if err != nil {
		log.Printf("failed to read yaml: %v", err)
		return
	}
	var m yaml.Node
	if err := yaml.Unmarshal(data, &m); err != nil {
		log.Printf("failed to unmarshal yaml: %v", err)
		return
	}
	fpData, _ := yaml.Marshal(*ee.Footprint)
	var fpNode yaml.Node
	yaml.Unmarshal(fpData, &fpNode)
	if m.Content[0].Kind == yaml.MappingNode {
		found := false
		for i := 0; i < len(m.Content[0].Content); i += 2 {
			if m.Content[0].Content[i].Value == "footprint" {
				m.Content[0].Content[i+1] = fpNode.Content[0]
				found = true
				break
			}
		}
		if !found {
			m.Content[0].Content = append(m.Content[0].Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "footprint"},
				fpNode.Content[0],
			)
		}
	}
	f, err := os.Create(ee.YamlPath)
	if err != nil {
		log.Printf("failed to write yaml: %v", err)
		return
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	enc.Encode(&m)
	log.Println("Footprint saved to", ee.YamlPath)
}

func (v *Viewer) InsertPointOnSegment(ee *EditorEntity, mx, my int, offsetX, offsetY float64) bool {
	if ee.Footprint == nil || len(*ee.Footprint) < 2 {
		return false
	}

	fp := *ee.Footprint
	bestDist := clickThreshold
	bestIdx := -1

	for i := 0; i < len(fp); i++ {
		p1 := fp[i]
		p2 := fp[(i+1)%len(fp)]

		sx1, sy1 := engine.CartesianToIso(p1.X, p1.Y)
		sx2, sy2 := engine.CartesianToIso(p2.X, p2.Y)

		dist := distanceToSegment(float64(mx), float64(my), sx1+offsetX, sy1+offsetY, sx2+offsetX, sy2+offsetY)
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}

	if bestIdx != -1 {
		p1 := fp[bestIdx]
		p2 := fp[(bestIdx+1)%len(fp)]

		newP := game.FootprintPoint{
			X: math.Round((p1.X+p2.X)/2*100) / 100,
			Y: math.Round((p1.Y+p2.Y)/2*100) / 100,
		}

		// Insert at bestIdx + 1
		newFp := make([]game.FootprintPoint, 0, len(fp)+1)
		newFp = append(newFp, fp[:bestIdx+1]...)
		newFp = append(newFp, newP)
		newFp = append(newFp, fp[bestIdx+1:]...)

		*ee.Footprint = newFp
		v.saveToYAML(ee)
		return true
	}

	return false
}

func (v *Viewer) AutoPerimeter(ee *EditorEntity) {
	if ee.Image == nil { return }
	w, h := ee.Image.Size()

	isSolid := func(x, y int) bool {
		if x < 0 || x >= w || y < 0 || y >= h { return false }
		r, g, b, a := ee.Image.At(x, y).RGBA()
		r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)
		if a8 < 50 { return false }
		if g8 > 160 && float64(g8) > float64(r8)*1.5 && float64(g8) > float64(b8)*1.5 { return false }
		return true
	}

	var start image.Point
	found := false
	for y := 0; y < h && !found; y++ {
		for x := 0; x < w && !found; x++ {
			if isSolid(x, y) {
				start = image.Point{x, y}
				found = true
			}
		}
	}
	if !found { return }

	var points []image.Point
	dirs := []image.Point{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}
	curr := start
	dirIdx := 7
	for {
		points = append(points, curr)
		nextFound := false
		for i := 0; i < 8; i++ {
			idx := (dirIdx + i) % 8
			nb := curr.Add(dirs[idx])
			if isSolid(nb.X, nb.Y) {
				curr = nb
				dirIdx = (idx + 5) % 8
				nextFound = true
				break
			}
		}
		if !nextFound || curr == start || len(points) > 3000 { break }
	}

	targetCount := 12
	if len(*ee.Footprint) > targetCount { targetCount = len(*ee.Footprint) }
	if targetCount < 8 { targetCount = 8 }

	step := len(points) / targetCount
	if step < 1 { step = 1 }

	pivotX := float64(w) / 2
	pivotY := float64(h) * 0.85
	if ee.Type == "Object" { pivotY = float64(h) * 0.9 }

	var newFp []game.FootprintPoint
	for i := 0; i < len(points); i += step {
		p := points[i]
		relX, relY := float64(p.X)-pivotX, float64(p.Y)-pivotY
		cx, cy := engine.IsoToCartesian(relX, relY)
		newFp = append(newFp, game.FootprintPoint{X: math.Round(cx*100)/100, Y: math.Round(cy*100)/100})
	}
	if len(newFp) > 2 {
		*ee.Footprint = newFp
		v.saveToYAML(ee)
	}
}

func distanceToSegment(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	if dx == 0 && dy == 0 {
		return math.Hypot(px-x1, py-y1)
	}
	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return math.Hypot(px-(x1+t*dx), py-(y1+t*dy))
}

