package coords

import (
	"github.com/go-vgo/robotgo"
)

// GetPrimaryScreen returns information about the primary display.
func GetPrimaryScreen() ScreenInfo {
	return GetScreen(0)
}

// GetScreenCount returns the number of available screens.
func GetScreenCount() int {
	// robotgo doesn't expose display count directly
	// We'll probe for screens until we get an invalid response
	count := 1
	for i := 1; i < 10; i++ {
		rect := robotgo.GetDisplayRect(i)
		if rect.W == 0 && rect.H == 0 {
			break
		}
		count++
	}
	return count
}

// GetAllScreens returns information about all available screens.
func GetAllScreens() []ScreenInfo {
	count := GetScreenCount()
	screens := make([]ScreenInfo, count)
	for i := 0; i < count; i++ {
		screens[i] = GetScreen(i)
	}
	return screens
}

// GetScreenAt returns the screen containing the given pixel coordinates.
// Returns the primary screen if no screen contains the point.
func GetScreenAt(x, y int) ScreenInfo {
	screens := GetAllScreens()
	for _, screen := range screens {
		if x >= screen.X && x < screen.X+screen.Width &&
			y >= screen.Y && y < screen.Y+screen.Height {
			return screen
		}
	}
	return GetPrimaryScreen()
}
