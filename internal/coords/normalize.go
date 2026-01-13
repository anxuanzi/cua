package coords

// Denormalize converts normalized 0-1000 coordinates to screen pixel coordinates.
// The normalized coordinates are resolution-independent:
//   - (0, 0) maps to top-left corner
//   - (1000, 1000) maps to bottom-right corner
//   - (500, 500) maps to screen center
func Denormalize(norm NormalizedPoint, screen ScreenInfo) Point {
	return Point{
		X: screen.X + (norm.X*screen.Width)/NormalizedMax,
		Y: screen.Y + (norm.Y*screen.Height)/NormalizedMax,
	}
}

// DenormalizeXY is a convenience function that takes x, y integers directly.
func DenormalizeXY(normX, normY int, screen ScreenInfo) (pixelX, pixelY int) {
	point := Denormalize(NormalizedPoint{X: normX, Y: normY}, screen)
	return point.X, point.Y
}

// Normalize converts screen pixel coordinates to normalized 0-1000 coordinates.
// This is useful for reporting positions back to the model.
func Normalize(pixel Point, screen ScreenInfo) NormalizedPoint {
	// Calculate local coordinates relative to this screen
	localX := pixel.X - screen.X
	localY := pixel.Y - screen.Y

	return NormalizedPoint{
		X: (localX * NormalizedMax) / screen.Width,
		Y: (localY * NormalizedMax) / screen.Height,
	}
}

// NormalizeXY is a convenience function that takes x, y integers directly.
func NormalizeXY(pixelX, pixelY int, screen ScreenInfo) (normX, normY int) {
	point := Normalize(Point{X: pixelX, Y: pixelY}, screen)
	return point.X, point.Y
}

// Clamp ensures normalized coordinates are within valid 0-1000 range.
func Clamp(norm NormalizedPoint) NormalizedPoint {
	return NormalizedPoint{
		X: clampInt(norm.X, 0, NormalizedMax),
		Y: clampInt(norm.Y, 0, NormalizedMax),
	}
}

func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
