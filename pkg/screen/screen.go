// Package screen provides cross-platform screen capture functionality.
//
// This package wraps github.com/kbinani/screenshot to provide a clean,
// consistent API for capturing screenshots on macOS, Windows, and Linux.
//
// # Basic Usage
//
//	// Capture the primary display
//	img, err := screen.CaptureDisplay(0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Capture a specific region
//	img, err := screen.CaptureRect(screen.Rect{X: 0, Y: 0, Width: 800, Height: 600})
//
//	// Get display information
//	displays := screen.Displays()
//	for i, d := range displays {
//	    fmt.Printf("Display %d: %dx%d at (%d,%d)\n", i, d.Width, d.Height, d.X, d.Y)
//	}
package screen

import (
	"errors"
	"image"
	"image/png"
	"io"

	"github.com/kbinani/screenshot"
)

// Rect represents a rectangle on screen in pixel coordinates.
type Rect struct {
	X      int // Left edge
	Y      int // Top edge
	Width  int
	Height int
}

// Display represents information about a connected display.
type Display struct {
	// Index is the display number (0 = primary).
	Index int

	// Bounds contains the display's position and size.
	Bounds Rect

	// Primary indicates if this is the primary display.
	Primary bool
}

// Common errors
var (
	// ErrNoDisplays indicates no displays were found.
	ErrNoDisplays = errors.New("screen: no displays found")

	// ErrInvalidDisplay indicates the display index is out of range.
	ErrInvalidDisplay = errors.New("screen: invalid display index")

	// ErrCaptureFailed indicates screenshot capture failed.
	ErrCaptureFailed = errors.New("screen: capture failed")

	// ErrInvalidRect indicates the capture rectangle is invalid.
	ErrInvalidRect = errors.New("screen: invalid capture rectangle")
)

// NumDisplays returns the number of active displays.
func NumDisplays() int {
	return screenshot.NumActiveDisplays()
}

// Displays returns information about all connected displays.
func Displays() []Display {
	n := screenshot.NumActiveDisplays()
	displays := make([]Display, n)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		displays[i] = Display{
			Index: i,
			Bounds: Rect{
				X:      bounds.Min.X,
				Y:      bounds.Min.Y,
				Width:  bounds.Dx(),
				Height: bounds.Dy(),
			},
			Primary: i == 0,
		}
	}

	return displays
}

// PrimaryDisplay returns information about the primary display.
func PrimaryDisplay() (Display, error) {
	if screenshot.NumActiveDisplays() == 0 {
		return Display{}, ErrNoDisplays
	}

	bounds := screenshot.GetDisplayBounds(0)
	return Display{
		Index: 0,
		Bounds: Rect{
			X:      bounds.Min.X,
			Y:      bounds.Min.Y,
			Width:  bounds.Dx(),
			Height: bounds.Dy(),
		},
		Primary: true,
	}, nil
}

// GetDisplayBounds returns the bounds of the specified display.
func GetDisplayBounds(displayIndex int) (Rect, error) {
	n := screenshot.NumActiveDisplays()
	if displayIndex < 0 || displayIndex >= n {
		return Rect{}, ErrInvalidDisplay
	}

	bounds := screenshot.GetDisplayBounds(displayIndex)
	return Rect{
		X:      bounds.Min.X,
		Y:      bounds.Min.Y,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}, nil
}

// CaptureDisplay captures the entire screen of the specified display.
// Display index 0 is the primary display.
func CaptureDisplay(displayIndex int) (*image.RGBA, error) {
	n := screenshot.NumActiveDisplays()
	if displayIndex < 0 || displayIndex >= n {
		return nil, ErrInvalidDisplay
	}

	img, err := screenshot.CaptureDisplay(displayIndex)
	if err != nil {
		return nil, wrapError(err)
	}

	return img, nil
}

// CapturePrimary captures the entire primary display.
// This is a convenience function equivalent to CaptureDisplay(0).
func CapturePrimary() (*image.RGBA, error) {
	return CaptureDisplay(0)
}

// CaptureRect captures a rectangular region of the screen.
// Coordinates are in global screen space (can span multiple displays).
func CaptureRect(rect Rect) (*image.RGBA, error) {
	if rect.Width <= 0 || rect.Height <= 0 {
		return nil, ErrInvalidRect
	}

	imgRect := image.Rect(rect.X, rect.Y, rect.X+rect.Width, rect.Y+rect.Height)
	img, err := screenshot.CaptureRect(imgRect)
	if err != nil {
		return nil, wrapError(err)
	}

	return img, nil
}

// Capture captures a rectangular region specified by coordinates.
// This is a convenience function equivalent to CaptureRect.
func Capture(x, y, width, height int) (*image.RGBA, error) {
	return CaptureRect(Rect{X: x, Y: y, Width: width, Height: height})
}

// CaptureAll captures all displays and returns them as a single combined image.
// Displays are arranged according to their actual screen positions.
func CaptureAll() (*image.RGBA, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, ErrNoDisplays
	}

	// For single display, just capture it
	if n == 1 {
		return CaptureDisplay(0)
	}

	// Find the bounding box of all displays
	minX, minY := 0, 0
	maxX, maxY := 0, 0

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		if i == 0 || bounds.Min.X < minX {
			minX = bounds.Min.X
		}
		if i == 0 || bounds.Min.Y < minY {
			minY = bounds.Min.Y
		}
		if i == 0 || bounds.Max.X > maxX {
			maxX = bounds.Max.X
		}
		if i == 0 || bounds.Max.Y > maxY {
			maxY = bounds.Max.Y
		}
	}

	// Capture the entire region
	return CaptureRect(Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	})
}

// SavePNG saves an image to a writer in PNG format.
func SavePNG(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}

// wrapError wraps screenshot errors with our error types.
func wrapError(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(ErrCaptureFailed, err)
}
