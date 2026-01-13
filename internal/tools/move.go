package tools

import (
	"context"
	"fmt"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
)

// MoveTool moves the mouse cursor without clicking.
type MoveTool struct {
	BaseTool
	// ScreenIndex specifies which screen to use (default: 0 = primary).
	ScreenIndex int
}

// NewMoveTool creates a new move tool.
func NewMoveTool() *MoveTool {
	return &MoveTool{ScreenIndex: 0}
}

func (t *MoveTool) Name() string {
	return "mouse_move"
}

func (t *MoveTool) Description() string {
	return `Move the mouse cursor to a position without clicking. Coordinates use a 0-1000 normalized scale where (0,0) is the top-left corner and (1000,1000) is the bottom-right corner. Use this for hover actions or to position the cursor before other actions.`
}

func (t *MoveTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"x": {
			Type:        "integer",
			Description: "X coordinate (0-1000 normalized scale)",
			Required:    true,
		},
		"y": {
			Type:        "integer",
			Description: "Y coordinate (0-1000 normalized scale)",
			Required:    true,
		},
		"screen_index": {
			Type:        "integer",
			Description: "Screen index for multi-monitor setups (0 = primary)",
			Required:    false,
			Default:     0,
		},
	}
}

func (t *MoveTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		X           int `json:"x"`
		Y           int `json:"y"`
		ScreenIndex int `json:"screen_index"`
	}

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide x and y coordinates (0-1000)"), nil
	}

	// Validate coordinates
	if args.X < 0 || args.X > coords.NormalizedMax {
		return ErrorResponse(fmt.Sprintf("x coordinate must be 0-%d, got %d", coords.NormalizedMax, args.X), ""), nil
	}
	if args.Y < 0 || args.Y > coords.NormalizedMax {
		return ErrorResponse(fmt.Sprintf("y coordinate must be 0-%d, got %d", coords.NormalizedMax, args.Y), ""), nil
	}

	// Get screen info
	screenIndex := args.ScreenIndex
	if screenIndex == 0 && t.ScreenIndex != 0 {
		screenIndex = t.ScreenIndex
	}
	screen := coords.GetScreen(screenIndex)

	// Denormalize coordinates
	pixel := coords.Denormalize(
		coords.NormalizedPoint{X: args.X, Y: args.Y},
		screen,
	)

	// Move cursor
	robotgo.Move(pixel.X, pixel.Y)

	return SuccessResponse(map[string]interface{}{
		"moved_to_pixel": map[string]int{"x": pixel.X, "y": pixel.Y},
		"normalized":     map[string]int{"x": args.X, "y": args.Y},
		"screen_index":   screenIndex,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *MoveTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
