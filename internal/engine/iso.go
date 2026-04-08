package engine

// TileWidth and TileHeight define the dimensions of a single isometric tile.
// A common ratio is 2:1 (e.g., 64x32 or 128x64).
const (
	TileWidth  = 64
	TileHeight = 32
)

// VerticalScale defines how many screen pixels correspond to 1 unit of Cartesian height (z).
const VerticalScale float64 = 16.0

// CartesianToIsoZ converts 3D world coordinates (x, y, z) to screen coordinates.
// x and y are the grid positions, z is the elevation.
func CartesianToIsoZ(x, y, z float64) (float64, float64) {
	isoX := (x - y) * (float64(TileWidth) / 2)
	isoY := (x + y) * (float64(TileHeight) / 2) - (z * VerticalScale)
	return isoX, isoY
}

// CartesianToIso converts world coordinates (x, y) to screen coordinates assuming z=0.
// x and y are the grid positions (0, 1, 2...).
func CartesianToIso(x, y float64) (float64, float64) {
	return CartesianToIsoZ(x, y, 0)
}

// IsoToCartesian converts screen coordinates back to grid positions.
func IsoToCartesian(screenX, screenY float64) (float64, float64) {
	halfW := float64(TileWidth) / 2
	halfH := float64(TileHeight) / 2
	x := (screenX/halfW + screenY/halfH) / 2
	y := (screenY/halfH - screenX/halfW) / 2
	return x, y
}
