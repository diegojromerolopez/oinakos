package game

import (
	"oinakos/internal/engine"
)

// GetActorSortY returns the value used for depth sorting actors.
func GetActorSortY(a *Actor) float64 {
	sortY := a.X + a.Y
	if a.State == ActorDead {
		sortY -= 100.0
	}
	return sortY
}

// GetObstacleSortY returns the value used for depth sorting obstacles.
func GetObstacleSortY(o *Obstacle) float64 {
	if o.Archetype == nil {
		return o.X + o.Y
	}
	sortY := o.X + o.Y
	if o.Archetype.Type == "static" || o.Archetype.Type == "well" {
		sortY += 2.0
	} else {
		p := o.GetFootprint()
		minX, minY, maxX, maxY := p.Bounds()
		sortY = (minX + maxX + minY + maxY) / 2
	}
	return sortY
}

// IsPointCoveredByObstacle checks if a screen-space point (isoX, isoY) is covered by an obstacle's sprite.
func IsPointCoveredByObstacle(o *Obstacle, isoX, isoY float64) bool {
	if o.Archetype == nil || o.Archetype.Image == nil || !o.Alive {
		return false
	}

	img := o.Archetype.Image
	sw, sh := img.Size()

	frameWidth := sw
	frameHeight := sh
	if o.Archetype.FrameCount > 1 {
		fpr := o.Archetype.FramesPerRow
		if fpr <= 0 {
			fpr = o.Archetype.FrameCount
		}
		numRows := (o.Archetype.FrameCount + fpr - 1) / fpr
		frameWidth = sw / fpr
		frameHeight = sh / numRows
	}

	currentFrame := 0
	if o.Archetype.FrameCount > 1 && o.Archetype.AnimationSpeed > 0 {
		currentFrame = (o.TickCounter / o.Archetype.AnimationSpeed) % o.Archetype.FrameCount
	}

	// Calculate sprite boundaries in screen space (relative to its own pivot)
	oIsoX, oIsoY := engine.CartesianToIso(o.X, o.Y)
	scale := 1.0
	pivotX := float64(frameWidth) * scale / 2
	pivotY := float64(frameHeight) * scale * 0.85

	// Obstacle draw top-left
	tx := oIsoX - pivotX
	ty := oIsoY - pivotY

	// Local coordinates in the sprite
	lx := int(isoX - tx)
	ly := int(isoY - ty)

	if lx < 0 || lx >= frameWidth || ly < 0 || ly >= frameHeight {
		return false
	}

	// Adjust lx, ly for sprite sheets
	if o.Archetype.FrameCount > 1 {
		fpr := o.Archetype.FramesPerRow
		if fpr <= 0 {
			fpr = o.Archetype.FrameCount
		}
		col := currentFrame % fpr
		row := currentFrame / fpr
		lx += col * frameWidth
		ly += row * frameHeight
	}

	// Check alpha
	clr := img.At(lx, ly)
	_, _, _, a := clr.RGBA()
	return a > 0x7FFF // Roughly > 0.5 alpha
}
