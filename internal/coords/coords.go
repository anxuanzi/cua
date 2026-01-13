// Package coords provides coordinate types and conversion utilities for screen interaction.
package coords

const (
	// NormalizedMax is the maximum value for normalized coordinates (0-1000 range).
	NormalizedMax = 1000
)

// Point represents a pixel coordinate on the screen.
type Point struct {
	X, Y int
}

// NormalizedPoint represents a coordinate in the 0-1000 normalized range.
// (0, 0) is top-left, (1000, 1000) is bottom-right.
type NormalizedPoint struct {
	X, Y int
}

// ScreenInfo contains information about a display screen.
type ScreenInfo struct {
	// Index is the display index (0 = primary).
	Index int

	// X, Y are the global offset coordinates for multi-monitor setups.
	X, Y int

	// Width is the screen width in pixels.
	Width int

	// Height is the screen height in pixels.
	Height int

	// ScaleFactor is the DPI scale factor (1.0 for standard, 2.0 for Retina).
	ScaleFactor float64

	// IsPrimary indicates if this is the primary display.
	IsPrimary bool
}
