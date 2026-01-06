// Package screen provides cross-platform screen capture functionality.
//
// This package wraps github.com/go-vgo/robotgo to provide a clean,
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

	"github.com/go-vgo/robotgo"
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
	return robotgo.DisplaysNum()
}

// Displays returns information about all connected displays.
func Displays() []Display {
	n := robotgo.DisplaysNum()
	displays := make([]Display, n)

	for i := 0; i < n; i++ {
		x, y, w, h := robotgo.GetDisplayBounds(i)
		displays[i] = Display{
			Index: i,
			Bounds: Rect{
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
			},
			Primary: i == 0,
		}
	}

	return displays
}

// PrimaryDisplay returns information about the primary display.
func PrimaryDisplay() (Display, error) {
	if robotgo.DisplaysNum() == 0 {
		return Display{}, ErrNoDisplays
	}

	x, y, w, h := robotgo.GetDisplayBounds(0)
	return Display{
		Index: 0,
		Bounds: Rect{
			X:      x,
			Y:      y,
			Width:  w,
			Height: h,
		},
		Primary: true,
	}, nil
}

// GetDisplayBounds returns the bounds of the specified display.
func GetDisplayBounds(displayIndex int) (Rect, error) {
	n := robotgo.DisplaysNum()
	if displayIndex < 0 || displayIndex >= n {
		return Rect{}, ErrInvalidDisplay
	}

	x, y, w, h := robotgo.GetDisplayBounds(displayIndex)
	return Rect{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
	}, nil
}

// CaptureDisplay captures the entire screen of the specified display.
// Display index 0 is the primary display.
func CaptureDisplay(displayIndex int) (*image.RGBA, error) {
	n := robotgo.DisplaysNum()
	if displayIndex < 0 || displayIndex >= n {
		return nil, ErrInvalidDisplay
	}

	// Get display bounds
	x, y, w, h := robotgo.GetDisplayBounds(displayIndex)

	// Capture the display region
	img, err := robotgo.CaptureImg(x, y, w, h)
	if err != nil {
		return nil, wrapError(err)
	}

	// Convert to RGBA if needed
	return toRGBA(img), nil
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

	img, err := robotgo.CaptureImg(rect.X, rect.Y, rect.Width, rect.Height)
	if err != nil {
		return nil, wrapError(err)
	}

	return toRGBA(img), nil
}

// Capture captures a rectangular region specified by coordinates.
// This is a convenience function equivalent to CaptureRect.
func Capture(x, y, width, height int) (*image.RGBA, error) {
	return CaptureRect(Rect{X: x, Y: y, Width: width, Height: height})
}

// CaptureAll captures all displays and returns them as a single combined image.
// Displays are arranged according to their actual screen positions.
func CaptureAll() (*image.RGBA, error) {
	n := robotgo.DisplaysNum()
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
		x, y, w, h := robotgo.GetDisplayBounds(i)
		if i == 0 || x < minX {
			minX = x
		}
		if i == 0 || y < minY {
			minY = y
		}
		if i == 0 || x+w > maxX {
			maxX = x + w
		}
		if i == 0 || y+h > maxY {
			maxY = y + h
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

// cachedScaleFactor stores the display scale factor to avoid recalculating.
var cachedScaleFactor float64

// logicalScreenSize stores the logical screen dimensions from the last screenshot.
// These are the dimensions used for Gemini's normalized coordinate system.
// Gemini outputs coords in 0-1000 range which must be denormalized to this size.
var logicalScreenSize struct {
	width  int
	height int
}

// imageSize stores the dimensions of the resized image sent to Gemini.
// This is needed if Gemini outputs pixel coordinates relative to the image it sees.
var imageSize struct {
	width  int
	height int
}

// ScaleFactor returns the display scale factor for the primary display.
// On Retina displays this is typically 2.0, on standard displays it's 1.0.
// This is calculated by comparing the logical display bounds with the
// actual captured image resolution. The result is cached for performance.
func ScaleFactor() float64 {
	if cachedScaleFactor > 0 {
		return cachedScaleFactor
	}

	// Get logical bounds from display
	_, _, logicalW, _ := robotgo.GetDisplayBounds(0)
	if logicalW == 0 {
		cachedScaleFactor = 1.0
		return 1.0
	}

	// Capture a small region to determine physical pixel ratio
	img, err := robotgo.CaptureImg(0, 0, 10, 10)
	if err != nil {
		cachedScaleFactor = 1.0
		return 1.0
	}

	physicalW := img.Bounds().Dx()
	if physicalW == 0 {
		cachedScaleFactor = 1.0
		return 1.0
	}

	// Scale is physical/logical
	cachedScaleFactor = float64(physicalW) / 10.0
	return cachedScaleFactor
}

// SetLogicalScreenSize stores the logical screen dimensions.
// This should be called by the screenshot tool after capturing.
// These dimensions are used to denormalize Gemini's 0-1000 coordinate output.
func SetLogicalScreenSize(width, height int) {
	logicalScreenSize.width = width
	logicalScreenSize.height = height
}

// LogicalScreenSize returns the logical screen dimensions from the last screenshot.
func LogicalScreenSize() (width, height int) {
	return logicalScreenSize.width, logicalScreenSize.height
}

// SetImageSize stores the dimensions of the resized image sent to the AI model.
// This should be called by the screenshot tool after resizing.
func SetImageSize(width, height int) {
	imageSize.width = width
	imageSize.height = height
}

// ImageSize returns the dimensions of the resized image sent to the AI model.
func ImageSize() (width, height int) {
	return imageSize.width, imageSize.height
}

// DenormalizeCoord converts Gemini's normalized 0-1000 coordinates to logical screen coordinates.
// Gemini outputs coordinates in a normalized 0-999 grid regardless of actual screen size.
// This function converts them to actual logical pixel coordinates for mouse input.
//
// Example: If screen is 1440x900 and Gemini outputs (500, 300):
//
//	x = 500 / 1000 * 1440 = 720
//	y = 300 / 1000 * 900  = 270
func DenormalizeCoord(modelX, modelY int) (logicalX, logicalY int) {
	w, h := logicalScreenSize.width, logicalScreenSize.height
	if w == 0 || h == 0 {
		// Fallback: try to get primary display dimensions
		if disp, err := PrimaryDisplay(); err == nil {
			w = disp.Bounds.Width
			h = disp.Bounds.Height
		} else {
			// Last resort fallback
			w, h = 1920, 1080
		}
	}

	// Denormalize: model outputs 0-999 normalized coords
	// Formula: actual = normalized / 1000 * dimension
	// Clamp to valid range [0, dimension-1]
	logicalX = clamp(int(float64(modelX)/1000.0*float64(w)+0.5), 0, w-1)
	logicalY = clamp(int(float64(modelY)/1000.0*float64(h)+0.5), 0, h-1)

	return logicalX, logicalY
}

// clamp restricts a value to be within [min, max].
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// ConvertModelCoord converts model output coordinates to screen coordinates.
// This function auto-detects whether the model output is:
// - Normalized (0-1000 range): Common for Gemini 2.5 Pro and CUA models
// - Image pixels: Actual pixel coordinates in the screenshot
//
// Auto-detection logic:
// 1. If coords > 1000 → must be image pixels (normalized max is ~999)
// 2. If coords >= image dimensions → must be normalized (can't be outside image)
// 3. If image > 1000px and coords < 1000 → likely normalized
// 4. Otherwise → treat as image pixels
//
// Returns the converted screen coordinates and the detected mode for logging.
func ConvertModelCoord(modelX, modelY int) (screenX, screenY int, mode string) {
	imgW, imgH := imageSize.width, imageSize.height

	// Fallback defaults
	if imgW == 0 || imgH == 0 {
		imgW, imgH = 1280, 720
	}

	// Case 1: Coordinates exceed normalized range (0-999)
	// Must be image pixel coordinates
	if modelX > 1000 || modelY > 1000 {
		screenX, screenY = ImageToScreenCoord(modelX, modelY)
		return screenX, screenY, "image_pixel"
	}

	// Case 2: Coordinates exceed image dimensions
	// Must be normalized (you can't have image pixels outside the image)
	if modelX >= imgW || modelY >= imgH {
		screenX, screenY = DenormalizeCoord(modelX, modelY)
		return screenX, screenY, "normalized"
	}

	// Case 3: Ambiguous - coords valid for both interpretations
	// Use heuristic based on image size:
	// - Large images (>1000px) with small coords → likely normalized
	// - Small images (<=1000px) → likely image pixels
	//
	// Research shows Gemini models typically output normalized 0-1000 coords,
	// so we prefer that interpretation when the image is larger than 1000px.
	if imgW > 1000 || imgH > 1000 {
		// Large image + small coords suggests normalized
		screenX, screenY = DenormalizeCoord(modelX, modelY)
		return screenX, screenY, "normalized"
	}

	// Small image (<=1000px in both dimensions)
	// Coords likely represent actual image pixels
	screenX, screenY = ImageToScreenCoord(modelX, modelY)
	return screenX, screenY, "image_pixel"
}

// ImageToScreenCoord converts image pixel coordinates to logical screen coordinates.
// Use this if the AI model outputs coordinates relative to the screenshot image it sees,
// rather than normalized 0-1000 coordinates.
//
// Example: If screen is 1512x982, image is 1280x831, and model outputs (640, 415):
//
//	x = 640 * (1512 / 1280) = 756
//	y = 415 * (1512 / 1280) = 490
func ImageToScreenCoord(imgX, imgY int) (screenX, screenY int) {
	imgW, imgH := imageSize.width, imageSize.height
	scrW, scrH := logicalScreenSize.width, logicalScreenSize.height

	// Fallbacks if not set
	if imgW == 0 || imgH == 0 {
		imgW, imgH = 1280, 720 // Default max dimension
	}
	if scrW == 0 || scrH == 0 {
		if disp, err := PrimaryDisplay(); err == nil {
			scrW = disp.Bounds.Width
			scrH = disp.Bounds.Height
		} else {
			scrW, scrH = 1920, 1080
		}
	}

	// Scale from image coords to screen coords
	scaleX := float64(scrW) / float64(imgW)
	scaleY := float64(scrH) / float64(imgH)

	screenX = clamp(int(float64(imgX)*scaleX+0.5), 0, scrW-1)
	screenY = clamp(int(float64(imgY)*scaleY+0.5), 0, scrH-1)

	return screenX, screenY
}

// PhysicalToLogical converts physical pixel coordinates (from screenshot)
// to logical coordinates (for mouse input).
// Use this when the AI model provides coordinates from a screenshot.
func PhysicalToLogical(physicalX, physicalY int) (logicalX, logicalY int) {
	scale := ScaleFactor()
	return int(float64(physicalX) / scale), int(float64(physicalY) / scale)
}

// LogicalToPhysical converts logical coordinates (mouse input)
// to physical pixel coordinates (screenshot).
func LogicalToPhysical(logicalX, logicalY int) (physicalX, physicalY int) {
	scale := ScaleFactor()
	return int(float64(logicalX) * scale), int(float64(logicalY) * scale)
}

// toRGBA converts any image.Image to *image.RGBA.
func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba
}

// wrapError wraps screenshot errors with our error types.
func wrapError(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(ErrCaptureFailed, err)
}
