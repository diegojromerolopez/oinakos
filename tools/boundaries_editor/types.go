package main

import (
	"image"
	"image/color"
	"oinakos/internal/engine"
	"oinakos/internal/game"
)

const (
	defaultScreenWidth  = 1280
	defaultScreenHeight = 720
	cameraSpeed         = 5.0
	vertexRadius        = 3.0
	polyLineWidth       = 2.0
	clickThreshold      = 10.0
	sidebarWidth        = 240
	thumbnailSize       = 64
	slotHeight          = 80
)

var (
	backgroundColor = color.RGBA{20, 20, 20, 255}
	sidebarColor    = color.RGBA{40, 40, 40, 255}
	footprintColor  = color.RGBA{0, 255, 255, 255} // Cyan
	vertexColor     = color.RGBA{255, 255, 0, 255}
	hoverColor      = color.RGBA{255, 255, 255, 255}
	selectedColor   = color.RGBA{0, 255, 0, 255} // Green border
	buttonColor     = color.RGBA{60, 60, 60, 255}
)

type EditorEntity struct {
	ID        string
	Type      string
	Image     engine.Image
	Footprint *[]game.FootprintPoint
	YamlPath  string
	DrawMain  func(screen engine.Image, graphics engine.Graphics, offsetX, offsetY float64)
}

func (e *EditorEntity) GetFootprint() engine.Polygon {
	if e.Footprint == nil || len(*e.Footprint) == 0 {
		return engine.Polygon{Points: []engine.Point{
			{X: -0.15, Y: -0.15}, {X: 0.15, Y: -0.15},
			{X: 0.15, Y: 0.15}, {X: -0.15, Y: 0.15},
		}}
	}
	poly := engine.Polygon{Points: make([]engine.Point, len(*e.Footprint))}
	for i, p := range *e.Footprint {
		poly.Points[i] = engine.Point{X: p.X, Y: p.Y}
	}
	return poly
}

type Viewer struct {
	entities      []*EditorEntity
	selectedIndex int
	graphics      engine.Graphics
	width, height int
	camX, camY    float64
	draggingIdx   int
	hoverIdx      int
	scrollOffset  int
	input         engine.Input
	addBtnRect    image.Rectangle
}
