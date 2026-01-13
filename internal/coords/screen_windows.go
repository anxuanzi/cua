//go:build windows

package coords

import (
	"github.com/go-vgo/robotgo"
)

// GetScreen returns information about a specific screen by index.
// On Windows, robotgo uses physical pixel coordinates with DPI awareness.
func GetScreen(index int) ScreenInfo {
	rect := robotgo.GetDisplayRect(index)

	return ScreenInfo{
		Index:       index,
		X:           rect.X,
		Y:           rect.Y,
		Width:       rect.W,
		Height:      rect.H,
		ScaleFactor: 1.0, // robotgo abstracts DPI on Windows
		IsPrimary:   index == 0,
	}
}

// IsRetinaDisplay always returns false on Windows.
// Windows uses DPI scaling differently than macOS.
func IsRetinaDisplay() bool {
	return false
}
