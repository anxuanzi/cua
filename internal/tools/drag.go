package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
)

// DragTool performs mouse drag operations.
type DragTool struct {
	BaseTool
	// ScreenIndex specifies which screen to use (default: 0 = primary).
	ScreenIndex int
}

// NewDragTool creates a new drag tool.
func NewDragTool() *DragTool {
	return &DragTool{ScreenIndex: 0}
}

func (t *DragTool) Name() string {
	return "mouse_drag"
}

func (t *DragTool) Description() string {
	return `Drag from one position to another. All coordinates use 0-1000 normalized scale. This performs a mouse press at the start position, moves to the end position, then releases.`
}

func (t *DragTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"start_x": {
			Type:        "integer",
			Description: "Starting X coordinate (0-1000 normalized scale)",
			Required:    true,
		},
		"start_y": {
			Type:        "integer",
			Description: "Starting Y coordinate (0-1000 normalized scale)",
			Required:    true,
		},
		"end_x": {
			Type:        "integer",
			Description: "Ending X coordinate (0-1000 normalized scale)",
			Required:    true,
		},
		"end_y": {
			Type:        "integer",
			Description: "Ending Y coordinate (0-1000 normalized scale)",
			Required:    true,
		},
		"button": {
			Type:        "string",
			Description: "Mouse button to use for dragging",
			Required:    false,
			Default:     "left",
			Enum:        []interface{}{"left", "right", "center"},
		},
		"screen_index": {
			Type:        "integer",
			Description: "Screen index for multi-monitor setups (0 = primary)",
			Required:    false,
			Default:     0,
		},
	}
}

func (t *DragTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		StartX      int    `json:"start_x"`
		StartY      int    `json:"start_y"`
		EndX        int    `json:"end_x"`
		EndY        int    `json:"end_y"`
		Button      string `json:"button"`
		ScreenIndex int    `json:"screen_index"`
	}
	args.Button = "left" // default

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide start_x, start_y, end_x, end_y coordinates (0-1000)"), nil
	}

	// Validate coordinates
	for _, coord := range []struct {
		name string
		val  int
	}{
		{"start_x", args.StartX}, {"start_y", args.StartY},
		{"end_x", args.EndX}, {"end_y", args.EndY},
	} {
		if coord.val < 0 || coord.val > coords.NormalizedMax {
			return ErrorResponse(fmt.Sprintf("%s must be 0-%d, got %d", coord.name, coords.NormalizedMax, coord.val), ""), nil
		}
	}

	// Get screen info
	screenIndex := args.ScreenIndex
	if screenIndex == 0 && t.ScreenIndex != 0 {
		screenIndex = t.ScreenIndex
	}
	screen := coords.GetScreen(screenIndex)

	// Denormalize coordinates
	startPixel := coords.Denormalize(coords.NormalizedPoint{X: args.StartX, Y: args.StartY}, screen)
	endPixel := coords.Denormalize(coords.NormalizedPoint{X: args.EndX, Y: args.EndY}, screen)

	// Perform drag: move to start, press, move to end, release
	robotgo.Move(startPixel.X, startPixel.Y)
	time.Sleep(50 * time.Millisecond)

	robotgo.Toggle(args.Button, "down")
	time.Sleep(50 * time.Millisecond)

	// Smooth drag with intermediate steps for better reliability
	steps := 10
	for i := 1; i <= steps; i++ {
		x := startPixel.X + (endPixel.X-startPixel.X)*i/steps
		y := startPixel.Y + (endPixel.Y-startPixel.Y)*i/steps
		robotgo.Move(x, y)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)
	robotgo.Toggle(args.Button, "up")

	return SuccessResponse(map[string]interface{}{
		"dragged_from_pixel": map[string]int{"x": startPixel.X, "y": startPixel.Y},
		"dragged_to_pixel":   map[string]int{"x": endPixel.X, "y": endPixel.Y},
		"normalized_from":    map[string]int{"x": args.StartX, "y": args.StartY},
		"normalized_to":      map[string]int{"x": args.EndX, "y": args.EndY},
		"button":             args.Button,
		"screen_index":       screenIndex,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *DragTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
