//go:build linux

package coords

import (
	"github.com/go-vgo/robotgo"
)

// GetScreen returns information about a specific screen by index.
// On Linux, behavior depends on the display server (X11 or Wayland).
func GetScreen(index int) ScreenInfo {
	rect := robotgo.GetDisplayRect(index)

	return ScreenInfo{
		Index:       index,
		X:           rect.X,
		Y:           rect.Y,
		Width:       rect.W,
		Height:      rect.H,
		ScaleFactor: 1.0, // robotgo uses native coordinates on Linux
		IsPrimary:   index == 0,
	}
}

// IsRetinaDisplay always returns false on Linux.
func IsRetinaDisplay() bool {
	return false
}
