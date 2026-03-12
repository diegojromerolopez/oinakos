package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"path"
	"oinakos/internal/engine"
)

func (gr *GameRenderer) getTileAt(x, y int) engine.Image {
	key := fmt.Sprintf("%d_%d", x, y)

	if tilePath, ok := gr.coordCache[key]; ok {
		return gr.tileCache[tilePath]
	}

	resolvedTile := gr.game.currentMapType.FloorTile
	highestPriority := -1

	for _, zone := range gr.game.currentMapType.FloorZones {
		if zone.Priority > highestPriority {
			if zone.Contains(float64(x), float64(y)) {
				resolvedTile = zone.Tile
				highestPriority = zone.Priority
			}
		}
	}

	switch resolvedTile {
	case "grass.png", "mud.png", "paved_ground.png", "dirt.png", "big_stones.png",
		"wheat_field.png", "desert_sand.png", "yellow_grass.png", "dry_ground.png",
		"water.png", "dark_water.png":
		
		hash := (x*73856093 ^ y*19349663)
		if hash < 0 {
			hash = -hash
		}
		variant := hash % 3
		if variant == 1 {
			resolvedTile = resolvedTile[:len(resolvedTile)-4] + "_2.png"
		} else if variant == 2 {
			resolvedTile = resolvedTile[:len(resolvedTile)-4] + "_3.png"
		}
	}

	gr.coordCache[key] = resolvedTile

	if _, ok := gr.tileCache[resolvedTile]; !ok {
		floorPath := path.Join("assets/images/floors", resolvedTile)
		loaded := gr.graphics.LoadSprite(gr.game.assets, floorPath, true)
		gr.tileCache[resolvedTile] = loaded
	}

	return gr.tileCache[resolvedTile]
}

func (gr *GameRenderer) drawFog(screen engine.Image) {
	g := gr.game
	if g.settings == nil || g.settings.FogOfWar == "none" {
		return
	}

	w, h := screen.Size()
	if gr.fogImage == nil {
		gr.fogImage = gr.graphics.NewImage(w, h)
	}

	gr.fogImage.Fill(color.Black)

	offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)
	px, py := g.playableCharacter.X, g.playableCharacter.Y
	isoX, isoY := engine.CartesianToIso(px, py)

	op := engine.NewDrawImageOptions()
	op.Blend = engine.BlendDestinationOut

	if g.settings.FogOfWar == "vision" {
		tw, th := gr.torchImage.Size()
		op.GeoM.Translate(isoX+offsetX-float64(tw/2), isoY+offsetY-float64(th/2))
		gr.fogImage.DrawImage(gr.torchImage, op)
	} else if g.settings.FogOfWar == "exploration" {
		for pt := range g.ExploredTiles {
			eisoX, eisoY := engine.CartesianToIso(float64(pt.X), float64(pt.Y))
			scrX, scrY := eisoX+offsetX, eisoY+offsetY
			
			if scrX < -64 || scrX > float64(g.width)+64 || scrY < -64 || scrY > float64(g.height)+64 {
				continue
			}
			
			rectOp := engine.NewDrawImageOptions()
			rectOp.Blend = engine.BlendDestinationOut
			rectOp.GeoM.Scale(64.0/3.0, 32.0/3.0)
			rectOp.GeoM.Translate(scrX-32, scrY-16)
			gr.fogImage.DrawImage(gr.emptyImage, rectOp)
		}
		
		tw, th := gr.torchImage.Size()
		op.GeoM.Reset()
		op.GeoM.Translate(isoX+offsetX-float64(tw/2), isoY+offsetY-float64(th/2))
		gr.fogImage.DrawImage(gr.torchImage, op)
	}

	screen.DrawImage(gr.fogImage, nil)
}

func generateTorchImage(g engine.Graphics, radius int) engine.Image {
	size := radius * 2
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - radius)
			dy := float64(y - radius)
			distSq := dx*dx + dy*dy
			radSq := float64(radius * radius)
			if distSq < radSq {
				alpha := uint8(255 * (1 - math.Sqrt(distSq/radSq)))
				img.SetRGBA(x, y, color.RGBA{255, 255, 255, alpha})
			}
		}
	}
	return g.NewImageFromImage(img)
}

func (gr *GameRenderer) drawObjectiveArrow(screen engine.Image) {
	g := gr.game
	targetX, targetY := 0.0, 0.0
	hasTarget := false

	switch g.currentMapType.Type {
	case ObjReachPortal, ObjReachZone, ObjReachBuilding, ObjProtectNPC, ObjDestroyBuilding:
		targetX, targetY = g.currentMapType.TargetPoint.X, g.currentMapType.TargetPoint.Y
		hasTarget = true
	case ObjKillVIP:
		for _, n := range g.npcs {
			if n.Alignment == AlignmentEnemy && n.IsAlive() {
				targetX, targetY = n.X, n.Y
				hasTarget = true
				break
			}
		}
	}

	if !hasTarget {
		return
	}

	dx := targetX - g.playableCharacter.X
	dy := targetY - g.playableCharacter.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 4.0 {
		return
	}

	arrowX := float64(g.width - 50)
	arrowY := 120.0

	screenDX := dx - dy
	screenDY := (dx + dy) * 0.5
	screenAngle := math.Atan2(screenDY, screenDX)

	size := 15.0
	p1 := engine.Point{X: arrowX + math.Cos(screenAngle)*size, Y: arrowY + math.Sin(screenAngle)*size}
	p2 := engine.Point{X: arrowX + math.Cos(screenAngle+2.5)*size*0.7, Y: arrowY + math.Sin(screenAngle+2.5)*size*0.7}
	p3 := engine.Point{X: arrowX + math.Cos(screenAngle-2.5)*size*0.7, Y: arrowY + math.Sin(screenAngle-2.5)*size*0.7}

	p1s := engine.Point{X: p1.X + 2, Y: p1.Y + 2}
	p2s := engine.Point{X: p2.X + 2, Y: p2.Y + 2}
	p3s := engine.Point{X: p3.X + 2, Y: p3.Y + 2}
	gr.graphics.DrawFilledPolygon(screen, []engine.Point{p1s, p2s, p3s}, color.RGBA{0, 0, 0, 150}, true)
	gr.graphics.DrawFilledPolygon(screen, []engine.Point{p1, p2, p3}, color.RGBA{220, 20, 60, 255}, true)
}

func (gr *GameRenderer) drawDebug(screen engine.Image, offsetX, offsetY float64) {
	red := color.RGBA{255, 0, 0, 255}
	green := color.RGBA{0, 255, 0, 255}
	cyan := color.RGBA{0, 255, 255, 255}

	drawPolygon := func(poly engine.Polygon, clr color.Color) {
		isoPoints := make([]engine.Point, len(poly.Points))
		for i, p := range poly.Points {
			ix, iy := engine.CartesianToIso(p.X, p.Y)
			isoPoints[i] = engine.Point{X: ix + offsetX, Y: iy + offsetY}
		}
		gr.graphics.DrawPolygon(screen, isoPoints, clr, 1.0)
	}

	for _, o := range gr.game.obstacles {
		drawPolygon(o.GetFootprint(), cyan)
	}

	for _, n := range gr.game.npcs {
		clr := red
		if n.Alignment == AlignmentAlly {
			clr = green
		} else if n.Alignment == AlignmentNeutral {
			clr = cyan
		}
		drawPolygon(n.GetFootprint(), clr)
	}

	drawPolygon(gr.game.playableCharacter.GetFootprint(), green)
}
