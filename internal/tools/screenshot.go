// Package tools provides ADK tool implementations for desktop automation.
//
// These tools are used by the CUA agents to interact with the desktop.
// Each tool follows the ADK FunctionTool pattern with:
//   - Args struct with json/jsonschema tags
//   - Result struct for structured output
//   - Handler function: func(ctx tool.Context, args Args) (Result, error)
package tools

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"

	"github.com/anxuanzi/cua/pkg/logging"
	"github.com/anxuanzi/cua/pkg/screen"
	"golang.org/x/image/draw"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// Screenshot configuration - can be adjusted for token optimization
var (
	// MaxScreenshotDimension is the maximum width or height for screenshots.
	// Images larger than this will be resized to fit while maintaining aspect ratio.
	// Lower values = fewer tokens but less detail. Default 1280 is good balance.
	MaxScreenshotDimension = 1280

	// ScreenshotQuality is the JPEG quality (1-100). Lower = smaller file but more artifacts.
	// 60 is a good balance for AI vision models.
	ScreenshotQuality = 60
)

var screenshotLog = logging.NewToolLogger("screenshot")

// ScreenshotArgs defines the arguments for the screenshot tool.
type ScreenshotArgs struct {
	// DisplayIndex is the display to capture (0 = primary). Optional.
	DisplayIndex int `json:"display_index,omitzero" jsonschema:"The display index to capture (0 for primary display)"`

	// Region is an optional region to capture instead of full screen.
	// If provided, captures only the specified rectangle.
	Region *ScreenshotRegion `json:"region,omitempty" jsonschema:"Optional region to capture instead of full screen"`
}

// ScreenshotRegion defines a rectangular region to capture.
type ScreenshotRegion struct {
	X      int `json:"x" jsonschema:"Left edge X coordinate"`
	Y      int `json:"y" jsonschema:"Top edge Y coordinate"`
	Width  int `json:"width" jsonschema:"Width in pixels"`
	Height int `json:"height" jsonschema:"Height in pixels"`
}

// ScreenshotResult contains the captured screenshot.
type ScreenshotResult struct {
	// Success indicates if the capture succeeded.
	Success bool `json:"success"`

	// ImageBase64 is the PNG image encoded as base64.
	// This can be sent to vision models for analysis.
	ImageBase64 string `json:"image_base64,omitempty"`

	// Width is the image width in physical pixels.
	Width int `json:"width,omitempty"`

	// Height is the image height in physical pixels.
	Height int `json:"height,omitempty"`

	// ScaleFactor is the display scale factor (e.g., 2.0 for Retina).
	// Coordinates from this image are in physical pixels.
	// The click tool automatically converts these to logical coordinates.
	ScaleFactor float64 `json:"scale_factor,omitempty"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// takeScreenshot handles the screenshot tool invocation.
func takeScreenshot(ctx tool.Context, args ScreenshotArgs) (ScreenshotResult, error) {
	var img *image.RGBA
	var err error

	// Get scale factor first (cached)
	scaleFactor := screen.ScaleFactor()

	// Capture based on arguments
	if args.Region != nil {
		logging.Info("[screenshot] Capturing region: x=%d, y=%d, w=%d, h=%d (scale=%.2f)",
			args.Region.X, args.Region.Y, args.Region.Width, args.Region.Height, scaleFactor)
		img, err = screen.CaptureRect(screen.Rect{
			X:      args.Region.X,
			Y:      args.Region.Y,
			Width:  args.Region.Width,
			Height: args.Region.Height,
		})
	} else {
		logging.Info("[screenshot] Capturing display %d (scale=%.2f)", args.DisplayIndex, scaleFactor)
		img, err = screen.CaptureDisplay(args.DisplayIndex)
	}

	if err != nil {
		logging.Error("[screenshot] Capture failed: %v", err)
		return ScreenshotResult{
			Success: false,
			Error:   fmt.Sprintf("failed to capture screen: %v", err),
		}, nil
	}

	originalBounds := img.Bounds()
	originalW, originalH := originalBounds.Dx(), originalBounds.Dy()
	logging.Info("[screenshot] Captured %dx%d physical pixels", originalW, originalH)

	// Resize if needed to reduce token usage
	resizedImg := resizeImage(img, MaxScreenshotDimension)
	resizedBounds := resizedImg.Bounds()
	resizedW, resizedH := resizedBounds.Dx(), resizedBounds.Dy()

	if resizedW != originalW || resizedH != originalH {
		logging.Info("[screenshot] Resized to %dx%d (max dimension: %d)", resizedW, resizedH, MaxScreenshotDimension)
	}

	// Encode to JPEG (much smaller than PNG)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: ScreenshotQuality}); err != nil {
		logging.Error("[screenshot] JPEG encode failed: %v", err)
		return ScreenshotResult{
			Success: false,
			Error:   fmt.Sprintf("failed to encode JPEG: %v", err),
		}, nil
	}

	// Convert to base64
	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Calculate the effective scale for coordinate mapping
	// The model sees resizedW x resizedH image
	// To convert image coords to logical screen coords:
	//   1. Scale up to original physical: multiply by (originalW / resizedW)
	//   2. Convert physical to logical: divide by scaleFactor
	// Combined: image_coords * (originalW / resizedW) / scaleFactor
	//
	// We store the multiplier so click can do: logical = image_coords * effectiveScale
	// effectiveScale = (originalW / resizedW) / scaleFactor
	effectiveScale := (float64(originalW) / float64(resizedW)) / scaleFactor

	// Store the effective scale for the click tool to use
	screen.SetEffectiveScale(effectiveScale)

	logging.Info("[screenshot] Success: %dx%d → %dx%d, %d bytes (%.1f KB), effective_scale=%.4f (display_scale=%.2f)",
		originalW, originalH, resizedW, resizedH, buf.Len(), float64(buf.Len())/1024, effectiveScale, scaleFactor)

	return ScreenshotResult{
		Success:     true,
		ImageBase64: base64Img,
		Width:       resizedW,
		Height:      resizedH,
		ScaleFactor: effectiveScale, // This is the multiplier to convert image coords to logical coords
	}, nil
}

// resizeImage resizes an image so its largest dimension is at most maxDim.
// Returns the original image if no resizing is needed.
func resizeImage(img *image.RGBA, maxDim int) *image.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// No resize needed
	if w <= maxDim && h <= maxDim {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newW, newH int
	if w > h {
		newW = maxDim
		newH = int(float64(h) * float64(maxDim) / float64(w))
	} else {
		newH = maxDim
		newW = int(float64(w) * float64(maxDim) / float64(h))
	}

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	return resized
}

// NewScreenshotTool creates the screenshot tool for ADK agents.
func NewScreenshotTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "screenshot",
			Description: "Captures a screenshot of the screen. Can capture the full display or a specific region. Returns the image as base64-encoded PNG.",
		},
		takeScreenshot,
	)
}
