// Package screen provides screenshot capture and image processing utilities.
package screen

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"

	"github.com/go-vgo/robotgo"
)

// CaptureResult contains the result of a screenshot capture.
type CaptureResult struct {
	Image          image.Image // Original captured image
	OriginalWidth  int         // Original width in pixels
	OriginalHeight int         // Original height in pixels
	ScreenIndex    int         // Screen index captured from
}

// Capture takes a screenshot of the specified screen.
// If screenIndex is -1, captures the primary screen.
func Capture(screenIndex int) (*CaptureResult, error) {
	// Set display for multi-monitor support
	if screenIndex >= 0 {
		robotgo.DisplayID = screenIndex
		defer func() { robotgo.DisplayID = -1 }()
	}

	// Capture the screen
	img, err := robotgo.CaptureImg()
	if err != nil {
		return nil, fmt.Errorf("failed to capture screen: %w", err)
	}

	bounds := img.Bounds()
	return &CaptureResult{
		Image:          img,
		OriginalWidth:  bounds.Dx(),
		OriginalHeight: bounds.Dy(),
		ScreenIndex:    screenIndex,
	}, nil
}

// CaptureRegion takes a screenshot of a specific region.
func CaptureRegion(x, y, width, height int) (*CaptureResult, error) {
	img, err := robotgo.CaptureImg(x, y, width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to capture region at (%d, %d) size %dx%d: %w", x, y, width, height, err)
	}

	bounds := img.Bounds()
	return &CaptureResult{
		Image:          img,
		OriginalWidth:  bounds.Dx(),
		OriginalHeight: bounds.Dy(),
		ScreenIndex:    -1,
	}, nil
}

// ProcessedScreenshot contains a processed screenshot ready for LLM consumption.
type ProcessedScreenshot struct {
	Base64         string // Base64-encoded PNG image
	OriginalWidth  int    // Original width before resize
	OriginalHeight int    // Original height before resize
	ScaledWidth    int    // Width after resize
	ScaledHeight   int    // Height after resize
	ScreenIndex    int    // Screen index captured from
}

// CaptureAndProcess captures a screenshot and processes it for LLM use.
// It resizes the image to fit within maxWidth x maxHeight while preserving aspect ratio.
func CaptureAndProcess(screenIndex int, maxWidth, maxHeight int) (*ProcessedScreenshot, error) {
	// Capture
	result, err := Capture(screenIndex)
	if err != nil {
		return nil, err
	}

	// Resize
	resized, scaledW, scaledH := Resize(result.Image, maxWidth, maxHeight)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	// Base64 encode
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &ProcessedScreenshot{
		Base64:         b64,
		OriginalWidth:  result.OriginalWidth,
		OriginalHeight: result.OriginalHeight,
		ScaledWidth:    scaledW,
		ScaledHeight:   scaledH,
		ScreenIndex:    screenIndex,
	}, nil
}

// EncodeToBase64PNG encodes an image to base64 PNG format.
func EncodeToBase64PNG(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// DecodeFromBase64PNG decodes a base64 PNG string to an image.
func DecodeFromBase64PNG(b64 string) (image.Image, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	return img, nil
}
