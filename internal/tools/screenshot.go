package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
	"golang.org/x/image/draw"
)

const (
	// MaxScreenshotWidth is the maximum width for screenshots sent to the model.
	MaxScreenshotWidth = 1280
	// MaxScreenshotHeight is the maximum height for screenshots sent to the model.
	MaxScreenshotHeight = 800
)

// ScreenshotTool captures screenshots of the screen.
type ScreenshotTool struct {
	BaseTool
	// ScreenIndex specifies which screen to capture (default: 0 = primary).
	ScreenIndex int
}

// NewScreenshotTool creates a new screenshot tool.
func NewScreenshotTool() *ScreenshotTool {
	return &ScreenshotTool{ScreenIndex: 0}
}

func (t *ScreenshotTool) Name() string {
	return "screen_capture"
}

func (t *ScreenshotTool) Description() string {
	return `Capture a screenshot of the current screen. Returns a base64-encoded PNG image along with screen dimensions. Use this tool frequently to understand the current state before taking actions. The screenshot is resized to fit within 1280x800 while preserving aspect ratio for efficient processing.`
}

func (t *ScreenshotTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"screen_index": {
			Type:        "integer",
			Description: "Screen index for multi-monitor setups (0 = primary)",
			Required:    false,
			Default:     0,
		},
	}
}

func (t *ScreenshotTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		ScreenIndex int `json:"screen_index"`
	}
	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide valid JSON with optional screen_index"), nil
	}

	// Use configured screen index if not specified
	screenIndex := args.ScreenIndex
	if screenIndex == 0 && t.ScreenIndex != 0 {
		screenIndex = t.ScreenIndex
	}

	// Set display for capture
	oldDisplayID := robotgo.DisplayID
	robotgo.DisplayID = screenIndex
	defer func() { robotgo.DisplayID = oldDisplayID }()

	// Capture screenshot
	img, err := robotgo.CaptureImg()
	if err != nil {
		return ErrorResponse("failed to capture screenshot: "+err.Error(), "Ensure screen permissions are granted"), nil
	}
	if img == nil {
		return ErrorResponse("failed to capture screenshot: nil image", "Ensure screen permissions are granted"), nil
	}

	// Get original dimensions
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	// Calculate scaled dimensions preserving aspect ratio
	newW, newH := calculateScaledDimensions(origW, origH, MaxScreenshotWidth, MaxScreenshotHeight)

	// Resize using high-quality CatmullRom scaling
	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return ErrorResponse("failed to encode screenshot: "+err.Error(), ""), nil
	}

	// Base64 encode
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Get screen info for additional context
	screen := coords.GetScreen(screenIndex)

	result := map[string]interface{}{
		"image_base64":    b64,
		"original_width":  origW,
		"original_height": origH,
		"scaled_width":    newW,
		"scaled_height":   newH,
		"screen_index":    screenIndex,
		"screen_x":        screen.X,
		"screen_y":        screen.Y,
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *ScreenshotTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}

// calculateScaledDimensions calculates new dimensions that fit within max bounds
// while preserving aspect ratio.
func calculateScaledDimensions(origW, origH, maxW, maxH int) (newW, newH int) {
	if origW <= maxW && origH <= maxH {
		return origW, origH
	}

	aspectRatio := float64(origW) / float64(origH)
	targetAspect := float64(maxW) / float64(maxH)

	if aspectRatio > targetAspect {
		// Width is the limiting factor
		newW = maxW
		newH = int(float64(maxW) / aspectRatio)
	} else {
		// Height is the limiting factor
		newH = maxH
		newW = int(float64(maxH) * aspectRatio)
	}

	return newW, newH
}
