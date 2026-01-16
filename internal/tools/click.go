package tools

import (
	"context"
	"time"

	"github.com/anxuanzi/cua/internal/coords"
	"github.com/go-vgo/robotgo"
)

// ClickTool performs mouse clicks at normalized coordinates (0-1000 scale).
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
	return `Click at a position on the screen. Coordinates are NORMALIZED to 0-1000 scale. (0,0) is top-left, (1000,1000) is bottom-right. Example: center of screen = (500, 500), top-right corner = (1000, 0).`
}

func (t *ClickTool) Parameters() map[string]ParameterSpec {
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
		return ErrorResponse("invalid arguments: "+err.Error(), "Provide x and y coordinates in 0-1000 normalized scale"), nil
	}

	// Validate normalized coordinates (allow slight overflow for edge cases)
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
	// Formula: screen_coord = (normalized / 1000) * screen_dimension
	screenX := screen.X + int(float64(args.X)/1000.0*float64(screen.Width))
	screenY := screen.Y + int(float64(args.Y)/1000.0*float64(screen.Height))

	// Move to position with human-like timing
	robotgo.Move(screenX, screenY)

	// Human-like delay after moving (150-200ms feels natural)
	time.Sleep(150 * time.Millisecond)

	// Perform click
	if args.Double {
		robotgo.Click(args.Button, true)
	} else {
		robotgo.Click(args.Button)
	}

	// Delay after clicking to let UI respond
	time.Sleep(100 * time.Millisecond)

	return SuccessResponse(map[string]interface{}{
		"clicked_at_screen": map[string]int{"x": screenX, "y": screenY},
		"normalized_coords": map[string]int{"x": args.X, "y": args.Y},
		"screen_dimensions": map[string]int{"width": screen.Width, "height": screen.Height},
		"button":            args.Button,
		"double_click":      args.Double,
		"screen_index":      screenIndex,
	}), nil
}

// Run implements the interfaces.Tool Run method by delegating to Execute.
func (t *ClickTool) Run(ctx context.Context, input string) (string, error) {
	return t.Execute(ctx, input)
}
