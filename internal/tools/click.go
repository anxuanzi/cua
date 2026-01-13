package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
)

// ClickTool performs mouse clicks at normalized coordinates.
type ClickTool struct {
	BaseTool
	// ScreenIndex specifies which screen to use (default: 0 = primary).
	ScreenIndex int
}

// NewClickTool creates a new click tool.
func NewClickTool() *ClickTool {
	return &ClickTool{ScreenIndex: 0}
}

func (t *ClickTool) Name() string {
	return "mouse_click"
}

func (t *ClickTool) Description() string {
	return `Click at a position on the screen. Coordinates use a 0-1000 normalized scale where (0,0) is the top-left corner and (1000,1000) is the bottom-right corner. This scale is resolution-independent and works the same on any screen size.`
}

func (t *ClickTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"x": {
			Type:        "integer",
			Description: "X coordinate (0-1000 normalized scale, where 0=left, 1000=right)",
			Required:    true,
		},
		"y": {
			Type:        "integer",
			Description: "Y coordinate (0-1000 normalized scale, where 0=top, 1000=bottom)",
			Required:    true,
		},
		"button": {
			Type:        "string",
			Description: "Mouse button to click",
			Required:    false,
			Default:     "left",
			Enum:        []interface{}{"left", "right", "center"},
		},
		"double": {
			Type:        "boolean",
			Description: "Whether to perform a double-click",
			Required:    false,
			Default:     false,
		},
		"screen_index": {
			Type:        "integer",
			Description: "Screen index for multi-monitor setups (0 = primary)",
			Required:    false,
			Default:     0,
		},
	}
}

func (t *ClickTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		X           int    `json:"x"`
		Y           int    `json:"y"`
		Button      string `json:"button"`
		Double      bool   `json:"double"`
		ScreenIndex int    `json:"screen_index"`
	}
	args.Button = "left" // default

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide x and y coordinates (0-1000)"), nil
	}

	// Validate coordinates
	if args.X < 0 || args.X > coords.NormalizedMax {
		return ErrorResponse(fmt.Sprintf("x coordinate must be 0-%d, got %d", coords.NormalizedMax, args.X), "Use normalized coordinates where 0=left, 1000=right"), nil
	}
	if args.Y < 0 || args.Y > coords.NormalizedMax {
		return ErrorResponse(fmt.Sprintf("y coordinate must be 0-%d, got %d", coords.NormalizedMax, args.Y), "Use normalized coordinates where 0=top, 1000=bottom"), nil
	}

	// Get screen info
	screenIndex := args.ScreenIndex
	if screenIndex == 0 && t.ScreenIndex != 0 {
		screenIndex = t.ScreenIndex
	}
	screen := coords.GetScreen(screenIndex)

	// Denormalize coordinates to screen pixels
	pixel := coords.Denormalize(
		coords.NormalizedPoint{X: args.X, Y: args.Y},
		screen,
	)

	// Move to position
	robotgo.Move(pixel.X, pixel.Y)

	// Small delay to ensure mouse is in position
	time.Sleep(50 * time.Millisecond)

	// Perform click
	if args.Double {
		robotgo.Click(args.Button, true)
	} else {
		robotgo.Click(args.Button)
	}

	return SuccessResponse(map[string]interface{}{
		"clicked_at_pixel": map[string]int{"x": pixel.X, "y": pixel.Y},
		"normalized":       map[string]int{"x": args.X, "y": args.Y},
		"button":           args.Button,
		"double_click":     args.Double,
		"screen_index":     screenIndex,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *ClickTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
