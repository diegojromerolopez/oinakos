package main

import (
	"image/color"
	"oinakos/internal/engine"
	"oinakos/internal/game"
)

const (
	screenWidth  = 1600
	screenHeight = 900
	sidebarWidth = 240
	slotHeight   = 80
	thumbSize    = 60
)

var (
	colorBG      = color.RGBA{15, 15, 15, 255}
	colorSide    = color.RGBA{35, 35, 35, 255}
	colorText    = color.White
	colorSelect  = color.RGBA{0, 255, 0, 255}
	colorOutline = color.RGBA{50, 50, 50, 255}
	colorModal   = color.RGBA{25, 25, 25, 240}
)

type EditorItem struct {
	ID        string
	Type      string // "obstacle", "npc"
	Image     engine.Image
	Archetype any
}

type MapElement struct {
	Item     *EditorItem
	X, Y     float64
	ID       string
	Selected bool
}

type MapEditor struct {
	Filename string
	MapData  *game.SaveData

	Graphics    engine.Graphics
	Input       engine.Input
	Renderer    *engine.Renderer
	Library     []*EditorItem
	Floors      []string
	FloorImages map[string]engine.Image

	PendingItem *EditorItem
	Selection   *MapElement
	FloorIdx    int

	CamX, CamY float64
	ScrollL    int
	ScrollR    int
	Mode       string // "DIALOG", "EDITOR"

	InName      string
	InWidth     string
	InHeight    string
	ActiveField int
}
