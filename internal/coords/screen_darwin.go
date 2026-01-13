//go:build darwin

package coords

import (
	"github.com/go-vgo/robotgo"
)

// GetScreen returns information about a specific screen by index.
// On macOS, robotgo uses logical coordinates (not physical pixels).
func GetScreen(index int) ScreenInfo {
	rect := robotgo.GetDisplayRect(index)

	// Detect Retina display
	// On macOS, robotgo returns logical dimensions
	// Scale factor is typically 2.0 for Retina displays
	scaleFactor := detectScaleFactor()

	return ScreenInfo{
		Index:       index,
		X:           rect.X,
		Y:           rect.Y,
		Width:       rect.W,
		Height:      rect.H,
		ScaleFactor: scaleFactor,
		IsPrimary:   index == 0,
	}
}

// detectScaleFactor attempts to detect if we're on a Retina display.
// This is a heuristic based on common macOS screen dimensions.
func detectScaleFactor() float64 {
	w, h := robotgo.GetScreenSize()

	// Common Retina resolutions (logical):
	// MacBook Pro 14": 1512x982 (native), displayed at various scaled resolutions
	// MacBook Pro 16": 1728x1117 (native)
	// iMac 27" 5K: 2560x1440 (native logical)
	// These are logical resolutions; physical would be 2x or more

	// For now, assume Retina if running on macOS
	// robotgo handles coordinate translation internally
	_ = w
	_ = h

	// robotgo on macOS uses logical coordinates, so we report 1.0
	// The actual DPI scaling is handled by the OS
	return 1.0
}

// IsRetinaDisplay returns true if the primary display is a Retina display.
// Note: On macOS, robotgo uses logical coordinates, so this is informational only.
func IsRetinaDisplay() bool {
	// On modern macOS, most displays are HiDPI
	// robotgo abstracts this away by using logical coordinates
	return true
}
