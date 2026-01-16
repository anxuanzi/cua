//go:build darwin

package coords

import (
	"github.com/go-vgo/robotgo"
)

// GetScreen returns information about a specific screen by index.
// On macOS, robotgo uses LOGICAL coordinates (not physical pixels).
// This is important: mouse operations (Move, Click) use logical coords,
// but CaptureImg returns physical (Retina) resolution.
func GetScreen(index int) ScreenInfo {
	rect := robotgo.GetDisplayRect(index)

	// On macOS, robotgo.GetDisplayRect returns LOGICAL dimensions
	// Mouse operations also use logical coordinates
	// Scale factor is detected separately in screenshot tool by comparing
	// capture size (physical) to logical dimensions

	return ScreenInfo{
		Index: index,
		X:     rect.X,
		Y:     rect.Y,
		// Width and Height are LOGICAL dimensions (what mouse coords use)
		Width:  rect.W,
		Height: rect.H,
		// ScaleFactor is informational - actual scale detected by screenshot tool
		// by comparing capture dimensions to logical dimensions
		ScaleFactor: detectScaleFactor(),
		IsPrimary:   index == 0,
	}
}

// detectScaleFactor returns the estimated display scale factor.
// On macOS, this is typically 2.0 for Retina displays.
// Note: The actual scale factor is more accurately calculated in the screenshot
// tool by comparing physical capture dimensions to logical screen dimensions.
func detectScaleFactor() float64 {
	// On modern macOS with Retina displays:
	// - robotgo.GetScreenSize() returns LOGICAL dimensions (e.g., 1512x982)
	// - robotgo.CaptureImg() captures at PHYSICAL resolution (e.g., 3024x1964)
	// - robotgo.Move() uses LOGICAL coordinates
	//
	// We assume 2.0x for Retina displays as a reasonable default.
	// The screenshot tool calculates the actual scale factor dynamically.
	return 2.0
}

// IsRetinaDisplay returns true if the primary display is a Retina display.
// On modern macOS, most built-in displays are HiDPI (Retina).
func IsRetinaDisplay() bool {
	return true
}
