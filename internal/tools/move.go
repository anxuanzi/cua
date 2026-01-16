package tools

import (
	"context"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
)

// MoveTool moves the mouse cursor to a position using normalized coordinates (0-1000 scale).
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
	return `Move the mouse cursor to a position on the screen. Coordinates are NORMALIZED to 0-1000 scale. (0,0) is top-left, (1000,1000) is bottom-right. Example: center of screen = (500, 500).`
}

func (t *MoveTool) Parameters() map[string]ParameterSpec {
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
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide x and y coordinates in 0-1000 normalized scale"), nil
	}

	// Validate normalized coordinates
	if args.X < 0 || args.X > 1000 {
		return ErrorResponse("x coordinate out of range", "Use normalized 0-1000 scale (0=left, 500=center, 1000=right)"), nil
	}
	if args.Y < 0 || args.Y > 1000 {
		return ErrorResponse("y coordinate out of range", "Use normalized 0-1000 scale (0=top, 500=center, 1000=bottom)"), nil
	}

	// Get screen info
	screenIndex := args.ScreenIndex
	if screenIndex == 0 && t.ScreenIndex != 0 {
		screenIndex = t.ScreenIndex
	}
	screen := coords.GetScreen(screenIndex)

	// Convert normalized coordinates (0-1000) to absolute screen coordinates
	// Standard mapping: 0=left/top, 1000=right/bottom (matches TuriX-CUA)
	screenX := screen.X + int(float64(args.X)/1000.0*float64(screen.Width))
	screenY := screen.Y + int(float64(args.Y)/1000.0*float64(screen.Height))

	// Move cursor
	robotgo.Move(screenX, screenY)

	return SuccessResponse(map[string]interface{}{
		"moved_to_screen":   map[string]int{"x": screenX, "y": screenY},
		"normalized_coords": map[string]int{"x": args.X, "y": args.Y},
		"screen_dimensions": map[string]int{"width": screen.Width, "height": screen.Height},
		"screen_index":      screenIndex,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *MoveTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
