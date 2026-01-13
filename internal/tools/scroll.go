package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
)

// ScrollTool performs scroll operations.
type ScrollTool struct {
	BaseTool
	// ScreenIndex specifies which screen to use (default: 0 = primary).
	ScreenIndex int
}

// NewScrollTool creates a new scroll tool.
func NewScrollTool() *ScrollTool {
	return &ScrollTool{ScreenIndex: 0}
}

func (t *ScrollTool) Name() string {
	return "mouse_scroll"
}

func (t *ScrollTool) Description() string {
	return `Scroll at a position on the screen. First moves the cursor to the specified position, then scrolls in the specified direction. Coordinates use 0-1000 normalized scale.`
}

func (t *ScrollTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"x": {
			Type:        "integer",
			Description: "X coordinate to scroll at (0-1000 normalized scale)",
			Required:    true,
		},
		"y": {
			Type:        "integer",
			Description: "Y coordinate to scroll at (0-1000 normalized scale)",
			Required:    true,
		},
		"direction": {
			Type:        "string",
			Description: "Direction to scroll",
			Required:    true,
			Enum:        []interface{}{"up", "down", "left", "right"},
		},
		"amount": {
			Type:        "integer",
			Description: "Number of scroll units (1-10, default: 3)",
			Required:    false,
			Default:     3,
		},
		"screen_index": {
			Type:        "integer",
			Description: "Screen index for multi-monitor setups (0 = primary)",
			Required:    false,
			Default:     0,
		},
	}
}

func (t *ScrollTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		X           int    `json:"x"`
		Y           int    `json:"y"`
		Direction   string `json:"direction"`
		Amount      int    `json:"amount"`
		ScreenIndex int    `json:"screen_index"`
	}
	args.Amount = 3 // default

	if err := ParseArgs(argsJSON, &args); err != nil {
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide x, y coordinates and direction"), nil
	}

	// Validate coordinates
	if args.X < 0 || args.X > coords.NormalizedMax {
		return ErrorResponse(fmt.Sprintf("x must be 0-%d, got %d", coords.NormalizedMax, args.X), ""), nil
	}
	if args.Y < 0 || args.Y > coords.NormalizedMax {
		return ErrorResponse(fmt.Sprintf("y must be 0-%d, got %d", coords.NormalizedMax, args.Y), ""), nil
	}

	// Validate direction
	validDirs := map[string]bool{"up": true, "down": true, "left": true, "right": true}
	if !validDirs[args.Direction] {
		return ErrorResponse("direction must be up, down, left, or right", ""), nil
	}

	// Clamp amount
	if args.Amount < 1 {
		args.Amount = 1
	}
	if args.Amount > 10 {
		args.Amount = 10
	}

	// Get screen info
	screenIndex := args.ScreenIndex
	if screenIndex == 0 && t.ScreenIndex != 0 {
		screenIndex = t.ScreenIndex
	}
	screen := coords.GetScreen(screenIndex)

	// Denormalize coordinates
	pixel := coords.Denormalize(coords.NormalizedPoint{X: args.X, Y: args.Y}, screen)

	// Move to position first
	robotgo.Move(pixel.X, pixel.Y)
	time.Sleep(50 * time.Millisecond)

	// Perform scroll
	robotgo.ScrollDir(args.Amount, args.Direction)

	return SuccessResponse(map[string]interface{}{
		"scrolled_at_pixel": map[string]int{"x": pixel.X, "y": pixel.Y},
		"normalized":        map[string]int{"x": args.X, "y": args.Y},
		"direction":         args.Direction,
		"amount":            args.Amount,
		"screen_index":      screenIndex,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *ScrollTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
