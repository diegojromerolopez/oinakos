package engine

type Camera struct {
	X, Y float64
}

func NewCamera(startX, startY float64) *Camera {
	return &Camera{X: startX, Y: startY}
}

func (c *Camera) Follow(targetX, targetY float64, lerp float64) {
	c.X += (targetX - c.X) * lerp
	c.Y += (targetY - c.Y) * lerp
}

func (c *Camera) SnapTo(targetX, targetY float64) {
	c.X = targetX
	c.Y = targetY
}

func (c *Camera) GetOffsets(screenWidth, screenHeight int) (float64, float64) {
	// Center the camera's focus point on the screen center
	return float64(screenWidth)/2 - c.X, float64(screenHeight)/2 - c.Y
}
