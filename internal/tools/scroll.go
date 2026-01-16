package tools

import (
	"context"
	"time"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
)

// ScrollTool performs scroll operations using normalized coordinates (0-1000 scale).
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
	return `Scroll at a position on the screen. Coordinates are NORMALIZED to 0-1000 scale. First moves the cursor to the specified position, then scrolls in the specified direction.`
}

func (t *ScrollTool) Parameters() map[string]ParameterSpec {
	return map[string]ParameterSpec{
		"x": {
			Type:        "integer",
			Description: "X coordinate normalized 0-1000 (0=left edge, 500=center, 1000=right edge)",
			Required:    true,
		},
		"y": {
			Type:        "integer",
			Description: "Y coordinate normalized 0-1000 (0=top edge, 500=center, 1000=bottom edge)",
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
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide x, y coordinates in 0-1000 scale and direction"), nil
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

	// Validate normalized coordinates
	if args.X < 0 || args.X > 1000 {
		return ErrorResponse("x coordinate out of range", "Use normalized 0-1000 scale"), nil
	}
	if args.Y < 0 || args.Y > 1000 {
		return ErrorResponse("y coordinate out of range", "Use normalized 0-1000 scale"), nil
	}

	// Get screen info
	screenIndex := args.ScreenIndex
	if screenIndex == 0 && t.ScreenIndex != 0 {
		screenIndex = t.ScreenIndex
	}
	screen := coords.GetScreen(screenIndex)

	// Convert normalized coordinates (0-1000) to absolute screen coordinates
	screenX := screen.X + int(float64(args.X)/1000.0*float64(screen.Width))
	screenY := screen.Y + int(float64(args.Y)/1000.0*float64(screen.Height))

	// Move to position first
	robotgo.Move(screenX, screenY)
	time.Sleep(50 * time.Millisecond)

	// Perform scroll
	robotgo.ScrollDir(args.Amount, args.Direction)

	return SuccessResponse(map[string]interface{}{
		"scrolled_at_screen": map[string]int{"x": screenX, "y": screenY},
		"normalized_coords":  map[string]int{"x": args.X, "y": args.Y},
		"screen_dimensions":  map[string]int{"width": screen.Width, "height": screen.Height},
		"direction":          args.Direction,
		"amount":             args.Amount,
		"screen_index":       screenIndex,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *ScrollTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
