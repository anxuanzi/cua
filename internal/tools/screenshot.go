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
	"image/png"

	"github.com/anxuanzi/cua/pkg/logging"
	"github.com/anxuanzi/cua/pkg/screen"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
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

	// Width is the image width in pixels.
	Width int `json:"width,omitempty"`

	// Height is the image height in pixels.
	Height int `json:"height,omitempty"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// takeScreenshot handles the screenshot tool invocation.
func takeScreenshot(ctx tool.Context, args ScreenshotArgs) (ScreenshotResult, error) {
	var img *image.RGBA
	var err error

	// Capture based on arguments
	if args.Region != nil {
		screenshotLog.Start("capture_region", args.Region.X, args.Region.Y, args.Region.Width, args.Region.Height)
		img, err = screen.CaptureRect(screen.Rect{
			X:      args.Region.X,
			Y:      args.Region.Y,
			Width:  args.Region.Width,
			Height: args.Region.Height,
		})
	} else {
		screenshotLog.Start("capture_display", args.DisplayIndex)
		img, err = screen.CaptureDisplay(args.DisplayIndex)
	}

	if err != nil {
		screenshotLog.Failure("screenshot", err)
		return ScreenshotResult{
			Success: false,
			Error:   fmt.Sprintf("failed to capture screen: %v", err),
		}, nil
	}

	bounds := img.Bounds()
	screenshotLog.Debug("captured image: %dx%d", bounds.Dx(), bounds.Dy())

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		screenshotLog.Failure("png_encode", err)
		return ScreenshotResult{
			Success: false,
			Error:   fmt.Sprintf("failed to encode PNG: %v", err),
		}, nil
	}

	// Convert to base64
	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())

	screenshotLog.Success("screenshot", fmt.Sprintf("%dx%d (%d bytes)", bounds.Dx(), bounds.Dy(), buf.Len()))
	return ScreenshotResult{
		Success:     true,
		ImageBase64: base64Img,
		Width:       bounds.Dx(),
		Height:      bounds.Dy(),
	}, nil
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
