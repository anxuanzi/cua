package tools

import (
	"fmt"

	"github.com/anxuanzi/cua/pkg/input"
	"github.com/anxuanzi/cua/pkg/logging"
	"github.com/anxuanzi/cua/pkg/screen"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// ClickArgs defines the arguments for the click tool.
type ClickArgs struct {
	// X is the X coordinate to click (in physical pixels from screenshot).
	// Automatically converted to logical coordinates for mouse input.
	X int `json:"x" jsonschema:"X coordinate in physical pixels (from screenshot)"`

	// Y is the Y coordinate to click (in physical pixels from screenshot).
	// Automatically converted to logical coordinates for mouse input.
	Y int `json:"y" jsonschema:"Y coordinate in physical pixels (from screenshot)"`

	// ClickType specifies the type of click: "left", "right", or "double".
	// Defaults to "left" if not specified.
	ClickType string `json:"click_type,omitempty" jsonschema:"Type of click: 'left', 'right', or 'double'. Defaults to 'left'"`
}

// ClickResult contains the result of a click operation.
type ClickResult struct {
	// Success indicates if the click succeeded.
	Success bool `json:"success"`

	// X is the X coordinate that was clicked.
	X int `json:"x"`

	// Y is the Y coordinate that was clicked.
	Y int `json:"y"`

	// ClickType is the type of click performed.
	ClickType string `json:"click_type"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// performClick handles the click tool invocation.
func performClick(ctx tool.Context, args ClickArgs) (ClickResult, error) {
	clickType := args.ClickType
	if clickType == "" {
		clickType = "left"
	}

	// Get the effective scale from the last screenshot
	// This includes both display scaling and any image resizing
	effectiveScale := screen.EffectiveScale()

	// Convert image coordinates to logical screen coordinates
	// logical = image_coords * effectiveScale
	logicalX := int(float64(args.X) * effectiveScale)
	logicalY := int(float64(args.Y) * effectiveScale)

	logging.Info("[click] %s at image(%d, %d) → logical(%d, %d) [effective_scale=%.4f]",
		clickType, args.X, args.Y, logicalX, logicalY, effectiveScale)

	point := input.Point{X: logicalX, Y: logicalY}
	var err error

	switch clickType {
	case "left":
		err = input.Click(point)
	case "right":
		err = input.RightClick(point)
	case "double":
		err = input.DoubleClick(point)
	default:
		logging.Error("[click] Invalid click type: %s", clickType)
		return ClickResult{
			Success:   false,
			X:         args.X,
			Y:         args.Y,
			ClickType: clickType,
			Error:     fmt.Sprintf("invalid click type: %s (must be 'left', 'right', or 'double')", clickType),
		}, nil
	}

	if err != nil {
		logging.Error("[click] Failed: %v", err)
		return ClickResult{
			Success:   false,
			X:         args.X,
			Y:         args.Y,
			ClickType: clickType,
			Error:     fmt.Sprintf("click failed: %v", err),
		}, nil
	}

	logging.Info("[click] Success: %s at logical(%d, %d)", clickType, logicalX, logicalY)
	return ClickResult{
		Success:   true,
		X:         args.X,
		Y:         args.Y,
		ClickType: clickType,
	}, nil
}

// NewClickTool creates the click tool for ADK agents.
func NewClickTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "click",
			Description: "Performs a mouse click at the specified screen coordinates. Supports left click, right click, and double click.",
		},
		performClick,
	)
}
